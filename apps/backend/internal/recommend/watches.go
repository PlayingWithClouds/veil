package recommend

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// watch is one playback session of one video, folded from its view,
// progress, finish and heatSpan events.
type watch struct {
	videoID   string
	startedAt time.Time
	lastAt    time.Time
	// playedSeconds is how long it played (the heatSpans' total); stopSeconds
	// the furthest position reported.
	playedSeconds   float64
	stopSeconds     float64
	finished        bool
	durationSeconds float64
}

// watchQuery folds the playback events since $since into one row per video
// and session; events without a session are grouped by day.
const watchQuery = `SELECT "subjectId", COALESCE("sessionId", substr("at", 1, 10)) AS "session",
	MIN("at") AS "startedAt", MAX("at") AS "lastAt",
	SUM(CASE WHEN "kind" = 'heatSpan'
		THEN MAX(0, json_extract("data", '` + pathEnd + `') - json_extract("data", '` + pathStart + `')) ELSE 0 END) AS "played",
	MAX(CASE WHEN "kind" IN ('progress', 'finish') THEN json_extract("data", '` + pathPosition + `') END) AS "stop",
	MAX("kind" = 'finish') AS "finished",
	(SELECT "durationSeconds" FROM "Video" WHERE "Video"."id" = "Event"."subjectId") AS "duration"
	FROM "Event"
	WHERE "kind" IN ('view', 'progress', 'finish', 'heatSpan') AND "subjectKind" = 'video'
	  AND "at" >= $since AND "deletedAt" IS NULL
	GROUP BY "subjectId", "session" ORDER BY "lastAt" DESC LIMIT $limit`

// watchSignals turns the window's watches into session-ending, continued or
// abandoned watch signals.
func (e *Engine) watchSignals(ctx context.Context, collected *signals, since string, now time.Time) error {
	rows, err := e.database.Query(ctx, watchQuery, db.Vars{"since": since, "limit": eventReadLimit})
	if err != nil {
		return fmt.Errorf("read watches: %w", err)
	}
	watches := make([]watch, 0, len(rows))
	for _, row := range rows {
		watches = append(watches, watchFromRow(row))
	}
	collected.watches = watches
	followed := followedWatches(watches)
	for index, entry := range watches {
		sinceUpdate := now.Sub(entry.lastAt)
		weight, counts := watchWeight(entry, followed[index], sinceUpdate)
		if !counts {
			continue
		}
		collected.items = append(collected.items, signal{
			item:        schema.RecordRef{Kind: schema.RecordKindVideo, Id: entry.videoID},
			weight:      weight,
			at:          entry.lastAt,
			sessionEnd:  endsSession(entry, followed[index], sinceUpdate),
			stopSeconds: entry.stopSeconds,
		})
	}
	return nil
}

// watchFromRow decodes one watchQuery row. A client that sends no heatSpans
// still gets a played time: the position reached, or the runtime when it
// finished.
func watchFromRow(row db.Row) watch {
	entry := watch{
		videoID:         rowString(row, "subjectId"),
		startedAt:       rowTime(row, "startedAt"),
		lastAt:          rowTime(row, "lastAt"),
		playedSeconds:   rowFloat(row, "played"),
		stopSeconds:     rowFloat(row, "stop"),
		finished:        rowInt(row, "finished") == 1,
		durationSeconds: rowFloat(row, "duration"),
	}
	if entry.playedSeconds > 0 {
		return entry
	}
	entry.playedSeconds = entry.stopSeconds
	if entry.finished && entry.durationSeconds > 0 {
		entry.playedSeconds = entry.durationSeconds
	}
	return entry
}

// followedWatches reports, per watch, whether another video played within
// sessionGap after it.
func followedWatches(watches []watch) []bool {
	order := make([]int, len(watches))
	for index := range order {
		order[index] = index
	}
	sort.Slice(order, func(left, right int) bool {
		return watches[order[left]].startedAt.Before(watches[order[right]].startedAt)
	})
	followed := make([]bool, len(watches))
	for _, index := range order {
		followed[index] = followedBy(watches[index], watches, order)
	}
	return followed
}

// followedBy reports whether a watch of another video started before
// sessionGap after entry and was still playing after entry stopped. order
// lists the watches by start.
func followedBy(entry watch, watches []watch, order []int) bool {
	deadline := entry.lastAt.Add(sessionGap)
	for _, index := range order {
		other := watches[index]
		if other.startedAt.After(deadline) {
			return false
		}
		if other.videoID != entry.videoID && other.lastAt.After(entry.lastAt) {
			return true
		}
	}
	return false
}

// watchWeight classifies one watch. A watch that ended the session is the
// strongest watch signal, finished or not: the video did its job. A watch
// another video followed counts for little, more when watched in full.
// counts is false for a fresh playback that may still be starting.
func watchWeight(entry watch, followed bool, sinceUpdate time.Duration) (weight float64, counts bool) {
	if isEarlyAbandon(entry.playedSeconds, entry.durationSeconds) {
		if sinceUpdate < abandonGrace {
			return 0, false
		}
		return weightEarlyAbandon, true
	}
	if endsSession(entry, followed, sinceUpdate) {
		return weightSessionEnd, true
	}
	if watchedInFull(entry.playedSeconds, entry.durationSeconds) {
		return weightCompletedWatch, true
	}
	return weightPartialWatch, true
}

// endsSession reports whether the watch was the last of its session: nothing
// else played within sessionGap after it and it played long enough to be
// more than a look.
func endsSession(entry watch, followed bool, sinceUpdate time.Duration) bool {
	if isEarlyAbandon(entry.playedSeconds, entry.durationSeconds) {
		return false
	}
	return !followed && sinceUpdate >= sessionGap && entry.playedSeconds >= sessionEndMinimumSeconds
}

// watchedFraction is the share of the runtime played; 0 when the runtime is
// unknown.
func watchedFraction(playedSeconds, durationSeconds float64) float64 {
	if durationSeconds <= 0 {
		return 0
	}
	return playedSeconds / durationSeconds
}

// watchedInFull reports whether a video played fullWatchSeconds or
// completedFraction of its runtime.
func watchedInFull(playedSeconds, durationSeconds float64) bool {
	if playedSeconds >= fullWatchSeconds {
		return true
	}
	return watchedFraction(playedSeconds, durationSeconds) >= completedFraction
}

// isEarlyAbandon reports whether a video played under earlyAbandonSeconds, or
// under earlyAbandonFraction of its runtime while also under
// earlyAbandonCeiling: a minute spent on a long video is a look, not a reject.
func isEarlyAbandon(playedSeconds, durationSeconds float64) bool {
	if playedSeconds < earlyAbandonSeconds {
		return true
	}
	if playedSeconds >= earlyAbandonCeiling {
		return false
	}
	return durationSeconds > 0 && watchedFraction(playedSeconds, durationSeconds) < earlyAbandonFraction
}

// watchState is how far the user got through an item.
type watchState int

const (
	watchNone watchState = iota
	watchPartial
	watchCompleted
)

// loadWatchStates classifies every viewed item: completed when a watch of the
// window played it in full, else partial when it was ever viewed.
func (e *Engine) loadWatchStates(ctx context.Context, watches []watch) (map[schema.RecordRef]watchState, error) {
	rows, err := e.database.Query(ctx,
		`SELECT "subjectKind", "subjectId" FROM "UserState"
		 WHERE "viewCount" > 0 AND "deletedAt" IS NULL AND "subjectKind" IN (SELECT value FROM json_each($mediaKinds))`,
		db.Vars{"mediaKinds": mediaKinds})
	if err != nil {
		return nil, fmt.Errorf("read watch states: %w", err)
	}
	out := make(map[schema.RecordRef]watchState, len(rows)+len(watches))
	for _, row := range rows {
		out[rowRef(row, "subjectKind", "subjectId")] = watchPartial
	}
	for _, entry := range watches {
		ref := schema.RecordRef{Kind: schema.RecordKindVideo, Id: entry.videoID}
		if entry.finished || watchedInFull(entry.playedSeconds, entry.durationSeconds) {
			out[ref] = watchCompleted
			continue
		}
		if out[ref] == watchNone {
			out[ref] = watchPartial
		}
	}
	return out, nil
}
