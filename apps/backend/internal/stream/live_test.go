package stream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// livePlaylist is a low-latency live media playlist like a cam site serves.
const livePlaylist = `#EXTM3U
#EXT-X-VERSION:6
#EXT-X-SERVER-CONTROL:CAN-BLOCK-RELOAD=YES,PART-HOLD-BACK=2.4
#EXT-X-PART-INF:PART-TARGET=0.8
#EXT-X-MEDIA-SEQUENCE:10
#EXTINF:1.6,
seg_10.m4s
#EXT-X-PART:DURATION=0.8,URI="part_11_0.m4s"
#EXT-X-PRELOAD-HINT:TYPE=PART,URI="part_11_1.m4s"
#EXT-X-RENDITION-REPORT:URI="other.m3u8",LAST-MSN=11
`

// TestStripLowLatencyTags checks that only the low-latency lines go.
func TestStripLowLatencyTags(t *testing.T) {
	stripped := stripLowLatencyTags(livePlaylist)
	for _, tag := range []string{"#EXT-X-PART", "PRELOAD-HINT", "RENDITION-REPORT", "SERVER-CONTROL"} {
		if strings.Contains(stripped, tag) {
			t.Errorf("playlist still contains %s:\n%s", tag, stripped)
		}
	}
	if !strings.Contains(stripped, "#EXTINF:1.6,\nseg_10.m4s") {
		t.Errorf("segment lost:\n%s", stripped)
	}
}

// TestLiveStreamBypassesCache plays a live session through the proxy and checks
// that the segment arrives while nothing lands in the blob store or kv cache.
func TestLiveStreamBypassesCache(t *testing.T) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			_, _ = w.Write([]byte(livePlaylist))
			return
		}
		_, _ = w.Write([]byte("segment-bytes"))
	}))
	defer origin.Close()

	cacheService := newTestCache(t)
	proxy := NewProxy(cacheService.service, cacheService.service.kv, "http://backend.test")
	manifestURL, err := proxy.LiveManifestURL(context.Background(), &plugins.ResolveResult{URL: origin.URL + "/live/index.m3u8"})
	if err != nil {
		t.Fatal(err)
	}

	manifestResponse := httptest.NewRecorder()
	proxy.ServeHTTP(manifestResponse, httptest.NewRequest(http.MethodGet, strings.TrimPrefix(manifestURL, "http://backend.test"), nil))
	if strings.Contains(manifestResponse.Body.String(), "PART") {
		t.Fatalf("low-latency tags survived:\n%s", manifestResponse.Body)
	}
	var segmentPath string
	for _, line := range strings.Split(manifestResponse.Body.String(), "\n") {
		if strings.HasPrefix(line, "/api/stream/seg?") {
			segmentPath = line
		}
	}
	if segmentPath == "" {
		t.Fatalf("no proxied segment in:\n%s", manifestResponse.Body)
	}

	segmentResponse := httptest.NewRecorder()
	proxy.ServeHTTP(segmentResponse, httptest.NewRequest(http.MethodGet, segmentPath, nil))
	if segmentResponse.Code != http.StatusOK || segmentResponse.Body.String() != "segment-bytes" {
		t.Fatalf("segment = %d %q, want the origin bytes", segmentResponse.Code, segmentResponse.Body)
	}

	cached, err := cacheService.service.kv.Entries(context.Background(), kvSegmentPrefix)
	if err != nil || len(cached) != 0 {
		t.Errorf("segment kv entries = %v (err %v), want none", cached, err)
	}
	filepath.WalkDir(cacheService.root, func(path string, entry os.DirEntry, err error) error {
		if err == nil && !entry.IsDir() && strings.Contains(path, blobPrefix) {
			t.Errorf("live stream wrote a blob: %s", path)
		}
		return nil
	})
}
