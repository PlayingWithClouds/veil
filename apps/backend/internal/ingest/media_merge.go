package ingest

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
)

// mergeableMediaTables are the content tables whose records can be merged when
// several sites carry the same thing. The merged-away record's source_url and
// external_id stay resolvable (see entity_identity), so a later listing of that
// page lands on the survivor.
var mergeableMediaTables = map[string]bool{"scene": true, "gallery": true, "collection": true}

// listColumns are the JSON id/URL lists a media merge unites, per table.
var listColumns = map[string][]string{
	"scene":      {"performers", "tags", "preview_images"},
	"gallery":    {"performers", "tags"},
	"collection": {"tags"},
}

// MergeMedia folds the scene, gallery or collection mergedID into survivorID in
// one transaction: streams, watch history, ratings, markers, memberships and
// every other reference move to the survivor, fields it lacks are copied over,
// the merged record's site identities stay resolvable and the record is deleted.
func (s *Service) MergeMedia(ctx context.Context, table, survivorID, mergedID string) error {
	if !mergeableMediaTables[table] {
		return fmt.Errorf("merge media: %s has no merge", table)
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
		if _, err := tx.Exec(ctx, fillStatement(table), vars); err != nil {
			return fmt.Errorf("fill %s fields: %w", table, err)
		}
		if err := keepMediaIdentities(ctx, tx, merged, vars); err != nil {
			return err
		}
		updates := map[string]any{}
		for _, column := range listColumns[table] {
			updates[column] = unionExact(jsonStringList(survivor[column]), jsonStringList(merged[column]))
		}
		survivorRecord, err := db.ParseRecordID(survivorID)
		if err != nil {
			return err
		}
		if err := tx.Merge(ctx, *survivorRecord, updates); err != nil {
			return fmt.Errorf("merge %s lists: %w", table, err)
		}
		for _, statement := range mediaRepointStatements(table) {
			if _, err := tx.Exec(ctx, statement, vars); err != nil {
				return fmt.Errorf("merge %s %s into %s: %w", table, mergedID, survivorID, err)
			}
		}
		return nil
	})
}

// keepMediaIdentities registers the merged record's source_url and external_id
// (and those it had registered itself) as identities of the survivor.
func keepMediaIdentities(ctx context.Context, tx *db.DB, merged db.Row, vars db.Vars) error {
	if _, err := tx.Exec(ctx, `UPDATE entity_identity SET record = $survivor WHERE record = $merged`, vars); err != nil {
		return fmt.Errorf("repoint identities: %w", err)
	}
	for _, kind := range identityKinds {
		value, _ := merged[kind].(string)
		if value == "" {
			continue
		}
		_, err := tx.Exec(ctx,
			`INSERT OR IGNORE INTO entity_identity (kind, value, record) VALUES ($kind, $value, $survivor)`,
			db.Vars{"kind": kind, "value": value, "survivor": vars["survivor"]})
		if err != nil {
			return fmt.Errorf("keep %s identity: %w", kind, err)
		}
	}
	return nil
}

// unionExact appends the values of extra missing from base, comparing exactly.
func unionExact(base, extra []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(base)+len(extra))
	for _, value := range append(append([]string{}, base...), extra...) {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

// mediaRepointStatements move everything that refers to $merged onto $survivor
// and finally delete $merged. Tables with a unique target keep the survivor's
// row when both exist.
func mediaRepointStatements(table string) []string {
	statements := specificMediaStatements(table)
	statements = append(statements,
		`UPDATE observation SET target = $survivor WHERE target = $merged`,
		`UPDATE changes SET canonical = $survivor WHERE canonical = $merged`,
		`UPDATE job SET target = $survivor WHERE target = $merged`,
		`UPDATE OR IGNORE collection_item SET media = $survivor WHERE media = $merged`,
		`DELETE FROM collection_item WHERE media = $merged`,
		`UPDATE OR IGNORE user_rating SET media = $survivor WHERE media = $merged`,
		`DELETE FROM user_rating WHERE media = $merged`,
		`UPDATE OR IGNORE watchlist SET media = $survivor WHERE media = $merged`,
		`DELETE FROM watchlist WHERE media = $merged`,
		`UPDATE OR IGNORE search_subscription_item SET media = $survivor WHERE media = $merged`,
		`DELETE FROM search_subscription_item WHERE media = $merged`,
		`UPDATE o_event SET media = $survivor WHERE media = $merged`,
		`UPDATE watch_history SET media = $survivor WHERE media = $merged`,
		`UPDATE recommendation_impression SET media = $survivor WHERE media = $merged`,
		`UPDATE search_event SET media = $survivor WHERE media = $merged`,
		fmt.Sprintf(`DELETE FROM %s WHERE id = $merged`, table),
	)
	return statements
}

// specificMediaStatements are the references only one kind of record has.
func specificMediaStatements(table string) []string {
	switch table {
	case "scene":
		return []string{
			`UPDATE OR IGNORE stream SET media = $survivor WHERE media = $merged`,
			`DELETE FROM stream WHERE media = $merged`,
			`UPDATE OR IGNORE performs_in SET media = $survivor WHERE media = $merged`,
			`DELETE FROM performs_in WHERE media = $merged`,
			`UPDATE scene_marker SET media = $survivor WHERE media = $merged`,
			`UPDATE image SET media = $survivor WHERE media = $merged`,
			// Heat buckets add up when both scenes were watched.
			`INSERT INTO scene_heat (scene, bucket, watched, scrubbed)
				SELECT $survivor, bucket, watched, scrubbed FROM scene_heat WHERE scene = $merged
				ON CONFLICT (scene, bucket) DO UPDATE SET
					watched = scene_heat.watched + excluded.watched,
					scrubbed = scene_heat.scrubbed + excluded.scrubbed`,
			`DELETE FROM scene_heat WHERE scene = $merged`,
			`UPDATE OR IGNORE scene_related SET scene = $survivor WHERE scene = $merged`,
			`UPDATE OR IGNORE scene_related SET related = $survivor WHERE related = $merged`,
			`DELETE FROM scene_related WHERE scene = $merged OR related = $merged OR scene = related`,
			`UPDATE gallery_video SET scene = $survivor WHERE scene = $merged`,
			`UPDATE OR IGNORE feed_discovery SET scene = $survivor WHERE scene = $merged`,
			`DELETE FROM feed_discovery WHERE scene = $merged`,
			`UPDATE preference_event SET chosen = $survivor WHERE chosen = $merged`,
			`UPDATE preference_event SET rejected = $survivor WHERE rejected = $merged`,
		}
	case "gallery":
		return append(galleryImageStatements(),
			`UPDATE OR IGNORE gallery_video SET gallery = $survivor WHERE gallery = $merged`,
			`DELETE FROM gallery_video WHERE gallery = $merged`,
			`UPDATE OR IGNORE performs_in SET media = $survivor WHERE media = $merged`,
			`DELETE FROM performs_in WHERE media = $merged`)
	case "collection":
		return []string{
			`UPDATE OR IGNORE collection_item SET collection = $survivor WHERE collection = $merged`,
			`DELETE FROM collection_item WHERE collection = $merged`,
			`UPDATE image SET media = $survivor WHERE media = $merged`,
		}
	}
	return nil
}

// galleryImageStatements move a merged gallery's images over when the survivor
// has none of its own. When it has, the merged gallery's pictures show the same
// set again, so the ones nobody saved or rated go; the rest stay attached.
func galleryImageStatements() []string {
	return []string{
		`DELETE FROM image WHERE media = $merged AND content
			AND EXISTS (SELECT 1 FROM image AS kept WHERE kept.media = $survivor AND kept.content)
			AND id NOT IN (SELECT media FROM collection_item)
			AND id NOT IN (SELECT media FROM user_rating)`,
		`UPDATE image SET media = $survivor WHERE media = $merged`,
	}
}

// MergeByID folds the record mergedID into survivorID, both of the same kind
// (performer, studio, tag, scene, gallery or collection), for a merge the user
// asked for.
func (s *Service) MergeByID(ctx context.Context, survivorID, mergedID string) error {
	table := TableFromID(survivorID)
	if table != TableFromID(mergedID) {
		return fmt.Errorf("merge: %s and %s are different kinds of record", survivorID, mergedID)
	}
	if nameKeyedTables[table] {
		return s.MergeRecords(ctx, table, survivorID, mergedID)
	}
	return s.MergeMedia(ctx, table, survivorID, mergedID)
}
