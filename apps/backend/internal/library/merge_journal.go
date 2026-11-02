package library

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// mergeJournalTable keeps one row per merge, to undo it.
const mergeJournalTable = "MergeJournal"

// change is one statement over the rows of table that where selects. The
// merge journal reads those rows before and after it runs.
type change struct {
	table     string
	where     string
	statement string
}

// updateRows is a change setting columns on table's rows matching where.
func updateRows(table, set, where string) change {
	return change{table: table, where: where, statement: `UPDATE ` + db.Quote(table) + ` SET ` + set + ` WHERE ` + where}
}

// deleteRows is a change deleting table's rows matching where.
func deleteRows(table, where string) change {
	return change{table: table, where: where, statement: `DELETE FROM ` + db.Quote(table) + ` WHERE ` + where}
}

// journalRow is one row a merge touched. For a changed row, Before and After
// hold the changed columns only; for a deleted row Before is the whole row
// and After is nil.
type journalRow struct {
	Table  string         `json:"table"`
	ID     string         `json:"id"`
	Before map[string]any `json:"before"`
	After  map[string]any `json:"after"`
}

// journal collects what a merge changes. A nil journal only runs changes,
// as purging does.
type journal struct {
	rows []journalRow
}

// apply runs changes in order, journaling the rows each one touches.
func (j *journal) apply(ctx context.Context, tx *db.DB, changes []change, vars db.Vars) error {
	for _, step := range changes {
		if err := j.run(ctx, tx, step, vars); err != nil {
			return err
		}
	}
	return nil
}

// run runs one change and journals its rows.
func (j *journal) run(ctx context.Context, tx *db.DB, step change, vars db.Vars) error {
	if j == nil {
		_, err := tx.Exec(ctx, step.statement, vars)
		return err
	}
	before, err := tx.Query(ctx, `SELECT * FROM `+db.Quote(step.table)+` WHERE `+step.where, vars)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, step.statement, vars); err != nil {
		return err
	}
	return j.compare(ctx, tx, step.table, before)
}

// update sets fields on one row, journaling it.
func (j *journal) update(ctx context.Context, tx *db.DB, table, id string, fields db.Vars) error {
	where := `"id" = $journalId`
	vars := db.Vars{"journalId": id}
	before, err := tx.Query(ctx, `SELECT * FROM `+db.Quote(table)+` WHERE `+where, vars)
	if err != nil {
		return err
	}
	if err := tx.Update(ctx, table, id, fields); err != nil {
		return err
	}
	return j.compare(ctx, tx, table, before)
}

// compare reads the rows again after a change and journals how they differ.
func (j *journal) compare(ctx context.Context, tx *db.DB, table string, before []db.Row) error {
	if len(before) == 0 {
		return nil
	}
	ids := make([]string, 0, len(before))
	for _, row := range before {
		ids = append(ids, fmt.Sprint(row["id"]))
	}
	after, err := tx.Query(ctx, `SELECT * FROM `+db.Quote(table)+` WHERE "id" IN (SELECT value FROM json_each($ids))`, db.Vars{"ids": ids})
	if err != nil {
		return err
	}
	afterByID := map[string]db.Row{}
	for _, row := range after {
		afterByID[fmt.Sprint(row["id"])] = row
	}
	for _, row := range before {
		j.add(table, row, afterByID[fmt.Sprint(row["id"])])
	}
	return nil
}

// add journals one row: whole when deleted, its changed columns otherwise.
func (j *journal) add(table string, before, after db.Row) {
	id := fmt.Sprint(before["id"])
	if after == nil {
		j.rows = append(j.rows, journalRow{Table: table, ID: id, Before: storedRow(before)})
		return
	}
	changedBefore, changedAfter := map[string]any{}, map[string]any{}
	for _, column := range bothColumns(before, after) {
		if sameValue(before[column], after[column]) {
			continue
		}
		changedBefore[column] = storedValue(before[column])
		changedAfter[column] = storedValue(after[column])
	}
	if len(changedBefore) == 0 {
		return
	}
	j.rows = append(j.rows, journalRow{Table: table, ID: id, Before: changedBefore, After: changedAfter})
}

// save stores the journal of a merge.
func (j *journal) save(ctx context.Context, tx *db.DB, kind schema.RecordKind, keepID string, merged db.Row) error {
	encoded, err := json.Marshal(j.rows)
	if err != nil {
		return err
	}
	_, err = tx.Insert(ctx, mergeJournalTable, db.Vars{
		"subjectKind": kind,
		"keptId":      keepID,
		"mergedId":    merged["id"],
		"mergedName":  recordLabel(merged),
		"rows":        json.RawMessage(encoded),
		"writtenBy":   schema.ActorUser,
	})
	return err
}

// recordLabel is a record row's name, else its title; nil without either.
func recordLabel(row db.Row) any {
	if name, named := row["name"].(string); named && name != "" {
		return name
	}
	if title, titled := row["title"].(string); titled && title != "" {
		return title
	}
	return nil
}

// bothColumns lists the columns of both rows; a NULL column is absent from a row.
func bothColumns(first, second db.Row) []string {
	columns := []string{}
	for column := range first {
		columns = append(columns, column)
	}
	for column := range second {
		if _, known := first[column]; !known {
			columns = append(columns, column)
		}
	}
	return columns
}

// sameValue compares two column values as SQLite returned them.
func sameValue(first, second any) bool {
	firstBytes, firstIsBytes := first.([]byte)
	secondBytes, secondIsBytes := second.([]byte)
	if firstIsBytes || secondIsBytes {
		return firstIsBytes && secondIsBytes && bytes.Equal(firstBytes, secondBytes)
	}
	return reflect.DeepEqual(first, second)
}

// storedRow is a whole row as the journal stores it.
func storedRow(row db.Row) map[string]any {
	stored := map[string]any{}
	for column, value := range row {
		stored[column] = storedValue(value)
	}
	return stored
}

// storedValue is a column value as JSON keeps it: blobs as {"bytes": base64},
// since JSON would bring them back as text.
func storedValue(value any) any {
	if blob, isBlob := value.([]byte); isBlob {
		return map[string]any{"bytes": base64.StdEncoding.EncodeToString(blob)}
	}
	return value
}

// columnValue is a journaled value as it is written back.
func columnValue(value any) any {
	wrapped, isObject := value.(map[string]any)
	if !isObject {
		return numberValue(value)
	}
	encoded, _ := wrapped["bytes"].(string)
	blob, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil
	}
	return blob
}

// numberValue binds a json.Number as an integer when whole, else a float.
func numberValue(value any) any {
	number, isNumber := value.(json.Number)
	if !isNumber {
		return value
	}
	if whole, err := number.Int64(); err == nil {
		return whole
	}
	float, _ := number.Float64()
	return float
}

// sameAsJournaled reports whether a current column value is still the one
// the merge left.
func sameAsJournaled(current, journaled any) bool {
	return sameValue(storedComparable(current), storedComparable(columnValue(journaled)))
}

// storedComparable makes a value comparable across the JSON round trip:
// numbers as float64, the rest as they are.
func storedComparable(value any) any {
	switch typed := value.(type) {
	case int64:
		return float64(typed)
	case int:
		return float64(typed)
	}
	return value
}
