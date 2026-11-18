package dedupe

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestPassMergesTiedPerformers merges a performer into the one whose alias
// is their name when a video credits both, and leaves a namesake with
// another birthdate apart.
func TestPassMergesTiedPerformers(t *testing.T) {
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	insertPerformer := func(name string, aliases []string, birthdate string) string {
		attributes := map[string]any{}
		if birthdate != "" {
			attributes["birthdate"] = birthdate
		}
		id, err := database.Insert(ctx, "Performer", db.Vars{"name": name, "nameKey": names.Key(name), "aliases": aliases, "attributes": attributes, "links": []string{}, "tags": []string{}})
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	full := insertPerformer("Lena Paul", []string{"Lena"}, "1993-07-02")
	short := insertPerformer("Lena", []string{}, "")
	namesake := insertPerformer("Lena", []string{"Lena Paul"}, "1980-01-01")
	for _, credits := range [][]string{{full}, {full, short}, {namesake}} {
		if _, err := database.Insert(ctx, "Video", db.Vars{"title": "Scene", "status": schema.MediaStatusAvailable, "performers": credits, "tags": []string{}, "durationSeconds": 600}); err != nil {
			t.Fatal(err)
		}
	}
	lib := library.New(database)
	if err := New(lib, cache.New(ctx, database), fakeImages{}).Pass(ctx); err != nil {
		t.Fatal(err)
	}
	if survivor, _ := lib.Performer(ctx, full); survivor == nil {
		t.Fatal("the performer with more credits is gone")
	}
	if gone, _ := lib.Performer(ctx, short); gone != nil {
		t.Error("the tied performer wasn't merged")
	}
	if kept, _ := lib.Performer(ctx, namesake); kept == nil {
		t.Error("a namesake with another birthdate was merged")
	}
}
