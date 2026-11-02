package cams

import (
	"net/url"
	"strings"
	"testing"
)

// TestChooseVariantPicksTheHighestBandwidthForSource checks the source pick
// and that a media playlist has no variants.
func TestChooseVariantPicksTheHighestBandwidthForSource(t *testing.T) {
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

// TestParseVariantsFindsTheSeparateAudioRendition checks audio groups.
func TestParseVariantsFindsTheSeparateAudioRendition(t *testing.T) {
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

// TestChooseVariant checks the pick for each target height.
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
	unknown := []streamVariant{{URL: "a", Bandwidth: 1}, {URL: "b", Bandwidth: 9}}
	if got := chooseVariant(unknown, 480); got.URL != "b" {
		t.Errorf("without heights the best bitrate wins, got %+v", got)
	}
	mixed := []streamVariant{{URL: "unknown", Bandwidth: 9}, {URL: "360", Height: 360, Bandwidth: 1}}
	if got := chooseVariant(mixed, 480); got.URL != "360" {
		t.Errorf("a rendition without a height must not win once others have one, got %+v", got)
	}
}

// TestParseQuality checks accepted and rejected qualities.
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
