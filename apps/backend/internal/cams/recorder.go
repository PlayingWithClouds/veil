package cams

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/storage"
)

const (
	// RecordingPrefix is the blob key prefix recordings are stored under. It is
	// outside every cache prefix, so the cache evictor never touches them.
	RecordingPrefix = storage.RecordingPrefix

	// DefaultMaxRecording bounds one recording when no setting says otherwise.
	DefaultMaxRecording = 4 * time.Hour
	// minimumRecordingSeconds is the shortest recording kept as a scene.
	minimumRecordingSeconds = 5.0
	// playlistFlushSegments is how many new segments pass between writes of the
	// playlist, so a crash loses at most that much footage.
	playlistFlushSegments = 10
	// maxSegmentBytes bounds one downloaded segment.
	maxSegmentBytes = 256 << 20
	// finishTimeout bounds turning a finished capture into a scene.
	finishTimeout = 2 * time.Minute

	statusRecording = "recording"
	statusFinished  = "finished"
	statusFailed    = "failed"

	reasonStopped     = "stopped"
	reasonOffline     = "offline"
	reasonMaxDuration = "max_duration"
	reasonRestart     = "restart"
	reasonTooShort    = "too_short"
	reasonError       = "error"
)

// ErrOffline is returned when a room cannot be recorded because it is not live.
var ErrOffline = fmt.Errorf("room is offline")

// BlobStore is the part of the blob store the recorder writes recordings to.
type BlobStore interface {
	Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]storage.File, error)
	DirectURL(key string) string
}

// Fetcher downloads a URL with the headers a stream needs. A response with a
// status other than 200 is an error.
type Fetcher interface {
	Fetch(ctx context.Context, rawURL string, headers map[string]string) (*http.Response, error)
}

// SceneWriter turns a finished recording into a library scene.
type SceneWriter interface {
	IngestScrapeResults(ctx context.Context, pluginName, sourceURL string, results []*plugins.ScrapeResult) (*ingest.IngestResult, error)
	MarkDetailFetched(ctx context.Context, sceneID string) error
}

// Recorder captures live rooms to the blob store and files each capture as a
// scene. It pulls the room's HLS stream itself, so a recording continues after
// the page that started it is closed.
type Recorder struct {
	service *Service
	blobs   BlobStore
	fetcher Fetcher
	scenes  SceneWriter
	// baseContext ends with the server; captures in progress are finished then.
	baseContext context.Context

	// maxDuration returns the longest a recording may run.
	maxDuration func() time.Duration
	// offlineGrace is how long a room may be unreachable before the recording ends.
	offlineGrace time.Duration
	// retryDelay is the wait before trying to resolve the stream again.
	retryDelay time.Duration
	// pollMinimum and pollMaximum clamp the wait between playlist refreshes.
	pollMinimum time.Duration
	pollMaximum time.Duration

	mutex  sync.Mutex
	active map[string]*capture
}

// NewRecorder returns a Recorder; ctx is the server's lifetime.
func NewRecorder(ctx context.Context, service *Service, blobs BlobStore, fetcher Fetcher, scenes SceneWriter, maxDuration func() time.Duration) *Recorder {
	return &Recorder{
		service: service, blobs: blobs, fetcher: fetcher, scenes: scenes, baseContext: ctx,
		maxDuration:  maxDuration,
		offlineGrace: 90 * time.Second,
		retryDelay:   3 * time.Second,
		pollMinimum:  time.Second,
		pollMaximum:  4 * time.Second,
		active:       map[string]*capture{},
	}
}

// storedSegment is a segment written to the blob store.
type storedSegment struct {
	name          string
	duration      float64
	initName      string
	discontinuity bool
}

// track is one stream of a capture: the video, or a separate audio rendition.
type track struct {
	// segmentPrefix and initPrefix start the names of the track's blob files.
	segmentPrefix string
	initPrefix    string
	// playlistName is the file holding the track's playlist when the capture has
	// two tracks.
	playlistName string

	segments        []storedSegment
	durationSeconds float64
	sizeBytes       int64
	initNames       map[string]string
	seen            map[string]bool
	pendingBreak    bool
	flushedSegments int
}

// newTrack returns an empty track writing files named with the prefixes.
func newTrack(segmentPrefix, initPrefix, playlistName string) *track {
	return &track{
		segmentPrefix: segmentPrefix, initPrefix: initPrefix, playlistName: playlistName,
		initNames: map[string]string{}, seen: map[string]bool{},
	}
}

// capture is one recording in progress.
type capture struct {
	recordID  string
	camera    camRecord
	startedAt time.Time
	context   context.Context
	cancel    context.CancelFunc
	finished  chan struct{}
	// stopRequested separates a stop by the user from the server shutting down.
	stopRequested atomic.Bool
	// targetHeight is the rendition asked for (0 = source), picked again on every
	// reconnect because a stream's renditions can change.
	targetHeight int

	// mutex guards the tracks, their counters and qualityLabel.
	mutex sync.Mutex
	// qualityLabel names the rendition being captured, like "480p".
	qualityLabel string
	video        *track
	// audio is the separate audio rendition, nil for streams with muxed audio.
	audio *track
}

// blobDirectory is the blob key prefix holding the capture's files.
func (recording *capture) blobDirectory() string {
	_, identifier, _ := strings.Cut(recording.recordID, ":")
	return RecordingPrefix + identifier + "/"
}

// quality returns the label of the rendition being captured, nil before the first is known.
func (recording *capture) quality() *string {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	if recording.qualityLabel == "" {
		return nil
	}
	label := recording.qualityLabel
	return &label
}

// totals returns the footage length (the video's) and the bytes stored in all tracks.
func (recording *capture) totals() (durationSeconds float64, sizeBytes int64) {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	durationSeconds = recording.video.durationSeconds
	sizeBytes = recording.video.sizeBytes
	if recording.audio != nil {
		sizeBytes += recording.audio.sizeBytes
	}
	return durationSeconds, sizeBytes
}

// snapshot describes the capture as it is now.
func (recording *capture) snapshot() *model.CamRecording {
	durationSeconds, sizeBytes := recording.totals()
	return &model.CamRecording{
		ID: recording.recordID, CamModelID: recording.camera.ID, DisplayName: recording.camera.DisplayName,
		ThumbnailURL: recording.camera.ThumbnailURL, Quality: recording.quality(), Status: statusRecording, StartedAt: formatTime(recording.startedAt),
		DurationSeconds: durationSeconds, SizeBytes: float64(sizeBytes),
	}
}

// formatTime renders a timestamp the way the database stores them.
func formatTime(moment time.Time) string {
	return moment.UTC().Format("2006-01-02T15:04:05.000Z")
}

// streamSession is a resolved, currently valid way to read the room's stream.
type streamSession struct {
	mediaURL string
	// audioURL is the separate audio rendition's playlist, "" when audio is muxed.
	audioURL string
	headers  map[string]string
}

// Start begins recording a room at a quality (see ParseQuality) and returns the
// recording. A room that is already being recorded returns the running
// recording. A room that is not live fails with ErrOffline.
func (r *Recorder) Start(ctx context.Context, cameraID, quality string) (*model.CamRecording, error) {
	targetHeight, err := ParseQuality(quality)
	if err != nil {
		return nil, err
	}
	camera, err := r.service.load(ctx, cameraID)
	if err != nil {
		return nil, err
	}
	recording, existing := r.reserve(camera, targetHeight)
	if existing {
		return recording.snapshot(), nil
	}
	session, err := r.openSession(ctx, recording)
	if err != nil {
		r.release(cameraID)
		return nil, fmt.Errorf("%w: %v", ErrOffline, err)
	}
	if err := r.insertRow(r.baseContext, recording); err != nil {
		r.release(cameraID)
		return nil, err
	}
	r.service.recordOnline(ctx, cameraID, true)
	go r.run(recording, session)
	return recording.snapshot(), nil
}

// reserve registers a capture for the room unless one is running, so two
// recordings of one room can never overlap. existing reports the latter.
func (r *Recorder) reserve(camera *camRecord, targetHeight int) (recording *capture, existing bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if running := r.active[camera.ID]; running != nil {
		return running, true
	}
	captureContext, cancel := context.WithCancel(r.baseContext)
	recording = &capture{
		recordID: db.NewRecordID("cam_recording").String(), camera: *camera, startedAt: time.Now(),
		context: captureContext, cancel: cancel, finished: make(chan struct{}),
		video: newTrack("seg", "init", "video.m3u8"), targetHeight: targetHeight,
	}
	r.active[camera.ID] = recording
	return recording, false
}

// release forgets a capture that never started.
func (r *Recorder) release(cameraID string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if recording := r.active[cameraID]; recording != nil {
		recording.cancel()
		delete(r.active, cameraID)
	}
}

// Stop ends the recording of a room, waits until its scene exists and returns
// the finished recording, or nil when the room is not being recorded.
func (r *Recorder) Stop(ctx context.Context, cameraID string) (*model.CamRecording, error) {
	r.mutex.Lock()
	recording := r.active[cameraID]
	r.mutex.Unlock()
	if recording == nil {
		return nil, nil
	}
	recording.stopRequested.Store(true)
	recording.cancel()
	select {
	case <-recording.finished:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return r.loadRecording(ctx, recording.recordID)
}

// IsRecording reports whether a room is being recorded.
func (r *Recorder) IsRecording(cameraID string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.active[cameraID] != nil
}

// Active returns the recordings in progress, oldest first.
func (r *Recorder) Active() []*model.CamRecording {
	r.mutex.Lock()
	captures := make([]*capture, 0, len(r.active))
	for _, recording := range r.active {
		captures = append(captures, recording)
	}
	r.mutex.Unlock()
	snapshots := make([]*model.CamRecording, 0, len(captures))
	for _, recording := range captures {
		snapshots = append(snapshots, recording.snapshot())
	}
	sortRecordings(snapshots)
	return snapshots
}

// run captures the room until it ends, then files the result.
func (r *Recorder) run(recording *capture, session *streamSession) {
	defer close(recording.finished)
	reason := r.captureLoop(recording, session)
	finishContext, cancel := context.WithTimeout(context.WithoutCancel(r.baseContext), finishTimeout)
	defer cancel()
	r.finish(finishContext, recording, reason)
	r.release(recording.camera.ID)
}

// captureLoop keeps reading the room's stream, reconnecting with a freshly
// resolved URL whenever it breaks, and returns why the recording ended.
func (r *Recorder) captureLoop(recording *capture, session *streamSession) string {
	deadline := recording.startedAt.Add(r.maxDuration())
	lastActivity := time.Now()
	for {
		if session == nil {
			reason, reopened := r.reopen(recording, deadline, lastActivity)
			if reason != "" {
				return reason
			}
			session = reopened
			recording.markBreak(recording.video)
			recording.markBreak(recording.audio)
		}
		reason := r.followPlaylist(recording, session, deadline, &lastActivity)
		if reason != "" {
			return reason
		}
		session = nil
	}
}

// reopen resolves a fresh stream, retrying until it works or a reason to end
// the recording applies.
func (r *Recorder) reopen(recording *capture, deadline, lastActivity time.Time) (reason string, session *streamSession) {
	for {
		if reason := r.endReason(recording, deadline); reason != "" {
			return reason, nil
		}
		if time.Since(lastActivity) > r.offlineGrace {
			return reasonOffline, nil
		}
		session, err := r.openSession(recording.context, recording)
		if err == nil {
			return "", session
		}
		sleepContext(recording.context, r.retryDelay)
	}
}

// endReason returns why the recording must end now, or "".
func (r *Recorder) endReason(recording *capture, deadline time.Time) string {
	if recording.context.Err() != nil {
		if recording.stopRequested.Load() {
			return reasonStopped
		}
		return reasonRestart
	}
	if !time.Now().Before(deadline) {
		return reasonMaxDuration
	}
	return ""
}

// followPlaylist polls the media playlist and stores new segments. It returns
// the reason the recording ends, or "" when the stream broke and a fresh URL
// is needed.
func (r *Recorder) followPlaylist(recording *capture, session *streamSession, deadline time.Time, lastActivity *time.Time) string {
	failures := 0
	for {
		if reason := r.endReason(recording, deadline); reason != "" {
			return reason
		}
		playlist, err := r.fetchPlaylist(recording.context, session.mediaURL, session.headers)
		if err != nil {
			failures++
			if failures >= 2 || time.Since(*lastActivity) > r.offlineGrace {
				return ""
			}
			sleepContext(recording.context, r.pollMinimum)
			continue
		}
		failures = 0
		*lastActivity = time.Now()
		r.storeNewSegments(recording, session, recording.video, playlist)
		r.followAudio(recording, session)
		if playlist.Ended {
			return reasonOffline
		}
		sleepContext(recording.context, r.pollInterval(playlist.TargetDuration))
	}
}

// followAudio stores the new segments of the separate audio rendition, if the
// stream has one. Its playlist failing is not the stream failing: the video
// decides that, the audio just misses a stretch.
func (r *Recorder) followAudio(recording *capture, session *streamSession) {
	if session.audioURL == "" {
		return
	}
	playlist, err := r.fetchPlaylist(recording.context, session.audioURL, session.headers)
	if err != nil {
		return
	}
	recording.mutex.Lock()
	if recording.audio == nil {
		recording.audio = newTrack("aseg", "ainit", "audio.m3u8")
	}
	audio := recording.audio
	recording.mutex.Unlock()
	r.storeNewSegments(recording, session, audio, playlist)
}

// pollInterval is half the target segment length, within the poll bounds.
func (r *Recorder) pollInterval(targetDuration float64) time.Duration {
	interval := time.Duration(targetDuration * float64(time.Second) / 2)
	if interval < r.pollMinimum {
		return r.pollMinimum
	}
	if interval > r.pollMaximum {
		return r.pollMaximum
	}
	return interval
}

// openSession resolves a fresh stream URL for the room and picks the rendition
// to record from its master playlist, so a reconnect also notices renditions
// that changed.
func (r *Recorder) openSession(ctx context.Context, recording *capture) (*streamSession, error) {
	camera := &recording.camera
	plugin, ok := r.service.sites.Plugin(camera.Plugin)
	if !ok {
		return nil, fmt.Errorf("cam plugin %q is not available", camera.Plugin)
	}
	result, err := r.service.sites.Stream(ctx, plugin, camera.Username)
	if err != nil {
		return nil, err
	}
	session := &streamSession{mediaURL: result.URL, headers: result.Headers}
	text, err := r.fetchText(ctx, session.mediaURL, session.headers)
	if err != nil {
		return nil, err
	}
	base, err := url.Parse(result.URL)
	if err != nil {
		return nil, err
	}
	variants := parseVariants(text, base)
	if chosen := chooseVariant(variants, recording.targetHeight); chosen != nil {
		session.mediaURL = chosen.URL
		session.audioURL = chosen.AudioURL
		recording.setQualityLabel(chosen.Label())
	}
	return session, nil
}

// setQualityLabel notes which rendition is being captured.
func (recording *capture) setQualityLabel(label string) {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	recording.qualityLabel = label
}

// fetchPlaylist downloads and parses a media playlist.
func (r *Recorder) fetchPlaylist(ctx context.Context, playlistURL string, headers map[string]string) (*mediaPlaylist, error) {
	text, err := r.fetchText(ctx, playlistURL, headers)
	if err != nil {
		return nil, err
	}
	base, err := url.Parse(playlistURL)
	if err != nil {
		return nil, err
	}
	return parseMediaPlaylist(text, base)
}

// fetchText downloads a small text resource.
func (r *Recorder) fetchText(ctx context.Context, rawURL string, headers map[string]string) (string, error) {
	data, err := r.fetchBytes(ctx, rawURL, headers)
	return string(data), err
}

// fetchBytes downloads a resource into memory.
func (r *Recorder) fetchBytes(ctx context.Context, rawURL string, headers map[string]string) ([]byte, error) {
	response, err := r.fetcher.Fetch(ctx, rawURL, headers)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(io.LimitReader(response.Body, maxSegmentBytes))
}

// storeNewSegments stores every segment of the playlist the track does not hold yet.
func (r *Recorder) storeNewSegments(recording *capture, session *streamSession, stored *track, playlist *mediaPlaylist) {
	for _, segment := range playlist.Segments {
		if recording.hasSeen(stored, segment.Identity) {
			continue
		}
		if segment.Duration <= 0 {
			segment.Duration = playlist.TargetDuration
		}
		if err := r.storeSegment(recording, session, stored, segment); err != nil {
			if recording.context.Err() != nil {
				return
			}
			log.Printf("cams: record %s: segment %s: %v", recording.camera.Username, segment.Identity, err)
			recording.markBreak(stored)
		}
	}
	recording.mutex.Lock()
	due := len(stored.segments)-stored.flushedSegments >= playlistFlushSegments
	recording.mutex.Unlock()
	if due {
		r.flushPlaylist(recording)
	}
}

// hasSeen reports whether the track handled the segment before, and remembers it.
func (recording *capture) hasSeen(stored *track, identity string) bool {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	if stored.seen[identity] {
		return true
	}
	stored.seen[identity] = true
	return false
}

// markBreak notes that the next segment stored in the track follows a gap.
func (recording *capture) markBreak(stored *track) {
	if stored == nil {
		return
	}
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	if len(stored.segments) > 0 {
		stored.pendingBreak = true
	}
}

// storeSegment downloads one segment (and its init section when new) into the
// recording's blob directory.
func (r *Recorder) storeSegment(recording *capture, session *streamSession, stored *track, segment playlistSegment) error {
	initName, err := r.storeInitSection(recording, session, stored, segment.InitURL)
	if err != nil {
		return err
	}
	data, err := r.fetchBytes(recording.context, segment.URL, session.headers)
	if err != nil {
		return err
	}
	recording.mutex.Lock()
	name := fmt.Sprintf("%s-%05d%s", stored.segmentPrefix, len(stored.segments), segmentExtension(segment.Identity))
	recording.mutex.Unlock()
	if err := r.blobs.Put(recording.context, recording.blobDirectory()+name, bytes.NewReader(data), int64(len(data)), "video/mp2t"); err != nil {
		return err
	}
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	stored.segments = append(stored.segments, storedSegment{
		name: name, duration: segment.Duration, initName: initName,
		discontinuity: segment.Discontinuity || stored.pendingBreak,
	})
	stored.pendingBreak = false
	stored.durationSeconds += segment.Duration
	stored.sizeBytes += int64(len(data))
	return nil
}

// storeInitSection stores an fMP4 initialization section once and returns its
// file name; it returns "" for streams without one.
func (r *Recorder) storeInitSection(recording *capture, session *streamSession, stored *track, initURL string) (string, error) {
	if initURL == "" {
		return "", nil
	}
	identity := urlIdentity(initURL)
	recording.mutex.Lock()
	name, known := stored.initNames[identity]
	recording.mutex.Unlock()
	if known {
		return name, nil
	}
	data, err := r.fetchBytes(recording.context, initURL, session.headers)
	if err != nil {
		return "", err
	}
	recording.mutex.Lock()
	name = fmt.Sprintf("%s-%d.mp4", stored.initPrefix, len(stored.initNames))
	recording.mutex.Unlock()
	if err := r.blobs.Put(recording.context, recording.blobDirectory()+name, bytes.NewReader(data), int64(len(data)), "video/mp4"); err != nil {
		return "", err
	}
	recording.mutex.Lock()
	stored.initNames[identity] = name
	recording.mutex.Unlock()
	return name, nil
}

// segmentExtension is the file extension to keep for a segment, defaulting to MPEG-TS.
func segmentExtension(identity string) string {
	extension := strings.ToLower(path.Ext(identity))
	switch extension {
	case ".m4s", ".mp4", ".aac", ".ts":
		return extension
	}
	return ".ts"
}

// sleepContext waits for duration or until ctx ends.
func sleepContext(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}
