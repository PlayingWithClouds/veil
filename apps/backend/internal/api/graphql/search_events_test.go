package graphql

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/searchrank"
)

// TestRecordSearchEventsFilesImpressionsWithAQuery checks only impressions
// carrying a query reach the search history, with the source as their plugin.
func TestRecordSearchEventsFilesImpressionsWithAQuery(t *testing.T) {
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	resolver := &Resolver{searchHistory: searchrank.NewHistory(database)}

	query, source, clicked, position := "blue lagoon", "alpha", true, 3
	resolver.recordSearchEvents(ctx, []*model.ImpressionInput{
		{MediaID: "scene:one", Query: &query, Source: &source, Clicked: &clicked, Position: &position},
		{MediaID: "scene:feed"},
	})

	rows, err := database.Query(ctx, `SELECT media, plugin, kind, position FROM search_event`, nil)
	if err != nil || len(rows) != 1 {
		t.Fatalf("rows = %v, %v", rows, err)
	}
	if rows[0]["media"] != "scene:one" || rows[0]["plugin"] != "alpha" || rows[0]["kind"] != "click" || db.AsInt(rows[0]["position"]) != 3 {
		t.Fatalf("row = %v", rows[0])
	}
}
