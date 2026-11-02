package cams

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/playingwithclouds/veil/internal/livebuffer"
)

// bufferSegments plays segments 0..last of the 720p rendition to the room's
// live buffer, as a viewer's player would, and waits until they are on disk.
func (f *fixture) bufferSegments(t *testing.T, last int) {
	t.Helper()
	if f.service.services.Buffer == nil {
		f.service.services.Buffer = livebuffer.New(f.ctx, f.blobs)
	}
	var text strings.Builder
	text.WriteString("#EXTM3U\n#EXT-X-TARGETDURATION:2\n")
	for sequence := 0; sequence <= last; sequence++ {
		fmt.Fprintf(&text, "#EXTINF:2.000,\nseg/%d.ts?token=1\n", sequence)
	}
	playlistURL := f.live.server.URL + "/live/high.m3u8?token=1"
	base, _ := url.Parse(playlistURL)
	playlist, err := livebuffer.ParsePlaylist(text.String(), base)
	if err != nil {
		t.Fatal(err)
	}
	source := livebuffer.Source{PlaylistURL: playlistURL, Kind: livebuffer.KindVideo, Fetch: f.service.fetcher.Fetch}
	window := f.service.services.Buffer.Observe(f.roomID, source, playlist, true)
	for _, segment := range window.Segments {
		if _, err := f.service.services.Buffer.Read(f.ctx, f.roomID, segment.Name); err != nil {
			t.Fatal(err)
		}
	}
}

// recordFromPast starts a recording secondsAgo in the past, lets it run until
// it holds the footage, stops it and returns its job id and rendition.
func (f *fixture) recordFromPast(t *testing.T, quality string, secondsAgo int, seconds float64) (jobID string, qualityLabel string) {
	t.Helper()
	jobID = f.start(t, RecordOptions{Quality: quality, FromSecondsAgo: secondsAgo})
	waitFor(t, "the footage", func() bool { return f.recordedSeconds() >= seconds })
	qualityLabel = f.service.ActiveRecordings()[0].Quality
	f.stop(t)
	return jobID, qualityLabel
}

// TestRecordingStartsInTheBufferedPastAndGoesLive checks that a recording
// from the past begins with buffered footage and continues live in the
// buffered rendition, seamlessly.
func TestRecordingStartsInTheBufferedPastAndGoesLive(t *testing.T) {
	f := newFixture(t, [][]int{{6, 7, 8}, {7, 8, 9}, {8, 9, 10}, {9, 10, 11}}, true)
	f.bufferSegments(t, 7)

	jobID, quality := f.recordFromPast(t, "480", 6, 14)

	for number, want := range []string{"data-5", "data-6", "data-7", "data-8", "data-9"} {
		if got := f.read(t, jobID, fmt.Sprintf("seg-%05d.ts", number)); got != want {
			t.Errorf("segment %d = %q, want %q", number, got, want)
		}
	}
	if quality != "720p" {
		t.Errorf("recorded %s, want the buffered 720p rather than the requested 480", quality)
	}
	if playlist := f.read(t, jobID, playlistName); strings.Contains(playlist, "#EXT-X-DISCONTINUITY") {
		t.Errorf("seamless hand-over from the buffer to live was marked as a break:\n%s", playlist)
	}
}

// TestRecordingFromPastIsCappedToTheBuffer asks for more past than buffered.
func TestRecordingFromPastIsCappedToTheBuffer(t *testing.T) {
	f := newFixture(t, [][]int{{6, 7, 8}, {7, 8, 9}, {8, 9, 10}}, true)
	f.bufferSegments(t, 7)

	jobID, _ := f.recordFromPast(t, QualitySource, 3600, 22)

	if got := f.read(t, jobID, "seg-00000.ts"); got != "data-0" {
		t.Errorf("first segment = %q, want the oldest buffered one", got)
	}
	if got := f.read(t, jobID, "seg-00008.ts"); got != "data-8" {
		t.Errorf("segment after the buffered ones = %q, want the live data-8", got)
	}
}

// TestRecordingFromPastWithoutABufferStartsLive asks for past footage of a
// room nobody watched.
func TestRecordingFromPastWithoutABufferStartsLive(t *testing.T) {
	f := newFixture(t, [][]int{{6, 7, 8}, {7, 8, 9}, {8, 9, 10}}, true)

	jobID, _ := f.recordFromPast(t, QualitySource, 120, 6)

	if got := f.read(t, jobID, "seg-00000.ts"); got != "data-6" {
		t.Errorf("first segment = %q, want the first live one", got)
	}
}

// TestRecordingMarksTheGapBetweenBufferAndLive checks the break when the
// buffer's newest segment already left the live playlist.
func TestRecordingMarksTheGapBetweenBufferAndLive(t *testing.T) {
	f := newFixture(t, [][]int{{6, 7, 8}, {7, 8, 9}, {8, 9, 10}}, true)
	f.bufferSegments(t, 3)

	jobID, _ := f.recordFromPast(t, QualitySource, 4, 10)

	playlist := f.read(t, jobID, playlistName)
	if !strings.Contains(playlist, "#EXT-X-DISCONTINUITY\n#EXTINF:2.000,\nseg-00002.ts") {
		t.Errorf("the missing footage between the buffer and live is not marked:\n%s", playlist)
	}
	if got := f.read(t, jobID, "seg-00002.ts"); got != "data-6" {
		t.Errorf("first live segment = %q, want data-6", got)
	}
}
