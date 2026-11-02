package recommend

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// tagRarity knows how many tagged items carry each tag, so a tag on most of
// them (HD, amateur, ...) can count for less than a rare one. A nil rarity
// weighs every tag 1.
type tagRarity struct {
	taggedItems int
	itemCounts  map[string]int
}

// weight is the tag's weight in [rarityFloor, 1]: 1 for a tag on
// rarityFullShare of the tagged items or fewer, falling with the log of its
// share to the floor for a tag on all of them.
func (rarity *tagRarity) weight(tagID string) float64 {
	if rarity == nil || rarity.taggedItems < rarityMinItems {
		return 1
	}
	return shareWeight(rarity.itemCounts[tagID], rarity.taggedItems, rarityFullShare, rarityFloor)
}

// shareWeight turns "count of total carry it" into a weight: 1 up to
// fullShare, then falling with the log of the share down to floor at 100%.
func shareWeight(count, total int, fullShare, floor float64) float64 {
	if total <= 0 || count <= 0 {
		return 1
	}
	share := math.Min(1, float64(count)/float64(total))
	weight := math.Log(1/share) / math.Log(1/fullShare)
	return math.Max(floor, math.Min(1, weight))
}

// Queries behind tagRarity, over videos and galleries together.
const (
	taggedItemCountQuery = `SELECT
		(SELECT COUNT(*) FROM "Video" WHERE "deletedAt" IS NULL AND json_array_length("tags") > 0) +
		(SELECT COUNT(*) FROM "Gallery" WHERE "deletedAt" IS NULL AND json_array_length("tags") > 0) AS "count"`
	tagItemCountQuery = `SELECT "tag", SUM("items") AS "items" FROM (
		SELECT "tagged".value AS "tag", COUNT(*) AS "items" FROM "Video", json_each("Video"."tags") AS "tagged"
		WHERE "Video"."deletedAt" IS NULL GROUP BY "tagged".value
		UNION ALL
		SELECT "tagged".value, COUNT(*) FROM "Gallery", json_each("Gallery"."tags") AS "tagged"
		WHERE "Gallery"."deletedAt" IS NULL GROUP BY "tagged".value
	) GROUP BY "tag"`
)

// loadTagRarity counts the items carrying each tag. It scans both media
// tables, which is why libraryStats caches it.
func (e *Engine) loadTagRarity(ctx context.Context) (*tagRarity, error) {
	taggedItems, err := e.database.Int(ctx, taggedItemCountQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("count tagged items: %w", err)
	}
	rows, err := e.database.Query(ctx, tagItemCountQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("count items per tag: %w", err)
	}
	rarity := &tagRarity{taggedItems: taggedItems, itemCounts: make(map[string]int, len(rows))}
	for _, row := range rows {
		rarity.itemCounts[rowString(row, "tag")] = rowInt(row, "items")
	}
	return rarity, nil
}

// libraryStats are the library-wide aggregates ranking reads: slow to change
// and costing a scan of the library, so they are reused for libraryStatsTTL.
type libraryStats struct {
	loadedAt   time.Time
	rarity     *tagRarity
	attributes *performerAttributes
}

// statsCache memoizes the library stats.
type statsCache struct {
	mutex sync.Mutex
	stats *libraryStats
}

// libraryStats returns the cached aggregates, reloading them once older than
// libraryStatsTTL.
func (e *Engine) libraryStats(ctx context.Context) (*libraryStats, error) {
	e.stats.mutex.Lock()
	defer e.stats.mutex.Unlock()
	now := e.now()
	if e.stats.stats != nil && now.Sub(e.stats.stats.loadedAt) < libraryStatsTTL {
		return e.stats.stats, nil
	}
	rarity, err := e.loadTagRarity(ctx)
	if err != nil {
		return nil, err
	}
	attributes, err := e.loadPerformerAttributes(ctx)
	if err != nil {
		return nil, err
	}
	e.stats.stats = &libraryStats{loadedAt: now, rarity: rarity, attributes: attributes}
	return e.stats.stats, nil
}
