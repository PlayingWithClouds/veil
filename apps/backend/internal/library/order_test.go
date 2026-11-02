package library

import (
	"context"
	"fmt"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// TestNewestTakesTurnsBetweenSites stores three videos of one site after two
// of another on the same day, and expects the default order to alternate
// sites, newest first within each.
func TestNewestTakesTurnsBetweenSites(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	storeListed(t, library, "other", "o1", "o2")
	storeListed(t, library, "busy", "b1", "b2", "b3")

	videos, err := library.Videos(ctx, MediaQuery{})
	if err != nil {
		t.Fatal(err)
	}
	titles := []string{}
	for _, video := range videos {
		titles = append(titles, video.Title)
	}
	want := "[b3 o2 b2 o1 b1]"
	if fmt.Sprint(titles) != want {
		t.Errorf("order = %v, want %s", titles, want)
	}
}

// storeListed stores listing cards of plugin, titled and identified by ids, in order.
func storeListed(t *testing.T, library *Library, plugin string, ids ...string) {
	t.Helper()
	for _, id := range ids {
		item := plugins.DiscoveredItem{MediaType: plugins.MediaTypeScene, ExternalID: id, SourceURL: "https://" + plugin + ".test/" + id, Title: id}
		if _, err := library.IngestListed(context.Background(), plugin, item); err != nil {
			t.Fatal(err)
		}
	}
}
