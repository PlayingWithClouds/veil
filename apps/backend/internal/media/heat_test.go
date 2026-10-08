package media

import (
	"testing"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
)

func TestHeatIncrementsSpreadSpansAndScrubs(t *testing.T) {
	// A 1000 s scene has 10 s buckets.
	increments := heatIncrements(1000,
		[]*model.HeatSpanInput{{FromSeconds: 5, ToSeconds: 25}},
		[]float64{12, 999, 2000, -3})

	byBucket := map[int]heatIncrement{}
	for _, increment := range increments {
		byBucket[increment.bucket] = increment
	}
	if got := byBucket[0].watched; got != 0.5 {
		t.Errorf("bucket 0 watched = %v, want 0.5", got)
	}
	if got := byBucket[1].watched; got != 1 {
		t.Errorf("bucket 1 watched = %v, want 1", got)
	}
	if got := byBucket[2].watched; got != 0.5 {
		t.Errorf("bucket 2 watched = %v, want 0.5", got)
	}
	if got := byBucket[1].scrubbed; got != 1 {
		t.Errorf("bucket 1 scrubbed = %v, want 1", got)
	}
	if got := byBucket[99].scrubbed; got != 1 {
		t.Errorf("bucket 99 scrubbed = %v, want 1", got)
	}
	if len(byBucket) != 4 {
		t.Errorf("touched %d buckets, want 4 (out-of-range scrubs ignored)", len(byBucket))
	}
}

func TestHeatIncrementsNeedARuntime(t *testing.T) {
	if got := heatIncrements(0, []*model.HeatSpanInput{{FromSeconds: 0, ToSeconds: 10}}, []float64{1}); got != nil {
		t.Errorf("got %v, want nil for an unknown runtime", got)
	}
}

func TestReplayBucketsIgnoreASingleViewing(t *testing.T) {
	watched := make([]float64, heatBuckets)
	for bucket := 0; bucket < 50; bucket++ {
		watched[bucket] = 1
	}
	if got := replayBuckets(watched, make([]float64, heatBuckets)); got != nil {
		t.Errorf("a single viewing produced a graph: %v", got)
	}
}

func TestReplayBucketsPeakWhereScrubbedAndReplayed(t *testing.T) {
	watched := make([]float64, heatBuckets)
	scrubbed := make([]float64, heatBuckets)
	for bucket := 40; bucket < 45; bucket++ {
		watched[bucket] = 3
	}
	scrubbed[42] = 2
	buckets := replayBuckets(watched, scrubbed)
	if buckets == nil {
		t.Fatal("expected a graph")
	}
	peak := 0
	for bucket, value := range buckets {
		if value > buckets[peak] {
			peak = bucket
		}
		if value < 0 || value > 1 {
			t.Fatalf("bucket %d = %v outside 0..1", bucket, value)
		}
	}
	if peak != 42 || buckets[42] != 1 {
		t.Errorf("peak at %d = %v, want bucket 42 at 1", peak, buckets[42])
	}
	if buckets[90] != 0 {
		t.Errorf("untouched bucket = %v, want 0", buckets[90])
	}
}

func TestBestMomentPrefersScrubsThenMarkers(t *testing.T) {
	scrubbed := make([]float64, heatBuckets)
	if got := bestMomentSeconds(scrubbed, nil, 1000); got != nil {
		t.Errorf("no evidence gave %v, want nil", *got)
	}
	if got := bestMomentSeconds(scrubbed, []float64{3, 400, 700}, 1000); got == nil || *got != 400 {
		t.Errorf("marker moment = %v, want 400 (first past the opening)", got)
	}
	if got := bestMomentSeconds(scrubbed, []float64{3, 5}, 1000); got == nil || *got != 3 {
		t.Errorf("early-only markers = %v, want the earliest, 3", got)
	}
	scrubbed[25] = 2
	if got := bestMomentSeconds(scrubbed, []float64{400}, 1000); got == nil || *got != 255 {
		t.Errorf("scrub moment = %v, want the middle of bucket 25, 255", got)
	}
	if got := bestMomentSeconds(scrubbed, nil, 0); got != nil {
		t.Errorf("unknown runtime gave %v, want nil", *got)
	}
}
