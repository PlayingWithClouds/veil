package stream

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/playingwithclouds/veil/internal/livebuffer"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// TestLiveRoomPlaylistCoversTheBufferedWindow plays a room through the proxy
// while the origin's playlist slides, and checks that the player's playlist
// keeps growing past the origin's window and that segments come from the
// buffer, fetched from the origin once.
func TestLiveRoomPlaylistCoversTheBufferedWindow(t *testing.T) {
	var mutex sync.Mutex
	first := 0
	segmentFetches := map[string]int{}
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			fmt.Fprintf(w, "#EXTM3U\n#EXT-X-TARGETDURATION:2\n#EXT-X-MEDIA-SEQUENCE:%d\n", first)
			for sequence := first; sequence < first+3; sequence++ {
				fmt.Fprintf(w, "#EXTINF:2.0,\nseg%d.ts?token=t\n", sequence)
			}
			return
		}
		segmentFetches[r.URL.Path]++
		fmt.Fprint(w, "bytes"+r.URL.Path)
	}))
	defer origin.Close()

	cacheService := newTestCache(t)
	proxy := NewProxy(cacheService.service, cacheService.service.kv, "http://backend.test")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	proxy.SetLiveBuffer(livebuffer.New(ctx, cacheService.service.storage))
	manifestURL, err := proxy.LiveManifestURL(ctx, &plugins.ResolveResult{URL: origin.URL + "/live/index.m3u8"}, "nova")
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := strings.TrimPrefix(manifestURL, "http://backend.test")

	var playlist string
	for _, start := range []int{0, 2, 4} {
		mutex.Lock()
		first = start
		mutex.Unlock()
		response := httptest.NewRecorder()
		proxy.ServeHTTP(response, httptest.NewRequest(http.MethodGet, manifestPath, nil))
		playlist = response.Body.String()
	}
	var segmentPaths []string
	for _, line := range strings.Split(playlist, "\n") {
		if strings.HasPrefix(line, "/api/stream/seg?") {
			segmentPaths = append(segmentPaths, line)
		}
	}
	if len(segmentPaths) != 7 {
		t.Fatalf("player playlist lists %d segments, want the 7 seen so far:\n%s", len(segmentPaths), playlist)
	}
	if !strings.Contains(playlist, "#EXT-X-MEDIA-SEQUENCE:0") || strings.Contains(playlist, "ENDLIST") {
		t.Errorf("playlist is not an open window from the start:\n%s", playlist)
	}
	for round := 0; round < 2; round++ {
		response := httptest.NewRecorder()
		proxy.ServeHTTP(response, httptest.NewRequest(http.MethodGet, segmentPaths[0], nil))
		if response.Body.String() != "bytes/live/seg0.ts" {
			t.Fatalf("oldest segment = %q", response.Body)
		}
	}
	mutex.Lock()
	defer mutex.Unlock()
	if segmentFetches["/live/seg0.ts"] != 1 {
		t.Errorf("origin served seg0 %d times, want once", segmentFetches["/live/seg0.ts"])
	}
}
