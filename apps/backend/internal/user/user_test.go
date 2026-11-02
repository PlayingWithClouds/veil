package user

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// openTestService opens the service on a fresh database holding one video,
// and returns the video's id.
func openTestService(t *testing.T) (*Service, string) {
	t.Helper()
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	videoID, err := database.Insert(ctx, "Video", db.Vars{"title": "Pool day", "status": "available", "performers": []string{}, "tags": []string{}})
	if err != nil {
		t.Fatal(err)
	}
	return New(database, ""), videoID
}

// TestStateEdits rates, favourites and blocks a video, and rejects a
// rating out of range and a record that doesn't exist.
func TestStateEdits(t *testing.T) {
	ctx := context.Background()
	service, videoID := openTestService(t)
	rating := 8.5
	if _, err := service.Rate(ctx, schema.RecordKindVideo, videoID, &rating); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetFavorite(ctx, schema.RecordKindVideo, videoID, true); err != nil {
		t.Fatal(err)
	}
	state, err := service.SetBlocked(ctx, schema.RecordKindVideo, videoID, true)
	if err != nil {
		t.Fatal(err)
	}
	if state.Rating == nil || *state.Rating != 8.5 || !state.Favorite || !state.Blocked {
		t.Errorf("state = %+v", state)
	}
	tooHigh := 11.0
	if _, err := service.Rate(ctx, schema.RecordKindVideo, videoID, &tooHigh); err == nil {
		t.Error("rating 11 was accepted")
	}
	if _, err := service.SetFavorite(ctx, schema.RecordKindVideo, "missing", true); err == nil {
		t.Error("state on a missing video was accepted")
	}
	favorite := true
	favorites, err := service.States(ctx, StateFilter{Favorite: &favorite}, 10, 0)
	if err != nil || len(favorites) != 1 {
		t.Errorf("favorites = %+v, err %v", favorites, err)
	}
}

// TestEventsUpdateState logs views and progress, then a finish, and checks
// the view count, resume point and event counts.
func TestEventsUpdateState(t *testing.T) {
	ctx := context.Background()
	service, videoID := openTestService(t)
	kind := schema.RecordKindVideo
	subject := func(eventKind schema.EventKind, data map[string]any) NewEvent {
		return NewEvent{Kind: eventKind, SubjectKind: &kind, SubjectID: &videoID, Data: data}
	}
	for _, event := range []NewEvent{
		subject(schema.EventKindView, nil),
		subject(schema.EventKindView, nil),
		subject(schema.EventKindProgress, map[string]any{"position": 42.0}),
		subject(schema.EventKindO, nil),
		{Kind: schema.EventKindSearch, Data: map[string]any{"text": "pool"}},
		{Kind: schema.EventKindSearch, Data: map[string]any{"text": "pool"}},
	} {
		if _, err := service.Record(ctx, event); err != nil {
			t.Fatal(err)
		}
	}
	state, _ := service.State(ctx, kind, videoID)
	if state.ViewCount != 2 || state.LastViewedAt == nil || state.Progress == nil || *state.Progress != 42 {
		t.Errorf("state = %+v", state)
	}
	oCount, err := service.CountEvents(ctx, EventQuery{Kinds: []schema.EventKind{schema.EventKindO}, SubjectID: &videoID})
	if err != nil || oCount != 1 {
		t.Errorf("o-count = %d, err %v", oCount, err)
	}
	searches, err := service.RecentSearches(ctx, 10)
	if err != nil || len(searches) != 1 || searches[0] != "pool" {
		t.Errorf("recent searches = %v, err %v", searches, err)
	}
	if _, err := service.Record(ctx, subject(schema.EventKindFinish, nil)); err != nil {
		t.Fatal(err)
	}
	state, _ = service.State(ctx, kind, videoID)
	if state.Progress != nil {
		t.Errorf("progress = %v after finishing", *state.Progress)
	}
	if _, err := service.Record(ctx, subject(schema.EventKindProgress, nil)); err == nil {
		t.Error("progress without a position was accepted")
	}
}
