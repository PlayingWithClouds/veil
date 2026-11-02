package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/user"
)

// TestCreditKeepsPerformerPage stores a credited performer's page on the
// site as their source, so the performer can be read from it later.
func TestCreditKeepsPerformerPage(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	scene := foundScene()
	scene[0].Scene.Performers = []plugins.ScenePerformer{{Name: "Ada Lane", SourceURL: "https://site.test/pornstar/ada-lane"}}
	ref, err := library.IngestFound(ctx, "site", scene)
	if err != nil {
		t.Fatal(err)
	}
	video, _ := library.Video(ctx, ref.Id)
	sources, err := library.Sources(ctx, schema.RecordKindPerformer, video.Performers[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) != 1 || sources[0].Url != "https://site.test/pornstar/ada-lane" {
		t.Errorf("performer sources = %+v", sources)
	}

	// Reading that page later, with the site's own id, lands on the same source.
	page := &plugins.Performer{Name: "Ada Lane", ExternalID: "ps-ada", SourceURL: "https://site.test/pornstar/ada-lane", Country: "DE"}
	id, err := library.upsertPerformer(ctx, "site", page)
	if err != nil {
		t.Fatal(err)
	}
	sources, _ = library.Sources(ctx, schema.RecordKindPerformer, id)
	if id != video.Performers[0] || len(sources) != 1 || sources[0].ExternalId != "ps-ada" {
		t.Errorf("after reading the page: performer %s, sources %+v", id, sources)
	}
}

// TestPerformerDuplicatesByAlias suggests a performer whose alias is the
// other's name, and not an unrelated one.
func TestPerformerDuplicatesByAlias(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ada, err := library.upsertPerformer(ctx, "site", &plugins.Performer{Name: "Ada Lane", Aliases: []string{"Ada L."}})
	if err != nil {
		t.Fatal(err)
	}
	alias, _ := library.upsertPerformer(ctx, "other", &plugins.Performer{Name: "Ada L", ExternalID: "a1"})
	library.upsertPerformer(ctx, "other", &plugins.Performer{Name: "Bea", ExternalID: "b1"})

	duplicates, err := library.PerformerDuplicates(ctx, ada)
	if err != nil {
		t.Fatal(err)
	}
	if len(duplicates) != 1 || duplicates[0].Id != alias {
		t.Errorf("duplicates = %+v, want only %s", duplicates, alias)
	}
}

// TestEditPerformerAttributesAndPicture replaces the attributes, sets and
// clears the picture.
func TestEditPerformerAttributesAndPicture(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	id, _ := library.upsertPerformer(ctx, "site", &plugins.Performer{Name: "Ada Lane", Country: "DE", ImagePath: "https://cdn.test/old.jpg"})

	height := int32(170)
	picture := "https://cdn.test/new.jpg"
	edit := NamedEdit{Attributes: &schema.PerformerAttributes{HeightCm: &height}, Image: &picture}
	if err := library.EditNamed(ctx, schema.RecordKindPerformer, id, edit); err != nil {
		t.Fatal(err)
	}
	performer, _ := library.Performer(ctx, id)
	if performer.Attributes.HeightCm == nil || *performer.Attributes.HeightCm != 170 || performer.Attributes.Country != nil {
		t.Errorf("attributes = %+v", performer.Attributes)
	}
	if performer.Image == nil || performer.Image.Uri != picture {
		t.Errorf("image = %+v", performer.Image)
	}

	cleared := ""
	if err := library.EditNamed(ctx, schema.RecordKindPerformer, id, NamedEdit{Image: &cleared}); err != nil {
		t.Fatal(err)
	}
	performer, _ = library.Performer(ctx, id)
	if performer.Image != nil {
		t.Errorf("image after clearing = %+v", performer.Image)
	}
}

// TestAttachIdentifiedProfile adds a site's profile to the performer it was
// identified as, and leaves a profile that is another performer's source.
func TestAttachIdentifiedProfile(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	mia, _ := library.upsertPerformer(ctx, "spankbang", &plugins.Performer{Name: "Mia Malkova", SourceURL: "https://spankbang.test/mia"})
	profile := &plugins.Performer{Name: "Mia Malkova", ExternalID: "mia-malkova", SourceURL: "https://freeones.test/mia-malkova", Country: "US"}

	attached, err := library.AttachPerformerProfile(ctx, mia, "freeones", profile)
	if err != nil || !attached {
		t.Fatalf("attached = %v, err = %v", attached, err)
	}
	performer, _ := library.Performer(ctx, mia)
	sources, _ := library.Sources(ctx, schema.RecordKindPerformer, mia)
	if performer.Attributes.Country == nil || len(sources) != 2 || sources[1].FetchedAt == nil {
		t.Errorf("after attaching: attributes %+v, sources %+v", performer.Attributes, sources)
	}

	other, _ := library.upsertPerformer(ctx, "eporner", &plugins.Performer{Name: "Mia M", ExternalID: "m1"})
	attached, err = library.AttachPerformerProfile(ctx, other, "freeones", profile)
	if err != nil || attached {
		t.Errorf("profile of another performer: attached = %v, err = %v", attached, err)
	}
}

// TestCreditPerformerFromTheirList credits a performer on a listed video
// once, and reads a favourite as the user marked it.
func TestCreditPerformerFromTheirList(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ada, _ := library.upsertPerformer(ctx, "site", &plugins.Performer{Name: "Ada Lane", ExternalID: "ps-ada"})
	ref, err := library.IngestListed(ctx, "site", plugins.DiscoveredItem{MediaType: plugins.MediaTypeScene, Title: "Pool day", ExternalID: "v1", SourceURL: "https://site.test/v1"})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := library.CreditPerformer(ctx, ref.Id, ada, "site", ""); err != nil {
			t.Fatal(err)
		}
	}
	video, _ := library.Video(ctx, ref.Id)
	if len(video.Performers) != 1 || video.Performers[0] != ada {
		t.Errorf("performers = %v, want [%s]", video.Performers, ada)
	}

	favorite, err := library.IsFavorite(ctx, schema.RecordKindPerformer, ada)
	if err != nil || favorite {
		t.Fatalf("favourite before marking = %v, %v", favorite, err)
	}
	user.New(library.database, "device").SetFavorite(ctx, schema.RecordKindPerformer, ada, true)
	if favorite, _ := library.IsFavorite(ctx, schema.RecordKindPerformer, ada); !favorite {
		t.Error("favourite not read after marking")
	}
}
