package recommend

import (
	"context"
	"strings"
	"testing"
	"time"
)

// seedWatchedScene stores a completed watch of scene:seed and links it to
// related scenes by source.
func seedWatchedScene(t *testing.T, engine *Engine, related map[string]string) {
	t.Helper()
	database := engine.database
	insertScene(t, database, "scene:seed", map[string]any{"title": "Seed Title", "performers": []string{"performer:seed"}})
	insertRow(t, database, "performer", "performer:seed", map[string]any{"name": "Seed Performer"})
	insertRow(t, database, "watch_history", "watch_history:seed", map[string]any{
		"media": "scene:seed", "progress_seconds": 600, "max_progress_seconds": 600,
		"duration_seconds": 600, "completed": true, "updated_at": ago(time.Hour),
	})
	rank := 0
	for relatedID, source := range related {
		insertScene(t, database, relatedID, map[string]any{})
		insertRow(t, database, "scene_related", "scene_related:"+strings.TrimPrefix(relatedID, "scene:"), map[string]any{
			"scene": "scene:seed", "related": relatedID, "source": source, "rank": rank,
		})
		rank++
	}
}

func TestFeedRanksSiteRelatedOfWatchedScenesFirst(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	seedWatchedScene(t, engine, map[string]string{"scene:site": "site", "scene:bytag": "tag"})
	insertScene(t, database, "scene:unrelated", map[string]any{})

	items := feed(t, engine, nil)
	if len(items) == 0 || items[0].SceneID != "scene:site" {
		t.Fatalf("site-related scene must lead the feed, got %+v", items)
	}
	if items[0].Source != SourceRelatedSite || items[0].Reason.Text != "Because you watched Seed Performer" {
		t.Errorf("site item: source %q, reason %q", items[0].Source, items[0].Reason.Text)
	}
	if items[1].SceneID != "scene:bytag" || items[1].Source != SourceRelatedTag {
		t.Errorf("tag-related scene must come second, got %+v", items[1])
	}
	seed := itemBySceneID(items, "scene:seed")
	if seed == nil || seed.Score >= itemBySceneID(items, "scene:unrelated").Score {
		t.Errorf("the completed scene must rank below unwatched filler: %+v", seed)
	}
}

func TestFeedExcludesBlockedDislikedAndInactiveStubs(t *testing.T) {
	ctx := context.Background()
	engine, database, repo := newTestEngine(t)
	seedWatchedScene(t, engine, map[string]string{"scene:blocked": "site", "scene:stub": "site", "scene:kept": "site"})
	if _, err := database.Exec(ctx, `UPDATE scene SET tags = '["tag:blocked"]' WHERE id = 'scene:blocked'`, nil); err != nil {
		t.Fatalf("tag blocked scene: %v", err)
	}
	if _, err := repo.AddBlock(ctx, "tag", "tag:blocked", nil); err != nil {
		t.Fatalf("block: %v", err)
	}
	insertScene(t, database, "scene:disliked", map[string]any{})
	insertRow(t, database, "user_rating", "user_rating:1", map[string]any{"media": "scene:disliked", "rating": 1, "updated_at": ago(time.Hour)})
	insertRow(t, database, "observation", "observation:stub", map[string]any{"target": "scene:stub", "plugin": "gone"})

	items := feed(t, engine, []string{"gone"})
	for _, excluded := range []string{"scene:blocked", "scene:stub", "scene:disliked"} {
		if itemBySceneID(items, excluded) != nil {
			t.Errorf("%s must not be recommended", excluded)
		}
	}
	if itemBySceneID(items, "scene:kept") == nil {
		t.Error("scene:kept must be recommended")
	}
	if itemBySceneID(feed(t, engine, nil), "scene:stub") == nil {
		t.Error("the stub must come back once its plugin is active")
	}
}

func TestColdStartServesSearchResultsThenNewest(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	searchedAt := time.Hour
	insertRow(t, database, "search_history", "search_history:1", map[string]any{
		"query": "Blondes", "normalized_query": "blondes", "created_at": ago(searchedAt),
	})
	insertScene(t, database, "scene:result1", map[string]any{"created_at": ago(searchedAt + time.Minute)})
	insertScene(t, database, "scene:result2", map[string]any{"created_at": ago(searchedAt + 30*time.Second)})
	insertScene(t, database, "scene:newer", map[string]any{"created_at": ago(10 * time.Minute)})
	insertScene(t, database, "scene:older", map[string]any{"created_at": ago(48 * time.Hour)})

	profile, err := engine.Profile(context.Background())
	if err != nil || !profile.Empty() {
		t.Fatalf("expected an empty profile, got %+v (%v)", profile, err)
	}
	items := feed(t, engine, nil)
	order := []string{}
	for _, item := range items {
		order = append(order, item.SceneID+"/"+item.Source)
	}
	want := []string{"scene:result1/search", "scene:result2/search", "scene:newer/newest", "scene:older/newest"}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Errorf("cold start order = %v, want %v", order, want)
	}
	if items[0].Reason.Text != "From your search “Blondes”" {
		t.Errorf("search reason = %q", items[0].Reason.Text)
	}
}

func TestFeedProposesNewSubscriptionFinds(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertRow(t, database, "search_subscription", "search_subscription:1", map[string]any{
		"query": "Redheads", "normalized_query": "redheads", "seen_at": ago(48 * time.Hour),
	})
	insertScene(t, database, "scene:new", map[string]any{})
	insertScene(t, database, "scene:seen", map[string]any{})
	insertRow(t, database, "search_subscription_item", "search_subscription_item:1", map[string]any{
		"subscription": "search_subscription:1", "media": "scene:new", "first_seen_at": ago(24 * time.Hour),
	})
	insertRow(t, database, "search_subscription_item", "search_subscription_item:2", map[string]any{
		"subscription": "search_subscription:1", "media": "scene:seen", "first_seen_at": ago(72 * time.Hour),
	})

	items := feed(t, engine, nil)
	fresh := itemBySceneID(items, "scene:new")
	if fresh == nil || fresh.Source != SourceSubscription || fresh.Reason.EntityID != "search_subscription:1" {
		t.Fatalf("new subscription find: %+v", fresh)
	}
	if fresh.Reason.Text != "New for “Redheads”" {
		t.Errorf("subscription reason = %q", fresh.Reason.Text)
	}
	if seen := itemBySceneID(items, "scene:seen"); seen == nil || seen.Source != SourceNewest {
		t.Errorf("an already-seen find is only newest filler: %+v", seen)
	}
}

func TestFeedProposesTopPerformerScenes(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertRow(t, database, "performer", "performer:star", map[string]any{"name": "Star"})
	insertScene(t, database, "scene:loved", map[string]any{"performers": []string{"performer:star"}})
	insertScene(t, database, "scene:more", map[string]any{"performers": []string{"performer:star"}})
	insertRow(t, database, "o_event", "o_event:1", map[string]any{"media": "scene:loved", "created_at": ago(time.Hour)})

	more := itemBySceneID(feed(t, engine, nil), "scene:more")
	if more == nil || more.Source != SourcePerformer || more.Reason.Text != "More from Star" {
		t.Errorf("performer candidate: %+v", more)
	}
}

func TestFeedPagesReuseTheRanking(t *testing.T) {
	ctx := context.Background()
	engine, database, _ := newTestEngine(t)
	for _, id := range []string{"scene:a", "scene:b", "scene:c", "scene:d"} {
		insertScene(t, database, id, map[string]any{})
	}
	first, _ := engine.Feed(ctx, nil, 2, 0, false)
	insertScene(t, database, "scene:late", map[string]any{"created_at": ago(0)})
	second, _ := engine.Feed(ctx, nil, 2, 2, false)
	if len(first) != 2 || len(second) != 2 {
		t.Fatalf("pages: %d + %d", len(first), len(second))
	}
	for _, item := range second {
		if item.SceneID == "scene:late" || itemBySceneID(first, item.SceneID) != nil {
			t.Errorf("page two must continue page one's ranking, got %s", item.SceneID)
		}
	}
	reloaded, _ := engine.Feed(ctx, nil, 2, 0, false)
	if reloaded[0].SceneID != first[0].SceneID || reloaded[1].SceneID != first[1].SceneID {
		t.Errorf("reloading page one within feedFirstPageTTL must keep the ranking, got %s, %s", reloaded[0].SceneID, reloaded[1].SceneID)
	}
	refreshed, _ := engine.Feed(ctx, nil, 1, 0, true)
	if refreshed[0].SceneID != "scene:late" {
		t.Errorf("a refresh must re-rank, got %s", refreshed[0].SceneID)
	}
}

func TestFirstPageReRanksOnceStale(t *testing.T) {
	ctx := context.Background()
	engine, database, _ := newTestEngine(t)
	insertScene(t, database, "scene:a", map[string]any{})
	engine.Feed(ctx, nil, 1, 0, false)
	insertScene(t, database, "scene:late", map[string]any{"created_at": ago(0)})
	engine.now = func() time.Time { return testNow.Add(feedFirstPageTTL + time.Minute) }
	stale, _ := engine.Feed(ctx, nil, 1, 0, false)
	if len(stale) == 0 || stale[0].SceneID != "scene:late" {
		t.Errorf("page one must re-rank after feedFirstPageTTL, got %+v", stale)
	}
}

func TestRowsGroupByReason(t *testing.T) {
	engine, _, _ := newTestEngine(t)
	seedWatchedScene(t, engine, map[string]string{"scene:r1": "site", "scene:r2": "site", "scene:r3": "performer"})

	rows, err := engine.Rows(context.Background(), nil, 5, 10)
	if err != nil {
		t.Fatalf("rows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected one row (newest filler is not a row), got %+v", rows)
	}
	if rows[0].Title != "Because you watched Seed Performer" || len(rows[0].Items) != 3 {
		t.Errorf("row: %q with %d items", rows[0].Title, len(rows[0].Items))
	}
}

func TestCategoriesFollowTagAffinity(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertScene(t, database, "scene:finished", map[string]any{"tags": []string{"tag:loved"}})
	insertScene(t, database, "scene:same", map[string]any{"tags": []string{"tag:loved"}})
	insertRow(t, database, "watch_history", "watch_history:1", map[string]any{
		"media": "scene:finished", "progress_seconds": 600, "max_progress_seconds": 600,
		"duration_seconds": 600, "completed": true, "updated_at": ago(time.Hour),
	})

	categories, err := engine.Categories(context.Background(), nil, 3, 5)
	if err != nil {
		t.Fatalf("categories: %v", err)
	}
	if len(categories) != 1 || categories[0].TagID != "tag:loved" {
		t.Fatalf("categories = %+v", categories)
	}
	if categories[0].Items[0].SceneID != "scene:same" {
		t.Errorf("the unwatched scene of the tag must lead, got %+v", categories[0].Items)
	}
}
