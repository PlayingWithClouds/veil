package graphql

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/api/graphql/model"
)

// findGalleryVideos ranks the videos a gallery's photo set may come from.
// A gallery has no runtime, so title, performers, date and the cover's
// likeness to each poster decide.
func (r *Resolver) findGalleryVideos(ctx context.Context, galleryID string, limit *int) ([]*model.AlikeCandidate, error) {
	gallery, err := r.repo.GetGallery(ctx, galleryID)
	if err != nil {
		return nil, err
	}
	if gallery == nil {
		return []*model.AlikeCandidate{}, nil
	}
	return r.rankAlike(ctx, r.galleryAlikeTarget(ctx, gallery), limit, func(discoveredItemWithPlugin) bool { return false }), nil
}

// galleryAlikeTarget projects a gallery onto the matcher's target shape.
func (r *Resolver) galleryAlikeTarget(ctx context.Context, gallery *model.Gallery) alike.Target {
	target := alike.Target{Title: gallery.Title, Performers: performerNames(gallery.Performers)}
	if gallery.Studio != nil {
		target.Studio = gallery.Studio.Name
	}
	if gallery.Date != nil {
		target.Date = *gallery.Date
	}
	if gallery.CoverPath != nil {
		target.PosterHash = r.hashPoster(ctx, *gallery.CoverPath)
	}
	return target
}

// linkGalleryVideo stores a picked match as a scene and links it to the gallery.
func (r *Resolver) linkGalleryVideo(ctx context.Context, galleryID, pluginName, url string) (*model.Scene, error) {
	sceneID, err := (&mutationResolver{r}).ResolvePluginResult(ctx, pluginName, url, nil)
	if err != nil {
		return nil, err
	}
	scene, err := r.repo.GetScene(ctx, sceneID)
	if err != nil {
		return nil, err
	}
	if scene == nil {
		return nil, fmt.Errorf("linkGalleryVideo: %s is not a video", url)
	}
	if err := r.repo.SetGalleryVideo(ctx, galleryID, sceneID); err != nil {
		return nil, err
	}
	return scene, nil
}

// galleryVideo loads the gallery's linked video, nil when none was picked.
func (r *Resolver) galleryVideo(ctx context.Context, galleryID string) (*model.Scene, error) {
	sceneID, err := r.repo.GalleryVideoID(ctx, galleryID)
	if err != nil || sceneID == "" {
		return nil, err
	}
	return r.repo.GetScene(ctx, sceneID)
}
