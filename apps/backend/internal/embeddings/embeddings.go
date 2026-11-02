// Package embeddings stores vectors for faces, frames and text, computed by
// clients or plugins (the backend runs no model), and finds the records
// nearest to a vector: "more like this", visual dedupe, semantic search with
// a client-embedded query, identifying a performer by an embedded face.
//
// Vectors are little-endian float32; vectors of different models never mix.
package embeddings

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// embeddingTable is the table vectors live in.
const embeddingTable = "Embedding"

// scanLimit bounds how many stored vectors one search compares: searches
// are brute force over SQLite, fine for a personal library.
const scanLimit = 50000

// Service stores and searches embeddings.
type Service struct {
	database *db.DB
}

// New returns the service on database.
func New(database *db.DB) *Service {
	return &Service{database: database}
}

// NewEmbedding is a vector to store for a record.
type NewEmbedding struct {
	Subject schema.RecordRef
	Kind    schema.EmbeddingKind
	Model   string
	Vector  []float32
	// AtSeconds is the frame time; Region the face box ({x, y, width, height}
	// as shares of the frame).
	AtSeconds *float64
	Region    map[string]any
}

// Add stores a vector.
func (s *Service) Add(ctx context.Context, embedding NewEmbedding) (*schema.Embedding, error) {
	if len(embedding.Vector) == 0 {
		return nil, fmt.Errorf("an embedding needs a vector")
	}
	if strings.TrimSpace(embedding.Model) == "" {
		return nil, fmt.Errorf("an embedding needs its model's name")
	}
	row := db.Vars{
		"subjectKind": embedding.Subject.Kind,
		"subjectId":   embedding.Subject.Id,
		"kind":        embedding.Kind,
		"model":       strings.TrimSpace(embedding.Model),
		"vector":      Encode(embedding.Vector),
	}
	if embedding.AtSeconds != nil {
		row["atSeconds"] = *embedding.AtSeconds
	}
	if embedding.Region != nil {
		row["region"] = embedding.Region
	}
	id, err := s.database.Insert(ctx, embeddingTable, row)
	if err != nil {
		return nil, err
	}
	return db.Get[schema.Embedding](ctx, s.database, embeddingTable, id)
}

// List returns a record's embeddings.
func (s *Service) List(ctx context.Context, subject schema.RecordRef) ([]schema.Embedding, error) {
	return db.QueryAs[schema.Embedding](ctx, s.database,
		`SELECT * FROM "Embedding" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL ORDER BY "atSeconds", "createdAt"`,
		db.Vars{"kind": subject.Kind, "id": subject.Id})
}

// Remove deletes one embedding; vectors are derived data, so they go for
// good.
func (s *Service) Remove(ctx context.Context, id string) error {
	return s.database.Delete(ctx, embeddingTable, id)
}

// Neighbour is a record near a searched vector, with its best similarity
// (cosine, -1..1) and the frame it was found at.
type Neighbour struct {
	Subject    schema.RecordRef
	Similarity float64
	AtSeconds  *float64
}

// Query is a nearest-neighbour search: vectors of one kind and model, of
// some record kinds (all when empty), leaving out one record.
type Query struct {
	Kind         schema.EmbeddingKind
	Model        string
	Vector       []float32
	SubjectKinds []schema.RecordKind
	Exclude      *schema.RecordRef
	Limit        int
}

// Nearest finds the records whose vectors are most similar to the query's,
// one entry per record.
func (s *Service) Nearest(ctx context.Context, query Query) ([]Neighbour, error) {
	rows, err := s.database.Query(ctx,
		`SELECT "subjectKind", "subjectId", "vector", "atSeconds" FROM "Embedding"
		 WHERE "kind" = $kind AND "model" = $model AND "deletedAt" IS NULL LIMIT $limit`,
		db.Vars{"kind": query.Kind, "model": query.Model, "limit": scanLimit})
	if err != nil {
		return nil, err
	}
	best := map[schema.RecordRef]Neighbour{}
	for _, row := range rows {
		neighbour, ok := scoreRow(row, query)
		if !ok {
			continue
		}
		if current, seen := best[neighbour.Subject]; !seen || neighbour.Similarity > current.Similarity {
			best[neighbour.Subject] = neighbour
		}
	}
	return topNeighbours(best, query.Limit), nil
}

// scoreRow compares one stored vector with the query; ok is false for rows
// the query leaves out.
func scoreRow(row db.Row, query Query) (Neighbour, bool) {
	subject := schema.RecordRef{Kind: schema.RecordKind(fmt.Sprint(row["subjectKind"])), Id: fmt.Sprint(row["subjectId"])}
	if query.Exclude != nil && subject == *query.Exclude {
		return Neighbour{}, false
	}
	if len(query.SubjectKinds) > 0 && !slices.Contains(query.SubjectKinds, subject.Kind) {
		return Neighbour{}, false
	}
	stored, isBytes := row["vector"].([]byte)
	if !isBytes {
		return Neighbour{}, false
	}
	similarity, comparable := Cosine(query.Vector, Decode(stored))
	if !comparable {
		return Neighbour{}, false
	}
	neighbour := Neighbour{Subject: subject, Similarity: similarity}
	if at, isNumber := row["atSeconds"].(float64); isNumber {
		neighbour.AtSeconds = &at
	}
	return neighbour, true
}

// topNeighbours sorts by similarity and keeps the first limit (20 when not
// positive).
func topNeighbours(best map[schema.RecordRef]Neighbour, limit int) []Neighbour {
	if limit <= 0 {
		limit = 20
	}
	neighbours := make([]Neighbour, 0, len(best))
	for _, neighbour := range best {
		neighbours = append(neighbours, neighbour)
	}
	sort.Slice(neighbours, func(first, second int) bool {
		return neighbours[first].Similarity > neighbours[second].Similarity
	})
	if len(neighbours) > limit {
		neighbours = neighbours[:limit]
	}
	return neighbours
}

// SimilarTo finds the records nearest to a record: each of its vectors of
// the kind and model is searched, and every neighbour keeps its best match.
func (s *Service) SimilarTo(ctx context.Context, subject schema.RecordRef, kind schema.EmbeddingKind, model string, limit int) ([]Neighbour, error) {
	own, err := s.List(ctx, subject)
	if err != nil {
		return nil, err
	}
	best := map[schema.RecordRef]Neighbour{}
	for _, embedding := range own {
		if embedding.Kind != kind || embedding.Model != model {
			continue
		}
		found, err := s.Nearest(ctx, Query{Kind: kind, Model: model, Vector: Decode(embedding.Vector), Exclude: &subject, Limit: limit})
		if err != nil {
			return nil, err
		}
		for _, neighbour := range found {
			if current, seen := best[neighbour.Subject]; !seen || neighbour.Similarity > current.Similarity {
				best[neighbour.Subject] = neighbour
			}
		}
	}
	return topNeighbours(best, limit), nil
}

// Encode packs a vector as little-endian float32.
func Encode(vector []float32) []byte {
	encoded := make([]byte, 4*len(vector))
	for index, value := range vector {
		binary.LittleEndian.PutUint32(encoded[4*index:], math.Float32bits(value))
	}
	return encoded
}

// Decode unpacks a little-endian float32 vector.
func Decode(encoded []byte) []float32 {
	vector := make([]float32, len(encoded)/4)
	for index := range vector {
		vector[index] = math.Float32frombits(binary.LittleEndian.Uint32(encoded[4*index:]))
	}
	return vector
}

// Cosine is the cosine similarity of two vectors; comparable is false when
// their lengths differ or either is all zeros.
func Cosine(first, second []float32) (float64, bool) {
	if len(first) != len(second) || len(first) == 0 {
		return 0, false
	}
	var dot, firstNorm, secondNorm float64
	for index := range first {
		dot += float64(first[index]) * float64(second[index])
		firstNorm += float64(first[index]) * float64(first[index])
		secondNorm += float64(second[index]) * float64(second[index])
	}
	if firstNorm == 0 || secondNorm == 0 {
		return 0, false
	}
	return dot / math.Sqrt(firstNorm*secondNorm), true
}
