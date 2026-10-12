package alike

const (
	// sameSceneDurationRatio is how close two runtimes must be (shorter over
	// longer) to be the same video: encodes and intros differ by seconds, not
	// minutes.
	sameSceneDurationRatio = 0.95

	// sameSceneTitleSimilarity is the title-token overlap that stands in for a
	// performer match when a copy credits no one.
	sameSceneTitleSimilarity = 0.9

	// sameSceneTitleMinimumTokens is the fewest title words that can stand in for a
	// performer match: a one-word title ("Anal") fits countless videos.
	sameSceneTitleMinimumTokens = 2

	// sameSceneRuntimeSlackSeconds is how far apart two runtimes may be to count
	// as the same to the second, which lets a looser title agree.
	sameSceneRuntimeSlackSeconds = 2

	// looseTitleSimilarity and looseTitleMinimumTokens are the title agreement
	// asked of two copies with the same runtime: sites reword titles ("dick" for
	// "cock"), add the cast or a series name, so only half the words must match,
	// but never from a title too short to mean anything.
	looseTitleSimilarity    = 0.5
	looseTitleMinimumTokens = 3
)

// SameScene reports whether two stored scenes are the same video on different
// sites. Unlike Score, which hunts for the full-length copy of a teaser, this is
// symmetric and strict, because a wrong merge hides a scene: both runtimes must
// be known and nearly equal, and then either the credited performers or the
// title tokens must agree. Titles that carry different numbers ("Part 14" and
// "Part 15") are different videos, and so are copies with different casts.
// Copies with the same runtime to the second need only a loose title agreement.
func SameScene(first, second Candidate) bool {
	return sameScene(first, second, true)
}

// SameSceneStrict is SameScene without the loose title rule for copies with the
// same runtime: for merging records for good, where a wrong match loses a scene
// instead of hiding one. The copies it leaves apart stay hidden from feeds.
func SameSceneStrict(first, second Candidate) bool {
	return sameScene(first, second, false)
}

// sameScene is the matcher behind SameScene and SameSceneStrict; allowLoose
// enables the loose title rule.
func sameScene(first, second Candidate, allowLoose bool) bool {
	if !similarDuration(first.DurationSeconds, second.DurationSeconds) {
		return false
	}
	firstTokens, secondTokens := titleTokens(first.Title), titleTokens(second.Title)
	if conflictingNumbers(firstTokens, secondTokens) {
		return false
	}
	if jaccard(nameSet(first.Performers), nameSet(second.Performers)) == 1 {
		return true
	}
	if castsConflict(first, second) {
		return false
	}
	firstSet, secondSet := tokenSet(firstTokens), tokenSet(secondTokens)
	if len(firstSet) >= sameSceneTitleMinimumTokens && len(secondSet) >= sameSceneTitleMinimumTokens &&
		jaccard(firstSet, secondSet) >= sameSceneTitleSimilarity {
		return true
	}
	return allowLoose && sameRuntimeToTheSecond(first.DurationSeconds, second.DurationSeconds) && looseTitleMatch(firstSet, secondSet)
}

// cast is who a copy features: its credited performers, else the ones its
// title names.
func cast(candidate Candidate) map[string]struct{} {
	if len(candidate.Performers) > 0 {
		return nameSet(candidate.Performers)
	}
	return nameSet(candidate.Named)
}

// castsConflict reports whether both copies feature someone and no one in
// common. Naming the same performer proves nothing (a performer has many videos
// of one length), but featuring different ones does.
func castsConflict(first, second Candidate) bool {
	firstCast, secondCast := cast(first), cast(second)
	if len(firstCast) == 0 || len(secondCast) == 0 {
		return false
	}
	for member := range firstCast {
		if _, shared := secondCast[member]; shared {
			return false
		}
	}
	return true
}

// sameRuntimeToTheSecond reports whether two runtimes differ by at most
// sameSceneRuntimeSlackSeconds.
func sameRuntimeToTheSecond(firstSeconds, secondSeconds int) bool {
	difference := firstSeconds - secondSeconds
	if difference < 0 {
		difference = -difference
	}
	return difference <= sameSceneRuntimeSlackSeconds
}

// looseTitleMatch reports whether the shorter title's words are all in the
// longer one, or half of all the words agree, provided the shorter title has
// enough words to mean anything.
func looseTitleMatch(first, second map[string]struct{}) bool {
	shorter, longer := first, second
	if len(shorter) > len(longer) {
		shorter, longer = longer, shorter
	}
	if len(shorter) < looseTitleMinimumTokens {
		return false
	}
	if containsAll(longer, shorter) {
		return true
	}
	return jaccard(first, second) >= looseTitleSimilarity
}

// containsAll reports whether every member of subset is in set.
func containsAll(set, subset map[string]struct{}) bool {
	for member := range subset {
		if _, found := set[member]; !found {
			return false
		}
	}
	return true
}

// conflictingNumbers reports whether each title carries a number the other lacks
// ("Parts 4" and "Parts 5", even when both say "Scene 01"): episode, part and ID
// numbers tell videos of one series or uploader apart. A title that merely has
// an extra number (a year, a resolution) does not conflict.
func conflictingNumbers(first, second []string) bool {
	firstNumbers, secondNumbers := numberSet(first), numberSet(second)
	return hasMemberOutside(firstNumbers, secondNumbers) && hasMemberOutside(secondNumbers, firstNumbers)
}

// hasMemberOutside reports whether set has a member that other lacks.
func hasMemberOutside(set, other map[string]struct{}) bool {
	for member := range set {
		if _, found := other[member]; !found {
			return true
		}
	}
	return false
}

// numberSet collects the all-digit tokens.
func numberSet(tokens []string) map[string]struct{} {
	numbers := map[string]struct{}{}
	for _, token := range tokens {
		if isNumber(token) {
			numbers[token] = struct{}{}
		}
	}
	return numbers
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

const (
	// sameGalleryMinimumImages is the smallest image count that identifies a
	// set: a handful of images is common to many unrelated galleries.
	sameGalleryMinimumImages = 5

	// sameGalleryMinimumTokens is the fewest title words a gallery needs before
	// its title can identify it.
	sameGalleryMinimumTokens = 3
)

// SameGallery reports whether two stored galleries are the same image set on
// different sites. Image counts must be known and equal, the titles must agree
// almost word for word (numbers included) and the casts must not conflict:
// sites reword little when they repost a set, and a wrong merge loses a set.
func SameGallery(first, second Candidate, firstImages, secondImages int) bool {
	if firstImages < sameGalleryMinimumImages || firstImages != secondImages {
		return false
	}
	firstTokens, secondTokens := titleTokens(first.Title), titleTokens(second.Title)
	if len(firstTokens) < sameGalleryMinimumTokens || len(secondTokens) < sameGalleryMinimumTokens {
		return false
	}
	if conflictingNumbers(firstTokens, secondTokens) || castsConflict(first, second) {
		return false
	}
	return jaccard(tokenSet(firstTokens), tokenSet(secondTokens)) >= sameSceneTitleSimilarity
}
