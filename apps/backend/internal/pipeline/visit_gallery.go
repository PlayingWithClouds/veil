package pipeline

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// VisitGallery is called when a gallery page is opened. Search and browse only
// ingest gallery stubs, so a gallery without images has its page fetched from
// its origin plugin (gallery:find), which lands the images. Concurrent visits
// of one gallery fetch once.
func (o *Orchestrator) VisitGallery(ctx context.Context, galleryID string) error {
	hasImages, err := o.ingestSvc.GalleryHasImages(ctx, galleryID)
	if err != nil || hasImages {
		return err
	}
	if _, alreadyFetching := o.detailInflight.LoadOrStore(galleryID, struct{}{}); alreadyFetching {
		return nil
	}
	defer o.detailInflight.Delete(galleryID)

	plugin, sourceURL, err := o.origin(ctx, galleryID)
	if err != nil {
		return err
	}
	results, err := o.runner.Find(ctx, plugin, plugins.MediaTypeGallery, sourceURL)
	if err != nil {
		return fmt.Errorf("find %s via %s: %w", sourceURL, plugin.Meta.Name, err)
	}
	return o.ingestScrapeResult(ctx, plugin.Meta.Name, sourceURL, results)
}
