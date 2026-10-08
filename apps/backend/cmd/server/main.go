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
	searchapi "github.com/playingwithclouds/veil/internal/api/search"
	sseapi "github.com/playingwithclouds/veil/internal/api/sse"
	"github.com/playingwithclouds/veil/internal/blobcache"
	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/discovery"
	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/media"
	"github.com/playingwithclouds/veil/internal/netdns"
	"github.com/playingwithclouds/veil/internal/outbound"
	"github.com/playingwithclouds/veil/internal/pipeline"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/pluginstore"
	"github.com/playingwithclouds/veil/internal/recommend"
	"github.com/playingwithclouds/veil/internal/resolve"
	"github.com/playingwithclouds/veil/internal/settings"
	"github.com/playingwithclouds/veil/internal/storage"
	"github.com/playingwithclouds/veil/internal/stream"
	"github.com/playingwithclouds/veil/internal/subscriptions"
	"github.com/playingwithclouds/veil/internal/suggest"
	"github.com/playingwithclouds/veil/internal/updater"
	"github.com/playingwithclouds/veil/internal/vpn"
)

// pluginUpdateInterval is how often installed plugins are checked for newer
// npm versions (besides at every start).
const pluginUpdateInterval = 6 * time.Hour

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
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/x-icon"
	}
	if err := store.Put(ctx, key, resp.Body, resp.ContentLength, ct); err != nil {
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

// enableBlobEncryption encrypts stream-cache blobs (streams and downloads) at
// rest with encodedKey, a base64 AES-256 key the host app provides (the phone
// app derives it from the Android Keystore). An empty key leaves them plain.
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
	return store.WithEncryption(key, "stream-cache/")
}

// cacheMaxBytes returns the cache size limit: CACHE_MAX_BYTES when it is a
// positive number of bytes, else blobcache.DefaultMaxBytes.
func cacheMaxBytes() int64 {
	configured := os.Getenv("CACHE_MAX_BYTES")
	if configured == "" {
		return blobcache.DefaultMaxBytes
	}
	maxBytes, err := strconv.ParseInt(configured, 10, 64)
	if err != nil || maxBytes <= 0 {
		log.Printf("CACHE_MAX_BYTES %q is not a positive byte count, using %d", configured, blobcache.DefaultMaxBytes)
		return blobcache.DefaultMaxBytes
	}
	return maxBytes
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
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

	// Storage.
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

	// Settings (must init before anything that reads them).
	settingsSvc := settings.NewService(database)
	if err := settingsSvc.Init(ctx); err != nil {
		log.Fatalf("settings init: %v", err)
	}
	cfg := settingsSvc.Cached()

	// Repository (subscriptions only) and job queue.
	repo := media.NewRepository(database)

	ingestSvc := ingest.New(database)

	queue := jobs.NewQueue(database, cfg.MaxJobRetries)

	// Crash recovery: reset any jobs left in "running" state.
	if err := queue.ResetStuck(ctx); err != nil {
		log.Fatalf("reset stuck jobs: %v", err)
	}

	// Plugin system: bundled plugins (scripts/bundle-plugins.ts) run in the
	// embedded JS runtime, the same on a server and in the phone app.
	runner := plugins.NewEmbeddedRunner()

	hub := subscriptions.NewHub()

	// Installed plugin bundles, one folder each; hot-reloaded on change.
	pluginDir := envOr("PLUGIN_DIR", filepath.Join(dataDir, "plugins"))
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		log.Fatalf("plugin dir: %v", err)
	}

	// Plugins come from the GitHub "plugins" release: installed on demand,
	// updated in the background.
	// The phone app also ships a bundled set to seed from, so it works offline
	// on first launch.
	pluginStore := pluginstore.New(pluginDir, os.Getenv("PLUGIN_INDEX"))
	if seedDir := os.Getenv("PLUGIN_SEED_DIR"); seedDir != "" {
		if err := pluginStore.Seed(seedDir); err != nil {
			log.Printf("plugin seed: %v", err)
		}
	}

	// Pipeline orchestrator (created before registry so onLoad can reference it).
	orch := pipeline.New(queue, ingestSvc, nil, runner, hub, storeClient, settingsSvc)

	registry, err := plugins.NewRegistry(pluginDir, runner,
		plugins.WithIconFetcher(func(ctx context.Context, name, iconURL string) string {
			return cachePluginIcon(ctx, storeClient, name, iconURL)
		}),
		// Plugins inherit the backend's environment, FLARESOLVERR_URL included.
		plugins.WithSolver(strings.TrimSpace(os.Getenv("FLARESOLVERR_URL")) != ""),
	)
	if err != nil {
		log.Fatalf("plugins: %v", err)
	}
	registry.RestoreDisabled(cfg.DisabledPlugins)
	orch.SetRegistry(registry)
	go pluginStore.Run(ctx, pluginUpdateInterval)

	// Worker pool: size from settings.
	allKinds := []jobs.Kind{jobs.KindScrape, jobs.KindEnrich, jobs.KindDownload, jobs.KindSpeedCheck}
	concurrency := cfg.MaxConcurrentJobs
	if concurrency <= 0 {
		concurrency = 4
	}
	worker := jobs.NewWorker(queue, allKinds, concurrency, func(ctx context.Context, j *jobs.Job) {
		orch.RunJob(ctx, j)
	})
	worker.SetVpnGate(func() bool {
		if !settingsSvc.Cached().RequireVpn {
			return true
		}
		connected, _ := vpn.IsConnected()
		return connected
	})
	go worker.Start(ctx)

	// Search subscriptions: saved searches re-run on their own schedule.
	discoverySvc := discovery.New(ingestSvc, registry, runner)
	searchScheduler := updater.NewScheduler(repo, discoverySvc)
	go searchScheduler.Start(ctx)

	// Inject per-plugin settings (API keys etc.) as env vars on each invocation.
	runner.SetSettingsProvider(settingsSvc)

	// Stream cache: HLS segment + MP4 files in the blob store, indexed in the kv cache.
	streamCache := stream.NewCacheService(cacheClient, storeClient)

	// Stream resolver service (cache → domain plugin → wildcard plugin → direct).
	streamResolver := resolve.New(registry, runner)
	streamResolver.SetCache(streamCache)
	orch.SetStreamResolver(streamResolver)
	orch.SetStreamCache(streamCache)

	// GraphQL API.
	resolver := gqlapi.NewResolver(repo, ingestSvc, queue, registry, runner, orch, hub, streamResolver, settingsSvc, nil)
	gqlHandler, err := gqlapi.NewHandler(resolver)
	if err != nil {
		log.Fatalf("graphql schema: %v", err)
	}

	// HLS proxy: rewrites manifest URIs to local proxy URLs, write-through caches segments.
	streamProxy := stream.NewProxy(streamCache, cacheClient, publicURL)
	resolver.SetStreamProxy(streamProxy)

	// Cache size limit: stream and image cache blobs beyond it are evicted,
	// least recently used first. Downloads are pinned and don't count.
	streamCache.SetPinnedURLs(queue.DownloadURLs)
	cacheEvictor := blobcache.New(storeClient, cacheMaxBytes(),
		streamCache.EvictionSource(), imgapi.EvictionSource())
	storeClient.OnPut(cacheEvictor.NoteWrite)
	go func() {
		reconcileBlobs(ctx, streamCache, queue)
		cacheEvictor.Run(ctx)
	}()

	imgHandler := imgapi.New(storeClient)
	resolver.SetImageCache(imgHandler)
	resolver.SetSearchScheduler(searchScheduler)
	resolver.SetEntityScenes(updater.NewEntityScenes(searchScheduler, cacheClient))
	recommender := recommend.New(database, repo)
	resolver.SetRecommender(recommender)
	resolver.SetPluginStore(pluginStore)
	// Search suggestions: build the name/phrase index now, not on the first keystroke.
	suggestions := suggest.New(database, recommender, repo)
	resolver.SetSuggestions(suggestions)
	go suggestions.Warm(ctx)

	mux := http.NewServeMux()
	mux.Handle("/graphql", gqlHandler)
	mux.Handle("/api/events", sseapi.New(hub))
	mux.Handle("/api/search", searchapi.New(ingestSvc, discoverySvc, repo))
	mux.Handle("/api/img", imgHandler)
	mux.Handle("/api/stream/", streamProxy)
	mux.Handle(storage.RoutePrefix, storeClient.Handler())
	mux.HandleFunc("/api/schema", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/graphql")
		fmt.Fprint(w, gqlapi.SchemaSDL)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		// HOST empty = every interface; the phone app binds to 127.0.0.1.
		Addr:    net.JoinHostPort(os.Getenv("HOST"), port),
		Handler: corsMiddleware(mux),
	}

	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	log.Printf("listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
