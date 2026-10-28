// Package livebuffer keeps the last minutes of live HLS streams (cam rooms) on
// disk so a viewer can rewind and a recording can start in the past. Only
// metadata lives in memory; segment bytes are blobs under Prefix.
package livebuffer

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	// Prefix is the blob key prefix buffers are stored under. It is outside the
	// cache prefixes, so the cache evictor neither counts nor deletes them.
	Prefix = "live-buffer/"

	// WindowSeconds is how much footage a rendition keeps.
	WindowSeconds = 180.0
	// IdleTimeout is how long a room's buffer survives without a viewer or a
	// recording touching it.
	IdleTimeout = 2 * time.Minute

	// sweepInterval is how often idle rooms and renditions are looked for.
	sweepInterval = 30 * time.Second
	// downloadTimeout bounds fetching one segment.
	downloadTimeout = 30 * time.Second
	// downloadWorkers bounds segment downloads across all rooms.
	downloadWorkers = 8

	// KindVideo and KindAudio tell a rendition's playlist apart: a stream may
	// carry its audio in a separate playlist.
	KindVideo = "video"
	KindAudio = "audio"
)

// Store is the part of the blob store the buffer writes to.
type Store interface {
	Put(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	DeletePrefix(ctx context.Context, prefix string) error
}

// Fetcher downloads a URL with the headers a stream needs. A response with a
// status other than 200 is an error.
type Fetcher func(ctx context.Context, rawURL string, headers map[string]string) (*http.Response, error)

// Manager owns the buffers of all rooms.
type Manager struct {
	store       Store
	baseContext context.Context
	// now, window and idleTimeout are fields so tests can control them.
	now         func() time.Time
	window      float64
	idleTimeout time.Duration

	mutex sync.Mutex
	rooms map[string]*room

	// queue holds the downloads waiting for a worker, the newest last: workers
	// take from the end, so the live edge a player is waiting for comes before
	// the backlog of older segments. queueChanged wakes idle workers.
	queueMutex   sync.Mutex
	queueChanged *sync.Cond
	queue        []queuedDownload
	workersBegun bool
}

// queuedDownload is a download waiting for a worker.
type queuedDownload struct {
	room *room
	pendingDownload
}

// New returns a Manager; ctx is the server's lifetime and ends downloads.
func New(ctx context.Context, store Store) *Manager {
	manager := &Manager{
		store: store, baseContext: ctx, now: time.Now, window: WindowSeconds, idleTimeout: IdleTimeout,
		rooms: map[string]*room{},
	}
	manager.queueChanged = sync.NewCond(&manager.queueMutex)
	return manager
}

// room is the buffer of one cam room.
type room struct {
	identifier  string
	lastTouched time.Time
	tracks      map[string]*track
	// files maps a blob name (relative to the room directory) to its state.
	files map[string]*file
}

// track is one rendition of a room: a playlist whose segments are buffered.
type track struct {
	key                    string
	kind                   string
	entries                []*entry
	known                  map[string]*entry
	retired                map[string]bool
	inits                  map[string]*file
	initCounter            int
	// lastSequence is the sequence number of the newest segment, -1 when empty.
	lastSequence   int
	lastIdentity   string
	targetDuration float64
	ended          bool
	lastObserved   time.Time
}

// entry is a buffered segment.
type entry struct {
	sequence  int
	identity  string
	sourceURL string
	duration  float64
	// discontinuityNumber is the origin's break count up to the segment
	// (Segment.DiscontinuityNumber), alike across a stream's renditions.
	discontinuityNumber int
	file                *file
	initFile            *file
}

// file is a blob being downloaded or already stored.
type file struct {
	name        string
	done        chan struct{}
	failed      bool
	removed     bool
	contentType string
}

// directory is the blob key prefix of the room's files.
func (buffered *room) directory() string {
	return Prefix + buffered.identifier + "/"
}

// TrackKey names a rendition by its playlist URL: the host and token are left
// out so a reconnect lands in the same rendition.
func TrackKey(playlistURL string) string {
	sum := sha1.Sum([]byte(URLIdentity(playlistURL)))
	return hex.EncodeToString(sum[:])[:10]
}

// roomIdentifier turns a record id ("cam_model:x") into a directory name.
func roomIdentifier(roomID string) string {
	_, identifier, found := strings.Cut(roomID, ":")
	if !found {
		identifier = roomID
	}
	return identifier
}

// Reset deletes every buffer on disk. Run at startup: buffers do not survive a
// restart, so what is left is garbage.
func (m *Manager) Reset(ctx context.Context) error {
	return m.store.DeletePrefix(ctx, Prefix)
}

// Run sweeps idle buffers until ctx ends.
func (m *Manager) Run(ctx context.Context) {
	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.Sweep()
		}
	}
}

// Sweep drops rooms nobody touched for the idle timeout and renditions that
// were not observed for as long.
func (m *Manager) Sweep() {
	now := m.now()
	var dropped []*room
	var removedFiles []removedFile
	m.mutex.Lock()
	for identifier, buffered := range m.rooms {
		if now.Sub(buffered.lastTouched) > m.idleTimeout {
			dropped = append(dropped, buffered)
			delete(m.rooms, identifier)
			continue
		}
		for key, rendition := range buffered.tracks {
			if now.Sub(rendition.lastObserved) > m.idleTimeout {
				removedFiles = append(removedFiles, buffered.dropTrack(key)...)
			}
		}
	}
	for _, buffered := range dropped {
		for _, stored := range buffered.files {
			stored.removed = true
		}
	}
	m.mutex.Unlock()
	for _, buffered := range dropped {
		if err := m.store.DeletePrefix(m.baseContext, buffered.directory()); err != nil {
			log.Printf("livebuffer: drop %s: %v", buffered.identifier, err)
		}
	}
	m.deleteFiles(removedFiles)
}

// removedFile is a blob to delete once the lock is released.
type removedFile struct {
	key string
}

// deleteFiles deletes blobs that left their buffer.
func (m *Manager) deleteFiles(files []removedFile) {
	for _, removed := range files {
		if err := m.store.Delete(m.baseContext, removed.key); err != nil {
			log.Printf("livebuffer: delete %s: %v", removed.key, err)
		}
	}
}

// dropTrack forgets a rendition and returns its blobs. The caller holds the lock.
func (buffered *room) dropTrack(key string) []removedFile {
	rendition := buffered.tracks[key]
	delete(buffered.tracks, key)
	var removed []removedFile
	for _, segment := range rendition.entries {
		removed = append(removed, buffered.forget(segment.file)...)
	}
	for _, initFile := range rendition.inits {
		removed = append(removed, buffered.forget(initFile)...)
	}
	return removed
}

// forget marks a file as gone and lists its blob for deletion. The caller holds the lock.
func (buffered *room) forget(stored *file) []removedFile {
	stored.removed = true
	delete(buffered.files, stored.name)
	return []removedFile{{key: buffered.directory() + stored.name}}
}

// Touch marks a room as watched, keeping its buffer from idling out.
func (m *Manager) Touch(roomID string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if buffered := m.rooms[roomIdentifier(roomID)]; buffered != nil {
		buffered.lastTouched = m.now()
	}
}

// Has reports whether a room currently has a buffer.
func (m *Manager) Has(roomID string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.rooms[roomIdentifier(roomID)] != nil
}

// Source says where a playlist came from and how to download its segments.
type Source struct {
	// PlaylistURL is the rendition's media playlist URL.
	PlaylistURL string
	Kind        string
	Fetch       Fetcher
	Headers     map[string]string
}

// Observe records a fresh upstream playlist of a rendition: segments not seen
// before start downloading, old ones leave the window. create makes a buffer
// for a room that has none (a viewer); without it a room nobody watches stays
// unbuffered and Observe returns nil.
func (m *Manager) Observe(roomID string, source Source, playlist *Playlist, create bool) *Window {
	now := m.now()
	var removedFiles []removedFile
	m.mutex.Lock()
	buffered := m.rooms[roomIdentifier(roomID)]
	if buffered == nil {
		if !create {
			m.mutex.Unlock()
			return nil
		}
		buffered = &room{identifier: roomIdentifier(roomID), tracks: map[string]*track{}, files: map[string]*file{}}
		m.rooms[buffered.identifier] = buffered
	}
	buffered.lastTouched = now
	rendition := buffered.renditionFor(source, now)
	if rendition.missed(playlist) {
		removedFiles = buffered.restart(rendition)
	}
	downloads := m.addNewSegments(buffered, rendition, source, playlist)
	removedFiles = append(removedFiles, m.trim(buffered, rendition)...)
	window := rendition.window()
	m.mutex.Unlock()

	m.deleteFiles(removedFiles)
	m.enqueue(buffered, downloads)
	return window
}

// renditionFor returns the track of the source, creating it. The caller holds the lock.
func (buffered *room) renditionFor(source Source, now time.Time) *track {
	key := TrackKey(source.PlaylistURL)
	rendition := buffered.tracks[key]
	if rendition == nil {
		rendition = &track{
			key: key, known: map[string]*entry{}, retired: map[string]bool{}, inits: map[string]*file{}, lastSequence: -1,
		}
		buffered.tracks[key] = rendition
	}
	if source.Kind != "" {
		rendition.kind = source.Kind
	}
	rendition.lastObserved = now
	return rendition
}

// pendingDownload is one file to fetch.
type pendingDownload struct {
	target  *file
	url     string
	headers map[string]string
	fetch   Fetcher
}

// addNewSegments adds the playlist's unseen segments to the rendition and
// returns the downloads to start. The caller holds the lock.
func (m *Manager) addNewSegments(buffered *room, rendition *track, source Source, playlist *Playlist) []pendingDownload {
	rendition.ended = playlist.Ended
	rendition.targetDuration = playlist.TargetDuration
	rendition.pruneRetired(playlist)
	var downloads []pendingDownload
	for index, segment := range playlist.Segments {
		if rendition.known[segment.Identity] != nil || rendition.retired[segment.Identity] {
			continue
		}
		if segment.Duration <= 0 {
			segment.Duration = playlist.TargetDuration
		}
		added, initDownload := buffered.addSegment(rendition, segment, playlist.MediaSequence+index)
		downloads = append(downloads, pendingDownload{target: added.file, url: segment.URL, headers: source.Headers, fetch: source.Fetch})
		if initDownload != nil {
			initDownload.headers, initDownload.fetch = source.Headers, source.Fetch
			downloads = append(downloads, *initDownload)
		}
		rendition.lastIdentity = segment.Identity
	}
	return downloads
}

// missed reports whether the rendition lost track of the stream: its newest
// segment has left the origin's playlist, so segments in between were never
// seen (nobody polled it for a while, or the stream reconnected).
func (rendition *track) missed(playlist *Playlist) bool {
	return rendition.lastIdentity != "" && !playlistHas(playlist, rendition.lastIdentity)
}

// restart empties a rendition that missed segments, so it begins again at the
// live edge. Keeping the older part would leave a hole the player cannot
// cross, and marking it with a break of our own would put the rendition out
// of step with its audio or video counterpart, which players match by the
// origin's break count. It returns the blobs to delete. The caller holds the lock.
func (buffered *room) restart(rendition *track) []removedFile {
	var removed []removedFile
	for _, segment := range rendition.entries {
		removed = append(removed, buffered.forget(segment.file)...)
	}
	rendition.entries = nil
	rendition.known = map[string]*entry{}
	rendition.retired = map[string]bool{}
	rendition.lastIdentity = ""
	return append(removed, buffered.dropUnusedInits(rendition)...)
}

// pruneRetired forgets trimmed segments the upstream playlist no longer lists.
func (rendition *track) pruneRetired(playlist *Playlist) {
	if len(rendition.retired) == 0 {
		return
	}
	listed := map[string]bool{}
	for _, segment := range playlist.Segments {
		listed[segment.Identity] = true
	}
	for identity := range rendition.retired {
		if !listed[identity] {
			delete(rendition.retired, identity)
		}
	}
}

// playlistHas reports whether the playlist lists the segment.
func playlistHas(playlist *Playlist, identity string) bool {
	for _, segment := range playlist.Segments {
		if segment.Identity == identity {
			return true
		}
	}
	return false
}

// addSegment appends a segment to the rendition, creating its init section
// file when new. upstreamSequence is the number the origin gave the segment; one
// that does not move forward (the origin restarted its numbering) continues
// after the newest. The caller holds the lock.
func (buffered *room) addSegment(rendition *track, segment Segment, upstreamSequence int) (*entry, *pendingDownload) {
	sequence := upstreamSequence
	if sequence <= rendition.lastSequence {
		sequence = rendition.lastSequence + 1
	}
	rendition.lastSequence = sequence
	added := &entry{
		sequence: sequence, identity: segment.Identity, sourceURL: segment.URL,
		duration: segment.Duration, discontinuityNumber: segment.DiscontinuityNumber,
	}
	extension := segmentExtension(segment.Identity)
	added.file = buffered.newFile(fmt.Sprintf("%s-%06d%s", rendition.key, added.sequence, extension), segmentContentType(extension))
	rendition.entries = append(rendition.entries, added)
	rendition.known[segment.Identity] = added
	if segment.InitURL == "" {
		return added, nil
	}
	initIdentity := URLIdentity(segment.InitURL)
	added.initFile = rendition.inits[initIdentity]
	if added.initFile != nil {
		return added, nil
	}
	added.initFile = buffered.newFile(fmt.Sprintf("%s-init-%d.mp4", rendition.key, rendition.initCounter), "video/mp4")
	rendition.initCounter++
	rendition.inits[initIdentity] = added.initFile
	return added, &pendingDownload{target: added.initFile, url: segment.InitURL}
}

// newFile registers a file that is about to be downloaded. The caller holds the lock.
func (buffered *room) newFile(name, contentType string) *file {
	created := &file{name: name, done: make(chan struct{}), contentType: contentType}
	buffered.files[name] = created
	return created
}

// segmentExtension is the file extension to keep for a segment, defaulting to MPEG-TS.
func segmentExtension(identity string) string {
	extension := strings.ToLower(path.Ext(identity))
	switch extension {
	case ".m4s", ".mp4", ".aac", ".ts":
		return extension
	}
	return ".ts"
}

// segmentContentType is the media type of a segment with the extension.
func segmentContentType(extension string) string {
	switch extension {
	case ".ts":
		return "video/mp2t"
	case ".aac":
		return "audio/aac"
	}
	return "video/mp4"
}

// trim drops the oldest segments while the rendition holds more than the
// window, and init sections nothing uses anymore. It returns the blobs to
// delete. The caller holds the lock.
func (m *Manager) trim(buffered *room, rendition *track) []removedFile {
	total := 0.0
	for _, segment := range rendition.entries {
		total += segment.duration
	}
	var removed []removedFile
	for total > m.window && len(rendition.entries) > 1 {
		oldest := rendition.entries[0]
		rendition.entries = rendition.entries[1:]
		total -= oldest.duration
		delete(rendition.known, oldest.identity)
		rendition.retired[oldest.identity] = true
		removed = append(removed, buffered.forget(oldest.file)...)
	}
	return append(removed, buffered.dropUnusedInits(rendition)...)
}

// dropUnusedInits forgets init sections no buffered segment references. The caller holds the lock.
func (buffered *room) dropUnusedInits(rendition *track) []removedFile {
	used := map[*file]bool{}
	for _, segment := range rendition.entries {
		used[segment.initFile] = true
	}
	var removed []removedFile
	for identity, initFile := range rendition.inits {
		if used[initFile] {
			continue
		}
		delete(rendition.inits, identity)
		removed = append(removed, buffered.forget(initFile)...)
	}
	return removed
}

// enqueue schedules downloads, oldest first, and starts the workers on first use.
func (m *Manager) enqueue(buffered *room, downloads []pendingDownload) {
	if len(downloads) == 0 {
		return
	}
	m.queueMutex.Lock()
	defer m.queueMutex.Unlock()
	for _, pending := range downloads {
		m.queue = append(m.queue, queuedDownload{room: buffered, pendingDownload: pending})
	}
	if !m.workersBegun {
		m.workersBegun = true
		for worker := 0; worker < downloadWorkers; worker++ {
			go m.work()
		}
		go func() {
			<-m.baseContext.Done()
			m.queueChanged.Broadcast()
		}()
	}
	m.queueChanged.Broadcast()
}

// work downloads queued files, newest first, until the server shuts down.
func (m *Manager) work() {
	for {
		next, ok := m.take()
		if !ok {
			return
		}
		ctx, cancel := context.WithTimeout(m.baseContext, downloadTimeout)
		m.finishDownload(next.room, next.target, m.fetchInto(ctx, next.room, next.pendingDownload))
		cancel()
	}
}

// take blocks until a download is queued and removes the newest one. It reports
// false once the server shuts down; downloads still queued then fail.
func (m *Manager) take() (queuedDownload, bool) {
	m.queueMutex.Lock()
	defer m.queueMutex.Unlock()
	for len(m.queue) == 0 && m.baseContext.Err() == nil {
		m.queueChanged.Wait()
	}
	if m.baseContext.Err() != nil {
		for _, abandoned := range m.queue {
			m.finishDownload(abandoned.room, abandoned.target, fmt.Errorf("shutting down"))
		}
		m.queue = nil
		return queuedDownload{}, false
	}
	last := len(m.queue) - 1
	next := m.queue[last]
	m.queue = m.queue[:last]
	return next, true
}

// promote moves a queued download to the front of the line, because a player
// is waiting for it.
func (m *Manager) promote(wanted *file) {
	m.queueMutex.Lock()
	defer m.queueMutex.Unlock()
	for index, queued := range m.queue {
		if queued.target != wanted {
			continue
		}
		m.queue = append(append(m.queue[:index:index], m.queue[index+1:]...), queued)
		return
	}
}

// fetchInto downloads the file and stores it.
func (m *Manager) fetchInto(ctx context.Context, buffered *room, pending pendingDownload) error {
	response, err := pending.fetch(ctx, pending.url, pending.headers)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return m.store.Put(ctx, buffered.directory()+pending.target.name, response.Body, response.ContentLength, pending.target.contentType)
}

// finishDownload records the outcome and wakes everyone waiting for the file.
// A file whose buffer dropped it meanwhile is deleted again.
func (m *Manager) finishDownload(buffered *room, finished *file, err error) {
	m.mutex.Lock()
	finished.failed = err != nil
	removed := finished.removed
	m.mutex.Unlock()
	if err != nil {
		log.Printf("livebuffer: %s: %v", finished.name, err)
	}
	close(finished.done)
	if removed {
		m.deleteFiles([]removedFile{{key: buffered.directory() + finished.name}})
	}
}

// Open returns the bytes of a buffered file, waiting for its download. It
// fails when the file left the buffer or its download failed.
func (m *Manager) Open(ctx context.Context, roomID, name string) (io.ReadCloser, string, error) {
	m.mutex.Lock()
	buffered := m.rooms[roomIdentifier(roomID)]
	var wanted *file
	if buffered != nil {
		buffered.lastTouched = m.now()
		wanted = buffered.files[name]
	}
	m.mutex.Unlock()
	if wanted == nil {
		return nil, "", fmt.Errorf("%s is not buffered", name)
	}
	m.promote(wanted)
	select {
	case <-wanted.done:
	case <-ctx.Done():
		return nil, "", ctx.Err()
	}
	m.mutex.Lock()
	failed := wanted.failed || wanted.removed
	m.mutex.Unlock()
	if failed {
		return nil, "", fmt.Errorf("%s is not available", name)
	}
	reader, err := m.store.Get(ctx, buffered.directory()+name)
	return reader, wanted.contentType, err
}

// Read returns the bytes of a buffered file.
func (m *Manager) Read(ctx context.Context, roomID, name string) ([]byte, error) {
	reader, _, err := m.Open(ctx, roomID, name)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}
