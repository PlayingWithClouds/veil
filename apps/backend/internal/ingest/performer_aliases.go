package ingest

import (
	"context"
	"fmt"
	"log"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
)

// absorbAliases adds the given names to the performer's aliases, skipping its
// own name and the ones it has. Sites that know a performer under other stage
// names (FreeOnes) are how two records of one person come to light.
func (s *Service) absorbAliases(ctx context.Context, performerID string, aliases []string) error {
	performer, err := s.database.GetString(ctx, performerID)
	if err != nil || performer == nil {
		return err
	}
	ownName, _ := performer["name"].(string)
	current := jsonStringList(performer["aliases"])
	merged := unionStrings(current, aliases)
	merged = withoutName(merged, ownName)
	if len(merged) == len(current) {
		return nil
	}
	record, err := db.ParseRecordID(performerID)
	if err != nil {
		return err
	}
	return s.database.Merge(ctx, *record, map[string]any{"aliases": merged})
}

// withoutName drops the entries that spell name.
func withoutName(values []string, name string) []string {
	kept := make([]string, 0, len(values))
	for _, value := range values {
		if names.Key(value) == names.Key(name) {
			continue
		}
		kept = append(kept, value)
	}
	return kept
}

// mergeAliasTwins merges every other performer whose name is one of this
// performer's aliases into the better-used of the two, and returns the id that
// holds them all. Only full names count (two words or more): a single-word
// alias like "Angel" is shared by many people. Returns performerID when
// nothing matched.
func (s *Service) mergeAliasTwins(ctx context.Context, performerID string) (string, error) {
	performer, err := s.database.GetString(ctx, performerID)
	if err != nil || performer == nil {
		return performerID, err
	}
	survivorID := performerID
	for _, alias := range jsonStringList(performer["aliases"]) {
		key := names.Key(alias)
		if !isFullName(key) {
			continue
		}
		twinIDs, err := s.database.Strings(ctx,
			`SELECT id FROM performer WHERE name_key = $key AND id <> $id`, db.Vars{"key": key, "id": survivorID})
		if err != nil {
			return survivorID, fmt.Errorf("find performers called %q: %w", alias, err)
		}
		for _, twinID := range twinIDs {
			survivorID, err = s.mergeBetter(ctx, "performer", survivorID, twinID)
			if err != nil {
				return survivorID, err
			}
		}
	}
	return survivorID, nil
}

// isFullName reports whether a comparison key has at least two words.
func isFullName(key string) bool {
	for _, character := range key {
		if character == ' ' {
			return true
		}
	}
	return false
}

// mergeBetter merges two records of a name-keyed table, keeping the one that is
// used more (ties: the first), and returns the survivor's id.
func (s *Service) mergeBetter(ctx context.Context, table, firstID, secondID string) (string, error) {
	survivorID, mergedID := firstID, secondID
	firstUses, err := s.useCount(ctx, table, firstID)
	if err != nil {
		return firstID, err
	}
	secondUses, err := s.useCount(ctx, table, secondID)
	if err != nil {
		return firstID, err
	}
	if secondUses > firstUses {
		survivorID, mergedID = secondID, firstID
	}
	if err := s.MergeRecords(ctx, table, survivorID, mergedID); err != nil {
		return firstID, err
	}
	log.Printf("ingest: merged %s %s into %s (alias match)", table, mergedID, survivorID)
	return survivorID, nil
}

// useCount is how much a record of a name-keyed table is used (see creditCount).
func (s *Service) useCount(ctx context.Context, table, recordID string) (int, error) {
	return s.database.Int(ctx,
		fmt.Sprintf(`SELECT %s FROM %[2]s WHERE id = $id`, creditCount(table), table), db.Vars{"id": recordID})
}

// PerformerName is a performer record's id and name.
type PerformerName struct {
	ID   string
	Name string
}

// PerformersWithoutProfile lists the performers no observation of pluginName
// describes yet, the most credited first.
func (s *Service) PerformersWithoutProfile(ctx context.Context, pluginName string) ([]PerformerName, error) {
	rows, err := s.database.Query(ctx,
		`SELECT id, name FROM performer
		 WHERE NOT EXISTS (SELECT 1 FROM observation WHERE observation.target = performer.id AND observation.plugin = $plugin)
		 ORDER BY `+creditCount("performer")+` DESC, created_at, id`,
		db.Vars{"plugin": pluginName})
	if err != nil {
		return nil, fmt.Errorf("find performers without a %s profile: %w", pluginName, err)
	}
	performers := make([]PerformerName, 0, len(rows))
	for _, row := range rows {
		performers = append(performers, PerformerName{ID: rowText(row, "id"), Name: rowText(row, "name")})
	}
	return performers, nil
}
