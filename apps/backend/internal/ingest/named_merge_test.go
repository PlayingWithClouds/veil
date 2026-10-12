package ingest

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// insertLegacyPerformer inserts a performer the way old ingest did: no name_key.
func insertLegacyPerformer(t *testing.T, ctx context.Context, database *db.DB, fields map[string]any) string {
	t.Helper()
	id, err := database.Insert(ctx, "performer", fields)
	if err != nil {
		t.Fatalf("insert performer: %v", err)
	}
	return id.String()
}

// insertCreditedScene inserts a scene credited to the performers, with the
// performs_in rows ingest would write.
func insertCreditedScene(t *testing.T, ctx context.Context, database *db.DB, sourceURL string, performerIDs ...string) string {
	t.Helper()
	sceneID, err := database.Insert(ctx, "scene", map[string]any{"source_url": sourceURL, "title": sourceURL, "performers": performerIDs})
	if err != nil {
		t.Fatalf("insert scene: %v", err)
	}
	for _, performerID := range performerIDs {
		_, err := database.Insert(ctx, "performs_in", map[string]any{"performer": performerID, "media": sceneID.String()})
		if err != nil {
			t.Fatalf("insert credit: %v", err)
		}
	}
	return sceneID.String()
}

// TestRepairPerformersMergesSpellingDuplicates checks the cleanup of performers
// that old ingest split by spelling: credits and references move to the record
// with the most credits, its data is completed and the other site's page still
// resolves to it.
func TestRepairPerformersMergesSpellingDuplicates(t *testing.T) {
	service, database, ctx := openTestService(t)
	survivorID := insertLegacyPerformer(t, ctx, database, map[string]any{
		"name": "Isla Summer", "source_url": "https://eporner.test/pornstar/isla", "external_id": "eporner-ps-isla",
	})
	mergedID := insertLegacyPerformer(t, ctx, database, map[string]any{
		"name": "ISLA SUMMER", "source_url": "https://hqporner.test/actress/isla", "external_id": "hqporner-actress-isla",
		"country": "US", "favorite": true, "tags": []string{"tag:one"},
	})
	insertCreditedScene(t, ctx, database, "https://a.test/1", survivorID)
	insertCreditedScene(t, ctx, database, "https://a.test/2", survivorID)
	onlyMerged := insertCreditedScene(t, ctx, database, "https://a.test/3", mergedID)
	both := insertCreditedScene(t, ctx, database, "https://a.test/4", mergedID, survivorID)
	if _, err := database.Insert(ctx, "watchlist", map[string]any{"media": mergedID}); err != nil {
		t.Fatalf("insert watchlist: %v", err)
	}
	if _, err := database.Insert(ctx, "blocklist_entry", map[string]any{"kind": "performer", "target": mergedID}); err != nil {
		t.Fatalf("insert blocklist: %v", err)
	}

	merged, err := service.RepairNamedRecords(ctx)
	if err != nil {
		t.Fatalf("repair: %v", err)
	}
	if merged != 1 {
		t.Fatalf("merged = %d, want 1", merged)
	}

	if performers := countRows(t, ctx, database, `SELECT count(*) FROM performer`, nil); performers != 1 {
		t.Fatalf("performer rows = %d, want 1", performers)
	}
	survivor, err := service.GetByID(ctx, survivorID)
	if err != nil || survivor == nil {
		t.Fatalf("survivor missing: %v", err)
	}
	if survivor["country"] != "US" || !db.AsBool(survivor["favorite"]) {
		t.Fatalf("merged fields not carried over: country=%v favorite=%v", survivor["country"], survivor["favorite"])
	}
	assertSceneCredits(t, ctx, service, onlyMerged, survivorID)
	assertSceneCredits(t, ctx, service, both, survivorID)
	if credits := countRows(t, ctx, database, `SELECT count(*) FROM performs_in WHERE performer = $id`, db.Vars{"id": survivorID}); credits != 4 {
		t.Fatalf("survivor credits = %d, want 4", credits)
	}
	if moved := countRows(t, ctx, database, `SELECT count(*) FROM watchlist WHERE media = $id`, db.Vars{"id": survivorID}); moved != 1 {
		t.Fatalf("watchlist rows on survivor = %d, want 1", moved)
	}
	if moved := countRows(t, ctx, database, `SELECT count(*) FROM blocklist_entry WHERE target = $id`, db.Vars{"id": survivorID}); moved != 1 {
		t.Fatalf("blocklist rows on survivor = %d, want 1", moved)
	}

	page := ingestPerformerPage(t, service, "hqporner", &plugins.Performer{
		Name: "ISLA SUMMER", SourceURL: "https://hqporner.test/actress/isla", ExternalID: "hqporner-actress-isla",
	})
	if page.IsNew || page.CanonicalID != survivorID {
		t.Fatalf("merged site's page resolved to %s (new=%v), want %s", page.CanonicalID, page.IsNew, survivorID)
	}

	again, err := service.RepairNamedRecords(ctx)
	if err != nil || again != 0 {
		t.Fatalf("second repair merged %d (err %v), want 0", again, err)
	}
}

// assertSceneCredits fails unless the scene's performers are exactly want.
func assertSceneCredits(t *testing.T, ctx context.Context, service *Service, sceneID string, want ...string) {
	t.Helper()
	scene, err := service.GetByID(ctx, sceneID)
	if err != nil || scene == nil {
		t.Fatalf("scene %s missing: %v", sceneID, err)
	}
	got := jsonStringList(scene["performers"])
	if len(got) != len(want) {
		t.Fatalf("scene %s performers = %v, want %v", sceneID, got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("scene %s performers = %v, want %v", sceneID, got, want)
		}
	}
}

// insertLegacyStudio inserts a studio the way old ingest did: no name_key.
func insertLegacyStudio(t *testing.T, ctx context.Context, database *db.DB, fields map[string]any) string {
	t.Helper()
	id, err := database.Insert(ctx, "studio", fields)
	if err != nil {
		t.Fatalf("insert studio: %v", err)
	}
	return id.String()
}

// insertSceneOfStudio inserts a scene of the studio.
func insertSceneOfStudio(t *testing.T, ctx context.Context, database *db.DB, sourceURL, studioID string) string {
	t.Helper()
	id, err := database.Insert(ctx, "scene", map[string]any{"source_url": sourceURL, "title": sourceURL, "studio": studioID})
	if err != nil {
		t.Fatalf("insert scene: %v", err)
	}
	return id.String()
}

// TestRepairMergesSpellingDuplicateStudios checks the cleanup of studios that old
// ingest split by spelling: scenes and parent links move to the studio with the
// most scenes, its data is completed, and the other site's channel page still
// resolves to it.
func TestRepairMergesSpellingDuplicateStudios(t *testing.T) {
	service, database, ctx := openTestService(t)
	survivorID := insertLegacyStudio(t, ctx, database, map[string]any{
		"name": "Cum4K", "source_url": "https://xhamster.test/channels/cum4k", "external_id": "xhamster-channel-cum4k",
	})
	mergedID := insertLegacyStudio(t, ctx, database, map[string]any{
		"name": "CUM 4K", "source_url": "https://spankbang.test/channel/cum4k/", "external_id": "spankbang-channel-cum4k",
		"image_path": "https://cdn.test/cum4k.png",
	})
	childID := insertLegacyStudio(t, ctx, database, map[string]any{"name": "Cum4K Originals", "parent": mergedID})
	insertSceneOfStudio(t, ctx, database, "https://a.test/1", survivorID)
	insertSceneOfStudio(t, ctx, database, "https://a.test/2", survivorID)
	onlyMerged := insertSceneOfStudio(t, ctx, database, "https://a.test/3", mergedID)

	merged, err := service.RepairNamedRecords(ctx)
	if err != nil {
		t.Fatalf("repair: %v", err)
	}
	if merged != 1 {
		t.Fatalf("merged = %d, want 1", merged)
	}

	if studios := countRows(t, ctx, database, `SELECT count(*) FROM studio WHERE name_key = 'cum4k'`, nil); studios != 1 {
		t.Fatalf("cum4k studios = %d, want 1", studios)
	}
	scene, err := service.GetByID(ctx, onlyMerged)
	if err != nil || scene == nil || scene["studio"] != survivorID {
		t.Fatalf("scene studio = %v (err %v), want %s", scene["studio"], err, survivorID)
	}
	child, err := service.GetByID(ctx, childID)
	if err != nil || child == nil || child["parent"] != survivorID {
		t.Fatalf("child parent = %v (err %v), want %s", child["parent"], err, survivorID)
	}
	survivor, err := service.GetByID(ctx, survivorID)
	if err != nil || survivor == nil || survivor["image_path"] != "https://cdn.test/cum4k.png" {
		t.Fatalf("image not carried over: %+v (err %v)", survivor, err)
	}

	page := ingestStudioPage(t, service, "spankbang", &plugins.Studio{
		Name: "CUM 4K", SourceURL: "https://spankbang.test/channel/cum4k/", ExternalID: "spankbang-channel-cum4k",
	})
	if page.IsNew || page.CanonicalID != survivorID {
		t.Fatalf("merged site's page resolved to %s (new=%v), want %s", page.CanonicalID, page.IsNew, survivorID)
	}
}

// TestMergingStudioIntoItsParentLeavesNoSelfParent checks that merging a studio
// with its own parent doesn't leave the survivor its own parent.
func TestMergingStudioIntoItsParentLeavesNoSelfParent(t *testing.T) {
	service, database, ctx := openTestService(t)
	parentID := insertLegacyStudio(t, ctx, database, map[string]any{"name": "Network"})
	childID := insertLegacyStudio(t, ctx, database, map[string]any{"name": "NETWORK", "parent": parentID})
	insertSceneOfStudio(t, ctx, database, "https://a.test/1", parentID)

	if err := service.MergeRecords(ctx, "studio", parentID, childID); err != nil {
		t.Fatalf("merge: %v", err)
	}

	survivor, err := service.GetByID(ctx, parentID)
	if err != nil || survivor == nil {
		t.Fatalf("survivor missing: %v", err)
	}
	if survivor["parent"] != nil {
		t.Fatalf("survivor parent = %v, want none", survivor["parent"])
	}
}
