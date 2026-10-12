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

	merged, err := service.RepairPerformers(ctx)
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

	again, err := service.RepairPerformers(ctx)
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
