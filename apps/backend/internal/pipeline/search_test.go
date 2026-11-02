package pipeline

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestSearchMemory keeps a site's answer to a search page, found again for
// the same query spelled differently, and not for another page or query.
func TestSearchMemory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	flow := &Pipeline{searches: cache.New(ctx, database)}

	answer := storedAnswer{Refs: []schema.RecordRef{{Kind: schema.RecordKindVideo, Id: "v1"}}, Exhausted: true}
	flow.storeSearch(ctx, searchKey(plugins.MediaTypeScene, "site", "Mia  Malkova", 0, 20), answer)

	stored, ok := flow.storedSearch(ctx, searchKey(plugins.MediaTypeScene, "site", " mia malkova", 0, 20))
	if !ok || len(stored.Refs) != 1 || stored.Refs[0].Id != "v1" || !stored.Exhausted {
		t.Errorf("stored answer = %+v, %v", stored, ok)
	}
	if _, ok := flow.storedSearch(ctx, searchKey(plugins.MediaTypeScene, "site", "mia malkova", 1, 20)); ok {
		t.Error("another page found the first page's answer")
	}
	if _, ok := flow.storedSearch(ctx, searchKey(plugins.MediaTypeScene, "site", "mia", 0, 20)); ok {
		t.Error("another query found the answer")
	}
}
