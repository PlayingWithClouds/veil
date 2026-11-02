package library

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// MergeableKinds are the kinds two records of can be merged.
var MergeableKinds = []schema.RecordKind{
	schema.RecordKindVideo, schema.RecordKindGallery, schema.RecordKindAudio, schema.RecordKindStory, schema.RecordKindPost,
	schema.RecordKindPerformer, schema.RecordKindStudio, schema.RecordKindTag, schema.RecordKindGroup,
}

// rowsMovedBySubject are the tables whose rows follow their record into
// the survivor of a merge.
var rowsMovedBySubject = []string{"Source", "Stream", "Segment", "Subtitle", "Script", "Fingerprint", "Embedding", "FaceScan", "MediaFile", "Event"}

// Merge folds the record mergeID into keepID for good: its sources, streams,
// files, extras, user state and events move over, its credits and tags join
// the survivor's, fields the survivor lacks are filled from it, every
// reference to it now names the survivor, and it's deleted. Merges can't be
// undone, so callers decide sameness strictly (see MergeAlike).
func (l *Library) Merge(ctx context.Context, kind schema.RecordKind, keepID, mergeID string) error {
	if !slices.Contains(MergeableKinds, kind) {
		return fmt.Errorf("%s records can't be merged", kind)
	}
	if keepID == mergeID {
		return fmt.Errorf("a record can't be merged into itself")
	}
	return l.database.Tx(ctx, func(tx *db.DB) error {
		keep, merged, err := loadPair(ctx, tx, kind, keepID, mergeID)
		if err != nil {
			return err
		}
		if err := fillFromMerged(ctx, tx, kind, keep, merged); err != nil {
			return err
		}
		if err := moveSubjectRows(ctx, tx, kind, keepID, mergeID); err != nil {
			return err
		}
		if err := mergeUserState(ctx, tx, kind, keepID, mergeID); err != nil {
			return err
		}
		if err := redirectReferences(ctx, tx, kind, keepID, mergeID); err != nil {
			return err
		}
		if err := deleteDependents(ctx, tx, kind, mergeID); err != nil {
			return err
		}
		return tx.Delete(ctx, tableFor(kind), mergeID)
	})
}

// loadPair loads both records of a merge as raw rows.
func loadPair(ctx context.Context, tx *db.DB, kind schema.RecordKind, keepID, mergeID string) (db.Row, db.Row, error) {
	query := `SELECT * FROM ` + db.Quote(tableFor(kind)) + ` WHERE "id" = $id`
	keep, err := tx.QueryRow(ctx, query, db.Vars{"id": keepID})
	if err != nil || keep == nil {
		return nil, nil, fmt.Errorf("no %s %s", kind, keepID)
	}
	merged, err := tx.QueryRow(ctx, query, db.Vars{"id": mergeID})
	if err != nil || merged == nil {
		return nil, nil, fmt.Errorf("no %s %s", kind, mergeID)
	}
	return keep, merged, nil
}

// mergeKeptColumns are never copied from the merged record.
var mergeKeptColumns = map[string]bool{"id": true, "createdAt": true, "updatedAt": true, "deletedAt": true, "nameKey": true, "name": true, "title": true}

// unionColumns are id lists whose entries both records keep.
var unionColumns = map[string]bool{"performers": true, "tags": true, "aliases": true, "links": true, "attachments": true}

// fillFromMerged sets the survivor's empty fields from the merged record and
// joins their lists; a performer, studio or tag also keeps the merged name
// as an alias.
func fillFromMerged(ctx context.Context, tx *db.DB, kind schema.RecordKind, keep, merged db.Row) error {
	fields := db.Vars{}
	for column, value := range merged {
		if mergeKeptColumns[column] || value == nil {
			continue
		}
		if unionColumns[column] {
			fields[column] = json.RawMessage(unionJSON(keep[column], value))
			continue
		}
		if keep[column] == nil {
			fields[column] = value
		}
	}
	if name, named := merged["name"].(string); named && name != keep["name"] {
		aliases := fields["aliases"]
		if aliases == nil {
			aliases = keep["aliases"]
		}
		fields["aliases"] = json.RawMessage(unionJSON(aliases, fmt.Sprintf("[%q]", name)))
	}
	if len(fields) == 0 {
		return nil
	}
	return tx.Update(ctx, tableFor(kind), keep["id"].(string), fields)
}

// unionJSON joins two JSON arrays (texts, or a RawMessage from an earlier
// union), keeping the first's order and dropping repeats.
func unionJSON(first, second any) string {
	joined := []any{}
	seen := map[string]bool{}
	for _, list := range []any{first, second} {
		for _, element := range decodeJSONList(list) {
			key := fmt.Sprint(element)
			if seen[key] {
				continue
			}
			seen[key] = true
			joined = append(joined, element)
		}
	}
	encoded, _ := db.StoredValue(joined).(string)
	return encoded
}

// decodeJSONList reads a JSON array column value; anything else is empty.
func decodeJSONList(value any) []any {
	text := ""
	switch typed := value.(type) {
	case string:
		text = typed
	case json.RawMessage:
		text = string(typed)
	}
	list := []any{}
	if text == "" {
		return list
	}
	if json.Unmarshal([]byte(text), &list) != nil {
		return []any{}
	}
	return list
}

// moveSubjectRows points the merged record's own rows at the survivor.
// Comments come along with their sources; user state and collection items
// are merged separately because they're one per record.
func moveSubjectRows(ctx context.Context, tx *db.DB, kind schema.RecordKind, keepID, mergeID string) error {
	vars := db.Vars{"kind": kind, "keep": keepID, "merge": mergeID}
	statements := []string{}
	for _, table := range rowsMovedBySubject {
		statements = append(statements, `UPDATE `+db.Quote(table)+` SET "subjectId" = $keep WHERE "subjectKind" = $kind AND "subjectId" = $merge`)
	}
	statements = append(statements,
		`DELETE FROM "CollectionItem" WHERE "subjectKind" = $kind AND "subjectId" = $merge AND "collection" IN
		   (SELECT "collection" FROM "CollectionItem" WHERE "subjectKind" = $kind AND "subjectId" = $keep)`,
		`UPDATE "CollectionItem" SET "subjectId" = $keep WHERE "subjectKind" = $kind AND "subjectId" = $merge`,
		`DELETE FROM "Alike" WHERE "subjectKind" = $kind AND ("firstId" = $merge OR "secondId" = $merge)`,
	)
	return execAll(ctx, tx, statements, vars)
}

// mergeUserState combines both records' user state into the survivor's:
// either one's rating, favourite and block, both views, both notes.
func mergeUserState(ctx context.Context, tx *db.DB, kind schema.RecordKind, keepID, mergeID string) error {
	vars := db.Vars{"kind": kind, "keep": keepID, "merge": mergeID}
	merged, err := tx.Int(ctx, `SELECT COUNT(*) FROM "UserState" WHERE "subjectKind" = $kind AND "subjectId" = $merge`, vars)
	if err != nil || merged == 0 {
		return err
	}
	kept, err := tx.Int(ctx, `SELECT COUNT(*) FROM "UserState" WHERE "subjectKind" = $kind AND "subjectId" = $keep`, vars)
	if err != nil {
		return err
	}
	if kept == 0 {
		_, err = tx.Exec(ctx, `UPDATE "UserState" SET "subjectId" = $keep WHERE "subjectKind" = $kind AND "subjectId" = $merge`, vars)
		return err
	}
	return execAll(ctx, tx, []string{combineStates, `DELETE FROM "UserState" WHERE "subjectKind" = $kind AND "subjectId" = $merge`}, vars)
}

// combineStates folds the merged record's state row into the survivor's.
const combineStates = `UPDATE "UserState" AS "kept" SET
	"rating" = COALESCE("kept"."rating", "other"."rating"),
	"favorite" = MAX("kept"."favorite", "other"."favorite"),
	"blocked" = MAX("kept"."blocked", "other"."blocked"),
	"notes" = CASE WHEN "kept"."notes" IS NULL THEN "other"."notes"
		WHEN "other"."notes" IS NULL THEN "kept"."notes"
		ELSE "kept"."notes" || char(10) || "other"."notes" END,
	"progress" = COALESCE("kept"."progress", "other"."progress"),
	"viewCount" = "kept"."viewCount" + "other"."viewCount",
	"lastViewedAt" = MAX(COALESCE("kept"."lastViewedAt", ''), COALESCE("other"."lastViewedAt", '')),
	"fields" = json_patch("other"."fields", "kept"."fields")
	FROM (SELECT * FROM "UserState" WHERE "subjectKind" = $kind AND "subjectId" = $merge) AS "other"
	WHERE "kept"."subjectKind" = $kind AND "kept"."subjectId" = $keep`

// redirectReferences makes every record that named the merged one name the
// survivor.
func redirectReferences(ctx context.Context, tx *db.DB, kind schema.RecordKind, keepID, mergeID string) error {
	vars := db.Vars{"kind": kind, "keep": keepID, "merge": mergeID}
	statements := []string{redirectAttachments}
	switch kind {
	case schema.RecordKindVideo:
		statements = append(statements,
			`UPDATE "Gallery" SET "video" = $keep WHERE "video" = $merge`,
			`UPDATE "Clip" SET "video" = $keep WHERE "video" = $merge`,
			`DELETE FROM "GroupItem" WHERE "video" = $merge AND "group" IN (SELECT "group" FROM "GroupItem" WHERE "video" = $keep)`,
			`UPDATE "GroupItem" SET "video" = $keep WHERE "video" = $merge`)
	case schema.RecordKindGallery:
		statements = append(statements, `UPDATE "Video" SET "gallery" = $keep WHERE "gallery" = $merge`)
	case schema.RecordKindStudio:
		statements = append(statements, redirectColumn([]string{"Studio"}, "parent")...)
		statements = append(statements, redirectColumn(append([]string{"Group"}, mediaTables...), "studio")...)
	case schema.RecordKindPerformer:
		statements = append(statements, redirectColumn([]string{"Post"}, "creator")...)
		statements = append(statements, redirectColumn([]string{"CamModel"}, "performer")...)
		statements = append(statements, redirectInLists(append([]string{"Segment", "Comment"}, mediaTables...), "performers", "mentions")...)
	case schema.RecordKindTag:
		statements = append(statements, redirectColumn([]string{"Segment"}, "tag")...)
		statements = append(statements, redirectInLists(append([]string{"Performer", "Studio"}, mediaTables...), "tags")...)
		statements = append(statements, redirectInLists([]string{"Category"}, "matchTags")...)
	case schema.RecordKindGroup:
		statements = append(statements,
			`DELETE FROM "GroupItem" WHERE "group" = $merge AND "video" IN (SELECT "video" FROM "GroupItem" WHERE "group" = $keep)`,
			`UPDATE "GroupItem" SET "group" = $keep WHERE "group" = $merge`)
	}
	return execAll(ctx, tx, statements, vars)
}

// redirectAttachments makes posts that attach the merged record attach the
// survivor.
const redirectAttachments = `UPDATE "Post" SET "attachments" =
	(SELECT json_group_array(CASE
		WHEN json_extract(value, '$.kind') = $kind AND json_extract(value, '$.id') = $merge
		THEN json_object('kind', $kind, 'id', $keep) ELSE json(value) END) FROM json_each("attachments"))
	WHERE EXISTS (SELECT 1 FROM json_each("attachments")
		WHERE json_extract(value, '$.kind') = $kind AND json_extract(value, '$.id') = $merge)`

// redirectColumn builds the statements that point a reference column at the
// survivor.
func redirectColumn(tables []string, column string) []string {
	statements := make([]string, 0, len(tables))
	for _, table := range tables {
		statements = append(statements, `UPDATE `+db.Quote(table)+` SET `+db.Quote(column)+` = $keep WHERE `+db.Quote(column)+` = $merge`)
	}
	return statements
}

// redirectInLists builds the statements that replace the merged id with the
// survivor's in JSON id lists, without repeating the survivor. Each table
// gets the first of columns it has; Comment has mentions, the rest the first.
func redirectInLists(tables []string, columns ...string) []string {
	statements := []string{}
	for _, table := range tables {
		column := columns[0]
		if table == "Comment" && len(columns) > 1 {
			column = columns[1]
		}
		quoted := db.Quote(column)
		statements = append(statements, `UPDATE `+db.Quote(table)+` SET `+quoted+` =
			(SELECT json_group_array("id") FROM (SELECT DISTINCT CASE WHEN value = $merge THEN $keep ELSE value END AS "id"
			 FROM json_each(`+quoted+`) ORDER BY "key"))
			WHERE EXISTS (SELECT 1 FROM json_each(`+quoted+`) WHERE value = $merge)`)
	}
	return statements
}
