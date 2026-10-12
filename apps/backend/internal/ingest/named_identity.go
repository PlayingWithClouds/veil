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

// nameKeyedTables are the tables whose records are matched by spelling-insensitive
// name (the name_key column) as well as by site identity.
var nameKeyedTables = map[string]bool{"performer": true, "studio": true, "tag": true}

// hasSiteIdentity reports whether records of table carry a source_url and
// external_id of their own. Tags are known by name only.
func hasSiteIdentity(table string) bool {
	return table != "tag"
}

// nameKey is the comparison form of a name in table. Studios and tags compare
// without spaces, so a brand written "JulesJordan" and "Jules Jordan" is one
// studio and "blow job" and "blowjob" one tag; people compare word by word,
// since joining their names would merge different ones.
func nameKey(table, name string) string {
	if table == "studio" || table == "tag" {
		return names.CompactKey(name)
	}
	return names.Key(name)
}

// findByNameKey returns the oldest record of table whose name matches name in
// spelling-insensitive form (case, accents, punctuation) or lists it among its
// aliases, preferring an exact spelling. A name without letters or digits
// only matches exactly.
func (s *Service) findByNameKey(ctx context.Context, table, name string) (db.RecordID, bool, error) {
	matches, err := s.database.Strings(ctx,
		fmt.Sprintf(`SELECT id FROM %[1]s
		 WHERE name = $name
		    OR ($key <> '' AND name_key = $key)
		    OR EXISTS (SELECT 1 FROM json_each(%[1]s.aliases) WHERE value = $name COLLATE NOCASE)
		 ORDER BY name = $name DESC, created_at, id LIMIT 1`, table),
		db.Vars{"key": nameKey(table, name), "name": name})
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

// findRegisteredIdentity returns the record of table that a merged-away record's
// source_url or external_id (see entity_identity) now resolves to, or "".
func (s *Service) findRegisteredIdentity(ctx context.Context, table, kind, value string) (string, error) {
	matches, err := s.database.Strings(ctx,
		`SELECT record FROM entity_identity WHERE kind = $kind AND value = $value AND record LIKE $prefix`,
		db.Vars{"kind": kind, "value": value, "prefix": table + ":%"})
	if err != nil {
		return "", fmt.Errorf("lookup %s identity: %w", table, err)
	}
	if len(matches) == 0 {
		return "", nil
	}
	return matches[0], nil
}

// findIdentityFallback resolves an observation that matched no record by its own
// identity: first through identities of merged records, then (name-keyed tables
// only) by the spelling-insensitive name, so a performer or channel page adopts
// the record a scene credit created (and the other way round). "" when
// nothing matches.
func (s *Service) findIdentityFallback(ctx context.Context, table string, obs map[string]any) (string, error) {
	for _, kind := range identityKinds {
		value, _ := obs[kind].(string)
		if value == "" || !hasSiteIdentity(table) {
			continue
		}
		record, err := s.findRegisteredIdentity(ctx, table, kind, value)
		if err != nil || record != "" {
			return record, err
		}
	}
	if !nameKeyedTables[table] {
		return "", nil
	}
	name, _ := obs["name"].(string)
	id, found, err := s.findByNameKey(ctx, table, strings.TrimSpace(name))
	if err != nil {
		return "", fmt.Errorf("lookup %s by name: %w", table, err)
	}
	if !found {
		return "", nil
	}
	return id.String(), nil
}

// registerAlternateIdentities records the observation's source_url and
// external_id as identities of the record when it holds different ones, so a
// later scrape of that page finds the record again instead of creating a
// duplicate.
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
	if nameKeyedTables[table] {
		fields["name_key"] = nameKey(table, name)
	}
	return fields
}
