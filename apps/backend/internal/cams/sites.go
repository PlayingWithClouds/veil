package cams

import (
	"context"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// PluginSites is the Sites implementation over the plugin registry and runner.
type PluginSites struct {
	registry *plugins.Registry
	runner   *plugins.Runner
}

// NewPluginSites returns the Sites backed by the loaded plugins.
func NewPluginSites(registry *plugins.Registry, runner *plugins.Runner) *PluginSites {
	return &PluginSites{registry: registry, runner: runner}
}

// Sources returns the active plugins with the cam:list capability.
func (sites *PluginSites) Sources() []*plugins.Plugin {
	return sites.registry.WithCapability(plugins.CapabilityCamList)
}

// Plugin returns an active cam plugin by name.
func (sites *PluginSites) Plugin(name string) (*plugins.Plugin, bool) {
	for _, plugin := range sites.Sources() {
		if plugin.Meta.Name == name {
			return plugin, true
		}
	}
	return nil, false
}

// List runs cam:list on the plugin.
func (sites *PluginSites) List(ctx context.Context, plugin *plugins.Plugin, args plugins.CamListArgs) (*plugins.CamListResult, error) {
	return sites.runner.CamList(ctx, plugin, args)
}

// Stream runs cam:stream on the plugin.
func (sites *PluginSites) Stream(ctx context.Context, plugin *plugins.Plugin, username string) (*plugins.ResolveResult, error) {
	return sites.runner.CamStream(ctx, plugin, username)
}
