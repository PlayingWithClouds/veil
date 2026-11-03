package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// TestListedUploaderCredits credits an audio card's uploader as its voice
// while the audio credits nobody (also one listed before cards named
// voices), and leaves the credits of the page read alone when the card
// comes again.
func TestListedUploaderCredits(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	card := plugins.DiscoveredItem{
		MediaType: plugins.MediaTypeAudio, ExternalID: "a1", SourceURL: "https://site.test/a1", Title: "Whispers",
	}
	if _, err := library.IngestListed(ctx, "site", card); err != nil {
		t.Fatal(err)
	}
	card.Studio = &plugins.StudioRef{Name: "Violet", ExternalID: "violet"}
	ref, err := library.IngestListed(ctx, "site", card)
	if err != nil {
		t.Fatal(err)
	}
	audio, err := library.Audio(ctx, ref.Id)
	if err != nil {
		t.Fatal(err)
	}
	if len(audio.Performers) != 1 || audio.Studio != nil {
		t.Fatalf("after the card, performers = %v, studio = %v", audio.Performers, audio.Studio)
	}

	page := &plugins.Audio{
		ExternalID: "a1", SourceURL: "https://site.test/a1", Title: "Whispers",
		Performers: []plugins.ScenePerformer{{Name: "Violet", ExternalID: "violet"}, {Name: "Rose"}},
	}
	if _, err := library.IngestFound(ctx, "site", []*plugins.ScrapeResult{{Type: plugins.MediaTypeAudio, Audio: page}}); err != nil {
		t.Fatal(err)
	}
	if _, err := library.IngestListed(ctx, "site", card); err != nil {
		t.Fatal(err)
	}
	audio, err = library.Audio(ctx, ref.Id)
	if err != nil {
		t.Fatal(err)
	}
	if len(audio.Performers) != 2 {
		t.Errorf("after the page and the card again, performers = %v, want both voices", audio.Performers)
	}
}
