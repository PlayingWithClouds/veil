package recommend

import (
	"context"
	"fmt"
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

func TestExploreFindsServeAdjacentTagsFromTheSites(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertRow(t, database, "tag", "tag:adjacent", map[string]any{"name": "Adjacent"})
	insertRow(t, database, "tag", "tag:disliked", map[string]any{"name": "Disliked"})
	insertScene(t, database, "scene:adjacent", map[string]any{})
	insertScene(t, database, "scene:hated", map[string]any{})
	insertScene(t, database, "scene:disliked", map[string]any{"tags": []string{"tag:disliked"}})
	insertRow(t, database, "user_rating", "user_rating:1", map[string]any{"media": "scene:disliked", "rating": 1, "updated_at": ago(time.Hour)})
	insertFind(t, database, "scene:adjacent", "explore", "tag:adjacent", 0, time.Hour)
	insertFind(t, database, "scene:hated", "explore", "tag:disliked", 0, time.Hour)

	items := feed(t, engine, nil)
	found := itemBySceneID(items, "scene:adjacent")
	if found == nil || found.Source != SourceExploreSite || found.Reason.Text != "Explore Adjacent" {
		t.Errorf("explore find: %+v", found)
	}
	if hated := itemBySceneID(items, "scene:hated"); hated != nil && hated.Source == SourceExploreSite {
		t.Errorf("a find for a disliked tag must not be proposed: %+v", hated)
	}
}

func TestPopularFindsNameTheirSite(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertScene(t, database, "scene:popular", map[string]any{})
	insertFind(t, database, "scene:popular", "popular", "www.pornhat.com", 0, time.Hour)

	popular := itemBySceneID(feed(t, engine, nil), "scene:popular")
	if popular == nil || popular.Source != SourcePopular || popular.Reason.Text != "Popular on www.pornhat.com" {
		t.Errorf("popular find: %+v", popular)
	}
}

func TestTasteFindsSurviveAThinProfile(t *testing.T) {
	// A single click leaves the performer's affinity tiny; the find still
	// carries more than the affinity-scaled prior of the old formula.
	engine, database, _ := newTestEngine(t)
	insertRow(t, database, "performer", "performer:star", map[string]any{"name": "Star"})
	insertScene(t, database, "scene:clicked", map[string]any{"performers": []string{"performer:star"}})
	insertRow(t, database, "o_event", "o_event:1", map[string]any{"media": "scene:clicked", "created_at": ago(time.Hour)})
	insertRow(t, database, "performer", "performer:other", map[string]any{"name": "Other"})
	insertScene(t, database, "scene:other", map[string]any{"performers": []string{"performer:other"}})
	insertRow(t, database, "watch_history", "watch_history:1", map[string]any{
		"media": "scene:other", "progress_seconds": 3, "max_progress_seconds": 3, "duration_seconds": 600, "updated_at": ago(time.Hour),
	})
	insertScene(t, database, "scene:found", map[string]any{})
	insertFind(t, database, "scene:found", "taste", "performer:star", 0, time.Hour)

	state, err := engine.takeSnapshot(context.Background(), nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	candidates, err := engine.tasteCandidates(context.Background(), state)
	if err != nil || len(candidates) != 1 {
		t.Fatalf("taste candidates = %+v (%v)", candidates, err)
	}
	affinity := state.profile.Performers["performer:star"]
	want := sourcePriors[SourceTaste] * (tasteAffinityFloor + (1-tasteAffinityFloor)*affinity)
	assertClose(t, "prior", candidates[0].prior, want)
	if candidates[0].prior < sourcePriors[SourceTaste]*tasteAffinityFloor {
		t.Errorf("prior %v below the floor", candidates[0].prior)
	}
}

func TestTasteFindsAreCappedPerEntity(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertRow(t, database, "performer", "performer:star", map[string]any{"name": "Star"})
	insertScene(t, database, "scene:loved", map[string]any{"performers": []string{"performer:star"}})
	insertRow(t, database, "o_event", "o_event:1", map[string]any{"media": "scene:loved", "created_at": ago(time.Hour)})
	for index := range tastePerEntityLimit + 5 {
		sceneID := fmt.Sprintf("scene:found%d", index)
		insertScene(t, database, sceneID, map[string]any{})
		insertFind(t, database, sceneID, "taste", "performer:star", index, time.Hour)
	}
	state, err := engine.takeSnapshot(context.Background(), nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	candidates, err := engine.tasteCandidates(context.Background(), state)
	if err != nil || len(candidates) != tastePerEntityLimit {
		t.Errorf("%d taste candidates (%v), want %d", len(candidates), err, tastePerEntityLimit)
	}
}
