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
	schema.RecordKindPerformer,
	schema.RecordKindStudio,
	schema.RecordKindTag,
}

// mediaTables are the tables of every media kind, for clearing references.
var mediaTables = []string{"Video", "Gallery", "Audio", "Story", "Post"}

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
		`SELECT "id", `+labelColumn(kind)+` AS "label", "deletedAt" FROM `+db.Quote(tableFor(kind))+`
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
	if kind == schema.RecordKindVideo || kind == schema.RecordKindGallery {
		return `"title"`
	}
	return `"name"`
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
// records. The event log keeps its entries.
func (l *Library) Purge(ctx context.Context, kind schema.RecordKind, id string) error {
	return l.database.Tx(ctx, func(tx *db.DB) error {
		if err := clearReferences(ctx, tx, kind, id); err != nil {
			return err
		}
		if err := deleteDependents(ctx, tx, kind, id); err != nil {
			return err
		}
		return tx.Delete(ctx, tableFor(kind), id)
	})
}

// deleteDependents deletes the rows that only exist for the record.
func deleteDependents(ctx context.Context, tx *db.DB, kind schema.RecordKind, id string) error {
	vars := db.Vars{"kind": kind, "id": id}
	statements := []string{
		`DELETE FROM "FieldSource" WHERE ("subjectKind" = $kind AND "subjectId" = $id)
		   OR "source" IN (SELECT "id" FROM "Source" WHERE "subjectKind" = $kind AND "subjectId" = $id)`,
		`DELETE FROM "Comment" WHERE "source" IN (SELECT "id" FROM "Source" WHERE "subjectKind" = $kind AND "subjectId" = $id)`,
		`DELETE FROM "Stream" WHERE "subjectKind" = $kind AND "subjectId" = $id`,
		`DELETE FROM "Source" WHERE "subjectKind" = $kind AND "subjectId" = $id`,
		`DELETE FROM "Revision" WHERE "subjectKind" = $kind AND "subjectId" = $id`,
		`DELETE FROM "UserState" WHERE "subjectKind" = $kind AND "subjectId" = $id`,
		`DELETE FROM "CollectionItem" WHERE "subjectKind" = $kind AND "subjectId" = $id`,
	}
	if kind == schema.RecordKindGallery {
		statements = append(statements, `DELETE FROM "Image" WHERE "gallery" = $id`)
	}
	if kind == schema.RecordKindVideo {
		statements = append(statements, `DELETE FROM "GroupItem" WHERE "video" = $id`, `DELETE FROM "Clip" WHERE "video" = $id`)
	}
	return execAll(ctx, tx, statements, vars)
}

// clearReferences removes the record's id from the records that point at it.
func clearReferences(ctx context.Context, tx *db.DB, kind schema.RecordKind, id string) error {
	switch kind {
	case schema.RecordKindVideo:
		return execAll(ctx, tx, []string{`UPDATE "Gallery" SET "video" = NULL WHERE "video" = $id`}, db.Vars{"id": id})
	case schema.RecordKindGallery:
		return execAll(ctx, tx, []string{`UPDATE "Video" SET "gallery" = NULL WHERE "gallery" = $id`}, db.Vars{"id": id})
	case schema.RecordKindStudio:
		statements := []string{`UPDATE "Studio" SET "parent" = NULL WHERE "parent" = $id`, `UPDATE "Group" SET "studio" = NULL WHERE "studio" = $id`}
		for _, table := range mediaTables {
			statements = append(statements, `UPDATE `+db.Quote(table)+` SET "studio" = NULL WHERE "studio" = $id`)
		}
		return execAll(ctx, tx, statements, db.Vars{"id": id})
	case schema.RecordKindPerformer:
		statements := []string{`UPDATE "CamModel" SET "performer" = NULL WHERE "performer" = $id`}
		statements = append(statements, removeFromLists(mediaTables, "performers")...)
		return execAll(ctx, tx, statements, db.Vars{"id": id})
	case schema.RecordKindTag:
		statements := []string{`UPDATE "Segment" SET "tag" = NULL WHERE "tag" = $id`}
		statements = append(statements, removeFromLists(append([]string{"Performer", "Studio"}, mediaTables...), "tags")...)
		statements = append(statements, removeFromLists([]string{"Category"}, "matchTags")...)
		return execAll(ctx, tx, statements, db.Vars{"id": id})
	}
	return nil
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
