package lookup

import (
	"testing"

	"github.com/playingwithclouds/veil/internal/schema"
)

// TestMergeCorroborates merges two plugins' findings: a record named twice
// keeps both pieces of evidence and outranks one named once with a higher
// first score.
func TestMergeCorroborates(t *testing.T) {
	ada := schema.RecordRef{Kind: schema.RecordKindPerformer, Id: "ada"}
	bea := schema.RecordRef{Kind: schema.RecordKindPerformer, Id: "bea"}
	merged := merge([][]finding{
		{{ref: bea, score: 0.85, evidence: schema.Evidence{Plugin: "faces"}}, {ref: ada, score: 0.8, evidence: schema.Evidence{Plugin: "faces"}}},
		{{ref: ada, score: 0.6, evidence: schema.Evidence{Plugin: "forum"}}},
	})
	if len(merged) != 2 || merged[0].Item != ada || len(merged[0].Evidence) != 2 {
		t.Fatalf("merged = %+v", merged)
	}
	if *merged[0].Score < 0.89 || *merged[0].Score > 0.91 {
		t.Errorf("score = %v, want 0.8 raised by one corroboration", *merged[0].Score)
	}
}
