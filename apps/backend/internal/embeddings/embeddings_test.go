package embeddings

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestNearestAndSimilarTo stores three frame vectors and finds the closest
// one to a query and to a record, never mixing models.
func TestNearestAndSimilarTo(t *testing.T) {
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	service := New(database)
	video := func(id string) schema.RecordRef { return schema.RecordRef{Kind: schema.RecordKindVideo, Id: id} }
	for _, stored := range []NewEmbedding{
		{Subject: video("a"), Kind: schema.EmbeddingKindVisual, Model: "clip-1", Vector: []float32{1, 0, 0}},
		{Subject: video("b"), Kind: schema.EmbeddingKindVisual, Model: "clip-1", Vector: []float32{0.9, 0.1, 0}},
		{Subject: video("c"), Kind: schema.EmbeddingKindVisual, Model: "clip-1", Vector: []float32{0, 0, 1}},
		{Subject: video("d"), Kind: schema.EmbeddingKindVisual, Model: "other", Vector: []float32{1, 0, 0}},
	} {
		if _, err := service.Add(ctx, stored); err != nil {
			t.Fatal(err)
		}
	}
	nearest, err := service.Nearest(ctx, Query{Kind: schema.EmbeddingKindVisual, Model: "clip-1", Vector: []float32{1, 0, 0}, Limit: 2})
	if err != nil || len(nearest) != 2 || nearest[0].Subject.Id != "a" || nearest[1].Subject.Id != "b" {
		t.Fatalf("nearest = %+v, err %v", nearest, err)
	}
	similar, err := service.SimilarTo(ctx, video("a"), schema.EmbeddingKindVisual, "clip-1", 1)
	if err != nil || len(similar) != 1 || similar[0].Subject.Id != "b" {
		t.Fatalf("similar = %+v, err %v", similar, err)
	}
	if _, comparable := Cosine([]float32{1, 0}, []float32{1, 0, 0}); comparable {
		t.Error("vectors of different lengths compared")
	}
}
