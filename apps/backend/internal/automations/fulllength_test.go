package automations

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// fakeKeeper records what the full-length hunt keeps.
type fakeKeeper struct {
	favourites []string
	added      map[string][]string
}

// SetFavorite records a favourite.
func (k *fakeKeeper) SetFavorite(_ context.Context, _ schema.RecordKind, id string, favorite bool) (*schema.UserState, error) {
	if favorite {
		k.favourites = append(k.favourites, id)
	}
	return &schema.UserState{}, nil
}

// AddItems records collection additions.
func (k *fakeKeeper) AddItems(_ context.Context, collectionID string, items []schema.RecordRef, _ *int) error {
	for _, item := range items {
		k.added[collectionID] = append(k.added[collectionID], item.Id)
	}
	return nil
}

// TestFullLengthHunt searches for a favourite trailer's full scene without
// the word "trailer", favourites the full scene, adds it to the trailer's
// collection, notifies, and doesn't search for the trailer again.
func TestFullLengthHunt(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	keeper := &fakeKeeper{added: map[string][]string{}}
	f.service.RegisterAction(ActionFullLength, f.service.FullLengthAction(keeper))
	clipID, err := f.database.Insert(ctx, "Video", db.Vars{"title": "Sunset Pool Party trailer", "status": schema.MediaStatusAvailable, "performers": []string{}, "tags": []string{}, "durationSeconds": 90})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.database.Insert(ctx, "UserState", db.Vars{"subjectKind": "video", "subjectId": clipID, "favorite": true, "blocked": false, "viewCount": 0, "fields": map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	collectionID, err := f.database.Insert(ctx, "Collection", db.Vars{"name": "Mine", "kind": schema.CollectionKindUser})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.database.Insert(ctx, "CollectionItem", db.Vars{"collection": collectionID, "subjectKind": "video", "subjectId": clipID, "position": 0}); err != nil {
		t.Fatal(err)
	}
	f.searcher.items = []plugins.DiscoveredItem{
		{MediaType: plugins.MediaTypeScene, ExternalID: "short", SourceURL: "https://site.test/short", Title: "Sunset Pool Party preview", DurationSeconds: 120},
		{MediaType: plugins.MediaTypeScene, ExternalID: "full", SourceURL: "https://site.test/full", Title: "Sunset Pool Party", DurationSeconds: 2400},
	}
	if err := f.service.EnsureBuiltIns(ctx); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		run, err := f.service.Execute(ctx, Request{AutomationID: FullLengthHuntID})
		if err != nil || run.Status != schema.RunStatusSucceeded {
			t.Fatalf("run = %+v, err %v", run, err)
		}
	}
	if len(f.searcher.queries) != 1 || f.searcher.queries[0] != "Sunset Pool Party" {
		t.Fatalf("queries = %v, want one search without the word trailer", f.searcher.queries)
	}
	if len(keeper.favourites) != 1 || len(keeper.added[collectionID]) != 1 || keeper.favourites[0] != keeper.added[collectionID][0] {
		t.Errorf("favourites = %v, added = %v", keeper.favourites, keeper.added)
	}
	full, _ := f.library.Video(ctx, keeper.favourites[0])
	if full == nil || full.Title != "Sunset Pool Party" {
		t.Errorf("kept %+v, want the full scene", full)
	}
	notifications, _ := f.notifier.Notifications(ctx, false, 10, 0)
	if len(notifications) != 1 || notifications[0].Kind != "fullLengthFound" {
		t.Errorf("notifications = %+v", notifications)
	}
}
