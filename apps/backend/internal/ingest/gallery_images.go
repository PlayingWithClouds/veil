package ingest

import (
	"context"
	"fmt"
	"log"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// syncGalleryImagesLogged materializes a gallery's scraped images into the image
// table and refreshes its image_count. Best-effort: failures are logged, not
// propagated, so a gallery still ingests when its images can't be written.
func (s *Service) syncGalleryImagesLogged(ctx context.Context, canonicalID, pluginName string, images []plugins.Image) {
	if err := s.syncGalleryImages(ctx, canonicalID, pluginName, images); err != nil {
		log.Printf("ingest: sync gallery images for %s: %v", canonicalID, err)
	}
}

// syncGalleryImages upserts the freshly scraped image set keyed by file_path
// and updates the gallery's image_count. Rows are content images (browsable,
// taggable), so a re-scrape must not wipe user data: existing rows are updated
// in place, and rows missing from the new set are only deleted when they carry
// no tags or performer links.
func (s *Service) syncGalleryImages(ctx context.Context, canonicalID, pluginName string, images []plugins.Image) error {
	galleryRID, err := db.ParseRecordID(canonicalID)
	if err != nil {
		return fmt.Errorf("invalid gallery id %q: %w", canonicalID, err)
	}

	scrapedPaths := make([]string, 0, len(images))
	count := 0
	for index, image := range images {
		if image.FilePath == "" {
			continue
		}
		count++
		scrapedPaths = append(scrapedPaths, image.FilePath)
		if err := s.upsertGalleryImage(ctx, *galleryRID, image, index); err != nil {
			return err
		}
	}

	if _, err := s.database.Exec(ctx,
		`DELETE FROM image WHERE media = $media AND type = 'gallery'
			AND file_path NOT IN (SELECT value FROM json_each($paths))
			AND json_array_length(tags) = 0 AND json_array_length(performers) = 0`,
		db.Vars{"media": *galleryRID, "paths": scrapedPaths}); err != nil {
		return fmt.Errorf("prune gallery images: %w", err)
	}

	if err := s.database.Merge(ctx, *galleryRID, map[string]any{"image_count": count}); err != nil {
		return fmt.Errorf("update image_count: %w", err)
	}
	return nil
}

// upsertGalleryImage updates the row matching this file_path or creates it.
func (s *Service) upsertGalleryImage(ctx context.Context, galleryRID db.RecordID, image plugins.Image, index int) error {
	imageType := image.Type
	if imageType == "" {
		imageType = "gallery"
	}
	position := image.Position
	if position == 0 {
		position = index
	}
	fields := map[string]any{
		"media":        galleryRID,
		"type":         imageType,
		"content":      true,
		"width":        image.Width,
		"height":       image.Height,
		"aspect_ratio": image.AspectRatio,
		"position":     position,
	}

	if existingID, found := s.findIDBy(ctx, "image", "file_path", image.FilePath); found {
		if err := s.database.Merge(ctx, existingID, fields); err != nil {
			return fmt.Errorf("update gallery image: %w", err)
		}
		return nil
	}

	fields["file_path"] = image.FilePath
	if _, err := s.database.Insert(ctx, "image", fields); err != nil {
		return fmt.Errorf("create gallery image: %w", err)
	}
	return nil
}
