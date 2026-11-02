package library

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Merges lists the merges into a record that can still be undone, newest
// first.
func (l *Library) Merges(ctx context.Context, kind schema.RecordKind, keptID string) ([]schema.MergeJournal, error) {
	return db.QueryAs[schema.MergeJournal](ctx, l.database,
		`SELECT * FROM "MergeJournal" WHERE "subjectKind" = $kind AND "keptId" = $kept
		   AND "undoneAt" IS NULL AND "deletedAt" IS NULL ORDER BY "createdAt" DESC`,
		db.Vars{"kind": kind, "kept": keptID})
}

// UndoMerge brings a merged record back as it was: the rows the merge
// deleted are inserted again, and every column it changed gets its value
// from before, unless something changed that column since. What reached the
// survivor after the merge stays with it.
func (l *Library) UndoMerge(ctx context.Context, journalID string) (*schema.MergeJournal, error) {
	err := l.database.Tx(ctx, func(tx *db.DB) error {
		entry, err := db.Get[schema.MergeJournal](ctx, tx, mergeJournalTable, journalID)
		if err != nil {
			return err
		}
		if entry == nil {
			return fmt.Errorf("no merge %s", journalID)
		}
		if entry.UndoneAt != nil {
			return fmt.Errorf("that merge was undone already")
		}
		rows, err := journalRows(entry.Rows)
		if err != nil {
			return err
		}
		for index := len(rows) - 1; index >= 0; index-- {
			if err := restoreRow(ctx, tx, rows[index]); err != nil {
				return fmt.Errorf("undo merge: %s %s: %w", rows[index].Table, rows[index].ID, err)
			}
		}
		return tx.Update(ctx, mergeJournalTable, journalID, db.Vars{"undoneAt": db.Now()})
	})
	if err != nil {
		return nil, err
	}
	return db.Get[schema.MergeJournal](ctx, l.database, mergeJournalTable, journalID)
}

// journalRows decodes a journal's rows, numbers kept exact.
func journalRows(encoded any) ([]journalRow, error) {
	text, err := journalText(encoded)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(text))
	decoder.UseNumber()
	rows := []journalRow{}
	return rows, decoder.Decode(&rows)
}

// journalText is a journal's rows column as JSON text, however it was read.
func journalText(encoded any) ([]byte, error) {
	switch typed := encoded.(type) {
	case string:
		return []byte(typed), nil
	case json.RawMessage:
		return typed, nil
	case []byte:
		return typed, nil
	}
	return json.Marshal(encoded)
}

// restoreRow undoes the merge's change to one row.
func restoreRow(ctx context.Context, tx *db.DB, row journalRow) error {
	current, err := tx.QueryRow(ctx, `SELECT * FROM `+db.Quote(row.Table)+` WHERE "id" = $id`, db.Vars{"id": row.ID})
	if err != nil {
		return err
	}
	if row.After == nil {
		return reinsert(ctx, tx, row, current)
	}
	if current == nil {
		// Deleted since the merge: nothing to put back.
		return nil
	}
	fields := db.Vars{}
	for column, before := range row.Before {
		if sameAsJournaled(current[column], row.After[column]) {
			fields[column] = columnValue(before)
		}
	}
	return tx.Update(ctx, row.Table, row.ID, fields)
}

// reinsert puts a row the merge deleted back, unless it exists again.
func reinsert(ctx context.Context, tx *db.DB, row journalRow, current db.Row) error {
	if current != nil {
		return nil
	}
	fields := db.Vars{}
	for column, value := range row.Before {
		fields[column] = columnValue(value)
	}
	_, err := tx.Insert(ctx, row.Table, fields)
	return err
}
