package cams

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestRecordingStoresEachSegmentOnceAndFilesAVideo records a sliding live
// playlist and checks the stored footage, the video and its media file.
func TestRecordingStoresEachSegmentOnceAndFilesAVideo(t *testing.T) {
	f := newFixture(t, [][]int{{0, 1, 2}, {1, 2, 3}, {2, 3, 4}, {3, 4, 5}}, true)
	f.live.discontinuityAt[3] = true
	performerID := insertPerformer(t, f, "Nova Star")
	if _, err := f.service.LinkPerformer(f.ctx, f.roomID, performerID); err != nil {
		t.Fatal(err)
	}

	jobID := f.start(t, RecordOptions{})
	waitFor(t, "six segments", func() bool { return f.recordedSeconds() >= 12 })
	f.stop(t)

	if job := f.waitJob(t, jobID); job.Status != jobs.StatusSucceeded {
		t.Fatalf("job ended %s: %s", job.Status, job.Error)
	}
	playlist := f.read(t, jobID, playlistName)
	if got := strings.Count(playlist, "#EXTINF"); got != 6 {
		t.Fatalf("playlist has %d segments, want 6:\n%s", got, playlist)
	}
	if !strings.Contains(playlist, "#EXT-X-DISCONTINUITY\n#EXTINF:2.000,\nseg-00003.ts") || !strings.HasSuffix(playlist, "#EXT-X-ENDLIST\n") {
		t.Fatalf("playlist lost the discontinuity or the end:\n%s", playlist)
	}
	if got := f.read(t, jobID, "seg-00004.ts"); got != "data-4" {
		t.Fatalf("segment 4 holds %q", got)
	}
	f.assertVideoFiled(t, jobID, performerID, 12)
}

// assertVideoFiled checks the one video recorded from Nova and its media file.
func (f *fixture) assertVideoFiled(t *testing.T, jobID, performerID string, duration float64) {
	t.Helper()
	videos := f.recordingVideos(t)
	if len(videos) != 1 {
		t.Fatalf("%d recorded videos, want 1", len(videos))
	}
	video := videos[0]
	if !strings.HasPrefix(video.Title, "Nova — live ") || video.Status != schema.MediaStatusAvailable {
		t.Errorf("video = %q %s", video.Title, video.Status)
	}
	if video.DurationSeconds == nil || *video.DurationSeconds != duration {
		t.Errorf("duration = %v, want %v", video.DurationSeconds, duration)
	}
	if len(video.Performers) != 1 || video.Performers[0] != performerID {
		t.Errorf("performers = %v, want the linked %s", video.Performers, performerID)
	}
	if video.Cover == nil || video.Cover.Uri != "blob://recordings/"+jobID+"/poster.jpg" {
		t.Errorf("cover = %+v", video.Cover)
	}
	f.assertCredits(t, video)
	f.assertMediaFile(t, jobID, video.Id, duration)
}

// assertCredits checks the studio and tags of a recording.
func (f *fixture) assertCredits(t *testing.T, video schema.Video) {
	t.Helper()
	studios, _ := f.database.Strings(f.ctx, `SELECT "name" FROM "Studio" WHERE "id" = $id`, db.Vars{"id": video.Studio})
	if len(studios) != 1 || studios[0] != "ALPHA" {
		t.Errorf("studio = %v, want the cam site", studios)
	}
	tags, _ := f.database.Strings(f.ctx,
		`SELECT "name" FROM "Tag" WHERE "id" IN (SELECT value FROM json_each($tags)) ORDER BY "name"`, db.Vars{"tags": video.Tags})
	if strings.Join(tags, ",") != "Live cam,redhead" {
		t.Errorf("tags = %v", tags)
	}
}

// assertMediaFile checks the media file of a recording.
func (f *fixture) assertMediaFile(t *testing.T, jobID, videoID string, duration float64) {
	t.Helper()
	files, err := db.QueryAs[schema.MediaFile](f.ctx, f.database,
		`SELECT * FROM "MediaFile" WHERE "subjectId" = $video`, db.Vars{"video": videoID})
	if err != nil || len(files) != 1 {
		t.Fatalf("media files = %+v %v", files, err)
	}
	file := files[0]
	location, _ := f.service.libraryLocation(f.ctx)
	if file.Location != location.Id || file.Path != "recordings/"+jobID+"/index.m3u8" || file.MimeType != "application/x-mpegURL" {
		t.Errorf("media file = %+v", file)
	}
	if file.SubjectKind != schema.RecordKindVideo || file.SizeBytes == 0 || file.Missing || file.DurationSeconds == nil || *file.DurationSeconds != duration {
		t.Errorf("media file = %+v", file)
	}
	if file.Quality == nil || *file.Quality != "720p" || file.Height == nil || *file.Height != 720 {
		t.Errorf("media file quality = %v %v", file.Quality, file.Height)
	}
}

// TestRecordingReconnectsWithAFreshStreamURL expires the token mid-recording
// and checks that a freshly resolved one continues without duplicates.
func TestRecordingReconnectsWithAFreshStreamURL(t *testing.T) {
	f := newFixture(t, [][]int{{0, 1}, {0, 1, 2}, {2, 3}, {2, 3, 4}}, true)
	jobID := f.start(t, RecordOptions{})
	waitFor(t, "first segments", func() bool { return f.recordedSeconds() >= 6 })
	calls := f.sites.streamCalls.Load()
	f.live.expireToken("2")
	waitFor(t, "segments after the reconnect", func() bool { return f.recordedSeconds() >= 10 })
	f.stop(t)

	if f.sites.streamCalls.Load() <= calls {
		t.Fatal("the stream was never resolved again")
	}
	playlist := f.read(t, jobID, playlistName)
	if got := strings.Count(playlist, "#EXTINF"); got != 5 {
		t.Fatalf("playlist has %d segments, want 5 (no duplicates after the reconnect):\n%s", got, playlist)
	}
	if !strings.Contains(playlist, "#EXT-X-DISCONTINUITY") {
		t.Fatalf("the reconnect left no discontinuity:\n%s", playlist)
	}
}

// TestRecordingEndsAtTheMaximumDuration checks the cut at the longest recording.
func TestRecordingEndsAtTheMaximumDuration(t *testing.T) {
	f := newFixture(t, [][]int{{0, 1, 2}}, true)
	f.service.timing.maxRecording = 150 * time.Millisecond
	jobID, err := f.service.StartRecording(f.ctx, f.roomID, RecordOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if job := f.waitJob(t, jobID); job.Status != jobs.StatusSucceeded {
		t.Fatalf("job ended %s: %s", job.Status, job.Error)
	}
	if len(f.recordingVideos(t)) != 1 {
		t.Fatal("the cut recording was not filed")
	}
}

// TestRecordingSurvivesTheRoomGoingOffline takes the room offline and checks
// that the recording ends after the grace period and is still filed.
func TestRecordingSurvivesTheRoomGoingOffline(t *testing.T) {
	f := newFixture(t, [][]int{{0, 1, 2}}, true)
	jobID := f.start(t, RecordOptions{})
	waitFor(t, "three segments", func() bool { return f.recordedSeconds() >= 6 })
	f.sites.refuseStreams.Store(true)
	f.live.setDown(true)

	if job := f.waitJob(t, jobID); job.Status != jobs.StatusSucceeded {
		t.Fatalf("job ended %s: %s", job.Status, job.Error)
	}
	if videos := f.recordingVideos(t); len(videos) != 1 || *videos[0].DurationSeconds != 6 {
		t.Fatalf("videos = %+v", videos)
	}
}

// TestOneRecordingPerRoomAndNoneOffline checks that a second start returns
// the running job and that an offline room isn't recorded.
func TestOneRecordingPerRoomAndNoneOffline(t *testing.T) {
	f := newFixture(t, [][]int{{0, 1, 2}}, true)
	first := f.start(t, RecordOptions{})
	second, err := f.service.StartRecording(f.ctx, f.roomID, RecordOptions{})
	if err != nil || second != first || len(f.service.ActiveRecordings()) != 1 {
		t.Fatalf("a second recording started: %s %v", second, err)
	}
	f.stop(t)

	f.sites.refuseStreams.Store(true)
	f.service.online = map[string]cachedOnline{}
	if _, err := f.service.StartRecording(f.ctx, f.roomID, RecordOptions{}); !errors.Is(err, ErrOffline) {
		t.Fatalf("starting an offline room = %v, want ErrOffline", err)
	}
	if _, err := f.service.StartRecording(f.ctx, f.roomID, RecordOptions{Quality: "huge"}); err == nil {
		t.Fatal("accepted quality huge")
	}
}

// TestRecordingTooShortToKeepIsDropped stops a recording after one segment.
func TestRecordingTooShortToKeepIsDropped(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, true)
	jobID := f.start(t, RecordOptions{})
	waitFor(t, "one segment", func() bool { return f.recordedSeconds() >= 2 })
	f.stop(t)

	if job := f.waitJob(t, jobID); job.Status != jobs.StatusSucceeded {
		t.Fatalf("job ended %s: %s", job.Status, job.Error)
	}
	if len(f.recordingVideos(t)) != 0 {
		t.Fatal("a 2 s clip became a video")
	}
	if files, _ := f.blobs.List(f.ctx, blobDirectory(jobID)); len(files) != 0 {
		t.Fatalf("files left behind: %v", files)
	}
}

// TestCancellingTheJobFilesTheRecording cancels through the job runner, as
// the jobs UI does, and checks the footage still becomes a video.
func TestCancellingTheJobFilesTheRecording(t *testing.T) {
	f := newFixture(t, [][]int{{0, 1, 2, 3}}, true)
	jobID := f.start(t, RecordOptions{})
	waitFor(t, "four segments", func() bool { return f.recordedSeconds() >= 8 })
	if err := f.runner.Cancel(f.ctx, jobID); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the video", func() bool { return len(f.recordingVideos(t)) == 1 })
	if job, _ := f.queue.Get(f.ctx, jobID); job.Status != jobs.StatusCancelled {
		t.Errorf("job is %s, want cancelled", job.Status)
	}
}

// TestRecordingResumesAfterARestart interrupts a job's attempt the way a
// server shutdown does and checks that the next attempt goes on in the same
// recording, after a break.
func TestRecordingResumesAfterARestart(t *testing.T) {
	sequences := make([]int, 12)
	for index := range sequences {
		sequences[index] = index
	}
	f := newFixture(t, [][]int{sequences, {12, 13, 14}}, false)
	jobID, err := f.service.StartRecording(f.ctx, f.roomID, RecordOptions{})
	if err != nil {
		t.Fatal(err)
	}
	f.runAttemptUntilInterrupted(t, 24)

	if _, flushed := f.service.flushedAt(f.ctx, jobID); !flushed {
		t.Fatal("the interrupted attempt left no playlist")
	}
	f.runner.Start(f.ctx)
	waitFor(t, "the resumed recording", func() bool { return f.recordedSeconds() >= 30 })
	f.stop(t)

	playlist := f.read(t, jobID, playlistName)
	if got := strings.Count(playlist, "#EXTINF"); got != 15 {
		t.Fatalf("playlist has %d segments, want 15:\n%s", got, playlist)
	}
	if !strings.Contains(playlist, "#EXT-X-DISCONTINUITY\n#EXTINF:2.000,\nseg-00012.ts") {
		t.Errorf("the restart left no break:\n%s", playlist)
	}
	if videos := f.recordingVideos(t); len(videos) != 1 || *videos[0].DurationSeconds != 30 {
		t.Fatalf("videos = %+v", videos)
	}
}

// runAttemptUntilInterrupted claims the queued record job, runs it until it
// holds seconds of footage, then ends its context like a shutdown, and puts
// the job back in the queue like the next start does.
func (f *fixture) runAttemptUntilInterrupted(t *testing.T, seconds float64) {
	t.Helper()
	job, err := f.queue.Claim(f.ctx, []jobs.Kind{schema.JobKindRecord})
	if err != nil || job == nil {
		t.Fatalf("claim = %+v %v", job, err)
	}
	attemptContext, cancel := context.WithCancel(f.ctx)
	attempt := make(chan error, 1)
	go func() {
		attempt <- f.service.RecordTask()(attemptContext, job, func(float64) {})
	}()
	waitFor(t, "the footage", func() bool { return f.recordedSeconds() >= seconds })
	cancel()
	if err := <-attempt; !errors.Is(err, context.Canceled) {
		t.Fatalf("interrupted attempt returned %v", err)
	}
	if err := f.queue.ResetStuck(f.ctx); err != nil {
		t.Fatal(err)
	}
}

// TestALateResumeFilesWhatWasRecorded checks that a job resumed long after
// its attempt stopped files the old footage instead of appending a new show.
func TestALateResumeFilesWhatWasRecorded(t *testing.T) {
	sequences := make([]int, 12)
	for index := range sequences {
		sequences[index] = index
	}
	f := newFixture(t, [][]int{sequences, {50, 51}}, false)
	jobID, err := f.service.StartRecording(f.ctx, f.roomID, RecordOptions{})
	if err != nil {
		t.Fatal(err)
	}
	f.runAttemptUntilInterrupted(t, 24)
	time.Sleep(f.service.timing.offlineGrace + 50*time.Millisecond)
	f.runner.Start(f.ctx)

	if job := f.waitJob(t, jobID); job.Status != jobs.StatusSucceeded {
		t.Fatalf("job ended %s: %s", job.Status, job.Error)
	}
	playlist := f.read(t, jobID, playlistName)
	if strings.Contains(playlist, "seg-00012") || !strings.HasSuffix(playlist, "#EXT-X-ENDLIST\n") {
		t.Errorf("late resume appended footage or left the playlist open:\n%s", playlist)
	}
	if len(f.recordingVideos(t)) != 1 {
		t.Fatal("the interrupted recording was not filed")
	}
}

// TestRecordingKeepsASeparateAudioRenditionInSync records a stream with its
// audio in its own playlist.
func TestRecordingKeepsASeparateAudioRenditionInSync(t *testing.T) {
	f := newFixture(t, [][]int{{0, 1, 2}, {1, 2, 3}, {2, 3, 4}}, true)
	f.live.setAudio()
	jobID := f.start(t, RecordOptions{})
	waitFor(t, "five segments", func() bool { return f.recordedSeconds() >= 10 })
	f.stop(t)

	master := f.read(t, jobID, playlistName)
	if !strings.Contains(master, `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio"`) || !strings.Contains(master, "video.m3u8") {
		t.Fatalf("index is not a master joining audio and video:\n%s", master)
	}
	if got := strings.Count(f.read(t, jobID, "audio.m3u8"), "#EXTINF"); got < 5 {
		t.Fatalf("audio playlist has %d segments, want at least 5", got)
	}
	if got := f.read(t, jobID, "aseg-00002.ts"); got != "audio-2" {
		t.Fatalf("audio segment holds %q", got)
	}
	if got := f.read(t, jobID, "seg-00002.ts"); got != "data-2" {
		t.Fatalf("video segment holds %q", got)
	}
}

// TestRecordingRecordsOnlyTheChosenRendition checks which rendition each
// quality records.
func TestRecordingRecordsOnlyTheChosenRendition(t *testing.T) {
	cases := []struct {
		quality string
		want    string
	}{
		{QualitySource, "data-0@720p"},
		{"480", "mid-0@480p"},
		{"480p", "mid-0@480p"},
		{"600", "mid-0@480p"},
		{"240", "low-0@360p"},
		{"1080", "data-0@720p"},
	}
	for _, testCase := range cases {
		f := newFixture(t, [][]int{{0, 1, 2}}, true)
		if got := f.recordAt(t, testCase.quality); got != testCase.want {
			t.Errorf("quality %q recorded %q, want %q", testCase.quality, got, testCase.want)
		}
	}
}

// recordAt records a few segments at a quality and returns the first stored
// segment and the rendition label.
func (f *fixture) recordAt(t *testing.T, quality string) string {
	t.Helper()
	jobID := f.start(t, RecordOptions{Quality: quality})
	waitFor(t, "three segments", func() bool { return f.recordedSeconds() >= 6 })
	label := f.service.ActiveRecordings()[0].Quality
	f.stop(t)
	return f.read(t, jobID, "seg-00000.ts") + "@" + label
}

// TestRecordingRepicksTheRenditionAfterAReconnect checks the quality holds
// across a reconnect.
func TestRecordingRepicksTheRenditionAfterAReconnect(t *testing.T) {
	f := newFixture(t, [][]int{{0, 1}, {0, 1, 2}, {2, 3}, {2, 3, 4}}, true)
	if _, err := f.service.SetAutoRecord(f.ctx, f.roomID, false, "480"); err != nil {
		t.Fatal(err)
	}
	jobID := f.start(t, RecordOptions{})
	waitFor(t, "first segments", func() bool { return f.recordedSeconds() >= 6 })
	f.live.expireToken("2")
	waitFor(t, "segments after the reconnect", func() bool { return f.recordedSeconds() >= 10 })
	f.stop(t)

	if got := f.read(t, jobID, "seg-00004.ts"); got != "mid-4" {
		t.Fatalf("the segment after the reconnect came from %q", got)
	}
}
