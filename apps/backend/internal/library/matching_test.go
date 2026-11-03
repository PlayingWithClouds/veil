package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestMatchingNarrowsStories keeps only the listed stories a search's filter
// matches, in their order.
func TestMatchingNarrowsStories(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	refs := []schema.RecordRef{}
	for _, title := range []string{"Summer at the lake", "Office hours", "Lake house"} {
		ref, err := library.IngestListed(ctx, "site", plugins.DiscoveredItem{
			MediaType: plugins.MediaTypeStory, ExternalID: title, SourceURL: "https://site.test/" + title, Title: title,
		})
		if err != nil {
			t.Fatal(err)
		}
		refs = append(refs, ref)
	}

	text := "lake"
	matching, err := library.Matching(ctx, refs, schema.Filter{Text: &text})
	if err != nil {
		t.Fatal(err)
	}
	if len(matching) != 2 || matching[0] != refs[0] || matching[1] != refs[2] {
		t.Errorf("matching = %+v, want the first and last of %+v", matching, refs)
	}
}
