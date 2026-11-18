package recommend

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/playingwithclouds/veil/internal/db"
)

// Scene tags are what the phone's scene model saw in a video: the tags of its
// ai segments, each weighted by how much of the runtime it covers, so a video
// that is mostly one act counts as that act more than one with a few seconds
// of it. They join the site tags in the taste profile and the affinity match.

// loadSceneTags fills in the scene tags of the videos.
func (e *Engine) loadSceneTags(ctx context.Context, videos []*mediaFeatures) error {
	if len(videos) == 0 {
		return nil
	}
	byID := make(map[string]*mediaFeatures, len(videos))
	ids := make([]string, 0, len(videos))
	for _, video := range videos {
		byID[video.ref.Id] = video
		ids = append(ids, video.ref.Id)
	}
	rows, err := e.database.Query(ctx,
		`SELECT "subjectId", "tag", "startSeconds", "endSeconds" FROM "Segment"
		 WHERE "subjectKind" = 'video' AND "subjectId" IN (SELECT value FROM json_each($ids))
		   AND "origin" = 'ai' AND "tag" IS NOT NULL AND "endSeconds" IS NOT NULL AND "deletedAt" IS NULL`,
		db.Vars{"ids": ids})
	if err != nil {
		return fmt.Errorf("read scene tags: %w", err)
	}
	stretches := map[string][]taggedSegment{}
	for _, row := range rows {
		videoID := rowString(row, "subjectId")
		stretch := span{start: rowFloat(row, "startSeconds"), end: rowFloat(row, "endSeconds")}
		stretches[videoID] = append(stretches[videoID], taggedSegment{span: stretch, tagID: rowString(row, "tag")})
	}
	for videoID, segments := range stretches {
		video := byID[videoID]
		video.sceneTags = sceneTagShares(segments, float64(video.durationSeconds))
	}
	return nil
}

// sceneTagShares is the share of the runtime each tag covers (overlaps of
// one tag counted once), tags under sceneTagMinimumShare left out. Without a
// known runtime the last segment's end stands in for it.
func sceneTagShares(segments []taggedSegment, durationSeconds float64) map[string]float64 {
	runtime := durationSeconds
	if runtime <= 0 {
		for _, segment := range segments {
			runtime = math.Max(runtime, segment.end)
		}
	}
	if runtime <= 0 {
		return nil
	}
	byTag := map[string][]span{}
	for _, segment := range segments {
		byTag[segment.tagID] = append(byTag[segment.tagID], segment.span)
	}
	shares := map[string]float64{}
	for tagID, spans := range byTag {
		share := math.Min(1, coveredSeconds(spans)/runtime)
		if share >= sceneTagMinimumShare {
			shares[tagID] = share
		}
	}
	if len(shares) == 0 {
		return nil
	}
	return shares
}

// coveredSeconds is how much of the timeline the spans cover, overlaps
// counted once.
func coveredSeconds(spans []span) float64 {
	ordered := append([]span(nil), spans...)
	sort.Slice(ordered, func(left, right int) bool { return ordered[left].start < ordered[right].start })
	total, reached := 0.0, math.Inf(-1)
	for _, stretch := range ordered {
		start := math.Max(stretch.start, reached)
		if stretch.end > start {
			total += stretch.end - start
		}
		reached = math.Max(reached, stretch.end)
	}
	return total
}

// addSceneTags gives a signalled video's scene tags a share of its weight,
// by how much of it each covers; tags the site gave it too already got theirs.
func (taste *profile) addSceneTags(item *mediaFeatures, weight float64) {
	siteTags := toSet(item.tags)
	for tagID, share := range item.sceneTags {
		if siteTags[tagID] {
			continue
		}
		taste.tags[tagID] += sceneTagWeight * weight * share * taste.rarity.weight(tagID)
	}
}

// sceneAffinity is the share-weighted mean affinity of a video's scene tags,
// each weighted by its rarity like the site tags; a blocked tag keeps its full
// weight.
func sceneAffinity(sceneTags map[string]float64, affinities map[string]float64, rarity *tagRarity) float64 {
	sum, shares := 0.0, 0.0
	for tagID, share := range sceneTags {
		affinity := affinities[tagID]
		if affinity > affinityBlocked {
			affinity *= rarity.weight(tagID)
		}
		sum += share * affinity
		shares += share
	}
	if shares == 0 {
		return 0
	}
	return clampFloat(sum/shares, -1, 1)
}

// itemTagAffinity is the tag term of the affinity match: the site tags, mixed
// with the scene tags when the video has any.
func itemTagAffinity(item *mediaFeatures, taste *profile) float64 {
	siteTerm := tagAffinity(item.tags, taste.tags, taste.rarity)
	if len(item.sceneTags) == 0 {
		return siteTerm
	}
	sceneTerm := sceneAffinity(item.sceneTags, taste.tags, taste.rarity)
	if len(item.tags) == 0 {
		return sceneTerm
	}
	return (1-sceneMatchShare)*siteTerm + sceneMatchShare*sceneTerm
}
