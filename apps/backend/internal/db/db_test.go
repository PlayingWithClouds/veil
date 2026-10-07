package db

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	database, err := Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

func TestOpenIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	for attempt := range 2 {
		database, err := Open(context.Background(), path)
		if err != nil {
			t.Fatalf("open attempt %d: %v", attempt, err)
		}
		database.Close()
	}
}

func TestInsertGetDecodesColumnTypes(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	tagID := NewRecordID("tag")
	sceneID, err := database.Insert(ctx, "scene", map[string]any{
		"source_url": "https://example.com/1",
		"title":      "First",
		"tags":       []RecordID{tagID},
		"organized":  true,
		"not_a_col":  "dropped",
	})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	row, err := database.Get(ctx, sceneID)
	if err != nil || row == nil {
		t.Fatalf("get: %v %v", row, err)
	}
	if row["id"] != sceneID.String() {
		t.Errorf("id = %v, want %s", row["id"], sceneID)
	}
	if row["organized"] != true {
		t.Errorf("organized = %#v, want true", row["organized"])
	}
	tags, isList := row["tags"].([]any)
	if !isList || len(tags) != 1 || tags[0] != tagID.String() {
		t.Errorf("tags = %#v", row["tags"])
	}
	if _, present := row["details"]; present {
		t.Errorf("NULL column should be absent, got %#v", row["details"])
	}
}

func TestNamedParametersAndJSONMembership(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	tagID := NewRecordID("tag")
	for index, tags := range [][]RecordID{{tagID}, {}} {
		if _, err := database.Insert(ctx, "scene", map[string]any{
			"source_url": "https://example.com/" + string(rune('a'+index)),
			"tags":       tags,
		}); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
	count, err := database.Int(ctx,
		`SELECT count(*) FROM scene WHERE EXISTS (SELECT 1 FROM json_each(scene.tags) WHERE value = $tag)`,
		Vars{"tag": tagID, "unused": 1})
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1", count)
	}
	inList, err := database.Int(ctx,
		`SELECT count(*) FROM scene WHERE source_url IN (SELECT value FROM json_each($urls))`,
		Vars{"urls": []string{"https://example.com/a", "https://example.com/b"}})
	if err != nil || inList != 2 {
		t.Errorf("IN json_each = %d, %v; want 2", inList, err)
	}
}

type testJob struct {
	ID        *RecordID      `json:"id,omitempty"`
	Kind      string         `json:"kind"`
	Target    *RecordID      `json:"target,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
	Attempts  int            `json:"attempts"`
	RunAt     *time.Time     `json:"run_at,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

func TestStructRoundTrip(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	target := NewRecordID("scene")
	runAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	jobID, err := database.Insert(ctx, "job", testJob{
		Kind:      "scrape",
		Target:    &target,
		Payload:   map[string]any{"url": "https://example.com", "depth": 2},
		Attempts:  1,
		RunAt:     &runAt,
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	job, err := QueryOneAs[testJob](ctx, database, "SELECT * FROM job WHERE id = $id", Vars{"id": jobID})
	if err != nil || job == nil {
		t.Fatalf("query: %v %v", job, err)
	}
	if job.ID == nil || *job.ID != jobID {
		t.Errorf("id = %v, want %v", job.ID, jobID)
	}
	if job.Target == nil || *job.Target != target {
		t.Errorf("target = %v, want %v", job.Target, target)
	}
	if job.Payload["url"] != "https://example.com" || job.Payload["depth"] != float64(2) {
		t.Errorf("payload = %#v", job.Payload)
	}
	if job.RunAt == nil || !job.RunAt.Equal(runAt) {
		t.Errorf("run_at = %v, want %v", job.RunAt, runAt)
	}
}

func TestMergeBumpsUpdatedAt(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	tagID, err := database.Insert(ctx, "tag", map[string]any{"name": "a", "updated_at": "2000-01-01T00:00:00.000Z"})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	if err := database.Merge(ctx, tagID, map[string]any{"description": "changed"}); err != nil {
		t.Fatalf("merge: %v", err)
	}
	row, _ := database.Get(ctx, tagID)
	if row["description"] != "changed" {
		t.Errorf("description = %#v", row["description"])
	}
	if row["updated_at"] == "2000-01-01T00:00:00.000Z" {
		t.Errorf("updated_at was not bumped")
	}
}

func TestTxRollsBack(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	_ = database.Tx(ctx, func(tx *DB) error {
		if _, err := tx.Insert(ctx, "tag", map[string]any{"name": "rolled back"}); err != nil {
			return err
		}
		return context.Canceled
	})
	count, _ := database.Int(ctx, "SELECT count(*) FROM tag", nil)
	if count != 0 {
		t.Errorf("count = %d, want 0 after rollback", count)
	}
}
