package searchrank

import "math"

// QueryProfile is what the search knows about the shape of a query.
type QueryProfile struct {
	WordCount int
	// EntityCoverage is the share of the words that recognised performer and
	// studio names cover, TagCoverage the share covered by tag names.
	EntityCoverage float64
	TagCoverage    float64
	// LibraryMatches is how many library scenes the query matches; a negative
	// value means it was not counted.
	LibraryMatches int
}

// Specificity rates in [0, 1] how narrow a query is: 0 is a topic ("anal"),
// 1 asks for one particular video. It starts from the word count, rises with
// the share of the words a recognised performer or studio covers, falls with
// the share that are tag names and with a large number of library matches,
// and rises slightly when the library has only a handful.
func Specificity(profile QueryProfile) float64 {
	if profile.WordCount == 0 {
		return 0
	}
	value := wordSpecificity[min(profile.WordCount, len(wordSpecificity))-1]
	value += SpecificEntityBonus * profile.EntityCoverage
	value -= BroadTagPenalty * profile.TagCoverage
	value += libraryAdjustment(profile.LibraryMatches)
	return max(0, min(1, value))
}

// libraryAdjustment is the specificity change for a library match count:
// LibraryNarrowBonus at LibraryFewMatches or fewer, -LibraryBroadPenalty at
// LibraryManyMatches or more, log-scaled between. No change for zero matches
// (the library says nothing about the query) or an uncounted query.
func libraryAdjustment(matches int) float64 {
	if matches <= 0 {
		return 0
	}
	broadness := math.Log(float64(matches)/LibraryFewMatches) / math.Log(float64(LibraryManyMatches)/LibraryFewMatches)
	broadness = max(0, min(1, broadness))
	return LibraryNarrowBonus*(1-broadness) - LibraryBroadPenalty*broadness
}

// PersonalWeight is the weight of the taste term for a query of the given
// specificity: PersonalBroadFactor times WeightPersonal for the broadest,
// PersonalSpecificFactor times for the most specific, linear in between. Broad
// queries have many near-equal results where taste is a good tie-breaker;
// a specific one has the answer the user named, which taste must not outvote.
func PersonalWeight(specificity float64) float64 {
	specificity = max(0, min(1, specificity))
	factor := PersonalBroadFactor + (PersonalSpecificFactor-PersonalBroadFactor)*specificity
	return WeightPersonal * factor
}
