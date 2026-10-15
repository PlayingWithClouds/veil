package face

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// testService builds a Service on a real database with two fake recognition
// plugins, whose answers are given per plugin name.
func testService(t *testing.T, replies map[string]*plugins.FaceIdentifyResult, failing map[string]bool) (*Service, *db.DB, *ingest.Service, *[]string) {
	t.Helper()
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	library := ingest.New(database)
	loaded := []string{}
	var services []*plugins.Plugin
	for name := range replies {
		services = append(services, &plugins.Plugin{Meta: plugins.PluginMeta{Name: name}})
	}
	service := &Service{
		database: database,
		library:  library,
		services: func() []*plugins.Plugin { return services },
		identify: func(ctx context.Context, plugin *plugins.Plugin, args plugins.FaceIdentifyArgs) (*plugins.FaceIdentifyResult, error) {
			if failing[plugin.Meta.Name] {
				return nil, errors.New("site down")
			}
			return replies[plugin.Meta.Name], nil
		},
		loadPicture: func(ctx context.Context, pictureURL string) ([]byte, error) {
			loaded = append(loaded, pictureURL)
			return []byte("\xff\xd8\xff\xe0 jpeg bytes"), nil
		},
	}
	return service, database, library, &loaded
}

func ingestSceneWithPoster(t *testing.T, library *ingest.Service, poster string) string {
	t.Helper()
	result, err := library.IngestScrapeResults(context.Background(), "alpha", "https://example.test/scene/1", []*plugins.ScrapeResult{
		{Type: plugins.MediaTypeScene, Scene: &plugins.Scene{
			Type: plugins.MediaTypeScene, SourceURL: "https://example.test/scene/1", Title: "A scene", PosterPath: poster,
			Downloads: []plugins.Download{{URL: "https://cdn.test/a.mp4"}},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return result.CanonicalID
}

// TestIdentifyMergesRanksAndMarksLibrary checks the merge by name_key (best
// similarity kept, links united), the ranking, a failing service reported
// without failing the run, and the library marker.
func TestIdentifyMergesRanksAndMarksLibrary(t *testing.T) {
	replies := map[string]*plugins.FaceIdentifyResult{
		"first": {Candidates: []plugins.FaceCandidate{
			{Name: "Lexi Lore", Similarity: 0.45, ProfileURLs: []string{"https://a.test/lexi"}},
			{Name: "Goldie Baby", Similarity: 0.47},
		}},
		"second": {Candidates: []plugins.FaceCandidate{
			{Name: "lexi  lore", Similarity: 0.6, ImageURL: "https://b.test/lexi.jpg", ProfileURLs: []string{"https://b.test/lexi", "https://a.test/lexi"}},
		}},
		"broken": {},
	}
	service, _, library, loaded := testService(t, replies, map[string]bool{"broken": true})
	ctx := context.Background()
	sceneID := ingestSceneWithPoster(t, library, "https://cdn.test/poster.jpg")
	if _, err := library.CreditPerformer(ctx, sceneID, "Goldie Baby"); err != nil {
		t.Fatal(err)
	}

	result, err := service.Identify(ctx, sceneID, "")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(*loaded, []string{"https://cdn.test/poster.jpg"}) {
		t.Errorf("picture loaded = %v", *loaded)
	}
	if len(result.Candidates) != 2 {
		t.Fatalf("candidates = %+v", result.Candidates)
	}
	top := result.Candidates[0]
	if top.Name != "lexi  lore" || top.Similarity != 0.6 || top.ImageURL != "https://b.test/lexi.jpg" {
		t.Errorf("top candidate = %+v", top)
	}
	if !slices.Equal(top.ProfileURLs, []string{"https://a.test/lexi", "https://b.test/lexi"}) && !slices.Equal(top.ProfileURLs, []string{"https://b.test/lexi", "https://a.test/lexi"}) {
		t.Errorf("profile links = %v", top.ProfileURLs)
	}
	if top.PerformerID != "" {
		t.Errorf("Lexi Lore is not in the library yet: %q", top.PerformerID)
	}
	if result.Candidates[1].Name != "Goldie Baby" || result.Candidates[1].PerformerID == "" {
		t.Errorf("Goldie Baby should be marked as in the library: %+v", result.Candidates[1])
	}
	if len(result.Sources) != 2 || len(result.Failures) != 1 {
		t.Errorf("sources %v failures %v", result.Sources, result.Failures)
	}
}

// TestPictureSelection picks the stored image per media kind and credits an
// image's gallery.
func TestPictureSelection(t *testing.T) {
	service, database, library, _ := testService(t, map[string]*plugins.FaceIdentifyResult{"only": {}}, nil)
	ctx := context.Background()

	sceneID := ingestSceneWithPoster(t, library, "https://cdn.test/poster.jpg")
	if _, err := database.Exec(ctx, `UPDATE scene SET thumbnail_path = $path WHERE id = $id`, db.Vars{"id": sceneID, "path": "https://cdn.test/best.jpg"}); err != nil {
		t.Fatal(err)
	}
	if got, err := service.pictureURL(ctx, sceneID, ""); err != nil || got != "https://cdn.test/best.jpg" {
		t.Errorf("scene picture = %q, %v", got, err)
	}

	galleryID, err := database.Insert(ctx, "gallery", map[string]any{"title": "G", "source_url": "https://example.test/g"})
	if err != nil {
		t.Fatal(err)
	}
	var imageIDs []string
	for position, file := range []string{"https://cdn.test/1.jpg", "https://cdn.test/2.jpg"} {
		imageID, err := database.Insert(ctx, "image", map[string]any{"media": galleryID.String(), "file_path": file, "content": true, "position": position})
		if err != nil {
			t.Fatal(err)
		}
		imageIDs = append(imageIDs, imageID.String())
	}
	if got, _ := service.pictureURL(ctx, galleryID.String(), ""); got != "https://cdn.test/1.jpg" {
		t.Errorf("gallery default picture = %q", got)
	}
	if got, _ := service.pictureURL(ctx, galleryID.String(), imageIDs[1]); got != "https://cdn.test/2.jpg" {
		t.Errorf("gallery chosen picture = %q", got)
	}
	if got, _ := service.pictureURL(ctx, imageIDs[1], ""); got != "https://cdn.test/2.jpg" {
		t.Errorf("image picture = %q", got)
	}

	performerID, err := service.CreditPerformer(ctx, imageIDs[0], "Lexi Lore")
	if err != nil {
		t.Fatal(err)
	}
	row, err := database.Get(ctx, galleryID)
	if err != nil {
		t.Fatal(err)
	}
	if credited, _ := row["performers"].([]any); len(credited) != 1 || credited[0] != performerID {
		t.Errorf("gallery performers = %v, want [%s]", row["performers"], performerID)
	}
}

// TestIdentifyWithoutPlugins explains itself when no service is installed.
func TestIdentifyWithoutPlugins(t *testing.T) {
	service, _, _, _ := testService(t, map[string]*plugins.FaceIdentifyResult{}, nil)
	if _, err := service.Identify(context.Background(), "scene:x", ""); err == nil {
		t.Error("expected an error without face plugins")
	}
}
