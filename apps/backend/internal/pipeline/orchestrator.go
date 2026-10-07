package pipeline

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/resolve"
	"github.com/playingwithclouds/veil/internal/settings"
	"github.com/playingwithclouds/veil/internal/storage"
	"github.com/playingwithclouds/veil/internal/stream"
	"github.com/playingwithclouds/veil/internal/subscriptions"
)

// Orchestrator routes plugin output through the ingestion pipeline.
type Orchestrator struct {
	queue          *jobs.Queue
	ingestSvc      *ingest.Service
	registry       *plugins.Registry
	runner         *plugins.Runner
	hub            *subscriptions.Hub
	store          *storage.Client
	streamResolver *resolve.Service
	streamCache    *stream.CacheService
	settings       *settings.Service
	// Scene ids with a detail fetch / related fill in progress (VisitScene).
	detailInflight  sync.Map
	relatedInflight sync.Map
}

func New(
	queue *jobs.Queue,
	ingestSvc *ingest.Service,
	registry *plugins.Registry,
	runner *plugins.Runner,
	hub *subscriptions.Hub,
	store *storage.Client,
	svc *settings.Service,
) *Orchestrator {
	return &Orchestrator{
		queue:     queue,
		ingestSvc: ingestSvc,
		registry:  registry,
		runner:    runner,
		hub:       hub,
		store:     store,
		settings:  svc,
	}
}

// SetStreamResolver wires in the stream resolver (breaks the init cycle between orchestrator and resolver).
func (o *Orchestrator) SetStreamResolver(r *resolve.Service) {
	o.streamResolver = r
}

// SetStreamCache wires in the stream cache service.
func (o *Orchestrator) SetStreamCache(c *stream.CacheService) {
	o.streamCache = c
}

// SetRegistry wires in the plugin registry after construction (breaks the init cycle).
func (o *Orchestrator) SetRegistry(r *plugins.Registry) {
	o.registry = r
}

// RunJob executes a job and updates the queue accordingly.
func (o *Orchestrator) RunJob(ctx context.Context, j *jobs.Job) {
	var err error
	switch j.Kind {
	case jobs.KindScrape:
		err = o.runScrape(ctx, j)
	case jobs.KindEnrich:
		err = o.runEnrich(ctx, j)
	case jobs.KindDownload:
		err = o.runDownload(ctx, j)
	case jobs.KindSpeedCheck:
		err = o.runSpeedCheck(ctx, j)
	default:
		log.Printf("orchestrator: unknown job kind: %s", j.Kind)
		return
	}

	if errors.Is(err, errDownloadDeferred) {
		// Not a failure: hold the download and re-check shortly.
		_ = o.queue.Defer(ctx, j.ID, 15*time.Second)
		return
	}

	if err != nil {
		log.Printf("job %s (%s/%s) failed: %v", j.StringID(), j.Kind, j.PluginName, err)
		_ = o.queue.Fail(ctx, j, err, backoffDelay(j.Kind, j.Attempts, o.settings))
		o.hub.Publish(subscriptions.TopicJobUpdated, subscriptions.JobUpdatedEvent{
			ID:        j.StringID(),
			Kind:      string(j.Kind),
			Status:    "failed",
			Error:     err.Error(),
			UpdatedAt: subscriptions.NowRFC3339(),
		})
	} else if j.Kind == jobs.KindDownload {
		// Keep download jobs so the queue page can show the completed download.
		_ = o.queue.MarkCompleted(ctx, j.ID)
		o.hub.Publish(subscriptions.TopicJobUpdated, subscriptions.JobUpdatedEvent{
			ID:        j.StringID(),
			Kind:      string(j.Kind),
			Status:    "completed",
			UpdatedAt: subscriptions.NowRFC3339(),
		})
	} else {
		_ = o.queue.Complete(ctx, j.ID)
		o.hub.Publish(subscriptions.TopicJobUpdated, subscriptions.JobUpdatedEvent{
			ID:        j.StringID(),
			Kind:      string(j.Kind),
			Status:    "completed",
			UpdatedAt: subscriptions.NowRFC3339(),
		})
	}
}

// backoffDelay computes the exponential retry delay for a job kind and attempt count.
// attempts is 1-based (first run = 1). Returns 0 if no per-kind limit is configured.
func backoffDelay(kind jobs.Kind, attempts int, svc *settings.Service) time.Duration {
	if svc == nil || attempts <= 0 {
		return 0
	}
	cfg := svc.Cached()
	lim, ok := cfg.KindLimits[string(kind)]
	if !ok || lim.RetryInitialMs <= 0 {
		return 0
	}
	mult := lim.RetryMultiplier
	if mult <= 1.0 {
		mult = 2.0
	}
	delayMs := float64(lim.RetryInitialMs)
	for i := 1; i < attempts; i++ {
		delayMs *= mult
	}
	if lim.RetryMaxMs > 0 && delayMs > float64(lim.RetryMaxMs) {
		delayMs = float64(lim.RetryMaxMs)
	}
	return time.Duration(delayMs) * time.Millisecond
}
