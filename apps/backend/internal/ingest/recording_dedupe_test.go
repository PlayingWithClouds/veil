package ingest

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// TestRecordingsOfOneModelAreNotMergedAsDuplicates covers two live recordings
// with the same performer and runtime: they are different footage.
func TestRecordingsOfOneModelAreNotMergedAsDuplicates(t *testing.T) {
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := New(database)

	var sceneIDs []string
	for _, name := range []string{"a", "b"} {
		result, err := service.IngestScrapeResults(ctx, "alpha", RecordingSourcePrefix+name, []*plugins.ScrapeResult{{
			Type: plugins.MediaTypeScene,
			Scene: &plugins.Scene{
				Type: plugins.MediaTypeScene, SourceURL: RecordingSourcePrefix + name, Title: "Nova — live " + name,
				Duration: 14400, Performers: []plugins.ScenePerformer{{Name: "Nova"}},
			},
		}})
		if err != nil {
			t.Fatal(err)
		}
		sceneIDs = append(sceneIDs, result.CanonicalID)
	}
	if sceneIDs[0] == sceneIDs[1] {
		t.Fatal("the second recording was merged into the first at ingest")
	}
	if _, err := service.MergeDuplicateMedia(ctx); err != nil {
		t.Fatal(err)
	}
	count, _ := database.Int(ctx, `SELECT count(*) FROM scene`, nil)
	if count != 2 {
		t.Fatalf("%d scenes left after the startup merge, want 2", count)
	}
}
