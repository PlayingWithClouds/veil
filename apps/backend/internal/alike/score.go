package alike

// Signal weights for the composite match score. Few scenes credit performers,
// so runtime and title carry most matches: the runtime when both are known
// (nearly every listing has one), the title's distinctive words always. The
// poster hash only counts when it is close, since different sites pick
// different frames and unrelated posters still share a coarse layout.
const (
	weightPerformer = 0.35
	weightDuration  = 0.30
	weightTitle     = 0.25
	weightPoster    = 0.10

	// dateBonus is added when release dates match exactly — a tiebreaker only.
	dateBonus = 0.03

	// posterSimilarityFloor is the hash similarity below which posters count
	// as unrelated; above it the similarity is rescaled to 0..1.
	posterSimilarityFloor = 0.8

	// sameSceneScore is the floor for a candidate SameScene recognises as the
	// target's own video (runtime and title or cast agree).
	sameSceneScore = 0.95
)

// Target describes the scene the user is currently viewing.
type Target struct {
	Performers      []string
	Studio          string
	Title           string
	DurationSeconds int
	PosterHash      uint64
	Date            string
}

// Candidate is one potential alternate source, filled in progressively: at
// search time only Title/PosterHash/Date are known; after a stage-2 scrape the
// performers and duration are populated too.
type Candidate struct {
	Performers []string
	// Named are the known performers the title mentions, for a copy that
	// credits no one; SameScene uses them only to tell two videos apart.
	Named           []string
	Title           string
	DurationSeconds int
	PosterHash      uint64
	Date            string
}

// Score rates how likely a candidate is the same scene as the target, in 0..1.
// Only the signals present on both sides contribute, and the result is
// renormalized over their weights — so a stage-1 candidate scored on poster and
// title alone is still comparable to a fully scraped one.
func Score(target Target, candidate Candidate) float64 {
	weightedSum := 0.0
	totalWeight := 0.0
	add := func(weight, similarity float64) {
		weightedSum += weight * similarity
		totalWeight += weight
	}

	if len(target.Performers) > 0 && len(candidate.Performers) > 0 {
		add(weightPerformer, jaccard(nameSet(target.Performers), nameSet(candidate.Performers)))
	}
	if target.DurationSeconds > 0 && candidate.DurationSeconds > 0 {
		add(weightDuration, durationScore(target.DurationSeconds, candidate.DurationSeconds))
	}
	if target.Title != "" && candidate.Title != "" {
		add(weightTitle, titleSimilarity(target.Title, candidate.Title))
	}
	if target.PosterHash != 0 && candidate.PosterHash != 0 {
		add(weightPoster, posterSimilarity(target.PosterHash, candidate.PosterHash))
	}
	if totalWeight == 0 {
		return 0
	}

	score := weightedSum / totalWeight
	if target.Date != "" && target.Date == candidate.Date {
		score += dateBonus
	}
	if SameScene(targetAsCandidate(target), candidate) {
		score = max(score, sameSceneScore)
	}
	return clamp01(score)
}

// titleSimilarity is the share of the shorter title's distinctive words the
// other title has too, so a copy whose site appends a cast or tag list still
// matches, but never from a single shared word.
func titleSimilarity(first, second string) float64 {
	firstTokens, secondTokens := tokenSet(titleTokens(first)), tokenSet(titleTokens(second))
	smaller := min(len(firstTokens), len(secondTokens))
	if smaller == 0 {
		return 0
	}
	shared := 0
	for token := range firstTokens {
		if _, found := secondTokens[token]; found {
			shared++
		}
	}
	if shared < 2 {
		return 0
	}
	return float64(shared) / float64(smaller)
}

// posterSimilarity rescales the hash similarity above posterSimilarityFloor
// to 0..1; anything below is unrelated (0).
func posterSimilarity(first, second uint64) float64 {
	similarity := HammingSimilarity(first, second)
	if similarity <= posterSimilarityFloor {
		return 0
	}
	return (similarity - posterSimilarityFloor) / (1 - posterSimilarityFloor)
}

// targetAsCandidate views the target as a Candidate for SameScene.
func targetAsCandidate(target Target) Candidate {
	return Candidate{Performers: target.Performers, Title: target.Title, DurationSeconds: target.DurationSeconds, PosterHash: target.PosterHash, Date: target.Date}
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
