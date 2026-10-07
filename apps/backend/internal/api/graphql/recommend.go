package graphql

import (
	"context"
	"log"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/recommend"
)

// Defaults for the recommendation queries' optional size arguments.
const (
	defaultFeedLimit     = 25
	defaultRowLimit      = 8
	defaultPerRow        = 12
	defaultCategoryLimit = 8
	defaultPerCategory   = 12
)

// derefInt returns *value, or fallback when it is nil.
func derefInt(value *int, fallback int) int {
	if value == nil {
		return fallback
	}
	return *value
}

// itemSceneIDs lists the items' scene ids in order.
func itemSceneIDs(items []recommend.Item) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, item.SceneID)
	}
	return out
}

// recommendedScenes loads the items' scenes and pairs each with its source
// and reason. Items whose scene is gone are skipped.
func (r *Resolver) recommendedScenes(ctx context.Context, items []recommend.Item) ([]*model.RecommendedScene, error) {
	scenes, err := r.repo.ScenesByIDs(ctx, itemSceneIDs(items))
	if err != nil {
		return nil, err
	}
	scenesByID := make(map[string]*model.Scene, len(scenes))
	for _, scene := range scenes {
		scenesByID[scene.ID] = scene
	}
	out := make([]*model.RecommendedScene, 0, len(items))
	for _, item := range items {
		scene, found := scenesByID[item.SceneID]
		if !found {
			continue
		}
		out = append(out, &model.RecommendedScene{Scene: scene, Source: item.Source, Reason: reasonModel(item.Reason), Score: item.Score})
	}
	return out, nil
}

// recommendationRow converts one engine row, loading its scenes.
func (r *Resolver) recommendationRow(ctx context.Context, row recommend.Row) (*model.RecommendationRow, error) {
	items, err := r.recommendedScenes(ctx, row.Items)
	if err != nil {
		return nil, err
	}
	return &model.RecommendationRow{Key: row.Key, Title: row.Title, Reason: reasonModel(row.Reason), Items: items}, nil
}

// recommendedCategory converts one engine category; nil when its tag is gone.
func (r *Resolver) recommendedCategory(ctx context.Context, category recommend.Category) (*model.RecommendedCategory, error) {
	tag, err := r.repo.GetTag(ctx, category.TagID)
	if err != nil || tag == nil {
		return nil, err
	}
	scenes, err := r.repo.ScenesByIDs(ctx, itemSceneIDs(category.Items))
	if err != nil {
		return nil, err
	}
	return &model.RecommendedCategory{Tag: tag, Scenes: scenes}, nil
}

// reasonModel converts a reason; empty entity fields become null.
func reasonModel(reason recommend.Reason) *model.RecommendationReason {
	out := &model.RecommendationReason{Kind: reason.Kind, Text: reason.Text}
	if reason.EntityID != "" {
		entityID := reason.EntityID
		out.EntityID = &entityID
	}
	if reason.EntityName != "" {
		entityName := reason.EntityName
		out.EntityName = &entityName
	}
	return out
}

// impressionsFromInput converts the mutation input; absent fields stay empty.
func impressionsFromInput(inputs []*model.ImpressionInput) []recommend.Impression {
	out := make([]recommend.Impression, 0, len(inputs))
	for _, input := range inputs {
		impression := recommend.Impression{SceneID: input.MediaID, Position: input.Position}
		if input.Source != nil {
			impression.Source = *input.Source
		}
		if input.Surface != nil {
			impression.Surface = *input.Surface
		}
		if input.Clicked != nil {
			impression.Clicked = *input.Clicked
		}
		out = append(out, impression)
	}
	return out
}

// topTagName returns the name of the user's highest-affinity tag, "" when
// there is none yet.
func (r *queryResolver) topTagName(ctx context.Context) string {
	tagIDs, err := r.recommender.TopTags(ctx, 1)
	if err != nil {
		log.Printf("recommendedBrowse top tag: %v", err)
		return ""
	}
	if len(tagIDs) == 0 {
		return ""
	}
	tag, err := r.repo.GetTag(ctx, tagIDs[0])
	if err != nil || tag == nil {
		return ""
	}
	return tag.Name
}

// runRecommendedBrowse pulls fresh items from every scene:list-capable plugin
// using two fetches per provider — a query list for the top taste
// category and a catalog list for diversity — interleaved category-first.
func (r *queryResolver) runRecommendedBrowse(ctx context.Context, limit *int) ([]*model.PluginSearchResult, error) {
	max := 40
	if limit != nil && *limit > 0 {
		max = *limit
	}

	browsePlugins := adultPlugins(r.registry.WithCapability(plugins.CapabilitySceneList))
	if len(browsePlugins) == 0 {
		return []*model.PluginSearchResult{}, nil
	}

	term := r.topTagName(ctx)

	type pluginPair struct {
		name     string
		category []plugins.DiscoveredItem
		main     []plugins.DiscoveredItem
	}
	ch := make(chan pluginPair, len(browsePlugins))

	for _, plugin := range browsePlugins {
		go func(plugin *plugins.Plugin) {
			pair := pluginPair{name: plugin.Meta.Name}
			if term != "" {
				if res, err := r.runner.List(ctx, plugin, plugins.MediaTypeScene, plugins.ListArgs{Query: term, Limit: max}); err == nil {
					pair.category = res.Items
				} else {
					log.Printf("recommendedBrowse category %q via %s: %v", term, plugin.Meta.Name, err)
				}
			}
			if res, err := r.runner.List(ctx, plugin, plugins.MediaTypeScene, plugins.ListArgs{Limit: max}); err == nil {
				pair.main = res.Items
			} else {
				log.Printf("recommendedBrowse main via %s: %v", plugin.Meta.Name, err)
			}
			ch <- pair
		}(plugin)
	}

	seen := make(map[string]bool)
	toPersist := make([]discoveredItemWithPlugin, 0, len(browsePlugins)*max)
	collect := func(name string, items []plugins.DiscoveredItem) []*model.PluginSearchResult {
		out := make([]*model.PluginSearchResult, 0, len(items))
		for _, item := range items {
			if seen[item.ExternalID] {
				continue
			}
			seen[item.ExternalID] = true
			out = append(out, pluginSearchResult(name, item))
			toPersist = append(toPersist, discoveredItemWithPlugin{plugin: name, item: item})
		}
		return out
	}

	pairs := make([]pluginPair, 0, len(browsePlugins))
	for range browsePlugins {
		pairs = append(pairs, <-ch)
	}

	// Collect all category items first so the shared seen set attributes an item
	// appearing in both lists to the taste category rather than the main page.
	var categoryResults, mainResults []*model.PluginSearchResult
	for _, pair := range pairs {
		categoryResults = append(categoryResults, collect(pair.name, pair.category)...)
	}
	for _, pair := range pairs {
		mainResults = append(mainResults, collect(pair.name, pair.main)...)
	}

	r.persistDiscoveredItems(toPersist)

	results := interleaveResults(categoryResults, mainResults)
	if len(results) > max {
		results = results[:max]
	}
	return results, nil
}

// interleaveResults alternates between the two lists so the feed stays diverse
// while keeping the taste-category items ranked first.
func interleaveResults(primary, secondary []*model.PluginSearchResult) []*model.PluginSearchResult {
	out := make([]*model.PluginSearchResult, 0, len(primary)+len(secondary))
	i, j := 0, 0
	for i < len(primary) || j < len(secondary) {
		if i < len(primary) {
			out = append(out, primary[i])
			i++
		}
		if j < len(secondary) {
			out = append(out, secondary[j])
			j++
		}
	}
	return out
}

// relatedSubscriptionLimit caps the related scenes relatedChanged emits.
const relatedSubscriptionLimit = 48
