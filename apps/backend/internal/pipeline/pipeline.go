// Package pipeline moves data between the plugins and the library: it reads a
// video's or gallery's page the first time it's opened, lists a plugin's
// catalog or search results into the library as stubs, and turns a stream
// into a URL the player can load.
package pipeline

import (
	"context"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// searchTimeout bounds how long one plugin may take to answer a search.
const searchTimeout = 20 * time.Second

// Pipeline connects plugins and the library.
type Pipeline struct {
	library  *library.Library
	registry *plugins.Registry
	runner   *plugins.Runner
	// visiting holds one lock per record being visited, so two requests for
	// the same page fetch it once.
	visiting sync.Map
	// identified holds when each plugin last answered about a performer
	// ("plugin|id" → time.Time), so opening a page doesn't ask every site again.
	identified sync.Map
	// searches keeps each site's answer to a search page for searchFresh.
	searches *cache.Client
	// queue takes the slow work a page visit starts: listing a performer's videos.
	queue *jobs.Queue
	// reading holds the performers whose profiles CollectFavorite is reading.
	reading sync.Map
	// collections tallies each favourite's collection until it's done (id → *collection).
	collections sync.Map
	// notifier hears when a favourite's collection is done; nil in tests.
	notifier Notifier
}

// New returns a pipeline that keeps search answers in searches and queues
// slow work on queue.
func New(lib *library.Library, registry *plugins.Registry, runner *plugins.Runner, searches *cache.Client, queue *jobs.Queue) *Pipeline {
	return &Pipeline{library: lib, registry: registry, runner: runner, searches: searches, queue: queue}
}

// Visit reads the page behind each of a record's sources that hasn't been
// read yet, through the source's plugin. Sources whose plugin is gone or
// can't read pages are skipped; a failing page is logged, not returned, so
// one broken site doesn't hide the others.
func (p *Pipeline) Visit(ctx context.Context, kind schema.RecordKind, id string) error {
	return p.visit(ctx, kind, id, false, nil)
}

// Refresh reads the page behind every one of a record's sources again, read
// before or not, like Visit otherwise.
func (p *Pipeline) Refresh(ctx context.Context, kind schema.RecordKind, id string) error {
	return p.visit(ctx, kind, id, true, nil)
}

// RefreshFrom is Refresh for the sources of the given plugins only; none
// given means every source.
func (p *Pipeline) RefreshFrom(ctx context.Context, kind schema.RecordKind, id string, only []string) error {
	return p.visit(ctx, kind, id, true, only)
}

// visit reads a record's unread sources, or all of them when again is set,
// leaving out the sources of plugins not in only (when only names any), one
// visit per record at a time.
func (p *Pipeline) visit(ctx context.Context, kind schema.RecordKind, id string, again bool, only []string) error {
	lock, _ := p.visiting.LoadOrStore(id, &sync.Mutex{})
	lock.(*sync.Mutex).Lock()
	defer lock.(*sync.Mutex).Unlock()
	sources, err := p.library.Sources(ctx, kind, id)
	if err != nil {
		return err
	}
	for _, source := range sources {
		if source.FetchedAt != nil && !again {
			continue
		}
		if !pluginWanted(only, source.Plugin) {
			continue
		}
		p.readSource(ctx, kind, source)
	}
	return nil
}

// pluginWanted reports whether a plugin is among only; an empty only wants
// every plugin.
func pluginWanted(only []string, plugin string) bool {
	if len(only) == 0 {
		return true
	}
	return slices.Contains(only, plugin)
}

// readSource fetches and stores one source's page.
func (p *Pipeline) readSource(ctx context.Context, kind schema.RecordKind, source schema.Source) {
	plugin, entity, ok := p.finder(source.Plugin, kind)
	if !ok {
		return
	}
	results, err := p.runner.Find(ctx, plugin, entity, source.Url)
	if err != nil {
		log.Printf("visit %s %s: %v", source.Plugin, source.Url, err)
		return
	}
	if _, err := p.library.IngestFound(ctx, source.Plugin, results); err != nil {
		log.Printf("visit %s %s: store: %v", source.Plugin, source.Url, err)
		return
	}
	// Media pages mark their source as they're stored; performer and studio pages don't.
	if err := p.library.MarkFetched(ctx, source.Id); err != nil {
		log.Printf("visit %s %s: mark read: %v", source.Plugin, source.Url, err)
	}
	if kind != schema.RecordKindPerformer {
		return
	}
	if err := p.QueuePerformerVideos(ctx, source.SubjectId, source.Plugin); err != nil {
		log.Printf("visit %s %s: queue videos: %v", source.Plugin, source.Url, err)
	}
}

// finder returns the active plugin that reads pages of this kind.
func (p *Pipeline) finder(pluginName string, kind schema.RecordKind) (*plugins.Plugin, plugins.MediaType, bool) {
	entity, ok := mediaTypeOf(kind)
	if !ok || !p.registry.IsActive(pluginName) {
		return nil, "", false
	}
	plugin, found := p.registry.Get(pluginName)
	if !found || !plugin.Serves(plugins.FindCapability(entity)) {
		return nil, "", false
	}
	return plugin, entity, true
}

// List stores one page of a plugin's catalog (query empty) or search results
// (query set) for entity and returns what the items became.
func (p *Pipeline) List(ctx context.Context, pluginName string, entity plugins.MediaType, query string, limit, offset int) ([]schema.RecordRef, error) {
	plugin, found := p.registry.Get(pluginName)
	if !found || !p.registry.IsActive(pluginName) {
		return nil, fmt.Errorf("plugin %q isn't active", pluginName)
	}
	if !plugin.Serves(plugins.ListCapability(entity)) {
		return nil, fmt.Errorf("%s can't list %ss", pluginName, entity)
	}
	result, err := p.runner.List(ctx, plugin, entity, plugins.ListArgs{Query: query, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	return p.storeListed(ctx, pluginName, result.Items), nil
}

// storeListed stores listing items as stubs, skipping the ones that fail.
func (p *Pipeline) storeListed(ctx context.Context, pluginName string, items []plugins.DiscoveredItem) []schema.RecordRef {
	refs := []schema.RecordRef{}
	for _, item := range items {
		ref, err := p.library.IngestListed(ctx, pluginName, item)
		if err != nil {
			log.Printf("store %s item %s: %v", pluginName, item.SourceURL, err)
			continue
		}
		refs = append(refs, ref)
	}
	return refs
}

// interleave merges per-plugin result lists round-robin, dropping repeats,
// so no single site fills the top of the list.
func interleave(lists [][]schema.RecordRef) []schema.RecordRef {
	merged := []schema.RecordRef{}
	seen := map[string]bool{}
	for position := 0; ; position++ {
		added := false
		for _, list := range lists {
			if position >= len(list) {
				continue
			}
			added = true
			if !seen[list[position].Id] {
				seen[list[position].Id] = true
				merged = append(merged, list[position])
			}
		}
		if !added {
			return merged
		}
	}
}

// mediaTypeOf maps a record kind to the plugin entity that reads its pages.
func mediaTypeOf(kind schema.RecordKind) (plugins.MediaType, bool) {
	switch kind {
	case schema.RecordKindVideo:
		return plugins.MediaTypeScene, true
	case schema.RecordKindGallery:
		return plugins.MediaTypeGallery, true
	case schema.RecordKindPerformer:
		return plugins.MediaTypePerformer, true
	case schema.RecordKindStudio:
		return plugins.MediaTypeStudio, true
	case schema.RecordKindAudio:
		return plugins.MediaTypeAudio, true
	case schema.RecordKindStory:
		return plugins.MediaTypeStory, true
	case schema.RecordKindPost:
		return plugins.MediaTypePost, true
	}
	return "", false
}
