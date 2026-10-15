package graphql

import (
	"github.com/playingwithclouds/veil/internal/api/imgcache"
	"github.com/playingwithclouds/veil/internal/cams"
	"github.com/playingwithclouds/veil/internal/face"
	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/media"
	"github.com/playingwithclouds/veil/internal/pipeline"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/pluginstore"
	"github.com/playingwithclouds/veil/internal/recommend"
	"github.com/playingwithclouds/veil/internal/resolve"
	"github.com/playingwithclouds/veil/internal/settings"
	"github.com/playingwithclouds/veil/internal/sitesuggest"
	"github.com/playingwithclouds/veil/internal/storage"
	"github.com/playingwithclouds/veil/internal/stream"
	"github.com/playingwithclouds/veil/internal/subscriptions"
	"github.com/playingwithclouds/veil/internal/suggest"
	"github.com/playingwithclouds/veil/internal/updater"
)

// Resolver is the root gqlgen resolver. It holds every service dependency and is
// split into query/mutation/subscription sub-resolvers by the generated code.
type Resolver struct {
	repo         *media.Repository
	ingestSvc    *ingest.Service
	queue        *jobs.Queue
	registry     *plugins.Registry
	runner       *plugins.Runner
	orchestrator *pipeline.Orchestrator
	hub          *subscriptions.Hub
	streamSvc    *resolve.Service
	settings     *settings.Service
	streamProxy  *stream.Proxy
	imgCache     *imgcache.Handler
	searchRuns   *updater.Scheduler
	entityScenes *updater.EntityScenes
	recommender  *recommend.Engine
	feeder       *updater.Feeder
	pluginStore  *pluginstore.Store
	suggestions  *suggest.Service
	siteSuggest  *sitesuggest.Service
	blobStore    *storage.Client
	cams         *cams.Service
	camRecorder  *cams.Recorder
	faces        *face.Service
}

// SetCams wires in the live cam service.
func (r *Resolver) SetCams(service *cams.Service) { r.cams = service }

// SetBlobStore wires in the blob store, where generated scene thumbnails are kept.
func (r *Resolver) SetBlobStore(store *storage.Client) { r.blobStore = store }

func NewResolver(
	repo *media.Repository,
	ingestSvc *ingest.Service,
	queue *jobs.Queue,
	registry *plugins.Registry,
	runner *plugins.Runner,
	orchestrator *pipeline.Orchestrator,
	hub *subscriptions.Hub,
	streamSvc *resolve.Service,
	svc *settings.Service,
	streamProxy *stream.Proxy,
) *Resolver {
	return &Resolver{
		repo:         repo,
		ingestSvc:    ingestSvc,
		queue:        queue,
		registry:     registry,
		runner:       runner,
		orchestrator: orchestrator,
		hub:          hub,
		streamSvc:    streamSvc,
		settings:     svc,
		streamProxy:  streamProxy,
	}
}

func (r *Resolver) SetStreamProxy(p *stream.Proxy) { r.streamProxy = p }

// SetImageCache wires in the image cache so persisted results can prewarm their
// thumbnails.
func (r *Resolver) SetImageCache(h *imgcache.Handler) { r.imgCache = h }

// SetSearchScheduler wires in the search subscription scheduler.
func (r *Resolver) SetSearchScheduler(scheduler *updater.Scheduler) { r.searchRuns = scheduler }

// SetEntityScenes wires in the studio/performer scene fetcher.
func (r *Resolver) SetEntityScenes(fetcher *updater.EntityScenes) { r.entityScenes = fetcher }

// SetRecommender wires in the recommendation engine.
func (r *Resolver) SetRecommender(engine *recommend.Engine) { r.recommender = engine }

// SetFeeder wires in the feeder that pulls a short site-filtered feed's sites.
func (r *Resolver) SetFeeder(feeder *updater.Feeder) { r.feeder = feeder }

// SetPluginStore wires in the npm plugin store (install, update, uninstall).
func (r *Resolver) SetPluginStore(store *pluginstore.Store) { r.pluginStore = store }

// SetSuggestions wires in the search suggestion engine.
func (r *Resolver) SetSuggestions(service *suggest.Service) { r.suggestions = service }

// SetSiteSuggestions wires in the sites' own search-box suggestions.
func (r *Resolver) SetSiteSuggestions(service *sitesuggest.Service) { r.siteSuggest = service }

// SetCamRecorder wires in the live cam recorder.
func (r *Resolver) SetCamRecorder(recorder *cams.Recorder) { r.camRecorder = recorder }
