package db

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/google/uuid"
)

// NewID returns a fresh record id. Ids are uuids so records from different
// devices never collide.
func NewID() string {
	return uuid.NewString()
}

// Get loads the row with id from table into T; nil when there is none or it
// is trashed.
func Get[T any](ctx context.Context, d *DB, table, id string) (*T, error) {
	query := "SELECT * FROM " + Quote(table) + ` WHERE "id" = $id AND "deletedAt" IS NULL`
	return QueryOneAs[T](ctx, d, query, Vars{"id": id})
}

// Insert writes value (a struct or map) as a new row of table and returns its
// id. Fields left out of the JSON form (omitempty and empty) take the column
// default; an empty id gets a new one, empty createdAt/updatedAt the current
// time.
func (d *DB) Insert(ctx context.Context, table string, value any) (string, error) {
	fields, err := toFields(value)
	if err != nil {
		return "", err
	}
	if id, _ := fields["id"].(string); id == "" {
		fields["id"] = NewID()
	}
	stampIfEmpty(fields, "createdAt")
	stampIfEmpty(fields, "updatedAt")
	columns := sortedKeys(fields)
	quoted := make([]string, len(columns))
	placeholders := make([]string, len(columns))
	for index, column := range columns {
		quoted[index] = Quote(column)
		placeholders[index] = "$" + parameterName(index)
	}
	query := "INSERT INTO " + Quote(table) + " (" + strings.Join(quoted, ", ") + ") VALUES (" + strings.Join(placeholders, ", ") + ")"
	if _, err := d.Exec(ctx, query, positionalVars(columns, fields)); err != nil {
		return "", fmt.Errorf("insert into %s: %w", table, err)
	}
	return fields["id"].(string), nil
}

// Update sets the given fields on the row with id and bumps updatedAt.
func (d *DB) Update(ctx context.Context, table, id string, fields Vars) error {
	if len(fields) == 0 {
		return nil
	}
	changes := Vars{}
	for column, value := range fields {
		changes[column] = value
	}
	changes["updatedAt"] = Now()
	columns := sortedKeys(changes)
	assignments := make([]string, len(columns))
	for index, column := range columns {
		assignments[index] = Quote(column) + " = $" + parameterName(index)
	}
	vars := positionalVars(columns, changes)
	vars["id"] = id
	query := "UPDATE " + Quote(table) + " SET " + strings.Join(assignments, ", ") + ` WHERE "id" = $id`
	if _, err := d.Exec(ctx, query, vars); err != nil {
		return fmt.Errorf("update %s: %w", table, err)
	}
	return nil
}

// Trash soft-deletes the row with id: it disappears from reads until purged.
func (d *DB) Trash(ctx context.Context, table, id string) error {
	return d.Update(ctx, table, id, Vars{"deletedAt": Now()})
}

// Delete removes the row with id for good. Records the user can see are
// trashed instead (Trash); this is for bookkeeping rows such as jobs.
func (d *DB) Delete(ctx context.Context, table, id string) error {
	_, err := d.Exec(ctx, "DELETE FROM "+Quote(table)+` WHERE "id" = $id`, Vars{"id": id})
	return err
}

// toFields turns a struct or map into column values through its JSON form.
func toFields(value any) (Vars, error) {
	if fields, isVars := value.(Vars); isVars {
		copied := Vars{}
		for key, field := range fields {
			copied[key] = field
		}
		return copied, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	fields := Vars{}
	if err := json.Unmarshal(encoded, &fields); err != nil {
		return nil, err
	}
	fillEmptyLists(fields, reflect.TypeOf(value))
	return fields, nil
}

// fillEmptyLists writes nil or omitted list fields of a struct as [], since
// list columns are NOT NULL and an empty list means the same as none.
func fillEmptyLists(fields Vars, valueType reflect.Type) {
	for valueType.Kind() == reflect.Pointer {
		valueType = valueType.Elem()
	}
	for name, kind := range fieldKinds(valueType) {
		if kind != reflect.Slice || fields[name] != nil {
			continue
		}
		fields[name] = []any{}
	}
}

// stampIfEmpty sets a timestamp column to now when it is missing or empty.
func stampIfEmpty(fields Vars, column string) {
	if text, _ := fields[column].(string); text == "" {
		fields[column] = Now()
	}
}

// positionalVars binds each column's value to the parameter named by its
// position, since column names may not be valid parameter names.
func positionalVars(columns []string, fields Vars) Vars {
	vars := Vars{}
	for index, column := range columns {
		vars[parameterName(index)] = jsonBindValue(fields[column])
	}
	return vars
}

// jsonBindValue binds a value decoded from JSON: numbers that are whole
// become integers, so INTEGER columns don't receive 3.0.
func jsonBindValue(value any) any {
	number, isNumber := value.(float64)
	if isNumber && number == float64(int64(number)) {
		return int64(number)
	}
	return bindValue(value)
}

// parameterName is the placeholder name for the column at index.
func parameterName(index int) string {
	return fmt.Sprintf("p%d", index)
}

// sortedKeys returns the map's keys in a stable order.
func sortedKeys(fields Vars) []string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
