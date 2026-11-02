package pipeline

import (
	"context"
	"errors"
	"fmt"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// ErrNoPageListing means a plugin can't list the scenes on a site page: it
// isn't active or doesn't declare scene:list:page.
var ErrNoPageListing = errors.New("plugin can't list site pages")

// ListPage stores the scenes on one site page (a performer's, studio's or
// tag's page, as its Source row gives it) through the plugin that owns the
// page, and returns what they became.
func (p *Pipeline) ListPage(ctx context.Context, pluginName, url string, limit int) ([]schema.RecordRef, error) {
	plugin, found := p.registry.Get(pluginName)
	if !found || !p.registry.IsActive(pluginName) || !plugin.Meta.Has(plugins.CapabilitySceneListPage) {
		return nil, fmt.Errorf("%s: %w", pluginName, ErrNoPageListing)
	}
	result, err := p.runner.List(ctx, plugin, plugins.MediaTypeScene, plugins.ListArgs{URL: url, Limit: limit})
	if err != nil {
		return nil, err
	}
	return p.storeListed(ctx, pluginName, result.Items), nil
}
