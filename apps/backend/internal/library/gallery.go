package library

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// upsertGallery stores a plugin gallery, like upsertVideo does a scene, and
// replaces its images when the page listed them.
func (l *Library) upsertGallery(ctx context.Context, plugin string, gallery *plugins.Gallery, fetched bool) (string, error) {
	site, ok := identity(plugin, gallery.ExternalID, gallery.SourceURL)
	if !ok {
		return "", fmt.Errorf("%s returned a gallery without id or URL", plugin)
	}
	fields, err := l.galleryFields(ctx, plugin, gallery, fetched)
	if err != nil {
		return "", err
	}
	source, err := l.findSource(ctx, site)
	if err != nil {
		return "", err
	}
	var id, sourceID string
	if source == nil {
		id, sourceID, err = l.createGallery(ctx, site, fields, gallery)
	} else {
		id, sourceID = source.SubjectId, source.Id
		_, err = l.recorder.Update(ctx, schema.RecordKindGallery, id, fields, provenance.Plugin(plugin, sourceID))
	}
	if err != nil {
		return "", err
	}
	return id, l.afterGalleryWrite(ctx, id, sourceID, gallery, fetched)
}

// createGallery inserts a new gallery and its first source, which owns the
// fields it filled.
func (l *Library) createGallery(ctx context.Context, site siteIdentity, fields db.Vars, gallery *plugins.Gallery) (string, string, error) {
	filled := columnsOf(fields)
	fields["status"] = string(schema.MediaStatusAvailable)
	fillListDefaults(fields, "performers", "tags")
	if _, hasTitle := fields["title"]; !hasTitle {
		fields["title"] = site.URL
	}
	if _, hasCount := fields["imageCount"]; !hasCount {
		fields["imageCount"] = 0
	}
	id, err := l.database.Insert(ctx, tableGallery, fields)
	if err != nil {
		return "", "", err
	}
	sourceID, err := l.addSource(ctx, schema.RecordKindGallery, id, site, gallery.Title, assetOf(gallery.CoverPath))
	if err != nil {
		return "", "", err
	}
	return id, sourceID, l.recorder.Claim(ctx, schema.RecordKindGallery, id, filled, provenance.Plugin(site.Plugin, sourceID))
}

// afterGalleryWrite marks the source fetched and replaces the images.
func (l *Library) afterGalleryWrite(ctx context.Context, id, sourceID string, gallery *plugins.Gallery, fetched bool) error {
	if !fetched {
		return nil
	}
	if err := l.markFetched(ctx, sourceID); err != nil {
		return err
	}
	if len(gallery.Images) == 0 {
		return nil
	}
	return l.replaceImages(ctx, id, gallery.Images)
}

// galleryFields are the columns a gallery fills; a listing card leaves the
// credits alone.
func (l *Library) galleryFields(ctx context.Context, plugin string, gallery *plugins.Gallery, fetched bool) (db.Vars, error) {
	fields := db.Vars{}
	setText(fields, "title", gallery.Title)
	setText(fields, "details", gallery.Details)
	setText(fields, "date", isoDate(gallery.Date))
	setAsset(fields, "cover", galleryCover(gallery))
	if gallery.ImageCount > 0 {
		fields["imageCount"] = gallery.ImageCount
	}
	if err := l.setStudio(ctx, fields, plugin, gallery.Studio); err != nil {
		return nil, err
	}
	if !fetched {
		return fields, nil
	}
	return fields, l.setCredits(ctx, fields, performerNames(gallery.Performers), gallery.Tags)
}

// galleryCover is the gallery's cover, else its first image.
func galleryCover(gallery *plugins.Gallery) string {
	if gallery.CoverPath != "" || len(gallery.Images) == 0 {
		return gallery.CoverPath
	}
	return gallery.Images[0].FilePath
}

// replaceImages swaps the gallery's images for the ones its page lists and
// updates the count. Images are the site's, so the old rows go for good.
func (l *Library) replaceImages(ctx context.Context, galleryID string, images []plugins.Image) error {
	return l.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM "Image" WHERE "gallery" = $gallery`, db.Vars{"gallery": galleryID}); err != nil {
			return err
		}
		for position, image := range images {
			if err := insertImage(ctx, tx, galleryID, position, image); err != nil {
				return err
			}
		}
		return tx.Update(ctx, tableGallery, galleryID, db.Vars{"imageCount": len(images)})
	})
}

// insertImage stores one gallery image at position.
func insertImage(ctx context.Context, tx *db.DB, galleryID string, position int, image plugins.Image) error {
	row := schema.Image{Gallery: galleryID, Position: int32(position), Asset: schema.Asset{Uri: image.FilePath}}
	if image.Width > 0 {
		width := int32(image.Width)
		row.Width = &width
	}
	if image.Height > 0 {
		height := int32(image.Height)
		row.Height = &height
	}
	_, err := tx.Insert(ctx, tableImage, row)
	return err
}
