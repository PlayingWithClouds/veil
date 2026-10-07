package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"sort"
	"strings"
	"sync"
)

// Get loads one row by id; nil when it does not exist.
func (d *DB) Get(ctx context.Context, id RecordID) (Row, error) {
	if err := validTable(id.Table); err != nil {
		return nil, err
	}
	return d.QueryRow(ctx, "SELECT * FROM "+id.Table+" WHERE id = $id", Vars{"id": id})
}

// GetString is Get for a "table:id" string.
func (d *DB) GetString(ctx context.Context, id string) (Row, error) {
	parsed, err := ParseRecordID(id)
	if err != nil {
		return nil, err
	}
	return d.Get(ctx, *parsed)
}

// Insert creates a row in table from fields (a map or a json-tagged struct)
// and returns its id. A non-empty "id" field is used as the id; otherwise a
// random one is generated. Fields without a matching column are dropped.
func (d *DB) Insert(ctx context.Context, table string, fields any) (RecordID, error) {
	values, err := toFields(fields)
	if err != nil {
		return RecordID{}, err
	}
	id := NewRecordID(table)
	if explicit, ok := idField(values); ok {
		id = explicit
	}
	delete(values, "id")
	columns, err := d.columns(ctx, table)
	if err != nil {
		return RecordID{}, err
	}
	names := []string{"id"}
	placeholders := []string{"?"}
	args := []any{id.String()}
	for _, name := range sortedKeys(values) {
		declaredType, known := columns[name]
		if !known {
			warnUnknownColumn(table, name)
			continue
		}
		names = append(names, quoteIdentifier(name))
		placeholders = append(placeholders, "?")
		args = append(args, columnValue(declaredType, values[name]))
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(names, ", "), strings.Join(placeholders, ", "))
	if _, err := d.executor.ExecContext(ctx, query, args...); err != nil {
		return RecordID{}, fmt.Errorf("insert %s: %w", table, err)
	}
	return id, nil
}

// Merge sets the given fields on an existing row. Fields without a matching
// column are dropped; "id" is ignored.
func (d *DB) Merge(ctx context.Context, id RecordID, fields any) error {
	values, err := toFields(fields)
	if err != nil {
		return err
	}
	delete(values, "id")
	if len(values) == 0 {
		return nil
	}
	columns, err := d.columns(ctx, id.Table)
	if err != nil {
		return err
	}
	assignments := []string{}
	args := []any{}
	for _, name := range sortedKeys(values) {
		declaredType, known := columns[name]
		if !known {
			warnUnknownColumn(id.Table, name)
			continue
		}
		assignments = append(assignments, quoteIdentifier(name)+" = ?")
		args = append(args, columnValue(declaredType, values[name]))
	}
	if len(assignments) == 0 {
		return nil
	}
	args = append(args, id.String())
	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", id.Table, strings.Join(assignments, ", "))
	if _, err := d.executor.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("merge %s: %w", id, err)
	}
	return nil
}

// Delete removes one row by id.
func (d *DB) Delete(ctx context.Context, id RecordID) error {
	if err := validTable(id.Table); err != nil {
		return err
	}
	_, err := d.Exec(ctx, "DELETE FROM "+id.Table+" WHERE id = $id", Vars{"id": id})
	return err
}

// columnValue encodes value for a column of declaredType: JSON columns always
// get JSON text, everything else goes through bindValue.
func columnValue(declaredType string, value any) any {
	if strings.EqualFold(declaredType, "JSON") {
		return encodeJSON(value)
	}
	return bindValue(value)
}

// toFields turns a map or a json-tagged struct into a column map. Structs go
// through JSON, so omitempty fields are left to their column defaults.
func toFields(fields any) (map[string]any, error) {
	if values, ok := fields.(map[string]any); ok {
		return maps.Clone(values), nil
	}
	encoded, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("encode fields: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	values := map[string]any{}
	if err := decoder.Decode(&values); err != nil {
		return nil, fmt.Errorf("fields must encode to an object: %w", err)
	}
	return values, nil
}

// idField reads an explicit id from a field map.
func idField(values map[string]any) (RecordID, bool) {
	switch typed := values["id"].(type) {
	case RecordID:
		return typed, !typed.IsZero()
	case *RecordID:
		if typed != nil && !typed.IsZero() {
			return *typed, true
		}
	case string:
		if parsed, err := ParseRecordID(typed); err == nil {
			return *parsed, true
		}
	}
	return RecordID{}, false
}

// validTable rejects anything that isn't a plain identifier, since table
// names are interpolated into SQL.
func validTable(table string) error {
	if table == "" {
		return fmt.Errorf("empty table name")
	}
	for _, character := range table {
		isLetter := character >= 'a' && character <= 'z'
		isDigit := character >= '0' && character <= '9'
		if !isLetter && !isDigit && character != '_' {
			return fmt.Errorf("invalid table name %q", table)
		}
	}
	return nil
}

func quoteIdentifier(name string) string {
	return `"` + name + `"`
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

var warnedColumns sync.Map

// warnUnknownColumn logs, once per table/column, a field that has no column.
func warnUnknownColumn(table, column string) {
	key := table + "." + column
	if _, alreadyWarned := warnedColumns.LoadOrStore(key, true); alreadyWarned {
		return
	}
	log.Printf("db: dropping field %s (no such column)", key)
}
