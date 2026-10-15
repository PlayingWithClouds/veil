package cams

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/plugins"
)

const (
	playlistName = "index.m3u8"
	posterName   = "poster.jpg"
	// recordingTagLimit caps how many of the room's own tags a scene takes.
	recordingTagLimit = 8
	// recordingListLimit caps the finished recordings returned for one query.
	recordingListLimit = 100
)

// recordingRow mirrors the cam_recording table joined with the model's name.
type recordingRow struct {
	ID              string  `json:"id"`
	CamModel        string  `json:"cam_model"`
	Scene           *string `json:"scene"`
	Status          string  `json:"status"`
	StopReason      *string `json:"stop_reason"`
	StartedAt       string  `json:"started_at"`
	EndedAt         *string `json:"ended_at"`
	DurationSeconds float64 `json:"duration_seconds"`
	SizeBytes       int64   `json:"size_bytes"`
	DisplayName     string  `json:"display_name"`
	ThumbnailURL    *string `json:"thumbnail_url"`
	Quality         *string `json:"quality"`
}

// toModel maps a stored recording to the GraphQL model.
func (row recordingRow) toModel() *model.CamRecording {
	return &model.CamRecording{
		ID: row.ID, CamModelID: row.CamModel, DisplayName: row.DisplayName, ThumbnailURL: row.ThumbnailURL, Quality: row.Quality, Status: row.Status,
		StopReason: row.StopReason, StartedAt: row.StartedAt, EndedAt: row.EndedAt,
		DurationSeconds: row.DurationSeconds, SizeBytes: float64(row.SizeBytes), SceneID: row.Scene,
	}
}

const recordingSelect = `SELECT r.*, m.display_name, m.thumbnail_url FROM cam_recording r JOIN cam_model m ON m.id = r.cam_model`

// insertRow stores the row of a capture that has just started.
func (r *Recorder) insertRow(ctx context.Context, recording *capture) error {
	_, err := r.service.database.Exec(ctx,
		`INSERT INTO cam_recording (id, cam_model, status, started_at, quality) VALUES ($id, $model, $status, $startedAt, $quality)`,
		db.Vars{"id": recording.recordID, "model": recording.camera.ID, "status": statusRecording,
			"startedAt": formatTime(recording.startedAt), "quality": derefOrNil(recording.quality())})
	return err
}

// derefOrNil binds an absent string as NULL.
func derefOrNil(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

// loadRecording returns one stored recording.
func (r *Recorder) loadRecording(ctx context.Context, recordID string) (*model.CamRecording, error) {
	row, err := db.QueryOneAs[recordingRow](ctx, r.service.database, recordingSelect+` WHERE r.id = $id`, db.Vars{"id": recordID})
	if err != nil || row == nil {
		return nil, err
	}
	return row.toModel(), nil
}

// Recordings returns the recordings in progress followed by the finished ones,
// newest first. cameraID narrows it to one model; activeOnly skips finished ones.
func (r *Recorder) Recordings(ctx context.Context, cameraID string, activeOnly bool) ([]*model.CamRecording, error) {
	var recordings []*model.CamRecording
	for _, active := range r.Active() {
		if cameraID == "" || active.CamModelID == cameraID {
			recordings = append(recordings, active)
		}
	}
	if activeOnly {
		return recordings, nil
	}
	rows, err := db.QueryAs[recordingRow](ctx, r.service.database,
		recordingSelect+` WHERE r.status = 'finished' AND ($model = '' OR r.cam_model = $model)
		 ORDER BY r.started_at DESC LIMIT $limit`,
		db.Vars{"model": cameraID, "limit": recordingListLimit})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		recordings = append(recordings, row.toModel())
	}
	return recordings, nil
}

// sortRecordings orders recordings oldest first.
func sortRecordings(recordings []*model.CamRecording) {
	sort.SliceStable(recordings, func(first, second int) bool {
		return recordings[first].StartedAt < recordings[second].StartedAt
	})
}

// finish writes the final playlist and files the capture as a scene. A capture
// without enough footage is deleted instead.
func (r *Recorder) finish(ctx context.Context, recording *capture, reason string) {
	duration, size := recording.totals()
	recording.mutex.Lock()
	segmentCount := len(recording.video.segments)
	recording.mutex.Unlock()

	if segmentCount == 0 || duration < minimumRecordingSeconds {
		r.discard(ctx, recording, reasonTooShort)
		return
	}
	sceneID, err := r.publish(ctx, recording)
	if err != nil {
		log.Printf("cams: finish recording of %s: %v", recording.camera.Username, err)
		r.updateRow(ctx, recording, statusFailed, reasonError, "", duration, size)
		return
	}
	r.updateRow(ctx, recording, statusFinished, reason, sceneID, duration, size)
}

// discard deletes a capture's files and marks it failed.
func (r *Recorder) discard(ctx context.Context, recording *capture, reason string) {
	files, err := r.blobs.List(ctx, recording.blobDirectory())
	if err == nil {
		for _, file := range files {
			_ = r.blobs.Delete(ctx, file.Key)
		}
	}
	r.updateRow(ctx, recording, statusFailed, reason, "", 0, 0)
}

// updateRow stores how a capture ended.
func (r *Recorder) updateRow(ctx context.Context, recording *capture, status, reason, sceneID string, duration float64, size int64) {
	var scene any
	if sceneID != "" {
		scene = sceneID
	}
	_, err := r.service.database.Exec(ctx,
		`UPDATE cam_recording SET status = $status, stop_reason = $reason, scene = $scene, ended_at = $endedAt,
		 duration_seconds = $duration, size_bytes = $size, quality = COALESCE($quality, quality) WHERE id = $id`,
		db.Vars{"id": recording.recordID, "status": status, "reason": reason, "scene": scene,
			"endedAt": formatTime(time.Now()), "duration": duration, "size": size, "quality": derefOrNil(recording.quality())})
	if err != nil {
		log.Printf("cams: update recording %s: %v", recording.recordID, err)
	}
}

// publish stores the final playlist and poster and creates the scene.
func (r *Recorder) publish(ctx context.Context, recording *capture) (string, error) {
	if err := r.writePlaylist(ctx, recording, true); err != nil {
		return "", fmt.Errorf("write playlist: %w", err)
	}
	posterURL := r.storePoster(ctx, recording)
	return r.createScene(ctx, recording, posterURL)
}

// flushPlaylist writes the playlists so far, so a crash keeps the footage.
func (r *Recorder) flushPlaylist(recording *capture) {
	if err := r.writePlaylist(recording.context, recording, false); err != nil {
		log.Printf("cams: record %s: write playlist: %v", recording.camera.Username, err)
		return
	}
	recording.mutex.Lock()
	recording.video.flushedSegments = len(recording.video.segments)
	if recording.audio != nil {
		recording.audio.flushedSegments = len(recording.audio.segments)
	}
	recording.mutex.Unlock()
}

// writePlaylist stores the recording's VOD playlists; ended closes them. A
// stream with a separate audio rendition gets one playlist per track and a
// master (index.m3u8) joining them; otherwise index.m3u8 is the one playlist.
func (r *Recorder) writePlaylist(ctx context.Context, recording *capture, ended bool) error {
	recording.mutex.Lock()
	videoText := renderPlaylist(recording.video.segments, ended)
	audioText := ""
	if recording.audio != nil && len(recording.audio.segments) > 0 {
		audioText = renderPlaylist(recording.audio.segments, ended)
	}
	recording.mutex.Unlock()

	directory := recording.blobDirectory()
	if audioText == "" {
		return r.putText(ctx, directory+playlistName, videoText)
	}
	if err := r.putText(ctx, directory+recording.video.playlistName, videoText); err != nil {
		return err
	}
	if err := r.putText(ctx, directory+recording.audio.playlistName, audioText); err != nil {
		return err
	}
	return r.putText(ctx, directory+playlistName, renderMaster(recording.video.playlistName, recording.audio.playlistName))
}

// putText stores a text blob.
func (r *Recorder) putText(ctx context.Context, key, text string) error {
	return r.blobs.Put(ctx, key, strings.NewReader(text), int64(len(text)), "application/x-mpegURL")
}

// renderMaster builds the master playlist joining a video and an audio playlist.
func renderMaster(videoPlaylist, audioPlaylist string) string {
	return "#EXTM3U\n#EXT-X-VERSION:7\n" +
		fmt.Sprintf("#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"audio\",NAME=\"Audio\",DEFAULT=YES,AUTOSELECT=YES,URI=%q\n", audioPlaylist) +
		"#EXT-X-STREAM-INF:BANDWIDTH=5000000,AUDIO=\"audio\"\n" + videoPlaylist + "\n"
}

// renderPlaylist builds an HLS playlist over stored segments with relative URIs.
func renderPlaylist(segments []storedSegment, ended bool) string {
	longest := 1.0
	for _, segment := range segments {
		longest = math.Max(longest, segment.duration)
	}
	var builder strings.Builder
	builder.WriteString("#EXTM3U\n#EXT-X-VERSION:7\n")
	fmt.Fprintf(&builder, "#EXT-X-TARGETDURATION:%d\n#EXT-X-MEDIA-SEQUENCE:0\n#EXT-X-PLAYLIST-TYPE:VOD\n", int(math.Ceil(longest)))
	currentInit := ""
	for _, segment := range segments {
		if segment.discontinuity {
			builder.WriteString("#EXT-X-DISCONTINUITY\n")
			currentInit = ""
		}
		if segment.initName != currentInit {
			currentInit = segment.initName
			if currentInit != "" {
				fmt.Fprintf(&builder, "#EXT-X-MAP:URI=%q\n", currentInit)
			}
		}
		fmt.Fprintf(&builder, "#EXTINF:%.3f,\n%s\n", segment.duration, segment.name)
	}
	if ended {
		builder.WriteString("#EXT-X-ENDLIST\n")
	}
	return builder.String()
}

// storePoster keeps the room's latest thumbnail with the recording and returns
// its URL, or "" when there is none to fetch.
func (r *Recorder) storePoster(ctx context.Context, recording *capture) string {
	camera := recording.camera
	if latest, err := r.service.load(ctx, camera.ID); err == nil {
		camera = *latest
	}
	if camera.ThumbnailURL == nil || *camera.ThumbnailURL == "" {
		return ""
	}
	data, err := r.fetchBytes(ctx, *camera.ThumbnailURL, nil)
	if err != nil || len(data) == 0 {
		return ""
	}
	key := recording.blobDirectory() + posterName
	if err := r.blobs.Put(ctx, key, bytes.NewReader(data), int64(len(data)), "image/jpeg"); err != nil {
		return ""
	}
	return r.blobs.DirectURL(key)
}

// createScene files the recording as a scene credited to the model's
// performer, with the recording as its one stream. It returns the scene id.
func (r *Recorder) createScene(ctx context.Context, recording *capture, posterURL string) (string, error) {
	duration, size := recording.totals()

	camera := recording.camera
	sourceURL := ingest.RecordingSourcePrefix + recording.recordID
	scene := &plugins.Scene{
		Type:       plugins.MediaTypeScene,
		ExternalID: recording.recordID,
		SourceURL:  sourceURL,
		Title:      fmt.Sprintf("%s — live %s", camera.DisplayName, recording.startedAt.Local().Format("2006-01-02 15:04")),
		Date:       recording.startedAt.UTC().Format(time.RFC3339),
		Duration:   int(math.Round(duration)),
		Studio:     r.siteStudio(camera),
		Performers: []plugins.ScenePerformer{{Name: camera.DisplayName}},
		Tags:       recordingTags(camera),
		PosterPath: posterURL,
		Downloads: []plugins.Download{{
			Label: "Recording", URL: r.blobs.DirectURL(recording.blobDirectory() + playlistName),
			Format: "hls", SizeBytes: size,
		}},
	}
	result, err := r.scenes.IngestScrapeResults(ctx, camera.Plugin, sourceURL,
		[]*plugins.ScrapeResult{{Type: plugins.MediaTypeScene, Scene: scene}})
	if err != nil {
		return "", err
	}
	if err := r.scenes.MarkDetailFetched(ctx, result.CanonicalID); err != nil {
		return "", err
	}
	return result.CanonicalID, nil
}

// siteStudio names the cam site as the scene's studio.
func (r *Recorder) siteStudio(camera camRecord) *plugins.StudioRef {
	name := camera.Plugin
	if plugin, ok := r.service.sites.Plugin(camera.Plugin); ok && plugin.Meta.DisplayName != "" {
		name = plugin.Meta.DisplayName
	}
	studio := &plugins.StudioRef{Name: name}
	if parsed, err := url.Parse(camera.SourceURL); err == nil && parsed.Host != "" {
		studio.SourceURL = parsed.Scheme + "://" + parsed.Host
	}
	return studio
}

// recordingTags are the tags of a recorded scene: "Live cam" and some of the
// tags the room carried.
func recordingTags(camera camRecord) []string {
	tags := []string{"Live cam"}
	for _, tag := range camera.Tags {
		if len(tags) > recordingTagLimit {
			break
		}
		tags = append(tags, tag)
	}
	return tags
}

// Recover files the recordings a crash or restart interrupted, using the
// playlists they had flushed; ones with nothing stored are marked failed.
func (r *Recorder) Recover(ctx context.Context) {
	rows, err := db.QueryAs[recordingRow](ctx, r.service.database, recordingSelect+` WHERE r.status = 'recording'`, nil)
	if err != nil {
		log.Printf("cams: recover recordings: %v", err)
		return
	}
	for _, row := range rows {
		r.recoverOne(ctx, row)
	}
}

// recoverOne finishes one interrupted recording.
func (r *Recorder) recoverOne(ctx context.Context, row recordingRow) {
	camera, err := r.service.load(ctx, row.CamModel)
	if err != nil {
		return
	}
	startedAt, _ := time.Parse(time.RFC3339, row.StartedAt)
	recording := &capture{recordID: row.ID, camera: *camera, startedAt: startedAt, context: ctx, video: newTrack("seg", "init", "video.m3u8")}
	if err := r.restoreSegments(ctx, recording); err != nil {
		log.Printf("cams: recover %s: %v", row.ID, err)
	}
	r.finish(ctx, recording, reasonRestart)
}

// restoreSegments reads the flushed playlists back into the capture.
func (r *Recorder) restoreSegments(ctx context.Context, recording *capture) error {
	directory := recording.blobDirectory()
	indexText, err := r.readText(ctx, directory+playlistName)
	if err != nil {
		return err
	}
	sizes := map[string]int64{}
	files, _ := r.blobs.List(ctx, directory)
	for _, file := range files {
		sizes[strings.TrimPrefix(file.Key, directory)] = file.Size
	}
	recording.video = newTrack("seg", "init", "video.m3u8")
	if !strings.Contains(indexText, "#EXT-X-STREAM-INF") {
		return restoreTrack(recording.video, indexText, sizes)
	}
	videoText, err := r.readText(ctx, directory+recording.video.playlistName)
	if err != nil {
		return err
	}
	if err := restoreTrack(recording.video, videoText, sizes); err != nil {
		return err
	}
	audioText, err := r.readText(ctx, directory+"audio.m3u8")
	if err != nil {
		return nil
	}
	recording.audio = newTrack("aseg", "ainit", "audio.m3u8")
	return restoreTrack(recording.audio, audioText, sizes)
}

// readText reads a stored text blob.
func (r *Recorder) readText(ctx context.Context, key string) (string, error) {
	reader, err := r.blobs.Get(ctx, key)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	return string(data), err
}

// restoreTrack fills a track from the text of its flushed playlist.
func restoreTrack(restored *track, playlistText string, sizes map[string]int64) error {
	base, _ := url.Parse("http://recording.invalid/")
	playlist, err := parseMediaPlaylist(playlistText, base)
	if err != nil {
		return err
	}
	for _, segment := range playlist.Segments {
		name := strings.TrimPrefix(segment.Identity, "/")
		initName := ""
		if segment.InitURL != "" {
			initName = strings.TrimPrefix(urlIdentity(segment.InitURL), "/")
		}
		restored.segments = append(restored.segments, storedSegment{
			name: name, duration: segment.Duration, initName: initName, discontinuity: segment.Discontinuity,
		})
		restored.durationSeconds += segment.Duration
		restored.sizeBytes += sizes[name]
	}
	return nil
}
