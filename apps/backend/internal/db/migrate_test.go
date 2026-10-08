package db

import (
	"context"
	"testing"
)

// TestRefetchAyloScenesMigration checks that 009 reopens only visited aylo
// scenes that have no stream.
func TestRefetchAyloScenesMigration(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)

	scenes := map[string]string{
		"scene:phnostream": "https://www.pornhub.com/view_video.php?viewkey=1",
		"scene:rtnostream": "https://www.redtube.com/2",
		"scene:phstream":   "https://www.pornhub.com/view_video.php?viewkey=3",
		"scene:epnostream": "https://www.eporner.com/video-4/",
	}
	for id, sourceURL := range scenes {
		if _, err := database.Exec(ctx,
			`INSERT INTO scene (id, source_url, detail_fetched_at) VALUES ($id, $url, '2026-01-01T00:00:00Z')`,
			Vars{"id": id, "url": sourceURL}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := database.Exec(ctx,
		`INSERT INTO stream (id, media, url) VALUES ('stream:1', 'scene:phstream', 'https://www.pornhub.com/view_video.php?viewkey=3')`,
		nil); err != nil {
		t.Fatal(err)
	}

	if _, err := database.Exec(ctx, `DELETE FROM _migration WHERE version = 9`, nil); err != nil {
		t.Fatal(err)
	}
	if err := database.migrate(ctx); err != nil {
		t.Fatal(err)
	}

	reopened, err := database.Strings(ctx, `SELECT id FROM scene WHERE detail_fetched_at IS NULL ORDER BY id`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(reopened) != 2 || reopened[0] != "scene:phnostream" || reopened[1] != "scene:rtnostream" {
		t.Fatalf("reopened scenes = %v, want [scene:phnostream scene:rtnostream]", reopened)
	}
}
