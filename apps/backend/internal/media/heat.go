package media

import (
	"math"
	"sort"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
)

const (
	// heatBuckets is how many equal slices of a scene's runtime heat is kept in.
	heatBuckets = 100
	// scrubWeight is how much one scrub to a slice counts against one replay of it.
	scrubWeight = 3.0
	// minHeatEvidence is the least a slice's replay score must reach before the graph is shown at all.
	minHeatEvidence = 2.0
	// minMarkerMomentSeconds keeps a marker-based moment off the opening seconds, which are rarely the best frame.
	minMarkerMomentSeconds = 10.0
)

// heatIncrement is what one viewing session adds to a single bucket.
type heatIncrement struct {
	bucket   int
	watched  float64
	scrubbed float64
}

// heatBucketOf returns the bucket a position falls in, clamped to the runtime.
func heatBucketOf(seconds, duration float64) int {
	bucket := int(seconds / duration * heatBuckets)
	if bucket < 0 {
		return 0
	}
	if bucket >= heatBuckets {
		return heatBuckets - 1
	}
	return bucket
}

// heatIncrements turns one session's playback spans and scrub targets into per-bucket additions.
// A span adds the share of each bucket it covers; a scrub adds one to the bucket it lands in.
func heatIncrements(duration float64, spans []*model.HeatSpanInput, scrubs []float64) []heatIncrement {
	if duration <= 0 {
		return nil
	}
	bucketWidth := duration / heatBuckets
	merged := map[int]*heatIncrement{}
	entry := func(bucket int) *heatIncrement {
		if merged[bucket] == nil {
			merged[bucket] = &heatIncrement{bucket: bucket}
		}
		return merged[bucket]
	}
	for _, span := range spans {
		from := math.Max(0, span.FromSeconds)
		to := math.Min(duration, span.ToSeconds)
		for bucket := heatBucketOf(from, duration); bucket <= heatBucketOf(to, duration) && to > from; bucket++ {
			bucketStart := float64(bucket) * bucketWidth
			overlap := math.Min(to, bucketStart+bucketWidth) - math.Max(from, bucketStart)
			if overlap > 0 {
				entry(bucket).watched += overlap / bucketWidth
			}
		}
	}
	for _, scrub := range scrubs {
		if scrub >= 0 && scrub <= duration {
			entry(heatBucketOf(scrub, duration)).scrubbed++
		}
	}
	out := make([]heatIncrement, 0, len(merged))
	for _, increment := range merged {
		out = append(out, *increment)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].bucket < out[j].bucket })
	return out
}

// replayBuckets turns the stored per-bucket totals into the graph: replay intensity 0 to 1 per
// bucket. Watching a slice once is not a replay, so only coverage beyond one counts, plus scrubs
// to it. Nil when no slice has enough evidence yet.
func replayBuckets(watched, scrubbed []float64) []float64 {
	scores := make([]float64, heatBuckets)
	for bucket := range scores {
		scores[bucket] = math.Max(0, watched[bucket]-1) + scrubWeight*scrubbed[bucket]
	}
	scores = smoothHeat(scores)
	highest := 0.0
	for _, score := range scores {
		highest = math.Max(highest, score)
	}
	if highest < minHeatEvidence {
		return nil
	}
	for bucket := range scores {
		scores[bucket] /= highest
	}
	return scores
}

// smoothHeat blends each bucket with its neighbours so a single scrub reads as a hill, not a needle.
func smoothHeat(scores []float64) []float64 {
	out := make([]float64, len(scores))
	for index := range scores {
		sum, weight := scores[index]*2, 2.0
		if index > 0 {
			sum += scores[index-1]
			weight++
		}
		if index < len(scores)-1 {
			sum += scores[index+1]
			weight++
		}
		out[index] = sum / weight
	}
	return out
}

// bestMomentSeconds picks the scene's thumbnail moment: the middle of the most-scrubbed bucket,
// else the earliest marker past the opening. Nil when the runtime is unknown or neither exists.
func bestMomentSeconds(scrubbed []float64, markerSeconds []float64, duration float64) *float64 {
	if duration <= 0 {
		return nil
	}
	bestBucket, bestScrubs := -1, 0.0
	for bucket, count := range scrubbed {
		if count > bestScrubs {
			bestBucket, bestScrubs = bucket, count
		}
	}
	if bestBucket >= 0 {
		moment := (float64(bestBucket) + 0.5) / heatBuckets * duration
		return &moment
	}
	return momentFromMarkers(markerSeconds, duration)
}

// momentFromMarkers is the earliest marker at least minMarkerMomentSeconds in, else the earliest of all.
func momentFromMarkers(markerSeconds []float64, duration float64) *float64 {
	inside := []float64{}
	for _, seconds := range markerSeconds {
		if seconds >= 0 && seconds <= duration {
			inside = append(inside, seconds)
		}
	}
	sort.Float64s(inside)
	for _, seconds := range inside {
		if seconds >= minMarkerMomentSeconds {
			return &seconds
		}
	}
	if len(inside) == 0 {
		return nil
	}
	return &inside[0]
}
