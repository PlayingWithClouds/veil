package recommend

import (
	"context"
	"encoding/binary"
	"math"
	"path/filepath"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// fixture is a fresh database with an engine on it.
type fixture struct {
	t        *testing.T
	ctx      context.Context
	database *db.DB
	engine   *Engine
	now      time.Time
}

// newFixture opens an engine on a fresh database.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return &fixture{t: t, ctx: ctx, database: database, engine: New(database), now: time.Now()}
}

// insert writes a row and returns its id.
func (f *fixture) insert(table string, row db.Vars) string {
	f.t.Helper()
	id, err := f.database.Insert(f.ctx, table, row)
	if err != nil {
		f.t.Fatal(err)
	}
	return id
}

// named inserts a performer, studio or tag by name.
func (f *fixture) named(table, name string) string {
	row := db.Vars{"name": name, "nameKey": name, "aliases": []string{}}
	if table != "Tag" {
		row["tags"] = []string{}
	}
	if table == "Performer" {
		row["links"] = []string{}
		row["attributes"] = map[string]any{}
	}
	return f.insert(table, row)
}

// mediaRow is a video or gallery to insert; zero fields are left out.
type mediaRow struct {
	title      string
	performers []string
	tags       []string
	studio     string
	duration   float64
	width      int
	height     int
	date       string
	gallery    string
	video      string
	images     int
}

// video inserts a video.
func (f *fixture) video(row mediaRow) string {
	vars := mediaVars(row)
	if row.duration > 0 {
		vars["durationSeconds"] = row.duration
	}
	if row.width > 0 {
		vars["width"] = row.width
		vars["height"] = row.height
	}
	if row.gallery != "" {
		vars["gallery"] = row.gallery
	}
	return f.insert("Video", vars)
}

// galleryRow inserts a gallery.
func (f *fixture) galleryRow(row mediaRow) string {
	vars := mediaVars(row)
	vars["imageCount"] = row.images
	if row.video != "" {
		vars["video"] = row.video
	}
	return f.insert("Gallery", vars)
}

// mediaVars are the columns videos and galleries share.
func mediaVars(row mediaRow) db.Vars {
	vars := db.Vars{"title": row.title, "status": "available", "performers": nonNil(row.performers), "tags": nonNil(row.tags)}
	if row.studio != "" {
		vars["studio"] = row.studio
	}
	if row.date != "" {
		vars["date"] = row.date
	}
	return vars
}

// nonNil turns a nil list into an empty one.
func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

// event logs an event about a subject ago before now.
func (f *fixture) event(kind schema.EventKind, subjectKind schema.RecordKind, subjectID string, ago time.Duration, data map[string]any) {
	row := db.Vars{"at": db.FormatTime(f.now.Add(-ago)), "kind": kind, "subjectKind": subjectKind, "subjectId": subjectID}
	if data != nil {
		row["data"] = data
	}
	f.insert("Event", row)
}

// watched logs a session-ending watch of a whole video ago before now.
func (f *fixture) watched(videoID string, seconds float64, ago time.Duration) {
	f.event(schema.EventKindView, schema.RecordKindVideo, videoID, ago+time.Duration(seconds)*time.Second, nil)
	f.event(schema.EventKindHeatSpan, schema.RecordKindVideo, videoID, ago, map[string]any{"start": 0, "end": seconds})
	f.event(schema.EventKindFinish, schema.RecordKindVideo, videoID, ago, map[string]any{"position": seconds})
}

// state writes a user state row.
func (f *fixture) state(kind schema.RecordKind, id string, fields db.Vars) {
	row := db.Vars{"subjectKind": kind, "subjectId": id, "favorite": false, "blocked": false, "viewCount": 0, "fields": map[string]any{}}
	for name, value := range fields {
		row[name] = value
	}
	f.insert("UserState", row)
}

// recommendations reads the first page of recommendations.
func (f *fixture) recommendations(limit int) []schema.Match {
	f.t.Helper()
	matches, err := f.engine.Recommendations(f.ctx, schema.Filter{}, limit, 0)
	if err != nil {
		f.t.Fatal(err)
	}
	return matches
}

// positionOf is the 0-based position of id among matches; -1 when absent.
func positionOf(matches []schema.Match, id string) int {
	for index, match := range matches {
		if match.Item.Id == id {
			return index
		}
	}
	return -1
}

// evidenceOf is the details of a match's evidence.
func evidenceOf(match schema.Match) []string {
	out := []string{}
	for _, entry := range match.Evidence {
		if entry.Detail != nil {
			out = append(out, *entry.Detail)
		}
	}
	return out
}

// encodeVector writes float32s little-endian, as Embedding stores them.
func encodeVector(values ...float32) []byte {
	out := make([]byte, 4*len(values))
	for index, value := range values {
		binary.LittleEndian.PutUint32(out[index*4:], math.Float32bits(value))
	}
	return out
}
