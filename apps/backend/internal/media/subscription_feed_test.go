package media

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
)

func TestSubscriptionFeed(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	for _, slug := range []string{"old", "shared", "fresh"} {
		insertRow(t, database, "scene", "scene:"+slug, map[string]any{"source_url": "https://example.test/" + slug, "title": slug})
	}
	first, _, err := repo.SubscribeSearch(ctx, "first", nil, 0)
	if err != nil {
		t.Fatalf("subscribe first: %v", err)
	}
	second, _, err := repo.SubscribeSearch(ctx, "second", nil, 0)
	if err != nil {
		t.Fatalf("subscribe second: %v", err)
	}
	repo.RecordSearchSubscriptionRun(ctx, first.ID, []string{"scene:old", "scene:shared"}, nil)
	database.Exec(ctx, "UPDATE search_subscription_item SET first_seen_at = '1999-01-01T00:00:00.000Z'", nil)
	database.Exec(ctx, "UPDATE search_subscription SET seen_at = '2000-01-01T00:00:00.000Z'", nil)
	repo.RecordSearchSubscriptionRun(ctx, second.ID, []string{"scene:shared", "scene:fresh"}, nil)
	database.Exec(ctx, "UPDATE search_subscription SET seen_at = '2000-01-01T00:00:00.000Z' WHERE id = $id", map[string]any{"id": second.ID})

	items, err := repo.SubscriptionFeed(ctx, nil, 10, 0, nil)
	if err != nil {
		t.Fatalf("feed: %v", err)
	}
	if len(items) != 3 || items[0].Scene.ID != "scene:fresh" || !items[0].IsNew {
		t.Fatalf("feed must list each scene once, newest find first: %+v", items)
	}
	for _, item := range items {
		if item.Scene.ID == "scene:shared" && (item.SubscriptionID != first.ID || item.IsNew) {
			t.Fatalf("a scene found by several subscriptions belongs to the first finder: %+v", item)
		}
	}

	newOnly := true
	items, _ = repo.SubscriptionFeed(ctx, &model.SubscriptionFeedFilter{NewOnly: &newOnly}, 10, 0, nil)
	if len(items) != 1 || items[0].Scene.ID != "scene:fresh" {
		t.Fatalf("new only: %+v", items)
	}

	items, _ = repo.SubscriptionFeed(ctx, &model.SubscriptionFeedFilter{SubscriptionID: &second.ID}, 10, 0, nil)
	if len(items) != 2 {
		t.Fatalf("one subscription's feed: %+v", items)
	}

	insertRow(t, database, "watch_history", "watch_history:1", map[string]any{"media": "scene:fresh"})
	unwatched := true
	items, _ = repo.SubscriptionFeed(ctx, &model.SubscriptionFeedFilter{UnwatchedOnly: &unwatched}, 10, 0, nil)
	if len(items) != 2 {
		t.Fatalf("unwatched only: %+v", items)
	}

	items, _ = repo.SubscriptionFeed(ctx, &model.SubscriptionFeedFilter{Kinds: []model.SubscriptionKind{model.SubscriptionKindPerformer}}, 10, 0, nil)
	if len(items) != 0 {
		t.Fatalf("kind filter: %+v", items)
	}
}
