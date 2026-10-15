package cams

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/storage"
)

// liveServer is a fake HLS origin. Each playlist request shows the next
// window of segments; the last window repeats once the list is exhausted.
type liveServer struct {
	server *httptest.Server

	mutex sync.Mutex
	// windows lists, per playlist request, the sequence numbers shown.
	windows [][]int
	// discontinuityAt lists sequence numbers preceded by a discontinuity tag.
	discontinuityAt map[int]bool
	requests        int
	// validToken is the only token the origin accepts; others get 403.
	validToken string
	// down makes every request fail with 503.
	down bool
}

// newLiveServer starts the origin.
func newLiveServer(t *testing.T, windows [][]int) *liveServer {
	t.Helper()
	live := &liveServer{windows: windows, discontinuityAt: map[int]bool{}, validToken: "1"}
	live.server = httptest.NewServer(http.HandlerFunc(live.serve))
	t.Cleanup(live.server.Close)
	return live
}

// serve answers master, media and segment requests.
func (live *liveServer) serve(writer http.ResponseWriter, request *http.Request) {
	live.mutex.Lock()
	defer live.mutex.Unlock()
	if live.down {
		http.Error(writer, "down", http.StatusServiceUnavailable)
		return
	}
	if request.URL.Query().Get("token") != live.validToken {
		http.Error(writer, "expired", http.StatusForbidden)
		return
	}
	switch {
	case strings.HasSuffix(request.URL.Path, "/master.m3u8"):
		fmt.Fprint(writer, "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=500000\nlow.m3u8?token="+live.validToken+
			"\n#EXT-X-STREAM-INF:BANDWIDTH=2000000\nhigh.m3u8?token="+live.validToken+"\n")
	case strings.HasSuffix(request.URL.Path, "/high.m3u8"):
		live.writeMedia(writer)
	case strings.HasSuffix(request.URL.Path, "/low.m3u8"):
		http.Error(writer, "the low variant must not be recorded", http.StatusTeapot)
	case strings.HasSuffix(request.URL.Path, ".ts"):
		fmt.Fprint(writer, "data-"+strings.TrimSuffix(request.URL.Path[strings.LastIndex(request.URL.Path, "/")+1:], ".ts"))
	default:
		http.NotFound(writer, request)
	}
}

// writeMedia writes the media playlist for the current request.
func (live *liveServer) writeMedia(writer io.Writer) {
	step := live.requests
	if step >= len(live.windows) {
		step = len(live.windows) - 1
	}
	live.requests++
	window := live.windows[step]
	fmt.Fprintf(writer, "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:2\n#EXT-X-MEDIA-SEQUENCE:%d\n", window[0])
	for _, sequence := range window {
		if live.discontinuityAt[sequence] {
			fmt.Fprint(writer, "#EXT-X-DISCONTINUITY\n")
		}
		fmt.Fprintf(writer, "#EXTINF:2.000,\nseg/%d.ts?token=%s\n", sequence, live.validToken)
	}
}

// expireToken invalidates the current token and returns the new valid one.
func (live *liveServer) expireToken(next string) {
	live.mutex.Lock()
	defer live.mutex.Unlock()
	live.validToken = next
}

// setDown switches the origin off or on.
func (live *liveServer) setDown(down bool) {
	live.mutex.Lock()
	defer live.mutex.Unlock()
	live.down = down
}

// requestCount is how many playlist requests were answered.
func (live *liveServer) requestCount() int {
	live.mutex.Lock()
	defer live.mutex.Unlock()
	return live.requests
}

// liveSites resolves a room to the fake origin with the origin's current token.
type liveSites struct {
	fakeSites
	live          *liveServer
	streamCalls   atomic.Int32
	refuseStreams atomic.Bool
}

// Stream returns the origin's master playlist URL carrying its current token.
func (sites *liveSites) Stream(ctx context.Context, plugin *plugins.Plugin, username string) (*plugins.ResolveResult, error) {
	sites.streamCalls.Add(1)
	if sites.refuseStreams.Load() {
		return nil, fmt.Errorf("offline")
	}
	sites.live.mutex.Lock()
	token := sites.live.validToken
	sites.live.mutex.Unlock()
	return &plugins.ResolveResult{URL: sites.live.server.URL + "/live/master.m3u8?token=" + token, MimeType: "application/x-mpegURL"}, nil
}

// recorderFixture is a Recorder over a real database, blob store and ingest.
type recorderFixture struct {
	service  *Service
	recorder *Recorder
	sites    *liveSites
	live     *liveServer
	blobs    *storage.Client
	roomID   string
}

// newRecorderFixture builds a recorder against a fake origin with one known room.
func newRecorderFixture(t *testing.T, windows [][]int) *recorderFixture {
	t.Helper()
	live := newLiveServer(t, windows)
	sites := &liveSites{fakeSites: fakeSites{rooms: map[string][]plugins.CamRoom{"alpha": {room("nova", 10)}}}, live: live}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	database, err := db.Open(ctx, t.TempDir()+"/test.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	service := New(database, sites, fakeProxy{})
	blobs, err := storage.New(t.TempDir(), "http://backend.test")
	if err != nil {
		t.Fatal(err)
	}
	thumbnail := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "jpeg-bytes")
	}))
	t.Cleanup(thumbnail.Close)
	stored, err := service.upsertRoom(ctx, "alpha", plugins.CamRoom{
		Username: "nova", DisplayName: "Nova", SourceURL: "https://site.example/nova/", ThumbnailURL: thumbnail.URL + "/nova.jpg",
		Tags: []string{"redhead"}, Online: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	recorder := NewRecorder(ctx, service, blobs, NewHTTPFetcher(), ingest.New(database), func() time.Duration { return time.Hour })
	recorder.offlineGrace = 300 * time.Millisecond
	recorder.retryDelay = 10 * time.Millisecond
	recorder.pollMinimum = 5 * time.Millisecond
	recorder.pollMaximum = 10 * time.Millisecond
	return &recorderFixture{service: service, recorder: recorder, sites: sites, live: live, blobs: blobs, roomID: stored.ID}
}

// waitFor polls until condition holds or the test times out.
func waitFor(t *testing.T, description string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", description)
}

// recordedSeconds is how much footage the running recording has stored.
func (fixture *recorderFixture) recordedSeconds() float64 {
	active := fixture.recorder.Active()
	if len(active) == 0 {
		return 0
	}
	return active[0].DurationSeconds
}

// playlistOf reads the stored playlist of a finished recording.
func (fixture *recorderFixture) playlistOf(t *testing.T, recordID string) string {
	t.Helper()
	_, identifier, _ := strings.Cut(recordID, ":")
	reader, err := fixture.blobs.Get(context.Background(), RecordingPrefix+identifier+"/"+playlistName)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, _ := io.ReadAll(reader)
	return string(data)
}

func TestRecorderStoresEachSegmentOnceAndFilesAScene(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{0, 1, 2}, {1, 2, 3}, {2, 3, 4}, {3, 4, 5}})
	fixture.live.discontinuityAt[3] = true
	ctx := context.Background()

	started, err := fixture.recorder.Start(ctx, fixture.roomID)
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, "six segments", func() bool { return fixture.recordedSeconds() >= 12 })
	finished, err := fixture.recorder.Stop(ctx, fixture.roomID)
	if err != nil || finished == nil {
		t.Fatalf("stop: %v %v", finished, err)
	}

	if finished.ID != started.ID || finished.Status != statusFinished || finished.DurationSeconds != 12 {
		t.Fatalf("unexpected recording %+v", finished)
	}
	playlist := fixture.playlistOf(t, finished.ID)
	if got := strings.Count(playlist, "#EXTINF"); got != 6 {
		t.Fatalf("playlist has %d segments, want 6:\n%s", got, playlist)
	}
	if !strings.Contains(playlist, "#EXT-X-DISCONTINUITY\n#EXTINF:2.000,\nseg-00003.ts") || !strings.HasSuffix(playlist, "#EXT-X-ENDLIST\n") {
		t.Fatalf("playlist lost the discontinuity or the end:\n%s", playlist)
	}
	_, identifier, _ := strings.Cut(finished.ID, ":")
	reader, err := fixture.blobs.Get(ctx, RecordingPrefix+identifier+"/seg-00004.ts")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(reader)
	reader.Close()
	if string(body) != "data-4" {
		t.Fatalf("segment 4 holds %q", body)
	}
	fixture.assertSceneFiled(t, finished.SceneID, 12)
}

// assertSceneFiled checks the scene made from a recording of Nova.
func (fixture *recorderFixture) assertSceneFiled(t *testing.T, sceneID *string, duration int) {
	t.Helper()
	if sceneID == nil {
		t.Fatal("recording has no scene")
	}
	ctx := context.Background()
	scene, err := fixture.service.database.GetString(ctx, *sceneID)
	if err != nil || scene == nil {
		t.Fatalf("scene missing: %v", err)
	}
	title, _ := scene["title"].(string)
	if !strings.HasPrefix(title, "Nova — live ") {
		t.Fatalf("title %q", title)
	}
	if scene["duration_seconds"] != int64(duration) {
		t.Fatalf("duration %v", scene["duration_seconds"])
	}
	poster, _ := scene["poster_path"].(string)
	if !strings.HasPrefix(poster, "http://backend.test/api/blob/recordings/") {
		t.Fatalf("poster %q", poster)
	}
	performers, err := fixture.service.database.Strings(ctx,
		`SELECT p.name FROM performer p WHERE p.id IN (SELECT value FROM json_each((SELECT performers FROM scene WHERE id = $id)))`, db.Vars{"id": *sceneID})
	if err != nil || len(performers) != 1 || performers[0] != "Nova" {
		t.Fatalf("performers %v %v", performers, err)
	}
	urls, _ := fixture.service.database.Strings(ctx, `SELECT url FROM stream WHERE media = $id`, db.Vars{"id": *sceneID})
	if len(urls) != 1 || !strings.HasSuffix(urls[0], "/"+playlistName) || !strings.Contains(urls[0], "/api/blob/recordings/") {
		t.Fatalf("streams %v", urls)
	}
	detailFetched, _ := fixture.service.database.Int(ctx, `SELECT count(*) FROM scene WHERE id = $id AND detail_fetched_at IS NOT NULL`, db.Vars{"id": *sceneID})
	if detailFetched != 1 {
		t.Fatal("scene would be fetched from its recording: URL as if it were a site page")
	}
}

func TestRecorderReconnectsWithAFreshStreamURL(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{0, 1}, {0, 1, 2}, {2, 3}, {2, 3, 4}})
	ctx := context.Background()
	if _, err := fixture.recorder.Start(ctx, fixture.roomID); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "first segments", func() bool { return fixture.recordedSeconds() >= 6 })
	// The token the recorder holds stops working; only a freshly resolved one does.
	fixture.live.expireToken("2")
	waitFor(t, "segments after the reconnect", func() bool { return fixture.recordedSeconds() >= 10 })
	finished, err := fixture.recorder.Stop(ctx, fixture.roomID)
	if err != nil || finished == nil {
		t.Fatalf("stop: %v %v", finished, err)
	}

	if fixture.sites.streamCalls.Load() < 2 {
		t.Fatal("the stream was never resolved again")
	}
	playlist := fixture.playlistOf(t, finished.ID)
	if got := strings.Count(playlist, "#EXTINF"); got != 5 {
		t.Fatalf("playlist has %d segments, want 5 (no duplicates after the reconnect):\n%s", got, playlist)
	}
	if !strings.Contains(playlist, "#EXT-X-DISCONTINUITY") {
		t.Fatalf("the reconnect left no discontinuity:\n%s", playlist)
	}
}

func TestRecorderEndsAtTheMaximumDuration(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{0, 1, 2}})
	fixture.recorder.maxDuration = func() time.Duration { return 150 * time.Millisecond }
	if _, err := fixture.recorder.Start(context.Background(), fixture.roomID); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the recording to end by itself", func() bool { return !fixture.recorder.IsRecording(fixture.roomID) })

	recordings, err := fixture.recorder.Recordings(context.Background(), fixture.roomID, false)
	if err != nil || len(recordings) != 1 {
		t.Fatalf("recordings %v %v", recordings, err)
	}
	if recordings[0].StopReason == nil || *recordings[0].StopReason != reasonMaxDuration || recordings[0].SceneID == nil {
		t.Fatalf("unexpected recording %+v", recordings[0])
	}
}

func TestRecorderEndsWhenTheRoomStaysOffline(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{0, 1, 2}})
	if _, err := fixture.recorder.Start(context.Background(), fixture.roomID); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "three segments", func() bool { return fixture.recordedSeconds() >= 6 })
	fixture.sites.refuseStreams.Store(true)
	fixture.live.setDown(true)
	waitFor(t, "the recording to end", func() bool { return !fixture.recorder.IsRecording(fixture.roomID) })

	recordings, _ := fixture.recorder.Recordings(context.Background(), fixture.roomID, false)
	if len(recordings) != 1 || recordings[0].StopReason == nil || *recordings[0].StopReason != reasonOffline || recordings[0].SceneID == nil {
		t.Fatalf("unexpected recordings %+v", recordings)
	}
}

func TestRecorderStartsOneRecordingPerRoomAndRefusesOfflineRooms(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{0, 1, 2}})
	ctx := context.Background()
	first, err := fixture.recorder.Start(ctx, fixture.roomID)
	if err != nil {
		t.Fatal(err)
	}
	second, err := fixture.recorder.Start(ctx, fixture.roomID)
	if err != nil || second.ID != first.ID || len(fixture.recorder.Active()) != 1 {
		t.Fatalf("a second recording started: %+v %v", second, err)
	}
	if _, err := fixture.recorder.Stop(ctx, fixture.roomID); err != nil {
		t.Fatal(err)
	}

	fixture.sites.refuseStreams.Store(true)
	if _, err := fixture.recorder.Start(ctx, fixture.roomID); err == nil {
		t.Fatal("an offline room started recording")
	}
	if fixture.recorder.IsRecording(fixture.roomID) {
		t.Fatal("a failed start left the room reserved")
	}
}

func TestRecorderDropsARecordingTooShortToKeep(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{0}})
	ctx := context.Background()
	started, err := fixture.recorder.Start(ctx, fixture.roomID)
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, "one segment", func() bool { return fixture.recordedSeconds() >= 2 })
	finished, err := fixture.recorder.Stop(ctx, fixture.roomID)
	if err != nil || finished == nil {
		t.Fatalf("stop: %v %v", finished, err)
	}
	if finished.Status != statusFailed || finished.SceneID != nil {
		t.Fatalf("a 2 s clip became a scene: %+v", finished)
	}
	_, identifier, _ := strings.Cut(started.ID, ":")
	files, _ := fixture.blobs.List(ctx, RecordingPrefix+identifier+"/")
	if len(files) != 0 {
		t.Fatalf("files left behind: %v", files)
	}
}

func TestRecoverFilesAnInterruptedRecording(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}})
	ctx := context.Background()
	if _, err := fixture.recorder.Start(ctx, fixture.roomID); err != nil {
		t.Fatal(err)
	}
	// Twelve segments cross the flush threshold, so a playlist exists on disk.
	waitFor(t, "a flushed playlist", func() bool { return fixture.recordedSeconds() >= 24 })
	active := fixture.recorder.Active()[0]
	// Simulate the crash: the capture vanishes without finishing.
	fixture.recorder.mutex.Lock()
	crashed := fixture.recorder.active[fixture.roomID]
	delete(fixture.recorder.active, fixture.roomID)
	fixture.recorder.mutex.Unlock()
	crashed.stopRequested.Store(true)
	crashed.cancel()
	<-crashed.finished
	if _, err := fixture.service.database.Exec(ctx, `UPDATE cam_recording SET status = 'recording', scene = NULL WHERE id = $id`, db.Vars{"id": active.ID}); err != nil {
		t.Fatal(err)
	}

	fixture.recorder.Recover(ctx)

	recordings, err := fixture.recorder.Recordings(ctx, fixture.roomID, false)
	if err != nil || len(recordings) != 1 || recordings[0].SceneID == nil || recordings[0].Status != statusFinished {
		t.Fatalf("recording not recovered: %+v %v", recordings, err)
	}
}
