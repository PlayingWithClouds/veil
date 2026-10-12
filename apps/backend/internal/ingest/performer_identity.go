package ingest

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
)

// identityKinds are the observation fields that identify a record on one site.
var identityKinds = []string{"source_url", "external_id"}

// findPerformerByName returns the oldest performer whose name matches name in
// spelling-insensitive form (case, accents, punctuation) or lists it among its
// aliases, preferring an exact spelling. A name without letters or digits
// only matches exactly.
func (s *Service) findPerformerByName(ctx context.Context, name string) (db.RecordID, bool, error) {
	key := names.Key(name)
	matches, err := s.database.Strings(ctx,
		`SELECT id FROM performer
		 WHERE name = $name
		    OR ($key <> '' AND name_key = $key)
		    OR EXISTS (SELECT 1 FROM json_each(performer.aliases) WHERE value = $name COLLATE NOCASE)
		 ORDER BY name = $name DESC, created_at, id LIMIT 1`,
		db.Vars{"key": key, "name": name})
	if err != nil {
		return db.RecordID{}, false, err
	}
	if len(matches) == 0 {
		return db.RecordID{}, false, nil
	}
	parsed, err := db.ParseRecordID(matches[0])
	if err != nil {
		return db.RecordID{}, false, err
	}
	return *parsed, true, nil
}

// findPerformerFallback resolves a performer observation that matched no record
// by its own identity: first through identities of merged records, then by the
// spelling-insensitive name, so a performer page adopts the record a scene
// credit created (and the other way round). "" when nothing matches.
func (s *Service) findPerformerFallback(ctx context.Context, obs map[string]any) (string, error) {
	for _, kind := range identityKinds {
		value, _ := obs[kind].(string)
		if value == "" {
			continue
		}
		matches, err := s.database.Strings(ctx,
			`SELECT record FROM entity_identity WHERE kind = $kind AND value = $value AND record LIKE 'performer:%'`,
			db.Vars{"kind": kind, "value": value})
		if err != nil {
			return "", fmt.Errorf("lookup performer identity: %w", err)
		}
		if len(matches) > 0 {
			return matches[0], nil
		}
	}
	name, _ := obs["name"].(string)
	id, found, err := s.findPerformerByName(ctx, strings.TrimSpace(name))
	if err != nil {
		return "", fmt.Errorf("lookup performer by name: %w", err)
	}
	if !found {
		return "", nil
	}
	return id.String(), nil
}

// registerAlternateIdentities records the observation's source_url and
// external_id as identities of the performer when the record holds different
// ones, so a later scrape of that page finds the record again instead of
// creating a duplicate.
func (s *Service) registerAlternateIdentities(ctx context.Context, canonicalID string, obs map[string]any) error {
	current, err := s.database.GetString(ctx, canonicalID)
	if err != nil || current == nil {
		return err
	}
	for _, kind := range identityKinds {
		value, _ := obs[kind].(string)
		if value == "" || value == current[kind] {
			continue
		}
		_, err := s.database.Exec(ctx,
			`INSERT OR IGNORE INTO entity_identity (kind, value, record) VALUES ($kind, $value, $record)`,
			db.Vars{"kind": kind, "value": value, "record": canonicalID})
		if err != nil {
			return fmt.Errorf("register %s identity: %w", kind, err)
		}
	}
	return nil
}

// namedRecordFields are the columns of a new row in a name-keyed table.
func namedRecordFields(table, name string) map[string]any {
	fields := map[string]any{"name": name}
	if table == "performer" {
		fields["name_key"] = names.Key(name)
	}
	return fields
}
