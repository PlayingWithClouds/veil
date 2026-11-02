package library

import (
	"context"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
)

// performerIDs finds or creates a performer for every credited name and
// returns their ids in credit order, without duplicates.
func (l *Library) performerIDs(ctx context.Context, credited []string) ([]string, error) {
	ids := []string{}
	seen := map[string]bool{}
	for _, name := range credited {
		id, err := l.findOrCreateNamed(ctx, tablePerformer, name)
		if err != nil {
			return nil, err
		}
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids, nil
}

// tagIDs finds or creates a tag for every name and returns their ids.
func (l *Library) tagIDs(ctx context.Context, tagNames []string) ([]string, error) {
	ids := []string{}
	seen := map[string]bool{}
	for _, name := range tagNames {
		id, err := l.findOrCreateNamed(ctx, tableTag, name)
		if err != nil {
			return nil, err
		}
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
	}
	return ids, nil
}

// findOrCreateNamed returns the id of the performer, studio or tag whose
// folded name matches, creating one when none does. A blank name yields "".
func (l *Library) findOrCreateNamed(ctx context.Context, table, name string) (string, error) {
	name = strings.TrimSpace(name)
	key := names.Key(name)
	if key == "" {
		return "", nil
	}
	existing, err := l.findByNameKey(ctx, table, key)
	if err != nil || existing != "" {
		return existing, err
	}
	return l.database.Insert(ctx, table, newNamedRecord(table, name, key))
}

// newNamedRecord is the column set of a fresh performer, studio or tag
// known only by its name.
func newNamedRecord(table, name, key string) db.Vars {
	record := db.Vars{"name": name, "nameKey": key, "aliases": []string{}}
	switch table {
	case tablePerformer:
		record["attributes"] = map[string]any{}
		record["links"] = []string{}
		record["tags"] = []string{}
	case tableStudio:
		record["tags"] = []string{}
	}
	return record
}

// findByNameKey returns the id of the first row of table with key, or "".
// Live rows come first; a trashed one still matches, so a credit of a
// trashed performer or tag stays on it (hidden) instead of recreating it.
func (l *Library) findByNameKey(ctx context.Context, table, key string) (string, error) {
	ids, err := l.database.Strings(ctx,
		`SELECT "id" FROM `+db.Quote(table)+` WHERE "nameKey" = $key ORDER BY "deletedAt" IS NOT NULL, "createdAt" LIMIT 1`,
		db.Vars{"key": key})
	if err != nil || len(ids) == 0 {
		return "", err
	}
	return ids[0], nil
}
