package follows

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/notify"
	"github.com/playingwithclouds/veil/internal/pipeline"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// fakeDiscoverer lists fixed scenes into the library as a plugin would.
type fakeDiscoverer struct {
	library *library.Library
	// pages maps a page URL to the external ids of the scenes on it.
	pages map[string][]string
	// searches maps a search text to the external ids it finds.
	searches map[string][]string
	listed   []string
}

// ListPage stores the page's scenes as stubs of plugin "site"; other
// plugins can't list pages.
func (f *fakeDiscoverer) ListPage(ctx context.Context, pluginName, url string, limit int) ([]schema.RecordRef, error) {
	if pluginName != "site" {
		return nil, fmt.Errorf("%s: %w", pluginName, pipeline.ErrNoPageListing)
	}
	f.listed = append(f.listed, url)
	return f.store(ctx, f.pages[url])
}

// Search stores the search's scenes as stubs.
func (f *fakeDiscoverer) Search(ctx context.Context, entity plugins.MediaType, query string, limit int) []schema.RecordRef {
	refs, _ := f.store(ctx, f.searches[query])
	return refs
}

// store ingests listing cards for externalIDs.
func (f *fakeDiscoverer) store(ctx context.Context, externalIDs []string) ([]schema.RecordRef, error) {
	refs := []schema.RecordRef{}
	for _, externalID := range externalIDs {
		item := plugins.DiscoveredItem{MediaType: plugins.MediaTypeScene, ExternalID: externalID, SourceURL: "https://site.test/" + externalID, Title: "Scene " + externalID}
		ref, err := f.library.IngestListed(ctx, "site", item)
		if err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	// Let the clock pass the stubs' millisecond, as a real listing would.
	time.Sleep(2 * time.Millisecond)
	return refs, nil
}

// fixture is a follow service with its collaborators.
type fixture struct {
	service    *Service
	library    *library.Library
	discoverer *fakeDiscoverer
	notifier   *notify.Service
	queue      *jobs.Queue
	hooked     [][]schema.RecordRef
}

// newFixture opens everything on a fresh database.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	lib := library.New(database)
	f := &fixture{library: lib, notifier: notify.New(database), queue: jobs.NewQueue(database, 3)}
	f.discoverer = &fakeDiscoverer{library: lib, pages: map[string][]string{}, searches: map[string][]string{}}
	f.service = New(Options{Library: lib, Queue: f.queue, Discoverer: f.discoverer, Notifier: f.notifier,
		OnNewMedia: func(ctx context.Context, follow schema.Follow, media []schema.RecordRef) {
			f.hooked = append(f.hooked, media)
		}})
	return f
}

// addPerformer stores a performer listed on plugin at url.
func (f *fixture) addPerformer(t *testing.T, plugin, url string) schema.RecordRef {
	t.Helper()
	ref, err := f.library.IngestListed(context.Background(), plugin, plugins.DiscoveredItem{MediaType: plugins.MediaTypePerformer, ExternalID: url, SourceURL: url, Title: "Ada Lane"})
	if err != nil {
		t.Fatal(err)
	}
	return ref
}

// addCreditedScene stores a scene page crediting Ada Lane.
func (f *fixture) addCreditedScene(t *testing.T, externalID string) {
	t.Helper()
	scene := &plugins.Scene{ExternalID: externalID, SourceURL: "https://other.test/" + externalID, Title: "Credited " + externalID, Performers: []plugins.ScenePerformer{{Name: "Ada Lane"}}}
	if _, err := f.library.IngestFound(context.Background(), "other", []*plugins.ScrapeResult{{Type: plugins.MediaTypeScene, Scene: scene}}); err != nil {
		t.Fatal(err)
	}
}

// runFollow runs a follow and expects count new media.
func (f *fixture) runFollow(t *testing.T, id string, count int) {
	t.Helper()
	media, err := f.service.RunFollow(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if len(media) != count {
		t.Errorf("run found %d new media, want %d: %+v", len(media), count, media)
	}
}

// TestPerformerFollow lists the performer's page, reports what it brought in
// once, and later picks up credited media from elsewhere.
func TestPerformerFollow(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	performer := f.addPerformer(t, "site", "https://site.test/ada")
	f.addPerformer(t, "noPages", "https://nopages.test/ada")
	f.addCreditedScene(t, "old")
	f.discoverer.pages["https://site.test/ada"] = []string{"s1", "s2"}
	time.Sleep(2 * time.Millisecond)

	follow, err := f.service.FollowRecord(ctx, performer, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := f.service.FollowRecord(ctx, performer, 0, true)
	if again.Id != follow.Id || follow.Name != "Ada Lane" || follow.IntervalHours != defaultIntervalHours {
		t.Errorf("follow = %+v, again = %+v", follow, again)
	}

	f.runFollow(t, follow.Id, 2)
	f.runFollow(t, follow.Id, 0)
	f.addCreditedScene(t, "new")
	f.runFollow(t, follow.Id, 1)

	if len(f.discoverer.listed) != 3 {
		t.Errorf("listed pages %v, want the site page each run", f.discoverer.listed)
	}
	if len(f.hooked) != 2 {
		t.Errorf("hook fired %d times, want 2", len(f.hooked))
	}
	notifications, _ := f.notifier.Notifications(ctx, false, 10, 0)
	if len(notifications) != 2 || notifications[1].Title != "2 new for Ada Lane" {
		t.Errorf("notifications = %+v", notifications)
	}
	stored, _ := f.service.Get(ctx, follow.Id)
	if stored.LastRunAt == nil || stored.NextRunAt == nil || *stored.NextRunAt <= *stored.LastRunAt || stored.LastError != nil {
		t.Errorf("run not recorded: %+v", stored)
	}
}

// TestSearchFollow runs a search follow through the plugins' search.
func TestSearchFollow(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.discoverer.searches["Scene"] = []string{"a", "b", "c"}
	follow, err := f.service.FollowSearch(ctx, "Scene", 6, false)
	if err != nil {
		t.Fatal(err)
	}
	f.runFollow(t, follow.Id, 3)
	notifications, _ := f.notifier.Notifications(ctx, false, 10, 0)
	if len(notifications) != 0 {
		t.Errorf("follow without notify notified: %+v", notifications)
	}
	media, err := f.service.NewSince(ctx, follow.Id, "", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(media) != 3 {
		t.Errorf("NewSince = %+v, want the 3 found", media)
	}
}

// TestEnqueueDue queues each due follow once and skips disabled ones.
func TestEnqueueDue(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	first, _ := f.service.FollowSearch(ctx, "one", 0, false)
	second, _ := f.service.FollowSearch(ctx, "two", 0, false)
	disabled := false
	if _, err := f.service.Update(ctx, second.Id, FollowEdit{Enabled: &disabled}); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		queued, err := f.service.EnqueueDue(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if queued != 1 {
			t.Errorf("queued %d follows, want 1", queued)
		}
	}
	jobList, _ := f.queue.List(ctx, jobs.Query{Limit: 10})
	if len(jobList) != 1 || jobList[0].Payload[payloadFollowID] != first.Id {
		t.Errorf("jobs = %+v, want one run of the first follow", jobList)
	}
	if err := f.service.RunTask()(ctx, &jobList[0], func(float64) {}); err != nil {
		t.Fatal(err)
	}
	queued, _ := f.service.EnqueueDue(ctx)
	if queued != 0 {
		t.Errorf("follow due again right after its run")
	}
}
