package searchrank

import "testing"

// TestSpecificityGrowsWithWordsAndNamedEntities checks one topic word is the
// broadest, more words and a named performer narrow the query.
func TestSpecificityGrowsWithWordsAndNamedEntities(t *testing.T) {
	topic := Specificity(QueryProfile{WordCount: 1, TagCoverage: 1, LibraryMatches: -1})
	twoWords := Specificity(QueryProfile{WordCount: 2, LibraryMatches: -1})
	person := Specificity(QueryProfile{WordCount: 2, EntityCoverage: 1, LibraryMatches: -1})
	title := Specificity(QueryProfile{WordCount: 6, LibraryMatches: -1})
	if !(topic < twoWords && twoWords < person && person < title) {
		t.Fatalf("topic %v, two words %v, person %v, title %v", topic, twoWords, person, title)
	}
	if topic != 0 || title > 1 {
		t.Fatalf("bounds: topic %v, title %v", topic, title)
	}
}

// TestSpecificityFallsWithManyLibraryMatches checks a query matching a lot of
// the library is broader than one matching a handful, and zero matches say
// nothing.
func TestSpecificityFallsWithManyLibraryMatches(t *testing.T) {
	base := QueryProfile{WordCount: 3}
	base.LibraryMatches = 2
	few := Specificity(base)
	base.LibraryMatches = 40
	some := Specificity(base)
	base.LibraryMatches = LibraryManyMatches
	many := Specificity(base)
	base.LibraryMatches = 0
	none := Specificity(base)
	base.LibraryMatches = -1
	unknown := Specificity(base)
	if !(few > some && some > many) {
		t.Fatalf("few %v, some %v, many %v", few, some, many)
	}
	if none != unknown {
		t.Fatalf("no match %v should equal uncounted %v", none, unknown)
	}
}

// TestPersonalWeightLeansOnTasteForBroadQueries checks the bounds and the
// direction of the taste weight.
func TestPersonalWeightLeansOnTasteForBroadQueries(t *testing.T) {
	broad, specific := PersonalWeight(0), PersonalWeight(1)
	if broad != WeightPersonal*PersonalBroadFactor || specific != WeightPersonal*PersonalSpecificFactor {
		t.Fatalf("broad %v, specific %v", broad, specific)
	}
	if middle := PersonalWeight(0.5); middle >= broad || middle <= specific {
		t.Fatalf("middle %v not between %v and %v", middle, specific, broad)
	}
	if PersonalWeight(-3) != broad || PersonalWeight(7) != specific {
		t.Fatal("specificity is clamped to [0, 1]")
	}
}
