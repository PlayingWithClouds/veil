package ingest

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

const testSceneURL = "https://example.test/scene/1"

func openTestService(t *testing.T) (*Service, *db.DB, context.Context) {
	t.Helper()
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return New(database), database, ctx
}

func ingestScene(t *testing.T, ctx context.Context, service *Service, pluginName string, scene *plugins.Scene) *IngestResult {
	t.Helper()
	result, err := service.IngestScrapeResults(ctx, pluginName, scene.SourceURL, []*plugins.ScrapeResult{
		{Type: plugins.MediaTypeScene, Scene: scene},
	})
	if err != nil {
		t.Fatalf("ingest scene: %v", err)
	}
	return result
}

func countRows(t *testing.T, ctx context.Context, database *db.DB, query string, vars db.Vars) int {
	t.Helper()
	count, err := database.Int(ctx, query, vars)
	if err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return count
}

func TestSceneCreateThenMergeBySourceURL(t *testing.T) {
	service, database, ctx := openTestService(t)

	first := ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: testSceneURL, Title: "First Title",
		Downloads: []plugins.Download{{URL: "https://cdn.test/a.mp4"}},
	})
	if !first.IsNew || first.Mode != "new" {
		t.Fatalf("first ingest: %+v", first)
	}

	second := ingestScene(t, ctx, service, "beta", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: testSceneURL, Title: "Other Title",
		Details: "Filled later", Duration: 600,
		Downloads: []plugins.Download{{URL: "https://cdn.test/a.mp4"}, {URL: "https://cdn.test/b.mp4"}},
	})
	if second.IsNew || second.CanonicalID != first.CanonicalID || second.Mode != "auto" {
		t.Fatalf("second ingest: %+v (first %s)", second, first.CanonicalID)
	}

	if scenes := countRows(t, ctx, database, `SELECT count(*) FROM scene`, nil); scenes != 1 {
		t.Fatalf("scene rows = %d, want 1", scenes)
	}
	scene, err := service.GetByID(ctx, first.CanonicalID)
	if err != nil || scene == nil {
		t.Fatalf("get scene: %v", err)
	}
	if scene["title"] != "First Title" {
		t.Fatalf("title overwritten: %v", scene["title"])
	}
	if scene["details"] != "Filled later" || db.AsInt(scene["duration_seconds"]) != 600 {
		t.Fatalf("empty fields not filled: %+v", scene)
	}

	secondChanges, err := database.Strings(ctx,
		`SELECT key FROM changes WHERE observation = $observation ORDER BY key`,
		db.Vars{"observation": second.ObservationID})
	if err != nil {
		t.Fatalf("list changes: %v", err)
	}
	if !slices.Equal(secondChanges, []string{"details", "duration_seconds"}) {
		t.Fatalf("second observation changes = %v", secondChanges)
	}
	change, err := database.QueryRow(ctx,
		`SELECT * FROM changes WHERE observation = $observation AND key = 'details'`,
		db.Vars{"observation": second.ObservationID})
	if err != nil || change["value"] != "Filled later" || change["plugin"] != "beta" || change["canonical"] != first.CanonicalID {
		t.Fatalf("details change row: %+v (err %v)", change, err)
	}
	if created := countRows(t, ctx, database, `SELECT count(*) FROM changes WHERE observation = $observation AND key = 'title'`,
		db.Vars{"observation": first.ObservationID}); created != 1 {
		t.Fatalf("create observation title change rows = %d, want 1", created)
	}

	if streams := countRows(t, ctx, database, `SELECT count(*) FROM stream WHERE media = $media`,
		db.Vars{"media": first.CanonicalID}); streams != 2 {
		t.Fatalf("stream rows = %d, want 2", streams)
	}
}

func TestObservationDataIsFlattenedOnRead(t *testing.T) {
	service, _, ctx := openTestService(t)
	result := ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: testSceneURL, Title: "T",
		Downloads: []plugins.Download{{URL: "https://cdn.test/a.mp4"}},
	})

	withObservation, err := service.GetByIDWithObs(ctx, result.CanonicalID)
	if err != nil {
		t.Fatalf("get with obs: %v", err)
	}
	downloads, isList := withObservation["downloads"].([]any)
	if !isList || len(downloads) != 1 {
		t.Fatalf("downloads overlay: %+v", withObservation["downloads"])
	}

	pluginName, sourceURL := service.PrimarySource(ctx, withObservation)
	if pluginName != "alpha" || sourceURL != testSceneURL {
		t.Fatalf("primary source = %q %q", pluginName, sourceURL)
	}

	merged, err := service.ManualMerge(ctx, result.CanonicalID, result.ObservationID, "tester")
	if err != nil || merged.Mode != "manual" {
		t.Fatalf("manual merge: %+v %v", merged, err)
	}
}

func TestPerformerCreditsWrittenOnce(t *testing.T) {
	service, database, ctx := openTestService(t)
	scene := &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: testSceneURL, Title: "T",
		Performers: []plugins.ScenePerformer{{Name: "Jane", As: "J", Order: 1}, {Name: "Kim"}},
		Tags:       []string{"outdoor", "outdoor"},
		Studio:     &plugins.StudioRef{Name: "Studio A"},
	}
	result := ingestScene(t, ctx, service, "alpha", scene)
	ingestScene(t, ctx, service, "alpha", scene)

	if credits := countRows(t, ctx, database, `SELECT count(*) FROM performs_in WHERE media = $media`,
		db.Vars{"media": result.CanonicalID}); credits != 2 {
		t.Fatalf("performs_in rows = %d, want 2", credits)
	}
	if performers := countRows(t, ctx, database, `SELECT count(*) FROM performer`, nil); performers != 2 {
		t.Fatalf("performer rows = %d, want 2", performers)
	}
	credit, err := database.QueryRow(ctx,
		`SELECT performs_in.* FROM performs_in JOIN performer ON performer.id = performs_in.performer WHERE performer.name = 'Jane'`, nil)
	if err != nil || credit["credited_as"] != "J" || db.AsInt(credit["billing_order"]) != 1 {
		t.Fatalf("Jane credit: %+v (err %v)", credit, err)
	}

	sceneRow, err := service.GetByID(ctx, result.CanonicalID)
	if err != nil {
		t.Fatalf("get scene: %v", err)
	}
	performerIDs, _ := sceneRow["performers"].([]any)
	tagIDs, _ := sceneRow["tags"].([]any)
	if len(performerIDs) != 2 || len(tagIDs) != 1 {
		t.Fatalf("denormalized links: performers=%v tags=%v", performerIDs, tagIDs)
	}
	studioID, _ := sceneRow["studio"].(string)
	if TableFromID(studioID) != "studio" {
		t.Fatalf("studio link: %v", sceneRow["studio"])
	}
}

func TestTagAliasResolvesToCanonicalTag(t *testing.T) {
	service, database, ctx := openTestService(t)
	tagID, err := database.Insert(ctx, "tag", map[string]any{"name": "Outdoors", "aliases": []string{"outside"}})
	if err != nil {
		t.Fatalf("insert tag: %v", err)
	}
	resolved := service.resolveNamedRecords(ctx, "tag", []string{"outside", "Outdoors"})
	if len(resolved) != 1 || resolved[0] != tagID {
		t.Fatalf("resolved = %v, want %s once", resolved, tagID)
	}
	if tags := countRows(t, ctx, database, `SELECT count(*) FROM tag`, nil); tags != 1 {
		t.Fatalf("tag rows = %d, want 1", tags)
	}
}

func TestTagsMatchCaseInsensitivelyAndSkipNoise(t *testing.T) {
	service, database, ctx := openTestService(t)

	result := ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: testSceneURL, Title: "Scene",
		Tags: []string{"anal", "Anal", "Uncategorized", "Sex", "big ass"},
	})
	ingestScene(t, ctx, service, "beta", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: "https://example.test/scene/2", Title: "Other",
		Tags: []string{"ANAL", "Big Ass"},
	})

	if tags := countRows(t, ctx, database, `SELECT count(*) FROM tag`, nil); tags != 2 {
		t.Fatalf("tag rows = %d, want 2 (anal, big ass)", tags)
	}
	names, err := database.Strings(ctx,
		`SELECT tag.name FROM scene, json_each(scene.tags) AS entry JOIN tag ON tag.id = entry.value
		 WHERE scene.id = $id ORDER BY entry.key`, db.Vars{"id": result.CanonicalID})
	if err != nil {
		t.Fatalf("scene tags: %v", err)
	}
	if !slices.Equal(names, []string{"anal", "big ass"}) {
		t.Fatalf("scene tags = %v, want [anal big ass]", names)
	}
}

func TestDiscoveredItemStudioLinksAndBackfillsLogo(t *testing.T) {
	service, database, ctx := openTestService(t)

	channel := &plugins.StudioRef{Name: "Xander-Vision", SourceURL: "https://example.test/channels/xander"}
	sceneID, err := service.IngestDiscoveredItem(ctx, "alpha", plugins.DiscoveredItem{
		Title: "Stub", MediaType: plugins.MediaTypeScene, SourceURL: testSceneURL, ExternalID: "alpha-1",
		Studio: channel,
	})
	if err != nil {
		t.Fatalf("ingest stub: %v", err)
	}
	if linked := countRows(t, ctx, database,
		`SELECT count(*) FROM scene JOIN studio ON studio.id = scene.studio WHERE scene.id = $id AND studio.name = 'Xander-Vision'`,
		db.Vars{"id": sceneID}); linked != 1 {
		t.Fatalf("stub not linked to its channel")
	}

	// A later listing carrying the logo fills it in on the same studio.
	withLogo := *channel
	withLogo.ImagePath = "https://example.test/logo.jpg"
	if _, err := service.IngestDiscoveredItem(ctx, "alpha", plugins.DiscoveredItem{
		Title: "Stub 2", MediaType: plugins.MediaTypeScene, SourceURL: "https://example.test/scene/2", ExternalID: "alpha-2",
		Studio: &withLogo,
	}); err != nil {
		t.Fatalf("ingest second stub: %v", err)
	}
	if studios := countRows(t, ctx, database,
		`SELECT count(*) FROM studio WHERE image_path = 'https://example.test/logo.jpg'`, nil); studios != 1 {
		t.Fatalf("logo not backfilled onto the one studio")
	}
	if studios := countRows(t, ctx, database, `SELECT count(*) FROM studio`, nil); studios != 1 {
		t.Fatalf("studio rows = %d, want 1", studios)
	}
}

func TestPluginMarkersAreGlobalAndDeduplicated(t *testing.T) {
	service, database, ctx := openTestService(t)
	scene := &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: testSceneURL, Title: "T",
		Markers: []plugins.SceneMarkerInput{{Seconds: 12, Tag: "kiss"}, {Seconds: 30, Label: "intro"}},
	}
	result := ingestScene(t, ctx, service, "alpha", scene)
	ingestScene(t, ctx, service, "alpha", scene)

	rows, err := database.Query(ctx, `SELECT * FROM scene_marker WHERE media = $media ORDER BY seconds`,
		db.Vars{"media": result.CanonicalID})
	if err != nil {
		t.Fatalf("list markers: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("marker rows = %d, want 2", len(rows))
	}
	for _, row := range rows {
		if row["personal"] != false || row["created_by"] != "plugin:alpha" {
			t.Fatalf("marker not global/plugin-owned: %+v", row)
		}
	}
}

func TestCollectionMembersAndGalleryImages(t *testing.T) {
	service, database, ctx := openTestService(t)
	sceneResult := ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: testSceneURL, Title: "T",
	})

	collection := &plugins.Collection{
		Type: plugins.MediaTypeCollection, SourceURL: "https://example.test/c/1", Name: "C",
		MemberURLs: []string{"https://example.test/missing", testSceneURL},
	}
	collectionResults := []*plugins.ScrapeResult{{Type: plugins.MediaTypeCollection, Collection: collection}}
	result, err := service.IngestScrapeResults(ctx, "alpha", collection.SourceURL, collectionResults)
	if err != nil {
		t.Fatalf("ingest collection: %v", err)
	}
	if !slices.Equal(result.UnresolvedMemberURLs, []string{"https://example.test/missing"}) {
		t.Fatalf("unresolved = %v", result.UnresolvedMemberURLs)
	}
	collection.MemberURLs = []string{testSceneURL}
	if _, err := service.IngestScrapeResults(ctx, "alpha", collection.SourceURL, collectionResults); err != nil {
		t.Fatalf("re-ingest collection: %v", err)
	}
	item, err := database.QueryRow(ctx, `SELECT * FROM collection_item WHERE media = $media`,
		db.Vars{"media": sceneResult.CanonicalID})
	if err != nil || item == nil || db.AsInt(item["position"]) != 0 {
		t.Fatalf("collection item: %+v (err %v)", item, err)
	}
	if items := countRows(t, ctx, database, `SELECT count(*) FROM collection_item`, nil); items != 1 {
		t.Fatalf("collection_item rows = %d, want 1", items)
	}

	gallery := &plugins.Gallery{
		Type: plugins.MediaTypeGallery, SourceURL: "https://example.test/g/1", Title: "G",
		Images: []plugins.Image{{FilePath: "/g/1.jpg"}, {FilePath: "/g/2.jpg"}},
	}
	galleryResults := []*plugins.ScrapeResult{{Type: plugins.MediaTypeGallery, Gallery: gallery}}
	galleryResult, err := service.IngestScrapeResults(ctx, "alpha", gallery.SourceURL, galleryResults)
	if err != nil {
		t.Fatalf("ingest gallery: %v", err)
	}
	// Tag one image so the prune keeps it even after it disappears from the scrape.
	if _, err := database.Exec(ctx, `UPDATE image SET tags = '["tag:x"]' WHERE file_path = '/g/1.jpg'`, nil); err != nil {
		t.Fatalf("tag image: %v", err)
	}
	gallery.Images = []plugins.Image{{FilePath: "/g/3.jpg"}}
	if _, err := service.IngestScrapeResults(ctx, "alpha", gallery.SourceURL, galleryResults); err != nil {
		t.Fatalf("re-ingest gallery: %v", err)
	}
	paths, err := database.Strings(ctx, `SELECT file_path FROM image WHERE media = $media ORDER BY file_path`,
		db.Vars{"media": galleryResult.CanonicalID})
	if err != nil || !slices.Equal(paths, []string{"/g/1.jpg", "/g/3.jpg"}) {
		t.Fatalf("gallery images = %v (err %v)", paths, err)
	}
	galleryRow, err := service.GetByID(ctx, galleryResult.CanonicalID)
	if err != nil || db.AsInt(galleryRow["image_count"]) != 1 {
		t.Fatalf("image_count: %+v (err %v)", galleryRow, err)
	}
}

func TestListAndCountSearchIsLiteral(t *testing.T) {
	service, _, ctx := openTestService(t)
	for index, title := range []string{"100% Real", "Plain", "Another 100%"} {
		ingestScene(t, ctx, service, "alpha", &plugins.Scene{
			Type: plugins.MediaTypeScene, SourceURL: testSceneURL + string(rune('a'+index)), Title: title,
		})
	}
	rows, err := service.ListByTable(ctx, "scene", "100%", 10, 0)
	if err != nil || len(rows) != 2 {
		t.Fatalf("list = %d rows (err %v), want 2", len(rows), err)
	}
	count, err := service.CountByTable(ctx, "scene", "real")
	if err != nil || count != 1 {
		t.Fatalf("count = %d (err %v), want 1", count, err)
	}
	total, err := service.CountByTable(ctx, "scene", "")
	if err != nil || total != 3 {
		t.Fatalf("total = %d (err %v), want 3", total, err)
	}
}
