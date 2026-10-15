// Package discovery runs keyword searches across plugins and ingests every
// hit as a stub. The live search endpoint and scheduled search subscriptions
// share it.
package discovery

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// listFunc lists one entity type of a plugin, handing each batch of items to
// onItems as the plugin finds them (every item once, the final result's
// included); plugins.Runner.ListStream in production.
type listFunc func(ctx context.Context, plugin *plugins.Plugin, entity plugins.MediaType, args plugins.ListArgs, onItems func([]plugins.DiscoveredItem)) (*plugins.ItemsResult, error)

// Service searches plugins and ingests their results.
type Service struct {
	ingestSvc *ingest.Service
	registry  *plugins.Registry
	runner    *plugins.Runner
	list      listFunc
}

// New creates a discovery service.
func New(ingestSvc *ingest.Service, registry *plugins.Registry, runner *plugins.Runner) *Service {
	return &Service{ingestSvc: ingestSvc, registry: registry, runner: runner, list: runner.ListStream}
}

// Options describes one search.
type Options struct {
	Query string
	Limit int
	// Offset paginates each plugin's listing when Paginate is set (infinite
	// scroll feeds); plain keyword searches leave it off.
	Offset   int
	Paginate bool
	// Entity, when set, lists only that entity type (the feeder wants scenes);
	// empty lists every type each plugin searches.
	Entity plugins.MediaType
}

// Hit is one plugin result after ingest. ID is the canonical record id, empty
// when the stub could not be ingested.
type Hit struct {
	Plugin string
	ID     string
	Item   plugins.DiscoveredItem
}

// Plugins returns the active search-capable plugins, restricted to the named
// sources when any are given.
func (s *Service) Plugins(sources []string) []*plugins.Plugin {
	all := s.registry.WithAnyCapability(plugins.CapabilitySceneList, plugins.CapabilityGalleryList)
	return filterBySources(all, sources)
}

// listJob is one list call: a plugin's listing of one entity type.
type listJob struct {
	plugin *plugins.Plugin
	entity plugins.MediaType
}

// listBatch is what a listJob delivered: a batch of items while it runs, or
// its end (finished, with the error if it failed).
type listBatch struct {
	job      listJob
	items    []plugins.DiscoveredItem
	finished bool
	err      error
}

// pluginOutcome tallies what one plugin's list calls have returned so far.
type pluginOutcome struct {
	itemCount int
	firstErr  error
}

// Run searches searchPlugins and calls onHit, one at a time, for every
// ingested result. Each (plugin, entity type) list runs concurrently, and
// every batch a plugin emits is ingested and handed on as soon as it
// arrives, so results stream in while plugins are still crawling pages or
// sites. A plugin counts as failed only when none of its lists returned
// anything and one errored; the failures are logged and come back joined.
// Returns once every list is done or ctx is cancelled.
func (s *Service) Run(ctx context.Context, options Options, searchPlugins []*plugins.Plugin, onHit func(Hit)) error {
	args := plugins.ListArgs{Query: options.Query, Limit: options.Limit}
	if options.Paginate {
		args.Offset = options.Offset
	}
	jobs := listJobs(searchPlugins, options.Entity)
	batches := make(chan listBatch, len(jobs))
	for _, job := range jobs {
		go s.runJob(ctx, job, args, batches)
	}

	outcomes := make(map[string]*pluginOutcome, len(searchPlugins))
	for finished := 0; finished < len(jobs); {
		select {
		case <-ctx.Done():
			return errors.Join(append(pluginFailures(options.Query, searchPlugins, outcomes), ctx.Err())...)
		case batch := <-batches:
			recordOutcome(outcomes, batch)
			if batch.finished {
				finished++
				continue
			}
			s.ingestHits(ctx, batch.job.plugin.Meta.Name, batch.items, onHit)
		}
	}
	return errors.Join(pluginFailures(options.Query, searchPlugins, outcomes)...)
}

// runJob runs one list call, sending each item batch and finally its end.
// Sends give up once ctx ends, so an abandoned search can't block the plugin.
func (s *Service) runJob(ctx context.Context, job listJob, args plugins.ListArgs, batches chan<- listBatch) {
	send := func(batch listBatch) {
		select {
		case batches <- batch:
		case <-ctx.Done():
		}
	}
	_, err := s.list(ctx, job.plugin, job.entity, args, func(items []plugins.DiscoveredItem) {
		send(listBatch{job: job, items: items})
	})
	send(listBatch{job: job, finished: true, err: err})
}

// listJobs expands each plugin into one job per entity type it lists for
// search, only the wanted type when one is given.
func listJobs(searchPlugins []*plugins.Plugin, wanted plugins.MediaType) []listJob {
	var jobs []listJob
	for _, plugin := range searchPlugins {
		for _, entity := range plugin.SearchListEntities() {
			if wanted != "" && entity != wanted {
				continue
			}
			jobs = append(jobs, listJob{plugin: plugin, entity: entity})
		}
	}
	return jobs
}

// recordOutcome adds a batch to its plugin's tally, keeping the first error.
func recordOutcome(outcomes map[string]*pluginOutcome, result listBatch) {
	name := result.job.plugin.Meta.Name
	outcome := outcomes[name]
	if outcome == nil {
		outcome = &pluginOutcome{}
		outcomes[name] = outcome
	}
	outcome.itemCount += len(result.items)
	if result.err != nil && outcome.firstErr == nil {
		outcome.firstErr = result.err
	}
}

// pluginFailures logs and returns an error for each plugin that has errored
// without returning anything, so one failing entity type doesn't discard the
// results of another.
func pluginFailures(query string, searchPlugins []*plugins.Plugin, outcomes map[string]*pluginOutcome) []error {
	var failures []error
	for _, plugin := range searchPlugins {
		outcome := outcomes[plugin.Meta.Name]
		if outcome == nil || outcome.itemCount > 0 || outcome.firstErr == nil {
			continue
		}
		log.Printf("search %q via plugin %s: %v", query, plugin.Meta.Name, outcome.firstErr)
		failures = append(failures, fmt.Errorf("%s: %w", plugin.Meta.Name, outcome.firstErr))
	}
	return failures
}

// ListPage lists the scenes on one of plugin's site pages (a channel,
// performer or category page, see plugins.CapabilitySceneListPage), ingests
// them as stubs and calls onHit for each.
func (s *Service) ListPage(ctx context.Context, plugin *plugins.Plugin, pageURL string, limit int, onHit func(Hit)) error {
	_, err := s.list(ctx, plugin, plugins.MediaTypeScene, plugins.ListArgs{URL: pageURL, Limit: limit}, func(items []plugins.DiscoveredItem) {
		s.ingestHits(ctx, plugin.Meta.Name, items, onHit)
	})
	return err
}

// ingestHits stores each item as a stub (details are only fetched when a
// scene is visited) and hands it to onHit.
func (s *Service) ingestHits(ctx context.Context, pluginName string, items []plugins.DiscoveredItem, onHit func(Hit)) {
	for _, item := range items {
		canonicalID, err := s.ingestSvc.IngestDiscoveredItem(ctx, pluginName, item)
		if err != nil {
			log.Printf("search: ingest stub %q: %v", item.SourceURL, err)
		}
		onHit(Hit{Plugin: pluginName, ID: canonicalID, Item: item})
	}
}

// SplitSources parses a comma-separated plugin list, dropping blanks.
func SplitSources(sources string) []string {
	var names []string
	for _, name := range strings.Split(sources, ",") {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			names = append(names, trimmed)
		}
	}
	return names
}

// filterBySources keeps only plugins named in sources. An empty list returns
// the input unchanged (search all providers).
func filterBySources(all []*plugins.Plugin, sources []string) []*plugins.Plugin {
	if len(sources) == 0 {
		return all
	}
	wanted := make(map[string]bool, len(sources))
	for _, name := range sources {
		wanted[name] = true
	}
	filtered := make([]*plugins.Plugin, 0, len(all))
	for _, plugin := range all {
		if wanted[plugin.Meta.Name] {
			filtered = append(filtered, plugin)
		}
	}
	return filtered
}
