package searchrank

import "testing"

// TestEditDistanceCountsTypos checks insertion, deletion, substitution and
// adjacent swaps each cost one.
func TestEditDistanceCountsTypos(t *testing.T) {
	cases := []struct {
		first, second string
		want          int
	}{
		{"khalifa", "khalifa", 0},
		{"khalifa", "khalfia", 1},
		{"khalifa", "khalifaa", 1},
		{"khalifa", "khalia", 1},
		{"khalifa", "khelifa", 1},
		{"khalifa", "kahlifa", 1},
		{"khalifa", "kalifah", 2},
	}
	for _, testCase := range cases {
		if got := EditDistance(testCase.first, testCase.second); got != testCase.want {
			t.Errorf("EditDistance(%q, %q) = %d, want %d", testCase.first, testCase.second, got, testCase.want)
		}
	}
}

// TestWithinTyposScalesWithLength checks short words tolerate none, ordinary
// words one and long words two, and the first letter must agree.
func TestWithinTyposScalesWithLength(t *testing.T) {
	if WithinTypos("mia", "mja") {
		t.Fatal("three letters tolerate no typo")
	}
	if !WithinTypos("lagoon", "lagon") || WithinTypos("lagoon", "lagnn") {
		t.Fatal("six letters tolerate exactly one")
	}
	if !WithinTypos("backstage", "bakcstge") {
		t.Fatal("nine letters tolerate two")
	}
	if WithinTypos("date", "mate") {
		t.Fatal("a different first letter is another word")
	}
	if WithinTypos("lagoon", "lagoon") {
		t.Fatal("identical words are not typos")
	}
}

// TestTextMatchTypoEarnsLessThanExact checks a misspelled word still earns
// credit but never as much as the right spelling.
func TestTextMatchTypoEarnsLessThanExact(t *testing.T) {
	fields := TextFields{Title: "Garden party", Performers: "Mia Khalifa"}
	exact := TextMatch(QueryWords("mia khalifa"), fields)
	typo := TextMatch(QueryWords("mia khalfia"), fields)
	none := TextMatch(QueryWords("mia zzzzzzz"), fields)
	if !(exact > typo && typo > none) {
		t.Fatalf("exact %v, typo %v, none %v", exact, typo, none)
	}
}
