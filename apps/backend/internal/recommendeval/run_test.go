package recommendeval

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// baseTime is day 0 of the synthetic history.
var baseTime = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

// dayAt renders the stored timestamp of a moment on a day of the history.
func dayAt(day int, hour, minute int) string {
	return db.FormatTime(baseTime.Add(time.Duration(day)*24*time.Hour + time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute))
}

// insertRow inserts a record with an explicit id and fails the test on error.
func insertRow(t *testing.T, database *db.DB, table, id string, fields map[string]any) {
	t.Helper()
	fields["id"] = id
	if _, err := database.Insert(context.Background(), table, fields); err != nil {
		t.Fatalf("insert %s: %v", id, err)
	}
}

// insertScene stores a scene created on a day of the history.
func insertScene(t *testing.T, database *db.DB, id string, createdDay int, fields map[string]any) {
	t.Helper()
	fields["title"] = id
	fields["source_url"] = "https://example.test/" + id
	fields["created_at"] = dayAt(createdDay, 0, 0)
	insertRow(t, database, "scene", id, fields)
}

// insertWatch stores a watch of a scene from start to end on a day.
func insertWatch(t *testing.T, database *db.DB, sceneID string, day, startHour, startMinute, endHour, endMinute, progress int) {
	t.Helper()
	insertRow(t, database, "watch_history", "watch_history:"+sceneID[len("scene:"):], map[string]any{
		"media": sceneID, "started_at": dayAt(day, startHour, startMinute), "updated_at": dayAt(day, endHour, endMinute),
		"progress_seconds": progress, "max_progress_seconds": progress, "played_seconds": progress, "duration_seconds": 600,
	})
}

// buildHistory writes a database whose user watched scene:seed on day 0 and,
// on day 3, scene:target (a site-related scene of the seed, then left alone),
// scene:interrupted (followed 20 minutes later by scene:after); on day 4 they
// opened scene:late, which did not exist before day 4; scene:last, watched on day 6, is the
// latest activity, so whether it ended its session is not yet decided.
func buildHistory(t *testing.T, path string) {
	t.Helper()
	database, err := db.Open(context.Background(), path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()
	insertScene(t, database, "scene:seed", 0, map[string]any{"tags": []string{"tag:a"}})
	insertScene(t, database, "scene:target", 0, map[string]any{"tags": []string{"tag:a"}})
	insertScene(t, database, "scene:interrupted", 0, map[string]any{})
	insertScene(t, database, "scene:after", 0, map[string]any{})
	insertScene(t, database, "scene:last", 0, map[string]any{})
	insertScene(t, database, "scene:late", 4, map[string]any{"tags": []string{"tag:a"}, "detail_fetched_at": dayAt(4, 9, 0)})
	for filler := 0; filler < 60; filler++ {
		insertScene(t, database, fmt.Sprintf("scene:filler%d", filler), 0, map[string]any{})
	}
	insertRow(t, database, "scene_related", "scene_related:1", map[string]any{
		"scene": "scene:seed", "related": "scene:target", "source": "site", "rank": 0, "created_at": dayAt(0, 0, 0)})
	insertWatch(t, database, "scene:seed", 0, 10, 0, 10, 10, 600)
	insertWatch(t, database, "scene:target", 3, 9, 0, 9, 5, 300)
	insertWatch(t, database, "scene:interrupted", 3, 12, 0, 12, 20, 100)
	insertWatch(t, database, "scene:after", 3, 12, 40, 12, 50, 300)
	insertWatch(t, database, "scene:late", 4, 9, 0, 9, 10, 300)
	insertWatch(t, database, "scene:last", 6, 10, 0, 10, 5, 300)
	insertRow(t, database, "recommendation_impression", "recommendation_impression:1", map[string]any{
		"media": "scene:target", "kind": "click", "source": "related:site", "surface": "home", "shown_at": dayAt(3, 8, 55)})
	insertRow(t, database, "o_event", "o_event:1", map[string]any{"media": "scene:target", "created_at": dayAt(5, 0, 0)})
}

// fileDigest hashes a file's content.
func fileDigest(t *testing.T, path string) [sha256.Size]byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return sha256.Sum256(content)
}

// kindOf finds a variant's report of one target kind.
func kindOf(t *testing.T, report *Report, variant, kind string) KindReport {
	t.Helper()
	for _, candidate := range report.Variants {
		if candidate.Name != variant {
			continue
		}
		for _, entry := range candidate.Kinds {
			if entry.Kind == kind {
				return entry
			}
		}
	}
	t.Fatalf("no %s/%s in report", variant, kind)
	return KindReport{}
}

func TestRunScoresRankingsAgainstTheFollowingWindow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "veil.db")
	buildHistory(t, path)
	before := fileDigest(t, path)

	report, err := Run(context.Background(), Config{
		DatabasePath: path, Window: 7 * 24 * time.Hour, Step: 24 * time.Hour,
		From: baseTime.Add(24 * time.Hour), To: baseTime.Add(24 * time.Hour),
		Variants: []string{VariantDefault, VariantNoLearning, VariantNewest, VariantRandom}, Repeats: 1, Seed: 7,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if fileDigest(t, path) != before {
		t.Error("the evaluated database was modified")
	}
	if len(report.Cutoffs) != 1 {
		t.Fatalf("cutoffs = %d, want 1", len(report.Cutoffs))
	}

	sessionEnd := kindOf(t, report, VariantDefault, TargetSessionEnd)
	if sessionEnd.Reachable != 2 || sessionEnd.Unreachable != 1 {
		t.Errorf("session-ending watches: reachable %d, unreachable %d, want 2 and 1 (target and after; late is new)", sessionEnd.Reachable, sessionEnd.Unreachable)
	}
	watched := kindOf(t, report, VariantDefault, TargetWatch)
	if watched.Reachable != 4 || watched.Unreachable != 1 {
		t.Errorf("watches: reachable %d, unreachable %d, want 4 and 1", watched.Reachable, watched.Unreachable)
	}
	clicked := kindOf(t, report, VariantDefault, TargetClick)
	if clicked.Reachable != 1 || clicked.Unreachable != 1 {
		t.Errorf("clicks: reachable %d, unreachable %d, want 1 and 1", clicked.Reachable, clicked.Unreachable)
	}
	if clicked.HitRateAt24 != 1 || clicked.MRR != 1 {
		t.Errorf("the site-related scene of the seed must lead the default ranking: hit@24 %.2f, MRR %.2f", clicked.HitRateAt24, clicked.MRR)
	}
	if random := kindOf(t, report, VariantRandom, TargetClick); random.MRR >= 1 {
		t.Errorf("a random ranking found the click first (MRR %.2f)", random.MRR)
	}
	if sharedSource := report.Variants[0].SourceShare["related:site"]; sharedSource <= 0 {
		t.Errorf("source shares missing the site-related source: %v", report.Variants[0].SourceShare)
	}
}

func TestRunRejectsHistoryWithoutActivity(t *testing.T) {
	path := filepath.Join(t.TempDir(), "veil.db")
	database, err := db.Open(context.Background(), path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	database.Close()
	_, err = Run(context.Background(), Config{DatabasePath: path, Window: time.Hour, Step: time.Hour, Variants: []string{VariantDefault}})
	if err == nil {
		t.Fatal("an empty history must be an error")
	}
}

func TestPruneRebuildsTheStateAtTheCutoff(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "veil.db")
	buildHistory(t, path)
	database, err := db.Open(ctx, path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()
	// scene:target was first visited on day 3 at 13:00, after the cutoff, and had its tags from that visit.
	if _, err := database.Exec(ctx, `UPDATE scene SET detail_fetched_at = $at, tags = '["tag:a"]' WHERE id = 'scene:target'`,
		db.Vars{"at": dayAt(3, 13, 0)}); err != nil {
		t.Fatalf("visit: %v", err)
	}

	// Day 3, 12:10: the interrupted watch is ten minutes in.
	if err := pruneTo(ctx, database, baseTime.Add(3*24*time.Hour+12*time.Hour+10*time.Minute)); err != nil {
		t.Fatalf("prune: %v", err)
	}
	counts := map[string]int{
		`SELECT count(*) FROM scene WHERE id = 'scene:late'`:                                   0,
		`SELECT count(*) FROM o_event`:                                                         0,
		`SELECT count(*) FROM watch_history WHERE media = 'scene:after'`:                       0,
		`SELECT count(*) FROM watch_history WHERE media = 'scene:target'`:                      1,
		`SELECT count(*) FROM scene WHERE id = 'scene:target' AND detail_fetched_at IS NULL`:   1,
		`SELECT count(*) FROM scene WHERE id = 'scene:target' AND json_array_length(tags) = 0`: 1,
	}
	for query, want := range counts {
		got, err := database.Int(ctx, query, nil)
		if err != nil || got != want {
			t.Errorf("%s = %d (%v), want %d", query, got, err, want)
		}
	}
	for _, column := range []string{"max_progress_seconds", "played_seconds"} {
		value, err := database.Int(ctx, `SELECT `+column+` FROM watch_history WHERE media = 'scene:interrupted'`, nil)
		if err != nil || value != 100 {
			t.Errorf("interrupted watch %s = %d (%v), want 100 (it had played 100 s of the 600 the wall clock allows by then)", column, value, err)
		}
	}
}
