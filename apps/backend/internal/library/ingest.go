package library

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// IngestFound stores everything a plugin's "<entity>:find" returned for one
// page: the root record first (a video or gallery), then the performers,
// studios and tags it described on their own. It returns the root record.
// The root's source is marked fetched, so the page isn't read again.
func (l *Library) IngestFound(ctx context.Context, plugin string, results []*plugins.ScrapeResult) (schema.RecordRef, error) {
	var root schema.RecordRef
	for index, result := range results {
		ref, err := l.ingestResult(ctx, plugin, result)
		if err != nil {
			return root, err
		}
		if index == 0 {
			root = ref
		}
	}
	if root.Id == "" {
		return root, fmt.Errorf("%s returned nothing to store", plugin)
	}
	return root, nil
}

// ingestResult stores one scrape result and returns what it became.
func (l *Library) ingestResult(ctx context.Context, plugin string, result *plugins.ScrapeResult) (schema.RecordRef, error) {
	switch {
	case result.Scene != nil:
		id, err := l.upsertVideo(ctx, plugin, result.Scene, true)
		return schema.RecordRef{Kind: schema.RecordKindVideo, Id: id}, err
	case result.Gallery != nil:
		id, err := l.upsertGallery(ctx, plugin, result.Gallery, true)
		return schema.RecordRef{Kind: schema.RecordKindGallery, Id: id}, err
	case result.Performer != nil:
		id, err := l.upsertPerformer(ctx, plugin, result.Performer)
		return schema.RecordRef{Kind: schema.RecordKindPerformer, Id: id}, err
	case result.Studio != nil:
		id, err := l.upsertStudio(ctx, plugin, result.Studio)
		return schema.RecordRef{Kind: schema.RecordKindStudio, Id: id}, err
	case result.Tag != nil:
		id, err := l.upsertTag(ctx, plugin, result.Tag)
		return schema.RecordRef{Kind: schema.RecordKindTag, Id: id}, err
	}
	// Standalone images and site collections come back with their models.
	return schema.RecordRef{}, nil
}

// IngestListed stores one item of a listing or search as a stub: what the
// card showed, without reading its page. Videos and galleries become
// records; performers, studios and tags are found or created by name.
func (l *Library) IngestListed(ctx context.Context, plugin string, item plugins.DiscoveredItem) (schema.RecordRef, error) {
	switch item.MediaType {
	case plugins.MediaTypeScene:
		id, err := l.upsertVideo(ctx, plugin, sceneFromListing(item), false)
		return schema.RecordRef{Kind: schema.RecordKindVideo, Id: id}, err
	case plugins.MediaTypeGallery:
		id, err := l.upsertGallery(ctx, plugin, galleryFromListing(item), false)
		return schema.RecordRef{Kind: schema.RecordKindGallery, Id: id}, err
	case plugins.MediaTypePerformer:
		id, err := l.upsertPerformer(ctx, plugin, &plugins.Performer{ExternalID: item.ExternalID, SourceURL: item.SourceURL, Name: item.Title, ImagePath: item.PosterPath})
		return schema.RecordRef{Kind: schema.RecordKindPerformer, Id: id}, err
	case plugins.MediaTypeStudio:
		id, err := l.upsertStudio(ctx, plugin, &plugins.Studio{ExternalID: item.ExternalID, SourceURL: item.SourceURL, Name: item.Title, ImagePath: item.PosterPath})
		return schema.RecordRef{Kind: schema.RecordKindStudio, Id: id}, err
	case plugins.MediaTypeTag:
		id, err := l.upsertTag(ctx, plugin, &plugins.Tag{Name: item.Title})
		return schema.RecordRef{Kind: schema.RecordKindTag, Id: id}, err
	}
	return schema.RecordRef{}, fmt.Errorf("can't store a listed %q", item.MediaType)
}

// sceneFromListing turns a listing card into the scene it stands for.
func sceneFromListing(item plugins.DiscoveredItem) *plugins.Scene {
	return &plugins.Scene{
		ExternalID:    item.ExternalID,
		SourceURL:     item.SourceURL,
		Title:         item.Title,
		Date:          item.Date,
		Duration:      item.DurationSeconds,
		PosterPath:    item.PosterPath,
		PreviewVideo:  item.PreviewVideo,
		PreviewImages: item.PreviewImages,
		Studio:        item.Studio,
		Downloads:     item.Downloads,
	}
}

// galleryFromListing turns a listing card into the gallery it stands for.
func galleryFromListing(item plugins.DiscoveredItem) *plugins.Gallery {
	return &plugins.Gallery{
		ExternalID: item.ExternalID,
		SourceURL:  item.SourceURL,
		Title:      item.Title,
		Date:       item.Date,
		CoverPath:  item.PosterPath,
		Studio:     item.Studio,
	}
}
