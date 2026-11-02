package provenance

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Revisions lists a record's changes, newest first.
func (r *Recorder) Revisions(ctx context.Context, kind schema.RecordKind, id string, limit, offset int) ([]schema.Revision, error) {
	return db.QueryAs[schema.Revision](ctx, r.database,
		`SELECT * FROM "Revision" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL
		 ORDER BY "createdAt" DESC, rowid DESC LIMIT $limit OFFSET $offset`,
		db.Vars{"kind": kind, "id": id, "limit": limit, "offset": offset})
}

// Revision loads one revision.
func (r *Recorder) Revision(ctx context.Context, id string) (*schema.Revision, error) {
	return db.Get[schema.Revision](ctx, r.database, "Revision", id)
}

// Revert puts a field back to its value before the revision, as a user edit:
// the field is locked, so the plugin that changed it won't change it again.
// The revert is itself a revision.
func (r *Recorder) Revert(ctx context.Context, revisionID string) error {
	revision, err := r.Revision(ctx, revisionID)
	if err != nil {
		return err
	}
	if revision == nil {
		return fmt.Errorf("no revision %s", revisionID)
	}
	column := Column(revision.SubjectKind, revision.FieldOrdinal)
	if column == "" {
		return fmt.Errorf("revision %s names an unknown field", revisionID)
	}
	value, err := storedValue(revision.Before, recordTypes[revision.SubjectKind].columns[column])
	if err != nil {
		return err
	}
	_, err = r.Update(ctx, revision.SubjectKind, revision.SubjectId, db.Vars{column: value}, User())
	return err
}

// storedValue turns a revision's JSON value back into what the column stores.
func storedValue(value any, kind reflect.Kind) (any, error) {
	if value == nil {
		return nil, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	if db.IsJSONKind(kind) {
		return json.RawMessage(encoded), nil
	}
	var scalar any
	if err := json.Unmarshal(encoded, &scalar); err != nil {
		return nil, err
	}
	return scalar, nil
}

// FieldSources lists who owns each field of a record.
func (r *Recorder) FieldSources(ctx context.Context, kind schema.RecordKind, id string) ([]schema.FieldSource, error) {
	return db.QueryAs[schema.FieldSource](ctx, r.database,
		`SELECT * FROM "FieldSource" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL
		 ORDER BY "fieldOrdinal"`,
		db.Vars{"kind": kind, "id": id})
}

// SetLocked locks a field to its current value, or unlocks it. Unlocking
// releases the field: the next source that describes it takes it over.
func (r *Recorder) SetLocked(ctx context.Context, kind schema.RecordKind, id, column string, locked bool) error {
	ordinal, known := Ordinal(kind, column)
	if !known || !isOwnable(kind, column) {
		return fmt.Errorf("%s has no field %q to lock", kind, column)
	}
	return r.database.Tx(ctx, func(tx *db.DB) error {
		owners, err := fieldOwners(ctx, tx, kind, id)
		if err != nil {
			return err
		}
		if locked {
			return setOwner(ctx, tx, kind, id, ordinal, owners, User())
		}
		owner, owned := owners[ordinal]
		if !owned {
			return nil
		}
		return tx.Delete(ctx, "FieldSource", owner.Id)
	})
}
