package pipeline

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// ErrNoUploaderListing means a plugin can't list an uploader's videos: it
// isn't active or doesn't declare uploader:list.
var ErrNoUploaderListing = errors.New("plugin can't list an uploader's videos")

// ListUploader stores one page of the videos on an uploader's channel page
// through the plugin that reported them, names the uploader on the ones whose
// cards didn't, and returns the videos they became.
func (p *Pipeline) ListUploader(ctx context.Context, pluginName string, uploader plugins.Uploader, limit, offset int) ([]schema.RecordRef, error) {
	plugin, found := p.registry.Get(pluginName)
	if !found || !p.registry.IsActive(pluginName) || !plugin.Meta.Has(plugins.CapabilityUploaderList) {
		return nil, fmt.Errorf("%s: %w", pluginName, ErrNoUploaderListing)
	}
	result, err := p.runner.ListUploader(ctx, plugin, plugins.ListArgs{URL: uploader.URL, Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	videos := videosOf(p.storeListed(ctx, pluginName, result.Items))
	ids := make([]string, 0, len(videos))
	for _, video := range videos {
		ids = append(ids, video.Id)
	}
	if err := p.library.SetUploaderWhereMissing(ctx, pluginName, ids, uploader); err != nil {
		log.Printf("name uploader %q on %s videos: %v", uploader.Name, pluginName, err)
	}
	return videos, nil
}

// videosOf keeps the video references.
func videosOf(refs []schema.RecordRef) []schema.RecordRef {
	videos := make([]schema.RecordRef, 0, len(refs))
	for _, ref := range refs {
		if ref.Kind == schema.RecordKindVideo {
			videos = append(videos, ref)
		}
	}
	return videos
}
