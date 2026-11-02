package cams

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/playingwithclouds/veil/internal/livebuffer"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/storage"
)

const (
	// playlistFlushSegments is how many new segments pass between writes of the
	// playlist, so a crash loses at most that much footage.
	playlistFlushSegments = 10
	// maxSegmentBytes bounds one downloaded segment.
	maxSegmentBytes = 256 << 20

	// Why a capture ended.
	reasonStopped     = "stopped"
	reasonOffline     = "offline"
	reasonMaxDuration = "max_duration"
	// reasonInterrupted is the job's context ending without a stop: the
	// server shutting down, or the job cancelled through the job queue.
	reasonInterrupted = "interrupted"
)

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
	// continuesAfter is the last seeded segment: when the first live playlist
	// no longer lists it, footage is missing and the next segment starts a break.
	continuesAfter  string
	flushedSegments int
}

// newVideoTrack returns the empty video track of a capture.
func newVideoTrack() *track {
	return &track{segmentPrefix: "seg", initPrefix: "init", playlistName: "video.m3u8", initNames: map[string]string{}, seen: map[string]bool{}}
}

// newAudioTrack returns the empty separate audio track of a capture.
func newAudioTrack() *track {
	return &track{segmentPrefix: "aseg", initPrefix: "ainit", playlistName: "audio.m3u8", initNames: map[string]string{}, seen: map[string]bool{}}
}

// capture is one recording in progress.
type capture struct {
	// jobID is the record job's id; the recording's files are named after it.
	jobID     string
	camModel  schema.CamModel
	startedAt time.Time
	context   context.Context
	cancel    context.CancelFunc
	finished  chan struct{}
	// stopRequested separates a stop by the user from the job's context ending.
	stopRequested atomic.Bool
	// targetHeight is the rendition asked for (0 = source), picked again on every
	// reconnect because a stream's renditions can change.
	targetHeight int
	// seed is the buffered past the recording starts from, nil for a live start.
	seed *livebuffer.Seed

	// mutex guards the tracks, their counters, qualityLabel and height.
	mutex sync.Mutex
	// qualityLabel names the rendition being captured, like "480p".
	qualityLabel string
	height       int
	video        *track
	// audio is the separate audio rendition, nil for streams with muxed audio.
	audio *track
}

// blobDirectory is the blob key prefix holding the files of the recording of job.
func blobDirectory(jobID string) string {
	return storage.RecordingPrefix + jobID + "/"
}

// directory is the blob key prefix holding the capture's files.
func (recording *capture) directory() string {
	return blobDirectory(recording.jobID)
}

// quality returns the label of the rendition being captured, "" before the first is known.
func (recording *capture) quality() string {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	return recording.qualityLabel
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

// streamSession is a resolved, currently valid way to read the room's stream.
type streamSession struct {
	mediaURL string
	// audioURL is the separate audio rendition's playlist, "" when audio is muxed.
	audioURL string
	headers  map[string]string
}

// captureLoop keeps reading the room's stream, reconnecting with a freshly
// resolved URL whenever it breaks, and returns why the recording ended.
// Footage after a reconnect follows a break; the first connection continues
// a seed seamlessly (checkContinuation decides whether it does).
func (s *Service) captureLoop(recording *capture, deadline time.Time) string {
	lastActivity := time.Now()
	for connection := 0; ; connection++ {
		session, reason := s.reopen(recording, deadline, lastActivity)
		if reason != "" {
			return reason
		}
		if connection > 0 {
			recording.markBreak(recording.video)
			recording.markBreak(recording.audio)
		}
		if reason := s.followPlaylist(recording, session, deadline, &lastActivity); reason != "" {
			return reason
		}
	}
}

// reopen resolves a fresh stream, retrying until it works or a reason to end
// the recording applies.
func (s *Service) reopen(recording *capture, deadline, lastActivity time.Time) (*streamSession, string) {
	for {
		if reason := endReason(recording, deadline); reason != "" {
			return nil, reason
		}
		if time.Since(lastActivity) > s.timing.offlineGrace {
			return nil, reasonOffline
		}
		session, err := s.openSession(recording.context, recording)
		if err == nil {
			return session, ""
		}
		sleepContext(recording.context, s.timing.retryDelay)
	}
}

// endReason returns why the recording must end now, or "".
func endReason(recording *capture, deadline time.Time) string {
	if recording.context.Err() != nil && recording.stopRequested.Load() {
		return reasonStopped
	}
	if recording.context.Err() != nil {
		return reasonInterrupted
	}
	if !time.Now().Before(deadline) {
		return reasonMaxDuration
	}
	return ""
}

// followPlaylist polls the media playlist and stores new segments. It returns
// the reason the recording ends, or "" when the stream broke and a fresh URL
// is needed.
func (s *Service) followPlaylist(recording *capture, session *streamSession, deadline time.Time, lastActivity *time.Time) string {
	failures := 0
	for {
		if reason := endReason(recording, deadline); reason != "" {
			return reason
		}
		playlist, err := s.fetchPlaylist(recording.context, session.mediaURL, session.headers)
		if err != nil {
			failures++
			if failures >= 2 || time.Since(*lastActivity) > s.timing.offlineGrace {
				return ""
			}
			sleepContext(recording.context, s.timing.pollMinimum)
			continue
		}
		failures = 0
		*lastActivity = time.Now()
		window := s.observeBuffer(recording, session.mediaURL, livebuffer.KindVideo, session, playlist)
		s.storeNewSegments(recording, session, recording.video, playlist, window)
		s.followAudio(recording, session)
		if playlist.Ended {
			return reasonOffline
		}
		sleepContext(recording.context, s.pollInterval(playlist.TargetDuration))
	}
}

// followAudio stores the new segments of the separate audio rendition, if the
// stream has one. Its playlist failing is not the stream failing: the video
// decides that, the audio just misses a stretch.
func (s *Service) followAudio(recording *capture, session *streamSession) {
	if session.audioURL == "" {
		return
	}
	playlist, err := s.fetchPlaylist(recording.context, session.audioURL, session.headers)
	if err != nil {
		return
	}
	window := s.observeBuffer(recording, session.audioURL, livebuffer.KindAudio, session, playlist)
	s.storeNewSegments(recording, session, recording.audioTrack(), playlist, window)
}

// audioTrack returns the capture's audio track, creating it on first use.
func (recording *capture) audioTrack() *track {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	if recording.audio == nil {
		recording.audio = newAudioTrack()
	}
	return recording.audio
}

// observeBuffer shows a fresh playlist to the room's live buffer, so a viewer
// can still rewind while the room is recorded, and returns the buffered window
// (nil when the room is not buffered). The recorder never creates a buffer.
func (s *Service) observeBuffer(recording *capture, playlistURL, kind string, session *streamSession, playlist *livebuffer.Playlist) *livebuffer.Window {
	if s.services.Buffer == nil {
		return nil
	}
	source := livebuffer.Source{PlaylistURL: playlistURL, Kind: kind, Fetch: s.fetcher.Fetch, Headers: session.headers}
	return s.services.Buffer.Observe(recording.camModel.Id, source, playlist, false)
}

// pollInterval is half the target segment length, within the poll bounds.
func (s *Service) pollInterval(targetDuration float64) time.Duration {
	interval := time.Duration(targetDuration * float64(time.Second) / 2)
	return min(max(interval, s.timing.pollMinimum), s.timing.pollMaximum)
}

// openSession resolves a fresh stream URL for the room and picks the rendition
// to record from its master playlist, so a reconnect also notices renditions
// that changed. Token URLs expire within about a minute, so this runs on
// every reconnect.
func (s *Service) openSession(ctx context.Context, recording *capture) (*streamSession, error) {
	camModel := &recording.camModel
	plugin, ok := s.services.Sites.Plugin(camModel.Plugin)
	if !ok {
		return nil, fmt.Errorf("cam plugin %q is not available", camModel.Plugin)
	}
	result, err := s.services.Sites.Stream(ctx, plugin, camModel.Username)
	if err != nil {
		return nil, err
	}
	session := &streamSession{mediaURL: result.URL, headers: result.Headers}
	text, err := s.fetchBytes(ctx, session.mediaURL, session.headers)
	if err != nil {
		return nil, err
	}
	base, err := url.Parse(result.URL)
	if err != nil {
		return nil, err
	}
	chosen := recording.pickVariant(parseVariants(string(text), base))
	if chosen != nil {
		session.mediaURL = chosen.URL
		session.audioURL = chosen.AudioURL
		recording.setRendition(*chosen)
	}
	return session, nil
}

// pickVariant chooses the rendition to record: the one the seed was buffered
// in when the stream still offers it, else the one the requested quality names.
func (recording *capture) pickVariant(variants []streamVariant) *streamVariant {
	if recording.seed != nil {
		for _, variant := range variants {
			if livebuffer.TrackKey(variant.URL) == recording.seed.Video.TrackKey {
				return &variant
			}
		}
	}
	return chooseVariant(variants, recording.targetHeight)
}

// setRendition notes which rendition is being captured.
func (recording *capture) setRendition(variant streamVariant) {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	recording.qualityLabel = variant.Label()
	recording.height = variant.Height
}

// fetchPlaylist downloads and parses a media playlist.
func (s *Service) fetchPlaylist(ctx context.Context, playlistURL string, headers map[string]string) (*livebuffer.Playlist, error) {
	text, err := s.fetchBytes(ctx, playlistURL, headers)
	if err != nil {
		return nil, err
	}
	base, err := url.Parse(playlistURL)
	if err != nil {
		return nil, err
	}
	return livebuffer.ParsePlaylist(string(text), base)
}

// fetchBytes downloads a resource into memory.
func (s *Service) fetchBytes(ctx context.Context, rawURL string, headers map[string]string) ([]byte, error) {
	response, err := s.fetcher.Fetch(ctx, rawURL, headers)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(io.LimitReader(response.Body, maxSegmentBytes))
}

// storeNewSegments stores every segment of the playlist the track does not hold
// yet. window, when not nil, lets segments the live buffer already downloaded
// be read from disk instead of fetched again.
func (s *Service) storeNewSegments(recording *capture, session *streamSession, stored *track, playlist *livebuffer.Playlist, window *livebuffer.Window) {
	recording.checkContinuation(stored, playlist)
	for _, segment := range playlist.Segments {
		if recording.hasSeen(stored, segment.Identity) {
			continue
		}
		if segment.Duration <= 0 {
			segment.Duration = playlist.TargetDuration
		}
		err := s.storeSegment(recording, session, stored, segment, window)
		if err != nil && recording.context.Err() != nil {
			return
		}
		if err != nil {
			log.Printf("cams: record %s: segment %s: %v", recording.camModel.Username, segment.Identity, err)
			recording.markBreak(stored)
		}
	}
	if recording.flushDue(stored) {
		s.flushPlaylist(recording)
	}
}

// flushDue reports whether enough segments were added since the last flush.
func (recording *capture) flushDue(stored *track) bool {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	return len(stored.segments)-stored.flushedSegments >= playlistFlushSegments
}

// checkContinuation marks a break when the first live playlist after a seed no
// longer lists the seed's last segment: the footage in between is lost.
func (recording *capture) checkContinuation(stored *track, playlist *livebuffer.Playlist) {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	if stored.continuesAfter == "" {
		return
	}
	listed := false
	for _, segment := range playlist.Segments {
		listed = listed || segment.Identity == stored.continuesAfter
	}
	stored.pendingBreak = stored.pendingBreak || !listed
	stored.continuesAfter = ""
}

// hasSeen reports whether the track handled the segment before, and remembers
// it. Segments are known by URL path, so a reconnect with a new token or edge
// host doesn't store them twice.
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
func (s *Service) storeSegment(recording *capture, session *streamSession, stored *track, segment livebuffer.Segment, window *livebuffer.Window) error {
	initName, err := s.storeInitSection(recording, session, stored, segment.InitURL)
	if err != nil {
		return err
	}
	data, err := s.segmentBytes(recording, session, segment, window)
	if err != nil {
		return err
	}
	return s.appendSegment(recording, stored, segment, data, initName)
}

// segmentBytes returns a segment's bytes, from the live buffer when it holds
// the segment and from the origin otherwise.
func (s *Service) segmentBytes(recording *capture, session *streamSession, segment livebuffer.Segment, window *livebuffer.Window) ([]byte, error) {
	if window == nil {
		return s.fetchBytes(recording.context, segment.URL, session.headers)
	}
	name, buffered := window.NameOf(segment.Identity)
	if !buffered {
		return s.fetchBytes(recording.context, segment.URL, session.headers)
	}
	data, err := s.services.Buffer.Read(recording.context, recording.camModel.Id, name)
	if err != nil {
		return s.fetchBytes(recording.context, segment.URL, session.headers)
	}
	return data, nil
}

// appendSegment writes a segment's bytes to the recording and adds it to the track.
func (s *Service) appendSegment(recording *capture, stored *track, segment livebuffer.Segment, data []byte, initName string) error {
	recording.mutex.Lock()
	name := fmt.Sprintf("%s-%05d%s", stored.segmentPrefix, len(stored.segments), segmentExtension(segment.Identity))
	recording.mutex.Unlock()
	if err := s.services.Blobs.Put(recording.context, recording.directory()+name, bytes.NewReader(data), int64(len(data)), "video/mp2t"); err != nil {
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
func (s *Service) storeInitSection(recording *capture, session *streamSession, stored *track, initURL string) (string, error) {
	if initURL == "" {
		return "", nil
	}
	identity := livebuffer.URLIdentity(initURL)
	if name, known := recording.initName(stored, identity); known {
		return name, nil
	}
	data, err := s.fetchBytes(recording.context, initURL, session.headers)
	if err != nil {
		return "", err
	}
	return s.saveInitSection(recording, stored, identity, data)
}

// initName returns the file an init section was stored as.
func (recording *capture) initName(stored *track, identity string) (string, bool) {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	name, known := stored.initNames[identity]
	return name, known
}

// saveInitSection writes an init section to the recording under identity.
func (s *Service) saveInitSection(recording *capture, stored *track, identity string, data []byte) (string, error) {
	recording.mutex.Lock()
	name := fmt.Sprintf("%s-%d.mp4", stored.initPrefix, len(stored.initNames))
	recording.mutex.Unlock()
	if err := s.services.Blobs.Put(recording.context, recording.directory()+name, bytes.NewReader(data), int64(len(data)), "video/mp4"); err != nil {
		return "", err
	}
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	stored.initNames[identity] = name
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
