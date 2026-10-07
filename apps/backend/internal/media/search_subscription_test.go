package media

import (
	"context"
	"errors"
	"testing"
)

func TestSearchSubscriptionLifecycle(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	for _, slug := range []string{"first", "second", "third"} {
		insertRow(t, database, "scene", "scene:"+slug, map[string]any{"source_url": "https://example.test/" + slug, "title": slug})
	}

	subscription, created, err := repo.SubscribeSearch(ctx, "  Big Query ", []string{"eporner"}, 0)
	if err != nil || !created {
		t.Fatalf("subscribe: created=%v err=%v", created, err)
	}
	if subscription.IntervalHours != DefaultSearchSubscriptionHours || len(subscription.Sources) != 1 || !subscription.Enabled {
		t.Fatalf("defaults: %+v", subscription)
	}
	again, created, err := repo.SubscribeSearch(ctx, "big query", nil, 2)
	if err != nil || created || again.ID != subscription.ID {
		t.Fatalf("same normalized query must return the existing subscription: created=%v err=%v %+v", created, err, again)
	}

	due, _ := repo.DueSearchSubscriptions(ctx)
	if len(due) != 1 {
		t.Fatalf("a never-run subscription is due, got %d", len(due))
	}

	// First run: everything lands in the feed but counts as already seen.
	if err := repo.RecordSearchSubscriptionRun(ctx, subscription.ID, []string{"scene:first", "scene:second", "gallery:x"}, nil); err != nil {
		t.Fatalf("first run: %v", err)
	}
	subscription, _ = repo.SearchSubscription(ctx, subscription.ID)
	assertSceneCounts(t, repo, subscription.ID, 2, 0)
	if subscription.LastRunAt == nil {
		t.Fatal("first run must set last_run_at")
	}
	due, _ = repo.DueSearchSubscriptions(ctx)
	if len(due) != 0 {
		t.Fatal("a just-run subscription must not be due")
	}

	// Later run: only the unseen scene is new; a repeat is not added twice.
	database.Exec(ctx, "UPDATE search_subscription SET seen_at = '2000-01-01T00:00:00.000Z'", nil)
	database.Exec(ctx, "UPDATE search_subscription_item SET first_seen_at = '1999-01-01T00:00:00.000Z'", nil)
	if err := repo.RecordSearchSubscriptionRun(ctx, subscription.ID, []string{"scene:second", "scene:third"}, errors.New("missav: 403")); err != nil {
		t.Fatalf("second run: %v", err)
	}
	subscription, _ = repo.SearchSubscription(ctx, subscription.ID)
	assertSceneCounts(t, repo, subscription.ID, 3, 1)
	if subscription.LastError == nil || *subscription.LastError != "missav: 403" {
		t.Fatalf("last error: %v", subscription.LastError)
	}
	scenes, err := repo.SearchSubscriptionScenes(ctx, subscription.ID, 10, 0, nil)
	if err != nil || len(scenes) != 3 || scenes[0].ID != "scene:third" {
		t.Fatalf("feed must list newest first: %v %d", err, len(scenes))
	}

	if _, err := repo.MarkSearchSubscriptionSeen(ctx, subscription.ID); err != nil {
		t.Fatalf("mark seen: %v", err)
	}
	assertSceneCounts(t, repo, subscription.ID, 3, 0)

	interval := 12
	disabled := false
	subscription, err = repo.UpdateSearchSubscription(ctx, subscription.ID, &interval, &disabled)
	if err != nil || subscription.IntervalHours != 12 || subscription.Enabled {
		t.Fatalf("update: %v %+v", err, subscription)
	}

	if err := repo.DeleteSearchSubscription(ctx, subscription.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.SearchSubscription(ctx, subscription.ID); !errors.Is(err, ErrSearchSubscriptionNotFound) {
		t.Fatalf("deleted subscription lookup: %v", err)
	}
	remaining, _ := database.Int(ctx, "SELECT count(*) FROM search_subscription_item", nil)
	if remaining != 0 {
		t.Fatalf("feed items left after delete: %d", remaining)
	}
}

// assertSceneCounts checks a subscription's visible and unseen scene counts.
func assertSceneCounts(t *testing.T, repo *Repository, id string, wantTotal, wantUnseen int) {
	t.Helper()
	total, unseen, err := repo.SearchSubscriptionSceneCounts(context.Background(), id, nil)
	if err != nil || total != wantTotal || unseen != wantUnseen {
		t.Fatalf("counts: total=%d unseen=%d err=%v, want total=%d unseen=%d", total, unseen, err, wantTotal, wantUnseen)
	}
}
