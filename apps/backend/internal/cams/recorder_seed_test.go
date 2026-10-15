package cams

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/playingwithclouds/veil/internal/livebuffer"
)

// bufferSegments plays segments 0..last of the 720p rendition to the room's
// live buffer, as a viewer's player would, and waits until they are on disk.
func (fixture *recorderFixture) bufferSegments(t *testing.T, last int) {
	t.Helper()
	if fixture.recorder.buffer == nil {
		fixture.recorder.SetLiveBuffer(livebuffer.New(context.Background(), fixture.blobs))
	}
	var text strings.Builder
	text.WriteString("#EXTM3U\n#EXT-X-TARGETDURATION:2\n")
	for sequence := 0; sequence <= last; sequence++ {
		fmt.Fprintf(&text, "#EXTINF:2.000,\nseg/%d.ts?token=1\n", sequence)
	}
	playlistURL := fixture.live.server.URL + "/live/high.m3u8?token=1"
	base, _ := url.Parse(playlistURL)
	playlist, err := parseMediaPlaylist(text.String(), base)
	if err != nil {
		t.Fatal(err)
	}
	source := livebuffer.Source{PlaylistURL: playlistURL, Kind: livebuffer.KindVideo, Fetch: fixture.recorder.fetcher.Fetch}
	window := fixture.recorder.buffer.Observe(fixture.roomID, source, playlist, true)
	for _, segment := range window.Segments {
		if _, err := fixture.recorder.buffer.Read(context.Background(), fixture.roomID, segment.Name); err != nil {
			t.Fatal(err)
		}
	}
}

// recordedSegment reads a stored video segment of a recording.
func (fixture *recorderFixture) recordedSegment(t *testing.T, recordID string, number int) string {
	t.Helper()
	_, identifier, _ := strings.Cut(recordID, ":")
	reader, err := fixture.blobs.Get(context.Background(), fmt.Sprintf("%s%s/seg-%05d.ts", RecordingPrefix, identifier, number))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	data, _ := io.ReadAll(reader)
	return string(data)
}

// recordFromPast starts a recording secondsAgo in the past, lets it run until
// it holds the footage and returns it finished.
func (fixture *recorderFixture) recordFromPast(t *testing.T, quality string, secondsAgo int, seconds float64) (recordID string, qualityLabel string) {
	t.Helper()
	ctx := context.Background()
	if _, err := fixture.recorder.StartFrom(ctx, fixture.roomID, quality, secondsAgo); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the footage", func() bool { return fixture.recordedSeconds() >= seconds })
	finished, err := fixture.recorder.Stop(ctx, fixture.roomID)
	if err != nil || finished == nil || finished.Quality == nil {
		t.Fatalf("stop: %+v %v", finished, err)
	}
	return finished.ID, *finished.Quality
}

func TestRecordingStartsInTheBufferedPastAndGoesLive(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{6, 7, 8}, {7, 8, 9}, {8, 9, 10}, {9, 10, 11}})
	fixture.bufferSegments(t, 7)

	recordID, quality := fixture.recordFromPast(t, "480", 6, 14)

	for number, want := range []string{"data-5", "data-6", "data-7", "data-8", "data-9"} {
		if got := fixture.recordedSegment(t, recordID, number); got != want {
			t.Errorf("segment %d = %q, want %q", number, got, want)
		}
	}
	if quality != "720p" {
		t.Errorf("recorded %s, want the buffered 720p rather than the requested 480", quality)
	}
	if playlist := fixture.playlistOf(t, recordID); strings.Contains(playlist, "#EXT-X-DISCONTINUITY") {
		t.Errorf("seamless hand-over from the buffer to live was marked as a break:\n%s", playlist)
	}
}

func TestRecordingFromPastIsCappedToTheBuffer(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{6, 7, 8}, {7, 8, 9}, {8, 9, 10}})
	fixture.bufferSegments(t, 7)

	recordID, _ := fixture.recordFromPast(t, QualitySource, 3600, 22)

	if got := fixture.recordedSegment(t, recordID, 0); got != "data-0" {
		t.Errorf("first segment = %q, want the oldest buffered one", got)
	}
	if got := fixture.recordedSegment(t, recordID, 8); got != "data-8" {
		t.Errorf("segment after the buffered ones = %q, want the live data-8", got)
	}
}

func TestRecordingFromPastWithoutABufferStartsLive(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{6, 7, 8}, {7, 8, 9}, {8, 9, 10}})

	recordID, _ := fixture.recordFromPast(t, QualitySource, 120, 6)

	if got := fixture.recordedSegment(t, recordID, 0); got != "data-6" {
		t.Errorf("first segment = %q, want the first live one", got)
	}
}

func TestRecordingMarksTheGapBetweenBufferAndLive(t *testing.T) {
	fixture := newRecorderFixture(t, [][]int{{6, 7, 8}, {7, 8, 9}, {8, 9, 10}})
	fixture.bufferSegments(t, 3)

	recordID, _ := fixture.recordFromPast(t, QualitySource, 4, 10)

	playlist := fixture.playlistOf(t, recordID)
	if !strings.Contains(playlist, "#EXT-X-DISCONTINUITY\n#EXTINF:2.000,\nseg-00002.ts") {
		t.Errorf("the missing footage between the buffer and live is not marked:\n%s", playlist)
	}
	if got := fixture.recordedSegment(t, recordID, 2); got != "data-6" {
		t.Errorf("first live segment = %q, want data-6", got)
	}
}
