package recommend

import (
	"context"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// insertFind records a feeder find found age before testNow.
func insertFind(t *testing.T, database *db.DB, sceneID, source, entity string, rank int, age time.Duration) {
	t.Helper()
	_, err := database.Exec(context.Background(),
		`INSERT INTO feed_discovery (scene, source, entity, rank, found_at) VALUES ($scene, $source, $entity, $rank, $at)`,
		db.Vars{"scene": sceneID, "source": source, "entity": entity, "rank": rank, "at": ago(age)})
	if err != nil {
		t.Fatalf("insert find: %v", err)
	}
}

func TestColdStartFeedServesCatalogFinds(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertScene(t, database, "scene:catalog", map[string]any{})
	insertScene(t, database, "scene:stale", map[string]any{})
	insertFind(t, database, "scene:catalog", "catalog", "eporner", 0, time.Hour)
	insertFind(t, database, "scene:stale", "catalog", "eporner", 1, 2*feedDiscoveryWindow)

	items := feed(t, engine, nil)
	catalog := itemBySceneID(items, "scene:catalog")
	if catalog == nil || catalog.Source != SourceCatalog || catalog.Reason.Text != "Fresh on eporner" {
		t.Errorf("catalog find: %+v", catalog)
	}
	if stale := itemBySceneID(items, "scene:stale"); stale == nil || stale.Source != SourceNewest {
		t.Errorf("a find older than the window is only newest filler: %+v", stale)
	}
}

func TestTasteFindsFollowCurrentAffinity(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertRow(t, database, "performer", "performer:star", map[string]any{"name": "Star"})
	insertRow(t, database, "performer", "performer:faded", map[string]any{"name": "Faded"})
	insertScene(t, database, "scene:loved", map[string]any{"performers": []string{"performer:star"}})
	insertRow(t, database, "o_event", "o_event:1", map[string]any{"media": "scene:loved", "created_at": ago(time.Hour)})
	insertScene(t, database, "scene:found", map[string]any{})
	insertScene(t, database, "scene:orphan", map[string]any{})
	insertFind(t, database, "scene:found", "taste", "performer:star", 0, time.Hour)
	insertFind(t, database, "scene:orphan", "taste", "performer:faded", 0, time.Hour)

	items := feed(t, engine, nil)
	found := itemBySceneID(items, "scene:found")
	if found == nil || found.Source != SourceTaste || found.Reason.Text != "More from Star" {
		t.Errorf("taste find: %+v", found)
	}
	if orphan := itemBySceneID(items, "scene:orphan"); orphan != nil && orphan.Source == SourceTaste {
		t.Errorf("a find for an entity without affinity is no taste candidate: %+v", orphan)
	}
}

func TestShortPoolAsksForTopUp(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertScene(t, database, "scene:only", map[string]any{})
	asked := make(chan struct{}, 1)
	engine.OnLowPool(func() { asked <- struct{}{} })

	feed(t, engine, nil)
	select {
	case <-asked:
	case <-time.After(time.Second):
		t.Fatal("a ranking with one candidate did not ask for a top-up")
	}
}

func TestInvalidateDropsTheMemoizedRanking(t *testing.T) {
	ctx := context.Background()
	engine, database, _ := newTestEngine(t)
	insertScene(t, database, "scene:a", map[string]any{})
	if _, err := engine.Feed(ctx, nil, nil, 10, 0, false); err != nil {
		t.Fatalf("feed: %v", err)
	}
	insertScene(t, database, "scene:late", map[string]any{"created_at": ago(0)})
	engine.Invalidate()
	items, err := engine.Feed(ctx, nil, nil, 10, 0, false)
	if err != nil {
		t.Fatalf("feed: %v", err)
	}
	if itemBySceneID(items, "scene:late") == nil {
		t.Errorf("feed after Invalidate misses the new scene: %+v", items)
	}
}
