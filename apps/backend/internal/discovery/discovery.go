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

// Service searches plugins and ingests their results.
type Service struct {
	ingestSvc *ingest.Service
	registry  *plugins.Registry
	runner    *plugins.Runner
}

// New creates a discovery service.
func New(ingestSvc *ingest.Service, registry *plugins.Registry, runner *plugins.Runner) *Service {
	return &Service{ingestSvc: ingestSvc, registry: registry, runner: runner}
}

// Options describes one search.
type Options struct {
	Query string
	Limit int
	// Offset paginates each plugin's listing when Paginate is set (infinite
	// scroll feeds); plain keyword searches leave it off.
	Offset   int
	Paginate bool
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

// Run searches searchPlugins concurrently and calls onHit, one at a time, for
// every ingested result as each plugin finishes. A failing plugin is logged
// and skipped; the failures come back joined. Returns once every plugin is
// done or ctx is cancelled.
func (s *Service) Run(ctx context.Context, options Options, searchPlugins []*plugins.Plugin, onHit func(Hit)) error {
	type pluginResult struct {
		name  string
		items []plugins.DiscoveredItem
		err   error
	}
	results := make(chan pluginResult, len(searchPlugins))
	for _, plugin := range searchPlugins {
		go func(plugin *plugins.Plugin) {
			items, err := s.listPlugin(ctx, plugin, options)
			results <- pluginResult{name: plugin.Meta.Name, items: items, err: err}
		}(plugin)
	}

	var failures []error
	for range searchPlugins {
		select {
		case <-ctx.Done():
			return errors.Join(append(failures, ctx.Err())...)
		case result := <-results:
			if result.err != nil {
				log.Printf("search %q via plugin %s: %v", options.Query, result.name, result.err)
				failures = append(failures, fmt.Errorf("%s: %w", result.name, result.err))
				continue
			}
			s.ingestHits(ctx, result.name, result.items, onHit)
		}
	}
	return errors.Join(failures...)
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

// listPlugin lists stubs from one plugin for every entity it exposes to search
// (scenes, galleries, performers, studios).
func (s *Service) listPlugin(ctx context.Context, plugin *plugins.Plugin, options Options) ([]plugins.DiscoveredItem, error) {
	entities := plugin.SearchListEntities()
	if len(entities) == 0 {
		return nil, nil
	}
	args := plugins.ListArgs{Query: options.Query, Limit: options.Limit}
	if options.Paginate {
		args.Offset = options.Offset
	}
	var items []plugins.DiscoveredItem
	var firstErr error
	for _, entity := range entities {
		result, err := s.runner.List(ctx, plugin, entity, args)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		items = append(items, result.Items...)
	}
	// Only surface an error when nothing at all came back, so one failing entity
	// doesn't discard the other's results.
	if len(items) == 0 && firstErr != nil {
		return nil, firstErr
	}
	return items, nil
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
