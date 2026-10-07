package alike

// durationScore rates a candidate's runtime against the current scene's. It
// encodes the core intent — find the full-length copy, not another teaser:
//
//   - equal or longer than the current scene → ideal (1.0), because the current
//     scene is often the short cut we want to replace, and its full version can
//     be many times longer than the teaser;
//   - shorter → penalized in proportion to how much is missing.
//
// Wrong-match protection is left to the performer and poster signals, so a
// legitimately much-longer full-length copy is never penalized here.
//
// Returns 0 when either duration is unknown; callers treat that as "no signal".
func durationScore(targetSeconds, candidateSeconds int) float64 {
	if targetSeconds <= 0 || candidateSeconds <= 0 {
		return 0
	}

	if candidateSeconds >= targetSeconds {
		return 1
	}
	// Shorter than the current scene: linear falloff. Half the length → 0.5.
	return float64(candidateSeconds) / float64(targetSeconds)
}
