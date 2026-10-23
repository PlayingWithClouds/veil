package searchrank

import (
	"strings"
	"testing"
)

// testIndex knows a performer with an alias and an accent, a studio, a
// one-word performer and a tag.
func testIndex() *EntityIndex {
	index := NewEntityIndex()
	index.AddName(Entity{Kind: KindPerformer, ID: "performer:mia", Name: "Mia Khalifa"})
	index.AddName(Entity{Kind: KindPerformer, ID: "performer:anna", Name: "Anna"})
	index.AddName(Entity{Kind: KindPerformer, ID: "performer:zoe", Name: "Zoë Lane"})
	index.AddName(Entity{Kind: KindStudio, ID: "studio:brazzers", Name: "Brazzers"})
	index.AddName(Entity{Kind: KindTag, ID: "tag:blonde", Name: "Blonde"})
	index.AddName(Entity{Kind: KindTag, ID: "tag:mia", Name: "Mia"})
	index.AddAlias("performer:mia", "Mia K")
	index.AddAlias("performer:mia", "Anna")
	return index
}

// TestFindRecognisesNamesIgnoringSpelling checks case, accents and punctuation
// do not matter, the longest name wins and spans do not overlap.
func TestFindRecognisesNamesIgnoringSpelling(t *testing.T) {
	matches := testIndex().Find("BRAZZERS zoe-lane blonde")
	if len(matches) != 3 {
		t.Fatalf("matches = %+v", matches)
	}
	if matches[0].ID != "studio:brazzers" || matches[1].ID != "performer:zoe" || matches[2].ID != "tag:blonde" {
		t.Fatalf("matches = %+v", matches)
	}
	if matches[1].Start != 1 || matches[1].End != 3 {
		t.Fatalf("span = %d..%d", matches[1].Start, matches[1].End)
	}
	if got := testIndex().Find("mia khalifa"); len(got) != 1 || got[0].ID != "performer:mia" {
		t.Fatalf("longest name should beat the tag Mia: %+v", got)
	}
}

// TestFindAliasMarksMatch checks an alias resolves to its record and says so,
// and that a record's name is not taken over by another record's alias.
func TestFindAliasMarksMatch(t *testing.T) {
	matches := testIndex().Find("mia k blonde")
	if len(matches) != 2 || matches[0].ID != "performer:mia" || !matches[0].ViaAlias {
		t.Fatalf("matches = %+v", matches)
	}
	anna := testIndex().Find("anna")
	if len(anna) != 1 || anna[0].ID != "performer:anna" || anna[0].ViaAlias {
		t.Fatalf("a name beats another record's alias: %+v", anna)
	}
}

// TestFindSingleWordPerformerOnlyAsWholeQuery checks a performer with a
// one-word name is not recognised inside a longer query.
func TestFindSingleWordPerformerOnlyAsWholeQuery(t *testing.T) {
	if got := testIndex().Find("anna"); len(got) != 1 {
		t.Fatalf("whole query = %+v", got)
	}
	if got := testIndex().Find("anna in the garden"); len(got) != 0 {
		t.Fatalf("inside a longer query = %+v", got)
	}
}

// TestFindFuzzyFindsMisspelledNames checks a misspelled name is found as a
// fuzzy match, also inside a longer query and over an exact match of part of
// it, but not for unrelated or too short text.
func TestFindFuzzyFindsMisspelledNames(t *testing.T) {
	matches := testIndex().Find("mia khalfia")
	if len(matches) != 1 || matches[0].ID != "performer:mia" || !matches[0].Fuzzy {
		t.Fatalf("matches = %+v", matches)
	}
	inside := testIndex().Find("mia khalfia garden")
	if len(inside) != 1 || !inside[0].Fuzzy || inside[0].End != 2 {
		t.Fatalf("misspelled name inside a longer query: %+v", inside)
	}
	if got := testIndex().Find("zzz qqq"); len(got) != 0 {
		t.Fatalf("unrelated = %+v", got)
	}
	if got := testIndex().Find("brzr"); len(got) != 0 {
		t.Fatalf("too short to guess = %+v", got)
	}
}

// TestRewriteSpellsOutAliasAndTypo checks only aliased or misspelled names are
// replaced, the rest of the query stays, and nothing is returned otherwise.
func TestRewriteSpellsOutAliasAndTypo(t *testing.T) {
	index := testIndex()
	if got := Rewrite("hotel mia k", index.Find("hotel mia k")); got != "hotel mia khalifa" {
		t.Fatalf("alias rewrite = %q", got)
	}
	if got := Rewrite("mia khalfia", index.Find("mia khalfia")); got != "mia khalifa" {
		t.Fatalf("typo rewrite = %q", got)
	}
	if got := Rewrite("mia khalifa", index.Find("mia khalifa")); got != "" {
		t.Fatalf("exact name needs no rewrite, got %q", got)
	}
}

// TestCoverageCountsWordsOfKinds checks the share of query words covered.
func TestCoverageCountsWordsOfKinds(t *testing.T) {
	matches := testIndex().Find("mia khalifa blonde hotel")
	if got := Coverage(matches, 4, KindPerformer, KindStudio); got != 0.5 {
		t.Fatalf("performer coverage = %v", got)
	}
	if got := Coverage(matches, 4, KindTag); got != 0.25 {
		t.Fatalf("tag coverage = %v", got)
	}
}

// TestMentionedInUsesAliases checks a title naming an alias counts.
func TestMentionedInUsesAliases(t *testing.T) {
	match := testIndex().Find("mia khalifa")[0]
	if !match.MentionedIn("Hotel night with MIA K!") || !match.MentionedIn("Mia Khalifa's day") {
		t.Fatal("alias and possessive-free name should be found")
	}
	if strings.Contains("x", "y") || match.MentionedIn("Amiable chat") {
		t.Fatal("whole words only")
	}
}
