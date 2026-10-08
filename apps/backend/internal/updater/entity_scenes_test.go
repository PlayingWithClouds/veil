package updater

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/media"
)

// memoryMarks is an in-memory FetchMarks.
type memoryMarks map[string]string

// Get returns the stored mark or cache.ErrMiss.
func (m memoryMarks) Get(ctx context.Context, key string) (string, error) {
	value, found := m[key]
	if !found {
		return "", cache.ErrMiss
	}
	return value, nil
}

// Set stores the mark, ignoring its lifetime.
func (m memoryMarks) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	m[key] = "set"
	return nil
}

// titledStubs are search stubs with titles, none credited yet.
var titledStubs = []seedRow{
	{"scene", map[string]any{"id": "scene:mention", "source_url": "https://x.test/mention", "title": "Riley's big day"}},
	{"scene", map[string]any{"id": "scene:unrelated", "source_url": "https://x.test/unrelated", "title": "Someone else entirely"}},
	{"scene", map[string]any{"id": "scene:alias", "source_url": "https://x.test/alias", "title": "Starring R. Quinn tonight"}},
}

// performerSceneIDs lists the scenes credited to a performer, sorted.
func performerSceneIDs(t *testing.T, repo *media.Repository, performerID string) []string {
	t.Helper()
	ids, err := repo.LinkedSceneIDs(context.Background(), performerID)
	if err != nil {
		t.Fatalf("linked scenes: %v", err)
	}
	sort.Strings(ids)
	return ids
}

func TestEnsurePerformerSearchesEveryScenePlugin(t *testing.T) {
	ctx := context.Background()
	fake := &fakeDiscovery{
		plugins: testPlugins(),
		searchHits: map[string][]string{
			"xhamster": {"scene:foreign", "scene:mention", "scene:unrelated"},
			"eporner":  {"scene:credited", "scene:alias"},
		},
	}
	aliased := seedRow{"performer", map[string]any{"id": "performer:quinn", "name": "Quinn", "aliases": []string{"R. Quinn"}}}
	scheduler, repo := newTestScheduler(t, fake, append(titledStubs, aliased)...)
	marks := memoryMarks{}
	fetcher := NewEntityScenes(scheduler, marks)

	linked, err := fetcher.Ensure(ctx, "performer:riley")
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	// Both plugins are searched, not just the origin one (xhamster).
	sort.Strings(fake.searches)
	assertStrings(t, "searches", fake.searches, []string{"Riley@eporner", "Riley@xhamster"})
	// Credited stays, a stub naming her is credited; someone else's scene and
	// an unrelated stub are not.
	assertStrings(t, "credited", performerSceneIDs(t, repo, "performer:riley"), []string{"scene:credited", "scene:mention"})
	if linked != 2 {
		t.Fatalf("linked: got %d, want 2", linked)
	}

	// Within the interval the page opens without searching again.
	fake.searches = nil
	if _, err := fetcher.Ensure(ctx, "performer:riley"); err != nil {
		t.Fatalf("second ensure: %v", err)
	}
	assertStrings(t, "searches after a recent fetch", fake.searches, nil)

	// Aliases count as mentions.
	if _, err := fetcher.Ensure(ctx, "performer:quinn"); err != nil {
		t.Fatalf("ensure alias: %v", err)
	}
	assertStrings(t, "alias credited", performerSceneIDs(t, repo, "performer:quinn"), []string{"scene:alias"})
}

func TestEnsureStudioCreditsItsPageListing(t *testing.T) {
	ctx := context.Background()
	fake := &fakeDiscovery{plugins: testPlugins(), pageHits: map[string][]string{"eporner": {"scene:page"}}}
	scheduler, repo := newTestScheduler(t, fake)
	fetcher := NewEntityScenes(scheduler, memoryMarks{})

	if _, err := fetcher.Ensure(ctx, "studio:vixen"); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	assertStrings(t, "pages", fake.pages, []string{"eporner@https://www.eporner.com/channel/vixen/"})
	ids, err := repo.LinkedSceneIDs(ctx, "studio:vixen")
	if err != nil {
		t.Fatalf("linked scenes: %v", err)
	}
	sort.Strings(ids)
	assertStrings(t, "studio scenes", ids, []string{"scene:credited", "scene:page"})
}

func TestEnsureRejectsTags(t *testing.T) {
	scheduler, _ := newTestScheduler(t, &fakeDiscovery{plugins: testPlugins()})
	if _, err := NewEntityScenes(scheduler, memoryMarks{}).Ensure(context.Background(), "tag:milf"); err == nil {
		t.Fatal("tags are not fetched")
	}
}

func TestMentionsAnyName(t *testing.T) {
	cases := []struct {
		title string
		names []string
		want  bool
	}{
		{"Riley-Reid's day off", []string{"Riley Reid"}, true},
		{"RILEY REID", []string{"riley reid"}, true},
		{"Rileys day", []string{"Riley"}, false},
		{"Abby and Al", []string{"Al"}, false},
		{"Nothing here", []string{"Riley", "Quinn"}, false},
	}
	for _, testCase := range cases {
		if got := mentionsAnyName(testCase.title, testCase.names); got != testCase.want {
			t.Errorf("mentionsAnyName(%q, %v) = %v, want %v", testCase.title, testCase.names, got, testCase.want)
		}
	}
}
