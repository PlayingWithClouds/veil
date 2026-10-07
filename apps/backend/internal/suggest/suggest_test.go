package suggest

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/media"
	"github.com/playingwithclouds/veil/internal/recommend"
)

// testNow is the fixed clock the tests start at.
var testNow = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

// fixedTaste serves a fixed taste profile.
type fixedTaste struct {
	profile *recommend.Profile
}

// Profile returns the fixed profile.
func (taste *fixedTaste) Profile(context.Context) (*recommend.Profile, error) {
	return taste.profile, nil
}

// newTestService opens a fresh database and returns a service on a fixed
// clock with an empty taste profile the test may fill.
func newTestService(t testing.TB) (*Service, *db.DB, *media.Repository, *recommend.Profile) {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	repo := media.NewRepository(database)
	profile := &recommend.Profile{
		Tags:       map[string]float64{},
		Performers: map[string]float64{},
		Studios:    map[string]float64{},
	}
	service := New(database, &fixedTaste{profile: profile}, repo)
	service.now = func() time.Time { return testNow }
	return service, database, repo, profile
}

// insertRow inserts fields with an explicit id and fails the test on error.
func insertRow(t testing.TB, database *db.DB, table, id string, fields map[string]any) {
	t.Helper()
	fields["id"] = id
	if _, err := database.Insert(context.Background(), table, fields); err != nil {
		t.Fatalf("insert %s: %v", id, err)
	}
}

// insertScenes stores count scenes with the title and links.
func insertScenes(t testing.TB, database *db.DB, prefix string, count int, fields map[string]any) {
	t.Helper()
	for number := range count {
		row := map[string]any{"source_url": fmt.Sprintf("https://example.test/%s/%d", prefix, number)}
		for key, value := range fields {
			row[key] = value
		}
		insertRow(t, database, "scene", fmt.Sprintf("scene:%s%d", prefix, number), row)
	}
}

// suggestions runs Suggest and fails the test on error.
func suggestions(t testing.TB, service *Service, query string, limit int) []Suggestion {
	t.Helper()
	out, err := service.Suggest(context.Background(), query, limit)
	if err != nil {
		t.Fatalf("suggest %q: %v", query, err)
	}
	return out
}

// positionOf returns the index of the first suggestion with text, -1 when absent.
func positionOf(list []Suggestion, text string) int {
	for index, suggestion := range list {
		if suggestion.Text == text {
			return index
		}
	}
	return -1
}

func TestPrefixBeatsFuzzyEvenWhenLessPopular(t *testing.T) {
	service, database, _, profile := newTestService(t)
	insertRow(t, database, "performer", "performer:riley", map[string]any{"name": "Riley Reid", "image_path": "https://cdn.test/riley.jpg"})
	insertRow(t, database, "tag", "tag:rule", map[string]any{"name": "rule34"})
	insertScenes(t, database, "riley", 1, map[string]any{"performers": []string{"performer:riley"}})
	insertScenes(t, database, "rule", 400, map[string]any{"tags": []string{"tag:rule"}})
	profile.Tags["tag:rule"] = 1

	list := suggestions(t, service, "rile", 12)
	riley, rule := positionOf(list, "Riley Reid"), positionOf(list, "rule34")
	if riley != 0 {
		t.Fatalf("the prefix match must lead, got %+v", list)
	}
	if rule < 0 {
		t.Fatalf("the one-typo match must still be offered, got %+v", list)
	}
	if list[0].Kind != KindPerformer || list[0].EntityID != "performer:riley" || list[0].ImageURL != "https://cdn.test/riley.jpg" || list[0].Detail != "Performer · 1 video" {
		t.Errorf("performer suggestion fields: %+v", list[0])
	}
}

func TestWordPrefixAndAliasMatch(t *testing.T) {
	service, database, _, _ := newTestService(t)
	insertRow(t, database, "performer", "performer:riley", map[string]any{"name": "Riley Reid"})
	insertRow(t, database, "studio", "studio:bb", map[string]any{"name": "Brazzers", "aliases": []string{"BZ Network"}})

	if list := suggestions(t, service, "reid", 12); positionOf(list, "Riley Reid") != 0 {
		t.Errorf("word-prefix match missing: %+v", list)
	}
	list := suggestions(t, service, "bz net", 12)
	if positionOf(list, "Brazzers") != 0 || list[0].Kind != KindStudio {
		t.Errorf("alias match must suggest the studio by its name: %+v", list)
	}
}

func TestEntityBeatsQueryOfSameText(t *testing.T) {
	ctx := context.Background()
	service, database, _, _ := newTestService(t)
	insertRow(t, database, "tag", "tag:bigtits", map[string]any{"name": "Big Tits"})
	insertScenes(t, database, "a", 1, map[string]any{"title": "Big tits at school", "tags": []string{"tag:bigtits"}})
	insertScenes(t, database, "b", 1, map[string]any{"title": "big TITS in the office"})
	if _, err := service.RecordSearch(ctx, "big tits"); err != nil {
		t.Fatalf("record: %v", err)
	}

	list := suggestions(t, service, "big", 12)
	matches := 0
	for _, suggestion := range list {
		if matchText(suggestion.Text) == "big tits" {
			matches++
			if suggestion.Kind != KindTag || suggestion.Detail != "Tag · 1 video" {
				t.Errorf("the tag must stand in for the duplicate search: %+v", suggestion)
			}
		}
	}
	if matches != 1 {
		t.Errorf("want exactly one \"big tits\" suggestion, got %d in %+v", matches, list)
	}
}

func TestQueryCompletionsFromTitles(t *testing.T) {
	service, database, _, _ := newTestService(t)
	insertScenes(t, database, "a", 3, map[string]any{"title": "Stepsister caught in the kitchen"})
	insertScenes(t, database, "b", 1, map[string]any{"title": "Stepsister surprise"})

	list := suggestions(t, service, "stepsis", 12)
	if positionOf(list, "stepsister") != 0 || list[0].Kind != KindQuery {
		t.Fatalf("the most frequent completion must lead: %+v", list)
	}
	if positionOf(list, "stepsister caught") < 0 {
		t.Errorf("a two-word completion must be offered: %+v", list)
	}
	if positionOf(list, "stepsister surprise") >= 0 {
		t.Errorf("a phrase from a single title is noise: %+v", list)
	}
}

func TestBlockedEntitiesExcluded(t *testing.T) {
	ctx := context.Background()
	service, database, repo, profile := newTestService(t)
	insertRow(t, database, "performer", "performer:blocked", map[string]any{"name": "Blocked Person"})
	insertRow(t, database, "tag", "tag:blockbuster", map[string]any{"name": "Blockbuster"})
	insertScenes(t, database, "a", 2, map[string]any{"title": "Blocked Person stars", "performers": []string{"performer:blocked"}})
	profile.Performers["performer:blocked"] = 1
	if _, err := service.RecordSearch(ctx, "blocked person"); err != nil {
		t.Fatalf("record: %v", err)
	}
	if _, err := repo.AddBlock(ctx, "performer", "performer:blocked", nil); err != nil {
		t.Fatalf("block: %v", err)
	}

	for _, query := range []string{"block", ""} {
		for _, suggestion := range suggestions(t, service, query, 12) {
			if matchText(suggestion.Text) == "blocked person" {
				t.Errorf("query %q: blocked performer suggested as %+v", query, suggestion)
			}
		}
	}
	if list := suggestions(t, service, "block", 12); positionOf(list, "Blockbuster") < 0 {
		t.Errorf("unblocked tag missing: %+v", list)
	}
}

func TestEmptyQueryListsRecentSearchesThenTastePicks(t *testing.T) {
	ctx := context.Background()
	service, database, _, profile := newTestService(t)
	insertRow(t, database, "tag", "tag:liked", map[string]any{"name": "Liked Tag"})
	insertRow(t, database, "tag", "tag:loved", map[string]any{"name": "Loved Tag"})
	insertRow(t, database, "studio", "studio:disliked", map[string]any{"name": "Disliked Studio"})
	insertScenes(t, database, "a", 2, map[string]any{"tags": []string{"tag:liked"}})
	profile.Tags["tag:liked"] = 0.5
	profile.Tags["tag:loved"] = 1
	profile.Studios["studio:disliked"] = -1
	for index, query := range []string{"older search", "newer search", "Liked Tag"} {
		service.now = func() time.Time { return testNow.Add(time.Duration(index) * time.Minute) }
		if _, err := service.RecordSearch(ctx, query); err != nil {
			t.Fatalf("record: %v", err)
		}
	}

	list := suggestions(t, service, "  ", 12)
	want := []struct {
		kind Kind
		text string
	}{
		{KindRecent, "Liked Tag"}, {KindRecent, "newer search"}, {KindRecent, "older search"}, {KindTag, "Loved Tag"},
	}
	if len(list) != len(want) {
		t.Fatalf("want %d suggestions, got %+v", len(want), list)
	}
	for index, expected := range want {
		if list[index].Kind != expected.kind || list[index].Text != expected.text {
			t.Errorf("suggestion %d = %s %q, want %s %q", index, list[index].Kind, list[index].Text, expected.kind, expected.text)
		}
	}
	if list[3].Detail != "Tag" || list[3].EntityID != "tag:loved" {
		t.Errorf("taste pick fields: %+v", list[3])
	}
	if limited := suggestions(t, service, "", 2); len(limited) != 2 {
		t.Errorf("limit must hold, got %+v", limited)
	}
}

func TestEmptyQueryColdStartSuggestsPopularEntities(t *testing.T) {
	service, database, _, _ := newTestService(t)
	insertRow(t, database, "tag", "tag:popular", map[string]any{"name": "Popular"})
	insertRow(t, database, "tag", "tag:unused", map[string]any{"name": "Unused"})
	insertScenes(t, database, "a", 3, map[string]any{"tags": []string{"tag:popular"}})

	list := suggestions(t, service, "", 12)
	if len(list) != 1 || list[0].Text != "Popular" || list[0].Detail != "Tag · 3 videos" {
		t.Errorf("cold start must suggest entities with scenes: %+v", list)
	}
}

func TestRecordAndForgetSearch(t *testing.T) {
	ctx := context.Background()
	service, database, _, _ := newTestService(t)
	for _, query := range []string{" Foo   Bar ", "FOO bar", "foo bar"} {
		recorded, err := service.RecordSearch(ctx, query)
		if err != nil || !recorded {
			t.Fatalf("record %q: %v %v", query, recorded, err)
		}
	}
	if recorded, _ := service.RecordSearch(ctx, "   "); recorded {
		t.Error("a blank search must not be recorded")
	}
	row, err := database.QueryRow(ctx, `SELECT query, search_count FROM recent_search WHERE normalized_query = 'foo bar'`, nil)
	if err != nil || row == nil {
		t.Fatalf("history row: %v %v", row, err)
	}
	if row["query"] != "foo bar" || db.AsInt(row["search_count"]) != 3 {
		t.Errorf("history row = %v, want the latest spelling and 3 runs", row)
	}
	if list := suggestions(t, service, "fo", 12); positionOf(list, "foo bar") != 0 || list[0].Kind != KindRecent {
		t.Errorf("the past search must match its prefix: %+v", list)
	}

	forgotten, err := service.ForgetSearch(ctx, "FOO  BAR")
	if err != nil || !forgotten {
		t.Fatalf("forget: %v %v", forgotten, err)
	}
	if forgotten, _ := service.ForgetSearch(ctx, "foo bar"); forgotten {
		t.Error("forgetting twice must report false")
	}
	if list := suggestions(t, service, "", 12); len(list) != 0 {
		t.Errorf("forgotten search still suggested: %+v", list)
	}
}

func TestFuzzyPrefixDistance(t *testing.T) {
	cases := []struct {
		typed, candidate string
		maxEdits, want   int
	}{
		{"rile", "riley reid", 1, 0},
		{"rule", "rule34", 1, 0},
		{"rile", "rule34", 1, 1},
		{"brazers", "brazzers", 2, 1},
		{"blnode", "blonde", 2, 2},
		{"abcd", "wxyz", 1, 2},
	}
	for _, testCase := range cases {
		typed, candidate := []rune(testCase.typed), []rune(testCase.candidate)
		scratch := make([]int, 2*(len(typed)+testCase.maxEdits+1))
		if got := fuzzyPrefixDistance(typed, candidate, testCase.maxEdits, scratch); got != testCase.want {
			t.Errorf("fuzzyPrefixDistance(%q, %q) = %d, want %d", testCase.typed, testCase.candidate, got, testCase.want)
		}
	}
}
