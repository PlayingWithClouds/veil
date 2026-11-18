package recommend

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Verdicts are the user's own answers about a tag, performer, studio or site:
// the latest like event of each (see latestLikeQuery), read without a time
// window, since a "yes" or "not really" on the taste page stands until taken
// back. A tag the scene model labels and the user said yes to is a liked
// scene: it drives the shorts, the phone's compilation and card previews.

// verdicts holds the latest answer per record and per site: 1 for yes, -1
// for no; unanswered ones are absent.
type verdicts struct {
	records map[schema.RecordRef]float64
	sites   map[string]float64
}

// loadVerdicts reads the latest like of every record and site.
func (e *Engine) loadVerdicts(ctx context.Context) (*verdicts, error) {
	rows, err := e.database.Query(ctx, latestLikeQuery, db.Vars{"since": "", "limit": eventReadLimit})
	if err != nil {
		return nil, fmt.Errorf("read verdicts: %w", err)
	}
	out := &verdicts{records: map[schema.RecordRef]float64{}, sites: map[string]float64{}}
	for _, row := range rows {
		value := 1.0
		if rowFloat(row, "value") < 0 {
			value = -1
		}
		if row["subjectKind"] == nil {
			out.sites[rowString(row, "site")] = value
			continue
		}
		out.records[rowRef(row, "subjectKind", "subjectId")] = value
	}
	return out, nil
}

// of is the answer about one record as the API gives it: true for yes,
// false for no, nil when unanswered.
func (answers *verdicts) of(kind schema.RecordKind, id string) *bool {
	return answerOf(answers.records[schema.RecordRef{Kind: kind, Id: id}])
}

// answerOf turns a stored answer into the API's: nil for none.
func answerOf(value float64) *bool {
	if value == 0 {
		return nil
	}
	yes := value > 0
	return &yes
}

// likedSceneTagsQuery reads the live tags whose latest like is a yes and that
// the scene model put on a video, by name.
const likedSceneTagsQuery = `SELECT "tag".* FROM "Tag" AS "tag"
	WHERE "tag"."deletedAt" IS NULL AND "tag"."id" IN (SELECT value FROM json_each($liked))
	  AND EXISTS (SELECT 1 FROM "Segment" AS "segment" WHERE "segment"."tag" = "tag"."id"
	    AND "segment"."origin" = 'ai' AND "segment"."deletedAt" IS NULL)
	ORDER BY "tag"."name"`

// LikedSceneTags are the scene labels the user said yes to: tags of ai
// segments whose latest like is positive, by name.
func (e *Engine) LikedSceneTags(ctx context.Context) ([]schema.Tag, error) {
	answers, err := e.loadVerdicts(ctx)
	if err != nil {
		return nil, err
	}
	liked := []string{}
	for ref, value := range answers.records {
		if ref.Kind == schema.RecordKindTag && value > 0 {
			liked = append(liked, ref.Id)
		}
	}
	if len(liked) == 0 {
		return []schema.Tag{}, nil
	}
	tags, err := db.QueryAs[schema.Tag](ctx, e.database, likedSceneTagsQuery, db.Vars{"liked": liked})
	if err != nil {
		return nil, fmt.Errorf("read liked scene tags: %w", err)
	}
	return tags, nil
}
