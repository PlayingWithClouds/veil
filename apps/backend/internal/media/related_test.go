package media

import (
	"context"
	"slices"
	"testing"
)

func TestRelatedScenesOrderBySourceThenRankAndSkipBlocked(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	insertRow(t, database, "tag", "tag:blocked", map[string]any{"name": "blocked"})
	for _, slug := range []string{"visited", "site0", "site1", "performer0", "tag0", "blocked"} {
		fields := map[string]any{"source_url": "https://example.test/" + slug, "title": slug}
		if slug == "blocked" {
			fields["tags"] = []string{"tag:blocked"}
		}
		insertRow(t, database, "scene", "scene:"+slug, fields)
	}
	edges := []struct {
		related string
		source  string
		rank    int
	}{
		{"scene:tag0", "tag", 0},
		{"scene:performer0", "performer", 0},
		{"scene:site1", "site", 1},
		{"scene:blocked", "site", 2},
		{"scene:site0", "site", 0},
	}
	for index, edge := range edges {
		insertRow(t, database, "scene_related", "scene_related:"+string(rune('a'+index)), map[string]any{
			"scene": "scene:visited", "related": edge.related, "source": edge.source, "rank": edge.rank,
		})
	}
	if _, err := repo.AddBlock(ctx, "tag", "tag:blocked", nil); err != nil {
		t.Fatalf("block: %v", err)
	}

	scenes, err := repo.RelatedScenes(ctx, "scene:visited", 10, nil)
	if err != nil {
		t.Fatalf("related: %v", err)
	}
	titles := []string{}
	for _, scene := range scenes {
		titles = append(titles, scene.Title)
	}
	if !slices.Equal(titles, []string{"site0", "site1", "performer0", "tag0"}) {
		t.Errorf("related = %v", titles)
	}

	limited, _ := repo.RelatedScenes(ctx, "scene:visited", 2, nil)
	if len(limited) != 2 {
		t.Errorf("limit 2 returned %d", len(limited))
	}
}
