package limits

import (
	"context"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// openTestService opens a limit service on a fresh database.
func openTestService(t *testing.T) (*Service, *db.DB) {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return New(database), database
}

// logEvent inserts an event of kind in session at the given time.
func logEvent(t *testing.T, database *db.DB, kind schema.EventKind, session string, at time.Time) {
	t.Helper()
	row := db.Vars{"at": db.FormatTime(at), "kind": kind}
	if session != "" {
		row["sessionId"] = session
	}
	if _, err := database.Insert(context.Background(), "Event", row); err != nil {
		t.Fatal(err)
	}
}

// TestUsage counts heartbeats per session, skipping pauses over a minute,
// and counts the sessions that viewed something.
func TestUsage(t *testing.T) {
	ctx := context.Background()
	service, database := openTestService(t)
	now := time.Date(2026, 11, 4, 20, 0, 0, 0, time.UTC)
	start := now.Add(-time.Hour)
	logEvent(t, database, schema.EventKindView, "a", start)
	// Session a: 0, 30, 60 s, then a 5 min pause, then 90 s more.
	for _, offset := range []time.Duration{0, 30, 60, 360, 390, 420, 450} {
		logEvent(t, database, schema.EventKindProgress, "a", start.Add(offset*time.Second))
	}
	logEvent(t, database, schema.EventKindView, "b", start)
	logEvent(t, database, schema.EventKindHeatSpan, "b", start.Add(10*time.Second))
	logEvent(t, database, schema.EventKindHeatSpan, "b", start.Add(40*time.Second))
	logEvent(t, database, schema.EventKindProgress, "", start)
	// Yesterday, outside the day window.
	logEvent(t, database, schema.EventKindView, "c", now.AddDate(0, 0, -1))

	usage, err := service.Usage(ctx, schema.LimitWindowDay, now)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(usage.WatchMinutes-3) > 0.001 {
		t.Errorf("watch minutes = %v, want 3 (60 s + 90 s + 30 s)", usage.WatchMinutes)
	}
	if usage.Sessions != 2 {
		t.Errorf("sessions = %d, want 2", usage.Sessions)
	}
	week, _ := service.Usage(ctx, schema.LimitWindowWeek, now)
	if week.Sessions != 3 {
		t.Errorf("week sessions = %d, want 3 (the week started on Monday)", week.Sessions)
	}
}

// TestStatus reports which enabled limits are reached.
func TestStatus(t *testing.T) {
	ctx := context.Background()
	service, database := openTestService(t)
	now := time.Date(2026, 11, 4, 20, 0, 0, 0, time.UTC)
	logEvent(t, database, schema.EventKindView, "a", now.Add(-time.Minute))
	logEvent(t, database, schema.EventKindView, "b", now.Add(-time.Minute))
	sessions, err := service.Create(ctx, NewLimit{Kind: schema.LimitKindSessions, Window: schema.LimitWindowDay, Amount: 2, Action: schema.LimitActionLock, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(ctx, NewLimit{Kind: schema.LimitKindWatchTime, Window: schema.LimitWindowWeek, Amount: 60, Action: schema.LimitActionRemind, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Create(ctx, NewLimit{Kind: schema.LimitKindSessions, Window: schema.LimitWindowDay, Amount: 0}); err == nil {
		t.Error("a limit of 0 was accepted")
	}
	statuses, err := service.StatusAt(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(statuses) != 2 || !statuses[0].Reached || statuses[0].Used != 2 || statuses[1].Reached {
		t.Errorf("statuses = %+v, want the session limit reached only", statuses)
	}
	disabled := false
	if _, err := service.Update(ctx, sessions.Id, LimitEdit{Enabled: &disabled}); err != nil {
		t.Fatal(err)
	}
	statuses, _ = service.StatusAt(ctx, now)
	if len(statuses) != 1 {
		t.Errorf("disabled limit still measured: %+v", statuses)
	}
}
