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

// listFunc lists one entity type of a plugin; plugins.Runner.List in production.
type listFunc func(ctx context.Context, plugin *plugins.Plugin, entity plugins.MediaType, args plugins.ListArgs) (*plugins.ItemsResult, error)

// Service searches plugins and ingests their results.
type Service struct {
	ingestSvc *ingest.Service
	registry  *plugins.Registry
	runner    *plugins.Runner
	list      listFunc
}

// New creates a discovery service.
func New(ingestSvc *ingest.Service, registry *plugins.Registry, runner *plugins.Runner) *Service {
	return &Service{ingestSvc: ingestSvc, registry: registry, runner: runner, list: runner.List}
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

// listResult is what a finished listJob returned.
type listResult struct {
	job   listJob
	items []plugins.DiscoveredItem
	err   error
}

// pluginOutcome tallies what one plugin's list calls have returned so far.
type pluginOutcome struct {
	itemCount int
	firstErr  error
}

// Run searches searchPlugins and calls onHit, one at a time, for every
// ingested result. Each (plugin, entity type) list runs concurrently and its
// results are ingested and handed on as soon as that list finishes, so a
// plugin's slow galleries never hold back its scenes. A plugin counts as
// failed only when none of its lists returned anything and one errored; the
// failures are logged and come back joined. Returns once every list is done
// or ctx is cancelled.
func (s *Service) Run(ctx context.Context, options Options, searchPlugins []*plugins.Plugin, onHit func(Hit)) error {
	args := plugins.ListArgs{Query: options.Query, Limit: options.Limit}
	if options.Paginate {
		args.Offset = options.Offset
	}
	jobs := listJobs(searchPlugins, options.Entity)
	results := make(chan listResult, len(jobs))
	for _, job := range jobs {
		go func(job listJob) {
			result, err := s.list(ctx, job.plugin, job.entity, args)
			if err != nil {
				results <- listResult{job: job, err: err}
				return
			}
			results <- listResult{job: job, items: result.Items}
		}(job)
	}

	outcomes := make(map[string]*pluginOutcome, len(searchPlugins))
	for range jobs {
		select {
		case <-ctx.Done():
			return errors.Join(append(pluginFailures(options.Query, searchPlugins, outcomes), ctx.Err())...)
		case result := <-results:
			recordOutcome(outcomes, result)
			s.ingestHits(ctx, result.job.plugin.Meta.Name, result.items, onHit)
		}
	}
	return errors.Join(pluginFailures(options.Query, searchPlugins, outcomes)...)
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

// recordOutcome adds a finished list to its plugin's tally, keeping the first error.
func recordOutcome(outcomes map[string]*pluginOutcome, result listResult) {
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
	result, err := s.runner.List(ctx, plugin, plugins.MediaTypeScene, plugins.ListArgs{URL: pageURL, Limit: limit})
	if err != nil {
		return err
	}
	s.ingestHits(ctx, plugin.Meta.Name, result.Items, onHit)
	return nil
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
