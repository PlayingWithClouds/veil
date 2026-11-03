package library

import (
	"context"
	"slices"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// SimilarByTags lists the ids of up to limit other records of a media kind
// sharing the most tags with the record id, newest first among equals.
// Blocked and trashed records, and those sharing no tag, are left out.
func (l *Library) SimilarByTags(ctx context.Context, kind schema.RecordKind, id string, limit int) ([]string, error) {
	table := db.Quote(tableFor(kind))
	vars := db.Vars{"subject": id}
	vars["limit"], _ = pageBounds(limit, 0)
	ids, err := l.database.Strings(ctx, `SELECT "other"."id" FROM `+table+` AS "other", json_each("other"."tags") AS "tag"
		WHERE "other"."deletedAt" IS NULL AND "other"."id" != $subject
			AND "tag"."value" IN (SELECT value FROM json_each((SELECT "tags" FROM `+table+` WHERE "id" = $subject)))
		GROUP BY "other"."id"
		ORDER BY COUNT(*) DESC, "other"."createdAt" DESC
		LIMIT $limit`, vars)
	if err != nil || len(ids) == 0 {
		return ids, err
	}
	visible, err := l.MatchingIDs(ctx, kind, schema.Filter{}, ids)
	if err != nil {
		return nil, err
	}
	return slices.DeleteFunc(ids, func(candidate string) bool { return !slices.Contains(visible, candidate) }), nil
}

// loadAll loads the records of ids with load, in their order; missing ones are skipped.
func loadAll[T any](ctx context.Context, ids []string, load func(context.Context, string) (*T, error)) ([]T, error) {
	records := make([]T, 0, len(ids))
	for _, id := range ids {
		record, err := load(ctx, id)
		if err != nil {
			return nil, err
		}
		if record != nil {
			records = append(records, *record)
		}
	}
	return records, nil
}
