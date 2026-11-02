package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestListingKeepsPreviewFrames stores a listing card's preview frames in
// order, blank ones dropped.
func TestListingKeepsPreviewFrames(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	item := plugins.DiscoveredItem{
		MediaType: plugins.MediaTypeScene, ExternalID: "f1", SourceURL: "https://site.test/f1", Title: "Frames",
		PreviewImages: []string{"https://cdn.test/1.jpg", " ", "https://cdn.test/2.jpg"},
	}
	ref, err := library.IngestListed(ctx, "site", item)
	if err != nil {
		t.Fatal(err)
	}
	video, err := library.Video(ctx, ref.Id)
	if err != nil {
		t.Fatal(err)
	}
	if len(video.PreviewFrames) != 2 || video.PreviewFrames[1].Uri != "https://cdn.test/2.jpg" {
		t.Errorf("previewFrames = %+v", video.PreviewFrames)
	}
}

// TestFiltersByResolution counts a portrait 1080×1920 video as 1080p and
// leaves out videos of unknown size.
func TestFiltersByResolution(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	portrait := foundScene()
	portrait[0].Scene.Width, portrait[0].Scene.Height = 1080, 1920
	tall, err := library.IngestFound(ctx, "site", portrait)
	if err != nil {
		t.Fatal(err)
	}
	small := foundScene()
	small[0].Scene.ExternalID, small[0].Scene.Title, small[0].Scene.Duration = "v2", "Night swim", 1500
	small[0].Scene.Width, small[0].Scene.Height = 1280, 720
	if _, err := library.IngestFound(ctx, "site", small); err != nil {
		t.Fatal(err)
	}
	unknown := foundScene()
	unknown[0].Scene.ExternalID, unknown[0].Scene.Title, unknown[0].Scene.Duration = "v3", "Beach walk", 2400
	if _, err := library.IngestFound(ctx, "site", unknown); err != nil {
		t.Fatal(err)
	}

	minimum := int32(1080)
	assertVideos(t, library, "1080p and up", schema.Filter{MinHeight: &minimum}, tall.Id)
}

// TestTagCounts counts tags across a filter's results, most common first,
// and leaves out the tags the filter requires.
func TestTagCounts(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	first := foundScene()
	first[0].Scene.Tags = []string{"Pool", "Outdoor"}
	if _, err := library.IngestFound(ctx, "site", first); err != nil {
		t.Fatal(err)
	}
	second := foundScene()
	second[0].Scene.ExternalID, second[0].Scene.Title, second[0].Scene.Duration = "v2", "Night swim", 1500
	second[0].Scene.Tags = []string{"Pool", "Night"}
	if _, err := library.IngestFound(ctx, "site", second); err != nil {
		t.Fatal(err)
	}

	counts, err := library.TagCounts(ctx, schema.RecordKindVideo, schema.Filter{}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(counts) != 3 || counts[0].Name != "Pool" || counts[0].Count != 2 {
		t.Errorf("counts = %+v", counts)
	}

	narrowed, err := library.TagCounts(ctx, schema.RecordKindVideo, schema.Filter{IncludeTagIds: []string{counts[0].Id}}, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(narrowed) != 2 {
		t.Errorf("with Pool required, counts = %+v", narrowed)
	}
}
