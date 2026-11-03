package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/schema"
)

// countDefaults counts the categories of list and their subcategories.
func countDefaults(list []defaultCategory) int {
	count := 0
	for _, entry := range list {
		count += 1 + countDefaults(entry.children)
	}
	return count
}

// TestSeedDefaultCategoriesOnce seeds every default once, groups matching
// only through their subcategories; a second seed changes nothing, and
// AddDefaultCategories brings back only what was deleted.
func TestSeedDefaultCategoriesOnce(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	if err := library.SeedDefaultCategories(ctx); err != nil {
		t.Fatal(err)
	}
	categories, err := library.Categories(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(categories) != countDefaults(defaultCategories) {
		t.Fatalf("seeded %d categories, want %d", len(categories), countDefaults(defaultCategories))
	}
	body, _ := library.CategoryBySlug(ctx, "body")
	blonde, _ := library.CategoryBySlug(ctx, "blonde")
	if body == nil || len(body.MatchTerms) != 0 {
		t.Errorf("group Body = %+v, want no terms of its own", body)
	}
	if blonde == nil || blonde.Parent == nil || len(blonde.MatchTerms) < 2 {
		t.Errorf("Blonde = %+v, want under Hair with its spellings", blonde)
	}

	if err := library.DeleteCategory(ctx, blonde.Id); err != nil {
		t.Fatal(err)
	}
	if err := library.SeedDefaultCategories(ctx); err != nil {
		t.Fatal(err)
	}
	if restored, _ := library.CategoryBySlug(ctx, "blonde"); restored != nil {
		t.Error("a second seed brought back a deleted category")
	}
	added, err := library.AddDefaultCategories(ctx)
	if err != nil || added != 1 {
		t.Errorf("AddDefaultCategories added %d (%v), want 1", added, err)
	}
}

// TestDefaultCategoryMatchesSiteSpelling finds a video tagged "Pool" through
// the default Pool category.
func TestDefaultCategoryMatchesSiteSpelling(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	if err := library.SeedDefaultCategories(ctx); err != nil {
		t.Fatal(err)
	}
	pool, _ := library.CategoryBySlug(ctx, "pool")
	setting, _ := library.CategoryBySlug(ctx, "setting")
	assertVideos(t, library, "Pool", schema.Filter{CategoryIds: []string{pool.Id}}, ref.Id)
	assertVideos(t, library, "Setting group", schema.Filter{CategoryIds: []string{setting.Id}}, ref.Id)
}

// TestCategoryTilesCountOnce counts the video tagged "Outdoor" and "Pool" in
// both categories, once in their Setting group, and not in Body.
func TestCategoryTilesCountOnce(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	if _, err := library.IngestFound(ctx, "site", foundScene()); err != nil {
		t.Fatal(err)
	}
	if err := library.SeedDefaultCategories(ctx); err != nil {
		t.Fatal(err)
	}
	tiles, err := library.CategoryTiles(ctx, schema.MediaKindVideo)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{}
	for _, tile := range tiles {
		counts[tile.Category.Slug] = tile.Count
	}
	want := map[string]int{"pool": 1, "outdoor": 1, "setting": 1, "body": 0}
	for slug, count := range want {
		if counts[slug] != count {
			t.Errorf("%s counts %d, want %d", slug, counts[slug], count)
		}
	}
}
