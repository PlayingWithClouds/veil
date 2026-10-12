package ingest

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
)

// performerFillColumns are the performer columns a merge copies from the merged
// record when the survivor has none.
var performerFillColumns = []string{
	"details", "gender", "birthdate", "death_date", "country", "ethnicity", "eye_color",
	"hair_color", "height_cm", "weight_kg", "measurements", "fake_tits", "tattoos",
	"piercings", "career_length", "url", "twitter", "instagram", "image_path", "rating",
}

// MergePerformers folds the performer mergedID into survivorID in one
// transaction: every credit and reference moves to the survivor, fields the
// survivor lacks are copied over, the merged record's site identities stay
// resolvable (see entity_identity) and the merged record is deleted.
func (s *Service) MergePerformers(ctx context.Context, survivorID, mergedID string) error {
	if survivorID == mergedID {
		return fmt.Errorf("merge performers: %s into itself", survivorID)
	}
	return s.database.Tx(ctx, func(tx *db.DB) error {
		survivor, err := tx.GetString(ctx, survivorID)
		if err != nil || survivor == nil {
			return fmt.Errorf("merge performers: load survivor %s: %v", survivorID, err)
		}
		merged, err := tx.GetString(ctx, mergedID)
		if err != nil || merged == nil {
			return fmt.Errorf("merge performers: load merged %s: %v", mergedID, err)
		}
		vars := db.Vars{"survivor": survivorID, "merged": mergedID}
		if err := foldPerformerFields(ctx, tx, survivorID, survivor, merged, vars); err != nil {
			return err
		}
		for _, statement := range performerRepointStatements() {
			if _, err := tx.Exec(ctx, statement, vars); err != nil {
				return fmt.Errorf("merge performers %s into %s: %w", mergedID, survivorID, err)
			}
		}
		return nil
	})
}

// foldPerformerFields carries the merged record's own data over to the
// survivor: empty fields, the favorite flag, tags, aliases and site identities.
func foldPerformerFields(ctx context.Context, tx *db.DB, survivorID string, survivor, merged db.Row, vars db.Vars) error {
	if _, err := tx.Exec(ctx, fillStatement(), vars); err != nil {
		return fmt.Errorf("fill performer fields: %w", err)
	}
	// The merged row is deleted at the end, but its external_id is unique, so it
	// is released before the survivor may take it.
	if _, err := tx.Exec(ctx, `UPDATE performer SET external_id = NULL, source_url = NULL WHERE id = $merged`, vars); err != nil {
		return fmt.Errorf("release merged identities: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE entity_identity SET record = $survivor WHERE record = $merged`, vars); err != nil {
		return fmt.Errorf("repoint identities: %w", err)
	}
	updates := map[string]any{
		"aliases": mergedAliases(survivor, merged),
		"tags":    unionStrings(jsonStringList(survivor["tags"]), jsonStringList(merged["tags"])),
	}
	for _, kind := range identityKinds {
		if err := keepIdentity(ctx, tx, kind, survivor, merged, updates, vars); err != nil {
			return err
		}
	}
	survivorRecord, err := db.ParseRecordID(survivorID)
	if err != nil {
		return err
	}
	return tx.Merge(ctx, *survivorRecord, updates)
}

// keepIdentity gives the survivor the merged record's source_url or external_id
// when it has none of its own, else registers it as an extra identity.
func keepIdentity(ctx context.Context, tx *db.DB, kind string, survivor, merged db.Row, updates map[string]any, vars db.Vars) error {
	value, _ := merged[kind].(string)
	if value == "" {
		return nil
	}
	if current, _ := survivor[kind].(string); current == "" {
		updates[kind] = value
		return nil
	}
	_, err := tx.Exec(ctx,
		`INSERT OR IGNORE INTO entity_identity (kind, value, record) VALUES ($kind, $value, $survivor)`,
		db.Vars{"kind": kind, "value": value, "survivor": vars["survivor"]})
	if err != nil {
		return fmt.Errorf("keep %s identity: %w", kind, err)
	}
	return nil
}

// fillStatement copies each performerFillColumns value the survivor lacks from
// the merged record, and keeps a favorite flag set on either.
func fillStatement() string {
	assignments := make([]string, 0, len(performerFillColumns)+1)
	for _, column := range performerFillColumns {
		assignments = append(assignments, fmt.Sprintf(
			"%[1]s = coalesce(nullif(%[1]s, ''), (SELECT nullif(%[1]s, '') FROM performer WHERE id = $merged))", column))
	}
	assignments = append(assignments, "favorite = (favorite OR coalesce((SELECT favorite FROM performer WHERE id = $merged), 0))")
	return "UPDATE performer SET " + strings.Join(assignments, ", ") + " WHERE id = $survivor"
}

// mergedAliases is the survivor's aliases plus the merged record's aliases and,
// when it is spelled as a genuinely different name, its name.
func mergedAliases(survivor, merged db.Row) []string {
	aliases := unionStrings(jsonStringList(survivor["aliases"]), jsonStringList(merged["aliases"]))
	mergedName, _ := merged["name"].(string)
	survivorName, _ := survivor["name"].(string)
	if names.Key(mergedName) != names.Key(survivorName) {
		aliases = unionStrings(aliases, []string{mergedName})
	}
	return aliases
}

// unionStrings appends the values of extra missing from base, comparing names
// by names.Key so spellings of one name stay a single entry.
func unionStrings(base, extra []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(base)+len(extra))
	for _, value := range append(append([]string{}, base...), extra...) {
		key := names.Key(value)
		if key == "" {
			key = value
		}
		if value == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

// jsonStringList reads a decoded JSON column as strings.
func jsonStringList(value any) []string {
	items, _ := value.([]any)
	out := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			out = append(out, text)
		}
	}
	return out
}

// performerRepointStatements move everything that refers to $merged onto
// $survivor and finally delete $merged. Tables with a unique target keep the
// survivor's row when both exist.
func performerRepointStatements() []string {
	statements := []string{}
	for _, table := range []string{"scene", "gallery", "image"} {
		statements = append(statements, replacePerformerInList(table))
	}
	return append(statements,
		`UPDATE OR IGNORE performs_in SET performer = $survivor WHERE performer = $merged`,
		`DELETE FROM performs_in WHERE performer = $merged`,
		`UPDATE image SET media = $survivor WHERE media = $merged`,
		`UPDATE observation SET target = $survivor WHERE target = $merged`,
		`UPDATE changes SET canonical = $survivor WHERE canonical = $merged`,
		`UPDATE saved_filter SET performer_id = $survivor WHERE performer_id = $merged`,
		`UPDATE OR IGNORE collection_item SET media = $survivor WHERE media = $merged`,
		`DELETE FROM collection_item WHERE media = $merged`,
		`UPDATE OR IGNORE user_rating SET media = $survivor WHERE media = $merged`,
		`DELETE FROM user_rating WHERE media = $merged`,
		`UPDATE OR IGNORE watchlist SET media = $survivor WHERE media = $merged`,
		`DELETE FROM watchlist WHERE media = $merged`,
		`UPDATE OR IGNORE blocklist_entry SET target = $survivor WHERE target = $merged`,
		`DELETE FROM blocklist_entry WHERE target = $merged`,
		`UPDATE OR IGNORE search_subscription_item SET media = $survivor WHERE media = $merged`,
		`DELETE FROM search_subscription_item WHERE media = $merged`,
		`DELETE FROM search_subscription WHERE entity = $merged
			AND EXISTS (SELECT 1 FROM search_subscription WHERE entity = $survivor)`,
		`UPDATE search_subscription SET entity = $survivor WHERE entity = $merged`,
		`DELETE FROM search_subscription_item WHERE subscription NOT IN (SELECT id FROM search_subscription)`,
		`DELETE FROM performer WHERE id = $merged`,
	)
}

// replacePerformerInList rewrites a table's performers JSON list so $merged
// becomes $survivor, keeping the first occurrence order and no repeats.
func replacePerformerInList(table string) string {
	return fmt.Sprintf(`UPDATE %[1]s SET performers = (
		SELECT json_group_array(value) FROM (
			SELECT CASE WHEN entry.value = $merged THEN $survivor ELSE entry.value END AS value, min(entry.key) AS position
			FROM json_each(%[1]s.performers) AS entry
			GROUP BY CASE WHEN entry.value = $merged THEN $survivor ELSE entry.value END
			ORDER BY position))
		WHERE EXISTS (SELECT 1 FROM json_each(%[1]s.performers) WHERE value = $merged)`, table)
}

// RepairPerformers fills the name key of performers that predate it and merges
// the performers that share one, keeping the record with the most credits (then
// the better-cased name, then the oldest). Returns how many were merged away.
func (s *Service) RepairPerformers(ctx context.Context) (int, error) {
	if err := s.backfillPerformerNameKeys(ctx); err != nil {
		return 0, err
	}
	keys, err := s.database.Strings(ctx,
		`SELECT name_key FROM performer WHERE name_key <> '' GROUP BY name_key HAVING count(*) > 1`, nil)
	if err != nil {
		return 0, fmt.Errorf("find duplicate performers: %w", err)
	}
	mergedCount := 0
	for _, key := range keys {
		count, err := s.mergePerformerGroup(ctx, key)
		mergedCount += count
		if err != nil {
			return mergedCount, err
		}
	}
	if mergedCount > 0 {
		log.Printf("ingest: merged %d duplicate performers", mergedCount)
	}
	return mergedCount, nil
}

// backfillPerformerNameKeys sets name_key on performers that have none.
func (s *Service) backfillPerformerNameKeys(ctx context.Context) error {
	rows, err := s.database.Query(ctx, `SELECT id, name FROM performer WHERE name_key IS NULL`, nil)
	if err != nil {
		return fmt.Errorf("find performers without name key: %w", err)
	}
	return s.database.Tx(ctx, func(tx *db.DB) error {
		for _, row := range rows {
			name, _ := row["name"].(string)
			_, err := tx.Exec(ctx, `UPDATE performer SET name_key = $key WHERE id = $id`,
				db.Vars{"key": names.Key(name), "id": row["id"]})
			if err != nil {
				return fmt.Errorf("set performer name key: %w", err)
			}
		}
		return nil
	})
}

// mergePerformerGroup merges every performer with name key into the best one.
func (s *Service) mergePerformerGroup(ctx context.Context, key string) (int, error) {
	ids, err := s.database.Strings(ctx,
		`SELECT id FROM performer WHERE name_key = $key
		 ORDER BY (SELECT count(*) FROM performs_in WHERE performs_in.performer = performer.id) DESC,
		          (name <> lower(name) AND name <> upper(name)) DESC, created_at, id`,
		db.Vars{"key": key})
	if err != nil {
		return 0, fmt.Errorf("order duplicate performers: %w", err)
	}
	mergedCount := 0
	for _, mergedID := range ids[1:] {
		if err := s.MergePerformers(ctx, ids[0], mergedID); err != nil {
			return mergedCount, err
		}
		mergedCount++
	}
	return mergedCount, nil
}
