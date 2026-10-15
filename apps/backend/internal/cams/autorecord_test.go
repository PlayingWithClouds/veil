package cams

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"
)

// newWatcher returns the auto-record watcher over the fixture.
func newWatcher(fixture *recorderFixture) *AutoRecorder {
	watcher := NewAutoRecorder(fixture.service, fixture.recorder)
	watcher.stagger = time.Millisecond
	return watcher
}

func TestAutoRecordStartsOnlyForFollowedAutoRecordModelsThatAreLive(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{0, 1, 2}})
	ctx := context.Background()
	watcher := newWatcher(fixture)

	watcher.Check(ctx)
	if fixture.recorder.IsRecording(fixture.roomID) {
		t.Fatal("recorded a model without auto-record")
	}

	if _, err := fixture.service.SetAutoRecord(ctx, fixture.roomID, true, nil); err != nil {
		t.Fatal(err)
	}
	fixture.sites.refuseStreams.Store(true)
	fixture.service.mutex.Lock()
	fixture.service.online = map[string]cachedOnline{}
	fixture.service.mutex.Unlock()
	watcher.Check(ctx)
	if fixture.recorder.IsRecording(fixture.roomID) {
		t.Fatal("recorded an offline model")
	}

	fixture.sites.refuseStreams.Store(false)
	fixture.service.mutex.Lock()
	fixture.service.online = map[string]cachedOnline{}
	fixture.service.mutex.Unlock()
	watcher.Check(ctx)
	if !fixture.recorder.IsRecording(fixture.roomID) {
		t.Fatal("did not record a live model")
	}

	watcher.Check(ctx)
	if len(fixture.recorder.Active()) != 1 {
		t.Fatalf("%d recordings of one room", len(fixture.recorder.Active()))
	}
	if _, err := fixture.recorder.Stop(ctx, fixture.roomID); err != nil {
		t.Fatal(err)
	}
}

func TestUnfollowingTurnsAutoRecordOff(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{0}})
	ctx := context.Background()
	enabled, err := fixture.service.SetAutoRecord(ctx, fixture.roomID, true, nil)
	if err != nil || !enabled.AutoRecord || !enabled.Followed {
		t.Fatalf("auto-record did not follow the model: %+v %v", enabled, err)
	}
	unfollowed, err := fixture.service.SetFollowed(ctx, fixture.roomID, false)
	if err != nil || unfollowed.AutoRecord {
		t.Fatalf("auto-record survived unfollowing: %+v %v", unfollowed, err)
	}
	reloaded, _ := fixture.service.load(ctx, fixture.roomID)
	if reloaded.AutoRecord {
		t.Fatal("auto-record still stored after unfollowing")
	}
}

func TestParseMediaPlaylistReadsInitSectionsAndRejectsEncryption(t *testing.T) {
	base, _ := url.Parse("https://edge.example/live/index.m3u8?token=a")
	playlist, err := parseMediaPlaylist(strings.Join([]string{
		"#EXTM3U", "#EXT-X-TARGETDURATION:4", `#EXT-X-MAP:URI="init.mp4"`,
		"#EXTINF:4.0,", "a.m4s?token=a", "#EXT-X-DISCONTINUITY", "#EXTINF:3.5,", "b.m4s?token=a",
	}, "\n"), base)
	if err != nil {
		t.Fatal(err)
	}
	if len(playlist.Segments) != 2 || playlist.Segments[0].InitURL != "https://edge.example/live/init.mp4" ||
		playlist.Segments[0].Identity != "/live/a.m4s" || !playlist.Segments[1].Discontinuity || playlist.Segments[1].Duration != 3.5 {
		t.Fatalf("unexpected playlist %+v", playlist)
	}
	if _, err := parseMediaPlaylist("#EXTM3U\n#EXT-X-KEY:METHOD=AES-128,URI=\"k\"\n#EXTINF:2,\na.ts\n", base); err == nil {
		t.Fatal("encrypted playlist accepted")
	}
}

func TestBestVariantURLPicksTheHighestBandwidth(t *testing.T) {
	base, _ := url.Parse("https://edge.example/live/master.m3u8")
	master := "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360\nlow.m3u8\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=4000000,CODECS=\"avc1.4d401f,mp4a.40.2\"\nhigh/index.m3u8\n"
	if got := chooseVariant(parseVariants(master, base), 0); got == nil || got.URL != "https://edge.example/live/high/index.m3u8" || got.AudioURL != "" {
		t.Fatalf("picked %+v", got)
	}
	if got := chooseVariant(parseVariants("#EXTM3U\n#EXTINF:2,\na.ts\n", base), 0); got != nil {
		t.Fatalf("a media playlist has no variants, got %+v", got)
	}
}

func TestBestVariantFindsTheSeparateAudioRendition(t *testing.T) {
	base, _ := url.Parse("https://edge.example/v1/master.m3u8?token=a")
	master := strings.Join([]string{
		"#EXTM3U",
		`#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="aac_96",NAME="a",DEFAULT=YES,URI="/v1/audio_96.m3u8"`,
		`#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="aac_128",NAME="a",DEFAULT=YES,URI="/v1/audio_128.m3u8"`,
		`#EXT-X-STREAM-INF:BANDWIDTH=900000,AUDIO="aac_96"`, "/v1/low.m3u8",
		`#EXT-X-STREAM-INF:BANDWIDTH=5000000,CODECS="avc1.640028,mp4a.40.2",AUDIO="aac_128"`, "/v1/high.m3u8?s=1",
	}, "\n")
	got := chooseVariant(parseVariants(master, base), 0)
	if got == nil || got.URL != "https://edge.example/v1/high.m3u8?s=1" || got.AudioURL != "https://edge.example/v1/audio_128.m3u8" {
		t.Fatalf("picked %+v", got)
	}
}

func TestAutoRecordRecordsAtTheModelsQuality(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{0, 1, 2}})
	ctx := context.Background()
	quality := "480"
	model, err := fixture.service.SetAutoRecord(ctx, fixture.roomID, true, &quality)
	if err != nil || model.RecordQuality != "480" {
		t.Fatalf("quality not stored: %+v %v", model, err)
	}
	// Leaving the quality out keeps the one chosen.
	if model, err = fixture.service.SetAutoRecord(ctx, fixture.roomID, true, nil); err != nil || model.RecordQuality != "480" {
		t.Fatalf("quality lost: %+v %v", model, err)
	}
	if _, err := fixture.service.SetAutoRecord(ctx, fixture.roomID, true, ptr("tall")); err == nil {
		t.Fatal("accepted an invalid quality")
	}

	newWatcher(fixture).Check(ctx)
	waitFor(t, "the watcher's recording", func() bool { return fixture.recordedSeconds() >= 4 })
	active := fixture.recorder.Active()
	if len(active) != 1 || active[0].Quality == nil || *active[0].Quality != "480p" {
		t.Fatalf("the watcher did not record at 480p: %+v", active)
	}
	if _, err := fixture.recorder.Stop(ctx, fixture.roomID); err != nil {
		t.Fatal(err)
	}
}

// ptr returns a pointer to text.
func ptr(text string) *string { return &text }

func TestChooseVariant(t *testing.T) {
	variants := []streamVariant{
		{URL: "240", Height: 240, Bandwidth: 300000},
		{URL: "480", Height: 480, Bandwidth: 1000000},
		{URL: "480-fat", Height: 480, Bandwidth: 1500000},
		{URL: "1080", Height: 1080, Bandwidth: 5000000},
		{URL: "720", Height: 720, Bandwidth: 3000000},
	}
	cases := []struct {
		target int
		want   string
	}{
		{0, "1080"},      // source: the best
		{720, "720"},     // exact
		{1080, "1080"},   // exact, the top
		{600, "480-fat"}, // closest below, richer copy on a tie
		{2000, "1080"},   // above everything: the tallest
		{100, "240"},     // none below: the lowest above
	}
	for _, testCase := range cases {
		if got := chooseVariant(variants, testCase.target); got == nil || got.URL != testCase.want {
			t.Errorf("target %d chose %+v, want %s", testCase.target, got, testCase.want)
		}
	}
	if got := chooseVariant(nil, 480); got != nil {
		t.Errorf("a media playlist has no variants, got %+v", got)
	}
	unknown := []streamVariant{{URL: "a", Bandwidth: 1}, {URL: "b", Bandwidth: 9}}
	if got := chooseVariant(unknown, 480); got.URL != "b" {
		t.Errorf("without heights the best bitrate wins, got %+v", got)
	}
	mixed := []streamVariant{{URL: "unknown", Bandwidth: 9}, {URL: "360", Height: 360, Bandwidth: 1}}
	if got := chooseVariant(mixed, 480); got.URL != "360" {
		t.Errorf("a rendition without a height must not win once others have one, got %+v", got)
	}
}

func TestParseQuality(t *testing.T) {
	for text, want := range map[string]int{"": 0, "source": 0, "SOURCE": 0, "480": 480, "480p": 480, " 720P ": 720} {
		if got, err := ParseQuality(text); err != nil || got != want {
			t.Errorf("ParseQuality(%q) = %d, %v; want %d", text, got, err, want)
		}
	}
	for _, text := range []string{"huge", "-5", "0", "4k"} {
		if _, err := ParseQuality(text); err == nil {
			t.Errorf("ParseQuality(%q) accepted", text)
		}
	}
}
