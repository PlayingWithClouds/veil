package alike

const (
	// sameSceneDurationRatio is how close two runtimes must be (shorter over
	// longer) to be the same video: encodes and intros differ by seconds, not
	// minutes.
	sameSceneDurationRatio = 0.95

	// sameSceneTitleSimilarity is the title-token overlap that stands in for a
	// performer match when a copy credits no one.
	sameSceneTitleSimilarity = 0.9
)

// SameScene reports whether two stored scenes are the same video on different
// sites. Unlike Score, which hunts for the full-length copy of a teaser, this is
// symmetric and strict, because a wrong merge hides a scene: both runtimes must
// be known and nearly equal, and then either the credited performers or the
// title tokens must agree.
func SameScene(first, second Candidate) bool {
	if !similarDuration(first.DurationSeconds, second.DurationSeconds) {
		return false
	}
	if jaccard(nameSet(first.Performers), nameSet(second.Performers)) == 1 {
		return true
	}
	titleOverlap := jaccard(tokenSet(titleTokens(first.Title)), tokenSet(titleTokens(second.Title)))
	return titleOverlap >= sameSceneTitleSimilarity
}

// similarDuration reports whether both runtimes are known and within
// sameSceneDurationRatio of each other.
func similarDuration(firstSeconds, secondSeconds int) bool {
	if firstSeconds <= 0 || secondSeconds <= 0 {
		return false
	}
	shorter, longer := firstSeconds, secondSeconds
	if shorter > longer {
		shorter, longer = longer, shorter
	}
	return float64(shorter)/float64(longer) >= sameSceneDurationRatio
}
