package library

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/schema"
)

// AISegment is a stretch of a video a scene model labelled, its label given
// as a tag name ("Blowjob").
type AISegment struct {
	StartSeconds float64
	EndSeconds   float64
	Tag          string
}

// ReplaceAISegments swaps a video's or audio's ai segments for what a model
// found, finding or creating each tag by name. The user's and the plugins'
// segments stay; an empty list clears the ai ones.
func (l *Library) ReplaceAISegments(ctx context.Context, kind schema.RecordKind, id string, found []AISegment) ([]schema.Segment, error) {
	if kind != schema.RecordKindVideo && kind != schema.RecordKindAudio {
		return nil, fmt.Errorf("only videos and audio have segments")
	}
	segments, err := l.aiSegmentRows(ctx, kind, id, found)
	if err != nil {
		return nil, err
	}
	err = l.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM "Segment" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "origin" = $ai`,
			db.Vars{"kind": kind, "id": id, "ai": schema.SegmentOriginAi}); err != nil {
			return err
		}
		for _, segment := range segments {
			if _, err := tx.Insert(ctx, "Segment", segment); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return l.Segments(ctx, kind, id)
}

// aiSegmentRows checks the found segments and turns them into rows. Tags are
// written before the segments' transaction opens, since SQLite has one writer.
func (l *Library) aiSegmentRows(ctx context.Context, kind schema.RecordKind, id string, found []AISegment) ([]schema.Segment, error) {
	rows := make([]schema.Segment, 0, len(found))
	for _, segment := range found {
		if segment.StartSeconds < 0 || segment.EndSeconds <= segment.StartSeconds {
			return nil, fmt.Errorf("a segment must end after it starts (%v–%v)", segment.StartSeconds, segment.EndSeconds)
		}
		tagID, err := l.findOrCreateNamed(ctx, tableTag, segment.Tag)
		if err != nil {
			return nil, err
		}
		if tagID == "" {
			return nil, fmt.Errorf("an ai segment needs a tag")
		}
		end := segment.EndSeconds
		rows = append(rows, schema.Segment{
			SubjectKind:  kind,
			SubjectId:    id,
			StartSeconds: segment.StartSeconds,
			EndSeconds:   &end,
			Tag:          &tagID,
			Performers:   []string{},
			Origin:       schema.SegmentOriginAi,
		})
	}
	return rows, nil
}

// SceneMatch is a video whose ai segments carry one of the asked tags, with
// those segments and how many seconds they cover together.
type SceneMatch struct {
	VideoID         string
	Segments        []schema.Segment
	MatchingSeconds float64
}

// SceneMatches finds the videos whose ai segments carry any of the tags
// (by name), the most matching seconds first; with videoIDs, only those, in
// their order. Trashed and blocked videos are left out.
func (l *Library) SceneMatches(ctx context.Context, tagNames []string, videoIDs []string, limit, offset int) ([]SceneMatch, error) {
	tagIDs, err := l.existingTagIDs(ctx, tagNames)
	if err != nil || len(tagIDs) == 0 {
		return []SceneMatch{}, err
	}
	conditions, vars := FilterConditions(schema.RecordKindVideo, schema.Filter{})
	conditions = append(conditions, `"segment"."tag" IN (SELECT value FROM json_each($tags))`)
	vars["tags"] = tagIDs
	if len(videoIDs) > 0 {
		conditions = append(conditions, `"media"."id" IN (SELECT value FROM json_each($videoIds))`)
		vars["videoIds"] = videoIDs
	}
	rows, err := db.QueryAs[schema.Segment](ctx, l.database,
		`SELECT "segment".*`+MediaFrom(schema.RecordKindVideo)+
			` JOIN "Segment" AS "segment" ON "segment"."subjectKind" = 'video' AND "segment"."subjectId" = "media"."id"
			   AND "segment"."deletedAt" IS NULL AND "segment"."origin" = 'ai'
			 WHERE `+strings.Join(conditions, " AND ")+` ORDER BY "segment"."startSeconds"`, vars)
	if err != nil {
		return nil, fmt.Errorf("read scene matches: %w", err)
	}
	matches := groupSceneMatches(rows)
	if len(videoIDs) > 0 {
		sortByIDs(matches, videoIDs)
	} else {
		sortByMatchingSeconds(matches)
	}
	return pageSceneMatches(matches, limit, offset), nil
}

// existingTagIDs are the ids of the live tags with the names; names no tag
// has are left out (nothing is created).
func (l *Library) existingTagIDs(ctx context.Context, tagNames []string) ([]string, error) {
	keys := make([]string, 0, len(tagNames))
	for _, name := range tagNames {
		if key := names.Key(name); key != "" {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return nil, nil
	}
	return l.database.Strings(ctx,
		`SELECT "id" FROM "Tag" WHERE "nameKey" IN (SELECT value FROM json_each($keys)) AND "deletedAt" IS NULL`,
		db.Vars{"keys": keys})
}

// groupSceneMatches gathers segments (in time order) per video, in the order
// the videos first appear.
func groupSceneMatches(segments []schema.Segment) []SceneMatch {
	indexByVideo := map[string]int{}
	matches := []SceneMatch{}
	for _, segment := range segments {
		index, seen := indexByVideo[segment.SubjectId]
		if !seen {
			index = len(matches)
			indexByVideo[segment.SubjectId] = index
			matches = append(matches, SceneMatch{VideoID: segment.SubjectId})
		}
		matches[index].Segments = append(matches[index].Segments, segment)
	}
	for index := range matches {
		matches[index].MatchingSeconds = coveredSeconds(matches[index].Segments)
	}
	return matches
}

// coveredSeconds is how much of the timeline the segments (in start order)
// cover, overlaps counted once.
func coveredSeconds(segments []schema.Segment) float64 {
	total, reached := 0.0, 0.0
	for _, segment := range segments {
		if segment.EndSeconds == nil {
			continue
		}
		start := max(segment.StartSeconds, reached)
		if *segment.EndSeconds > start {
			total += *segment.EndSeconds - start
		}
		reached = max(reached, *segment.EndSeconds)
	}
	return total
}

// sortByMatchingSeconds puts the most matching videos first, ties by id.
func sortByMatchingSeconds(matches []SceneMatch) {
	sort.SliceStable(matches, func(left, right int) bool {
		if matches[left].MatchingSeconds != matches[right].MatchingSeconds {
			return matches[left].MatchingSeconds > matches[right].MatchingSeconds
		}
		return matches[left].VideoID < matches[right].VideoID
	})
}

// sortByIDs orders the matches as the ids are listed.
func sortByIDs(matches []SceneMatch, ids []string) {
	rank := make(map[string]int, len(ids))
	for index, id := range ids {
		rank[id] = index
	}
	sort.SliceStable(matches, func(left, right int) bool {
		return rank[matches[left].VideoID] < rank[matches[right].VideoID]
	})
}

// pageSceneMatches slices one page out of the matches.
func pageSceneMatches(matches []SceneMatch, limit, offset int) []SceneMatch {
	if offset < 0 || offset >= len(matches) {
		return []SceneMatch{}
	}
	if limit <= 0 {
		return matches[offset:]
	}
	return matches[offset:min(offset+limit, len(matches))]
}
