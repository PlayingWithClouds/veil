// Package provenance is the write path for records: every change to a
// record (media, performer, studio, tag, group, collection, …) goes through
// Recorder, which
//   - keeps per-field ownership (FieldSource): a plugin only writes fields no
//     other source owns, and never a field the user set (locked);
//   - records each changed field as a Revision (before, after, who), so any
//     change can be reverted.
//
// Fields are named by their OpenSchema ordinal (schema.FieldOrdinals), so
// renaming a field keeps its history.
package provenance

import (
	"reflect"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// recordType describes a stored kind of record.
type recordType struct {
	table string
	// columns maps each column to the kind of Go value it decodes into.
	columns map[string]reflect.Kind
	// owned: plugins fill its descriptive fields, so they get a FieldSource.
	owned bool
}

// recordTypes lists every kind of record. Media, performers, studios and
// groups are filled by plugins, so their fields have owners.
var recordTypes = map[schema.RecordKind]recordType{
	schema.RecordKindVideo:      {table: "Video", columns: db.ColumnKinds(schema.Video{}), owned: true},
	schema.RecordKindGallery:    {table: "Gallery", columns: db.ColumnKinds(schema.Gallery{}), owned: true},
	schema.RecordKindImage:      {table: "Image", columns: db.ColumnKinds(schema.Image{})},
	schema.RecordKindAudio:      {table: "Audio", columns: db.ColumnKinds(schema.Audio{}), owned: true},
	schema.RecordKindStory:      {table: "Story", columns: db.ColumnKinds(schema.Story{}), owned: true},
	schema.RecordKindPost:       {table: "Post", columns: db.ColumnKinds(schema.Post{}), owned: true},
	schema.RecordKindPerformer:  {table: "Performer", columns: db.ColumnKinds(schema.Performer{}), owned: true},
	schema.RecordKindStudio:     {table: "Studio", columns: db.ColumnKinds(schema.Studio{}), owned: true},
	schema.RecordKindCamModel:   {table: "CamModel", columns: db.ColumnKinds(schema.CamModel{})},
	schema.RecordKindGroup:      {table: "Group", columns: db.ColumnKinds(schema.Group{}), owned: true},
	schema.RecordKindTag:        {table: "Tag", columns: db.ColumnKinds(schema.Tag{})},
	schema.RecordKindCategory:   {table: "Category", columns: db.ColumnKinds(schema.Category{})},
	schema.RecordKindCollection: {table: "Collection", columns: db.ColumnKinds(schema.Collection{})},
	schema.RecordKindSegment:    {table: "Segment", columns: db.ColumnKinds(schema.Segment{})},
	schema.RecordKindClip:       {table: "Clip", columns: db.ColumnKinds(schema.Clip{})},
}

// bookkeepingColumns are never owned by a source; only deletedAt of them
// gets revisions.
var bookkeepingColumns = map[string]bool{
	"id":        true,
	"createdAt": true,
	"updatedAt": true,
	"deletedAt": true,
	"nameKey":   true,
}

// Table returns the table a record kind is stored in, or "" for an unknown
// kind.
func Table(kind schema.RecordKind) string {
	return recordTypes[kind].table
}

// Kinds lists every record kind.
func Kinds() []schema.RecordKind {
	kinds := make([]schema.RecordKind, 0, len(recordTypes))
	for kind := range recordTypes {
		kinds = append(kinds, kind)
	}
	return kinds
}

// Ordinal returns a column's field ordinal; ok is false for unknown columns.
func Ordinal(kind schema.RecordKind, column string) (uint32, bool) {
	ordinal, ok := schema.FieldOrdinals[Table(kind)][column]
	return ordinal, ok
}

// Column returns the column a field ordinal names, or "".
func Column(kind schema.RecordKind, ordinal uint32) string {
	for column, candidate := range schema.FieldOrdinals[Table(kind)] {
		if candidate == ordinal {
			return column
		}
	}
	return ""
}

// isOwnable reports whether a column of kind has a FieldSource.
func isOwnable(kind schema.RecordKind, column string) bool {
	return recordTypes[kind].owned && !bookkeepingColumns[column]
}

// hasRevisions reports whether changes to a column are recorded.
func hasRevisions(column string) bool {
	return column == "deletedAt" || !bookkeepingColumns[column]
}

// Author is who writes a change.
type Author struct {
	Actor schema.Actor
	// Name of the plugin or automation.
	Name string
	// SourceID is the source a plugin's values came from; "" when the record
	// was described without a site identity.
	SourceID        string
	AutomationRunID string
}

// User is the author of the user's own edits.
func User() Author {
	return Author{Actor: schema.ActorUser}
}

// Plugin is the author of values a plugin read from source.
func Plugin(name, sourceID string) Author {
	return Author{Actor: schema.ActorPlugin, Name: name, SourceID: sourceID}
}

// System is the author of the backend's own changes (merges, purges).
func System() Author {
	return Author{Actor: schema.ActorSystem}
}

// Recorder writes records with provenance and revisions.
type Recorder struct {
	database *db.DB
}

// New returns a recorder on database.
func New(database *db.DB) *Recorder {
	return &Recorder{database: database}
}
