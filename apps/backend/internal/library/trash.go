package library

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TrashableKinds are the kinds the user can trash. Images, sources and
// streams belong to the site and go with their record.
var TrashableKinds = []schema.RecordKind{
	schema.RecordKindVideo,
	schema.RecordKindGallery,
	schema.RecordKindAudio,
	schema.RecordKindStory,
	schema.RecordKindPost,
	schema.RecordKindPerformer,
	schema.RecordKindStudio,
	schema.RecordKindGroup,
	schema.RecordKindTag,
	schema.RecordKindCategory,
	schema.RecordKindCollection,
	schema.RecordKindSegment,
	schema.RecordKindClip,
}

// mediaTables are the tables of every media kind, for clearing references.
var mediaTables = []string{"Video", "Gallery", "Audio", "Story", "Post"}

// namedKinds are the kinds labelled by a name rather than a title.
var namedKinds = []schema.RecordKind{
	schema.RecordKindPerformer, schema.RecordKindStudio, schema.RecordKindTag,
	schema.RecordKindCategory, schema.RecordKindCollection,
}

// TrashedRecord is one record in the trash.
type TrashedRecord struct {
	Kind      schema.RecordKind
	ID        string
	Label     string
	DeletedAt string
}

// Trash moves a record to the trash: it disappears from every read until
// restored, and is purged once the retention passes.
func (l *Library) Trash(ctx context.Context, kind schema.RecordKind, id string) error {
	if !isTrashable(kind) {
		return fmt.Errorf("a %s can't be trashed", kind)
	}
	_, err := l.recorder.Update(ctx, kind, id, db.Vars{"deletedAt": db.Now()}, provenance.User())
	return err
}

// Restore takes a record out of the trash.
func (l *Library) Restore(ctx context.Context, kind schema.RecordKind, id string) error {
	if !isTrashable(kind) {
		return fmt.Errorf("a %s can't be trashed", kind)
	}
	_, err := l.recorder.Update(ctx, kind, id, db.Vars{"deletedAt": nil}, provenance.User())
	return err
}

// isTrashable reports whether kind is one of TrashableKinds.
func isTrashable(kind schema.RecordKind) bool {
	return slices.Contains(TrashableKinds, kind)
}

// Trashed lists the trashed records of kind, most recently trashed first.
func (l *Library) Trashed(ctx context.Context, kind schema.RecordKind, limit, offset int) ([]TrashedRecord, error) {
	if !isTrashable(kind) {
		return nil, fmt.Errorf("a %s can't be trashed", kind)
	}
	vars := db.Vars{}
	vars["limit"], vars["offset"] = pageBounds(limit, offset)
	rows, err := l.database.Query(ctx,
		`SELECT "id", COALESCE(`+labelColumn(kind)+`, '') AS "label", "deletedAt" FROM `+db.Quote(tableFor(kind))+`
		 WHERE "deletedAt" IS NOT NULL ORDER BY "deletedAt" DESC LIMIT $limit OFFSET $offset`, vars)
	if err != nil {
		return nil, err
	}
	trashed := make([]TrashedRecord, 0, len(rows))
	for _, row := range rows {
		trashed = append(trashed, TrashedRecord{Kind: kind, ID: row["id"].(string), Label: fmt.Sprint(row["label"]), DeletedAt: row["deletedAt"].(string)})
	}
	return trashed, nil
}

// labelColumn is the column that names a record of kind.
func labelColumn(kind schema.RecordKind) string {
	if slices.Contains(namedKinds, kind) {
		return `"name"`
	}
	return `"title"`
}

// PurgeTrash removes for good every record trashed before cutoff, with what
// hangs off it, and returns how many it removed.
func (l *Library) PurgeTrash(ctx context.Context, cutoff time.Time) (int, error) {
	purged := 0
	for _, kind := range TrashableKinds {
		ids, err := l.database.Strings(ctx,
			`SELECT "id" FROM `+db.Quote(tableFor(kind))+` WHERE "deletedAt" IS NOT NULL AND "deletedAt" < $cutoff`,
			db.Vars{"cutoff": db.FormatTime(cutoff)})
		if err != nil {
			return purged, err
		}
		for _, id := range ids {
			if err := l.Purge(ctx, kind, id); err != nil {
				return purged, err
			}
			purged++
		}
	}
	return purged, nil
}

// PurgeTrashed removes one trashed record for good, now. A record that isn't
// in the trash is refused, so nothing live is purged by mistake.
func (l *Library) PurgeTrashed(ctx context.Context, kind schema.RecordKind, id string) error {
	if !isTrashable(kind) {
		return fmt.Errorf("a %s can't be trashed", kind)
	}
	trashed, err := l.database.Int(ctx,
		`SELECT COUNT(*) FROM `+db.Quote(tableFor(kind))+` WHERE "id" = $id AND "deletedAt" IS NOT NULL`, db.Vars{"id": id})
	if err != nil {
		return err
	}
	if trashed == 0 {
		return fmt.Errorf("%s %s isn't in the trash", kind, id)
	}
	return l.Purge(ctx, kind, id)
}

// Purge removes one record for good: its sources and streams, the user's
// state, provenance and revisions, and every reference to it from other
// records. The event log keeps its entries; stored files stay on disk.
func (l *Library) Purge(ctx context.Context, kind schema.RecordKind, id string) error {
	return l.database.Tx(ctx, func(tx *db.DB) error {
		return purgeIn(ctx, tx, kind, id)
	})
}

// purgeIn purges one record inside an open transaction.
func purgeIn(ctx context.Context, tx *db.DB, kind schema.RecordKind, id string) error {
	if err := purgeCreatorPosts(ctx, tx, kind, id); err != nil {
		return err
	}
	if err := clearReferences(ctx, tx, kind, id); err != nil {
		return err
	}
	if err := deleteDependents(ctx, tx, kind, id); err != nil {
		return err
	}
	return tx.Delete(ctx, tableFor(kind), id)
}

// purgeCreatorPosts purges a performer's posts first: a post can't exist
// without its creator.
func purgeCreatorPosts(ctx context.Context, tx *db.DB, kind schema.RecordKind, id string) error {
	if kind != schema.RecordKindPerformer {
		return nil
	}
	postIDs, err := tx.Strings(ctx, `SELECT "id" FROM "Post" WHERE "creator" = $id`, db.Vars{"id": id})
	if err != nil {
		return err
	}
	for _, postID := range postIDs {
		if err := purgeIn(ctx, tx, schema.RecordKindPost, postID); err != nil {
			return err
		}
	}
	return nil
}

// subjectTables are the tables whose rows exist for one record they name by
// subjectKind and subjectId.
var subjectTables = []string{
	"Stream", "Revision", "UserState", "CollectionItem", "Segment", "Subtitle", "Script",
	"Fingerprint", "Embedding", "MediaFile",
}

// deleteDependents deletes the rows that only exist for the record.
func deleteDependents(ctx context.Context, tx *db.DB, kind schema.RecordKind, id string) error {
	vars := db.Vars{"kind": kind, "id": id}
	statements := []string{
		`DELETE FROM "FieldSource" WHERE ("subjectKind" = $kind AND "subjectId" = $id)
		   OR "source" IN (SELECT "id" FROM "Source" WHERE "subjectKind" = $kind AND "subjectId" = $id)`,
		`DELETE FROM "Comment" WHERE "source" IN (SELECT "id" FROM "Source" WHERE "subjectKind" = $kind AND "subjectId" = $id)`,
		`DELETE FROM "Alike" WHERE "subjectKind" = $kind AND ("firstId" = $id OR "secondId" = $id)`,
	}
	for _, table := range subjectTables {
		statements = append(statements, `DELETE FROM `+db.Quote(table)+` WHERE "subjectKind" = $kind AND "subjectId" = $id`)
	}
	statements = append(statements, `DELETE FROM "Source" WHERE "subjectKind" = $kind AND "subjectId" = $id`)
	statements = append(statements, kindDependents[kind]...)
	return execAll(ctx, tx, statements, vars)
}

// kindDependents are the rows only one kind of record has.
var kindDependents = map[schema.RecordKind][]string{
	schema.RecordKindGallery:    {`DELETE FROM "Image" WHERE "gallery" = $id`},
	schema.RecordKindVideo:      {`DELETE FROM "GroupItem" WHERE "video" = $id`, `DELETE FROM "Clip" WHERE "video" = $id`},
	schema.RecordKindStory:      {`DELETE FROM "StoryChapter" WHERE "story" = $id`},
	schema.RecordKindGroup:      {`DELETE FROM "GroupItem" WHERE "group" = $id`},
	schema.RecordKindCollection: {`DELETE FROM "CollectionItem" WHERE "collection" = $id`},
}

// clearReferences removes the record's id from the records that point at it.
func clearReferences(ctx context.Context, tx *db.DB, kind schema.RecordKind, id string) error {
	vars := db.Vars{"id": id, "kind": kind}
	statements := []string{removeFromAttachments}
	switch kind {
	case schema.RecordKindVideo:
		statements = append(statements, `UPDATE "Gallery" SET "video" = NULL WHERE "video" = $id`)
	case schema.RecordKindGallery:
		statements = append(statements, `UPDATE "Video" SET "gallery" = NULL WHERE "gallery" = $id`)
	case schema.RecordKindStudio:
		statements = append(statements, studioReferences()...)
	case schema.RecordKindPerformer:
		statements = append(statements, performerReferences()...)
	case schema.RecordKindTag:
		statements = append(statements, tagReferences()...)
	case schema.RecordKindCategory:
		statements = append(statements, `UPDATE "Category" SET "parent" = NULL WHERE "parent" = $id`)
	case schema.RecordKindCamModel:
		statements = append(statements, `UPDATE "Video" SET "recordedFrom" = NULL WHERE "recordedFrom" = $id`)
	}
	return execAll(ctx, tx, statements, vars)
}

// removeFromAttachments drops the record from the posts it's attached to.
const removeFromAttachments = `UPDATE "Post" SET "attachments" =
	(SELECT json_group_array(json(value)) FROM json_each("attachments")
	 WHERE NOT (json_extract(value, '$.kind') = $kind AND json_extract(value, '$.id') = $id))
	WHERE EXISTS (SELECT 1 FROM json_each("attachments")
	 WHERE json_extract(value, '$.kind') = $kind AND json_extract(value, '$.id') = $id)`

// studioReferences clear a studio from every record that names it.
func studioReferences() []string {
	statements := []string{`UPDATE "Studio" SET "parent" = NULL WHERE "parent" = $id`, `UPDATE "Group" SET "studio" = NULL WHERE "studio" = $id`}
	for _, table := range mediaTables {
		statements = append(statements, `UPDATE `+db.Quote(table)+` SET "studio" = NULL WHERE "studio" = $id`)
	}
	return statements
}

// performerReferences clear a performer from credits, cam rooms, segments
// and comment mentions.
func performerReferences() []string {
	statements := []string{`UPDATE "CamModel" SET "performer" = NULL WHERE "performer" = $id`}
	statements = append(statements, removeFromLists(append([]string{"Segment"}, mediaTables...), "performers")...)
	return append(statements, removeFromLists([]string{"Comment"}, "mentions")...)
}

// tagReferences clear a tag from every tag list and category.
func tagReferences() []string {
	statements := []string{`UPDATE "Segment" SET "tag" = NULL WHERE "tag" = $id`}
	statements = append(statements, removeFromLists(append([]string{"Performer", "Studio"}, mediaTables...), "tags")...)
	return append(statements, removeFromLists([]string{"Category"}, "matchTags")...)
}

// removeFromLists builds the statements that drop $id from a JSON id list
// column of each table.
func removeFromLists(tables []string, column string) []string {
	quoted := db.Quote(column)
	statements := make([]string, 0, len(tables))
	for _, table := range tables {
		statements = append(statements, `UPDATE `+db.Quote(table)+` SET `+quoted+` =
			(SELECT json_group_array(value) FROM json_each(`+quoted+`) WHERE value != $id)
			WHERE EXISTS (SELECT 1 FROM json_each(`+quoted+`) WHERE value = $id)`)
	}
	return statements
}

// execAll runs statements in order, stopping at the first error.
func execAll(ctx context.Context, tx *db.DB, statements []string, vars db.Vars) error {
	for _, statement := range statements {
		if _, err := tx.Exec(ctx, statement, vars); err != nil {
			return err
		}
	}
	return nil
}
