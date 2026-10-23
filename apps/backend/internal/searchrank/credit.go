package searchrank

// EntityCredit rates in [0, 1] how well a result belongs to the entities the
// query names. credited holds the ids the result is credited to (its stored
// performers, studio and tags, or the entity whose site page listed it); a
// result that is the entity's own record counts as credited. Where nothing is
// stored (a fresh stub), naming the entity in the title earns
// EntityMentionCredit. A match found only through a near-miss spelling counts
// FuzzyEntityFactor as much. The result is the mean over the matches, weighted
// by kind (EntityPerformerWeight, EntityStudioWeight, EntityTagWeight).
func EntityCredit(matches []EntityMatch, credited map[string]bool, ownID, title string) float64 {
	total, weights := 0.0, 0.0
	for _, match := range matches {
		weight := entityKindWeight(match.Kind)
		weights += weight
		total += weight * matchCredit(match, credited, ownID, title)
	}
	if weights == 0 {
		return 0
	}
	return total / weights
}

// matchCredit is the credit one recognised entity gives a result.
func matchCredit(match EntityMatch, credited map[string]bool, ownID, title string) float64 {
	credit := 0.0
	switch {
	case credited[match.ID] || ownID == match.ID:
		credit = 1
	case match.MentionedIn(title):
		credit = EntityMentionCredit
	}
	if match.Fuzzy {
		credit *= FuzzyEntityFactor
	}
	return credit
}

// entityKindWeight is the share of a kind in the mean entity credit.
func entityKindWeight(kind EntityKind) float64 {
	switch kind {
	case KindPerformer:
		return EntityPerformerWeight
	case KindStudio:
		return EntityStudioWeight
	}
	return EntityTagWeight
}
