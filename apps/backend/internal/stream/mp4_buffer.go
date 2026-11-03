package stream

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// mp4ChunkSize is the piece of a file fetched, stored and served at once.
	mp4ChunkSize = 1 << 20
	// mp4ReadAhead is how many chunks past the one being served are fetched
	// ahead, so the buffer runs at most this far ahead of the player.
	mp4ReadAhead = 24
	// mp4AheadWorkers bounds the read-ahead fetches running at once across
	// all files; the chunk a player waits for doesn't wait for them.
	mp4AheadWorkers = 4
	// mp4FetchTimeout bounds fetching one chunk.
	mp4FetchTimeout = 60 * time.Second
)

// errNoRanges is returned for an origin that answers a range request with
// the whole file: it can't be buffered in chunks.
var errNoRanges = errors.New("origin ignores range requests")

// mp4File is what the buffer knows of a file from its first chunk.
type mp4File struct {
	size        int64
	contentType string
}

// chunkFetch is a chunk download in progress; done closes when it ends.
type chunkFetch struct {
	done chan struct{}
	err  error
}

// mp4Buffer buffers the MP4 files being played: it fetches them from the
// origin in chunks with range requests, keeps the chunks in the stream cache
// (evicted like the rest of it) and serves them from there. The chunks after
// the one being served are fetched ahead, so playback has a buffer without
// the whole file being downloaded.
type mp4Buffer struct {
	cache *CacheService
	ahead chan struct{}

	mutex    sync.Mutex
	files    map[string]mp4File
	fetching map[string]*chunkFetch
	// relayed lists the files whose origin ignores range requests.
	relayed map[string]bool
}

// newMP4Buffer returns a buffer storing its chunks through cache.
func newMP4Buffer(cache *CacheService) *mp4Buffer {
	return &mp4Buffer{
		cache:    cache,
		ahead:    make(chan struct{}, mp4AheadWorkers),
		files:    map[string]mp4File{},
		fetching: map[string]*chunkFetch{},
		relayed:  map[string]bool{},
	}
}

// serve answers a player's request for a file, the range it asks for or all
// of it, from the buffer. It reports false when the origin can't be buffered
// (it ignores range requests or the first chunk failed); nothing was written
// then, so the caller relays the file instead.
func (b *mp4Buffer) serve(w http.ResponseWriter, r *http.Request, mediaURL string, headers map[string]string) bool {
	file, err := b.file(r.Context(), mediaURL, headers)
	if err != nil {
		if !errors.Is(err, errNoRanges) {
			log.Printf("mp4 buffer %s: %v", mediaURL, err)
		}
		return false
	}
	start, end, ok := parseRange(r.Header.Get("Range"), file.size)
	if !ok {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", file.size))
		http.Error(w, "range not satisfiable", http.StatusRequestedRangeNotSatisfiable)
		return true
	}
	w.Header().Set("Content-Type", file.contentType)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", strconv.FormatInt(end-start+1, 10))
	status := http.StatusOK
	if r.Header.Get("Range") != "" {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, file.size))
		status = http.StatusPartialContent
	}
	w.WriteHeader(status)
	b.write(w, r.Context(), mediaURL, headers, file, start, end)
	return true
}

// write sends bytes start..end of the file chunk by chunk, fetching ahead
// of each chunk, until it's sent or the player goes away.
func (b *mp4Buffer) write(w io.Writer, ctx context.Context, mediaURL string, headers map[string]string, file mp4File, start, end int64) {
	lastChunk := (file.size - 1) / mp4ChunkSize
	for index := start / mp4ChunkSize; index <= end/mp4ChunkSize; index++ {
		b.fetchAhead(mediaURL, headers, index+1, min(index+mp4ReadAhead, lastChunk))
		data, err := b.chunk(ctx, mediaURL, headers, index)
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("mp4 buffer %s chunk %d: %v", mediaURL, index, err)
			}
			return
		}
		chunkStart := index * mp4ChunkSize
		from := max(start, chunkStart) - chunkStart
		to := min(end+1, chunkStart+int64(len(data))) - chunkStart
		if from >= to {
			return
		}
		if _, err := w.Write(data[from:to]); err != nil {
			return
		}
	}
}

// file returns what's known of the file, fetching its first chunk to learn
// its size the first time. An origin found to ignore ranges isn't asked again.
func (b *mp4Buffer) file(ctx context.Context, mediaURL string, headers map[string]string) (mp4File, error) {
	b.mutex.Lock()
	file, known := b.files[mediaURL]
	relayed := b.relayed[mediaURL]
	b.mutex.Unlock()
	if known {
		return file, nil
	}
	if relayed {
		return mp4File{}, errNoRanges
	}
	_, err := b.chunk(ctx, mediaURL, headers, 0)
	if errors.Is(err, errNoRanges) {
		b.mutex.Lock()
		b.relayed[mediaURL] = true
		b.mutex.Unlock()
	}
	if err != nil {
		return mp4File{}, err
	}
	b.mutex.Lock()
	file, known = b.files[mediaURL]
	b.mutex.Unlock()
	if !known {
		return mp4File{}, errors.New("the first chunk named no file size")
	}
	return file, nil
}

// chunk returns one chunk: from the stream cache, else fetched from the
// origin (joining a fetch already running) and stored.
func (b *mp4Buffer) chunk(ctx context.Context, mediaURL string, headers map[string]string, index int64) ([]byte, error) {
	key := chunkKey(mediaURL, index)
	if data, ok := b.stored(ctx, key); ok && b.known(mediaURL) {
		return data, nil
	}
	fetch := b.startFetch(mediaURL, headers, index)
	select {
	case <-fetch.done:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	if fetch.err != nil {
		return nil, fetch.err
	}
	data, ok := b.stored(ctx, key)
	if !ok {
		return nil, errors.New("fetched chunk is gone")
	}
	return data, nil
}

// known reports whether the file's size is known, so its chunks are complete.
func (b *mp4Buffer) known(mediaURL string) bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	_, known := b.files[mediaURL]
	return known
}

// stored reads a chunk from the stream cache, marking it used.
func (b *mp4Buffer) stored(ctx context.Context, key string) ([]byte, bool) {
	if b.cache.storage.Touch(ctx, key) != nil {
		return nil, false
	}
	reader, err := b.cache.storage.Get(ctx, key)
	if err != nil {
		return nil, false
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, false
	}
	return data, true
}

// fetchAhead starts fetching the chunks from first to last that aren't
// stored or being fetched, a few at a time.
func (b *mp4Buffer) fetchAhead(mediaURL string, headers map[string]string, first, last int64) {
	for index := first; index <= last; index++ {
		if b.cache.storage.Exists(context.Background(), chunkKey(mediaURL, index)) {
			continue
		}
		b.mutex.Lock()
		_, running := b.fetching[chunkKey(mediaURL, index)]
		b.mutex.Unlock()
		if running {
			continue
		}
		select {
		case b.ahead <- struct{}{}:
		default:
			// Every worker is busy; the next chunk served asks again.
			return
		}
		fetch := b.startFetch(mediaURL, headers, index)
		go func() {
			<-fetch.done
			<-b.ahead
		}()
	}
}

// startFetch returns the running fetch of a chunk, starting one when there
// is none. It runs on its own, so a player leaving doesn't waste it.
func (b *mp4Buffer) startFetch(mediaURL string, headers map[string]string, index int64) *chunkFetch {
	key := chunkKey(mediaURL, index)
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if fetch, running := b.fetching[key]; running {
		return fetch
	}
	fetch := &chunkFetch{done: make(chan struct{})}
	b.fetching[key] = fetch
	go func() {
		fetch.err = b.fetch(mediaURL, headers, index)
		b.mutex.Lock()
		delete(b.fetching, key)
		b.mutex.Unlock()
		close(fetch.done)
	}()
	return fetch
}

// fetch downloads one chunk with a range request and stores it, noting the
// file's size and type from the answer.
func (b *mp4Buffer) fetch(mediaURL string, headers map[string]string, index int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), mp4FetchTimeout)
	defer cancel()
	defer b.cache.beginOperation()()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, mediaURL, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	for name, value := range headers {
		request.Header.Set(name, value)
	}
	start := index * mp4ChunkSize
	request.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, start+mp4ChunkSize-1))
	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		return errNoRanges
	}
	if response.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("HTTP %d", response.StatusCode)
	}
	size, ok := totalSize(response.Header.Get("Content-Range"))
	if !ok {
		return errNoRanges
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, mp4ChunkSize))
	if err != nil {
		return err
	}
	if err := b.cache.storage.Put(ctx, chunkKey(mediaURL, index), bytes.NewReader(data), int64(len(data)), "video/mp4"); err != nil {
		return err
	}
	b.remember(mediaURL, size, response.Header.Get("Content-Type"))
	return nil
}

// remember notes a file's size and type.
func (b *mp4Buffer) remember(mediaURL string, size int64, contentType string) {
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = "video/mp4"
	}
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.files[mediaURL] = mp4File{size: size, contentType: contentType}
}

// chunkKey is the blob key one chunk of a file is kept under.
func chunkKey(mediaURL string, index int64) string {
	return fmt.Sprintf("%smp4-chunks/%s/%d", blobPrefix, urlHash(mediaURL), index)
}

// totalSize reads the file size from a Content-Range header ("bytes 0-99/1234").
func totalSize(contentRange string) (int64, bool) {
	slash := strings.LastIndex(contentRange, "/")
	if slash == -1 {
		return 0, false
	}
	size, err := strconv.ParseInt(contentRange[slash+1:], 10, 64)
	if err != nil || size <= 0 {
		return 0, false
	}
	return size, true
}

// parseRange reads a Range header against a file of size bytes and returns
// the first and last byte asked for: all of it without a header. Only the
// first range of several is served.
func parseRange(header string, size int64) (start, end int64, ok bool) {
	if header == "" {
		return 0, size - 1, true
	}
	spec, found := strings.CutPrefix(header, "bytes=")
	if !found {
		return 0, 0, false
	}
	spec, _, _ = strings.Cut(spec, ",")
	first, last, found := strings.Cut(strings.TrimSpace(spec), "-")
	if !found {
		return 0, 0, false
	}
	if first == "" {
		return suffixRange(last, size)
	}
	start, err := strconv.ParseInt(first, 10, 64)
	if err != nil || start >= size {
		return 0, 0, false
	}
	end = size - 1
	if last != "" {
		parsed, err := strconv.ParseInt(last, 10, 64)
		if err != nil || parsed < start {
			return 0, 0, false
		}
		end = min(parsed, size-1)
	}
	return start, end, true
}

// suffixRange is the range of the last count bytes ("bytes=-500").
func suffixRange(count string, size int64) (start, end int64, ok bool) {
	length, err := strconv.ParseInt(count, 10, 64)
	if err != nil || length <= 0 {
		return 0, 0, false
	}
	return max(0, size-length), size - 1, true
}
