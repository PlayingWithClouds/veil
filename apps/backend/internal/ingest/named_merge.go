package ingest

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
)

// fillColumns are the columns a merge copies from the merged record when the
// survivor has none, per table.
var fillColumns = map[string][]string{
	"performer": {
		"details", "gender", "birthdate", "death_date", "country", "ethnicity", "eye_color",
		"hair_color", "height_cm", "weight_kg", "measurements", "fake_tits", "tattoos",
		"piercings", "career_length", "url", "twitter", "instagram", "image_path", "rating",
	},
	"studio": {"url", "image_path", "details", "parent"},
	"tag":    {"description", "category"},
	"scene": {
		"details", "date", "duration_seconds", "studio", "poster_path", "preview_video",
		"thumbnail_path", "thumbnail_seconds", "rating",
	},
	"gallery":    {"details", "date", "studio", "cover_path", "image_count", "rating"},
	"collection": {"details", "cover_path"},
}

// taggedTables are the tables with a JSON list of tag ids in "tags", and
// performerListTables those with one of performer ids in "performers".
var (
	taggedTables        = []string{"scene", "gallery", "image", "performer", "studio", "collection"}
	performerListTables = []string{"scene", "gallery", "image"}
)

// MergeRecords folds the record mergedID into survivorID of the same name-keyed
// table (performer, studio or tag) in one transaction: every credit and reference
// moves to the survivor, fields the survivor lacks are copied over, the merged
// record's site identities stay resolvable (see entity_identity) and the merged
// record is deleted.
func (s *Service) MergeRecords(ctx context.Context, table, survivorID, mergedID string) error {
	if !nameKeyedTables[table] {
		return fmt.Errorf("merge records: %s has no merge", table)
	}
	if survivorID == mergedID {
		return fmt.Errorf("merge %s: %s into itself", table, survivorID)
	}
	return s.database.Tx(ctx, func(tx *db.DB) error {
		survivor, err := tx.GetString(ctx, survivorID)
		if err != nil || survivor == nil {
			return fmt.Errorf("merge %s: load survivor %s: %v", table, survivorID, err)
		}
		merged, err := tx.GetString(ctx, mergedID)
		if err != nil || merged == nil {
			return fmt.Errorf("merge %s: load merged %s: %v", table, mergedID, err)
		}
		vars := db.Vars{"survivor": survivorID, "merged": mergedID}
		if err := foldFields(ctx, tx, table, survivorID, survivor, merged, vars); err != nil {
			return err
		}
		for _, statement := range repointStatements(table) {
			if _, err := tx.Exec(ctx, statement, vars); err != nil {
				return fmt.Errorf("merge %s %s into %s: %w", table, mergedID, survivorID, err)
			}
		}
		return nil
	})
}

// foldFields carries the merged record's own data over to the survivor: empty
// fields, the favorite flag, tags, aliases and site identities.
func foldFields(ctx context.Context, tx *db.DB, table, survivorID string, survivor, merged db.Row, vars db.Vars) error {
	if _, err := tx.Exec(ctx, fillStatement(table), vars); err != nil {
		return fmt.Errorf("fill %s fields: %w", table, err)
	}
	updates := map[string]any{"aliases": mergedAliases(table, survivor, merged)}
	if table != "tag" {
		updates["tags"] = unionStrings(jsonStringList(survivor["tags"]), jsonStringList(merged["tags"]))
	}
	if hasSiteIdentity(table) {
		if err := foldIdentities(ctx, tx, table, survivor, merged, updates, vars); err != nil {
			return err
		}
	}
	survivorRecord, err := db.ParseRecordID(survivorID)
	if err != nil {
		return err
	}
	return tx.Merge(ctx, *survivorRecord, updates)
}

// foldIdentities moves the merged record's site identities to the survivor.
func foldIdentities(ctx context.Context, tx *db.DB, table string, survivor, merged db.Row, updates map[string]any, vars db.Vars) error {
	// The merged row is deleted at the end, but its external_id is unique, so it
	// is released before the survivor may take it.
	if _, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s SET external_id = NULL, source_url = NULL WHERE id = $merged`, table), vars); err != nil {
		return fmt.Errorf("release merged identities: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE entity_identity SET record = $survivor WHERE record = $merged`, vars); err != nil {
		return fmt.Errorf("repoint identities: %w", err)
	}
	for _, kind := range identityKinds {
		if err := keepIdentity(ctx, tx, kind, survivor, merged, updates, vars); err != nil {
			return err
		}
	}
	return nil
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

// fillStatement copies each fillColumns value the survivor lacks from the
// merged record, and keeps the flags (a performer's favorite, a scene's
// organized mark) set on either.
func fillStatement(table string) string {
	assignments := make([]string, 0, len(fillColumns[table])+1)
	for _, column := range fillColumns[table] {
		assignments = append(assignments, fmt.Sprintf(
			"%[1]s = coalesce(nullif(%[1]s, ''), (SELECT nullif(%[1]s, '') FROM %[2]s WHERE id = $merged))", column, table))
	}
	switch table {
	case "performer":
		assignments = append(assignments, "favorite = (favorite OR coalesce((SELECT favorite FROM performer WHERE id = $merged), 0))")
	case "scene":
		assignments = append(assignments,
			"organized = (organized OR coalesce((SELECT organized FROM scene WHERE id = $merged), 0))",
			"view_count = max(view_count, coalesce((SELECT view_count FROM scene WHERE id = $merged), 0))")
	case "gallery":
		assignments = append(assignments, "organized = (organized OR coalesce((SELECT organized FROM gallery WHERE id = $merged), 0))")
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE id = $survivor", table, strings.Join(assignments, ", "))
}

// mergedAliases is the survivor's aliases plus the merged record's aliases and,
// when it is spelled as a genuinely different name, its name.
func mergedAliases(table string, survivor, merged db.Row) []string {
	aliases := unionStrings(jsonStringList(survivor["aliases"]), jsonStringList(merged["aliases"]))
	mergedName, _ := merged["name"].(string)
	survivorName, _ := survivor["name"].(string)
	if nameKey(table, mergedName) != nameKey(table, survivorName) {
		aliases = unionStrings(aliases, []string{mergedName})
	}
	return withoutName(aliases, survivorName)
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

// repointStatements move everything that refers to $merged onto $survivor and
// finally delete $merged. Tables with a unique target keep the survivor's row
// when both exist.
func repointStatements(table string) []string {
	statements := specificRepointStatements(table)
	return append(statements,
		`UPDATE image SET media = $survivor WHERE media = $merged`,
		`UPDATE observation SET target = $survivor WHERE target = $merged`,
		`UPDATE changes SET canonical = $survivor WHERE canonical = $merged`,
		fmt.Sprintf(`UPDATE saved_filter SET %[1]s_id = $survivor WHERE %[1]s_id = $merged`, table),
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
		`UPDATE OR IGNORE feed_discovery SET entity = $survivor WHERE entity = $merged`,
		`DELETE FROM feed_discovery WHERE entity = $merged`,
		fmt.Sprintf(`DELETE FROM %s WHERE id = $merged`, table),
	)
}

// specificRepointStatements are the references only one kind of record has:
// tag lists for a tag, credits for a performer, the studio column and the parent link for a studio.
func specificRepointStatements(table string) []string {
	if table == "tag" {
		statements := []string{}
		for _, listTable := range taggedTables {
			statements = append(statements, replaceIDInList(listTable, "tags"))
		}
		return append(statements, `UPDATE scene_marker SET tag = $survivor WHERE tag = $merged`)
	}
	if table == "performer" {
		statements := []string{}
		for _, listTable := range performerListTables {
			statements = append(statements, replaceIDInList(listTable, "performers"))
		}
		return append(statements,
			`UPDATE OR IGNORE performs_in SET performer = $survivor WHERE performer = $merged`,
			`DELETE FROM performs_in WHERE performer = $merged`)
	}
	return []string{
		`UPDATE scene SET studio = $survivor WHERE studio = $merged`,
		`UPDATE gallery SET studio = $survivor WHERE studio = $merged`,
		`UPDATE image SET studio = $survivor WHERE studio = $merged`,
		`UPDATE studio SET parent = $survivor WHERE parent = $merged`,
		// A studio and its own parent merged leave the survivor its own parent.
		`UPDATE studio SET parent = NULL WHERE id = $survivor AND parent = $survivor`,
	}
}

// replaceIDInList rewrites a table's JSON id list (column) so $merged becomes
// $survivor, keeping the first occurrence order and no repeats.
func replaceIDInList(table, column string) string {
	return fmt.Sprintf(`UPDATE %[1]s SET %[2]s = (
		SELECT json_group_array(value) FROM (
			SELECT CASE WHEN entry.value = $merged THEN $survivor ELSE entry.value END AS value, min(entry.key) AS position
			FROM json_each(%[1]s.%[2]s) AS entry
			GROUP BY CASE WHEN entry.value = $merged THEN $survivor ELSE entry.value END
			ORDER BY position))
		WHERE EXISTS (SELECT 1 FROM json_each(%[1]s.%[2]s) WHERE value = $merged)`, table, column)
}

// creditCount is the SQL that counts what a record is used for, which decides
// which of several duplicates survives: credits for a performer, scenes for a
// studio or tag.
func creditCount(table string) string {
	if table == "performer" {
		return "(SELECT count(*) FROM performs_in WHERE performs_in.performer = performer.id)"
	}
	if table == "tag" {
		return "(SELECT count(*) FROM scene, json_each(scene.tags) AS entry WHERE entry.value = tag.id)"
	}
	return "(SELECT count(*) FROM scene WHERE scene.studio = studio.id)"
}

// RepairNamedRecords fills the name key of performers and studios that predate
// it and merges the ones that share one, keeping the record that is used most
// (then the better-cased name, then the oldest). Returns how many were merged away.
func (s *Service) RepairNamedRecords(ctx context.Context) (int, error) {
	total := 0
	for _, table := range []string{"performer", "studio", "tag"} {
		count, err := s.repairTable(ctx, table)
		total += count
		if err != nil {
			return total, err
		}
	}
	return total, s.repairAliasTwins(ctx)
}

// repairAliasTwins merges performers that carry another performer's name as an
// alias (see mergeAliasTwins).
func (s *Service) repairAliasTwins(ctx context.Context) error {
	performerIDs, err := s.database.Strings(ctx, `SELECT id FROM performer WHERE json_array_length(aliases) > 0`, nil)
	if err != nil {
		return fmt.Errorf("find performers with aliases: %w", err)
	}
	for _, performerID := range performerIDs {
		if _, err := s.mergeAliasTwins(ctx, performerID); err != nil {
			return err
		}
	}
	return nil
}

// repairTable backfills one table's name keys and merges its duplicates.
func (s *Service) repairTable(ctx context.Context, table string) (int, error) {
	if err := s.backfillNameKeys(ctx, table); err != nil {
		return 0, err
	}
	keys, err := s.database.Strings(ctx,
		fmt.Sprintf(`SELECT name_key FROM %s WHERE name_key <> '' GROUP BY name_key HAVING count(*) > 1`, table), nil)
	if err != nil {
		return 0, fmt.Errorf("find duplicate %s records: %w", table, err)
	}
	mergedCount := 0
	for _, key := range keys {
		count, err := s.mergeGroup(ctx, table, key)
		mergedCount += count
		if err != nil {
			return mergedCount, err
		}
	}
	if mergedCount > 0 {
		log.Printf("ingest: merged %d duplicate %s records", mergedCount, table)
	}
	return mergedCount, nil
}

// backfillNameKeys sets name_key on the rows of table that have none.
func (s *Service) backfillNameKeys(ctx context.Context, table string) error {
	rows, err := s.database.Query(ctx, fmt.Sprintf(`SELECT id, name FROM %s WHERE name_key IS NULL`, table), nil)
	if err != nil {
		return fmt.Errorf("find %s rows without name key: %w", table, err)
	}
	return s.database.Tx(ctx, func(tx *db.DB) error {
		for _, row := range rows {
			name, _ := row["name"].(string)
			_, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s SET name_key = $key WHERE id = $id`, table),
				db.Vars{"key": nameKey(table, name), "id": row["id"]})
			if err != nil {
				return fmt.Errorf("set %s name key: %w", table, err)
			}
		}
		return nil
	})
}

// mergeGroup merges every record of table with name key into the best one.
func (s *Service) mergeGroup(ctx context.Context, table, key string) (int, error) {
	ids, err := s.database.Strings(ctx,
		fmt.Sprintf(`SELECT id FROM %[1]s WHERE name_key = $key
		 ORDER BY %[2]s DESC,
		          (name <> lower(name) AND name <> upper(name)) DESC, created_at, id`, table, creditCount(table)),
		db.Vars{"key": key})
	if err != nil {
		return 0, fmt.Errorf("order duplicate %s records: %w", table, err)
	}
	mergedCount := 0
	for _, mergedID := range ids[1:] {
		if err := s.MergeRecords(ctx, table, ids[0], mergedID); err != nil {
			return mergedCount, err
		}
		mergedCount++
	}
	return mergedCount, nil
}
