package recommend

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"sort"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// similarityKinds are the embedding kinds that say two media look or read
// alike. Face vectors identify people, which shared credits already cover.
var similarityKinds = []string{string(schema.EmbeddingKindVisual), string(schema.EmbeddingKindText)}

// modelKey identifies vectors that can be compared: one kind from one model.
type modelKey struct {
	kind  schema.EmbeddingKind
	model string
}

// probe is a vector to search with and the item it belongs to.
type probe struct {
	owner  schema.RecordRef
	vector []float32
}

// similarItem is an item close to one of the probes.
type similarItem struct {
	item       schema.RecordRef
	similarity float64
	// owner is the probe's item it is closest to.
	owner schema.RecordRef
	kind  schema.EmbeddingKind
}

// decodeVector reads little-endian float32s and scales them to unit length,
// so a dot product is the cosine similarity. nil when malformed or zero.
func decodeVector(raw []byte) []float32 {
	if len(raw) == 0 || len(raw)%4 != 0 {
		return nil
	}
	vector := make([]float32, len(raw)/4)
	norm := 0.0
	for index := range vector {
		vector[index] = math.Float32frombits(binary.LittleEndian.Uint32(raw[index*4:]))
		norm += float64(vector[index]) * float64(vector[index])
	}
	if norm == 0 || math.IsNaN(norm) || math.IsInf(norm, 0) {
		return nil
	}
	scale := float32(1 / math.Sqrt(norm))
	for index := range vector {
		vector[index] *= scale
	}
	return vector
}

// cosine is the dot product of two unit vectors; 0 when their sizes differ.
func cosine(first, second []float32) float64 {
	if len(first) != len(second) {
		return 0
	}
	sum := 0.0
	for index := range first {
		sum += float64(first[index]) * float64(second[index])
	}
	return sum
}

// loadProbes reads the visual and text vectors of the items, grouped by
// model. Embedding has no index on its subject, so this scans it once (the
// vectors themselves are only read for matching rows).
func (e *Engine) loadProbes(ctx context.Context, refs []schema.RecordRef) (map[modelKey][]probe, error) {
	out := map[modelKey][]probe{}
	if len(refs) == 0 {
		return out, nil
	}
	keys := make([]string, 0, len(refs))
	for _, ref := range refs {
		keys = append(keys, refKey(ref))
	}
	rows, err := e.database.Query(ctx,
		`SELECT "subjectKind", "subjectId", "kind", "model", "vector" FROM "Embedding"
		 WHERE "kind" IN (SELECT value FROM json_each($kinds)) AND "deletedAt" IS NULL
		   AND "subjectKind" || ':' || "subjectId" IN (SELECT value FROM json_each($keys))`,
		db.Vars{"kinds": similarityKinds, "keys": keys})
	if err != nil {
		return nil, fmt.Errorf("read item vectors: %w", err)
	}
	for _, row := range rows {
		vector := decodeVector(rowBytes(row, "vector"))
		if vector == nil {
			continue
		}
		key := modelKey{kind: schema.EmbeddingKind(rowString(row, "kind")), model: rowString(row, "model")}
		out[key] = append(out[key], probe{owner: rowRef(row, "subjectKind", "subjectId"), vector: vector})
	}
	return out, nil
}

// rowBytes reads a blob column; nil when NULL.
func rowBytes(row db.Row, column string) []byte {
	switch typed := row[column].(type) {
	case []byte:
		return typed
	case string:
		return []byte(typed)
	}
	return nil
}

// embeddingPageQuery reads one page of a model's media vectors by rowid.
const embeddingPageQuery = `SELECT rowid AS "position", "subjectKind", "subjectId", "vector" FROM "Embedding"
	WHERE "kind" = $kind AND "model" = $model AND "deletedAt" IS NULL AND rowid > $after
	  AND "subjectKind" IN (SELECT value FROM json_each($mediaKinds))
	ORDER BY rowid LIMIT $page`

// nearest finds the media whose vectors of one model are closest to any of
// the probes, best first, at most limit, skipping excluded items. It reads
// the model's vectors page by page (at most embeddingScanLimit), so memory
// stays bounded; an item with several vectors (frames) counts by its best.
func (e *Engine) nearest(ctx context.Context, key modelKey, probes []probe, excluded map[schema.RecordRef]bool, limit int) ([]similarItem, error) {
	best := map[schema.RecordRef]similarItem{}
	after := int64(0)
	for scanned := 0; scanned < embeddingScanLimit; scanned += embeddingPageSize {
		rows, err := e.database.Query(ctx, embeddingPageQuery, db.Vars{
			"kind": key.kind, "model": key.model, "after": after, "mediaKinds": mediaKinds, "page": embeddingPageSize,
		})
		if err != nil {
			return nil, fmt.Errorf("scan vectors: %w", err)
		}
		for _, row := range rows {
			scoreVector(best, row, key.kind, probes, excluded)
		}
		if len(rows) < embeddingPageSize {
			break
		}
		after = int64(rowInt(rows[len(rows)-1], "position"))
	}
	return bestSimilar(best, limit), nil
}

// scoreVector compares one stored vector with the probes and keeps it in
// best when it beats the item's previous best.
func scoreVector(best map[schema.RecordRef]similarItem, row db.Row, kind schema.EmbeddingKind, probes []probe, excluded map[schema.RecordRef]bool) {
	ref := rowRef(row, "subjectKind", "subjectId")
	if excluded[ref] {
		return
	}
	vector := decodeVector(rowBytes(row, "vector"))
	if vector == nil {
		return
	}
	for _, candidate := range probes {
		similarity := cosine(vector, candidate.vector)
		if similarity >= similarMinimumCosine && similarity > best[ref].similarity {
			best[ref] = similarItem{item: ref, similarity: similarity, owner: candidate.owner, kind: kind}
		}
	}
}

// bestSimilar lists the most similar items, best first, at most limit.
func bestSimilar(best map[schema.RecordRef]similarItem, limit int) []similarItem {
	out := make([]similarItem, 0, len(best))
	for _, item := range best {
		out = append(out, item)
	}
	sort.Slice(out, func(left, right int) bool {
		if out[left].similarity != out[right].similarity {
			return out[left].similarity > out[right].similarity
		}
		return out[left].item.Id < out[right].item.Id
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

// similarTo finds the items closest to the probes' items over every model
// they have vectors in, best first. The probes' own items are excluded.
func (e *Engine) similarTo(ctx context.Context, refs []schema.RecordRef, limit int) ([]similarItem, error) {
	probesByModel, err := e.loadProbes(ctx, refs)
	if err != nil {
		return nil, err
	}
	excluded := map[schema.RecordRef]bool{}
	for _, ref := range refs {
		excluded[ref] = true
	}
	merged := map[schema.RecordRef]similarItem{}
	for key, probes := range probesByModel {
		found, err := e.nearest(ctx, key, probes, excluded, limit)
		if err != nil {
			return nil, err
		}
		for _, item := range found {
			if item.similarity > merged[item.item].similarity {
				merged[item.item] = item
			}
		}
	}
	return bestSimilar(merged, limit), nil
}
