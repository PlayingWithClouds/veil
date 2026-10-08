package updater

import (
	"context"
	"errors"
	"path/filepath"
	"sort"
	"testing"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/discovery"
	"github.com/playingwithclouds/veil/internal/media"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// fakeDiscovery answers searches and page listings from canned media ids and
// records every call.
type fakeDiscovery struct {
	plugins    []*plugins.Plugin
	searchHits map[string][]string
	pageHits   map[string][]string
	pageErrors map[string]error
	// searches holds "query@plugin" per plugin searched; pages "plugin@url".
	searches []string
	pages    []string
}

// Plugins returns the fake's plugins, restricted to sources when given.
func (f *fakeDiscovery) Plugins(sources []string) []*plugins.Plugin {
	if len(sources) == 0 {
		return f.plugins
	}
	var matched []*plugins.Plugin
	for _, plugin := range f.plugins {
		for _, source := range sources {
			if plugin.Meta.Name == source {
				matched = append(matched, plugin)
				break
			}
		}
	}
	return matched
}

// Run reports each plugin's canned search hits.
func (f *fakeDiscovery) Run(ctx context.Context, options discovery.Options, searchPlugins []*plugins.Plugin, onHit func(discovery.Hit)) error {
	for _, plugin := range searchPlugins {
		f.searches = append(f.searches, options.Query+"@"+plugin.Meta.Name)
		for _, id := range f.searchHits[plugin.Meta.Name] {
			onHit(discovery.Hit{Plugin: plugin.Meta.Name, ID: id})
		}
	}
	return nil
}

// ListPage reports the plugin's canned page hits, or its canned error.
func (f *fakeDiscovery) ListPage(ctx context.Context, plugin *plugins.Plugin, pageURL string, limit int, onHit func(discovery.Hit)) error {
	f.pages = append(f.pages, plugin.Meta.Name+"@"+pageURL)
	if err := f.pageErrors[plugin.Meta.Name]; err != nil {
		return err
	}
	for _, id := range f.pageHits[plugin.Meta.Name] {
		onHit(discovery.Hit{Plugin: plugin.Meta.Name, ID: id})
	}
	return nil
}

// testPlugins are an eporner-like plugin that lists channel pages and a
// search-only xhamster-like one.
func testPlugins() []*plugins.Plugin {
	return []*plugins.Plugin{
		{Meta: plugins.PluginMeta{
			Name:         "eporner",
			Capabilities: []plugins.Capability{plugins.CapabilitySceneList, plugins.CapabilitySceneListPage},
			Domains:      []string{"eporner.com"},
		}},
		{Meta: plugins.PluginMeta{
			Name:         "xhamster",
			Capabilities: []plugins.Capability{plugins.CapabilitySceneList},
			Domains:      []string{"xhamster.com"},
		}},
	}
}

// seedRow is a row inserted into the scratch database.
type seedRow struct {
	table  string
	fields map[string]any
}

// newTestScheduler opens a scratch database seeded with a followed studio and
// performer, their credited scene and a few stubs, plus any extra rows.
func newTestScheduler(t *testing.T, fake *fakeDiscovery, extra ...seedRow) (*Scheduler, *media.Repository) {
	t.Helper()
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	rows := []seedRow{
		{"studio", map[string]any{"id": "studio:vixen", "name": "Vixen", "source_url": "https://www.eporner.com/channel/vixen/"}},
		{"performer", map[string]any{"id": "performer:riley", "name": "Riley"}},
		{"scene", map[string]any{"id": "scene:credited", "source_url": "https://x.test/credited", "studio": "studio:vixen", "performers": []string{"performer:riley"}}},
		{"scene", map[string]any{"id": "scene:foreign", "source_url": "https://x.test/foreign", "performers": []string{"performer:other"}}},
		{"scene", map[string]any{"id": "scene:stub", "source_url": "https://x.test/stub"}},
		{"scene", map[string]any{"id": "scene:page", "source_url": "https://x.test/page"}},
		{"observation", map[string]any{"id": "observation:credited", "target": "scene:credited", "plugin": "xhamster"}},
	}
	for _, row := range append(rows, extra...) {
		if _, err := database.Insert(ctx, row.table, row.fields); err != nil {
			t.Fatalf("insert %v: %v", row.fields["id"], err)
		}
	}
	repo := media.NewRepository(database)
	return NewScheduler(repo, fake), repo
}

// feedIDs lists a subscription's feed, sorted.
func feedIDs(t *testing.T, repo *media.Repository, subscriptionID string) []string {
	t.Helper()
	scenes, err := repo.SearchSubscriptionScenes(context.Background(), subscriptionID, 100, 0, nil)
	if err != nil {
		t.Fatalf("feed: %v", err)
	}
	ids := make([]string, 0, len(scenes))
	for _, scene := range scenes {
		ids = append(ids, scene.ID)
	}
	sort.Strings(ids)
	return ids
}

// assertStrings fails unless got equals want.
func assertStrings(t *testing.T, label string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %v, want %v", label, got, want)
	}
	for index := range got {
		if got[index] != want[index] {
			t.Fatalf("%s: got %v, want %v", label, got, want)
		}
	}
}

func TestRunStudioListsItsChannelPage(t *testing.T) {
	ctx := context.Background()
	fake := &fakeDiscovery{plugins: testPlugins(), pageHits: map[string][]string{"eporner": {"scene:page"}}}
	scheduler, repo := newTestScheduler(t, fake)
	subscription, _, err := repo.SubscribeEntity(ctx, model.SubscriptionKindStudio, "studio:vixen", 0)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	if err := scheduler.Run(ctx, subscription); err != nil {
		t.Fatalf("run: %v", err)
	}
	assertStrings(t, "pages", fake.pages, []string{"eporner@https://www.eporner.com/channel/vixen/"})
	assertStrings(t, "searches", fake.searches, nil)
	assertStrings(t, "feed", feedIDs(t, repo, subscription.ID), []string{"scene:credited", "scene:page"})
}

func TestRunPerformerFallsBackToKeywordSearch(t *testing.T) {
	ctx := context.Background()
	fake := &fakeDiscovery{
		plugins:    testPlugins(),
		searchHits: map[string][]string{"xhamster": {"scene:foreign", "scene:stub"}, "eporner": {"scene:page"}},
	}
	scheduler, repo := newTestScheduler(t, fake)
	subscription, _, err := repo.SubscribeEntity(ctx, model.SubscriptionKindPerformer, "performer:riley", 0)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	if err := scheduler.Run(ctx, subscription); err != nil {
		t.Fatalf("run: %v", err)
	}
	// No source URL: no page to list. The name is searched only where the
	// performer's scenes came from; the hit credited to someone else is dropped.
	assertStrings(t, "pages", fake.pages, nil)
	assertStrings(t, "searches", fake.searches, []string{"Riley@xhamster"})
	assertStrings(t, "feed", feedIDs(t, repo, subscription.ID), []string{"scene:credited", "scene:stub"})
}

func TestRunFailedPageListingFallsBackToSearch(t *testing.T) {
	ctx := context.Background()
	fake := &fakeDiscovery{
		plugins:    testPlugins(),
		pageErrors: map[string]error{"eporner": errors.New("layout changed")},
		searchHits: map[string][]string{"eporner": {"scene:stub"}},
	}
	scheduler, repo := newTestScheduler(t, fake)
	subscription, _, err := repo.SubscribeEntity(ctx, model.SubscriptionKindStudio, "studio:vixen", 0)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	if err := scheduler.Run(ctx, subscription); err != nil {
		t.Fatalf("run: %v", err)
	}
	// Origin plugins: eporner serves the channel URL, xhamster's scenes credit it.
	assertStrings(t, "searches", fake.searches, []string{"Vixen@eporner", "Vixen@xhamster"})
	assertStrings(t, "feed", feedIDs(t, repo, subscription.ID), []string{"scene:credited", "scene:stub"})
	subscription, _ = repo.SearchSubscription(ctx, subscription.ID)
	if subscription.LastError == nil {
		t.Fatal("the failed page listing is kept as last_error")
	}
}

func TestRunSearchUsesItsSources(t *testing.T) {
	ctx := context.Background()
	fake := &fakeDiscovery{plugins: testPlugins(), searchHits: map[string][]string{"xhamster": {"scene:stub"}}}
	scheduler, repo := newTestScheduler(t, fake)
	subscription, _, err := repo.SubscribeSearch(ctx, "big query", []string{"xhamster"}, 0)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	if err := scheduler.Run(ctx, subscription); err != nil {
		t.Fatalf("run: %v", err)
	}
	assertStrings(t, "searches", fake.searches, []string{"big query@xhamster"})
	assertStrings(t, "pages", fake.pages, nil)
	assertStrings(t, "feed", feedIDs(t, repo, subscription.ID), []string{"scene:stub"})
}
