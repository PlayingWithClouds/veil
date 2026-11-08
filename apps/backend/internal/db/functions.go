package db

import (
	"database/sql/driver"

	"modernc.org/sqlite"

	"github.com/playingwithclouds/veil/internal/searchtext"
)

// SQL functions every connection has. veil_search(text, query, words) is 1
// when text matches the search query as searchtext reads it (bare words
// required only when words is 1), else 0.
func init() {
	sqlite.MustRegisterDeterministicScalarFunction("veil_search", 3, veilSearch)
}

// veilSearch is veil_search; a NULL text is matched as empty.
func veilSearch(_ *sqlite.FunctionContext, args []driver.Value) (driver.Value, error) {
	text, _ := args[0].(string)
	query, _ := args[1].(string)
	words, _ := args[2].(int64)
	if searchtext.Parse(query).Matches(text, words == 1) {
		return int64(1), nil
	}
	return int64(0), nil
}
