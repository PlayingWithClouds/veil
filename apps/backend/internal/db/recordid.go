package db

import (
	"crypto/rand"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// RecordID identifies a row as "table:id". The full string is what the id
// column and every link column store, and what the GraphQL API exposes.
type RecordID struct {
	Table string
	ID    string
}

const recordIDAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// recordIDLength matches the length of the random ids SurrealDB generated, so
// ids look the same before and after the SQLite migration.
const recordIDLength = 20

// NewRecordID returns a fresh random id in table.
func NewRecordID(table string) RecordID {
	randomBytes := make([]byte, recordIDLength)
	if _, err := rand.Read(randomBytes); err != nil {
		panic(fmt.Sprintf("record id: %v", err))
	}
	for index, value := range randomBytes {
		randomBytes[index] = recordIDAlphabet[int(value)%len(recordIDAlphabet)]
	}
	return RecordID{Table: table, ID: string(randomBytes)}
}

// ParseRecordID parses "table:id". Both parts must be non-empty.
func ParseRecordID(value string) (*RecordID, error) {
	table, id, found := strings.Cut(value, ":")
	if !found || table == "" || id == "" {
		return nil, fmt.Errorf("invalid record id %q", value)
	}
	return &RecordID{Table: table, ID: id}, nil
}

// String renders the id as "table:id".
func (r RecordID) String() string {
	return r.Table + ":" + r.ID
}

// IsZero reports whether the id is unset.
func (r RecordID) IsZero() bool {
	return r.Table == "" && r.ID == ""
}

// Value stores the id as its "table:id" string.
func (r RecordID) Value() (driver.Value, error) {
	return r.String(), nil
}

// Scan reads a "table:id" string column.
func (r *RecordID) Scan(source any) error {
	var text string
	switch value := source.(type) {
	case string:
		text = value
	case []byte:
		text = string(value)
	default:
		return fmt.Errorf("record id: cannot scan %T", source)
	}
	parsed, err := ParseRecordID(text)
	if err != nil {
		return err
	}
	*r = *parsed
	return nil
}

// MarshalJSON encodes the id as its "table:id" string.
func (r RecordID) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON decodes a "table:id" string.
func (r *RecordID) UnmarshalJSON(data []byte) error {
	var text string
	if err := json.Unmarshal(data, &text); err != nil {
		return err
	}
	parsed, err := ParseRecordID(text)
	if err != nil {
		return err
	}
	*r = *parsed
	return nil
}
