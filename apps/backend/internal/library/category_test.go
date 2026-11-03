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

// TestCategoryTermsMatchTagNames finds a video through a term its tag's name
// folds to, in another spelling, and not through an unrelated term.
func TestCategoryTermsMatchTagNames(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	water, swimming := "Water", "Swimming"
	waterTerms := []string{"POOL", "  ", "pool", "beach"}
	inWater, err := library.CreateCategory(ctx, CategoryEdit{Name: &water, MatchTerms: &waterTerms})
	if err != nil {
		t.Fatal(err)
	}
	if len(inWater.MatchTerms) != 2 {
		t.Errorf("terms = %v, want pool and beach once each", inWater.MatchTerms)
	}
	swimmingTerms := []string{"swimsuit"}
	other, err := library.CreateCategory(ctx, CategoryEdit{Name: &swimming, MatchTerms: &swimmingTerms})
	if err != nil {
		t.Fatal(err)
	}
	assertVideos(t, library, "matching term", schema.Filter{CategoryIds: []string{inWater.Id}}, ref.Id)
	assertVideos(t, library, "unrelated term", schema.Filter{CategoryIds: []string{other.Id}})
}

// TestDeleteCategoryMovesChildrenUp deletes a category: its child moves to
// the top and its slug is free again.
func TestDeleteCategoryMovesChildrenUp(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	outdoors, pool := "Outdoors", "Pool"
	parent, err := library.CreateCategory(ctx, CategoryEdit{Name: &outdoors})
	if err != nil {
		t.Fatal(err)
	}
	child, err := library.CreateCategory(ctx, CategoryEdit{Name: &pool, ParentID: &parent.Id})
	if err != nil {
		t.Fatal(err)
	}
	if err := library.DeleteCategory(ctx, parent.Id); err != nil {
		t.Fatal(err)
	}
	moved, err := library.Category(ctx, child.Id)
	if err != nil || moved == nil || moved.Parent != nil {
		t.Errorf("child after delete = %+v, %v", moved, err)
	}
	if _, err := library.CreateCategory(ctx, CategoryEdit{Name: &outdoors}); err != nil {
		t.Errorf("the deleted category's name stayed taken: %v", err)
	}
}
