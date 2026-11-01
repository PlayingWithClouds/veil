package db

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/schema"
)

// openTestDB opens a fresh migrated database in a temp dir.
func openTestDB(t *testing.T) *DB {
	t.Helper()
	database, err := Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}

// TestOpenIsIdempotent opens the same file twice; the second run applies nothing.
func TestOpenIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	for attempt := range 2 {
		database, err := Open(context.Background(), path)
		if err != nil {
			t.Fatalf("open attempt %d: %v", attempt, err)
		}
		database.Close()
	}
}

// TestInsertGetRoundTrip writes a performer with lists, a value object and a
// map and reads it back through the generated type.
func TestInsertGetRoundTrip(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	country := "NL"
	performer := schema.Performer{
		Name:       "Ada",
		Aliases:    []string{"A.", "Ada L"},
		Image:      &schema.Asset{Uri: "blob://ada"},
		Attributes: schema.PerformerAttributes{Country: &country},
	}
	id, err := database.Insert(ctx, "Performer", performer)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Get[schema.Performer](ctx, database, "Performer", id)
	if err != nil || loaded == nil {
		t.Fatalf("get: %v %v", loaded, err)
	}
	if loaded.Id != id || loaded.Name != "Ada" || len(loaded.Aliases) != 2 {
		t.Errorf("loaded = %+v", loaded)
	}
	if loaded.Image == nil || loaded.Image.Uri != "blob://ada" {
		t.Errorf("image = %+v", loaded.Image)
	}
	if loaded.Attributes.Country == nil || *loaded.Attributes.Country != "NL" {
		t.Errorf("attributes = %+v", loaded.Attributes)
	}
	if loaded.CreatedAt == "" || len(loaded.Links) != 0 {
		t.Errorf("createdAt %q, links %v", loaded.CreatedAt, loaded.Links)
	}
}

// TestBoolsAndKeywordTables checks a boolean column and a table named after
// an SQL keyword.
func TestBoolsAndKeywordTables(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	camID, err := database.Insert(ctx, "CamModel", schema.CamModel{Plugin: "chaturbate", Username: "nova", DisplayName: "Nova", AutoRecord: true})
	if err != nil {
		t.Fatal(err)
	}
	cam, err := Get[schema.CamModel](ctx, database, "CamModel", camID)
	if err != nil || cam == nil || !cam.AutoRecord {
		t.Fatalf("cam = %+v, err %v", cam, err)
	}
	if _, err := database.Insert(ctx, "Group", schema.Group{Title: "Part one"}); err != nil {
		t.Fatalf("insert into Group: %v", err)
	}
}

// TestTrashHidesRow checks that a trashed row no longer loads.
func TestTrashHidesRow(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	id, err := database.Insert(ctx, "Tag", schema.Tag{Name: "outdoor"})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Trash(ctx, "Tag", id); err != nil {
		t.Fatal(err)
	}
	tag, err := Get[schema.Tag](ctx, database, "Tag", id)
	if err != nil || tag != nil {
		t.Fatalf("trashed tag = %+v, err %v", tag, err)
	}
}

// TestForeignKeysAreEnforced rejects a video pointing at a missing studio.
func TestForeignKeysAreEnforced(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	missing := NewID()
	video := schema.Video{Media: schema.Media{Title: "x", Studio: &missing, Status: schema.MediaStatusAvailable}}
	if _, err := database.Insert(ctx, "Video", video); err == nil {
		t.Fatal("insert with a missing studio succeeded")
	}
}

// TestUpdateBumpsUpdatedAt changes a field and the timestamp.
func TestUpdateBumpsUpdatedAt(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)
	id, err := database.Insert(ctx, "Studio", schema.Studio{Entity: schema.Entity{UpdatedAt: "2000-01-01T00:00:00.000Z"}, Name: "Old"})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Update(ctx, "Studio", id, Vars{"name": "New", "aliases": []string{"Old"}}); err != nil {
		t.Fatal(err)
	}
	studio, err := Get[schema.Studio](ctx, database, "Studio", id)
	if err != nil || studio.Name != "New" || len(studio.Aliases) != 1 || studio.UpdatedAt == "2000-01-01T00:00:00.000Z" {
		t.Fatalf("studio = %+v, err %v", studio, err)
	}
}
