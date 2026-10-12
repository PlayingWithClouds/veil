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
