package recommend

import "math"

// runtimeBandSeconds are the centres of the runtime bands, about a factor of
// two apart (3, 6, 12, 25, 50 and 90 minutes). A runtime is spread over its
// two neighbouring centres by position on the log scale, so a preference has
// no cliffs at band edges.
var runtimeBandSeconds = [runtimeBands]float64{180, 360, 720, 1500, 3000, 5400}

// runtimeBands is how many runtime bands there are.
const runtimeBands = 6

// runtimeShares splits a runtime over two adjacent bands: the lower band, and
// the share (0..1) that goes to the band above. ok is false for an unknown
// runtime.
func runtimeShares(durationSeconds int) (lowerBand int, upperShare float64, ok bool) {
	if durationSeconds <= 0 {
		return 0, 0, false
	}
	seconds := float64(durationSeconds)
	if seconds <= runtimeBandSeconds[0] {
		return 0, 0, true
	}
	last := runtimeBands - 1
	for band := 0; band < last; band++ {
		if seconds < runtimeBandSeconds[band+1] {
			span := math.Log(runtimeBandSeconds[band+1] / runtimeBandSeconds[band])
			return band, math.Log(seconds/runtimeBandSeconds[band]) / span, true
		}
	}
	return last - 1, 1, true
}

// addRuntime spreads a signalled item's weight over the bands of its runtime.
func (taste *profile) addRuntime(durationSeconds int, weight float64) {
	lowerBand, upperShare, ok := runtimeShares(durationSeconds)
	if !ok {
		return
	}
	taste.runtime[lowerBand] += weight * (1 - upperShare)
	taste.runtime[lowerBand+1] += weight * upperShare
}

// normalizeRuntime scales the bands into [-1, 1] by the larger of the
// strongest band and runtimeMinMass, so thin evidence stays a weak preference.
func (taste *profile) normalizeRuntime() {
	largest := runtimeMinMass
	for _, value := range taste.runtime {
		largest = math.Max(largest, math.Abs(value))
	}
	for band := range taste.runtime {
		taste.runtime[band] /= largest
	}
}

// runtimeAffinity is the profile's affinity for a runtime in [-1, 1], 0 when
// the runtime is unknown.
func (taste *profile) runtimeAffinity(durationSeconds int) float64 {
	lowerBand, upperShare, ok := runtimeShares(durationSeconds)
	if !ok {
		return 0
	}
	return taste.runtime[lowerBand]*(1-upperShare) + taste.runtime[lowerBand+1]*upperShare
}
