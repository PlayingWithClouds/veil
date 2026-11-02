package library

import (
	"context"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestUserEditSurvivesRescrape edits a title, scrapes the page again with
// another title, and expects the user's title to stay.
func TestUserEditSurvivesRescrape(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	title := "My title"
	if err := library.EditMedia(ctx, schema.RecordKindVideo, ref.Id, MediaEdit{Title: &title}); err != nil {
		t.Fatal(err)
	}
	again := foundScene()
	again[0].Scene.Title = "Site title"
	again[0].Scene.Details = "Now with details"
	if _, err := library.IngestFound(ctx, "site", again); err != nil {
		t.Fatal(err)
	}
	video, _ := library.Video(ctx, ref.Id)
	if video.Title != "My title" || video.Details == nil || *video.Details != "Now with details" {
		t.Errorf("video = %q / %v, want the user's title and the site's details", video.Title, video.Details)
	}
	owners, err := library.Recorder().FieldSources(ctx, schema.RecordKindVideo, ref.Id)
	if err != nil {
		t.Fatal(err)
	}
	if !hasLockedField(owners, 10) {
		t.Errorf("title isn't locked: %+v", owners)
	}
}

// hasLockedField reports whether the field with ordinal is locked by the user.
func hasLockedField(owners []schema.FieldSource, ordinal uint32) bool {
	for _, owner := range owners {
		if owner.FieldOrdinal == ordinal {
			return owner.Locked && owner.WrittenBy == schema.ActorUser
		}
	}
	return false
}

// TestSecondSiteOnlyFillsEmptyFields describes one performer on two sites
// and expects the first site to keep its fields and the second to add new ones.
func TestSecondSiteOnlyFillsEmptyFields(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	first := []*plugins.ScrapeResult{{Type: plugins.MediaTypePerformer, Performer: &plugins.Performer{ExternalID: "ada", Name: "Ada Lane", Details: "From A"}}}
	second := []*plugins.ScrapeResult{{Type: plugins.MediaTypePerformer, Performer: &plugins.Performer{ExternalID: "ada-b", Name: "Ada Lane", Details: "From B", Country: "NL"}}}
	ref, err := library.IngestFound(ctx, "a", first)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := library.IngestFound(ctx, "b", second); err != nil {
		t.Fatal(err)
	}
	performer, _ := library.Performer(ctx, ref.Id)
	if performer.Details == nil || *performer.Details != "From A" {
		t.Errorf("details = %v, want site a's", performer.Details)
	}
	if performer.Attributes.Country == nil || *performer.Attributes.Country != "NL" {
		t.Errorf("attributes = %+v, want site b's country", performer.Attributes)
	}
}

// TestRevertPutsTheOldValueBack changes a title by re-scraping, reverts the
// revision and expects the old title, locked.
func TestRevertPutsTheOldValueBack(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	again := foundScene()
	again[0].Scene.Title = "Renamed"
	if _, err := library.IngestFound(ctx, "site", again); err != nil {
		t.Fatal(err)
	}
	revisions, err := library.Recorder().Revisions(ctx, schema.RecordKindVideo, ref.Id, 10, 0)
	if err != nil || len(revisions) != 1 || revisions[0].Before != "Pool day" || revisions[0].After != "Renamed" {
		t.Fatalf("revisions = %+v, err %v", revisions, err)
	}
	if err := library.Recorder().Revert(ctx, revisions[0].Id); err != nil {
		t.Fatal(err)
	}
	if _, err := library.IngestFound(ctx, "site", again); err != nil {
		t.Fatal(err)
	}
	video, _ := library.Video(ctx, ref.Id)
	if video.Title != "Pool day" {
		t.Errorf("title = %q after revert and rescrape", video.Title)
	}
}

// TestTrashRestoreAndPurge trashes and restores a video, then purges a
// trashed performer and expects the video's credit to go with it.
func TestTrashRestoreAndPurge(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	if err := library.Trash(ctx, schema.RecordKindVideo, ref.Id); err != nil {
		t.Fatal(err)
	}
	if video, _ := library.Video(ctx, ref.Id); video != nil {
		t.Fatal("trashed video still loads")
	}
	trashed, err := library.Trashed(ctx, schema.RecordKindVideo, 0, 0)
	if err != nil || len(trashed) != 1 || trashed[0].Label != "Pool day" {
		t.Fatalf("trash = %+v, err %v", trashed, err)
	}
	if err := library.Restore(ctx, schema.RecordKindVideo, ref.Id); err != nil {
		t.Fatal(err)
	}
	video, _ := library.Video(ctx, ref.Id)
	if video == nil {
		t.Fatal("restored video doesn't load")
	}
	performerID := video.Performers[0]
	if err := library.Trash(ctx, schema.RecordKindPerformer, performerID); err != nil {
		t.Fatal(err)
	}
	purged, err := library.PurgeTrash(ctx, time.Now().Add(time.Minute))
	if err != nil || purged != 1 {
		t.Fatalf("purged %d, err %v", purged, err)
	}
	video, _ = library.Video(ctx, ref.Id)
	if len(video.Performers) != 1 || video.Performers[0] == performerID {
		t.Errorf("performers = %v, want the purged one gone", video.Performers)
	}
}
