package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestCreditedPerformers lists the voices of audio, the most credited first,
// and leaves out performers credited only on other kinds.
func TestCreditedPerformers(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	for _, entry := range []struct {
		id     string
		voices []string
	}{
		{"a1", []string{"Rose"}},
		{"a2", []string{"Rose", "Ivy"}},
	} {
		audio := &plugins.Audio{ExternalID: entry.id, SourceURL: "https://site.test/" + entry.id, Title: entry.id}
		for _, voice := range entry.voices {
			audio.Performers = append(audio.Performers, plugins.ScenePerformer{Name: voice})
		}
		if _, err := library.IngestFound(ctx, "site", []*plugins.ScrapeResult{{Type: plugins.MediaTypeAudio, Audio: audio}}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := library.IngestFound(ctx, "site", foundScene()); err != nil {
		t.Fatal(err)
	}

	voices, err := library.CreditedPerformers(ctx, schema.MediaKindAudio, 10)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, voice := range voices {
		names = append(names, voice.Name)
	}
	if len(names) != 2 || names[0] != "Rose" || names[1] != "Ivy" {
		t.Errorf("voices = %v, want [Rose Ivy]", names)
	}
}
