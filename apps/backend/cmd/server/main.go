package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	gqlapi "github.com/playingwithclouds/veil/internal/api/graphql"
	imgapi "github.com/playingwithclouds/veil/internal/api/imgcache"
	"github.com/playingwithclouds/veil/internal/blobcache"
	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/devices"
	"github.com/playingwithclouds/veil/internal/files"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/netdns"
	"github.com/playingwithclouds/veil/internal/outbound"
	"github.com/playingwithclouds/veil/internal/pipeline"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/pluginstore"
	"github.com/playingwithclouds/veil/internal/resolve"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/settings"
	"github.com/playingwithclouds/veil/internal/storage"
	"github.com/playingwithclouds/veil/internal/stream"
	"github.com/playingwithclouds/veil/internal/user"
)

// pluginUpdateInterval is how often installed plugins are checked for newer
// published versions (besides at every start).
const pluginUpdateInterval = 6 * time.Hour

// cachePluginIcon stores a plugin's icon in the blob store and returns its
// local URL, or the original URL when it can't be fetched.
func cachePluginIcon(ctx context.Context, store *storage.Client, name, iconURL string) string {
	key := "plugin-icons/" + name
	if store.Exists(ctx, key) {
		return store.DirectURL(key)
	}
	resp, err := outbound.Client.Get(iconURL) //nolint:gosec
	if err != nil || resp.StatusCode != http.StatusOK {
		return iconURL
	}
	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/x-icon"
	}
	if err := store.Put(ctx, key, resp.Body, resp.ContentLength, contentType); err != nil {
		log.Printf("plugin icon cache: %v", err)
		return iconURL
	}
	return store.DirectURL(key)
}

// envOr returns the environment variable key, or fallback when it is unset.
func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// reconcileBlobs drops stream cache entries and download links that point at
// blobs no longer on disk (e.g. deleted by hand), so they are fetched from the
// origin again instead of failing.
func reconcileBlobs(ctx context.Context, streamCache *stream.CacheService, queue *jobs.Queue) {
	removed, err := streamCache.Reconcile(ctx)
	if err != nil {
		log.Printf("stream cache reconcile: %v", err)
	} else if removed > 0 {
		log.Printf("stream cache reconcile: dropped %d entries for missing blobs", removed)
	}
	forgotten, err := queue.ForgetMissingDownloads(ctx, func(downloadURL string) bool {
		return streamCache.Complete(ctx, downloadURL)
	})
	if err != nil {
		log.Printf("downloads reconcile: %v", err)
	} else if forgotten > 0 {
		log.Printf("downloads reconcile: %d downloads lost their file", forgotten)
	}
}

// encryptedPrefixes are the blobs encrypted at rest: cached streams and
// downloads, cam recordings and the live buffer, uploads.
var encryptedPrefixes = []string{"stream-cache/", "recordings/", "live-buffer/", "uploads/"}

// enableBlobEncryption encrypts media blobs (encryptedPrefixes) at rest with
// encodedKey, a base64 AES-256 key the host app provides (the phone app
// derives it from the Android Keystore). An empty key leaves them plain.
// BLOB_ENCRYPTION_WRITES=off stops encrypting new blobs while still reading
// encrypted ones.
func enableBlobEncryption(store *storage.Client, encodedKey, writes string) error {
	if encodedKey == "" {
		return nil
	}
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return fmt.Errorf("BLOB_ENCRYPTION_KEY is not base64: %w", err)
	}
	if writes == "off" {
		// Keep reading blobs encrypted earlier, but write new ones plain.
		return store.WithEncryption(key)
	}
	return store.WithEncryption(key, encryptedPrefixes...)
}

// configuredCacheBytes returns CACHE_MAX_BYTES when it is a positive number
// of bytes, else nil (the cache location's stored quota applies).
func configuredCacheBytes() *int64 {
	configured := os.Getenv("CACHE_MAX_BYTES")
	if configured == "" {
		return nil
	}
	maxBytes, err := strconv.ParseInt(configured, 10, 64)
	if err != nil || maxBytes <= 0 {
		log.Printf("CACHE_MAX_BYTES %q is not a positive byte count, ignoring it", configured)
		return nil
	}
	return &maxBytes
}

// openDevice returns this install's device row: named DEVICE_NAME (else the
// host name) on first start, platform DEVICE_PLATFORM (default server).
func openDevice(ctx context.Context, database *db.DB) *devices.Service {
	platform, err := devices.ParsePlatform(os.Getenv("DEVICE_PLATFORM"))
	if err != nil {
		log.Fatalf("device: %v", err)
	}
	hostName, _ := os.Hostname()
	service, err := devices.Open(ctx, database, envOr("DEVICE_NAME", hostName), platform)
	if err != nil {
		log.Fatalf("device: %v", err)
	}
	return service
}

// openStorageLocations makes sure this device has its library and cache
// locations in the blob store, and returns the cache's quota and the
// location. CACHE_MAX_BYTES, when set, overrides the stored quota.
func openStorageLocations(ctx context.Context, deviceService *devices.Service, blobDir string) (*schema.StorageLocation, int64) {
	absolute := absolutePath(blobDir)
	if _, err := deviceService.EnsureLocation(ctx, schema.StorageKindLibrary, absolute, nil, nil); err != nil {
		log.Fatalf("storage locations: %v", err)
	}
	defaultQuota := int64(blobcache.DefaultMaxBytes)
	cacheLocation, err := deviceService.EnsureLocation(ctx, schema.StorageKindCache, absolute, configuredCacheBytes(), &defaultQuota)
	if err != nil {
		log.Fatalf("storage locations: %v", err)
	}
	if cacheLocation.QuotaBytes == nil {
		return cacheLocation, defaultQuota
	}
	return cacheLocation, *cacheLocation.QuotaBytes
}

// registerJobs gives the job runner its tasks, with the per-kind limits from
// settings.
func registerJobs(runner *jobs.Runner, flow *pipeline.Pipeline, configured settings.Settings) {
	runner.Register(jobs.KindScrape, jobLimits(configured, jobs.KindScrape), func(ctx context.Context, job *jobs.Job, progress func(float64)) error {
		if pipeline.IsPerformerVideosJob(job) {
			return flow.CollectPerformerVideos(ctx, job, progress)
		}
		if job.Subject == nil {
			return fmt.Errorf("a fetch job needs a subject")
		}
		return flow.Visit(ctx, job.Subject.Kind, job.Subject.Id)
	})
}

// jobLimits reads a kind's concurrency and retry backoff from settings.
func jobLimits(configured settings.Settings, kind jobs.Kind) jobs.Limits {
	limit := configured.KindLimits[string(kind)]
	concurrency := limit.MaxConcurrent
	if concurrency <= 0 {
		concurrency = configured.MaxConcurrentJobs
	}
	return jobs.Limits{
		Concurrency:     concurrency,
		RetryInitial:    time.Duration(limit.RetryInitialMs) * time.Millisecond,
		RetryMultiplier: limit.RetryMultiplier,
		RetryMax:        time.Duration(limit.RetryMaxMs) * time.Millisecond,
	}
}

// maintenanceInterval is how often trash and finished jobs are purged.
const maintenanceInterval = 6 * time.Hour

// finishedJobRetention is how long finished jobs stay listed.
const finishedJobRetention = 7 * 24 * time.Hour

// maintained are the services whose trash and finished work is purged.
type maintained struct {
	library *library.Library
	user    *user.Service
	files   *files.Service
	queue   *jobs.Queue
}

// runMaintenance purges what outlived its retention, at start and then every
// maintenanceInterval, until ctx ends.
func runMaintenance(ctx context.Context, services maintained, settingsService *settings.Service) {
	for {
		purgeExpired(ctx, services, settingsService.Cached().TrashRetentionDays)
		select {
		case <-ctx.Done():
			return
		case <-time.After(maintenanceInterval):
		}
	}
}

// purgeExpired removes trash (records, saved filters, files) older than
// retentionDays and finished jobs older than finishedJobRetention.
func purgeExpired(ctx context.Context, services maintained, retentionDays int) {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	if purged, err := services.library.PurgeTrash(ctx, cutoff); err != nil {
		log.Printf("maintenance: purge trash: %v", err)
	} else if purged > 0 {
		log.Printf("maintenance: purged %d trashed records", purged)
	}
	if _, err := services.user.PurgeTrash(ctx, cutoff); err != nil {
		log.Printf("maintenance: purge saved filters: %v", err)
	}
	if _, err := services.files.PurgeTrash(ctx, cutoff); err != nil {
		log.Printf("maintenance: purge files: %v", err)
	}
	if _, err := services.queue.PurgeFinished(ctx, time.Now().Add(-finishedJobRetention)); err != nil {
		log.Printf("maintenance: purge jobs: %v", err)
	}
}

// absolutePath makes a path absolute, keeping it as given when that fails.
func absolutePath(path string) string {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return absolute
}

// corsMiddleware lets the web app on another port call the API.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// openPlugins loads the installed plugins and starts the store that installs
// and updates them.
func openPlugins(ctx context.Context, dataDir string, runner *plugins.Runner, store *storage.Client, settingsService *settings.Service) (*plugins.Registry, *pluginstore.Store) {
	// Installed plugin bundles, one folder each; hot-reloaded on change.
	pluginDir := envOr("PLUGIN_DIR", filepath.Join(dataDir, "plugins"))
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		log.Fatalf("plugin dir: %v", err)
	}
	// Plugins come from the GitHub "plugins" release. The phone app also ships
	// a bundled set to seed from, so it works offline on first launch.
	pluginStore := pluginstore.New(pluginDir, os.Getenv("PLUGIN_INDEX"))
	if seedDir := os.Getenv("PLUGIN_SEED_DIR"); seedDir != "" {
		if err := pluginStore.Seed(seedDir); err != nil {
			log.Printf("plugin seed: %v", err)
		}
	}
	registry, err := plugins.NewRegistry(pluginDir, runner,
		plugins.WithIconFetcher(func(ctx context.Context, name, iconURL string) string {
			return cachePluginIcon(ctx, store, name, iconURL)
		}),
		// Plugins inherit the backend's environment, FLARESOLVERR_URL included.
		plugins.WithSolver(strings.TrimSpace(os.Getenv("FLARESOLVERR_URL")) != ""),
	)
	if err != nil {
		log.Fatalf("plugins: %v", err)
	}
	registry.RestoreDisabled(settingsService.Cached().DisabledPlugins)
	// Inject per-plugin settings (API keys etc.) as env vars on each invocation.
	runner.SetSettingsProvider(settingsService)
	go pluginStore.Run(ctx, pluginUpdateInterval)
	return registry, pluginStore
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	port := envOr("PORT", "8080")
	// Externally reachable origin; stream-proxy and blob URLs are built from it.
	publicURL := envOr("PUBLIC_URL", "http://localhost:"+port)
	// Everything the backend persists lives here: the SQLite database, blobs
	// and installed plugins.
	dataDir := envOr("DATA_DIR", "./data")
	// Hosts without a system resolver Go can find (Android) pass their DNS
	// servers, as a list or a file kept current by the app.
	netdns.Configure(os.Getenv("DNS_SERVERS"), os.Getenv("DNS_SERVERS_FILE"))

	database, err := db.Open(ctx, filepath.Join(dataDir, "veil.db"))
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer database.Close()

	cacheClient := cache.New(ctx, database)
	storeClient, err := storage.New(filepath.Join(dataDir, "blobs"), publicURL)
	if err != nil {
		log.Fatalf("storage: %v", err)
	}
	if err := enableBlobEncryption(storeClient, os.Getenv("BLOB_ENCRYPTION_KEY"), os.Getenv("BLOB_ENCRYPTION_WRITES")); err != nil {
		log.Fatalf("blob encryption: %v", err)
	}

	// Settings must load before anything that reads them.
	settingsService := settings.NewService(database)
	if err := settingsService.Init(ctx); err != nil {
		log.Fatalf("settings init: %v", err)
	}

	deviceService := openDevice(ctx, database)
	userService := user.New(database, deviceService.Current().Id)

	queue := jobs.NewQueue(database, settingsService.Cached().MaxJobRetries)
	queue.SetDevice(deviceService.Current().Id)
	// Crash recovery: jobs left running go back to the queue.
	if err := queue.ResetStuck(ctx); err != nil {
		log.Fatalf("reset stuck jobs: %v", err)
	}

	// Plugins run in the embedded JS runtime, the same on a server and in the
	// phone app.
	runner := plugins.NewEmbeddedRunner()
	registry, pluginStore := openPlugins(ctx, dataDir, runner, storeClient, settingsService)

	lib := library.New(database)
	if err := lib.SeedDefaultCategories(ctx); err != nil {
		log.Printf("default categories: %v", err)
	}
	flow := pipeline.New(lib, registry, runner, cacheClient, queue)

	// Stream cache: HLS segment + MP4 files in the blob store, indexed in the kv cache.
	streamCache := stream.NewCacheService(cacheClient, storeClient)
	// Stream resolver (cache → domain plugin → wildcard plugin → direct).
	streamResolver := resolve.New(registry, runner)
	streamResolver.SetCache(streamCache)
	// HLS proxy: rewrites manifest URIs to local proxy URLs, write-through caches segments.
	streamProxy := stream.NewProxy(streamCache, cacheClient, publicURL)

	// Cache size limit: stream and image cache blobs beyond it are evicted,
	// least recently used first. Downloads are pinned and don't count.
	streamCache.SetPinnedURLs(queue.DownloadURLs)
	blobDir := absolutePath(filepath.Join(dataDir, "blobs"))
	cacheLocation, cacheQuota := openStorageLocations(ctx, deviceService, blobDir)
	cacheEvictor := blobcache.New(storeClient, cacheQuota, streamCache.EvictionSource(), imgapi.EvictionSource())
	cacheEvictor.OnPass(func(report blobcache.Report) {
		if err := deviceService.SetUsedBytes(ctx, cacheLocation.Id, report.TotalBytes); err != nil {
			log.Printf("cache location: %v", err)
		}
	})
	storeClient.OnPut(cacheEvictor.NoteWrite)
	go func() {
		reconcileBlobs(ctx, streamCache, queue)
		cacheEvictor.Run(ctx)
	}()

	feature := openFeatures(ctx, core{
		database: database, library: lib, pipeline: flow, registry: registry, runner: runner, queue: queue,
		settings: settingsService, devices: deviceService, user: userService, store: storeClient,
		streamCache: streamCache, streamResolver: streamResolver, streamProxy: streamProxy,
		dataDir: dataDir, blobDir: blobDir, publicURL: publicURL,
	})
	defer feature.notify.Wait()

	// Every task is registered before the runner starts claiming jobs.
	jobRunner := jobs.NewRunner(queue)
	registerJobs(jobRunner, flow, settingsService.Cached())
	feature.registerTasks(jobRunner, settingsService.Cached())
	jobRunner.Start(ctx)
	feature.start(ctx)
	go runMaintenance(ctx, maintained{library: lib, user: userService, files: feature.files, queue: queue}, settingsService)

	resolver := gqlapi.NewResolver(gqlapi.Services{
		Library:     lib,
		Pipeline:    flow,
		Registry:    registry,
		Runner:      runner,
		PluginStore: pluginStore,
		Settings:    settingsService,
		Resolver:    streamResolver,
		Proxy:       streamProxy,
		User:        userService,
		Devices:     deviceService,
		Queue:       queue,
		JobRunner:   jobRunner,
		Evictor:     cacheEvictor,
		Notify:      feature.notify,
		Follows:     feature.follows,
		Limits:      feature.limits,
		Automations: feature.automations,
		Cams:        feature.cams,
		Files:       feature.files,
		Recommend:   feature.recommend,
		Embeddings:  feature.embeddings,
		Faces:       feature.faces,
		Lookup:      feature.lookup,
		Tokens:      feature.tokens,
		Accounts:    feature.accounts,
	})

	mux := http.NewServeMux()
	mux.Handle("/graphql", gqlapi.NewHandler(resolver))
	mux.Handle("/api/img", imgapi.New(storeClient))
	mux.Handle("/api/stream/", streamProxy)
	mux.Handle(storage.RoutePrefix, storeClient.Handler())
	mux.Handle(files.FileRoutePrefix, feature.files.Handler())
	mux.Handle(files.UploadRoute, feature.files.UploadHandler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	requireToken := func() bool { return settingsService.Cached().RequireAPIToken }
	server := &http.Server{
		// HOST empty = every interface; the phone app binds to 127.0.0.1.
		Addr:    net.JoinHostPort(os.Getenv("HOST"), port),
		Handler: corsMiddleware(feature.tokens.Middleware(requireToken, mux)),
	}
	go func() {
		<-ctx.Done()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownContext)
	}()

	log.Printf("listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
