package library

import (
	"regexp"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// setText sets a text column when the value isn't blank.
func setText(fields db.Vars, column, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	fields[column] = value
}

// setAsset sets an asset column to the URI when it isn't blank.
func setAsset(fields db.Vars, column, uri string) {
	asset := assetOf(uri)
	if asset == nil {
		return
	}
	fields[column] = asset
}

// assetOf wraps a URI as an asset; nil for a blank one.
func assetOf(uri string) *schema.Asset {
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return nil
	}
	return &schema.Asset{Uri: uri}
}

// fillListDefaults sets list columns a new row needs to an empty list when
// the fields don't carry them.
func fillListDefaults(fields db.Vars, columns ...string) {
	for _, column := range columns {
		if _, present := fields[column]; !present {
			fields[column] = []string{}
		}
	}
}

// columnsOf lists the columns fields sets.
func columnsOf(fields db.Vars) []string {
	columns := make([]string, 0, len(fields))
	for column := range fields {
		columns = append(columns, column)
	}
	return columns
}

// isoDatePattern is the date part of an ISO 8601 date or timestamp.
var isoDatePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}`)

// isoDate keeps the YYYY-MM-DD part of a site's date; "" when it has none.
func isoDate(value string) string {
	return isoDatePattern.FindString(strings.TrimSpace(value))
}
