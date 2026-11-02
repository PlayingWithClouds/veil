// Package cams serves live webcam rooms: it lists the rooms cam plugins have
// online, keeps a CamModel row for every room it has seen (so rooms can be
// followed, auto-recorded and shown offline), resolves a room to a live
// stream through the HLS proxy, and records rooms into ordinary videos.
//
// A recording is a `record` job (RecordTask). Its footage lands under
// recordings/<job id>/ in the blob store and is filed as a Video with
// recordedFrom set and one MediaFile pointing at the recording's playlist.
package cams

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/devices"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/livebuffer"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/settings"
	"github.com/playingwithclouds/veil/internal/storage"
)

// camModelTable is the table cam rooms are stored in.
const camModelTable = "CamModel"

// ErrNotFound is returned for a cam model id that does not exist.
var ErrNotFound = errors.New("cam model not found")

// ErrOffline is returned when a room cannot be watched or recorded because it
// is not live.
var ErrOffline = errors.New("room is offline")

// Sites is what the service needs from the plugin layer.
type Sites interface {
	// Sources returns the active plugins that list cam rooms.
	Sources() []*plugins.Plugin
	// Plugin returns one active cam plugin by name.
	Plugin(name string) (*plugins.Plugin, bool)
	List(ctx context.Context, plugin *plugins.Plugin, args plugins.CamListArgs) (*plugins.CamListResult, error)
	Stream(ctx context.Context, plugin *plugins.Plugin, username string) (*plugins.ResolveResult, error)
}

// LiveProxy turns a resolved live stream into a playable local URL.
type LiveProxy interface {
	LiveManifestURL(ctx context.Context, result *plugins.ResolveResult, roomID string) (string, error)
}

// Fetcher downloads a URL with the headers a stream needs. A response with a
// status other than 200 is an error.
type Fetcher interface {
	Fetch(ctx context.Context, rawURL string, headers map[string]string) (*http.Response, error)
}

// Services are the parts of the backend the cam service works with.
type Services struct {
	Database *db.DB
	// Library finds or creates the studio and tags of recordings.
	Library *library.Library
	// Sites is usually NewPluginSites over the plugin registry and runner.
	Sites Sites
	// Proxy is the stream proxy that serves live sessions.
	Proxy LiveProxy
	// Blobs stores recordings under storage.RecordingPrefix.
	Blobs *storage.Client
	// Devices names this device's library storage location, where recordings are filed.
	Devices *devices.Service
	// Queue holds the record jobs.
	Queue *jobs.Queue
	// Settings gives the longest a recording may run; nil uses the default.
	Settings *settings.Service
	// Buffer, when set, lets a recording start in the past and keeps a
	// watched room rewindable while it is recorded.
	Buffer *livebuffer.Manager
	// Fetcher downloads playlists and segments; nil uses NewHTTPFetcher.
	Fetcher Fetcher
}

// timing holds the waits of recordings and the watcher; tests shorten them.
type timing struct {
	// offlineGrace is how long a room may be unreachable before its recording ends.
	offlineGrace time.Duration
	// retryDelay is the wait before resolving a broken stream again.
	retryDelay time.Duration
	// pollMinimum and pollMaximum clamp the wait between playlist refreshes.
	pollMinimum time.Duration
	pollMaximum time.Duration
	// watchInterval is how often Run checks the watched models.
	watchInterval time.Duration
	// watchStagger spaces the checks of consecutive models, so a long list
	// does not hit a site in one burst.
	watchStagger time.Duration
	// maxRecording, when set, replaces the settings' longest recording.
	maxRecording time.Duration
}

// defaultTiming is the timing used against live sites.
var defaultTiming = timing{
	offlineGrace:  90 * time.Second,
	retryDelay:    3 * time.Second,
	pollMinimum:   time.Second,
	pollMaximum:   4 * time.Second,
	watchInterval: 2 * time.Minute,
	watchStagger:  700 * time.Millisecond,
}

// Service lists, stores, streams and records cam rooms.
type Service struct {
	services Services
	fetcher  Fetcher
	timing   timing
	now      func() time.Time

	mutex  sync.Mutex
	chunks map[string]cachedChunk
	online map[string]cachedOnline
	// active maps a cam model id to its recording running in this process.
	active map[string]*capture
	// seeds holds the buffered past a requested recording starts from, by cam
	// model id, until its job picks it up.
	seeds map[string]*livebuffer.Seed
	// announced maps a watched model's id to the live.startedAt OnLive was
	// last called for ("" while offline).
	announced map[string]string
	onLive    func(schema.CamModel)
}

// New returns the cam service over services.
func New(services Services) *Service {
	fetcher := services.Fetcher
	if fetcher == nil {
		fetcher = NewHTTPFetcher()
	}
	return &Service{
		services:  services,
		fetcher:   fetcher,
		timing:    defaultTiming,
		now:       time.Now,
		chunks:    map[string]cachedChunk{},
		online:    map[string]cachedOnline{},
		active:    map[string]*capture{},
		seeds:     map[string]*livebuffer.Seed{},
		announced: map[string]string{},
	}
}

// SetOnLive sets the hook called when a watched model (auto-recorded or
// followed) goes live. It runs on the watcher's goroutine, so it must not block.
func (s *Service) SetOnLive(hook func(schema.CamModel)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.onLive = hook
}

// Sources returns the cam plugins, for the site filter.
func (s *Service) Sources() []*plugins.Plugin {
	return s.services.Sites.Sources()
}

// sleepContext waits for duration or until ctx ends.
func sleepContext(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}
