package provenance

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Update writes fields to the record kind/id on behalf of author. Fields the
// author may not write are dropped (see mayWrite); unchanged ones are
// skipped; every change is recorded as a revision. A trashed record is
// updated too and stays trashed. It returns the columns that changed.
func (r *Recorder) Update(ctx context.Context, kind schema.RecordKind, id string, fields db.Vars, author Author) ([]string, error) {
	recordType, known := recordTypes[kind]
	if !known {
		return nil, fmt.Errorf("can't write a %s", kind)
	}
	var changed []string
	err := r.database.Tx(ctx, func(tx *db.DB) error {
		current, err := tx.QueryRow(ctx, `SELECT * FROM `+db.Quote(recordType.table)+` WHERE "id" = $id`, db.Vars{"id": id})
		if err != nil || current == nil {
			return err
		}
		owners, err := fieldOwners(ctx, tx, kind, id)
		if err != nil {
			return err
		}
		changes := db.Vars{}
		for column, value := range fields {
			if !mayWrite(kind, column, owners, author) {
				continue
			}
			before := revisionValue(current[column], recordType.columns[column])
			after := revisionValue(db.StoredValue(value), recordType.columns[column])
			if sameJSON(before, after) {
				continue
			}
			changes[column] = value
			changed = append(changed, column)
			if err := recordChange(ctx, tx, kind, id, column, before, after, author); err != nil {
				return err
			}
		}
		if len(changes) == 0 {
			return nil
		}
		if err := tx.Update(ctx, recordType.table, id, changes); err != nil {
			return err
		}
		return claimColumns(ctx, tx, kind, id, keys(changes), owners, author)
	})
	return changed, err
}

// Claim records author as the owner of columns of a record just created, so
// later writes from other sources leave them alone. Creation itself has no
// revision: the record's createdAt says when it appeared.
func (r *Recorder) Claim(ctx context.Context, kind schema.RecordKind, id string, columns []string, author Author) error {
	return r.database.Tx(ctx, func(tx *db.DB) error {
		owners, err := fieldOwners(ctx, tx, kind, id)
		if err != nil {
			return err
		}
		return claimColumns(ctx, tx, kind, id, columns, owners, author)
	})
}

// mayWrite decides whether author may write column. The user always may and
// locks the field by doing so; anyone else never overwrites a locked field,
// and a plugin only writes a field that no other source owns.
func mayWrite(kind schema.RecordKind, column string, owners map[uint32]schema.FieldSource, author Author) bool {
	if _, known := Ordinal(kind, column); !known {
		return false
	}
	if author.Actor == schema.ActorUser || !isOwnable(kind, column) {
		return true
	}
	ordinal, _ := Ordinal(kind, column)
	owner, owned := owners[ordinal]
	if !owned {
		return true
	}
	if owner.Locked {
		return false
	}
	if author.Actor != schema.ActorPlugin || owner.WrittenBy != schema.ActorPlugin {
		return true
	}
	return sourceOf(owner) == author.SourceID
}

// sourceOf is the source id a field owner names, or "".
func sourceOf(owner schema.FieldSource) string {
	if owner.Source == nil {
		return ""
	}
	return *owner.Source
}

// claimColumns makes author the owner of the ownable columns: plugins with
// their source, the user locked.
func claimColumns(ctx context.Context, tx *db.DB, kind schema.RecordKind, id string, columns []string, owners map[uint32]schema.FieldSource, author Author) error {
	for _, column := range columns {
		if !isOwnable(kind, column) {
			continue
		}
		ordinal, _ := Ordinal(kind, column)
		if err := setOwner(ctx, tx, kind, id, ordinal, owners, author); err != nil {
			return err
		}
	}
	return nil
}

// setOwner inserts or replaces the FieldSource of one field.
func setOwner(ctx context.Context, tx *db.DB, kind schema.RecordKind, id string, ordinal uint32, owners map[uint32]schema.FieldSource, author Author) error {
	fields := db.Vars{
		"writtenBy": author.Actor,
		"source":    nil,
		"locked":    author.Actor == schema.ActorUser,
	}
	if author.SourceID != "" {
		fields["source"] = author.SourceID
	}
	existing, owned := owners[ordinal]
	if owned {
		return tx.Update(ctx, "FieldSource", existing.Id, fields)
	}
	fields["subjectKind"] = kind
	fields["subjectId"] = id
	fields["fieldOrdinal"] = ordinal
	_, err := tx.Insert(ctx, "FieldSource", fields)
	return err
}

// recordChange writes the revision of one changed column.
func recordChange(ctx context.Context, tx *db.DB, kind schema.RecordKind, id, column string, before, after json.RawMessage, author Author) error {
	if !hasRevisions(column) {
		return nil
	}
	ordinal, _ := Ordinal(kind, column)
	revision := db.Vars{
		"subjectKind":  kind,
		"subjectId":    id,
		"fieldOrdinal": ordinal,
		"writtenBy":    author.Actor,
	}
	// An empty RawMessage would bind as "", which isn't valid JSON; NULL is.
	if before != nil {
		revision["before"] = before
	}
	if after != nil {
		revision["after"] = after
	}
	if author.Name != "" {
		revision["actorName"] = author.Name
	}
	if author.AutomationRunID != "" {
		revision["automationRunId"] = author.AutomationRunID
	}
	_, err := tx.Insert(ctx, "Revision", revision)
	return err
}

// fieldOwners loads a record's FieldSource rows by field ordinal.
func fieldOwners(ctx context.Context, tx *db.DB, kind schema.RecordKind, id string) (map[uint32]schema.FieldSource, error) {
	rows, err := db.QueryAs[schema.FieldSource](ctx, tx,
		`SELECT * FROM "FieldSource" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL`,
		db.Vars{"kind": kind, "id": id})
	if err != nil {
		return nil, err
	}
	owners := make(map[uint32]schema.FieldSource, len(rows))
	for _, row := range rows {
		owners[row.FieldOrdinal] = row
	}
	return owners, nil
}

// revisionValue renders a stored column value as JSON: JSON columns as they
// are, booleans as true/false, NULL as nil.
func revisionValue(stored any, kind reflect.Kind) json.RawMessage {
	if stored == nil {
		return nil
	}
	text, isText := stored.(string)
	if isText && db.IsJSONKind(kind) {
		return json.RawMessage(text)
	}
	if kind == reflect.Bool {
		stored = db.AsBool(stored)
	}
	encoded, err := json.Marshal(stored)
	if err != nil {
		return nil
	}
	return encoded
}

// sameJSON reports whether two JSON values are equal, ignoring formatting
// and number spelling (600 and 600.0).
func sameJSON(first, second json.RawMessage) bool {
	if first == nil || second == nil {
		return first == nil && second == nil
	}
	var firstValue, secondValue any
	if json.Unmarshal(first, &firstValue) != nil || json.Unmarshal(second, &secondValue) != nil {
		return string(first) == string(second)
	}
	return reflect.DeepEqual(firstValue, secondValue)
}

// keys returns the columns of fields.
func keys(fields db.Vars) []string {
	columns := make([]string, 0, len(fields))
	for column := range fields {
		columns = append(columns, column)
	}
	return columns
}
