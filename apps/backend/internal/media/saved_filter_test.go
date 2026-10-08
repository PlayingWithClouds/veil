package media

import (
	"context"
	"testing"
)

// TestSavedFilterKeepsSourcesAndDurations checks that a preset round-trips its sites and duration window.
func TestSavedFilterKeepsSourcesAndDurations(t *testing.T) {
	ctx := context.Background()
	repo, _ := newTestRepository(t)
	minDuration := 600
	maxDuration := 1800

	created, err := repo.CreateSavedFilter(ctx, "Mid length", SavedFilterInput{
		Sources:     []string{"eporner", "xhamster"},
		MinDuration: &minDuration,
		MaxDuration: &maxDuration,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if len(created.Filter.Sources) != 2 || *created.Filter.MinDuration != 600 || *created.Filter.MaxDuration != 1800 {
		t.Fatalf("created filter lost fields: %+v", created.Filter)
	}

	listed, err := repo.ListSavedFilters(ctx)
	if err != nil || len(listed) != 1 {
		t.Fatalf("list: %v %v", listed, err)
	}
	if got := listed[0].Filter.Sources; len(got) != 2 || got[0] != "eporner" || got[1] != "xhamster" {
		t.Fatalf("listed sources: %v", got)
	}
}
