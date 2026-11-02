package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// Row is one result row keyed by column name, with values as SQLite returns
// them (text, integers, floats, blobs). NULL is absent.
type Row = map[string]any

// Vars binds named $parameters in a query.
type Vars = map[string]any

// Query runs a SELECT and returns every row.
func (d *DB) Query(ctx context.Context, query string, vars Vars) ([]Row, error) {
	rows, err := d.executor.QueryContext(ctx, query, namedArgs(query, vars)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

// QueryRow runs a SELECT and returns its first row, or nil when it has none.
func (d *DB) QueryRow(ctx context.Context, query string, vars Vars) (Row, error) {
	rows, err := d.Query(ctx, query, vars)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0], nil
}

// Exec runs a statement and returns the number of affected rows.
func (d *DB) Exec(ctx context.Context, query string, vars Vars) (int64, error) {
	result, err := d.executor.ExecContext(ctx, query, namedArgs(query, vars)...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Int runs a query returning one integer (e.g. a COUNT). No rows yields 0.
func (d *DB) Int(ctx context.Context, query string, vars Vars) (int, error) {
	row, err := d.QueryRow(ctx, query, vars)
	if err != nil || row == nil {
		return 0, err
	}
	for _, value := range row {
		return AsInt(value), nil
	}
	return 0, nil
}

// Strings runs a single-column query and returns the non-NULL string values.
func (d *DB) Strings(ctx context.Context, query string, vars Vars) ([]string, error) {
	rows, err := d.executor.QueryContext(ctx, query, namedArgs(query, vars)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := []string{}
	for rows.Next() {
		var value sql.NullString
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		if value.Valid {
			values = append(values, value.String)
		}
	}
	return values, rows.Err()
}

// QueryAs runs a SELECT and decodes each row into T (see DecodeRow).
func QueryAs[T any](ctx context.Context, d *DB, query string, vars Vars) ([]T, error) {
	rows, err := d.Query(ctx, query, vars)
	if err != nil {
		return nil, err
	}
	out := make([]T, 0, len(rows))
	for _, row := range rows {
		var value T
		if err := DecodeRow(row, &value); err != nil {
			return nil, err
		}
		out = append(out, value)
	}
	return out, nil
}

// QueryOneAs is QueryAs for the first row; nil when there is none.
func QueryOneAs[T any](ctx context.Context, d *DB, query string, vars Vars) (*T, error) {
	row, err := d.QueryRow(ctx, query, vars)
	if err != nil || row == nil {
		return nil, err
	}
	var value T
	if err := DecodeRow(row, &value); err != nil {
		return nil, err
	}
	return &value, nil
}

// DecodeRow copies a row into target, a struct pointer, through its json
// tags. SQLite has no JSON or boolean storage, so the target's field types
// decide: text going into a slice, map or struct field is parsed as JSON, and
// an integer going into a bool field becomes true or false.
func DecodeRow(row Row, target any) error {
	kinds := fieldKinds(reflect.TypeOf(target).Elem())
	prepared := make(map[string]any, len(row))
	for column, value := range row {
		prepared[column] = prepareValue(kinds[column], value)
	}
	encoded, err := json.Marshal(prepared)
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, target)
}

// prepareValue converts one stored value for the field kind it decodes into.
func prepareValue(kind reflect.Kind, value any) any {
	text, isText := value.(string)
	switch kind {
	case reflect.Slice, reflect.Map, reflect.Struct, reflect.Interface:
		if isText && json.Valid([]byte(text)) {
			return json.RawMessage(text)
		}
	case reflect.Bool:
		return AsBool(value)
	}
	return value
}

// fieldKinds maps each json field name of a struct type, embedded structs
// included, to the kind of its value (pointers unwrapped). time.Time counts
// as a scalar.
func fieldKinds(structType reflect.Type) map[string]reflect.Kind {
	kinds := map[string]reflect.Kind{}
	if structType.Kind() != reflect.Struct {
		return kinds
	}
	for index := 0; index < structType.NumField(); index++ {
		field := structType.Field(index)
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			for name, kind := range fieldKinds(field.Type) {
				kinds[name] = kind
			}
			continue
		}
		name := jsonName(field)
		if name == "" {
			continue
		}
		kinds[name] = valueKind(field.Type)
	}
	return kinds
}

// ColumnKinds maps each column of a model struct (a value or pointer of the
// generated schema type) to the kind of Go value it decodes into: slices,
// maps, structs and interfaces are JSON columns, bools are 0/1.
func ColumnKinds(model any) map[string]reflect.Kind {
	modelType := reflect.TypeOf(model)
	for modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
	}
	return fieldKinds(modelType)
}

// IsJSONKind reports whether a column of this kind (see ColumnKinds) holds
// JSON text.
func IsJSONKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Slice, reflect.Map, reflect.Struct, reflect.Interface:
		return true
	}
	return false
}

// StoredValue is value in the form a write binds it: JSON text for
// collections and structs, text for times, plain scalars otherwise.
func StoredValue(value any) any {
	return bindValue(value)
}

// valueKind is the kind of t with pointers removed; time.Time is a string.
func valueKind(t reflect.Type) reflect.Kind {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == reflect.TypeOf(time.Time{}) {
		return reflect.String
	}
	if t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8 {
		return reflect.String
	}
	return t.Kind()
}

// jsonName is the name a field has in JSON, or "" when it is skipped.
func jsonName(field reflect.StructField) string {
	if !field.IsExported() {
		return ""
	}
	tag := field.Tag.Get("json")
	if tag == "-" {
		return ""
	}
	name, _, _ := strings.Cut(tag, ",")
	if name == "" {
		return field.Name
	}
	return name
}

// parameterPattern finds $name placeholders.
var parameterPattern = regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)

// namedArgs binds the vars the query references. Unreferenced vars are
// dropped, so callers can build one Vars map for several query shapes.
func namedArgs(query string, vars Vars) []any {
	if len(vars) == 0 {
		return nil
	}
	seen := map[string]bool{}
	args := []any{}
	for _, match := range parameterPattern.FindAllStringSubmatch(query, -1) {
		name := match[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		value, ok := vars[name]
		if !ok {
			continue
		}
		args = append(args, sql.Named(name, bindValue(value)))
	}
	return args
}

// bindValue converts a Go value into something SQLite stores: times become
// their text form, collections and structs become JSON text.
func bindValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case string, bool, int, int64, int32, float64, float32, []byte:
		return typed
	case time.Time:
		return FormatTime(typed)
	case *time.Time:
		if typed == nil {
			return nil
		}
		return FormatTime(*typed)
	case json.RawMessage:
		return string(typed)
	case json.Number:
		return string(typed)
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Pointer:
		if reflected.IsNil() {
			return nil
		}
		return bindValue(reflected.Elem().Interface())
	case reflect.String:
		return reflected.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflected.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(reflected.Uint())
	case reflect.Float32, reflect.Float64:
		return reflected.Float()
	case reflect.Bool:
		return reflected.Bool()
	}
	return encodeJSON(value)
}

// encodeJSON renders value as JSON text; a nil slice becomes "[]" and a nil
// map NULL.
func encodeJSON(value any) any {
	reflected := reflect.ValueOf(value)
	if reflected.Kind() == reflect.Slice && reflected.IsNil() {
		return "[]"
	}
	if reflected.Kind() == reflect.Map && reflected.IsNil() {
		return nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return string(encoded)
}

// scanRows reads every row as raw values; blobs stay bytes, other byte
// slices become text.
func scanRows(rows *sql.Rows) ([]Row, error) {
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	names := make([]string, len(columnTypes))
	declared := make([]string, len(columnTypes))
	for index, columnType := range columnTypes {
		names[index] = columnType.Name()
		declared[index] = strings.ToUpper(columnType.DatabaseTypeName())
	}
	out := []Row{}
	for rows.Next() {
		values := make([]any, len(names))
		pointers := make([]any, len(names))
		for index := range values {
			pointers[index] = &values[index]
		}
		if err := rows.Scan(pointers...); err != nil {
			return nil, err
		}
		row := make(Row, len(names))
		for index, name := range names {
			if values[index] == nil {
				continue
			}
			row[name] = textUnlessBlob(declared[index], values[index])
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// textUnlessBlob turns scanned bytes into a string unless the column is a BLOB.
func textUnlessBlob(declaredType string, value any) any {
	if raw, isBytes := value.([]byte); isBytes && declaredType != "BLOB" {
		return string(raw)
	}
	return value
}

// Quote renders a table or column name as a quoted SQL identifier, so names
// like "Group" or "createdAt" survive as written.
func Quote(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

// AsInt reads an integer-ish value (int64, float64, bool, numeric string).
func AsInt(value any) int {
	switch typed := value.(type) {
	case int64:
		return int(typed)
	case int:
		return typed
	case float64:
		return int(typed)
	case bool:
		if typed {
			return 1
		}
		return 0
	case string:
		var parsed int
		fmt.Sscan(typed, &parsed)
		return parsed
	}
	return 0
}

// AsBool reads a boolean-ish value (bool, or a 0/1 integer).
func AsBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case int64:
		return typed != 0
	case int:
		return typed != 0
	case float64:
		return typed != 0
	}
	return false
}
