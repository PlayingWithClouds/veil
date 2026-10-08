package media

import (
	"context"
	"testing"
)

func TestListScenesIncludesAndExcludesTagLists(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	insertRow(t, database, "tag", "tag:one", map[string]any{"name": "one"})
	insertRow(t, database, "tag", "tag:two", map[string]any{"name": "two"})
	insertRow(t, database, "scene", "scene:both", map[string]any{"title": "both", "source_url": "https://e.test/both", "tags": []string{"tag:one", "tag:two"}})
	insertRow(t, database, "scene", "scene:onlyone", map[string]any{"title": "onlyone", "source_url": "https://e.test/onlyone", "tags": []string{"tag:one"}})
	insertRow(t, database, "scene", "scene:none", map[string]any{"title": "none", "source_url": "https://e.test/none"})

	included, err := repo.ListScenes(ctx, EntityFilters{IncludeTagIDs: []string{"tag:one", "tag:two"}}, nil, nil)
	if err != nil {
		t.Fatalf("include: %v", err)
	}
	if len(included) != 1 || included[0].ID != "scene:both" {
		t.Errorf("include both tags = %v", sceneIDs(included))
	}

	excluded, err := repo.ListScenes(ctx, EntityFilters{ExcludeTagIDs: []string{"tag:two"}}, nil, nil)
	if err != nil {
		t.Fatalf("exclude: %v", err)
	}
	if len(excluded) != 2 {
		t.Errorf("exclude tag:two = %v, want onlyone and none", sceneIDs(excluded))
	}

	mixed, err := repo.ListScenes(ctx, EntityFilters{IncludeTagIDs: []string{"tag:one"}, ExcludeTagIDs: []string{"tag:two"}}, nil, nil)
	if err != nil {
		t.Fatalf("mixed: %v", err)
	}
	if len(mixed) != 1 || mixed[0].ID != "scene:onlyone" {
		t.Errorf("include one exclude two = %v", sceneIDs(mixed))
	}
}
