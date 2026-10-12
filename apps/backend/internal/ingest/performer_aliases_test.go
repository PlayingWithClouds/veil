package ingest

import (
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// TestProfileAliasMergesDifferentlySpelledPerformer checks that a site profile
// listing a stage name the library knows as a separate performer merges the two,
// keeping the better-used record and its credits.
func TestProfileAliasMergesDifferentlySpelledPerformer(t *testing.T) {
	service, database, ctx := openTestService(t)
	credited := insertLegacyPerformer(t, ctx, database, map[string]any{"name": "Bella Danger", "name_key": "bella danger"})
	sceneID := insertCreditedScene(t, ctx, database, "https://a.test/1", credited)

	result := ingestPerformerPage(t, service, "freeones", &plugins.Performer{
		Name: "Abella Danger", SourceURL: "https://freeones.test/abella-danger", ExternalID: "freeones-abella-danger",
		Aliases: []string{"Bella Danger", "Angel"}, Country: "US",
	})

	if result.IsNew || result.CanonicalID != credited {
		t.Fatalf("profile resolved to %s (new=%v), want the credited %s", result.CanonicalID, result.IsNew, credited)
	}
	if performers := countRows(t, ctx, database, `SELECT count(*) FROM performer`, nil); performers != 1 {
		t.Fatalf("performers = %d, want 1", performers)
	}
	assertSceneCredits(t, ctx, service, sceneID, credited)
	survivor, err := service.GetByID(ctx, credited)
	if err != nil || survivor == nil {
		t.Fatalf("survivor missing: %v", err)
	}
	if survivor["country"] != "US" {
		t.Fatalf("profile data not carried over: %v", survivor["country"])
	}
	aliases := jsonStringList(survivor["aliases"])
	if len(aliases) != 2 || aliases[0] != "Angel" || aliases[1] != "Abella Danger" {
		t.Fatalf("aliases = %v, want Angel and Abella Danger", aliases)
	}

	again := ingestPerformerPage(t, service, "freeones", &plugins.Performer{
		Name: "Abella Danger", SourceURL: "https://freeones.test/abella-danger", ExternalID: "freeones-abella-danger",
	})
	if again.CanonicalID != credited {
		t.Fatalf("second visit resolved to %s, want %s", again.CanonicalID, credited)
	}
}

// TestSingleWordAliasNeverMerges checks that a one-word alias, which many
// people share, does not merge a performer of that name.
func TestSingleWordAliasNeverMerges(t *testing.T) {
	service, database, ctx := openTestService(t)
	insertLegacyPerformer(t, ctx, database, map[string]any{"name": "Angel", "name_key": "angel"})

	ingestPerformerPage(t, service, "freeones", &plugins.Performer{
		Name: "Angel Smith", SourceURL: "https://freeones.test/angel-smith", Aliases: []string{"Angel"},
	})

	if performers := countRows(t, ctx, database, `SELECT count(*) FROM performer`, nil); performers != 2 {
		t.Fatalf("performers = %d, want 2", performers)
	}
}

// TestRepairMergesAliasTwins checks the startup repair for performers that
// gained an alias naming another record.
func TestRepairMergesAliasTwins(t *testing.T) {
	service, database, ctx := openTestService(t)
	first := insertLegacyPerformer(t, ctx, database, map[string]any{"name": "Kaci Starr", "aliases": []string{"Kaci Star"}})
	second := insertLegacyPerformer(t, ctx, database, map[string]any{"name": "Kaci Star"})
	insertCreditedScene(t, ctx, database, "https://a.test/1", second)

	merged, err := service.RepairNamedRecords(ctx)
	if err != nil || merged != 0 {
		t.Fatalf("repair merged %d by name (err %v), want 0", merged, err)
	}

	if performers := countRows(t, ctx, database, `SELECT count(*) FROM performer`, nil); performers != 1 {
		t.Fatalf("performers = %d, want 1", performers)
	}
	if kept := countRows(t, ctx, database, `SELECT count(*) FROM performer WHERE id = $id`, db.Vars{"id": second}); kept != 1 {
		t.Fatalf("the credited record %s did not survive (other was %s)", second, first)
	}
}

// TestMergeByIDChecksKinds checks the user-facing merge: same kind merges,
// different kinds are refused.
func TestMergeByIDChecksKinds(t *testing.T) {
	service, database, ctx := openTestService(t)
	first := insertLegacyPerformer(t, ctx, database, map[string]any{"name": "Kaci Star"})
	second := insertLegacyPerformer(t, ctx, database, map[string]any{"name": "Kaci Starr"})
	studio := insertLegacyStudio(t, ctx, database, map[string]any{"name": "Studio"})

	if err := service.MergeByID(ctx, first, studio); err == nil {
		t.Fatalf("a performer was merged into a studio")
	}
	if err := service.MergeByID(ctx, first, second); err != nil {
		t.Fatalf("merge: %v", err)
	}
	if performers := countRows(t, ctx, database, `SELECT count(*) FROM performer`, nil); performers != 1 {
		t.Fatalf("performers = %d, want 1", performers)
	}
	merged, err := service.GetByID(ctx, first)
	if err != nil || merged == nil {
		t.Fatalf("survivor missing: %v", err)
	}
	if aliases := jsonStringList(merged["aliases"]); len(aliases) != 1 || aliases[0] != "Kaci Starr" {
		t.Fatalf("aliases = %v, want the merged name kept as an alias", aliases)
	}
}
