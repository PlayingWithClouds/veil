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
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/netdns"
	"github.com/playingwithclouds/veil/internal/outbound"
	"github.com/playingwithclouds/veil/internal/pipeline"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/pluginstore"
	"github.com/playingwithclouds/veil/internal/resolve"
	"github.com/playingwithclouds/veil/internal/settings"
	"github.com/playingwithclouds/veil/internal/storage"
	"github.com/playingwithclouds/veil/internal/stream"
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

// corsMiddleware lets the web app on another port call the API.
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

	queue := jobs.NewQueue(database, settingsService.Cached().MaxJobRetries)
	// Crash recovery: jobs left running go back to the queue.
	if err := queue.ResetStuck(ctx); err != nil {
		log.Fatalf("reset stuck jobs: %v", err)
	}

	// Plugins run in the embedded JS runtime, the same on a server and in the
	// phone app.
	runner := plugins.NewEmbeddedRunner()
	registry, pluginStore := openPlugins(ctx, dataDir, runner, storeClient, settingsService)

	lib := library.New(database)
	flow := pipeline.New(lib, registry, runner)

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
	cacheEvictor := blobcache.New(storeClient, cacheMaxBytes(), streamCache.EvictionSource(), imgapi.EvictionSource())
	storeClient.OnPut(cacheEvictor.NoteWrite)
	go func() {
		reconcileBlobs(ctx, streamCache, queue)
		cacheEvictor.Run(ctx)
	}()

	resolver := gqlapi.NewResolver(gqlapi.Services{
		Library:     lib,
		Pipeline:    flow,
		Registry:    registry,
		Runner:      runner,
		PluginStore: pluginStore,
		Settings:    settingsService,
		Resolver:    streamResolver,
		Proxy:       streamProxy,
	})

	mux := http.NewServeMux()
	mux.Handle("/graphql", gqlapi.NewHandler(resolver))
	mux.Handle("/api/img", imgapi.New(storeClient))
	mux.Handle("/api/stream/", streamProxy)
	mux.Handle(storage.RoutePrefix, storeClient.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		// HOST empty = every interface; the phone app binds to 127.0.0.1.
		Addr:    net.JoinHostPort(os.Getenv("HOST"), port),
		Handler: corsMiddleware(mux),
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
