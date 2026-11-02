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

// Merge folds the record mergeID into keepID: its sources, streams, files,
// extras, user state and events move over, its credits and tags join the
// survivor's, fields the survivor lacks are filled from it, every reference
// to it now names the survivor, and it's deleted. Every row it changes is
// journaled (MergeJournal), so UndoMerge can bring the record back.
// Automatic matching still decides sameness strictly (see MergeAlike).
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
		changes := &journal{}
		if err := fillFromMerged(ctx, tx, changes, kind, keep, merged); err != nil {
			return err
		}
		vars := db.Vars{"kind": kind, "keep": keepID, "merge": mergeID, "id": mergeID}
		steps := moveSubjectRows()
		steps = append(steps, mergeUserState()...)
		steps = append(steps, redirectReferences(kind)...)
		steps = append(steps, dependents(kind)...)
		steps = append(steps, deleteRows(tableFor(kind), `"id" = $merge`))
		if err := changes.apply(ctx, tx, steps, vars); err != nil {
			return err
		}
		return changes.save(ctx, tx, kind, keepID, merged)
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
func fillFromMerged(ctx context.Context, tx *db.DB, changes *journal, kind schema.RecordKind, keep, merged db.Row) error {
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
	return changes.update(ctx, tx, tableFor(kind), keep["id"].(string), fields)
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
func moveSubjectRows() []change {
	const merged = `"subjectKind" = $kind AND "subjectId" = $merge`
	changes := []change{}
	for _, table := range rowsMovedBySubject {
		changes = append(changes, updateRows(table, `"subjectId" = $keep`, merged))
	}
	return append(changes,
		deleteRows("CollectionItem", merged+` AND "collection" IN
		   (SELECT "collection" FROM "CollectionItem" WHERE "subjectKind" = $kind AND "subjectId" = $keep)`),
		updateRows("CollectionItem", `"subjectId" = $keep`, merged),
		deleteRows("Alike", `"subjectKind" = $kind AND ("firstId" = $merge OR "secondId" = $merge)`),
	)
}

// mergeUserState combines both records' user state into the survivor's:
// either one's rating, favourite and block, both views, both notes. Without
// a state of its own the survivor takes the merged one's.
func mergeUserState() []change {
	const merged = `"subjectKind" = $kind AND "subjectId" = $merge`
	return []change{
		{table: "UserState", where: `"subjectKind" = $kind AND "subjectId" = $keep`, statement: combineStates},
		deleteRows("UserState", merged+` AND EXISTS
		   (SELECT 1 FROM "UserState" AS "kept" WHERE "kept"."subjectKind" = $kind AND "kept"."subjectId" = $keep)`),
		updateRows("UserState", `"subjectId" = $keep`, merged),
	}
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
func redirectReferences(kind schema.RecordKind) []change {
	changes := []change{redirectAttachments}
	switch kind {
	case schema.RecordKindVideo:
		changes = append(changes,
			updateRows("Gallery", `"video" = $keep`, `"video" = $merge`),
			updateRows("Clip", `"video" = $keep`, `"video" = $merge`),
			deleteRows("GroupItem", `"video" = $merge AND "group" IN (SELECT "group" FROM "GroupItem" WHERE "video" = $keep)`),
			updateRows("GroupItem", `"video" = $keep`, `"video" = $merge`))
	case schema.RecordKindGallery:
		changes = append(changes, updateRows("Video", `"gallery" = $keep`, `"gallery" = $merge`))
	case schema.RecordKindStudio:
		changes = append(changes, redirectColumn([]string{"Studio"}, "parent")...)
		changes = append(changes, redirectColumn(append([]string{"Group"}, mediaTables...), "studio")...)
	case schema.RecordKindPerformer:
		changes = append(changes, redirectColumn([]string{"Post"}, "creator")...)
		changes = append(changes, redirectColumn([]string{"CamModel"}, "performer")...)
		changes = append(changes, redirectInLists(append([]string{"Segment", "Comment"}, mediaTables...), "performers", "mentions")...)
	case schema.RecordKindTag:
		changes = append(changes, redirectColumn([]string{"Segment"}, "tag")...)
		changes = append(changes, redirectInLists(append([]string{"Performer", "Studio"}, mediaTables...), "tags")...)
		changes = append(changes, redirectInLists([]string{"Category"}, "matchTags")...)
	case schema.RecordKindGroup:
		changes = append(changes,
			deleteRows("GroupItem", `"group" = $merge AND "video" IN (SELECT "video" FROM "GroupItem" WHERE "group" = $keep)`),
			updateRows("GroupItem", `"group" = $keep`, `"group" = $merge`))
	}
	return changes
}

// redirectAttachments makes posts that attach the merged record attach the
// survivor.
var redirectAttachments = updateRows("Post", `"attachments" =
	(SELECT json_group_array(CASE
		WHEN json_extract(value, '$.kind') = $kind AND json_extract(value, '$.id') = $merge
		THEN json_object('kind', $kind, 'id', $keep) ELSE json(value) END) FROM json_each("attachments"))`,
	`EXISTS (SELECT 1 FROM json_each("attachments")
		WHERE json_extract(value, '$.kind') = $kind AND json_extract(value, '$.id') = $merge)`)

// redirectColumn builds the changes that point a reference column at the
// survivor.
func redirectColumn(tables []string, column string) []change {
	changes := make([]change, 0, len(tables))
	quoted := db.Quote(column)
	for _, table := range tables {
		changes = append(changes, updateRows(table, quoted+` = $keep`, quoted+` = $merge`))
	}
	return changes
}

// redirectInLists builds the changes that replace the merged id with the
// survivor's in JSON id lists, without repeating the survivor. Each table
// gets the first of columns it has; Comment has mentions, the rest the first.
func redirectInLists(tables []string, columns ...string) []change {
	changes := []change{}
	for _, table := range tables {
		column := columns[0]
		if table == "Comment" && len(columns) > 1 {
			column = columns[1]
		}
		quoted := db.Quote(column)
		changes = append(changes, updateRows(table, quoted+` =
			(SELECT json_group_array("id") FROM (SELECT DISTINCT CASE WHEN value = $merge THEN $keep ELSE value END AS "id"
			 FROM json_each(`+quoted+`) ORDER BY "key"))`,
			`EXISTS (SELECT 1 FROM json_each(`+quoted+`) WHERE value = $merge)`))
	}
	return changes
}
