package ingest

import (
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// ingestPerformerPage ingests a performer the way a plugin's performer page does.
func ingestPerformerPage(t *testing.T, service *Service, pluginName string, performer *plugins.Performer) *IngestResult {
	t.Helper()
	performer.Type = plugins.MediaTypePerformer
	result, err := service.IngestScrapeResults(t.Context(), pluginName, performer.SourceURL, []*plugins.ScrapeResult{
		{Type: plugins.MediaTypePerformer, Performer: performer},
	})
	if err != nil {
		t.Fatalf("ingest performer page: %v", err)
	}
	return result
}

// TestCreditSpellingsShareOnePerformer checks that credits differing only in
// case or punctuation resolve to one performer record.
func TestCreditSpellingsShareOnePerformer(t *testing.T) {
	service, database, ctx := openTestService(t)

	for index, credit := range []string{"Isla Summer", "ISLA SUMMER", "isla-summer"} {
		ingestScene(t, ctx, service, "alpha", &plugins.Scene{
			Type: plugins.MediaTypeScene, SourceURL: testSceneURL + string(rune('a'+index)), Title: "Scene",
			Performers: []plugins.ScenePerformer{{Name: credit}},
		})
	}

	if performers := countRows(t, ctx, database, `SELECT count(*) FROM performer`, nil); performers != 1 {
		t.Fatalf("performer rows = %d, want 1", performers)
	}
}

// TestPerformerPageAdoptsCreditRecord checks that a performer page finds the
// record a scene credit created instead of making a second one, and that a
// second site's page for the same name joins it too and is recognised again.
func TestPerformerPageAdoptsCreditRecord(t *testing.T) {
	service, database, ctx := openTestService(t)
	ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: testSceneURL, Title: "Scene",
		Performers: []plugins.ScenePerformer{{Name: "Zoe Bloom"}},
	})

	first := ingestPerformerPage(t, service, "eporner", &plugins.Performer{
		Name: "Zoe Bloom", SourceURL: "https://eporner.test/pornstar/zoe-bloom", ExternalID: "eporner-ps-zoe",
	})
	second := ingestPerformerPage(t, service, "hqporner", &plugins.Performer{
		Name: "zoe bloom", SourceURL: "https://hqporner.test/actress/zoe-bloom", ExternalID: "hqporner-actress-zoe",
	})
	again := ingestPerformerPage(t, service, "hqporner", &plugins.Performer{
		Name: "zoe bloom", SourceURL: "https://hqporner.test/actress/zoe-bloom", ExternalID: "hqporner-actress-zoe",
	})

	if first.IsNew || second.IsNew || again.IsNew {
		t.Fatalf("a page created a record: first=%v second=%v again=%v", first.IsNew, second.IsNew, again.IsNew)
	}
	if performers := countRows(t, ctx, database, `SELECT count(*) FROM performer`, nil); performers != 1 {
		t.Fatalf("performer rows = %d, want 1", performers)
	}
	if identities := countRows(t, ctx, database, `SELECT count(*) FROM entity_identity WHERE record = $id`,
		map[string]any{"id": first.CanonicalID}); identities != 2 {
		t.Fatalf("alternate identities = %d, want the second site's source_url and external_id", identities)
	}
}

// TestPerformerPageResolvesBeforeAnyCredit checks that the credit finds a
// record a performer page created.
func TestPerformerPageResolvesBeforeAnyCredit(t *testing.T) {
	service, database, ctx := openTestService(t)
	page := ingestPerformerPage(t, service, "eporner", &plugins.Performer{
		Name: "Alexis Texas", SourceURL: "https://eporner.test/pornstar/alexis-texas", ExternalID: "eporner-ps-alexis",
	})

	scene := ingestScene(t, ctx, service, "hqporner", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: testSceneURL, Title: "Scene",
		Performers: []plugins.ScenePerformer{{Name: "alexis texas"}},
	})

	if performers := countRows(t, ctx, database, `SELECT count(*) FROM performer`, nil); performers != 1 {
		t.Fatalf("performer rows = %d, want 1", performers)
	}
	credited := countRows(t, ctx, database,
		`SELECT count(*) FROM scene WHERE id = $scene AND EXISTS (SELECT 1 FROM json_each(scene.performers) WHERE value = $performer)`,
		map[string]any{"scene": scene.CanonicalID, "performer": page.CanonicalID})
	if credited != 1 {
		t.Fatal("scene is not credited to the performer page's record")
	}
}

// ingestStudioPage ingests a studio the way a plugin's channel page does.
func ingestStudioPage(t *testing.T, service *Service, pluginName string, studio *plugins.Studio) *IngestResult {
	t.Helper()
	studio.Type = plugins.MediaTypeStudio
	result, err := service.IngestScrapeResults(t.Context(), pluginName, studio.SourceURL, []*plugins.ScrapeResult{
		{Type: plugins.MediaTypeStudio, Studio: studio},
	})
	if err != nil {
		t.Fatalf("ingest studio page: %v", err)
	}
	return result
}

// TestStudioSpellingsShareOneStudio checks that case, spacing and punctuation
// variants of one brand resolve to a single studio.
func TestStudioSpellingsShareOneStudio(t *testing.T) {
	service, database, ctx := openTestService(t)

	for index, name := range []string{"Jules Jordan", "JULES JORDAN", "JulesJordan", "jules-jordan"} {
		ingestScene(t, ctx, service, "alpha", &plugins.Scene{
			Type: plugins.MediaTypeScene, SourceURL: testSceneURL + string(rune('a'+index)), Title: "Scene",
			Studio: &plugins.StudioRef{Name: name},
		})
	}

	if studios := countRows(t, ctx, database, `SELECT count(*) FROM studio`, nil); studios != 1 {
		t.Fatalf("studio rows = %d, want 1", studios)
	}
}

// TestChannelPageAdoptsCreditedStudio checks that a channel page finds the
// studio a scene credit created, that a second site's page for the same name
// joins it, and that the second page is recognised on its next scrape.
func TestChannelPageAdoptsCreditedStudio(t *testing.T) {
	service, database, ctx := openTestService(t)
	ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: testSceneURL, Title: "Scene",
		Studio: &plugins.StudioRef{Name: "Tushy"},
	})

	first := ingestStudioPage(t, service, "eporner", &plugins.Studio{
		Name: "Tushy", SourceURL: "https://eporner.test/channel/tushy/", ExternalID: "eporner-channel-tushy",
	})
	second := ingestStudioPage(t, service, "xhamster", &plugins.Studio{
		Name: "TUSHY", SourceURL: "https://xhamster.test/channels/tushy", ExternalID: "xhamster-channel-tushy",
	})
	again := ingestStudioPage(t, service, "xhamster", &plugins.Studio{
		Name: "TUSHY", SourceURL: "https://xhamster.test/channels/tushy", ExternalID: "xhamster-channel-tushy",
	})

	if first.IsNew || second.IsNew || again.IsNew {
		t.Fatalf("a page created a studio: first=%v second=%v again=%v", first.IsNew, second.IsNew, again.IsNew)
	}
	if studios := countRows(t, ctx, database, `SELECT count(*) FROM studio`, nil); studios != 1 {
		t.Fatalf("studio rows = %d, want 1", studios)
	}
}

// TestSceneStudioRefFindsMergedChannel checks that a scene whose studio ref
// carries the second site's channel URL resolves to the studio that page was
// registered on, instead of making a new one.
func TestSceneStudioRefFindsMergedChannel(t *testing.T) {
	service, database, ctx := openTestService(t)
	first := ingestStudioPage(t, service, "eporner", &plugins.Studio{
		Name: "Vixen", SourceURL: "https://eporner.test/channel/vixen/", ExternalID: "eporner-channel-vixen",
	})
	ingestStudioPage(t, service, "xhamster", &plugins.Studio{
		Name: "VIXEN", SourceURL: "https://xhamster.test/channels/vixen", ExternalID: "xhamster-channel-vixen",
	})

	scene := ingestScene(t, ctx, service, "xhamster", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: testSceneURL, Title: "Scene",
		// A different spelling, so only the registered page identity can match.
		Studio: &plugins.StudioRef{Name: "Vixen Media Group", SourceURL: "https://xhamster.test/channels/vixen"},
	})

	if studios := countRows(t, ctx, database, `SELECT count(*) FROM studio`, nil); studios != 1 {
		t.Fatalf("studio rows = %d, want 1", studios)
	}
	linked := countRows(t, ctx, database, `SELECT count(*) FROM scene WHERE id = $scene AND studio = $studio`,
		map[string]any{"scene": scene.CanonicalID, "studio": first.CanonicalID})
	if linked != 1 {
		t.Fatal("scene is not linked to the studio the channel page was registered on")
	}
}
