package files

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/devices"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/resolve"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/settings"
	"github.com/playingwithclouds/veil/internal/storage"
	"github.com/playingwithclouds/veil/internal/stream"
)

// testEnv is a service on a fresh database, blob store and device.
type testEnv struct {
	service     *Service
	database    *db.DB
	devices     *devices.Service
	queue       *jobs.Queue
	settings    *settings.Service
	store       *storage.Client
	streamCache *stream.CacheService
	added       []schema.MediaFile
}

// newTestEnv opens a test environment whose device has cache and library
// locations in the blob store.
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	t.Cleanup(cancel)
	deviceService, err := devices.Open(ctx, database, "test", schema.PlatformServer)
	if err != nil {
		t.Fatal(err)
	}
	blobDir := filepath.Join(t.TempDir(), "blobs")
	store, err := storage.New(blobDir, "http://veil.test")
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []schema.StorageKind{schema.StorageKindLibrary, schema.StorageKindCache} {
		if _, err := deviceService.EnsureLocation(ctx, kind, blobDir, nil, nil); err != nil {
			t.Fatal(err)
		}
	}
	settingsService := settings.NewService(database)
	if err := settingsService.Init(ctx); err != nil {
		t.Fatal(err)
	}
	queue := jobs.NewQueue(database, 3)
	queue.SetDevice(deviceService.Current().Id)
	env := &testEnv{database: database, devices: deviceService, queue: queue, settings: settingsService, store: store}
	env.streamCache = stream.NewCacheService(cache.New(ctx, database), store)
	env.service = New(Config{
		Database: database, Devices: deviceService, Queue: queue, Settings: settingsService,
		Store: store, BlobDir: blobDir, PublicURL: "http://veil.test",
		Resolver: resolve.New(nil, nil), StreamCache: env.streamCache,
		OnFileAdded: func(file schema.MediaFile) { env.added = append(env.added, file) },
	})
	return env
}

// libraryRoot adds a library location at a temp dir with a root folder
// "media" in it, and returns the root and its folder.
func (env *testEnv) libraryRoot(t *testing.T, template *string) (*schema.LibraryRoot, string) {
	t.Helper()
	ctx := context.Background()
	locationDir := t.TempDir()
	location, err := env.devices.AddLocation(ctx, schema.StorageKindLibrary, locationDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	folder := filepath.Join(locationDir, "media")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	interval := 60
	root, err := env.service.AddRoot(ctx, location.Id, "media", &interval, template)
	if err != nil {
		t.Fatal(err)
	}
	return root, folder
}

// runJob claims the queued job of kind, runs task on it and records the
// outcome like the runner does.
func (env *testEnv) runJob(t *testing.T, kind jobs.Kind, task jobs.Task) *jobs.Job {
	t.Helper()
	ctx := context.Background()
	job, err := env.queue.Claim(ctx, []jobs.Kind{kind})
	if err != nil || job == nil {
		t.Fatalf("claim %s: job %v, err %v", kind, job, err)
	}
	if err := task(ctx, job, func(float64) {}); err != nil {
		t.Fatalf("%s: %v", kind, err)
	}
	if err := env.queue.Succeed(ctx, job.ID); err != nil {
		t.Fatal(err)
	}
	finished, err := env.queue.Get(ctx, job.ID)
	if err != nil {
		t.Fatal(err)
	}
	return finished
}

// scan queues and runs a scan of root.
func (env *testEnv) scan(t *testing.T, rootID string) {
	t.Helper()
	if _, err := env.service.ScanRoot(context.Background(), rootID); err != nil {
		t.Fatal(err)
	}
	env.runJob(t, schema.JobKindLibraryScan, env.service.ScanTask())
}

// writeFile writes content to target, creating its folders.
func writeFile(t *testing.T, target string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

// patterned returns size bytes that differ per seed.
func patterned(size int, seed byte) []byte {
	content := make([]byte, size)
	for index := range content {
		content[index] = byte(index*7) + seed
	}
	return content
}

// newVideo inserts a video titled title.
func (env *testEnv) newVideo(t *testing.T, title string) schema.RecordRef {
	t.Helper()
	ref, err := env.service.createMedia(context.Background(), schema.RecordKindVideo, title)
	if err != nil {
		t.Fatal(err)
	}
	return *ref
}
