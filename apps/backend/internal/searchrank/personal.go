package searchrank

import "math"

// Affinities are the profile's affinities, each in [-1, 1], for what a result
// credits: one per performer and tag, one for the studio and one for the site
// that lists it. Zero means no signal.
type Affinities struct {
	Performers []float64
	Tags       []float64
	Studio     float64
	Site       float64
}

// Personal mixes affinities into one value in [-1, 1], the way the
// recommendation ranking does: one loved performer carries a result (an
// all-disliked cast sinks it), tag affinities add up damped by their count,
// then studio and site.
func Personal(affinities Affinities) float64 {
	total := PersonalWeightPerformer * strongest(affinities.Performers)
	total += PersonalWeightTag * damped(affinities.Tags)
	total += PersonalWeightStudio * affinities.Studio
	total += PersonalWeightSite * affinities.Site
	return math.Max(-1, math.Min(1, total))
}

// strongest is the highest value, or the lowest when none is positive.
func strongest(values []float64) float64 {
	highest, lowest := 0.0, 0.0
	for _, value := range values {
		highest = math.Max(highest, value)
		lowest = math.Min(lowest, value)
	}
	if highest > 0 {
		return highest
	}
	return lowest
}

// damped sums the values over the square root of their count, clamped to
// [-1, 1], so many mild tags do not outweigh one strong one.
func damped(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return math.Max(-1, math.Min(1, sum/math.Sqrt(float64(len(values)))))
}
