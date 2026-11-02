package recommend

import (
	"context"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Short is one entry of the shorts feed: a saved clip (item kind clip), a
// segment (kind segment), or a stretch of a video (kind video): its most
// replayed moment, or the whole of a short vertical video. StartSeconds and
// EndSeconds are the stretch to play in the video.
type Short struct {
	Match        schema.Match
	VideoID      string
	StartSeconds *float64
	EndSeconds   *float64
}

// shortCandidate is a short before ranking.
type shortCandidate struct {
	item    schema.RecordRef
	videoID string
	window  span
	base    float64
	detail  string
	title   string
	// tagID is a segment's tag, named in its detail once names are read.
	tagID string
	score float64
}

// Shorts builds the shorts feed: saved clips, segments, most replayed
// moments and short vertical videos of the videos matching filter (its sort
// is ignored: shorts are ranked by taste), best first and spread over
// videos. Nothing when filter.kinds leaves out videos.
func (e *Engine) Shorts(ctx context.Context, filter schema.Filter, limit int) ([]Short, error) {
	if limit <= 0 || !includesVideos(filter.Kinds) {
		return []Short{}, nil
	}
	candidates, err := e.shortCandidates(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err := e.scoreShorts(ctx, candidates); err != nil {
		return nil, err
	}
	if err := e.nameShortTags(ctx, candidates); err != nil {
		return nil, err
	}
	return shortsOf(spreadShorts(candidates, limit)), nil
}

// includesVideos reports whether a filter's kinds admit videos (none set
// admits every kind).
func includesVideos(kinds []schema.MediaKind) bool {
	return len(kinds) == 0 || slices.Contains(kinds, schema.MediaKindVideo)
}

// shortCandidates gathers every kind of short.
func (e *Engine) shortCandidates(ctx context.Context, filter schema.Filter) ([]*shortCandidate, error) {
	sources := []func(context.Context, schema.Filter) ([]*shortCandidate, error){
		e.clipShorts,
		e.segmentShorts,
		e.momentShorts,
		e.verticalShorts,
	}
	out := []*shortCandidate{}
	for _, source := range sources {
		found, err := source(ctx, filter)
		if err != nil {
			return nil, err
		}
		out = append(out, found...)
	}
	return out, nil
}

// filteredVideoRows runs a query over the videos matching filter: columns to
// select, a join of the caller's table, and the tail (ORDER BY, LIMIT
// $limit, bound to shortsReadLimit).
func (e *Engine) filteredVideoRows(ctx context.Context, filter schema.Filter, columns, join, condition, tail string, extra db.Vars) ([]db.Row, error) {
	conditions, vars := library.FilterConditions(schema.RecordKindVideo, filter)
	if condition != "" {
		conditions = append(conditions, condition)
	}
	vars["limit"] = shortsReadLimit
	for name, value := range extra {
		vars[name] = value
	}
	query := `SELECT ` + columns + library.MediaFrom(schema.RecordKindVideo) + join +
		` WHERE ` + strings.Join(conditions, " AND ") + ` ` + tail
	rows, err := e.database.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("read shorts: %w", err)
	}
	return rows, nil
}

// clipShorts are the clips the user saved, newest first.
func (e *Engine) clipShorts(ctx context.Context, filter schema.Filter) ([]*shortCandidate, error) {
	rows, err := e.filteredVideoRows(ctx, filter,
		`"clip"."id", "clip"."video", "clip"."startSeconds", "clip"."endSeconds", "clip"."title"`,
		` JOIN "Clip" AS "clip" ON "clip"."video" = "media"."id" AND "clip"."deletedAt" IS NULL`,
		"", `ORDER BY "clip"."createdAt" DESC LIMIT $limit`, nil)
	if err != nil {
		return nil, err
	}
	out := make([]*shortCandidate, 0, len(rows))
	for _, row := range rows {
		out = append(out, &shortCandidate{
			item:    schema.RecordRef{Kind: schema.RecordKindClip, Id: rowString(row, "id")},
			videoID: rowString(row, "video"),
			window:  span{start: rowFloat(row, "startSeconds"), end: rowFloat(row, "endSeconds")},
			base:    shortBaseClip,
			detail:  "Your clip",
			title:   rowString(row, "title"),
		})
	}
	return out, nil
}

// segmentShorts are the videos' segments, each cut to a short; a marker (no
// end) plays markerShortSeconds.
func (e *Engine) segmentShorts(ctx context.Context, filter schema.Filter) ([]*shortCandidate, error) {
	rows, err := e.filteredVideoRows(ctx, filter,
		`"segment"."id", "segment"."subjectId", "segment"."startSeconds", "segment"."endSeconds", "segment"."title", "segment"."tag", "media"."durationSeconds"`,
		` JOIN "Segment" AS "segment" ON "segment"."subjectKind" = 'video' AND "segment"."subjectId" = "media"."id" AND "segment"."deletedAt" IS NULL`,
		"", `ORDER BY "segment"."createdAt" DESC LIMIT $limit`, nil)
	if err != nil {
		return nil, err
	}
	out := make([]*shortCandidate, 0, len(rows))
	for _, row := range rows {
		start := rowFloat(row, "startSeconds")
		end := start + markerShortSeconds
		if row["endSeconds"] != nil {
			end = rowFloat(row, "endSeconds")
		}
		out = append(out, &shortCandidate{
			item:    schema.RecordRef{Kind: schema.RecordKindSegment, Id: rowString(row, "id")},
			videoID: rowString(row, "subjectId"),
			window:  shortWindow(span{start: start, end: end}, rowFloat(row, "durationSeconds")),
			base:    shortBaseSegment,
			detail:  segmentDetail(rowString(row, "title"), start),
			title:   rowString(row, "title"),
			tagID:   rowString(row, "tag"),
		})
	}
	return out, nil
}

// segmentDetail describes a segment by its title, else by where it starts.
func segmentDetail(title string, start float64) string {
	if title != "" {
		return "Chapter: " + title
	}
	return "Scene at " + formatClock(start)
}

// momentShorts are the most replayed moments of the most replayed videos.
func (e *Engine) momentShorts(ctx context.Context, filter schema.Filter) ([]*shortCandidate, error) {
	videoIDs, err := e.database.Strings(ctx,
		`SELECT "subjectId" FROM "Event"
		 WHERE "kind" IN ('heatSpan', 'o') AND "subjectKind" = 'video' AND "at" >= $since AND "deletedAt" IS NULL
		 GROUP BY "subjectId" HAVING COUNT(*) >= 2 ORDER BY COUNT(*) DESC LIMIT $limit`,
		db.Vars{"since": db.FormatTime(e.now().Add(-signalWindow)), "limit": momentVideoLimit})
	if err != nil {
		return nil, fmt.Errorf("read replayed videos: %w", err)
	}
	if len(videoIDs) == 0 {
		return nil, nil
	}
	rows, err := e.filteredVideoRows(ctx, filter, `"media"."id", "media"."durationSeconds"`, "",
		`"media"."id" IN (SELECT value FROM json_each($videoIds))`, `LIMIT $limit`, db.Vars{"videoIds": videoIDs})
	if err != nil {
		return nil, err
	}
	return e.peakShorts(ctx, rows)
}

// peakShorts finds the replay peak of each video row (id, durationSeconds).
func (e *Engine) peakShorts(ctx context.Context, rows []db.Row) ([]*shortCandidate, error) {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, rowString(row, "id"))
	}
	heat, err := e.loadHeat(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := []*shortCandidate{}
	for _, row := range rows {
		videoID := rowString(row, "id")
		peak, found := heatPeak(heat[videoID])
		if !found {
			continue
		}
		out = append(out, &shortCandidate{
			item:    schema.RecordRef{Kind: schema.RecordKindVideo, Id: videoID},
			videoID: videoID,
			window:  shortWindow(peak, rowFloat(row, "durationSeconds")),
			base:    shortBaseMoment,
			detail:  "Most replayed at " + formatClock(peak.start),
		})
	}
	return out, nil
}

// verticalShorts are short videos filmed upright, newest first.
func (e *Engine) verticalShorts(ctx context.Context, filter schema.Filter) ([]*shortCandidate, error) {
	rows, err := e.filteredVideoRows(ctx, filter, `"media"."id", "media"."durationSeconds"`, "",
		`"media"."height" > "media"."width" AND "media"."durationSeconds" > 0 AND "media"."durationSeconds" <= $verticalMax`,
		`ORDER BY "media"."createdAt" DESC LIMIT $limit`, db.Vars{"verticalMax": verticalMaxSeconds})
	if err != nil {
		return nil, err
	}
	out := make([]*shortCandidate, 0, len(rows))
	for _, row := range rows {
		videoID := rowString(row, "id")
		out = append(out, &shortCandidate{
			item:    schema.RecordRef{Kind: schema.RecordKindVideo, Id: videoID},
			videoID: videoID,
			window:  span{start: 0, end: rowFloat(row, "durationSeconds")},
			base:    shortBaseVertical,
			detail:  "Short vertical video",
		})
	}
	return out, nil
}

// scoreShorts scores each short by its kind, the taste match of its video,
// and whether it was skipped lately.
func (e *Engine) scoreShorts(ctx context.Context, candidates []*shortCandidate) error {
	if len(candidates) == 0 {
		return nil
	}
	taste, _, _, err := e.profileAt(ctx, e.now())
	if err != nil {
		return err
	}
	videos := make([]schema.RecordRef, 0, len(candidates))
	for _, candidate := range candidates {
		videos = append(videos, schema.RecordRef{Kind: schema.RecordKindVideo, Id: candidate.videoID})
	}
	features, err := e.loadFeatures(ctx, videos)
	if err != nil {
		return err
	}
	skipped, err := e.recentlySkipped(ctx)
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		candidate.score = shortScore(candidate, features, taste, skipped)
	}
	return nil
}

// shortScore is a short's kind base, plus the taste match of its video,
// minus the skip penalty.
func shortScore(candidate *shortCandidate, features map[schema.RecordRef]*mediaFeatures, taste *profile, skipped map[schema.RecordRef]bool) float64 {
	total := candidate.base
	if video, found := features[schema.RecordRef{Kind: schema.RecordKindVideo, Id: candidate.videoID}]; found {
		total += shortAffinityWeight * affinityMatch(video, taste)
	}
	if skipped[candidate.item] {
		total -= shortSkipPenalty
	}
	return total
}

// recentlySkipped lists the items skipped within shortSkipWindow.
func (e *Engine) recentlySkipped(ctx context.Context) (map[schema.RecordRef]bool, error) {
	rows, err := e.database.Query(ctx,
		`SELECT DISTINCT "subjectKind", "subjectId" FROM "Event"
		 WHERE "kind" = 'skip' AND "at" >= $since AND "deletedAt" IS NULL AND "subjectKind" IS NOT NULL
		 LIMIT $limit`,
		db.Vars{"since": db.FormatTime(e.now().Add(-shortSkipWindow)), "limit": eventReadLimit})
	if err != nil {
		return nil, fmt.Errorf("read skips: %w", err)
	}
	out := make(map[schema.RecordRef]bool, len(rows))
	for _, row := range rows {
		out[rowRef(row, "subjectKind", "subjectId")] = true
	}
	return out, nil
}

// nameShortTags describes untitled segments by their tag ("Marker: …").
func (e *Engine) nameShortTags(ctx context.Context, candidates []*shortCandidate) error {
	ids := map[string]bool{}
	for _, candidate := range candidates {
		if candidate.tagID != "" && candidate.title == "" {
			ids[candidate.tagID] = true
		}
	}
	names, err := e.entityNames(ctx, setKeys(ids))
	if err != nil {
		return err
	}
	for _, candidate := range candidates {
		if name := names[candidate.tagID]; name != "" && candidate.title == "" {
			candidate.detail = "Marker: " + name + " at " + formatClock(candidate.window.start)
		}
	}
	return nil
}

// spreadShorts orders the shorts best first, at most maxShortsPerVideo per
// video and never two of one video in a row while others are left.
func spreadShorts(candidates []*shortCandidate, limit int) []*shortCandidate {
	queue := slices.Clone(candidates)
	sort.SliceStable(queue, func(left, right int) bool { return queue[left].score > queue[right].score })
	perVideo := map[string]int{}
	out := []*shortCandidate{}
	for len(out) < limit && len(queue) > 0 {
		index := nextShort(queue, out, perVideo)
		if index < 0 {
			break
		}
		perVideo[queue[index].videoID]++
		out = append(out, queue[index])
		queue = slices.Delete(queue, index, index+1)
	}
	return out
}

// nextShort picks the best short under its video's cap, preferring one of a
// different video than the last; -1 when every video is capped.
func nextShort(queue, picked []*shortCandidate, perVideo map[string]int) int {
	lastVideo := ""
	if len(picked) > 0 {
		lastVideo = picked[len(picked)-1].videoID
	}
	fallback := -1
	for index, candidate := range queue {
		if perVideo[candidate.videoID] >= maxShortsPerVideo {
			continue
		}
		if candidate.videoID != lastVideo {
			return index
		}
		if fallback < 0 {
			fallback = index
		}
	}
	return fallback
}

// shortsOf turns ranked shorts into the feed's entries.
func shortsOf(candidates []*shortCandidate) []Short {
	out := make([]Short, 0, len(candidates))
	for _, candidate := range candidates {
		evidence := veilEvidence(candidate.detail)
		if candidate.title != "" {
			evidence.Title = stringPointer(candidate.title)
		}
		start, end := candidate.window.start, math.Max(candidate.window.end, candidate.window.start)
		out = append(out, Short{
			Match:        schema.Match{Item: candidate.item, Score: float32Pointer(normalizedScore(candidate.score)), Evidence: []schema.Evidence{evidence}},
			VideoID:      candidate.videoID,
			StartSeconds: &start,
			EndSeconds:   &end,
		})
	}
	return out
}
