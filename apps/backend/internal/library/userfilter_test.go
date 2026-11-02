package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/user"
)

// TestFiltersOnUserState filters by favourite and rating, sorts by rating,
// and hides everything crediting a blocked performer.
func TestFiltersOnUserState(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	service := user.New(library.DB(), "")
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	second := foundScene()
	second[0].Scene.ExternalID = "v2"
	second[0].Scene.Performers = nil
	other, err := library.IngestFound(ctx, "site", second)
	if err != nil {
		t.Fatal(err)
	}
	high, low := 9.0, 3.0
	service.Rate(ctx, schema.RecordKindVideo, ref.Id, &high)
	service.Rate(ctx, schema.RecordKindVideo, other.Id, &low)
	service.SetFavorite(ctx, schema.RecordKindVideo, other.Id, true)

	favorite := true
	assertVideos(t, library, "favourites", schema.Filter{Favorite: &favorite}, other.Id)
	minimum := float32(5)
	assertVideos(t, library, "rated 5+", schema.Filter{MinRating: &minimum}, ref.Id)
	assertVideos(t, library, "by rating", schema.Filter{Sort: &schema.Sort{Field: schema.SortFieldRating, Descending: true}}, ref.Id, other.Id)

	video, _ := library.Video(ctx, ref.Id)
	if _, err := service.SetBlocked(ctx, schema.RecordKindPerformer, video.Performers[0], true); err != nil {
		t.Fatal(err)
	}
	assertVideos(t, library, "with a blocked performer", schema.Filter{}, other.Id)
	performers, _ := library.Performers(ctx, NameFilter{Text: "ada"})
	if len(performers) != 0 {
		t.Errorf("blocked performer still listed: %+v", performers)
	}
}

// assertVideos lists videos with filter and expects exactly ids, in order.
func assertVideos(t *testing.T, library *Library, name string, filter schema.Filter, ids ...string) {
	t.Helper()
	videos, err := library.Videos(context.Background(), MediaQuery{Filter: filter})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	got := make([]string, len(videos))
	for index, video := range videos {
		got[index] = video.Id
	}
	if len(got) != len(ids) {
		t.Errorf("%s: videos %v, want %v", name, got, ids)
		return
	}
	for index := range ids {
		if got[index] != ids[index] {
			t.Errorf("%s: videos %v, want %v", name, got, ids)
			return
		}
	}
}
