package recommend

import (
	"context"
	"fmt"
	"sort"

	"github.com/playingwithclouds/veil/internal/db"
)

// sessionEnder is a scene whose watch ended a session, with where playback
// was left.
type sessionEnder struct {
	sceneID         string
	stopSeconds     int
	durationSeconds int
}

// marker is a tagged span (or point, end = start) of a scene.
type marker struct {
	sceneID string
	tagID   string
	start   float64
	end     float64
}

// span is a stretch of a scene's timeline in seconds.
type span struct {
	start float64
	end   float64
}

// overlaps reports whether the marker touches the span.
func (candidate marker) overlaps(window span) bool {
	return candidate.start <= window.end && candidate.end >= window.start
}

// sessionEnders lists the scenes with a session-ending watch signal, the
// most recent momentSceneLimit of them.
func sessionEnders(signals []signal) []sessionEnder {
	latest := map[string]signal{}
	for _, entry := range signals {
		if !entry.sessionEnd {
			continue
		}
		if previous, seen := latest[entry.sceneID]; !seen || entry.at.After(previous.at) {
			latest[entry.sceneID] = entry
		}
	}
	entries := make([]signal, 0, len(latest))
	for _, entry := range latest {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(left, right int) bool {
		if !entries[left].at.Equal(entries[right].at) {
			return entries[left].at.After(entries[right].at)
		}
		return entries[left].sceneID < entries[right].sceneID
	})
	if len(entries) > momentSceneLimit {
		entries = entries[:momentSceneLimit]
	}
	out := make([]sessionEnder, 0, len(entries))
	for _, entry := range entries {
		out = append(out, sessionEnder{sceneID: entry.sceneID, stopSeconds: entry.stopSeconds, durationSeconds: entry.durationSeconds})
	}
	return out
}

// Queries behind the moments: tagged markers and replay heat of the scenes.
const (
	momentMarkerQuery = `SELECT media, tag, seconds, end_seconds FROM scene_marker
		WHERE tag IS NOT NULL AND media IN (SELECT value FROM json_each($ids))`
	momentHeatQuery = `SELECT scene, bucket, watched, scrubbed FROM scene_heat
		WHERE scene IN (SELECT value FROM json_each($ids))`
)

// heatBucketCount is how many buckets a scene's replay heat is split into
// (internal/media's heatBuckets).
const heatBucketCount = 100

// loadMoments reads the markers and replay heat of the session-ending scenes
// and returns, per scene, the tags of its moments (see momentTags).
func (e *Engine) loadMoments(ctx context.Context, enders []sessionEnder) (map[string][]string, error) {
	if len(enders) == 0 {
		return map[string][]string{}, nil
	}
	ids := make([]string, 0, len(enders))
	for _, ender := range enders {
		ids = append(ids, ender.sceneID)
	}
	markerRows, err := e.database.Query(ctx, momentMarkerQuery, db.Vars{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("read moment markers: %w", err)
	}
	heatRows, err := e.database.Query(ctx, momentHeatQuery, db.Vars{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("read moment heat: %w", err)
	}
	return momentTags(enders, markersFromRows(markerRows), heatFromRows(heatRows)), nil
}

// markersFromRows decodes momentMarkerQuery rows; a marker without an end is a point.
func markersFromRows(rows []db.Row) []marker {
	out := make([]marker, 0, len(rows))
	for _, row := range rows {
		start := rowFloat(row, "seconds")
		end := start
		if row["end_seconds"] != nil {
			end = rowFloat(row, "end_seconds")
		}
		out = append(out, marker{sceneID: rowString(row, "media"), tagID: rowString(row, "tag"), start: start, end: end})
	}
	return out
}

// heatFromRows decodes momentHeatQuery rows into watched + scrubbed per bucket.
func heatFromRows(rows []db.Row) map[string][]float64 {
	out := map[string][]float64{}
	for _, row := range rows {
		bucket := rowInt(row, "bucket")
		if bucket < 0 || bucket >= heatBucketCount {
			continue
		}
		sceneID := rowString(row, "scene")
		if out[sceneID] == nil {
			out[sceneID] = make([]float64, heatBucketCount)
		}
		out[sceneID][bucket] += rowFloat(row, "watched") + rowFloat(row, "scrubbed")
	}
	return out
}

// momentTags finds, per session-ending scene, the tags of the markers near
// where playback stopped and near the replay-heat peak. The result is sorted
// and deduplicated; scenes with no such marker are absent.
func momentTags(enders []sessionEnder, markers []marker, heat map[string][]float64) map[string][]string {
	markersByScene := map[string][]marker{}
	for _, candidate := range markers {
		markersByScene[candidate.sceneID] = append(markersByScene[candidate.sceneID], candidate)
	}
	out := map[string][]string{}
	for _, ender := range enders {
		windows := momentWindows(ender, heat[ender.sceneID])
		tags := taggedWithin(markersByScene[ender.sceneID], windows)
		if len(tags) > 0 {
			out[ender.sceneID] = tags
		}
	}
	return out
}

// momentWindows are the stretches of a scene that matter: the last minutes
// before the stop point and the replay peak.
func momentWindows(ender sessionEnder, heat []float64) []span {
	windows := []span{}
	if ender.stopSeconds > 0 {
		stop := float64(ender.stopSeconds)
		windows = append(windows, span{start: stop - momentSecondsBefore, end: stop + momentSecondsAfter})
	}
	if peak, found := heatPeak(heat, ender.durationSeconds); found {
		windows = append(windows, peak)
	}
	return windows
}

// heatPeak is the span of the busiest heat bucket, widened by
// momentHeatMarginSeconds, when that bucket is at least momentHeatPeakRatio
// times the bucket average. Not found without heat or a runtime.
func heatPeak(heat []float64, durationSeconds int) (span, bool) {
	if len(heat) == 0 || durationSeconds <= 0 {
		return span{}, false
	}
	total, busiest := 0.0, 0
	for bucket, value := range heat {
		total += value
		if value > heat[busiest] {
			busiest = bucket
		}
	}
	if total <= 0 || heat[busiest] < momentHeatPeakRatio*total/float64(len(heat)) {
		return span{}, false
	}
	bucketSeconds := float64(durationSeconds) / float64(len(heat))
	return span{
		start: float64(busiest)*bucketSeconds - momentHeatMarginSeconds,
		end:   float64(busiest+1)*bucketSeconds + momentHeatMarginSeconds,
	}, true
}

// taggedWithin collects the distinct tags of the markers touching any window, sorted.
func taggedWithin(markers []marker, windows []span) []string {
	found := map[string]bool{}
	for _, candidate := range markers {
		for _, window := range windows {
			if candidate.overlaps(window) {
				found[candidate.tagID] = true
				break
			}
		}
	}
	tags := setKeys(found)
	sort.Strings(tags)
	return tags
}
