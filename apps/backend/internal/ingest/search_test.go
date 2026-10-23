package ingest

import (
	"strings"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// searchTitles runs a library search and returns the titles in order.
func searchTitles(t *testing.T, service *Service, table, query string) []string {
	t.Helper()
	records, err := service.SearchLibrary(t.Context(), table, query, 20)
	if err != nil {
		t.Fatalf("search %q: %v", query, err)
	}
	titles := make([]string, 0, len(records))
	for _, record := range records {
		title, _ := record["title"].(string)
		if title == "" {
			title, _ = record["name"].(string)
		}
		titles = append(titles, title)
	}
	return titles
}

// TestSearchLibraryScenesMatchPerformerStudioAndTags checks that the full-text
// index finds a scene by the names its title does not contain, ranks a title
// match above a tag match, and follows renames through the triggers.
func TestSearchLibraryScenesMatchPerformerStudioAndTags(t *testing.T) {
	service, database, ctx := openTestService(t)
	ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: "https://a.test/1", Title: "Lagoon Nights",
		Performers: []plugins.ScenePerformer{{Name: "Mia Rose"}}, Studio: &plugins.StudioRef{Name: "Blue Films"},
		Tags: []string{"Outdoor"},
	})
	ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: "https://a.test/2", Title: "Rose garden",
		Tags: []string{"Outdoor", "Rose"},
	})
	ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: "https://a.test/3", Title: "Unrelated",
	})

	if got := searchTitles(t, service, "scene", "rose"); len(got) != 2 || got[0] != "Rose garden" {
		t.Fatalf("rose = %v, want the title match first, then the performer match", got)
	}
	if got := searchTitles(t, service, "scene", "mia ro"); len(got) != 1 || got[0] != "Lagoon Nights" {
		t.Fatalf("performer prefix = %v", got)
	}
	if got := searchTitles(t, service, "scene", "blue films"); len(got) != 1 || got[0] != "Lagoon Nights" {
		t.Fatalf("studio = %v", got)
	}
	if got := searchTitles(t, service, "scene", "agoon"); len(got) != 1 || got[0] != "Lagoon Nights" {
		t.Fatalf("mid-word fallback = %v", got)
	}

	if _, err := database.Exec(ctx, `UPDATE performer SET name = 'Zoe Lane' WHERE name = 'Mia Rose'`, nil); err != nil {
		t.Fatalf("rename performer: %v", err)
	}
	if got := searchTitles(t, service, "scene", "zoe"); len(got) != 1 || got[0] != "Lagoon Nights" {
		t.Fatalf("after rename = %v", got)
	}
	if _, err := database.Exec(ctx, `DELETE FROM scene WHERE title = 'Lagoon Nights'`, nil); err != nil {
		t.Fatalf("delete scene: %v", err)
	}
	if got := searchTitles(t, service, "scene", "zoe"); len(got) != 0 {
		t.Fatalf("after delete = %v", got)
	}
}

// TestSearchLibraryCarriesIndexText checks that scene rows hand the callers
// the names behind the match, one per line.
func TestSearchLibraryCarriesIndexText(t *testing.T) {
	service, _, ctx := openTestService(t)
	ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: "https://a.test/1", Title: "Lagoon",
		Performers: []plugins.ScenePerformer{{Name: "Mia Rose"}, {Name: "Kim Lee"}},
	})
	records, err := service.SearchLibrary(ctx, "scene", "lagoon", 5)
	if err != nil || len(records) != 1 {
		t.Fatalf("search: %v %v", records, err)
	}
	names, _ := records[0]["search_performers"].(string)
	if lines := strings.Split(names, "\n"); len(lines) != 2 {
		t.Fatalf("search_performers = %q", names)
	}
}

// TestSearchLibraryNamesRankExactFirst checks that performers answering the
// query exactly come before prefixes (shorter first), word prefixes and inner matches.
func TestSearchLibraryNamesRankExactFirst(t *testing.T) {
	service, database, ctx := openTestService(t)
	for _, name := range []string{"Anna Ray", "Joanna", "Ann", "Hannah Ann Lee", "Annabel"} {
		if _, err := database.Insert(ctx, "performer", db.Vars{"name": name}); err != nil {
			t.Fatalf("insert %s: %v", name, err)
		}
	}
	got := searchTitles(t, service, "performer", "ann")
	want := []string{"Ann", "Annabel", "Anna Ray", "Hannah Ann Lee", "Joanna"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("order = %v, want %v", got, want)
	}
}

// TestFullTextQuery checks the MATCH expression quotes words and prefixes the last.
func TestFullTextQuery(t *testing.T) {
	if got := FullTextQuery(`Mia "Rose" OR -x*`); got != `"mia" "rose" "or" "x"*` {
		t.Fatalf("query = %s", got)
	}
	if FullTextQuery("  !! ") != "" {
		t.Fatal("no words must give an empty query")
	}
}

// TestSearchSceneCandidatesFindMisspelledWords checks the trigram index offers
// a scene for a misspelled word, follows renames through the triggers, and that
// the exact index count stays zero for the misspelling.
func TestSearchSceneCandidatesFindMisspelledWords(t *testing.T) {
	service, database, ctx := openTestService(t)
	ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: "https://a.test/1", Title: "Hotel night",
		Performers: []plugins.ScenePerformer{{Name: "Mia Khalifa"}},
	})
	ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: "https://a.test/2", Title: "Garden party",
	})

	if got := searchTitles(t, service, "scene", "mia khalfia"); len(got) != 0 {
		t.Fatalf("exact index found the misspelling: %v", got)
	}
	count, err := service.CountSceneMatches(ctx, "mia khalfia", 100)
	if err != nil || count != 0 {
		t.Fatalf("exact count = %d, %v", count, err)
	}
	if count, err := service.CountSceneMatches(ctx, "mia khalifa", 100); err != nil || count != 1 {
		t.Fatalf("exact count of the right spelling = %d, %v", count, err)
	}

	candidates, err := service.SearchSceneCandidates(ctx, "mia khalfia")
	if err != nil || len(candidates) == 0 || candidates[0]["title"] != "Hotel night" {
		t.Fatalf("candidates = %v, %v", candidates, err)
	}
	if performers, _ := candidates[0]["search_performers"].(string); performers != "Mia Khalifa" {
		t.Fatalf("candidate performers = %q", performers)
	}

	if _, err := database.Exec(ctx, `UPDATE performer SET name = 'Zoe Lane' WHERE name = 'Mia Khalifa'`, nil); err != nil {
		t.Fatalf("rename: %v", err)
	}
	candidates, err = service.SearchSceneCandidates(ctx, "zoe lanne")
	if err != nil || len(candidates) == 0 || candidates[0]["title"] != "Hotel night" {
		t.Fatalf("after rename = %v, %v", candidates, err)
	}
	if _, err := database.Exec(ctx, `DELETE FROM scene WHERE title = 'Hotel night'`, nil); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if after, _ := service.SearchSceneCandidates(ctx, "zoe lanne"); len(after) != 0 {
		t.Fatalf("deleted scene still offered: %v", after)
	}
}

// TestTrigramQueryUsesPiecesOfLongWords checks the expression holds every
// three-letter piece and skips words that are too short.
func TestTrigramQueryUsesPiecesOfLongWords(t *testing.T) {
	if got := TrigramQuery("Mia Köln"); got != `"mia" OR "kol" OR "oln"` {
		t.Fatalf("query = %s", got)
	}
	if got := TrigramQuery("to be"); got != "" {
		t.Fatalf("short words = %q", got)
	}
}

// TestNamedRecordsListsAliases checks the entity listing carries names and
// aliases and refuses other tables.
func TestNamedRecordsListsAliases(t *testing.T) {
	service, database, ctx := openTestService(t)
	if _, err := database.Insert(ctx, "tag", db.Vars{"name": "Blonde", "aliases": []string{"Blond"}}); err != nil {
		t.Fatalf("insert: %v", err)
	}
	records, err := service.NamedRecords(ctx, "tag")
	if err != nil || len(records) != 1 || records[0]["name"] != "Blonde" {
		t.Fatalf("records = %v, %v", records, err)
	}
	if aliases, _ := records[0]["aliases"].([]any); len(aliases) != 1 || aliases[0] != "Blond" {
		t.Fatalf("aliases = %v", records[0]["aliases"])
	}
	if _, err := service.NamedRecords(ctx, "scene"); err == nil {
		t.Fatal("scene is not a named table")
	}
}
