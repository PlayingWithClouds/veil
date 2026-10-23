package search

import (
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/recommend"
)

// TestLibraryHitReadsRowAndIndexText checks that a stored scene becomes a hit
// with the numbers clients sort by and the names behind the full-text match.
func TestLibraryHitReadsRowAndIndexText(t *testing.T) {
	hit := libraryHit(plugins.MediaTypeScene, "scene:one", map[string]any{
		"title": "Lagoon", "source_url": "https://a.test/1", "duration_seconds": int64(600),
		"view_count": int64(1200), "rating": 7.5, "date": "2025-01-02",
		"search_performers": "Mia Rose\nKim Lee", "search_studio": "Blue Films", "search_tags": "Outdoor",
	})
	if hit.Item.DurationSeconds != 600 || hit.Item.ViewCount != 1200 || hit.Item.Rating != 7.5 || hit.Item.Date != "2025-01-02" {
		t.Fatalf("item = %+v", hit.Item)
	}
	if len(hit.Performers) != 2 || hit.Performers[1] != "Kim Lee" || hit.Studio != "Blue Films" || len(hit.Tags) != 1 {
		t.Fatalf("names = %+v", hit)
	}
}

// TestLibraryHitLabelsEntitiesByName checks performers label with `name`.
func TestLibraryHitLabelsEntitiesByName(t *testing.T) {
	hit := libraryHit(plugins.MediaTypePerformer, "performer:one", map[string]any{"name": "Mia Rose"})
	if hit.Item.Title != "Mia Rose" || hit.Performers != nil {
		t.Fatalf("hit = %+v", hit)
	}
}

// TestPersonalAffinityNeedsProfile checks the term is zero without a profile
// and uses the listing site once there is one.
func TestPersonalAffinityNeedsProfile(t *testing.T) {
	handler := &Handler{}
	scene := plugins.DiscoveredItem{MediaType: plugins.MediaTypeScene}
	if got := handler.personalAffinity(t.Context(), nil, libraryHit(plugins.MediaTypeScene, "", nil)); got != 0 {
		t.Fatalf("no profile = %v", got)
	}
	profile := &recommend.Profile{Sites: map[string]float64{"alpha": 1}}
	hit := libraryHit(plugins.MediaTypeScene, "", nil)
	hit.Source, hit.Item = "alpha", scene
	if got := handler.personalAffinity(t.Context(), profile, hit); got <= 0 {
		t.Fatalf("liked site = %v", got)
	}
}
