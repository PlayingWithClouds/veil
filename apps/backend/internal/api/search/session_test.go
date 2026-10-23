package search

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/searchrank"
)

// openTestHandler opens a handler on a fresh database holding two scenes: one
// credited to Mia Khalifa (alias "Mia K"), one without credits.
func openTestHandler(t *testing.T) (*Handler, *ingest.Service) {
	t.Helper()
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	service := ingest.New(database)
	for _, scene := range []*plugins.Scene{
		{Type: plugins.MediaTypeScene, SourceURL: "https://a.test/1", Title: "Hotel night",
			Performers: []plugins.ScenePerformer{{Name: "Mia Khalifa"}}},
		{Type: plugins.MediaTypeScene, SourceURL: "https://a.test/2", Title: "Garden party"},
	} {
		if _, err := service.IngestScrapeResults(ctx, "alpha", scene.SourceURL, []*plugins.ScrapeResult{{Type: plugins.MediaTypeScene, Scene: scene}}); err != nil {
			t.Fatalf("ingest: %v", err)
		}
	}
	if _, err := database.Exec(ctx, `UPDATE performer SET aliases = '["Mia K"]' WHERE name = 'Mia Khalifa'`, nil); err != nil {
		t.Fatalf("alias: %v", err)
	}
	return &Handler{ingestSvc: service, entities: newEntityDirectory(service)}, service
}

// collectingEmitter returns an emitter gathering the titles it is asked to send.
func collectingEmitter(titles *[]string) libraryEmitter {
	return libraryEmitter{
		annotate: func(*searchrank.Hit) {},
		send: func(_, _ string, hit searchrank.Hit) {
			*titles = append(*titles, hit.Item.Title)
		},
	}
}

// TestEntityDirectoryRecognisesNamesAndAliases checks the index built from the
// library finds a performer by name, misspelling and alias.
func TestEntityDirectoryRecognisesNamesAndAliases(t *testing.T) {
	handler, _ := openTestHandler(t)
	index := handler.entities.Index(context.Background())
	for _, query := range []string{"mia khalifa", "mia khalfia", "hotel mia k"} {
		matches := index.Find(query)
		if len(matches) != 1 || matches[0].Kind != searchrank.KindPerformer || matches[0].Name != "Mia Khalifa" {
			t.Fatalf("%q matched %+v", query, matches)
		}
	}
	if alias := index.Find("hotel mia k"); !alias[0].ViaAlias {
		t.Fatal("alias use should be flagged")
	}
	if matches := index.Find("garden"); len(matches) != 0 {
		t.Fatalf("garden matched %+v", matches)
	}
}

// TestSendFuzzyScenesFindsMisspelledQuery checks the fallback sends the scene a
// typo away and not the unrelated one, skipping scenes already sent.
func TestSendFuzzyScenesFindsMisspelledQuery(t *testing.T) {
	handler, _ := openTestHandler(t)
	var titles []string
	handler.sendFuzzyScenes(context.Background(), "mia khalfia", 20, map[string]bool{}, collectingEmitter(&titles))
	if len(titles) != 1 || titles[0] != "Hotel night" {
		t.Fatalf("fuzzy scenes = %v", titles)
	}

	records, err := handler.ingestSvc.SearchLibrary(context.Background(), "scene", "hotel", 5)
	if err != nil || len(records) != 1 {
		t.Fatalf("lookup: %v %v", records, err)
	}
	titles = nil
	handler.sendFuzzyScenes(context.Background(), "mia khalfia", 20, map[string]bool{recordID(records[0]): true}, collectingEmitter(&titles))
	if len(titles) != 0 {
		t.Fatalf("already sent scene came again: %v", titles)
	}
}

// TestSendLibraryHitsTriesRewriteAfterQuery checks the rewritten query finds
// what the alias alone cannot, and nothing is sent twice.
func TestSendLibraryHitsTriesRewriteAfterQuery(t *testing.T) {
	handler, _ := openTestHandler(t)
	var titles []string
	sent := handler.sendLibraryHits(context.Background(), []string{"mia k", "mia khalifa"}, 20, collectingEmitter(&titles))
	if len(sent) != 1 {
		t.Fatalf("scenes sent = %v", sent)
	}
	hotel := 0
	for _, title := range titles {
		if title == "Hotel night" {
			hotel++
		}
	}
	if hotel != 1 {
		t.Fatalf("Hotel night sent %d times in %v", hotel, titles)
	}
}

// TestAnnotateCreditsStoredAndMentionedEntities checks a scene credited to the
// recognised performer gets full credit, an uncredited stub naming her in the
// title the mention credit, and an unrelated one none.
func TestAnnotateCreditsStoredAndMentionedEntities(t *testing.T) {
	handler, service := openTestHandler(t)
	ctx := context.Background()
	session := &searchSession{handler: handler, ctx: ctx, query: "mia khalifa"}
	session.matches = handler.entities.Index(ctx).Find("mia khalifa")

	credited, _ := service.SearchLibrary(ctx, "scene", "hotel", 5)
	hit := libraryHit(plugins.MediaTypeScene, recordID(credited[0]), credited[0])
	session.annotate(&hit, "")
	if hit.EntityCredit != 1 {
		t.Fatalf("credited scene = %v", hit.EntityCredit)
	}

	unrelated, _ := service.SearchLibrary(ctx, "scene", "garden", 5)
	stub := libraryHit(plugins.MediaTypeScene, recordID(unrelated[0]), unrelated[0])
	session.annotate(&stub, "")
	if stub.EntityCredit != 0 {
		t.Fatalf("unrelated scene = %v", stub.EntityCredit)
	}
	stub.Item.Title = "Mia Khalifa in the garden"
	session.annotate(&stub, "")
	if stub.EntityCredit != searchrank.EntityMentionCredit {
		t.Fatalf("title mention = %v", stub.EntityCredit)
	}
	session.annotate(&stub, session.matches[0].ID)
	if stub.EntityCredit != 1 {
		t.Fatalf("listed on her page = %v", stub.EntityCredit)
	}
}

// TestChoosePageEntities checks fuzzy matches and tags inside a longer query
// get no page listing, performers come first and at most two are listed.
func TestChoosePageEntities(t *testing.T) {
	performer := searchrank.EntityMatch{Entity: searchrank.Entity{Kind: searchrank.KindPerformer, ID: "performer:a"}, Start: 2, End: 4}
	studio := searchrank.EntityMatch{Entity: searchrank.Entity{Kind: searchrank.KindStudio, ID: "studio:a"}, Start: 0, End: 1}
	otherStudio := searchrank.EntityMatch{Entity: searchrank.Entity{Kind: searchrank.KindStudio, ID: "studio:b"}, Start: 1, End: 2}
	tag := searchrank.EntityMatch{Entity: searchrank.Entity{Kind: searchrank.KindTag, ID: "tag:a"}, Start: 4, End: 5}
	guess := searchrank.EntityMatch{Entity: searchrank.Entity{Kind: searchrank.KindPerformer, ID: "performer:g"}, Start: 0, End: 2, Fuzzy: true}

	picked := choosePageEntities([]searchrank.EntityMatch{studio, otherStudio, performer, tag}, 5)
	if len(picked) != 2 || picked[0].ID != "performer:a" || picked[1].ID != "studio:a" {
		t.Fatalf("picked = %+v", picked)
	}
	if got := choosePageEntities([]searchrank.EntityMatch{guess, tag}, 5); len(got) != 0 {
		t.Fatalf("guessed performer and tag inside a longer query = %+v", got)
	}
	wholeTag := searchrank.EntityMatch{Entity: tag.Entity, Start: 0, End: 1}
	if got := choosePageEntities([]searchrank.EntityMatch{wholeTag}, 1); len(got) != 1 {
		t.Fatalf("a tag that is the whole query = %+v", got)
	}
}
