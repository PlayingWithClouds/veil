package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// relatedItem is a listing card a scene page shows as related.
func relatedItem(id, title string) plugins.DiscoveredItem {
	return plugins.DiscoveredItem{MediaType: plugins.MediaTypeScene, ExternalID: id, SourceURL: "https://site.test/" + id, Title: title}
}

// TestSiteRelated stores a scene page's related videos as stubs, lists them
// in the site's order without the scene itself, and replaces them when the
// page is read again.
func TestSiteRelated(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	results := foundScene()
	results[0].Scene.Related = []plugins.DiscoveredItem{
		relatedItem("r1", "First"), relatedItem("v1", "Itself"), relatedItem("r2", "Second"),
	}
	ref, err := library.IngestFound(ctx, "site", results)
	if err != nil {
		t.Fatal(err)
	}
	videos, err := library.SiteRelated(ctx, ref.Id, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(videos) != 2 || videos[0].Title != "First" || videos[1].Title != "Second" {
		t.Fatalf("related = %+v", videos)
	}

	results = foundScene()
	results[0].Scene.Related = []plugins.DiscoveredItem{relatedItem("r3", "Third")}
	if _, err := library.IngestFound(ctx, "site", results); err != nil {
		t.Fatal(err)
	}
	videos, _ = library.SiteRelated(ctx, ref.Id, 10, 0)
	if len(videos) != 1 || videos[0].Title != "Third" {
		t.Fatalf("related after a new read = %+v", videos)
	}
}
