package recommend

import (
	"context"
	"testing"
	"time"
)

func TestWatchedPenaltyFadesForSessionEnders(t *testing.T) {
	day := 24 * time.Hour
	endedAt := func(age time.Duration) time.Time { return testNow.Add(-age) }
	cases := []struct {
		name  string
		state watchState
		ended time.Time
		want  float64
	}{
		{"not a session ender, completed", watchCompleted, time.Time{}, penaltyWatchedCompleted},
		{"not a session ender, partial", watchPartial, time.Time{}, penaltyWatchedPartial},
		{"unwatched", watchNone, endedAt(30 * day), 0},
		{"session ender, fresh", watchCompleted, endedAt(day), penaltyWatchedCompleted},
		{"session ender, at the end of the free days", watchCompleted, endedAt(rewatchFreeAfter), penaltyWatchedCompleted},
		{"session ender, halfway", watchCompleted, endedAt(rewatchFreeAfter + rewatchRampDuration/2), penaltyWatchedCompleted/2 - rewatchBonus/2},
		{"session ender, long ago", watchCompleted, endedAt(60 * day), -rewatchBonus},
		{"partial session ender, long ago", watchPartial, endedAt(60 * day), -rewatchBonus},
	}
	for _, testCase := range cases {
		got := watchedPenalty(testCase.state, testCase.ended, testNow)
		assertClose(t, testCase.name, got, testCase.want)
	}
}

func TestEndsSessionRequiresLastLongWatch(t *testing.T) {
	long := watch{progressSeconds: 300, durationSeconds: 1200}
	if !endsSession(long, false, time.Hour) {
		t.Error("the last long watch ends the session")
	}
	if endsSession(long, true, time.Hour) {
		t.Error("a watch another video followed does not")
	}
	if endsSession(long, false, 10*time.Minute) {
		t.Error("a session still open does not")
	}
	if endsSession(watch{progressSeconds: 30, durationSeconds: 1200}, false, time.Hour) {
		t.Error("a look under a minute does not")
	}
	if endsSession(watch{progressSeconds: 8, durationSeconds: 1200}, false, time.Hour) {
		t.Error("an early abandonment does not")
	}
}

// insertSessionEnder stores a scene watched for 10 minutes ending a session age ago.
func insertSessionEnder(t *testing.T, engine *Engine, sceneID string, age time.Duration, fields map[string]any) {
	t.Helper()
	fields["duration_seconds"] = 1200
	insertScene(t, engine.database, sceneID, fields)
	insertRow(t, engine.database, "watch_history", "watch_history:"+sceneID, map[string]any{
		"media": sceneID, "progress_seconds": 600, "max_progress_seconds": 600, "duration_seconds": 1200,
		"started_at": ago(age + 20*time.Minute), "updated_at": ago(age),
	})
}

func TestFeedLiftsTheWatchedPenaltyOfOldSessionEnders(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	day := 24 * time.Hour
	insertSessionEnder(t, engine, "scene:old-favourite", 20*day, map[string]any{})
	// Just as old, but another video followed it: the penalty stays.
	insertScene(t, database, "scene:old-followed", map[string]any{"duration_seconds": 1200})
	insertRow(t, database, "watch_history", "watch_history:followed", map[string]any{
		"media": "scene:old-followed", "progress_seconds": 600, "max_progress_seconds": 600, "duration_seconds": 1200,
		"started_at": ago(25*day + 20*time.Minute), "updated_at": ago(25 * day),
	})
	insertScene(t, database, "scene:next", map[string]any{})
	insertRow(t, database, "watch_history", "watch_history:next", map[string]any{
		"media": "scene:next", "progress_seconds": 700, "max_progress_seconds": 700, "duration_seconds": 1200,
		"started_at": ago(25*day - 10*time.Minute), "updated_at": ago(25*day - 25*time.Minute),
	})
	insertSessionEnder(t, engine, "scene:fresh-favourite", time.Hour, map[string]any{})

	profile, err := engine.Profile(context.Background())
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if _, ended := profile.sessionEnds["scene:old-favourite"]; !ended {
		t.Fatalf("old favourite must be a session ender: %v", profile.sessionEnds)
	}
	if _, ended := profile.sessionEnds["scene:old-followed"]; ended {
		t.Error("a watch another video followed is not a session ender")
	}
	state := &snapshot{
		now:     testNow,
		profile: profile,
		watched: map[string]watchState{"scene:old-favourite": watchPartial, "scene:old-followed": watchPartial, "scene:fresh-favourite": watchPartial},
	}
	if got := penalties("scene:old-favourite", state); got >= 0 {
		t.Errorf("an old session ender carries a bonus, got penalty %v", got)
	}
	assertClose(t, "followed watch keeps the penalty", penalties("scene:old-followed", state), penaltyWatchedPartial)
	assertClose(t, "fresh session ender keeps the penalty", penalties("scene:fresh-favourite", state), penaltyWatchedPartial)
}

func TestRowsOfferOldSessionEndersToWatchAgain(t *testing.T) {
	ctx := context.Background()
	engine, database, repo := newTestEngine(t)
	day := 24 * time.Hour
	insertSessionEnder(t, engine, "scene:a", 10*day, map[string]any{})
	insertSessionEnder(t, engine, "scene:b", 12*day, map[string]any{})
	insertSessionEnder(t, engine, "scene:c", 14*day, map[string]any{})
	insertSessionEnder(t, engine, "scene:recent", 2*day, map[string]any{})
	insertSessionEnder(t, engine, "scene:blocked", 11*day, map[string]any{"tags": []string{"tag:blocked"}})
	// Watched, but not a favourite: another video followed it.
	insertScene(t, database, "scene:plain", map[string]any{})
	insertRow(t, database, "watch_history", "watch_history:plain", map[string]any{
		"media": "scene:plain", "progress_seconds": 600, "max_progress_seconds": 600, "duration_seconds": 1200,
		"started_at": ago(10*day + 55*time.Minute), "updated_at": ago(10*day + 35*time.Minute),
	})
	if _, err := repo.AddBlock(ctx, "tag", "tag:blocked", nil); err != nil {
		t.Fatalf("block: %v", err)
	}

	rows, err := engine.Rows(ctx, nil, 10, 10)
	if err != nil {
		t.Fatalf("rows: %v", err)
	}
	var shelf *Row
	for index := range rows {
		if rows[index].Key == ReasonRewatch {
			shelf = &rows[index]
		}
	}
	if shelf == nil {
		t.Fatalf("expected a Watch again row, got %+v", rows)
	}
	if shelf.Title != "Watch again" || shelf.Reason.Kind != ReasonRewatch {
		t.Errorf("shelf = %q / %q", shelf.Title, shelf.Reason.Kind)
	}
	got := map[string]bool{}
	for _, item := range shelf.Items {
		got[item.SceneID] = true
		if item.Source != SourceRewatch {
			t.Errorf("item source = %q", item.Source)
		}
	}
	for _, want := range []string{"scene:a", "scene:b", "scene:c"} {
		if !got[want] {
			t.Errorf("%s missing from the shelf: %v", want, got)
		}
	}
	for _, unwanted := range []string{"scene:recent", "scene:blocked", "scene:plain"} {
		if got[unwanted] {
			t.Errorf("%s must not be on the shelf", unwanted)
		}
	}
}

func TestWithRewatchRowSkipsThinShelvesAndKeepsPosition(t *testing.T) {
	rows := []Row{{Key: "one"}, {Key: "two"}, {Key: "three"}}
	items := []Item{{SceneID: "scene:a"}, {SceneID: "scene:b"}, {SceneID: "scene:c"}}
	if got := withRewatchRow(rows, items[:2], 10, 10); len(got) != 3 {
		t.Errorf("a shelf under minRowItems must be left out, got %d rows", len(got))
	}
	got := withRewatchRow(rows, items, 3, 2)
	if len(got) != 3 || got[rewatchRowPosition].Key != ReasonRewatch || len(got[rewatchRowPosition].Items) != 2 {
		t.Errorf("shelf must sit at position %d within the row and item caps, got %+v", rewatchRowPosition, got)
	}
}
