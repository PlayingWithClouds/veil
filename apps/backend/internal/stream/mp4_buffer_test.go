package stream

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// countingOrigin serves content with range support (or without, when
// ranges is false) and counts the bytes it sent.
func countingOrigin(t *testing.T, content []byte, ranges bool) (*httptest.Server, *atomic.Int64) {
	t.Helper()
	var sent atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !ranges {
			request.Header.Del("Range")
		}
		counter := &countingWriter{ResponseWriter: writer, sent: &sent}
		http.ServeContent(counter, request, "video.mp4", time.Time{}, bytes.NewReader(content))
	}))
	t.Cleanup(server.Close)
	return server, &sent
}

// countingWriter counts the body bytes written through it.
type countingWriter struct {
	http.ResponseWriter
	sent *atomic.Int64
}

// Write counts and passes on.
func (writer *countingWriter) Write(data []byte) (int, error) {
	written, err := writer.ResponseWriter.Write(data)
	writer.sent.Add(int64(written))
	return written, err
}

// waitForFetches waits until the buffer has no chunk download running.
func waitForFetches(t *testing.T, buffer *mp4Buffer) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		buffer.mutex.Lock()
		running := len(buffer.fetching)
		buffer.mutex.Unlock()
		if running == 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("chunk fetches never finished")
}

// TestMP4BufferServesRangesAndReadsAheadOnlyAFewChunks checks that ranges are
// served right, and that the origin sends the chunks ahead of the player,
// not the whole file.
func TestMP4BufferServesRangesAndReadsAheadOnlyAFewChunks(t *testing.T) {
	content := make([]byte, 60*mp4ChunkSize)
	for index := range content {
		content[index] = byte(index * 7)
	}
	origin, sent := countingOrigin(t, content, true)
	buffer := newMP4Buffer(newTestCache(t).service)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/stream/mp4", nil)
	request.Header.Set("Range", "bytes=0-1023")
	if !buffer.serve(response, request, origin.URL, nil) {
		t.Fatal("buffer turned the file down")
	}
	if response.Code != http.StatusPartialContent || !bytes.Equal(response.Body.Bytes(), content[:1024]) {
		t.Fatalf("first range: %d, %d bytes", response.Code, response.Body.Len())
	}
	if got := response.Header().Get("Content-Range"); got != "bytes 0-1023/62914560" {
		t.Errorf("Content-Range = %q", got)
	}
	waitForFetches(t, buffer)
	if fetched := sent.Load(); fetched > (mp4ReadAhead+1)*mp4ChunkSize {
		t.Errorf("origin sent %d bytes, more than the chunks ahead", fetched)
	}

	middle := int64(30*mp4ChunkSize - 50)
	response = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/stream/mp4", nil)
	request.Header.Set("Range", "bytes=31457230-31457329")
	buffer.serve(response, request, origin.URL, nil)
	if !bytes.Equal(response.Body.Bytes(), content[middle:middle+100]) {
		t.Errorf("a range across two chunks came back wrong (%d bytes)", response.Body.Len())
	}
	waitForFetches(t, buffer)
	if fetched := sent.Load(); fetched >= int64(len(content)) {
		t.Errorf("origin sent the whole file (%d bytes)", fetched)
	}
}

// TestMP4BufferLeavesOriginsWithoutRangesToTheRelay checks that a file whose
// origin ignores range requests isn't buffered.
func TestMP4BufferLeavesOriginsWithoutRangesToTheRelay(t *testing.T) {
	origin, _ := countingOrigin(t, make([]byte, 3*mp4ChunkSize), false)
	buffer := newMP4Buffer(newTestCache(t).service)
	request := httptest.NewRequest(http.MethodGet, "/api/stream/mp4", nil)
	request.Header.Set("Range", "bytes=0-")
	if buffer.serve(httptest.NewRecorder(), request, origin.URL, nil) {
		t.Error("buffered an origin that ignores ranges")
	}
}

// TestParseRange checks the Range forms players send.
func TestParseRange(t *testing.T) {
	cases := []struct {
		header     string
		start, end int64
		ok         bool
	}{
		{"", 0, 999, true},
		{"bytes=0-", 0, 999, true},
		{"bytes=100-199", 100, 199, true},
		{"bytes=900-5000", 900, 999, true},
		{"bytes=-100", 900, 999, true},
		{"bytes=1000-", 0, 0, false},
		{"items=0-1", 0, 0, false},
	}
	for _, testCase := range cases {
		start, end, ok := parseRange(testCase.header, 1000)
		if ok != testCase.ok || (ok && (start != testCase.start || end != testCase.end)) {
			t.Errorf("%q = %d-%d %v", testCase.header, start, end, ok)
		}
	}
}
