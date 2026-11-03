package cams

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/devices"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/storage"
)

// fakeSites serves canned rooms per plugin and counts listing calls. Streams
// resolve to the live origin when one is set.
type fakeSites struct {
	mutex     sync.Mutex
	rooms     map[string][]plugins.CamRoom
	offline   map[string]bool
	listCalls int
	live      *liveServer
	// streamCalls counts cam:stream calls; refuseStreams makes them fail.
	streamCalls   atomic.Int32
	refuseStreams atomic.Bool
}

// Sources lists the plugins that have rooms.
func (sites *fakeSites) Sources() []*plugins.Plugin {
	var sources []*plugins.Plugin
	for _, name := range []string{"alpha", "beta"} {
		if _, ok := sites.rooms[name]; ok {
			sources = append(sources, &plugins.Plugin{Meta: plugins.PluginMeta{Name: name, DisplayName: strings.ToUpper(name)}})
		}
	}
	return sources
}

// Plugin returns one of Sources by name.
func (sites *fakeSites) Plugin(name string) (*plugins.Plugin, bool) {
	for _, plugin := range sites.Sources() {
		if plugin.Meta.Name == name {
			return plugin, true
		}
	}
	return nil, false
}

// List returns the page of the plugin's rooms at the args' offset and limit.
func (sites *fakeSites) List(ctx context.Context, plugin *plugins.Plugin, args plugins.CamListArgs) (*plugins.CamListResult, error) {
	sites.mutex.Lock()
	defer sites.mutex.Unlock()
	sites.listCalls++
	rooms := sites.rooms[plugin.Meta.Name]
	start := min(args.Offset, len(rooms))
	end := min(start+args.Limit, len(rooms))
	return &plugins.CamListResult{Rooms: rooms[start:end], Total: len(rooms)}, nil
}

// Stream resolves a room to the live origin's master playlist with its
// current token, or a made-up URL without an origin.
func (sites *fakeSites) Stream(ctx context.Context, plugin *plugins.Plugin, username string) (*plugins.ResolveResult, error) {
	sites.streamCalls.Add(1)
	sites.mutex.Lock()
	offline := sites.offline[username]
	sites.mutex.Unlock()
	if offline || sites.refuseStreams.Load() {
		return nil, errors.New("offline")
	}
	if sites.live == nil {
		return &plugins.ResolveResult{URL: "https://edge.example/" + username + ".m3u8", MimeType: "application/x-mpegURL"}, nil
	}
	token := sites.live.token()
	return &plugins.ResolveResult{URL: sites.live.server.URL + "/live/master.m3u8?token=" + token, MimeType: "application/x-mpegURL"}, nil
}

// setOffline marks a username offline or online.
func (sites *fakeSites) setOffline(username string, offline bool) {
	sites.mutex.Lock()
	defer sites.mutex.Unlock()
	sites.offline[username] = offline
}

// fakeProxy turns a stream into a recognisable local URL.
type fakeProxy struct{}

// LiveManifestURL wraps the stream URL.
func (fakeProxy) LiveManifestURL(ctx context.Context, result *plugins.ResolveResult, roomID string) (string, error) {
	return "http://backend.test/api/stream/manifest?live=" + result.URL, nil
}

// room is an online listing room.
func room(username string, viewers int) plugins.CamRoom {
	return plugins.CamRoom{
		Username: username, SourceURL: "https://site.example/" + username + "/",
		Viewers: viewers, Tags: []string{"tag"}, Online: true,
	}
}

// liveServer is a fake HLS origin. Each playlist request shows the next
// window of segments; the last window repeats once the list is exhausted.
type liveServer struct {
	server *httptest.Server

	mutex sync.Mutex
	// windows lists, per playlist request, the sequence numbers shown.
	windows [][]int
	// discontinuityAt lists sequence numbers preceded by a discontinuity tag.
	discontinuityAt map[int]bool
	requests        int
	// validToken is the only token the origin accepts; others get 403.
	validToken string
	// down makes every request fail with 503.
	down bool
	// withAudio adds a separate audio rendition to the master playlist.
	withAudio bool
}

// newLiveServer starts the origin.
func newLiveServer(t *testing.T, windows [][]int) *liveServer {
	t.Helper()
	live := &liveServer{windows: windows, discontinuityAt: map[int]bool{}, validToken: "1"}
	live.server = httptest.NewServer(http.HandlerFunc(live.serve))
	t.Cleanup(live.server.Close)
	return live
}

// serve answers master, media, segment and thumbnail requests.
func (live *liveServer) serve(writer http.ResponseWriter, request *http.Request) {
	live.mutex.Lock()
	defer live.mutex.Unlock()
	if strings.HasSuffix(request.URL.Path, ".jpg") {
		fmt.Fprint(writer, "jpeg-bytes")
		return
	}
	if live.down {
		http.Error(writer, "down", http.StatusServiceUnavailable)
		return
	}
	if request.URL.Query().Get("token") != live.validToken {
		http.Error(writer, "expired", http.StatusForbidden)
		return
	}
	live.answer(writer, request.URL.Path)
}

// answer writes the resource at urlPath. The caller holds the lock.
func (live *liveServer) answer(writer io.Writer, urlPath string) {
	switch {
	case strings.HasSuffix(urlPath, "/master.m3u8"):
		live.writeMaster(writer)
	case strings.HasSuffix(urlPath, "/audio.m3u8"):
		live.writeAudio(writer)
	case strings.HasSuffix(urlPath, "/high.m3u8"):
		live.writeMedia(writer, "seg")
	case strings.HasSuffix(urlPath, "/mid.m3u8"):
		live.writeMedia(writer, "segmid")
	case strings.HasSuffix(urlPath, "/low.m3u8"):
		live.writeMedia(writer, "seglow")
	case strings.HasSuffix(urlPath, ".ts"):
		fmt.Fprint(writer, segmentContent(urlPath))
	}
}

// segmentContent is what the origin serves for a segment path.
func segmentContent(urlPath string) string {
	number := strings.TrimSuffix(path.Base(urlPath), ".ts")
	switch {
	case strings.HasPrefix(urlPath, "/live/aseg/"):
		return "audio-" + number
	case strings.HasPrefix(urlPath, "/live/segmid/"):
		return "mid-" + number
	case strings.HasPrefix(urlPath, "/live/seglow/"):
		return "low-" + number
	}
	return "data-" + number
}

// writeMaster writes the master playlist with three renditions.
func (live *liveServer) writeMaster(writer io.Writer) {
	audioGroup := ""
	fmt.Fprint(writer, "#EXTM3U\n")
	if live.withAudio {
		fmt.Fprint(writer, "#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"a\",NAME=\"Audio\",DEFAULT=YES,URI=\"audio.m3u8?token="+live.validToken+"\"\n")
		audioGroup = ",AUDIO=\"a\""
	}
	token := "?token=" + live.validToken
	fmt.Fprint(writer,
		"#EXT-X-STREAM-INF:BANDWIDTH=500000,RESOLUTION=640x360"+audioGroup+"\nlow.m3u8"+token+
			"\n#EXT-X-STREAM-INF:BANDWIDTH=1000000,RESOLUTION=854x480"+audioGroup+"\nmid.m3u8"+token+
			"\n#EXT-X-STREAM-INF:BANDWIDTH=2000000,RESOLUTION=1280x720"+audioGroup+"\nhigh.m3u8"+token+"\n")
}

// writeMedia writes the media playlist for the current request, with segments
// under the directory.
func (live *liveServer) writeMedia(writer io.Writer, directory string) {
	step := min(live.requests, len(live.windows)-1)
	live.requests++
	window := live.windows[step]
	fmt.Fprintf(writer, "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:2\n#EXT-X-MEDIA-SEQUENCE:%d\n", window[0])
	for _, sequence := range window {
		if live.discontinuityAt[sequence] {
			fmt.Fprint(writer, "#EXT-X-DISCONTINUITY\n")
		}
		fmt.Fprintf(writer, "#EXTINF:2.000,\n%s/%d.ts?token=%s\n", directory, sequence, live.validToken)
	}
}

// writeAudio writes the audio playlist: the segments of the window the video
// playlist showed last.
func (live *liveServer) writeAudio(writer io.Writer) {
	step := min(max(live.requests-1, 0), len(live.windows)-1)
	window := live.windows[step]
	fmt.Fprintf(writer, "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:2\n#EXT-X-MEDIA-SEQUENCE:%d\n", window[0])
	for _, sequence := range window {
		fmt.Fprintf(writer, "#EXTINF:2.000,\naseg/%d.ts?token=%s\n", sequence, live.validToken)
	}
}

// token is the token the origin accepts now.
func (live *liveServer) token() string {
	live.mutex.Lock()
	defer live.mutex.Unlock()
	return live.validToken
}

// expireToken invalidates the current token in favour of next.
func (live *liveServer) expireToken(next string) {
	live.mutex.Lock()
	defer live.mutex.Unlock()
	live.validToken = next
}

// setDown switches the origin off or on.
func (live *liveServer) setDown(down bool) {
	live.mutex.Lock()
	defer live.mutex.Unlock()
	live.down = down
}

// setAudio switches the separate audio rendition on.
func (live *liveServer) setAudio() {
	live.mutex.Lock()
	defer live.mutex.Unlock()
	live.withAudio = true
}

// fixture is a cam service over a fresh database, blob store, device and job
// queue, with one stored room ("nova" on plugin alpha).
type fixture struct {
	service  *Service
	sites    *fakeSites
	live     *liveServer
	database *db.DB
	blobs    *storage.Client
	queue    *jobs.Queue
	runner   *jobs.Runner
	ctx      context.Context
	roomID   string
}

// newFixture builds the service against an origin serving windows; with
// startRunner the record jobs run on a job runner.
func newFixture(t *testing.T, windows [][]int, startRunner bool) *fixture {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	blobRoot := t.TempDir()
	blobs, err := storage.New(blobRoot, "http://backend.test")
	if err != nil {
		t.Fatal(err)
	}
	deviceService := openDevice(t, ctx, database, blobRoot)
	live := newLiveServer(t, windows)
	sites := &fakeSites{rooms: map[string][]plugins.CamRoom{"alpha": {room("nova", 10)}}, offline: map[string]bool{}, live: live}
	queue := jobs.NewQueue(database, 3)
	service := New(Services{
		Database: database, Library: library.New(database), Sites: sites, Proxy: fakeProxy{},
		Blobs: blobs, Devices: deviceService, Queue: queue,
	})
	service.timing = timing{
		offlineGrace: 300 * time.Millisecond, retryDelay: 10 * time.Millisecond,
		pollMinimum: 5 * time.Millisecond, pollMaximum: 10 * time.Millisecond,
		watchInterval: time.Hour, watchStagger: time.Millisecond, maxRecording: time.Hour,
	}
	stored := storeNova(t, ctx, service, live)
	runner := jobs.NewRunner(queue)
	runner.Register(schema.JobKindRecord, jobs.Limits{Concurrency: 2}, service.RecordTask())
	if startRunner {
		runner.Start(ctx)
	}
	return &fixture{service: service, sites: sites, live: live, database: database, blobs: blobs, queue: queue, runner: runner, ctx: ctx, roomID: stored.Id}
}

// openDevice creates this device with a library location at blobRoot.
func openDevice(t *testing.T, ctx context.Context, database *db.DB, blobRoot string) *devices.Service {
	t.Helper()
	deviceService, err := devices.Open(ctx, database, "test", schema.PlatformServer)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := deviceService.EnsureLocation(ctx, schema.StorageKindLibrary, blobRoot, nil, nil); err != nil {
		t.Fatal(err)
	}
	return deviceService
}

// storeNova stores the fixture's room as a listing would.
func storeNova(t *testing.T, ctx context.Context, service *Service, live *liveServer) *schema.CamModel {
	t.Helper()
	stored, err := service.upsertRoom(ctx, "alpha", plugins.CamRoom{
		Username: "nova", DisplayName: "Nova", SourceURL: "https://site.example/nova/",
		ThumbnailURL: live.server.URL + "/nova.jpg", Tags: []string{"redhead"}, Online: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	return stored
}

// waitFor polls until condition holds or the test times out.
func waitFor(t *testing.T, description string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", description)
}

// recordedSeconds is how much footage the running recording has stored.
func (f *fixture) recordedSeconds() float64 {
	active := f.service.ActiveRecordings()
	if len(active) == 0 {
		return 0
	}
	return active[0].DurationSeconds
}

// start begins recording the room and waits until the job captures.
func (f *fixture) start(t *testing.T, options RecordOptions) string {
	t.Helper()
	jobID, err := f.service.StartRecording(f.ctx, f.roomID, options)
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the recording to run", func() bool { return f.service.runningJob(f.roomID) == jobID })
	return jobID
}

// stop stops the room's recording.
func (f *fixture) stop(t *testing.T) {
	t.Helper()
	if err := f.service.StopRecording(f.ctx, f.roomID); err != nil {
		t.Fatal(err)
	}
}

// waitJob waits until the job has finished and returns it.
func (f *fixture) waitJob(t *testing.T, jobID string) *jobs.Job {
	t.Helper()
	var job *jobs.Job
	waitFor(t, "the job to finish", func() bool {
		job, _ = f.queue.Get(f.ctx, jobID)
		return job != nil && job.FinishedAt != nil
	})
	return job
}

// recordingVideos returns the videos recorded from the room.
func (f *fixture) recordingVideos(t *testing.T) []schema.Video {
	t.Helper()
	videos, err := db.QueryAs[schema.Video](f.ctx, f.database,
		`SELECT * FROM "Video" WHERE "recordedFrom" = $room ORDER BY "createdAt"`, db.Vars{"room": f.roomID})
	if err != nil {
		t.Fatal(err)
	}
	return videos
}

// read returns a stored file of the recording of job.
func (f *fixture) read(t *testing.T, jobID, name string) string {
	t.Helper()
	reader, err := f.blobs.Get(f.ctx, blobDirectory(jobID)+name)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	defer reader.Close()
	data, _ := io.ReadAll(reader)
	return string(data)
}
