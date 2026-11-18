package alike

const (
	// fullLengthMinimumSeconds is the shortest runtime a full-length scene
	// can have: anything shorter is another clip.
	fullLengthMinimumSeconds = 600

	// fullLengthMinimumRatio is how much longer than the clip the full
	// version must run.
	fullLengthMinimumRatio = 1.5
)

// FullLengthOf reports whether candidate looks like the full-length video a
// shorter clip (a teaser, trailer or excerpt) was cut from: it runs at least
// fullLengthMinimumRatio as long and fullLengthMinimumSeconds, the clip's
// title words are in its title or half of all words agree, neither the
// numbers in the titles nor the casts tell them apart, and when both credit
// performers they share one. Only used to suggest and collect the long
// version, never to merge.
func FullLengthOf(clip, candidate Candidate) bool {
	if clip.DurationSeconds <= 0 || candidate.DurationSeconds < fullLengthMinimumSeconds {
		return false
	}
	if float64(candidate.DurationSeconds) < float64(clip.DurationSeconds)*fullLengthMinimumRatio {
		return false
	}
	clipTokens, candidateTokens := titleTokens(clip.Title), titleTokens(candidate.Title)
	if conflictingNumbers(clipTokens, candidateTokens) || castsConflict(clip, candidate) {
		return false
	}
	if len(clip.Performers) > 0 && len(candidate.Performers) > 0 && !sharesMember(nameSet(clip.Performers), nameSet(candidate.Performers)) {
		return false
	}
	return looseTitleMatch(tokenSet(clipTokens), tokenSet(candidateTokens))
}

// sharesMember reports whether the sets have a member in common.
func sharesMember(first, second map[string]struct{}) bool {
	for member := range first {
		if _, found := second[member]; found {
			return true
		}
	}
	return false
}
