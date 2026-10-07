package graphql

import (
	"context"
	"log"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// Upper bound on a background persist sweep for one search/browse fetch. Generous
// because the sweep also downloads and resizes each result's thumbnail.
const persistDiscoveredTimeout = 2 * time.Minute

// discoveredItemWithPlugin pairs a discovered item with the plugin that produced
// it, so the background persist sweep knows which source to attribute.
type discoveredItemWithPlugin struct {
	plugin string
	item   plugins.DiscoveredItem
}

// fanOutSearch runs the given query against every plugin concurrently and
// returns each plugin's results paired with its name. One goroutine per plugin;
// a failing plugin contributes nothing rather than aborting the batch. Results
// are not deduped — callers decide how to merge.
func (r *Resolver) fanOutSearch(ctx context.Context, searchPlugins []*plugins.Plugin, query string, max int) []discoveredItemWithPlugin {
	type pluginResult struct {
		name  string
		items []plugins.DiscoveredItem
	}
	ch := make(chan pluginResult, len(searchPlugins))

	for _, plugin := range searchPlugins {
		go func(plugin *plugins.Plugin) {
			entity, ok := plugin.ContentListEntity()
			if !ok {
				ch <- pluginResult{name: plugin.Meta.Name}
				return
			}
			result, err := r.runner.List(ctx, plugin, entity, plugins.ListArgs{Query: query, Limit: max})
			if err != nil {
				log.Printf("fanOutSearch %q via %s: %v", query, plugin.Meta.Name, err)
				ch <- pluginResult{name: plugin.Meta.Name}
				return
			}
			ch <- pluginResult{name: plugin.Meta.Name, items: result.Items}
		}(plugin)
	}

	collected := make([]discoveredItemWithPlugin, 0, len(searchPlugins)*max)
	for range searchPlugins {
		res := <-ch
		for _, item := range res.items {
			collected = append(collected, discoveredItemWithPlugin{plugin: res.name, item: item})
		}
	}
	return collected
}

// filterPluginsByName keeps only plugins whose name is in the allowed set. An
// empty (or nil) allow list is treated as "no filter" and returns all plugins.
func filterPluginsByName(all []*plugins.Plugin, names []string) []*plugins.Plugin {
	if len(names) == 0 {
		return all
	}
	allowed := make(map[string]bool, len(names))
	for _, name := range names {
		allowed[name] = true
	}
	filtered := make([]*plugins.Plugin, 0, len(all))
	for _, plugin := range all {
		if allowed[plugin.Meta.Name] {
			filtered = append(filtered, plugin)
		}
	}
	return filtered
}

// adultPlugins returns all source plugins. Every source in veil is adult,
// so the historical NSFW partition is a no-op kept for call-site compatibility.
func adultPlugins(all []*plugins.Plugin) []*plugins.Plugin {
	return all
}

// persistDiscoveredItems upserts each fetched result as a minimal local scene
// stub in the background. This builds a local catalog from what the user browses,
// enabling fast initial paint on repeat fetches and a signal for recommendations.
//
// Best-effort: it runs detached from the request (results are already returned to
// the caller), upserts are idempotent by source_url, and failures are only logged.
func (r *Resolver) persistDiscoveredItems(items []discoveredItemWithPlugin) {
	if len(items) == 0 {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), persistDiscoveredTimeout)
		defer cancel()
		for _, entry := range items {
			_, err := r.ingestSvc.IngestDiscoveredItem(ctx, entry.plugin, entry.item)
			if err != nil {
				log.Printf("persist discovered item %q via %s: %v", entry.item.SourceURL, entry.plugin, err)
			}
			r.prefetchThumbnail(ctx, entry.item.PosterPath)
		}
	}()
}

// Upper bound on the background search-logging write.
const logSearchTimeout = 15 * time.Second

// logSearch records a search as a recommendation signal, in the background.
// Best-effort: it runs detached from the request and only logs failures.
func (r *Resolver) logSearch(query string, pluginNames []string, resultCount int) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), logSearchTimeout)
		defer cancel()
		if err := r.repo.LogSearch(ctx, query, pluginNames, resultCount); err != nil {
			log.Printf("log search %q: %v", query, err)
		}
	}()
}

// prefetchThumbnail warms the image cache for a result's poster so its first view
// is served (resized) from MinIO instead of the source CDN. Best-effort.
func (r *Resolver) prefetchThumbnail(ctx context.Context, posterURL string) {
	if r.imgCache == nil || posterURL == "" {
		return
	}
	if err := r.imgCache.Prefetch(ctx, posterURL); err != nil {
		log.Printf("prefetch thumbnail %q: %v", posterURL, err)
	}
}

// pluginSearchResult maps a plugin's raw discovered item onto the GraphQL model.
func pluginSearchResult(pluginName string, item plugins.DiscoveredItem) *model.PluginSearchResult {
	result := &model.PluginSearchResult{
		ExternalID:    item.ExternalID,
		Title:         item.Title,
		MediaType:     string(item.MediaType),
		Plugin:        pluginName,
		SourceURL:     item.SourceURL,
		PreviewImages: item.PreviewImages,
	}
	if result.PreviewImages == nil {
		result.PreviewImages = []string{}
	}
	if item.Date != "" {
		date := item.Date
		result.Date = &date
	}
	if item.PosterPath != "" {
		poster := item.PosterPath
		result.PosterURL = &poster
	}
	if item.PreviewVideo != "" {
		preview := item.PreviewVideo
		result.PreviewVideo = &preview
	}
	if item.DurationSeconds > 0 {
		duration := item.DurationSeconds
		result.DurationSeconds = &duration
	}
	return result
}
