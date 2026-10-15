package alike

const (
	// teaserSeconds is the runtime under which a target counts as a teaser,
	// whose full-length copy is wanted, so a longer candidate stays plausible.
	teaserSeconds = 180
	// teaserFullLengthScore is a longer candidate's score against a teaser.
	teaserFullLengthScore = 0.7
	// durationMismatchRatio is the shorter-over-longer ratio at which runtimes
	// stop meaning anything (score 0); sameSceneDurationRatio and up score 1.
	durationMismatchRatio = 0.7
)

// durationScore rates how well a candidate's runtime fits the current scene's,
// in 0..1: runtimes within a few percent are the same video (1), and the score
// falls to 0 by durationMismatchRatio. Against a teaser a longer candidate may
// be its full-length copy and keeps teaserFullLengthScore. Unknown runtimes
// score 0; Score leaves the signal out then.
func durationScore(targetSeconds, candidateSeconds int) float64 {
	if targetSeconds <= 0 || candidateSeconds <= 0 {
		return 0
	}
	shorter, longer := min(targetSeconds, candidateSeconds), max(targetSeconds, candidateSeconds)
	ratio := float64(shorter) / float64(longer)
	closeness := (ratio - durationMismatchRatio) / (sameSceneDurationRatio - durationMismatchRatio)
	closeness = clamp01(closeness)
	if targetSeconds < teaserSeconds && candidateSeconds > targetSeconds {
		return max(closeness, teaserFullLengthScore)
	}
	return closeness
}

// RuntimeRulesOut reports whether a listing's runtime alone shows it is not
// the target's video: both runtimes known and too far apart to score.
func RuntimeRulesOut(target Target, candidateSeconds int) bool {
	if target.DurationSeconds <= 0 || candidateSeconds <= 0 {
		return false
	}
	return durationScore(target.DurationSeconds, candidateSeconds) == 0
}
