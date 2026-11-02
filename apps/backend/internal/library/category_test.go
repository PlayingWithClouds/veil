package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/schema"
)

// TestCategoryFilterIncludesSubcategories maps a video's tag to a
// subcategory and finds the video through the parent category.
func TestCategoryFilterIncludesSubcategories(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	video, _ := library.Video(ctx, ref.Id)
	outdoors := "Outdoors"
	parent, err := library.CreateCategory(ctx, CategoryEdit{Name: &outdoors})
	if err != nil {
		t.Fatal(err)
	}
	pool := "Pool & beach"
	tags := video.Tags[:1]
	child, err := library.CreateCategory(ctx, CategoryEdit{Name: &pool, ParentID: &parent.Id, MatchTagIDs: &tags})
	if err != nil {
		t.Fatal(err)
	}
	if child.Slug != "pool-beach" {
		t.Errorf("slug = %q", child.Slug)
	}
	assertVideos(t, library, "parent category", schema.Filter{CategoryIds: []string{parent.Id}}, ref.Id)
	if _, err := library.EditCategory(ctx, parent.Id, CategoryEdit{ParentID: &child.Id}); err == nil {
		t.Error("a category was put inside its own child")
	}
	if _, err := library.CreateCategory(ctx, CategoryEdit{Name: &pool}); err == nil {
		t.Error("a second category got the same slug")
	}
}
