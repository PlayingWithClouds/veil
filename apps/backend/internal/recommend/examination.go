package recommend

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// learnMaxPooledRate caps the pooled reward rate after position correction,
// which can lift rewards above one per showing.
const learnMaxPooledRate = 0.9

// examinationModel is the fitted chance, per surface and position bucket,
// that a showing is looked at, relative to the best-placed bucket of its
// surface (1 = as likely as the top). Showings logged without a position
// take the surface's average.
type examinationModel struct {
	bySurface map[string]*surfaceExamination
}

// surfaceExamination is the examination of one surface's position buckets.
type surfaceExamination struct {
	byBucket map[int]float64
	// unknown is the propensity of showings without a position: the mean of
	// the known buckets weighted by their showings, 1 when none is known.
	unknown float64
}

// fitExamination estimates, for each surface, the examination of its position
// buckets from the cells' showings and rewards. The reward of a source in a
// bucket is modelled as quality(source) * examination(bucket); the two are
// re-estimated from each other, the examination shrunk toward a decaying
// prior curve. Surfaces without positions get no correction.
func fitExamination(cells map[outcomeCell]outcomeCounts) *examinationModel {
	surfaces := map[string]bool{}
	for cell := range cells {
		surfaces[cell.surface] = true
	}
	model := &examinationModel{bySurface: map[string]*surfaceExamination{}}
	for _, surface := range sortedKeys(surfaces) {
		model.bySurface[surface] = fitSurfaceExamination(cells, surface)
	}
	return model
}

// fitSurfaceExamination fits one surface; see fitExamination.
func fitSurfaceExamination(cells map[outcomeCell]outcomeCounts, surface string) *surfaceExamination {
	known := map[outcomeCell]outcomeCounts{}
	buckets := map[int]bool{}
	for cell, counts := range cells {
		if cell.surface == surface && cell.bucket != unknownPositionBucket {
			known[cell] = counts
			buckets[cell.bucket] = true
		}
	}
	fitted := &surfaceExamination{byBucket: map[int]float64{}, unknown: 1}
	if len(known) == 0 {
		return fitted
	}
	order := sortedBuckets(buckets)
	examination := map[int]float64{}
	for _, bucket := range order {
		examination[bucket] = positionPrior(bucket)
	}
	for iteration := 0; iteration < positionFitIterations; iteration++ {
		quality := sourceQuality(known, examination)
		examination = refitBuckets(known, order, quality)
	}
	fitted.byBucket = examination
	fitted.unknown = weightedMeanExamination(known, examination)
	return fitted
}

// sourceQuality is each source's reward per unit of examined showing.
func sourceQuality(cells map[outcomeCell]outcomeCounts, examination map[int]float64) map[string]float64 {
	rewards, exposures := map[string]float64{}, map[string]float64{}
	for cell, counts := range cells {
		rewards[cell.source] += counts.reward
		exposures[cell.source] += examination[cell.bucket] * counts.shows
	}
	quality := map[string]float64{}
	for source, exposure := range exposures {
		if exposure > 0 {
			quality[source] = rewards[source] / exposure
		}
	}
	return quality
}

// refitBuckets re-estimates each bucket's examination from the rewards it
// earned against the rewards its showings would earn at full examination,
// shrunk toward the prior curve; the result is scaled so the best-placed
// bucket is 1, made non-increasing with depth and floored.
func refitBuckets(cells map[outcomeCell]outcomeCounts, order []int, quality map[string]float64) map[int]float64 {
	earned, expected := map[int]float64{}, map[int]float64{}
	for cell, counts := range cells {
		earned[cell.bucket] += counts.reward
		expected[cell.bucket] += quality[cell.source] * counts.shows
	}
	raw := map[int]float64{}
	for _, bucket := range order {
		raw[bucket] = (earned[bucket] + positionPriorStrength*positionPrior(bucket)) / (expected[bucket] + positionPriorStrength)
	}
	reference := raw[order[0]]
	fitted := map[int]float64{}
	ceiling := 1.0
	for _, bucket := range order {
		value := clampFloat(raw[bucket]/reference, positionMinExposure, 1)
		ceiling = math.Min(ceiling, value)
		fitted[bucket] = ceiling
	}
	return fitted
}

// weightedMeanExamination is the mean examination of the known buckets,
// weighted by their showings.
func weightedMeanExamination(cells map[outcomeCell]outcomeCounts, examination map[int]float64) float64 {
	var shows, weighted float64
	for cell, counts := range cells {
		shows += counts.shows
		weighted += counts.shows * examination[cell.bucket]
	}
	if shows <= 0 {
		return 1
	}
	return weighted / shows
}

// positionPrior is the examination assumed for a bucket before the data says
// otherwise: it falls with depth.
func positionPrior(bucket int) float64 {
	return 1 / (1 + positionPriorSlope*float64(bucket))
}

// of is the examination of a showing at a position bucket of a surface; 1
// for a surface the model has not seen.
func (model *examinationModel) of(surface string, bucket int) float64 {
	fitted, found := model.bySurface[surface]
	if !found {
		return 1
	}
	if bucket == unknownPositionBucket {
		return fitted.unknown
	}
	examination, found := fitted.byBucket[bucket]
	if !found {
		return fitted.unknown
	}
	return examination
}

// correctedBySource sums the cells per source, each cell's reward divided by
// its examination so that a source is judged as if all its showings had been
// at the top. Showings and clicks are counted as they happened.
func (model *examinationModel) correctedBySource(cells map[outcomeCell]outcomeCounts) map[string]outcomeCounts {
	out := map[string]outcomeCounts{}
	for cell, counts := range cells {
		counts.reward /= model.of(cell.surface, cell.bucket)
		out[cell.source] = out[cell.source].plus(counts)
	}
	return out
}

// describe renders the fitted curves for the log, e.g. "home[1.00 0.62 0.41]".
func (model *examinationModel) describe() string {
	var parts []string
	for _, surface := range sortedKeys(model.bySurface) {
		fitted := model.bySurface[surface]
		if len(fitted.byBucket) == 0 {
			parts = append(parts, surface+"[no positions]")
			continue
		}
		var values []string
		for _, bucket := range sortedBuckets(boolSet(fitted.byBucket)) {
			values = append(values, fmt.Sprintf("%.2f", fitted.byBucket[bucket]))
		}
		parts = append(parts, fmt.Sprintf("%s[%s]", surface, strings.Join(values, " ")))
	}
	return strings.Join(parts, " ")
}

// boolSet is the set of a map's integer keys.
func boolSet(values map[int]float64) map[int]bool {
	out := map[int]bool{}
	for key := range values {
		out[key] = true
	}
	return out
}

// sortedBuckets lists a set of buckets in order.
func sortedBuckets(set map[int]bool) []int {
	buckets := make([]int, 0, len(set))
	for bucket := range set {
		buckets = append(buckets, bucket)
	}
	sort.Ints(buckets)
	return buckets
}
