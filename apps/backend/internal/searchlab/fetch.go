package searchlab

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// fetchTimeout bounds one plugin search.
const fetchTimeout = 2 * time.Minute

// Fetch runs every query on every configured plugin from pluginDir (installed
// bundles, e.g. DATA_DIR/plugins) and returns the frozen results. A failing
// plugin is recorded in its list instead of failing the run.
func Fetch(ctx context.Context, config *Config, pluginDir string, solver bool) (*Snapshot, error) {
	runner := plugins.NewEmbeddedRunner()
	registry, err := plugins.NewRegistry(pluginDir, runner, plugins.WithSolver(solver))
	if err != nil {
		return nil, fmt.Errorf("load plugins: %w", err)
	}
	snapshot := &Snapshot{FetchedAt: time.Now().UTC()}
	for _, query := range config.Queries {
		for _, pluginName := range config.Plugins {
			plugin, found := registry.Get(pluginName)
			if !found {
				return nil, fmt.Errorf("plugin %q is not installed in %s", pluginName, pluginDir)
			}
			snapshot.Lists = append(snapshot.Lists, fetchList(ctx, runner, plugin, query, config.Limit))
		}
	}
	return snapshot, nil
}

// fetchList runs one query on one plugin.
func fetchList(ctx context.Context, runner *plugins.Runner, plugin *plugins.Plugin, query Query, limit int) List {
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	started := time.Now()
	list := List{QueryID: query.ID, Plugin: plugin.Meta.Name}
	result, err := runner.List(ctx, plugin, plugins.MediaTypeScene, plugins.ListArgs{Query: query.Text, Limit: limit})
	if err != nil {
		list.Error = err.Error()
		log.Printf("%s on %s: %v", query.ID, plugin.Meta.Name, err)
		return list
	}
	list.Items = result.Items
	if len(list.Items) > limit {
		list.Items = list.Items[:limit]
	}
	log.Printf("%s on %s: %d results in %s", query.ID, plugin.Meta.Name, len(list.Items), time.Since(started).Round(time.Millisecond))
	return list
}

// relatedLimit bounds the related results kept per seed.
const relatedLimit = 12

// Expand adds, for every niche query of the snapshot, the related lists of
// its confident hits (see Seeds), replacing earlier expansions. Keyword
// lists stay as they are, so their grades stay valid.
func Expand(ctx context.Context, config *Config, snapshot *Snapshot, pluginDir string, solver bool) error {
	runner := plugins.NewEmbeddedRunner()
	registry, err := plugins.NewRegistry(pluginDir, runner, plugins.WithSolver(solver))
	if err != nil {
		return fmt.Errorf("load plugins: %w", err)
	}
	snapshot.Lists = keywordLists(snapshot.Lists)
	for _, query := range config.Queries {
		lists := snapshot.ListsFor(query.ID)
		if !IsNiche(query, lists) {
			log.Printf("%s: not niche, not expanded", query.ID)
			continue
		}
		for _, seed := range Seeds(query, lists) {
			plugin, found := registry.Get(seed.Plugin)
			if !found {
				return fmt.Errorf("plugin %q is not installed in %s", seed.Plugin, pluginDir)
			}
			snapshot.Lists = append(snapshot.Lists, fetchRelated(ctx, runner, plugin, query, seed))
		}
	}
	return nil
}

// fetchRelated reads a seed's page and keeps the site's related list.
func fetchRelated(ctx context.Context, runner *plugins.Runner, plugin *plugins.Plugin, query Query, seed Seed) List {
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	list := List{QueryID: query.ID, Plugin: plugin.Meta.Name, Kind: KindRelated,
		Seed: ItemKey(seed.Item), SeedTitle: seed.Item.Title, SeedConfidence: seed.Confidence}
	results, err := runner.Find(ctx, plugin, plugins.MediaTypeScene, seed.Item.SourceURL)
	if err != nil {
		list.Error = err.Error()
		log.Printf("%s related of %q: %v", query.ID, seed.Item.Title, err)
		return list
	}
	if len(results) > 0 && results[0].Scene != nil {
		list.Items = results[0].Scene.Related
	}
	if len(list.Items) > relatedLimit {
		list.Items = list.Items[:relatedLimit]
	}
	log.Printf("%s related of %q (confidence %.1f): %d results", query.ID, seed.Item.Title, seed.Confidence, len(list.Items))
	return list
}
