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
	case result.Audio != nil:
		id, err := l.upsertAudio(ctx, plugin, result.Audio, true)
		return schema.RecordRef{Kind: schema.RecordKindAudio, Id: id}, err
	case result.Story != nil:
		id, err := l.upsertStory(ctx, plugin, result.Story, true)
		return schema.RecordRef{Kind: schema.RecordKindStory, Id: id}, err
	case result.Post != nil:
		id, err := l.upsertPost(ctx, plugin, result.Post, true)
		return schema.RecordRef{Kind: schema.RecordKindPost, Id: id}, err
	case result.Collection != nil:
		id, err := l.upsertGroup(ctx, plugin, result.Collection)
		return schema.RecordRef{Kind: schema.RecordKindGroup, Id: id}, err
	}
	// Standalone images have no record of their own; they come with galleries.
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
	case plugins.MediaTypeAudio:
		id, err := l.upsertAudio(ctx, plugin, &plugins.Audio{
			ExternalID: item.ExternalID, SourceURL: item.SourceURL, Title: item.Title, Date: item.Date,
			Duration: item.DurationSeconds, CoverPath: item.PosterPath, Performers: uploaderOf(item), Downloads: item.Downloads,
		}, false)
		return schema.RecordRef{Kind: schema.RecordKindAudio, Id: id}, err
	case plugins.MediaTypeStory:
		id, err := l.upsertStory(ctx, plugin, &plugins.Story{
			ExternalID: item.ExternalID, SourceURL: item.SourceURL, Title: item.Title, Date: item.Date,
			CoverPath: item.PosterPath, Performers: uploaderOf(item),
		}, false)
		return schema.RecordRef{Kind: schema.RecordKindStory, Id: id}, err
	case plugins.MediaTypePost:
		id, err := l.upsertPost(ctx, plugin, postFromListing(item), false)
		return schema.RecordRef{Kind: schema.RecordKindPost, Id: id}, err
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

// postFromListing turns a listing card into the post it stands for; a
// creator platform's card names the creator as its channel (Studio).
func postFromListing(item plugins.DiscoveredItem) *plugins.Post {
	post := &plugins.Post{ExternalID: item.ExternalID, SourceURL: item.SourceURL, Title: item.Title, Date: item.Date, CoverPath: item.PosterPath}
	if item.Studio != nil {
		post.Creator = plugins.ScenePerformer{Name: item.Studio.Name, ExternalID: item.Studio.ExternalID, SourceURL: item.Studio.SourceURL}
	}
	return post
}

// uploaderOf is who an audio or story card names as its channel (Studio):
// on those sites the uploader is the voice or the author, so credited as
// a performer, like a creator post's creator.
func uploaderOf(item plugins.DiscoveredItem) []plugins.ScenePerformer {
	if item.Studio == nil || item.Studio.Name == "" {
		return nil
	}
	return []plugins.ScenePerformer{{Name: item.Studio.Name, ExternalID: item.Studio.ExternalID, SourceURL: item.Studio.SourceURL}}
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
