package media

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
)

// GalleryVideoID returns the scene id linked as the gallery's video, "" when
// none was picked.
func (r *Repository) GalleryVideoID(ctx context.Context, galleryID string) (string, error) {
	ids, err := r.database.Strings(ctx, `SELECT scene FROM gallery_video WHERE gallery = $gallery`, db.Vars{"gallery": galleryID})
	if err != nil {
		return "", fmt.Errorf("gallery video: %w", err)
	}
	if len(ids) == 0 {
		return "", nil
	}
	return ids[0], nil
}

// SetGalleryVideo links sceneID as the gallery's video, replacing an earlier link.
func (r *Repository) SetGalleryVideo(ctx context.Context, galleryID, sceneID string) error {
	_, err := r.database.Exec(ctx,
		`INSERT INTO gallery_video (gallery, scene) VALUES ($gallery, $scene)
		 ON CONFLICT (gallery) DO UPDATE SET scene = excluded.scene, created_at = excluded.created_at`,
		db.Vars{"gallery": galleryID, "scene": sceneID})
	if err != nil {
		return fmt.Errorf("set gallery video: %w", err)
	}
	return nil
}
