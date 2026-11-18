package recommend

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Cover tags are what the phone's scene model reads off a card's cover: the
// labels it is fairly sure of, as tag names with their probability, sent with
// the impression and choice events (data.coverTags). Comparing how often
// covers showing a label were picked against how often they were shown tells
// which acts and positions draw the user in, before a video is even watched.

// coverTagQuery reads the cover tags of the impressions and choices since
// $since, newest first.
const coverTagQuery = `SELECT "kind", json_extract("data", '$.coverTags') AS "coverTags" FROM "Event"
	WHERE "kind" IN ('impression', 'choice') AND "at" >= $since AND "deletedAt" IS NULL
	  AND json_extract("data", '$.coverTags') IS NOT NULL
	ORDER BY "at" DESC LIMIT $limit`

// coverTally is the probability mass of a tag over the shown and the picked
// covers.
type coverTally struct {
	shown  float64
	chosen float64
}

// loadCoverTagLifts reads the cover tags of the last ignoredImpressionWindow
// and returns each known tag's click lift by tag id.
func (e *Engine) loadCoverTagLifts(ctx context.Context, now time.Time) (map[string]float64, error) {
	rows, err := e.database.Query(ctx, coverTagQuery, db.Vars{"since": db.FormatTime(now.Add(-ignoredImpressionWindow)), "limit": eventReadLimit})
	if err != nil {
		return nil, fmt.Errorf("read cover tags: %w", err)
	}
	tallies := map[string]*coverTally{}
	shownCount, chosenCount := 0, 0
	for _, row := range rows {
		chosen := rowString(row, "kind") == string(schema.EventKindChoice)
		if chosen {
			chosenCount++
		} else {
			shownCount++
		}
		tallyCoverTags(tallies, decodeCoverTags(rowString(row, "coverTags")), chosen)
	}
	lifts := coverTagLifts(tallies, shownCount, chosenCount)
	return e.liftsByTagID(ctx, lifts)
}

// decodeCoverTags parses one event's coverTags (tag name → probability);
// malformed data is empty.
func decodeCoverTags(text string) map[string]float64 {
	tags := map[string]float64{}
	if err := json.Unmarshal([]byte(text), &tags); err != nil {
		return map[string]float64{}
	}
	return tags
}

// tallyCoverTags adds one cover's tags to the tallies, keyed by folded name.
func tallyCoverTags(tallies map[string]*coverTally, tags map[string]float64, chosen bool) {
	for name, probability := range tags {
		key := names.Key(name)
		if key == "" || probability <= 0 {
			continue
		}
		tally := tallies[key]
		if tally == nil {
			tally = &coverTally{}
			tallies[key] = tally
		}
		if chosen {
			tally.chosen += probability
		} else {
			tally.shown += probability
		}
	}
}

// coverTagLifts is, per tag shown often enough, how much likelier a cover
// with it was picked than covers overall, in [-1, 1]: its pick rate pulled
// toward the overall rate by coverTagPrior showings, relative to that rate.
func coverTagLifts(tallies map[string]*coverTally, shownCount, chosenCount int) map[string]float64 {
	lifts := map[string]float64{}
	if shownCount == 0 || chosenCount == 0 {
		return lifts
	}
	overall := float64(chosenCount) / float64(shownCount)
	for key, tally := range tallies {
		if tally.shown < coverTagMinimumShown {
			continue
		}
		rate := (tally.chosen + coverTagPrior*overall) / (tally.shown + coverTagPrior)
		lifts[key] = clampFloat(rate/overall-1, -1, 1)
	}
	return lifts
}

// liftsByTagID maps lifts by folded tag name onto the live tags with those
// names; names no tag has are dropped.
func (e *Engine) liftsByTagID(ctx context.Context, lifts map[string]float64) (map[string]float64, error) {
	out := map[string]float64{}
	if len(lifts) == 0 {
		return out, nil
	}
	keys := setKeys(boolSet(lifts))
	rows, err := e.database.Query(ctx,
		`SELECT "id", "nameKey" FROM "Tag" WHERE "nameKey" IN (SELECT value FROM json_each($keys)) AND "deletedAt" IS NULL`,
		db.Vars{"keys": keys})
	if err != nil {
		return nil, fmt.Errorf("read cover tag ids: %w", err)
	}
	for _, row := range rows {
		out[rowString(row, "id")] = lifts[rowString(row, "nameKey")]
	}
	return out, nil
}

// boolSet is the key set of a map.
func boolSet(values map[string]float64) map[string]bool {
	out := make(map[string]bool, len(values))
	for key := range values {
		out[key] = true
	}
	return out
}

// addCoverTags adds the cover tags' click lifts to the tag affinities.
func (taste *profile) addCoverTags(lifts map[string]float64) {
	for tagID, lift := range lifts {
		taste.tags[tagID] += coverTagWeight * lift
	}
}
