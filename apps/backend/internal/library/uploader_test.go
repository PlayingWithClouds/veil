package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestSceneUploaderStaysOnSource stores a scene's uploader on its source, not
// as a studio, and keeps it when a later card names nobody.
func TestSceneUploaderStaysOnSource(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	scene := &plugins.Scene{
		ExternalID: "v1", SourceURL: "https://tube.test/v1", Title: "Clip",
		Uploader: &plugins.Uploader{Name: "  Jane  ", URL: "https://tube.test/users/jane", ExternalID: "jane"},
	}
	found := []*plugins.ScrapeResult{{Type: plugins.MediaTypeScene, Scene: scene}}
	ref, err := library.IngestFound(ctx, "tube", found)
	if err != nil {
		t.Fatal(err)
	}
	video, err := library.Video(ctx, ref.Id)
	if err != nil {
		t.Fatal(err)
	}
	if video.Studio != nil {
		t.Errorf("studio = %v, want none: the uploader is not a studio", video.Studio)
	}
	if _, err := library.IngestListed(ctx, "tube", plugins.DiscoveredItem{
		MediaType: plugins.MediaTypeScene, ExternalID: "v1", SourceURL: "https://tube.test/v1", Title: "Clip",
	}); err != nil {
		t.Fatal(err)
	}
	sources, err := library.Sources(ctx, schema.RecordKindVideo, ref.Id)
	if err != nil {
		t.Fatal(err)
	}
	uploader := sources[0].Uploader
	if len(sources) != 1 || uploader == nil || uploader.Name != "Jane" || *uploader.Url != "https://tube.test/users/jane" {
		t.Fatalf("source uploader = %+v, want Jane with her page", uploader)
	}
}

// TestSetUploaderWhereMissing names the uploader on listed videos whose
// cards didn't, and leaves one an earlier read named.
func TestSetUploaderWhereMissing(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	named, err := library.IngestListed(ctx, "tube", plugins.DiscoveredItem{
		MediaType: plugins.MediaTypeScene, ExternalID: "a", SourceURL: "https://tube.test/a", Title: "A",
		Uploader: &plugins.Uploader{Name: "Other"},
	})
	if err != nil {
		t.Fatal(err)
	}
	bare, err := library.IngestListed(ctx, "tube", plugins.DiscoveredItem{
		MediaType: plugins.MediaTypeScene, ExternalID: "b", SourceURL: "https://tube.test/b", Title: "B",
	})
	if err != nil {
		t.Fatal(err)
	}
	uploader := plugins.Uploader{Name: "Jane", URL: "https://tube.test/users/jane"}
	if err := library.SetUploaderWhereMissing(ctx, "tube", []string{named.Id, bare.Id}, uploader); err != nil {
		t.Fatal(err)
	}
	for id, want := range map[string]string{named.Id: "Other", bare.Id: "Jane"} {
		sources, err := library.Sources(ctx, schema.RecordKindVideo, id)
		if err != nil {
			t.Fatal(err)
		}
		if got := sources[0].Uploader; got == nil || got.Name != want {
			t.Errorf("uploader of %s = %+v, want %s", id, got, want)
		}
	}
}
