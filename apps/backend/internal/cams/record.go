package cams

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/schema"
)

const (
	// DefaultMaxRecording bounds one recording when no setting says otherwise.
	DefaultMaxRecording = 4 * time.Hour
	// finishTimeout bounds turning a finished capture into a video.
	finishTimeout = 2 * time.Minute
	// qualityPayloadKey names the record job's payload field holding the quality.
	qualityPayloadKey = "quality"
)

// RecordOptions say how a recording starts.
type RecordOptions struct {
	// Quality is "source" or a target height (see ParseQuality); empty uses
	// the model's recordQuality.
	Quality string
	// FromSecondsAgo starts the recording that far in the past, from the
	// room's live buffer (all of it when less is buffered, none when the room
	// is not buffered). The buffered rendition then wins over Quality.
	FromSecondsAgo int
}

// ActiveRecording describes a recording running in this process.
type ActiveRecording struct {
	JobID           string
	CamModelID      string
	StartedAt       time.Time
	DurationSeconds float64
	SizeBytes       int64
	// Quality names the rendition being captured, like "720p"; "" until known.
	Quality string
}

// StartRecording queues a record job for the room and returns its id. A room
// already being recorded returns the running job. A room that is not live
// fails with ErrOffline.
func (s *Service) StartRecording(ctx context.Context, id string, options RecordOptions) (string, error) {
	camModel, err := s.load(ctx, id)
	if err != nil {
		return "", err
	}
	quality, err := recordQuality(options.Quality, camModel)
	if err != nil {
		return "", err
	}
	if running := s.runningJob(id); running != "" {
		return running, nil
	}
	if !s.checkOnline(ctx, camModel) {
		return "", fmt.Errorf("%w: %s", ErrOffline, camModel.DisplayName)
	}
	s.holdSeed(id, options.FromSecondsAgo)
	return s.services.Queue.Enqueue(ctx, jobs.Job{
		Kind:       schema.JobKindRecord,
		Subject:    &schema.RecordRef{Kind: schema.RecordKindCamModel, Id: id},
		PluginName: camModel.Plugin,
		Payload:    map[string]any{qualityPayloadKey: quality},
		DedupeKey:  recordDedupeKey(id),
	})
}

// recordQuality is the normalized quality a recording runs at: the requested
// one, else the model's, else source.
func recordQuality(requested string, camModel *schema.CamModel) (string, error) {
	if requested == "" && camModel.RecordQuality != nil {
		requested = *camModel.RecordQuality
	}
	return NormalizeQuality(requested)
}

// recordDedupeKey keeps one record job per room, so one room never has two captures.
func recordDedupeKey(camModelID string) string {
	return "record:" + camModelID
}

// runningJob is the id of the room's recording running in this process, or "".
func (s *Service) runningJob(camModelID string) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	recording := s.active[camModelID]
	if recording == nil {
		return ""
	}
	return recording.jobID
}

// holdSeed keeps the buffered past a recording starts from until its job
// picks it up. The buffer slides, so it is taken now, at the request.
func (s *Service) holdSeed(camModelID string, secondsAgo int) {
	if s.services.Buffer == nil || secondsAgo <= 0 {
		return
	}
	seed := s.services.Buffer.SeedFrom(camModelID, float64(secondsAgo))
	if seed == nil {
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.seeds[camModelID] = seed
}

// StopRecording ends the room's recording and waits until it is filed as a
// video. A recording waiting in the queue is cancelled; footage an earlier
// attempt left is filed. A room not being recorded is no error.
func (s *Service) StopRecording(ctx context.Context, id string) error {
	s.mutex.Lock()
	recording := s.active[id]
	s.mutex.Unlock()
	if recording == nil {
		return s.cancelQueued(ctx, id)
	}
	recording.stopRequested.Store(true)
	recording.cancel()
	select {
	case <-recording.finished:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// cancelQueued cancels the room's queued record job and files what an
// interrupted attempt of it recorded.
func (s *Service) cancelQueued(ctx context.Context, camModelID string) error {
	job, err := db.QueryOneAs[jobs.Job](ctx, s.services.Database,
		`SELECT * FROM "Job" WHERE "dedupeKey" = $key AND "status" = $queued`,
		db.Vars{"key": recordDedupeKey(camModelID), "queued": jobs.StatusQueued})
	if err != nil || job == nil {
		return err
	}
	if _, err := s.services.Queue.Cancel(ctx, job.ID); err != nil {
		return err
	}
	if _, flushed := s.flushedAt(ctx, job.ID); !flushed {
		return nil
	}
	camModel, err := s.load(ctx, camModelID)
	if err != nil {
		return err
	}
	recording := s.newCapture(ctx, job.ID, *camModel, 0, job.CreatedAt)
	defer recording.cancel()
	if err := s.restoreSegments(ctx, recording); err != nil {
		return err
	}
	_, err = s.finish(ctx, recording)
	return ignoreTooShort(err)
}

// ActiveRecordings returns the recordings running in this process, oldest first.
func (s *Service) ActiveRecordings() []ActiveRecording {
	s.mutex.Lock()
	captures := make([]*capture, 0, len(s.active))
	for _, recording := range s.active {
		captures = append(captures, recording)
	}
	s.mutex.Unlock()
	active := make([]ActiveRecording, 0, len(captures))
	for _, recording := range captures {
		duration, size := recording.totals()
		active = append(active, ActiveRecording{
			JobID: recording.jobID, CamModelID: recording.camModel.Id, StartedAt: recording.startedAt,
			DurationSeconds: duration, SizeBytes: size, Quality: recording.quality(),
		})
	}
	sort.SliceStable(active, func(first, second int) bool {
		return active[first].StartedAt.Before(active[second].StartedAt)
	})
	return active
}

// RecordTask is the task of `record` jobs: it records the room the job's
// subject names into recordings/<job id>/ until the room goes offline (after
// a grace period), the recording is stopped or cancelled, or it reaches the
// longest recording the settings allow; then it files the footage as a
// video. When the server stops mid-recording the footage stays on disk and
// the job, queued again at the next start, goes on from it.
func (s *Service) RecordTask() jobs.Task {
	return func(ctx context.Context, job *jobs.Job, progress func(float64)) error {
		if job.Subject == nil || job.Subject.Kind != schema.RecordKindCamModel {
			return errors.New("a record job needs a cam model subject")
		}
		camModel, err := s.load(ctx, job.Subject.Id)
		if err != nil {
			return err
		}
		targetHeight, err := ParseQuality(payloadString(job.Payload, qualityPayloadKey))
		if err != nil {
			return err
		}
		recording, err := s.reserve(ctx, job, *camModel, targetHeight)
		if err != nil {
			return err
		}
		defer s.release(recording)
		return s.record(ctx, job, recording, progress)
	}
}

// payloadString reads a string field of a job payload, "" when absent.
func payloadString(payload map[string]any, key string) string {
	value, _ := payload[key].(string)
	return value
}

// newCapture returns a capture of the room for the job, in a context of its own
// so a stop can end it without ending the job.
func (s *Service) newCapture(ctx context.Context, jobID string, camModel schema.CamModel, targetHeight int, startedAt time.Time) *capture {
	captureContext, cancel := context.WithCancel(ctx)
	return &capture{
		jobID: jobID, camModel: camModel, startedAt: startedAt, context: captureContext, cancel: cancel,
		finished: make(chan struct{}), targetHeight: targetHeight, video: newVideoTrack(),
	}
}

// reserve registers the job's capture as the room's running recording.
func (s *Service) reserve(ctx context.Context, job *jobs.Job, camModel schema.CamModel, targetHeight int) (*capture, error) {
	recording := s.newCapture(ctx, job.ID, camModel, targetHeight, time.Now())
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.active[camModel.Id] != nil {
		recording.cancel()
		return nil, fmt.Errorf("%s is already being recorded", camModel.DisplayName)
	}
	s.active[camModel.Id] = recording
	recording.seed = s.seeds[camModel.Id]
	delete(s.seeds, camModel.Id)
	return recording, nil
}

// release forgets a capture once it ended and wakes StopRecording.
func (s *Service) release(recording *capture) {
	s.mutex.Lock()
	if s.active[recording.camModel.Id] == recording {
		delete(s.active, recording.camModel.Id)
	}
	s.mutex.Unlock()
	recording.cancel()
	close(recording.finished)
}

// record runs one attempt of a record job: it resumes footage an earlier
// attempt left, captures, and files the result.
func (s *Service) record(ctx context.Context, job *jobs.Job, recording *capture, progress func(float64)) error {
	resumed, err := s.resume(ctx, job, recording)
	if err != nil {
		return err
	}
	if resumed && s.resumedTooLate(ctx, job.ID) {
		return s.file(ctx, recording, reasonOffline)
	}
	if !resumed {
		s.copySeed(recording)
	}
	maxDuration := s.maxDuration()
	duration, _ := recording.totals()
	deadline := time.Now().Add(maxDuration - time.Duration(duration*float64(time.Second)))
	stopProgress := reportProgress(recording, maxDuration, progress)
	reason := s.captureLoop(recording, deadline)
	stopProgress()
	if reason == reasonInterrupted && !s.jobCancelled(ctx, job.ID) {
		return s.suspend(ctx, recording)
	}
	return s.file(ctx, recording, reason)
}

// resume restores the footage an interrupted attempt of the job flushed. It
// reports whether there was any.
func (s *Service) resume(ctx context.Context, job *jobs.Job, recording *capture) (bool, error) {
	if _, flushed := s.flushedAt(ctx, job.ID); !flushed {
		return false, nil
	}
	recording.startedAt = job.CreatedAt
	if err := s.restoreSegments(ctx, recording); err != nil {
		return false, fmt.Errorf("resume recording: %w", err)
	}
	return true, nil
}

// resumedTooLate reports whether the interrupted attempt stopped writing longer
// ago than a room may be unreachable: the show it recorded is over, so a new
// one must not be appended to it.
func (s *Service) resumedTooLate(ctx context.Context, jobID string) bool {
	flushedAt, _ := s.flushedAt(ctx, jobID)
	return time.Since(flushedAt) > s.timing.offlineGrace
}

// jobCancelled reports whether the job was cancelled through the job queue,
// which ends its context like a server shutdown does.
func (s *Service) jobCancelled(ctx context.Context, jobID string) bool {
	job, err := s.services.Queue.Get(context.WithoutCancel(ctx), jobID)
	return err == nil && job != nil && job.Status == jobs.StatusCancelled
}

// suspend flushes the footage of a capture the server interrupted, so the
// job's next attempt goes on from it, and fails the attempt.
func (s *Service) suspend(ctx context.Context, recording *capture) error {
	finishContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), finishTimeout)
	defer cancel()
	if err := s.writePlaylist(finishContext, recording, false); err != nil {
		log.Printf("cams: suspend recording %s: %v", recording.jobID, err)
	}
	return ctx.Err()
}

// file turns the ended capture into a video. A capture that never got enough
// footage is an error when the room was offline (so the job retries) and
// none when the user stopped it.
func (s *Service) file(ctx context.Context, recording *capture, reason string) error {
	finishContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), finishTimeout)
	defer cancel()
	videoID, err := s.finish(finishContext, recording)
	if errors.Is(err, errTooShort) && reason == reasonOffline {
		return fmt.Errorf("%w: nothing recorded of %s", ErrOffline, recording.camModel.DisplayName)
	}
	if err != nil {
		return ignoreTooShort(err)
	}
	log.Printf("cams: recording of %s filed as video %s (%s)", recording.camModel.DisplayName, videoID, reason)
	return nil
}

// ignoreTooShort drops errTooShort: a recording too short to keep is no failure.
func ignoreTooShort(err error) error {
	if errors.Is(err, errTooShort) {
		return nil
	}
	return err
}

// maxDuration is the longest a recording may run, from the settings.
func (s *Service) maxDuration() time.Duration {
	if s.timing.maxRecording > 0 {
		return s.timing.maxRecording
	}
	if s.services.Settings == nil {
		return DefaultMaxRecording
	}
	minutes := s.services.Settings.Cached().CamRecordingMaxMinutes
	if minutes <= 0 {
		return DefaultMaxRecording
	}
	return time.Duration(minutes) * time.Minute
}

// reportProgress reports the recorded share of the longest recording every
// few seconds until the returned stop is called.
func reportProgress(recording *capture, maxDuration time.Duration, progress func(float64)) func() {
	done := make(chan struct{})
	go progressLoop(recording, maxDuration, progress, done)
	return func() { close(done) }
}

// progressLoop is reportProgress's goroutine; it ends when done closes.
func progressLoop(recording *capture, maxDuration time.Duration, progress func(float64), done chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
		}
		duration, _ := recording.totals()
		progress(duration / maxDuration.Seconds())
	}
}
