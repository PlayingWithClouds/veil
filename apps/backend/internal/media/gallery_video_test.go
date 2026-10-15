package media

import (
	"context"
	"testing"
)

func TestGalleryVideoLinkReplacesThePick(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestRepository(t)

	if id, err := repo.GalleryVideoID(ctx, "gallery:set"); err != nil || id != "" {
		t.Fatalf("unlinked gallery: %q, %v", id, err)
	}
	for _, sceneID := range []string{"scene:first", "scene:second"} {
		if err := repo.SetGalleryVideo(ctx, "gallery:set", sceneID); err != nil {
			t.Fatalf("link %s: %v", sceneID, err)
		}
	}
	if id, err := repo.GalleryVideoID(ctx, "gallery:set"); err != nil || id != "scene:second" {
		t.Fatalf("linked video = %q, %v; want the later pick", id, err)
	}
}
