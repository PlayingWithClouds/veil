package ingest

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
)

// insertLegacyTag inserts a tag the way old ingest did: no name_key.
func insertLegacyTag(t *testing.T, ctx context.Context, database *db.DB, name string) string {
	t.Helper()
	id, err := database.Insert(ctx, "tag", map[string]any{"name": name})
	if err != nil {
		t.Fatalf("insert tag: %v", err)
	}
	return id.String()
}

// TestTagSpellingsShareOneTag checks that tags differing only in spacing,
// punctuation or case resolve to one record, and that an alias still resolves.
func TestTagSpellingsShareOneTag(t *testing.T) {
	service, database, ctx := openTestService(t)

	first := service.resolveNamedRecords(ctx, "tag", []string{"Blow Job"})
	for _, spelling := range []string{"blowjob", "blow-job", "BLOWJOB"} {
		again := service.resolveNamedRecords(ctx, "tag", []string{spelling})
		if len(again) != 1 || again[0] != first[0] {
			t.Fatalf("%q resolved to %v, want %v", spelling, again, first)
		}
	}
	if tags := countRows(t, ctx, database, `SELECT count(*) FROM tag`, nil); tags != 1 {
		t.Fatalf("tag rows = %d, want 1", tags)
	}

	if _, err := database.Exec(ctx, `UPDATE tag SET aliases = '["bj"]'`, nil); err != nil {
		t.Fatalf("set alias: %v", err)
	}
	aliased := service.resolveNamedRecords(ctx, "tag", []string{"BJ"})
	if len(aliased) != 1 || aliased[0] != first[0] {
		t.Fatalf("alias resolved to %v, want %v", aliased, first)
	}
}

// TestRepairMergesSpellingDuplicateTags checks the cleanup of tags old ingest
// split by spacing and punctuation: every list that carried the merged tag now
// carries the tag with the most scenes, once.
func TestRepairMergesSpellingDuplicateTags(t *testing.T) {
	service, database, ctx := openTestService(t)
	survivorID := insertLegacyTag(t, ctx, database, "blowjob")
	mergedID := insertLegacyTag(t, ctx, database, "blow job")
	otherID := insertLegacyTag(t, ctx, database, "anal")
	for _, sourceURL := range []string{"https://a.test/1", "https://a.test/2"} {
		if _, err := database.Insert(ctx, "scene", map[string]any{"source_url": sourceURL, "title": sourceURL, "tags": []string{survivorID}}); err != nil {
			t.Fatalf("insert scene: %v", err)
		}
	}
	sceneID, err := database.Insert(ctx, "scene", map[string]any{"source_url": "https://a.test/3", "title": "3", "tags": []string{otherID, mergedID, survivorID}})
	if err != nil {
		t.Fatalf("insert scene: %v", err)
	}
	performerID := insertLegacyPerformer(t, ctx, database, map[string]any{"name": "Someone Else", "tags": []string{mergedID}})
	if _, err := database.Insert(ctx, "blocklist_entry", map[string]any{"kind": "tag", "target": mergedID}); err != nil {
		t.Fatalf("insert blocklist: %v", err)
	}

	merged, err := service.RepairNamedRecords(ctx)
	if err != nil {
		t.Fatalf("repair: %v", err)
	}
	if merged != 1 {
		t.Fatalf("merged = %d, want 1", merged)
	}

	if tags := countRows(t, ctx, database, `SELECT count(*) FROM tag WHERE name_key = 'blowjob'`, nil); tags != 1 {
		t.Fatalf("blowjob tags = %d, want 1", tags)
	}
	scene, err := service.GetByID(ctx, sceneID.String())
	if err != nil || scene == nil {
		t.Fatalf("scene missing: %v", err)
	}
	if got := jsonStringList(scene["tags"]); len(got) != 2 || got[0] != otherID || got[1] != survivorID {
		t.Fatalf("scene tags = %v, want [%s %s]", got, otherID, survivorID)
	}
	performer, err := service.GetByID(ctx, performerID)
	if err != nil || performer == nil {
		t.Fatalf("performer missing: %v", err)
	}
	if got := jsonStringList(performer["tags"]); len(got) != 1 || got[0] != survivorID {
		t.Fatalf("performer tags = %v, want [%s]", got, survivorID)
	}
	if blocked := countRows(t, ctx, database, `SELECT count(*) FROM blocklist_entry WHERE target = $id`, db.Vars{"id": survivorID}); blocked != 1 {
		t.Fatalf("blocklist rows on survivor = %d, want 1", blocked)
	}

	again, err := service.RepairNamedRecords(ctx)
	if err != nil || again != 0 {
		t.Fatalf("second repair merged %d (err %v), want 0", again, err)
	}
}
