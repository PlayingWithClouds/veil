package library

import (
	"context"
	"strings"

	"github.com/playingwithclouds/veil/internal/schema"
)

// MatchingIDs returns which of ids are videos or galleries (by kind) that
// match filter, in no particular order. Trashed and blocked media never match.
func (l *Library) MatchingIDs(ctx context.Context, kind schema.RecordKind, filter schema.Filter, ids []string) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}
	conditions, vars := filterConditions(kind, filter)
	conditions = append(conditions, `"media"."id" IN (SELECT value FROM json_each($matchIds))`)
	vars["matchIds"] = ids
	return l.database.Strings(ctx, `SELECT "media"."id"`+mediaSelect(kind)+` WHERE `+strings.Join(conditions, " AND "), vars)
}

// AddedBetween lists the ids of the videos or galleries (by kind) matching
// filter that were created at or after since and before until (both
// db.TimeFormat), newest first.
func (l *Library) AddedBetween(ctx context.Context, kind schema.RecordKind, filter schema.Filter, since, until string, limit int) ([]string, error) {
	conditions, vars := filterConditions(kind, filter)
	conditions = append(conditions, `"media"."createdAt" >= $addedSince`, `"media"."createdAt" < $addedUntil`)
	vars["addedSince"] = since
	vars["addedUntil"] = until
	vars["limit"], _ = pageBounds(limit, 0)
	return l.database.Strings(ctx, `SELECT "media"."id"`+mediaSelect(kind)+` WHERE `+strings.Join(conditions, " AND ")+
		` ORDER BY "media"."createdAt" DESC, "media".rowid DESC LIMIT $limit`, vars)
}
