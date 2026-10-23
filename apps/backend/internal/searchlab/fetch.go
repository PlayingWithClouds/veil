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
