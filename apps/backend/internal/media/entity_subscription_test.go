package media

import (
	"context"
	"errors"
	"sort"
	"testing"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// seedFollowable creates a studio, performer and tag with credited, foreign
// and uncredited scenes around them.
func seedFollowable(t *testing.T, database *db.DB) {
	insertRow(t, database, "studio", "studio:vixen", map[string]any{
		"name": "Vixen", "image_path": "https://cdn.test/vixen.png", "source_url": "https://www.eporner.com/channel/vixen/",
	})
	insertRow(t, database, "studio", "studio:other", map[string]any{"name": "Other"})
	insertRow(t, database, "performer", "performer:riley", map[string]any{"name": "Riley"})
	insertRow(t, database, "tag", "tag:milf", map[string]any{"name": "milf"})
	insertRow(t, database, "scene", "scene:credited", map[string]any{
		"source_url": "https://example.test/credited", "studio": "studio:vixen",
		"performers": []string{"performer:riley"}, "tags": []string{"tag:milf"},
	})
	insertRow(t, database, "scene", "scene:foreign", map[string]any{
		"source_url": "https://example.test/foreign", "studio": "studio:other",
		"performers": []string{"performer:other"}, "tags": []string{"tag:other"},
	})
	insertRow(t, database, "scene", "scene:stub", map[string]any{"source_url": "https://example.test/stub"})
}

func TestSubscribeEntity(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	seedFollowable(t, database)

	subscription, created, err := repo.SubscribeEntity(ctx, model.SubscriptionKindStudio, "studio:vixen", 0)
	if err != nil || !created {
		t.Fatalf("subscribe: created=%v err=%v", created, err)
	}
	if subscription.Kind != model.SubscriptionKindStudio || subscription.Query != "Vixen" || subscription.IntervalHours != DefaultSearchSubscriptionHours {
		t.Fatalf("subscription: %+v", subscription)
	}
	target := subscription.Target
	if target == nil || target.ID != "studio:vixen" || target.Name != "Vixen" || target.ImageURL == nil || *target.ImageURL != "https://cdn.test/vixen.png" {
		t.Fatalf("target: %+v", target)
	}

	again, created, err := repo.SubscribeEntity(ctx, model.SubscriptionKindStudio, "studio:vixen", 2)
	if err != nil || created || again.ID != subscription.ID {
		t.Fatalf("a followed target must return the existing subscription: created=%v err=%v", created, err)
	}
	found, err := repo.SubscriptionForTarget(ctx, "studio:vixen")
	if err != nil || found.ID != subscription.ID {
		t.Fatalf("lookup by target: %v %+v", err, found)
	}
	if _, err := repo.SubscriptionForTarget(ctx, "studio:other"); !errors.Is(err, ErrSearchSubscriptionNotFound) {
		t.Fatalf("unfollowed target: %v", err)
	}

	// A search for the same words is a separate subscription.
	search, created, err := repo.SubscribeSearch(ctx, "vixen", nil, 0)
	if err != nil || !created || search.Kind != model.SubscriptionKindSearch || search.Target != nil {
		t.Fatalf("search next to a studio subscription: created=%v err=%v %+v", created, err, search)
	}

	for _, invalid := range []struct {
		kind     model.SubscriptionKind
		targetID string
	}{
		{model.SubscriptionKindPerformer, "studio:vixen"},
		{model.SubscriptionKindSearch, "studio:vixen"},
		{model.SubscriptionKindStudio, "studio:missing"},
		{model.SubscriptionKindStudio, "scene:credited"},
	} {
		if _, _, err := repo.SubscribeEntity(ctx, invalid.kind, invalid.targetID, 0); err == nil {
			t.Fatalf("subscribe %s %s must fail", invalid.kind, invalid.targetID)
		}
	}

	all, _ := repo.SearchSubscriptions(ctx)
	if len(all) != 2 {
		t.Fatalf("subscriptions: %d", len(all))
	}
}

func TestEntitySubscriptionTargetOutlivesRecord(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	seedFollowable(t, database)
	subscription, _, err := repo.SubscribeEntity(ctx, model.SubscriptionKindTag, "tag:milf", 0)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	database.Exec(ctx, "DELETE FROM tag WHERE id = 'tag:milf'", nil)
	subscription, _ = repo.SearchSubscription(ctx, subscription.ID)
	if subscription.Target == nil || subscription.Target.Name != "milf" || subscription.Target.ImageURL != nil {
		t.Fatalf("deleted target keeps its stored name: %+v", subscription.Target)
	}
	if _, err := repo.EntityOrigin(ctx, "tag:milf"); !errors.Is(err, ErrSubscriptionTargetNotFound) {
		t.Fatalf("origin of a deleted target: %v", err)
	}
}

func TestLinkedSceneIDs(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	seedFollowable(t, database)

	for _, entityID := range []string{"studio:vixen", "performer:riley", "tag:milf"} {
		linked, err := repo.LinkedSceneIDs(ctx, entityID)
		if err != nil || len(linked) != 1 || linked[0] != "scene:credited" {
			t.Fatalf("%s linked scenes: %v %v", entityID, err, linked)
		}
		kept, err := repo.KeepLinkedOrUncredited(ctx, entityID,
			[]string{"scene:credited", "scene:foreign", "scene:stub", "gallery:x"})
		sort.Strings(kept)
		if err != nil || len(kept) != 2 || kept[0] != "scene:credited" || kept[1] != "scene:stub" {
			t.Fatalf("%s kept: %v %v", entityID, err, kept)
		}
	}
	if _, err := repo.LinkedSceneIDs(ctx, "scene:credited"); err == nil {
		t.Fatal("scenes cannot be followed")
	}
}

func TestEntityOrigin(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	seedFollowable(t, database)
	insertRow(t, database, "observation", "observation:studio", map[string]any{"target": "studio:vixen", "plugin": "eporner"})
	insertRow(t, database, "observation", "observation:scene", map[string]any{"target": "scene:credited", "plugin": "xhamster"})
	insertRow(t, database, "observation", "observation:foreign", map[string]any{"target": "scene:foreign", "plugin": "tnaflix"})

	origin, err := repo.EntityOrigin(ctx, "studio:vixen")
	if err != nil {
		t.Fatalf("origin: %v", err)
	}
	if origin.Name != "Vixen" || origin.SourceURL != "https://www.eporner.com/channel/vixen/" {
		t.Fatalf("origin identity: %+v", origin)
	}
	if len(origin.Observers) != 1 || origin.Observers[0] != "eporner" {
		t.Fatalf("observers: %v", origin.Observers)
	}
	if len(origin.SceneSources) != 1 || origin.SceneSources[0] != "xhamster" {
		t.Fatalf("scene sources: %v", origin.SceneSources)
	}
}
