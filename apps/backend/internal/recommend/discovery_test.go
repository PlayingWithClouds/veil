package recommend

import (
	"context"
	"testing"
	"time"
)

func TestFeedShowsOneCopyOfAVideoSeveralSitesCarry(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertRow(t, database, "performer", "performer:a", map[string]any{"name": "A"})
	insertRow(t, database, "performer", "performer:b", map[string]any{"name": "B"})
	insertScene(t, database, "scene:siteone", map[string]any{"title": "First Title", "performers": []string{"performer:a"}, "duration_seconds": 1200})
	insertScene(t, database, "scene:sitetwo", map[string]any{"title": "Other Words Entirely", "performers": []string{"performer:a"}, "duration_seconds": 1210})
	insertScene(t, database, "scene:otherruntime", map[string]any{"title": "Different", "performers": []string{"performer:a"}, "duration_seconds": 600})
	insertScene(t, database, "scene:otherperformer", map[string]any{"title": "Another", "performers": []string{"performer:b"}, "duration_seconds": 1200})
	insertScene(t, database, "scene:unknownruntime", map[string]any{"title": "Stub", "performers": []string{"performer:a"}})

	items := feed(t, engine, nil)
	copies := 0
	for _, sceneID := range []string{"scene:siteone", "scene:sitetwo"} {
		if itemBySceneID(items, sceneID) != nil {
			copies++
		}
	}
	if copies != 1 {
		t.Errorf("exactly one of the two copies must be listed, got %d", copies)
	}
	for _, sceneID := range []string{"scene:otherruntime", "scene:otherperformer", "scene:unknownruntime"} {
		if itemBySceneID(items, sceneID) == nil {
			t.Errorf("%s is not a copy and must stay", sceneID)
		}
	}
}

func TestFeedFilteredDropsScenesCarryingTags(t *testing.T) {
	ctx := context.Background()
	engine, database, _ := newTestEngine(t)
	insertRow(t, database, "tag", "tag:bad", map[string]any{"name": "bad"})
	insertRow(t, database, "studio", "studio:tagged", map[string]any{"name": "Tagged", "tags": []string{"tag:bad"}})
	insertScene(t, database, "scene:own", map[string]any{"tags": []string{"tag:bad"}})
	insertScene(t, database, "scene:viastudio", map[string]any{"studio": "studio:tagged"})
	insertScene(t, database, "scene:clean", map[string]any{})

	items, err := engine.FeedFiltered(ctx, nil, nil, FeedFilter{ExcludeTagIDs: []string{"tag:bad"}}, 100, 0, true)
	if err != nil {
		t.Fatalf("feed: %v", err)
	}
	if len(items) != 1 || items[0].SceneID != "scene:clean" {
		t.Errorf("only the clean scene may remain, got %+v", items)
	}
}

func TestTasteListsLikedThenDislikedEntries(t *testing.T) {
	ctx := context.Background()
	engine, database, _ := newTestEngine(t)
	insertRow(t, database, "tag", "tag:loved", map[string]any{"name": "loved"})
	insertRow(t, database, "tag", "tag:hated", map[string]any{"name": "hated"})
	insertScene(t, database, "scene:good", map[string]any{"tags": []string{"tag:loved"}})
	insertScene(t, database, "scene:bad", map[string]any{"tags": []string{"tag:hated"}})
	insertRow(t, database, "user_rating", "user_rating:good", map[string]any{"media": "scene:good", "rating": 10, "updated_at": ago(time.Hour)})
	insertRow(t, database, "user_rating", "user_rating:bad", map[string]any{"media": "scene:bad", "rating": 1, "updated_at": ago(time.Hour)})

	taste, err := engine.Taste(ctx, 5)
	if err != nil {
		t.Fatalf("taste: %v", err)
	}
	if taste.SignalCount == 0 || len(taste.Tags) != 2 {
		t.Fatalf("taste = %+v", taste)
	}
	if taste.Tags[0].Name != "loved" || taste.Tags[0].Affinity <= 0 {
		t.Errorf("liked tag must lead: %+v", taste.Tags[0])
	}
	if taste.Tags[1].Name != "hated" || taste.Tags[1].Affinity >= 0 {
		t.Errorf("disliked tag must follow: %+v", taste.Tags[1])
	}
}
