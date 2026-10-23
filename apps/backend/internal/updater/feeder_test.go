package updater

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/media"
	"github.com/playingwithclouds/veil/internal/recommend"
)

// fixedTaste is a TasteSource answering with one profile.
type fixedTaste struct{ profile *recommend.Profile }

// Profile returns the fixed profile.
func (taste fixedTaste) Profile(ctx context.Context) (*recommend.Profile, error) {
	return taste.profile, nil
}

// emptyProfile is a cold-start profile.
func emptyProfile() *recommend.Profile {
	return &recommend.Profile{Tags: map[string]float64{}, Performers: map[string]float64{}, Studios: map[string]float64{}}
}

// newTestFeeder opens a scratch database with a studio that has an eporner
// channel page, and returns a feeder over it plus a counter of onFed calls.
func newTestFeeder(t *testing.T, fake *fakeDiscovery, profile *recommend.Profile) (*Feeder, *db.DB, memoryMarks, *int) {
	t.Helper()
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	studio := map[string]any{"id": "studio:vixen", "name": "Vixen", "source_url": "https://www.eporner.com/channel/vixen/"}
	if _, err := database.Insert(ctx, "studio", studio); err != nil {
		t.Fatalf("insert studio: %v", err)
	}
	marks := memoryMarks{}
	fedCount := 0
	scheduler := NewScheduler(media.NewRepository(database), fake)
	feeder := NewFeeder(scheduler, database, fixedTaste{profile}, marks, func() { fedCount++ })
	return feeder, database, marks, &fedCount
}

// finds lists feed_discovery rows as "source/entity/scene#rank", ordered.
func finds(t *testing.T, database *db.DB) []string {
	t.Helper()
	rows, err := database.Strings(context.Background(),
		`SELECT source || '/' || entity || '/' || scene || '#' || rank FROM feed_discovery ORDER BY source, entity, rank`, nil)
	if err != nil {
		t.Fatalf("finds: %v", err)
	}
	return rows
}

func TestFeederRecordsCatalogScenesPerPlugin(t *testing.T) {
	fake := &fakeDiscovery{plugins: testPlugins(), searchHits: map[string][]string{
		"eporner":  {"scene:a", "gallery:skip", "scene:b", "scene:a"},
		"xhamster": {"scene:c"},
	}}
	feeder, database, marks, fedCount := newTestFeeder(t, fake, emptyProfile())

	stored, err := feeder.Run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if stored != 3 {
		t.Errorf("stored %d, want 3", stored)
	}
	assertStrings(t, "finds", finds(t, database), []string{
		"catalog/eporner/scene:a#0", "catalog/eporner/scene:b#1", "catalog/xhamster/scene:c#0",
	})
	assertStrings(t, "catalog searches", fake.searches, []string{"@eporner", "@xhamster"})
	if _, ran := marks[feedLastRunKey]; !ran || *fedCount != 1 {
		t.Errorf("a run that stored scenes is the interval's run: marks %v, fed %d", marks, *fedCount)
	}
	if marks[feedCatalogPageKey+"eporner"] != "1" {
		t.Errorf("next eporner catalog page = %q, want 1", marks[feedCatalogPageKey+"eporner"])
	}
}

func TestFeederPullsTopEntityPages(t *testing.T) {
	fake := &fakeDiscovery{plugins: testPlugins(), pageHits: map[string][]string{"eporner": {"scene:page"}}}
	profile := emptyProfile()
	profile.Studios["studio:vixen"] = 1
	profile.Studios["studio:disliked"] = -1
	feeder, database, _, _ := newTestFeeder(t, fake, profile)

	if _, err := feeder.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
	assertStrings(t, "finds", finds(t, database), []string{"taste/studio:vixen/scene:page#0"})
	assertStrings(t, "pages", fake.pages, []string{"eporner@https://www.eporner.com/channel/vixen/"})
}

func TestFeederPullsAdjacentTagPagesAsExploration(t *testing.T) {
	// Tags have no page of their own, so they are searched by name.
	fake := &fakeDiscovery{plugins: testPlugins(), searchHits: map[string][]string{"eporner": {"scene:fresh"}}}
	profile := emptyProfile()
	profile.AdjacentTags = []string{"tag:near", "tag:far", "tag:nearer", "tag:nearest"}
	feeder, database, _, _ := newTestFeeder(t, fake, profile)
	for _, name := range []string{"near", "far", "nearer", "nearest"} {
		if _, err := database.Insert(context.Background(), "tag", map[string]any{"id": "tag:" + name, "name": name}); err != nil {
			t.Fatalf("insert tag: %v", err)
		}
	}
	if _, err := database.Insert(context.Background(), "scene", map[string]any{"id": "scene:fresh", "source_url": "https://x.test/fresh"}); err != nil {
		t.Fatalf("insert scene: %v", err)
	}

	if _, err := feeder.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
	explored := 0
	for _, find := range finds(t, database) {
		if strings.HasPrefix(find, "explore/tag:") && strings.HasSuffix(find, "/scene:fresh#0") {
			explored++
		}
	}
	if explored != exploreTagLimit {
		t.Errorf("%d adjacent tags explored, want %d: %v", explored, exploreTagLimit, finds(t, database))
	}
}

func TestFeederListsPopularPagesTakingTurns(t *testing.T) {
	fake := &fakeDiscovery{plugins: testPlugins(), pageHits: map[string][]string{"eporner": {"scene:hot"}}}
	feeder, database, marks, _ := newTestFeeder(t, fake, emptyProfile())
	feeder.popularPages = []string{
		"https://www.eporner.com/best-videos/",
		"https://www.eporner.com/top-rated/",
		"https://www.eporner.com/cat/all/",
		"https://elsewhere.test/popular/",
	}

	for run := 0; run < 2; run++ {
		if _, err := feeder.Run(context.Background()); err != nil {
			t.Fatalf("run: %v", err)
		}
	}
	assertStrings(t, "pages", fake.pages, []string{
		"eporner@https://www.eporner.com/best-videos/", "eporner@https://www.eporner.com/top-rated/",
		"eporner@https://www.eporner.com/cat/all/",
	})
	assertStrings(t, "finds", finds(t, database), []string{"popular/www.eporner.com/scene:hot#0"})
	if marks[feedPopularIndexKey] != "0" {
		t.Errorf("next popular page = %q, want 0 after wrapping around", marks[feedPopularIndexKey])
	}
}

func TestFeederRetriesARunThatFoundNothing(t *testing.T) {
	feeder, _, marks, fedCount := newTestFeeder(t, &fakeDiscovery{plugins: testPlugins()}, emptyProfile())

	if _, err := feeder.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
	if _, ran := marks[feedLastRunKey]; ran || *fedCount != 0 {
		t.Errorf("an empty run must stay due: marks %v, fed %d", marks, *fedCount)
	}
	if !feeder.due(context.Background()) {
		t.Error("feeder is not due after an empty run")
	}
}

func TestFeedSitesListsOnlyTheNamedSitesOnePageDeeper(t *testing.T) {
	fake := &fakeDiscovery{plugins: testPlugins(), searchHits: map[string][]string{
		"eporner":  {"scene:a"},
		"xhamster": {"scene:c"},
	}}
	feeder, database, marks, fedCount := newTestFeeder(t, fake, emptyProfile())

	for run := 0; run < 2; run++ {
		if _, err := feeder.FeedSites(context.Background(), []string{"xhamster"}); err != nil {
			t.Fatalf("feed sites: %v", err)
		}
	}
	assertStrings(t, "finds", finds(t, database), []string{"catalog/xhamster/scene:c#0"})
	assertStrings(t, "searches", fake.searches, []string{"@xhamster", "@xhamster"})
	if marks[feedCatalogPageKey+"xhamster"] != "2" || *fedCount != 2 {
		t.Errorf("xhamster next page %q, fed %d", marks[feedCatalogPageKey+"xhamster"], *fedCount)
	}
	if _, listed := marks[feedCatalogPageKey+"eporner"]; listed {
		t.Error("an unnamed site's catalog page moved")
	}
}
