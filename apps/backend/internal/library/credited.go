package library

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// CreditedPerformers lists up to limit performers credited on media of a
// kind (the voices of audio, the authors of stories): favourites first,
// then by how much they're credited on. Trashed and blocked performers are
// left out.
func (l *Library) CreditedPerformers(ctx context.Context, kind schema.MediaKind, limit int) ([]schema.Performer, error) {
	recordKind, err := recordKindOf(kind)
	if err != nil {
		return nil, err
	}
	vars := db.Vars{}
	vars["limit"], _ = pageBounds(limit, 0)
	ids, err := l.database.Strings(ctx, `SELECT "performer"."id"
		FROM `+db.Quote(tableFor(recordKind))+` AS "media", json_each("media"."performers") AS "credit"
		JOIN "Performer" AS "performer" ON "performer"."id" = "credit"."value" AND "performer"."deletedAt" IS NULL
		LEFT JOIN "UserState" AS "state" ON "state"."subjectKind" = 'performer'
			AND "state"."subjectId" = "performer"."id" AND "state"."deletedAt" IS NULL
		WHERE "media"."deletedAt" IS NULL AND COALESCE("state"."blocked", 0) = 0
		GROUP BY "performer"."id"
		ORDER BY MAX(COALESCE("state"."favorite", 0)) DESC, COUNT(*) DESC, "performer"."name"
		LIMIT $limit`, vars)
	if err != nil {
		return nil, err
	}
	return loadAll(ctx, ids, l.Performer)
}

// MediaTagCounts counts tags across the media of a kind matching filter,
// like TagCounts for a record kind.
func (l *Library) MediaTagCounts(ctx context.Context, kind schema.MediaKind, filter schema.Filter, limit int) ([]TagCount, error) {
	recordKind, err := recordKindOf(kind)
	if err != nil {
		return nil, err
	}
	return l.TagCounts(ctx, recordKind, filter, limit)
}

// recordKindOf is the record kind of a media kind.
func recordKindOf(kind schema.MediaKind) (schema.RecordKind, error) {
	recordKind, known := mediaKinds[kind]
	if !known {
		return "", fmt.Errorf("unknown media kind %q", kind)
	}
	return recordKind, nil
}
