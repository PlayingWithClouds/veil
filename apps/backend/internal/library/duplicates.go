package library

import (
	"context"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/schema"
)

// PerformerDuplicates lists performers that are probably the same person as
// the one with id: another performer whose name or an alias folds to the same
// key as this one's name or an alias. Only a suggestion: merging stays the
// user's call. Trashed and blocked performers are left out.
func (l *Library) PerformerDuplicates(ctx context.Context, id string) ([]schema.Performer, error) {
	performer, err := l.Performer(ctx, id)
	if err != nil || performer == nil {
		return nil, err
	}
	keys := nameKeys(performer.Name, performer.Aliases)
	candidates, err := db.QueryAs[schema.Performer](ctx, l.database,
		`SELECT * FROM "Performer" AS "candidate"
		WHERE "candidate"."id" != $id AND "candidate"."deletedAt" IS NULL
			AND ("candidate"."nameKey" IN (SELECT value FROM json_each($keys)) OR "candidate"."aliases" != '[]')
			AND NOT EXISTS (SELECT 1 FROM "UserState" AS "blocked"
				WHERE "blocked"."subjectKind" = 'performer' AND "blocked"."subjectId" = "candidate"."id"
					AND "blocked"."blocked" = 1 AND "blocked"."deletedAt" IS NULL)
		ORDER BY "candidate"."name"`,
		db.Vars{"id": id, "keys": mapKeys(keys)})
	if err != nil {
		return nil, err
	}
	matches := []schema.Performer{}
	for _, candidate := range candidates {
		if sharesKey(keys, nameKeys(candidate.Name, candidate.Aliases)) {
			matches = append(matches, candidate)
		}
	}
	return matches, nil
}

// nameKeys folds a name and its aliases into comparison keys.
func nameKeys(name string, aliases []string) map[string]bool {
	keys := map[string]bool{}
	for _, value := range append([]string{name}, aliases...) {
		if key := names.Key(value); key != "" {
			keys[key] = true
		}
	}
	return keys
}

// sharesKey reports whether two key sets have a key in common.
func sharesKey(first, second map[string]bool) bool {
	for key := range second {
		if first[key] {
			return true
		}
	}
	return false
}

// mapKeys lists a set's keys, for binding as a JSON array.
func mapKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	return keys
}
