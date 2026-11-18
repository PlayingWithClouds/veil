package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// performerPage is a performer's profile page as a plugin's performer:find
// returns it.
func performerPage(externalID, birthdate, details string) []*plugins.ScrapeResult {
	return []*plugins.ScrapeResult{{Type: plugins.MediaTypePerformer, Performer: &plugins.Performer{
		ExternalID: externalID, SourceURL: "https://profiles.test/" + externalID, Name: "Ada Lane", Birthdate: birthdate, Details: details,
	}}}
}

// TestPreferredSiteTakesFieldsOver lets a profile database replace a tube
// site's facts about a performer, while a later tube site can't replace the
// database's.
func TestPreferredSiteTakesFieldsOver(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "xhamster", performerPage("x1", "1990-01-01", "Tube bio"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := library.IngestFound(ctx, "freeones", performerPage("f1", "1995-04-12", "Database bio")); err != nil {
		t.Fatal(err)
	}
	if _, err := library.IngestFound(ctx, "eporner", performerPage("e1", "1980-01-01", "Other tube bio")); err != nil {
		t.Fatal(err)
	}
	performer, _ := library.Performer(ctx, ref.Id)
	if performer == nil || performer.Attributes.Birthdate == nil || *performer.Attributes.Birthdate != "1995-04-12" {
		t.Errorf("birthdate = %v, want the database's", performer.Attributes.Birthdate)
	}
	if performer.Details == nil || *performer.Details != "Database bio" {
		t.Errorf("details = %v, want the database's", performer.Details)
	}
}
