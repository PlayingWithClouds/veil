package alike

// Signal weights for the composite match score. Performers and duration are the
// most reliable cross-site identifiers of the same scene; the poster hash is a
// title-independent content signal; the title itself is only a light nudge.
const (
	weightPerformer = 0.45
	weightDuration  = 0.30
	weightPoster    = 0.20
	weightTitle     = 0.05

	// dateBonus is added when release dates match exactly — a tiebreaker only.
	dateBonus = 0.03
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
	Performers      []string
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

	if len(target.Performers) > 0 && len(candidate.Performers) > 0 {
		overlap := jaccard(nameSet(target.Performers), nameSet(candidate.Performers))
		weightedSum += weightPerformer * overlap
		totalWeight += weightPerformer
	}

	if duration := durationScore(target.DurationSeconds, candidate.DurationSeconds); duration > 0 {
		weightedSum += weightDuration * duration
		totalWeight += weightDuration
	}

	if poster := HammingSimilarity(target.PosterHash, candidate.PosterHash); poster > 0 {
		weightedSum += weightPoster * poster
		totalWeight += weightPoster
	}

	if title := jaccard(tokenSet(titleTokens(target.Title)), tokenSet(titleTokens(candidate.Title))); title > 0 {
		weightedSum += weightTitle * title
		totalWeight += weightTitle
	}

	if totalWeight == 0 {
		return 0
	}

	score := weightedSum / totalWeight
	if target.Date != "" && target.Date == candidate.Date {
		score += dateBonus
	}
	return clamp01(score)
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
