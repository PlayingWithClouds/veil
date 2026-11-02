package graphql

import (
	"context"
	"fmt"
	"math"
	"slices"

	"github.com/playingwithclouds/veil/internal/embeddings"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/schema"
)

// namedMergeKinds are the kinds the user merges by their own judgement.
// Media merges need proof and go through mergeAlike.
var namedMergeKinds = []schema.RecordKind{
	schema.RecordKindPerformer, schema.RecordKindStudio, schema.RecordKindTag, schema.RecordKindGroup,
}

// mergeNamed merges two performers, studios, tags or groups; two media
// records only when they're provably the same.
func (r *Resolver) mergeNamed(ctx context.Context, kind schema.RecordKind, keepID, mergeID string) error {
	if slices.Contains(namedMergeKinds, kind) {
		return r.library.Merge(ctx, kind, keepID, mergeID)
	}
	same, err := r.library.ProvablySame(ctx, kind, keepID, mergeID)
	if err != nil {
		return err
	}
	if !same {
		return fmt.Errorf("the two %ss aren't provably the same: they need a matching fingerprint first", kind)
	}
	return r.library.Merge(ctx, kind, keepID, mergeID)
}

// queuedJob loads the job a service just queued.
func (r *Resolver) queuedJob(jobID string, err error) (*jobs.Job, error) {
	if err != nil {
		return nil, err
	}
	return r.queue.Get(context.Background(), jobID)
}

// float32s narrows a vector from the API.
func float32s(vector []float64) []float32 {
	narrowed := make([]float32, len(vector))
	for index, value := range vector {
		narrowed[index] = float32(value)
	}
	return narrowed
}

// neighbourMatches turns embedding neighbours into matches; the score is
// the cosine similarity clamped to 0..1.
func neighbourMatches(neighbours []embeddings.Neighbour) []*schema.Match {
	matches := make([]*schema.Match, 0, len(neighbours))
	for _, neighbour := range neighbours {
		score := float32(math.Max(0, math.Min(1, neighbour.Similarity)))
		detail := fmt.Sprintf("%d%% similar", int(score*100))
		matches = append(matches, &schema.Match{
			Item:     neighbour.Subject,
			Score:    &score,
			Evidence: []schema.Evidence{{Plugin: "veil", Detail: &detail}},
		})
	}
	return matches
}
