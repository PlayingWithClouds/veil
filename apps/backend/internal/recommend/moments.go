package recommend

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// span is a stretch of a video's timeline in seconds.
type span struct {
	start float64
	end   float64
}

// overlaps reports whether the two spans touch.
func (window span) overlaps(other span) bool {
	return window.start <= other.end && window.end >= other.start
}

// maxHeatSpanSeconds drops spans longer than any plausible playback, so one
// malformed event cannot paint a whole video hot.
const maxHeatSpanSeconds = 6 * 3600

// heatQuery reads the heatSpans and o positions of the videos in $ids.
const heatQuery = `SELECT "subjectId", "kind",
	json_extract("data", '` + pathStart + `') AS "start", json_extract("data", '` + pathEnd + `') AS "end",
	json_extract("data", '` + pathPosition + `') AS "position"
	FROM "Event"
	WHERE "subjectKind" = 'video' AND "subjectId" IN (SELECT value FROM json_each($ids))
	  AND "kind" IN ('heatSpan', 'o') AND "deletedAt" IS NULL
	ORDER BY "at" DESC LIMIT $limit`

// loadHeat builds each video's replay heat: per heatBucketSeconds bucket, how
// often it played, plus heatOWeight before every o position.
func (e *Engine) loadHeat(ctx context.Context, videoIDs []string) (map[string][]float64, error) {
	out := map[string][]float64{}
	if len(videoIDs) == 0 {
		return out, nil
	}
	rows, err := e.database.Query(ctx, heatQuery, db.Vars{"ids": videoIDs, "limit": heatSpanReadLimit * len(videoIDs)})
	if err != nil {
		return nil, fmt.Errorf("read replay heat: %w", err)
	}
	for _, row := range rows {
		videoID := rowString(row, "subjectId")
		out[videoID] = addHeat(out[videoID], heatEventSpan(row))
	}
	return out, nil
}

// heatEventSpan is the stretch one heat row warms and by how much: the span
// played, or the lead-in to an o. A zero weight warms nothing.
func heatEventSpan(row db.Row) weightedSpan {
	if rowString(row, "kind") == string(schema.EventKindO) {
		if row["position"] == nil {
			return weightedSpan{}
		}
		position := rowFloat(row, "position")
		return weightedSpan{span: span{start: math.Max(0, position-heatOLeadSeconds), end: position}, weight: heatOWeight}
	}
	played := span{start: rowFloat(row, "start"), end: rowFloat(row, "end")}
	if played.start < 0 || played.end <= played.start || played.end-played.start > maxHeatSpanSeconds {
		return weightedSpan{}
	}
	return weightedSpan{span: played, weight: 1}
}

// weightedSpan is a span that adds weight to every bucket it touches.
type weightedSpan struct {
	span
	weight float64
}

// addHeat adds a weighted span to a heat array, growing it as needed.
func addHeat(buckets []float64, warmed weightedSpan) []float64 {
	if warmed.weight == 0 {
		return buckets
	}
	first := int(warmed.start / heatBucketSeconds)
	last := int(math.Ceil(warmed.end/heatBucketSeconds)) - 1
	last = max(last, first)
	for len(buckets) <= last {
		buckets = append(buckets, 0)
	}
	for bucket := first; bucket <= last; bucket++ {
		buckets[bucket] += warmed.weight
	}
	return buckets
}

// heatPeak is the most replayed stretch: the busiest bucket widened to its
// neighbours of at least heatRegionShare of it. Not found unless the busiest
// bucket played heatMinimumReplays times and heatPeakRatio times the mean
// (over the timeline up to the last warm bucket).
func heatPeak(buckets []float64) (span, bool) {
	if len(buckets) == 0 {
		return span{}, false
	}
	total, busiest := 0.0, 0
	for bucket, value := range buckets {
		total += value
		if value > buckets[busiest] {
			busiest = bucket
		}
	}
	peak := buckets[busiest]
	if peak < heatMinimumReplays || peak < heatPeakRatio*total/float64(len(buckets)) {
		return span{}, false
	}
	left, right := busiest, busiest
	for left > 0 && buckets[left-1] >= heatRegionShare*peak {
		left--
	}
	for right < len(buckets)-1 && buckets[right+1] >= heatRegionShare*peak {
		right++
	}
	return span{start: float64(left) * heatBucketSeconds, end: float64(right+1) * heatBucketSeconds}, true
}

// shortWindow fits a moment into a short: at most shortMaxSeconds from its
// start, at least shortMinSeconds, and inside the runtime when it is known.
func shortWindow(moment span, durationSeconds float64) span {
	end := math.Min(moment.end, moment.start+shortMaxSeconds)
	end = math.Max(end, moment.start+shortMinSeconds)
	if durationSeconds > 0 {
		end = math.Min(end, durationSeconds)
	}
	return span{start: moment.start, end: end}
}

// BestMoment is where a video is worth starting: its most replayed point,
// else the start of its first tagged segment (any segment when none is
// tagged). ok is false when there is neither.
func (e *Engine) BestMoment(ctx context.Context, videoID string) (seconds float64, ok bool, err error) {
	heat, err := e.loadHeat(ctx, []string{videoID})
	if err != nil {
		return 0, false, err
	}
	if peak, found := heatPeak(heat[videoID]); found {
		return peak.start, true, nil
	}
	row, err := e.database.QueryRow(ctx,
		`SELECT "startSeconds" FROM "Segment"
		 WHERE "subjectKind" = 'video' AND "subjectId" = $id AND "deletedAt" IS NULL
		 ORDER BY "tag" IS NULL, "startSeconds" LIMIT 1`,
		db.Vars{"id": videoID})
	if err != nil || row == nil {
		return 0, false, err
	}
	return rowFloat(row, "startSeconds"), true, nil
}

// sessionEnder is a video whose watch ended a session, with where playback
// was left.
type sessionEnder struct {
	videoID     string
	stopSeconds float64
}

// sessionEnders lists the videos with a session-ending watch, the most recent
// momentSceneLimit of them.
func sessionEnders(items []signal) []sessionEnder {
	latest := map[string]signal{}
	for _, entry := range items {
		if !entry.sessionEnd {
			continue
		}
		if previous, seen := latest[entry.item.Id]; !seen || entry.at.After(previous.at) {
			latest[entry.item.Id] = entry
		}
	}
	entries := make([]signal, 0, len(latest))
	for _, entry := range latest {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(left, right int) bool {
		return entries[left].at.After(entries[right].at)
	})
	if len(entries) > momentSceneLimit {
		entries = entries[:momentSceneLimit]
	}
	out := make([]sessionEnder, 0, len(entries))
	for _, entry := range entries {
		out = append(out, sessionEnder{videoID: entry.item.Id, stopSeconds: entry.stopSeconds})
	}
	return out
}

// taggedSegment is a tagged stretch (or point) of a video.
type taggedSegment struct {
	span
	tagID string
}

// loadMoments reads the tagged segments and replay heat of the session-ending
// videos and returns, per video, the tags of the segments near where playback
// stopped and near the replay peak, sorted.
func (e *Engine) loadMoments(ctx context.Context, enders []sessionEnder) (map[schema.RecordRef][]string, error) {
	out := map[schema.RecordRef][]string{}
	if len(enders) == 0 {
		return out, nil
	}
	ids := make([]string, 0, len(enders))
	for _, ender := range enders {
		ids = append(ids, ender.videoID)
	}
	segments, err := e.loadTaggedSegments(ctx, ids)
	if err != nil {
		return nil, err
	}
	heat, err := e.loadHeat(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, ender := range enders {
		tags := taggedWithin(segments[ender.videoID], momentWindows(ender, heat[ender.videoID]))
		if len(tags) > 0 {
			out[schema.RecordRef{Kind: schema.RecordKindVideo, Id: ender.videoID}] = tags
		}
	}
	return out, nil
}

// loadTaggedSegments reads the tagged segments of the videos, by video id. A
// segment without an end is a point.
func (e *Engine) loadTaggedSegments(ctx context.Context, videoIDs []string) (map[string][]taggedSegment, error) {
	rows, err := e.database.Query(ctx,
		`SELECT "subjectId", "tag", "startSeconds", "endSeconds" FROM "Segment"
		 WHERE "subjectKind" = 'video' AND "subjectId" IN (SELECT value FROM json_each($ids))
		   AND "tag" IS NOT NULL AND "deletedAt" IS NULL`,
		db.Vars{"ids": videoIDs})
	if err != nil {
		return nil, fmt.Errorf("read moment segments: %w", err)
	}
	out := map[string][]taggedSegment{}
	for _, row := range rows {
		start := rowFloat(row, "startSeconds")
		end := start
		if row["endSeconds"] != nil {
			end = rowFloat(row, "endSeconds")
		}
		videoID := rowString(row, "subjectId")
		out[videoID] = append(out[videoID], taggedSegment{span: span{start: start, end: end}, tagID: rowString(row, "tag")})
	}
	return out, nil
}

// momentWindows are the stretches of a video that matter: the last minutes
// before the stop point and the replay peak.
func momentWindows(ender sessionEnder, heat []float64) []span {
	windows := []span{}
	if ender.stopSeconds > 0 {
		windows = append(windows, span{start: ender.stopSeconds - momentSecondsBefore, end: ender.stopSeconds + momentSecondsAfter})
	}
	if peak, found := heatPeak(heat); found {
		windows = append(windows, peak)
	}
	return windows
}

// taggedWithin collects the distinct tags of the segments touching any
// window, sorted.
func taggedWithin(segments []taggedSegment, windows []span) []string {
	found := map[string]bool{}
	for _, segment := range segments {
		if touchesAny(segment.span, windows) {
			found[segment.tagID] = true
		}
	}
	return setKeys(found)
}

// touchesAny reports whether stretch overlaps any of windows.
func touchesAny(stretch span, windows []span) bool {
	for _, window := range windows {
		if stretch.overlaps(window) {
			return true
		}
	}
	return false
}
