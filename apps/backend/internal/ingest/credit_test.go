package ingest

import (
	"slices"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// TestCreditPerformerCreatesAndMatchesByName credits a new name (creating the
// performer), then a differently spelled one, which must reuse the record.
func TestCreditPerformerCreatesAndMatchesByName(t *testing.T) {
	service, database, ctx := openTestService(t)
	scene := ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: testSceneURL, Title: "Some Scene",
		Downloads: []plugins.Download{{URL: "https://cdn.test/a.mp4"}},
	})

	if id, err := service.FindPerformer(ctx, "Lexi Lore"); err != nil || id != "" {
		t.Fatalf("performer should not exist yet: %q, %v", id, err)
	}
	firstID, err := service.CreditPerformer(ctx, scene.CanonicalID, "Lexi Lore")
	if err != nil {
		t.Fatal(err)
	}
	secondID, err := service.CreditPerformer(ctx, scene.CanonicalID, "  lexi  lore ")
	if err != nil {
		t.Fatal(err)
	}
	if firstID != secondID {
		t.Fatalf("spelling variant created a second performer: %s vs %s", firstID, secondID)
	}
	if found, _ := service.FindPerformer(ctx, "LEXI LORE"); found != firstID {
		t.Fatalf("FindPerformer = %q, want %q", found, firstID)
	}

	row, err := database.GetString(ctx, scene.CanonicalID)
	if err != nil {
		t.Fatal(err)
	}
	if performers := jsonStringList(row["performers"]); !slices.Equal(performers, []string{firstID}) {
		t.Errorf("scene performers = %v", performers)
	}
	credits := countRows(t, ctx, database, `SELECT COUNT(*) FROM performs_in WHERE media = $media`, db.Vars{"media": scene.CanonicalID})
	if credits != 1 {
		t.Errorf("credit rows = %d, want 1", credits)
	}
}

// TestCreditPerformerRejectsOtherRecords refuses media that cannot be credited.
func TestCreditPerformerRejectsOtherRecords(t *testing.T) {
	service, _, ctx := openTestService(t)
	if _, err := service.CreditPerformer(ctx, "tag:abc", "Lexi Lore"); err == nil {
		t.Error("crediting a tag should fail")
	}
	if _, err := service.CreditPerformer(ctx, "scene:missing", "Lexi Lore"); err == nil {
		t.Error("crediting a missing scene should fail")
	}
}
