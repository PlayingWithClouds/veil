// Package library stores and reads the core records: videos, galleries and
// their images, performers, studios, tags, and the sources and streams they
// come with. Plugin results enter through Ingest*, the user's changes through
// Edit*, Trash and Restore; everything else reads.
//
// Identity: a record found on a site has one Source row per site
// (plugin + externalId, unique). Ingesting the same site record again
// updates the record it already points at. Performers, studios and tags
// without a site identity match by their folded name (names.Key).
//
// Updates of records go through provenance.Recorder, so plugins only
// overwrite fields their source owns and every change has a revision.
package library

import (
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Table names, exactly as the schema names the models.
const (
	tableVideo     = "Video"
	tableGallery   = "Gallery"
	tableImage     = "Image"
	tablePerformer = "Performer"
	tableStudio    = "Studio"
	tableTag       = "Tag"
	tableSource    = "Source"
	tableStream    = "Stream"
)

// Library is the store of core records.
type Library struct {
	database *db.DB
	recorder *provenance.Recorder
}

// New returns a library on database.
func New(database *db.DB) *Library {
	return &Library{database: database, recorder: provenance.New(database)}
}

// DB exposes the database for callers that run their own queries.
func (l *Library) DB() *db.DB {
	return l.database
}

// Recorder exposes the write path, for revisions and field provenance.
func (l *Library) Recorder() *provenance.Recorder {
	return l.recorder
}

// tableFor returns the table a record kind lives in, or "" for kinds the
// library doesn't store yet.
func tableFor(kind schema.RecordKind) string {
	return provenance.Table(kind)
}
