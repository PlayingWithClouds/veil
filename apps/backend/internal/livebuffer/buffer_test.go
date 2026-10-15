package livebuffer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/storage"
)

// fixture is a Manager over a real blob store and a fake origin that serves
// "bytes-<name>" for every path.
type fixture struct {
	manager *Manager
	store   *storage.Client
	origin  *httptest.Server
	clock   *fakeClock
}

// fakeClock is a settable time source.
type fakeClock struct {
	mutex   sync.Mutex
	current time.Time
}

// Now returns the clock's time.
func (clock *fakeClock) Now() time.Time {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()
	return clock.current
}

// Advance moves the clock forward.
func (clock *fakeClock) Advance(duration time.Duration) {
	clock.mutex.Lock()
	defer clock.mutex.Unlock()
	clock.current = clock.current.Add(duration)
}

// newFixture builds a manager with a 10 second window.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	store, err := storage.New(t.TempDir(), "http://backend.test")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	origin := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprint(writer, "bytes-"+strings.Trim(request.URL.Path, "/"))
	}))
	t.Cleanup(origin.Close)
	clock := &fakeClock{current: time.Unix(1_000_000, 0)}
	manager := New(ctx, store)
	manager.now = clock.Now
	manager.window = 10
	return &fixture{manager: manager, store: store, origin: origin, clock: clock}
}

// source is a rendition at the fake origin.
func (f *fixture) source(playlist string, kind string) Source {
	return Source{
		PlaylistURL: f.origin.URL + "/live/" + playlist + ".m3u8?token=1", Kind: kind,
		Fetch: func(ctx context.Context, rawURL string, headers map[string]string) (*http.Response, error) {
			request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
			if err != nil {
				return nil, err
			}
			return http.DefaultClient.Do(request)
		},
	}
}

// playlistOf builds a 2 second segment playlist over the sequence numbers.
func (f *fixture) playlistOf(t *testing.T, sequences ...int) *Playlist {
	t.Helper()
	var text strings.Builder
	fmt.Fprintf(&text, "#EXTM3U\n#EXT-X-TARGETDURATION:2\n#EXT-X-MEDIA-SEQUENCE:%d\n", sequences[0])
	for _, sequence := range sequences {
		fmt.Fprintf(&text, "#EXTINF:2.000,\nseg/%d.ts?token=1\n", sequence)
	}
	base, _ := url.Parse(f.origin.URL + "/live/video.m3u8")
	playlist, err := ParsePlaylist(text.String(), base)
	if err != nil {
		t.Fatal(err)
	}
	return playlist
}

// waitStored waits until every segment of the window can be read.
func (f *fixture) waitStored(t *testing.T, roomID string, window *Window) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, segment := range window.Segments {
		if _, err := f.manager.Read(ctx, roomID, segment.Name); err != nil {
			t.Fatalf("segment %s: %v", segment.Identity, err)
		}
	}
}

func TestObserveBuffersAndTrimsToWindow(t *testing.T) {
	f := newFixture(t)
	room := "cam_model:nova"
	var window *Window
	for start := 0; start <= 8; start += 2 {
		window = f.manager.Observe(room, f.source("video", KindVideo), f.playlistOf(t, start, start+1, start+2), true)
	}
	f.waitStored(t, room, window)
	if seconds := window.Seconds(); seconds > 10 || seconds < 8 {
		t.Fatalf("window holds %.0f s, want at most 10", seconds)
	}
	if got := window.Segments[len(window.Segments)-1].Identity; got != "/live/seg/10.ts" {
		t.Errorf("newest segment = %s", got)
	}
	if window.MediaSequence == 0 {
		t.Errorf("media sequence did not advance after trimming")
	}
	files, _ := f.store.List(context.Background(), Prefix)
	if len(files) != len(window.Segments) {
		t.Errorf("%d blobs on disk for %d buffered segments", len(files), len(window.Segments))
	}
	data, err := f.manager.Read(context.Background(), room, window.Segments[0].Name)
	if err != nil || !strings.HasPrefix(string(data), "bytes-live/seg/") {
		t.Errorf("segment bytes = %q (%v)", data, err)
	}
}

func TestTrimmedSegmentsAreNotReAdded(t *testing.T) {
	f := newFixture(t)
	room := "cam_model:nova"
	f.manager.Observe(room, f.source("video", KindVideo), f.playlistOf(t, 0, 1, 2, 3, 4, 5, 6, 7), true)
	window := f.manager.Observe(room, f.source("video", KindVideo), f.playlistOf(t, 0, 1, 2, 3, 4, 5, 6, 7), true)
	if len(window.Segments) != 5 {
		t.Fatalf("buffered %d segments, want 5 after trimming to 10 s", len(window.Segments))
	}
	if window.Segments[0].Identity != "/live/seg/3.ts" {
		t.Errorf("oldest = %s, trimmed segments came back", window.Segments[0].Identity)
	}
}

func TestMissedSegmentsMarkDiscontinuity(t *testing.T) {
	f := newFixture(t)
	room := "cam_model:nova"
	f.manager.Observe(room, f.source("video", KindVideo), f.playlistOf(t, 0, 1, 2), true)
	window := f.manager.Observe(room, f.source("video", KindVideo), f.playlistOf(t, 9, 10), true)
	rendered := window.Render(func(name, sourceURL string) string { return name })
	if strings.Count(rendered, "#EXT-X-DISCONTINUITY\n") != 1 {
		t.Fatalf("playlist should mark exactly the gap:\n%s", rendered)
	}
	if window.Segments[3].Discontinuity != true || window.Segments[2].Discontinuity {
		t.Errorf("break not on the first segment after the gap: %+v", window.Segments)
	}
}

func TestRenderIsASlidingLivePlaylist(t *testing.T) {
	f := newFixture(t)
	window := f.manager.Observe("cam_model:nova", f.source("video", KindVideo), f.playlistOf(t, 4, 5), true)
	rendered := window.Render(func(name, sourceURL string) string { return "/x/" + name })
	for _, want := range []string{"#EXT-X-MEDIA-SEQUENCE:4", "#EXT-X-DISCONTINUITY-SEQUENCE:0", "#EXT-X-TARGETDURATION:2", "#EXTINF:2.000,\n/x/"} {
		if !strings.Contains(rendered, want) {
			t.Errorf("playlist lacks %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "ENDLIST") || strings.Contains(rendered, "PLAYLIST-TYPE") {
		t.Errorf("a live window must stay open:\n%s", rendered)
	}
}

func TestObserveWithoutCreateIgnoresUnwatchedRooms(t *testing.T) {
	f := newFixture(t)
	if window := f.manager.Observe("cam_model:nova", f.source("video", KindVideo), f.playlistOf(t, 0), false); window != nil {
		t.Fatalf("an unwatched room was buffered")
	}
	f.manager.Observe("cam_model:nova", f.source("video", KindVideo), f.playlistOf(t, 0), true)
	if window := f.manager.Observe("cam_model:nova", f.source("video", KindVideo), f.playlistOf(t, 0, 1), false); window == nil {
		t.Fatalf("a watched room stopped being buffered")
	}
}

func TestSweepDropsIdleRoomsAndKeepsTouchedOnes(t *testing.T) {
	f := newFixture(t)
	window := f.manager.Observe("cam_model:idle", f.source("video", KindVideo), f.playlistOf(t, 0, 1), true)
	busyWindow := f.manager.Observe("cam_model:busy", f.source("video", KindVideo), f.playlistOf(t, 0, 1), true)
	f.waitStored(t, "cam_model:idle", window)
	f.waitStored(t, "cam_model:busy", busyWindow)
	f.clock.Advance(IdleTimeout - time.Second)
	f.manager.Observe("cam_model:busy", f.source("video", KindVideo), f.playlistOf(t, 0, 1, 2), true)
	f.clock.Advance(2 * time.Second)
	f.manager.Sweep()
	if f.manager.Has("cam_model:idle") {
		t.Errorf("idle room kept")
	}
	if !f.manager.Has("cam_model:busy") {
		t.Errorf("touched room dropped")
	}
	files, _ := f.store.List(context.Background(), Prefix+"idle/")
	if len(files) != 0 {
		t.Errorf("idle room left %d blobs", len(files))
	}
	files, _ = f.store.List(context.Background(), Prefix+"busy/")
	if len(files) == 0 {
		t.Errorf("busy room lost its blobs")
	}
}

func TestSweepDropsRenditionsNoLongerObserved(t *testing.T) {
	f := newFixture(t)
	room := "cam_model:nova"
	f.manager.Observe(room, f.source("old", KindVideo), f.playlistOf(t, 0, 1), true)
	f.clock.Advance(IdleTimeout - time.Second)
	f.manager.Observe(room, f.source("new", KindVideo), f.playlistOf(t, 5, 6), true)
	f.clock.Advance(2 * time.Second)
	f.manager.Sweep()
	windows := f.manager.Windows(room)
	if len(windows) != 1 || windows[0].TrackKey != TrackKey(f.source("new", KindVideo).PlaylistURL) {
		t.Fatalf("windows after sweep = %+v, want only the new rendition", windows)
	}
}

func TestResetClearsLeftoverBuffers(t *testing.T) {
	f := newFixture(t)
	if err := f.store.Put(context.Background(), Prefix+"old/seg.ts", strings.NewReader("x"), 1, "video/mp2t"); err != nil {
		t.Fatal(err)
	}
	if err := f.store.Put(context.Background(), "recordings/keep/seg.ts", strings.NewReader("x"), 1, "video/mp2t"); err != nil {
		t.Fatal(err)
	}
	if err := f.manager.Reset(context.Background()); err != nil {
		t.Fatal(err)
	}
	if files, _ := f.store.List(context.Background(), Prefix); len(files) != 0 {
		t.Errorf("buffers survived the reset: %v", files)
	}
	if !f.store.Exists(context.Background(), "recordings/keep/seg.ts") {
		t.Errorf("reset touched a recording")
	}
}

func TestSeedFromKeepsTheNewestSecondsCappedToTheBuffer(t *testing.T) {
	f := newFixture(t)
	f.manager.window = 100
	room := "cam_model:nova"
	f.manager.Observe(room, f.source("video", KindVideo), f.playlistOf(t, 0, 1, 2, 3, 4, 5, 6, 7), true)
	f.manager.Observe(room, f.source("audio", KindAudio), f.playlistOf(t, 0, 1, 2, 3, 4, 5, 6, 7), true)

	seed := f.manager.SeedFrom(room, 6)
	if seed == nil || len(seed.Video.Segments) != 3 || seed.Seconds != 6 {
		t.Fatalf("6 s seed = %+v", seed)
	}
	if seed.Video.Segments[2].Identity != "/live/seg/7.ts" || seed.Video.Segments[0].Identity != "/live/seg/5.ts" {
		t.Errorf("seed is not the newest segments: %+v", seed.Video.Segments)
	}
	if seed.Audio == nil || len(seed.Audio.Segments) != 3 {
		t.Errorf("audio seed = %+v, want 3 segments", seed.Audio)
	}
	capped := f.manager.SeedFrom(room, 600)
	if capped == nil || capped.Seconds != 16 {
		t.Errorf("a request beyond the buffer should yield all 16 s, got %+v", capped)
	}
	if f.manager.SeedFrom(room, 0) != nil || f.manager.SeedFrom("cam_model:other", 30) != nil {
		t.Errorf("no seed expected without a request or a buffer")
	}
	if got := f.manager.BufferedSeconds(room); got != 16 {
		t.Errorf("buffered seconds = %.0f, want 16", got)
	}
}

func TestOpenFailsForSegmentsThatLeftTheBuffer(t *testing.T) {
	f := newFixture(t)
	room := "cam_model:nova"
	first := f.manager.Observe(room, f.source("video", KindVideo), f.playlistOf(t, 0, 1), true)
	f.manager.Observe(room, f.source("video", KindVideo), f.playlistOf(t, 2, 3, 4, 5, 6, 7), true)
	if _, _, err := f.manager.Open(context.Background(), room, first.Segments[0].Name); err == nil {
		t.Errorf("a trimmed segment can still be opened")
	}
	reader, _, err := f.manager.Open(context.Background(), room, f.manager.Windows(room)[0].Segments[0].Name)
	if err != nil {
		t.Fatal(err)
	}
	reader.Close()
}
