package ingest

import (
	"context"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// ingestImage upserts a standalone content image by source_url (falling back to
// file_path) and resolves its studio/performers/tags relations. The row is
// marked content = true so it surfaces in browse/search, unlike asset images.
func (s *Service) ingestImage(ctx context.Context, pluginName string, image *plugins.ImageContent) (*IngestResult, error) {
	identity, field := image.SourceURL, "source_url"
	if identity == "" {
		identity, field = image.FilePath, "file_path"
	}

	obs := s.baseObs(image, pluginName, image.SourceURL)
	// The NDJSON discriminator ("image") collides with the image table's asset
	// type column; standalone content images use their own asset type.
	obs["type"] = "standalone"
	obs["content"] = true

	result, err := s.upsert(ctx, "image", field, identity, obs, pluginName)
	if err != nil {
		return nil, err
	}

	// applyMerge only fills empty columns, so promote an existing asset row
	// (content = false) explicitly.
	if !result.IsNew {
		if err := s.setField(ctx, result.CanonicalID, "content", true); err != nil {
			return nil, err
		}
	}

	s.linkStudio(ctx, "image", result.CanonicalID, image.Studio, pluginName)
	s.linkPerformers(ctx, "image", result.CanonicalID, image.Performers, pluginName)
	s.linkTags(ctx, "image", result.CanonicalID, image.Tags, pluginName)
	return result, nil
}
