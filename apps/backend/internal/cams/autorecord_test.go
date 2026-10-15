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

	if _, err := fixture.service.SetAutoRecord(ctx, fixture.roomID, true); err != nil {
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
	enabled, err := fixture.service.SetAutoRecord(ctx, fixture.roomID, true)
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
	if got := bestVariantURL(master, base); got != "https://edge.example/live/high/index.m3u8" {
		t.Fatalf("picked %q", got)
	}
	if got := bestVariantURL("#EXTM3U\n#EXTINF:2,\na.ts\n", base); got != "" {
		t.Fatalf("a media playlist has no variants, got %q", got)
	}
}
