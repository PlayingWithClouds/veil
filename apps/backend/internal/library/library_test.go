package library

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// openTestLibrary opens a library on a fresh migrated database.
func openTestLibrary(t *testing.T) *Library {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return New(database)
}

// foundScene is a scene page as a plugin's scene:find returns it.
func foundScene() []*plugins.ScrapeResult {
	return []*plugins.ScrapeResult{{
		Type: plugins.MediaTypeScene,
		Scene: &plugins.Scene{
			ExternalID: "v1",
			SourceURL:  "https://site.test/v1",
			Title:      "Pool day",
			Date:       "2024-05-01T10:00:00Z",
			Duration:   600,
			PosterPath: "https://cdn.test/v1.jpg",
			Studio:     &plugins.StudioRef{Name: "Sunny Studio"},
			Performers: []plugins.ScenePerformer{{Name: "Ada Lane"}, {Name: "ada lane"}, {Name: "Bea"}},
			Tags:       []string{"Outdoor", "outdoor", "Pool"},
			Downloads:  []plugins.Download{{URL: "https://cdn.test/v1.m3u8", Quality: "1080p"}},
		},
	}}
}

// TestIngestFoundScene stores a scene page with its studio, credits, tags,
// source and stream.
func TestIngestFoundScene(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	video, err := library.Video(ctx, ref.Id)
	if err != nil || video == nil {
		t.Fatalf("video = %v, err %v", video, err)
	}
	if video.Title != "Pool day" || video.Date == nil || *video.Date != "2024-05-01" || *video.DurationSeconds != 600 {
		t.Errorf("video = %+v", video)
	}
	if len(video.Performers) != 2 || len(video.Tags) != 2 || video.Studio == nil {
		t.Errorf("credits: performers %v, tags %v, studio %v", video.Performers, video.Tags, video.Studio)
	}
	sources, err := library.Sources(ctx, schema.RecordKindVideo, ref.Id)
	if err != nil || len(sources) != 1 || sources[0].FetchedAt == nil {
		t.Fatalf("sources = %+v, err %v", sources, err)
	}
	streams, err := library.Streams(ctx, schema.RecordKindVideo, ref.Id)
	if err != nil || len(streams) != 1 || streams[0].MimeType != "application/x-mpegURL" {
		t.Fatalf("streams = %+v, err %v", streams, err)
	}
}

// TestIngestTwiceUpdatesTheSameVideo re-ingests a page and expects one video,
// one source and one stream.
func TestIngestTwiceUpdatesTheSameVideo(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	first, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	again := foundScene()
	again[0].Scene.Title = "Pool day (remastered)"
	second, err := library.IngestFound(ctx, "site", again)
	if err != nil {
		t.Fatal(err)
	}
	if first.Id != second.Id {
		t.Fatalf("second ingest made a new video: %s vs %s", first.Id, second.Id)
	}
	videos, _ := library.Videos(ctx, MediaQuery{})
	streams, _ := library.Streams(ctx, schema.RecordKindVideo, first.Id)
	if len(videos) != 1 || videos[0].Title != "Pool day (remastered)" || len(streams) != 1 {
		t.Errorf("videos %d (%q), streams %d", len(videos), videos[0].Title, len(streams))
	}
}

// TestListedStubBecomesTheFoundVideo stores a listing card, then its page,
// and expects the page to fill the same record.
func TestListedStubBecomesTheFoundVideo(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	stub, err := library.IngestListed(ctx, "site", plugins.DiscoveredItem{MediaType: plugins.MediaTypeScene, ExternalID: "v1", SourceURL: "https://site.test/v1", Title: "Pool day"})
	if err != nil {
		t.Fatal(err)
	}
	sources, _ := library.Sources(ctx, schema.RecordKindVideo, stub.Id)
	if len(sources) != 1 || sources[0].FetchedAt != nil {
		t.Fatalf("stub sources = %+v", sources)
	}
	found, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil || found.Id != stub.Id {
		t.Fatalf("found %s, stub %s, err %v", found.Id, stub.Id, err)
	}
}

// TestFilters lists videos by performer, tag and studio, and names by text.
func TestFilters(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	video, _ := library.Video(ctx, ref.Id)
	text := "pool"
	minimum := int32(300)
	for name, filter := range map[string]schema.Filter{
		"performer": {PerformerIds: video.Performers[:1]},
		"tags":      {IncludeTagIds: video.Tags},
		"studio":    {StudioIds: []string{*video.Studio}},
		"text":      {Text: &text},
		"duration":  {MinDurationSeconds: &minimum},
		"site":      {Sources: []string{"site"}},
	} {
		videos, err := library.Videos(ctx, MediaQuery{Filter: filter})
		if err != nil || len(videos) != 1 {
			t.Errorf("%s filter: %d videos, err %v", name, len(videos), err)
		}
	}
	maximum := int32(300)
	for name, filter := range map[string]schema.Filter{
		"excluded tag": {ExcludeTagIds: video.Tags[:1]},
		"missing tag":  {IncludeTagIds: []string{video.Tags[0], "other"}},
		"duration":     {MaxDurationSeconds: &maximum},
		"other site":   {Sources: []string{"elsewhere"}},
	} {
		videos, err := library.Videos(ctx, MediaQuery{Filter: filter})
		if err != nil || len(videos) != 0 {
			t.Errorf("%s filter: %d videos, err %v", name, len(videos), err)
		}
	}
	performers, err := library.Performers(ctx, NameFilter{Text: "ada"})
	if err != nil || len(performers) != 1 || performers[0].Name != "Ada Lane" {
		t.Errorf("performers = %+v, err %v", performers, err)
	}
	count, err := library.CountMedia(ctx, schema.RecordKindVideo, schema.Filter{PerformerIds: video.Performers[:1]})
	if err != nil || count != 1 {
		t.Errorf("count = %d, err %v", count, err)
	}
}

// TestIngestGalleryImages stores a gallery page and its images, then a
// shorter image list replacing them.
func TestIngestGalleryImages(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	page := func(urls ...string) []*plugins.ScrapeResult {
		images := []plugins.Image{}
		for _, url := range urls {
			images = append(images, plugins.Image{Type: "gallery", FilePath: url})
		}
		return []*plugins.ScrapeResult{{Type: plugins.MediaTypeGallery, Gallery: &plugins.Gallery{ExternalID: "g1", Title: "Set", Images: images}}}
	}
	ref, err := library.IngestFound(ctx, "site", page("a.jpg", "b.jpg", "c.jpg"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := library.IngestFound(ctx, "site", page("x.jpg", "y.jpg")); err != nil {
		t.Fatal(err)
	}
	gallery, _ := library.Gallery(ctx, ref.Id)
	images, err := library.Images(ctx, ref.Id)
	if err != nil || len(images) != 2 || images[0].Asset.Uri != "x.jpg" || gallery.ImageCount != 2 {
		t.Fatalf("images = %+v, count %d, err %v", images, gallery.ImageCount, err)
	}
	if gallery.Cover == nil || gallery.Cover.Uri != "x.jpg" {
		t.Errorf("cover = %+v, want the latest page's first image", gallery.Cover)
	}
}

// TestPerformerPageFillsTheCreditedPerformer stores a performer page whose
// name matches a credit and expects one performer with the page's facts.
func TestPerformerPageFillsTheCreditedPerformer(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	results := foundScene()
	results = append(results, &plugins.ScrapeResult{Type: plugins.MediaTypePerformer, Performer: &plugins.Performer{
		ExternalID: "ada", SourceURL: "https://site.test/ada", Name: "Ada Lane", Country: "NL", HeightCm: 170, Twitter: "https://x.test/ada",
	}})
	if _, err := library.IngestFound(ctx, "site", results); err != nil {
		t.Fatal(err)
	}
	performers, _ := library.Performers(ctx, NameFilter{Text: "Ada"})
	if len(performers) != 1 {
		t.Fatalf("performers = %+v", performers)
	}
	ada := performers[0]
	if ada.Attributes.Country == nil || *ada.Attributes.Country != "NL" || *ada.Attributes.HeightCm != 170 || len(ada.Links) != 1 {
		t.Errorf("ada = %+v", ada)
	}
}
