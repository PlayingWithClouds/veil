package library

import (
	"context"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// StudioDuplicates lists studios that are probably the same company as the
// one with id: another studio whose name or an alias folds to the same key as
// this one's name or an alias. Trashed studios are left out.
func (l *Library) StudioDuplicates(ctx context.Context, id string) ([]schema.Studio, error) {
	studio, err := l.Studio(ctx, id)
	if err != nil || studio == nil {
		return nil, err
	}
	keys := nameKeys(studio.Name, studio.Aliases)
	candidates, err := db.QueryAs[schema.Studio](ctx, l.database,
		`SELECT * FROM "Studio" AS "candidate"
		WHERE "candidate"."id" != $id AND "candidate"."deletedAt" IS NULL
			AND ("candidate"."nameKey" IN (SELECT value FROM json_each($keys)) OR "candidate"."aliases" != '[]')
		ORDER BY "candidate"."name"`,
		db.Vars{"id": id, "keys": mapKeys(keys)})
	if err != nil {
		return nil, err
	}
	matches := []schema.Studio{}
	for _, candidate := range candidates {
		if sharesKey(keys, nameKeys(candidate.Name, candidate.Aliases)) {
			matches = append(matches, candidate)
		}
	}
	return matches, nil
}

// SharedVideoCount counts the live videos crediting both performers.
func (l *Library) SharedVideoCount(ctx context.Context, firstID, secondID string) (int, error) {
	return l.database.Int(ctx,
		`SELECT COUNT(*) FROM "Video" WHERE "deletedAt" IS NULL
		 AND EXISTS (SELECT 1 FROM json_each("performers") WHERE value = $first)
		 AND EXISTS (SELECT 1 FROM json_each("performers") WHERE value = $second)`,
		db.Vars{"first": firstID, "second": secondID})
}

// CreditCount counts the live videos crediting a performer, or made by a
// studio: who survives a merge.
func (l *Library) CreditCount(ctx context.Context, kind schema.RecordKind, id string) (int, error) {
	condition := `EXISTS (SELECT 1 FROM json_each("performers") WHERE value = $id)`
	if kind == schema.RecordKindStudio {
		condition = `"studio" = $id`
	}
	return l.database.Int(ctx, `SELECT COUNT(*) FROM "Video" WHERE "deletedAt" IS NULL AND `+condition, db.Vars{"id": id})
}

// PeopleAfter lists up to limit live performers or studios created after
// the cursor (a createdAt), oldest first: the dedupe sweep walks them in turns.
func (l *Library) PeopleAfter(ctx context.Context, kind schema.RecordKind, cursor string, limit int) ([]string, []string, error) {
	rows, err := l.database.Query(ctx,
		`SELECT "id", "createdAt" FROM `+db.Quote(tableFor(kind))+` WHERE "deletedAt" IS NULL AND "createdAt" > $cursor
		 ORDER BY "createdAt" LIMIT $limit`,
		db.Vars{"cursor": cursor, "limit": limit})
	if err != nil {
		return nil, nil, err
	}
	ids, stamps := make([]string, 0, len(rows)), make([]string, 0, len(rows))
	for _, row := range rows {
		id, _ := row["id"].(string)
		stamp, _ := row["createdAt"].(string)
		ids, stamps = append(ids, id), append(stamps, stamp)
	}
	return ids, stamps, nil
}
