package recommend

import (
	"context"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/media"
)

// testNow is the fixed clock every engine test runs at.
var testNow = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

// newTestEngine opens a fresh database and returns an engine on a fixed clock.
func newTestEngine(t *testing.T) (*Engine, *db.DB, *media.Repository) {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	repo := media.NewRepository(database)
	engine := New(database, repo)
	engine.now = func() time.Time { return testNow }
	return engine, database, repo
}

// insertRow inserts fields with an explicit id and fails the test on error.
func insertRow(t *testing.T, database *db.DB, table, id string, fields map[string]any) {
	t.Helper()
	fields["id"] = id
	if _, err := database.Insert(context.Background(), table, fields); err != nil {
		t.Fatalf("insert %s: %v", id, err)
	}
}

// insertScene inserts a scene stored a day before testNow, with extra fields.
func insertScene(t *testing.T, database *db.DB, id string, fields map[string]any) {
	t.Helper()
	fields["source_url"] = "https://example.test/" + id
	if _, found := fields["title"]; !found {
		fields["title"] = id
	}
	if _, found := fields["created_at"]; !found {
		fields["created_at"] = ago(24 * time.Hour)
	}
	insertRow(t, database, "scene", id, fields)
}

// ago renders the stored timestamp age before testNow.
func ago(age time.Duration) string {
	return db.FormatTime(testNow.Add(-age))
}

// assertClose fails when got differs from want by more than a rounding error.
func assertClose(t *testing.T, label string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}

// feed returns the first page of a freshly ranked feed, failing the test on
// error.
func feed(t *testing.T, engine *Engine, disabledPlugins []string) []Item {
	t.Helper()
	items, err := engine.Feed(context.Background(), disabledPlugins, 100, 0, true)
	if err != nil {
		t.Fatalf("feed: %v", err)
	}
	return items
}

// itemBySceneID finds a scene in the feed; nil when absent.
func itemBySceneID(items []Item, sceneID string) *Item {
	for index := range items {
		if items[index].SceneID == sceneID {
			return &items[index]
		}
	}
	return nil
}
