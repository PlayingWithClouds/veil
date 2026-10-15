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
}

// toModel maps a stored recording to the GraphQL model.
func (row recordingRow) toModel() *model.CamRecording {
	return &model.CamRecording{
		ID: row.ID, CamModelID: row.CamModel, DisplayName: row.DisplayName, Status: row.Status,
		StopReason: row.StopReason, StartedAt: row.StartedAt, EndedAt: row.EndedAt,
		DurationSeconds: row.DurationSeconds, SizeBytes: float64(row.SizeBytes), SceneID: row.Scene,
	}
}

const recordingSelect = `SELECT r.*, m.display_name FROM cam_recording r JOIN cam_model m ON m.id = r.cam_model`

// insertRow stores the row of a capture that has just started.
func (r *Recorder) insertRow(ctx context.Context, recording *capture) error {
	_, err := r.service.database.Exec(ctx,
		`INSERT INTO cam_recording (id, cam_model, status, started_at) VALUES ($id, $model, $status, $startedAt)`,
		db.Vars{"id": recording.recordID, "model": recording.camera.ID, "status": statusRecording, "startedAt": formatTime(recording.startedAt)})
	return err
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
	recording.mutex.Lock()
	duration := recording.durationSeconds
	size := recording.sizeBytes
	segmentCount := len(recording.segments)
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
		 duration_seconds = $duration, size_bytes = $size WHERE id = $id`,
		db.Vars{"id": recording.recordID, "status": status, "reason": reason, "scene": scene,
			"endedAt": formatTime(time.Now()), "duration": duration, "size": size})
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

// flushPlaylist writes the playlist so far, so a crash keeps the footage.
func (r *Recorder) flushPlaylist(recording *capture) {
	if err := r.writePlaylist(recording.context, recording, false); err != nil {
		log.Printf("cams: record %s: write playlist: %v", recording.camera.Username, err)
		return
	}
	recording.mutex.Lock()
	recording.flushedSegments = len(recording.segments)
	recording.mutex.Unlock()
}

// writePlaylist stores the recording's VOD playlist; ended closes it.
func (r *Recorder) writePlaylist(ctx context.Context, recording *capture, ended bool) error {
	recording.mutex.Lock()
	text := renderPlaylist(recording.segments, ended)
	recording.mutex.Unlock()
	return r.blobs.Put(ctx, recording.blobDirectory()+playlistName, strings.NewReader(text), int64(len(text)), "application/x-mpegURL")
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
	recording.mutex.Lock()
	duration := recording.durationSeconds
	size := recording.sizeBytes
	recording.mutex.Unlock()

	camera := recording.camera
	sourceURL := ingest.RecordingSourcePrefix + recording.recordID
	scene := &plugins.Scene{
		Type:       plugins.MediaTypeScene,
		ExternalID: recording.recordID,
		SourceURL:  sourceURL,
		Title:      fmt.Sprintf("%s — live %s", camera.DisplayName, recording.startedAt.Local().Format("2006-01-02 15:04")),
		Date:       recording.startedAt.UTC().Format("2006-01-02"),
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
	recording := &capture{recordID: row.ID, camera: *camera, startedAt: startedAt, context: ctx}
	if err := r.restoreSegments(ctx, recording); err != nil {
		log.Printf("cams: recover %s: %v", row.ID, err)
	}
	r.finish(ctx, recording, reasonRestart)
}

// restoreSegments reads the flushed playlist back into the capture.
func (r *Recorder) restoreSegments(ctx context.Context, recording *capture) error {
	reader, err := r.blobs.Get(ctx, recording.blobDirectory()+playlistName)
	if err != nil {
		return err
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	base, _ := url.Parse("http://recording.invalid/")
	playlist, err := parseMediaPlaylist(string(data), base)
	if err != nil {
		return err
	}
	sizes := map[string]int64{}
	files, _ := r.blobs.List(ctx, recording.blobDirectory())
	for _, file := range files {
		sizes[strings.TrimPrefix(file.Key, recording.blobDirectory())] = file.Size
	}
	for _, segment := range playlist.Segments {
		name := strings.TrimPrefix(segment.Identity, "/")
		initName := strings.TrimPrefix(urlIdentity(segment.InitURL), "/")
		if segment.InitURL == "" {
			initName = ""
		}
		recording.segments = append(recording.segments, storedSegment{
			name: name, duration: segment.Duration, initName: initName, discontinuity: segment.Discontinuity,
		})
		recording.durationSeconds += segment.Duration
		recording.sizeBytes += sizes[name]
	}
	return nil
}
