package ingest

import (
	"slices"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

func relatedItem(slug string) plugins.DiscoveredItem {
	return plugins.DiscoveredItem{
		Title:           "Related " + slug,
		MediaType:       plugins.MediaTypeScene,
		SourceURL:       "https://example.test/scene/" + slug,
		ExternalID:      "test-" + slug,
		PosterPath:      "https://cdn.example.test/" + slug + ".jpg",
		DurationSeconds: 600,
	}
}

// relatedEdges returns sceneID's edges as "source:related-title" in rank order.
func relatedEdges(t *testing.T, service *Service, database *db.DB, sceneID string) []string {
	t.Helper()
	rows, err := database.Query(t.Context(),
		`SELECT scene_related.source, scene.title FROM scene_related
		 JOIN scene ON scene.id = scene_related.related
		 WHERE scene_related.scene = $scene ORDER BY scene_related.source, scene_related.rank`,
		db.Vars{"scene": sceneID})
	if err != nil {
		t.Fatalf("query edges: %v", err)
	}
	edges := []string{}
	for _, row := range rows {
		edges = append(edges, row["source"].(string)+":"+row["title"].(string))
	}
	return edges
}

func TestFindResultRelatedBecomesStubsAndEdges(t *testing.T) {
	service, database, ctx := openTestService(t)
	result := ingestScene(t, ctx, service, "testplugin", &plugins.Scene{
		Type: plugins.MediaTypeScene, ExternalID: "test-1", SourceURL: testSceneURL, Title: "Visited",
		Related: []plugins.DiscoveredItem{
			relatedItem("a"),
			// The scene itself in its own related list is skipped.
			{Title: "Visited", MediaType: plugins.MediaTypeScene, SourceURL: testSceneURL, ExternalID: "test-1"},
			relatedItem("b"),
			relatedItem("a"), // duplicate
		},
	})

	edges := relatedEdges(t, service, database, result.CanonicalID)
	if !slices.Equal(edges, []string{"site:Related a", "site:Related b"}) {
		t.Errorf("edges = %v", edges)
	}
	stub, _ := database.QueryRow(ctx, `SELECT * FROM scene WHERE source_url = $url`,
		db.Vars{"url": "https://example.test/scene/a"})
	if stub == nil || db.AsInt(stub["duration_seconds"]) != 600 || stub["poster_path"] == nil {
		t.Errorf("related stub = %v", stub)
	}
	if countRows(t, ctx, database, `SELECT count(*) FROM observation WHERE json_extract(data, '$.related') IS NOT NULL`, nil) != 0 {
		t.Error("related list leaked into observation data")
	}
	if count, _ := service.RelatedCount(ctx, result.CanonicalID); count != 2 {
		t.Errorf("related count = %d, want 2", count)
	}
}

func TestLinkRelatedContinuesRankPerSource(t *testing.T) {
	service, database, ctx := openTestService(t)
	sceneID := ingestScene(t, ctx, service, "testplugin", &plugins.Scene{
		Type: plugins.MediaTypeScene, ExternalID: "test-1", SourceURL: testSceneURL, Title: "Visited",
	}).CanonicalID

	for _, batch := range [][]plugins.DiscoveredItem{{relatedItem("a")}, {relatedItem("b"), relatedItem("a")}} {
		if _, err := service.LinkRelated(ctx, sceneID, "testplugin", RelatedSourcePerformer, batch); err != nil {
			t.Fatalf("link: %v", err)
		}
	}
	linked, err := service.LinkRelated(ctx, sceneID, "testplugin", RelatedSourceTag, []plugins.DiscoveredItem{relatedItem("c")})
	if err != nil || linked != 1 {
		t.Fatalf("link tag = %d, %v", linked, err)
	}
	ranks, _ := database.Query(ctx, `SELECT source, rank FROM scene_related WHERE scene = $scene ORDER BY source, rank`,
		db.Vars{"scene": sceneID})
	got := []string{}
	for _, row := range ranks {
		got = append(got, row["source"].(string)+"#"+string(rune('0'+db.AsInt(row["rank"]))))
	}
	if !slices.Equal(got, []string{"performer#0", "performer#1", "tag#0"}) {
		t.Errorf("ranks = %v", got)
	}
}

func TestSceneSearchTermsKeepCreditOrder(t *testing.T) {
	service, _, ctx := openTestService(t)
	sceneID := ingestScene(t, ctx, service, "testplugin", &plugins.Scene{
		Type: plugins.MediaTypeScene, ExternalID: "test-1", SourceURL: testSceneURL, Title: "Visited",
		Performers: []plugins.ScenePerformer{{Name: "Zoe"}, {Name: "Anna"}},
		Tags:       []string{"outdoor", "amateur"},
	}).CanonicalID

	performers, tags, title, err := service.SceneSearchTerms(ctx, sceneID)
	if err != nil {
		t.Fatalf("terms: %v", err)
	}
	if !slices.Equal(performers, []string{"Zoe", "Anna"}) || !slices.Equal(tags, []string{"outdoor", "amateur"}) || title != "Visited" {
		t.Errorf("terms = %v %v %q", performers, tags, title)
	}
}

func TestDetailFetchedFlag(t *testing.T) {
	service, _, ctx := openTestService(t)
	sceneID := ingestScene(t, ctx, service, "testplugin", &plugins.Scene{
		Type: plugins.MediaTypeScene, ExternalID: "test-1", SourceURL: testSceneURL, Title: "Visited",
	}).CanonicalID
	if fetched, _ := service.DetailFetched(ctx, sceneID); fetched {
		t.Fatal("stub reported as fetched")
	}
	if err := service.MarkDetailFetched(ctx, sceneID); err != nil {
		t.Fatalf("mark: %v", err)
	}
	if fetched, _ := service.DetailFetched(ctx, sceneID); !fetched {
		t.Error("fetched flag not set")
	}
}

func TestSameExternalIDUnderAnotherURLMergesIntoStub(t *testing.T) {
	service, database, ctx := openTestService(t)
	stub, err := service.IngestDiscoveredItem(ctx, "testplugin", plugins.DiscoveredItem{
		Title: "Video", MediaType: plugins.MediaTypeScene, ExternalID: "test-42",
		SourceURL: "https://example.test/video-42/slug/",
	})
	if err != nil {
		t.Fatalf("stub: %v", err)
	}
	detail := ingestScene(t, ctx, service, "testplugin", &plugins.Scene{
		Type: plugins.MediaTypeScene, ExternalID: "test-42", SourceURL: "https://example.test/hd/42/Slug/",
		Title: "Video", Tags: []string{"outdoor"},
	})
	if detail.CanonicalID != stub {
		t.Errorf("detail ingested as %s, want the stub %s", detail.CanonicalID, stub)
	}
	if count := countRows(t, ctx, database, `SELECT count(*) FROM scene`, nil); count != 1 {
		t.Errorf("scene rows = %d, want 1", count)
	}
}
