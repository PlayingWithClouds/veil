package alike

const (
	// sameSceneDurationRatio is how close two runtimes must be (shorter over
	// longer) to be the same video: encodes and intros differ by seconds, not
	// minutes.
	sameSceneDurationRatio = 0.95

	// sameSceneTitleSimilarity is the title-token overlap that stands in for a
	// performer match when a copy credits no one.
	sameSceneTitleSimilarity = 0.9

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
// "Part 15") are different videos. Copies with the same runtime to the second
// need only a loose title agreement.
func SameScene(first, second Candidate) bool {
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
	firstSet, secondSet := tokenSet(firstTokens), tokenSet(secondTokens)
	if jaccard(firstSet, secondSet) >= sameSceneTitleSimilarity {
		return true
	}
	return sameRuntimeToTheSecond(first.DurationSeconds, second.DurationSeconds) && looseTitleMatch(firstSet, secondSet)
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

// conflictingNumbers reports whether both titles carry numbers and share none:
// episode, part and ID numbers tell videos of one series or uploader apart.
func conflictingNumbers(first, second []string) bool {
	firstNumbers, secondNumbers := numberSet(first), numberSet(second)
	if len(firstNumbers) == 0 || len(secondNumbers) == 0 {
		return false
	}
	for number := range firstNumbers {
		if _, shared := secondNumbers[number]; shared {
			return false
		}
	}
	return true
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
