package media

import (
	"context"
	"path/filepath"
	"sort"
	"testing"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

func newTestRepository(t *testing.T) (*Repository, *db.DB) {
	t.Helper()
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return NewRepository(database), database
}

// insertRow inserts fields with an explicit id and fails the test on error.
func insertRow(t *testing.T, database *db.DB, table, id string, fields map[string]any) {
	t.Helper()
	fields["id"] = id
	if _, err := database.Insert(context.Background(), table, fields); err != nil {
		t.Fatalf("insert %s: %v", id, err)
	}
}

func stringPointer(value string) *string {
	return &value
}

// tagMatchesByID maps each listed id to its tagMatch value.
func tagMatchesByID[T any](items []T, idOf func(T) string, matchOf func(T) *string) map[string]string {
	out := map[string]string{}
	for _, item := range items {
		match := "<nil>"
		if matchOf(item) != nil {
			match = *matchOf(item)
		}
		out[idOf(item)] = match
	}
	return out
}

func assertMatches(t *testing.T, got, want map[string]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for id, match := range want {
		if got[id] != match {
			t.Fatalf("%s: got %q, want %q (all: %v)", id, got[id], match, got)
		}
	}
}

// seedTaggedEntities creates tag:target, a studio and a performer carrying it,
// and an unrelated studio/performer.
func seedTaggedEntities(t *testing.T, database *db.DB) {
	insertRow(t, database, "tag", "tag:target", map[string]any{"name": "target"})
	insertRow(t, database, "tag", "tag:other", map[string]any{"name": "other"})
	insertRow(t, database, "studio", "studio:tagged", map[string]any{"name": "Tagged", "tags": []string{"tag:target"}})
	insertRow(t, database, "studio", "studio:plain", map[string]any{"name": "Plain", "tags": []string{"tag:other"}})
	insertRow(t, database, "performer", "performer:tagged", map[string]any{"name": "Tagged", "tags": []string{"tag:other", "tag:target"}})
	insertRow(t, database, "performer", "performer:plain", map[string]any{"name": "Plain"})
}

func TestListScenesInheritedTagMatch(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	seedTaggedEntities(t, database)
	insertRow(t, database, "scene", "scene:direct", map[string]any{"source_url": "a", "tags": []string{"tag:target"}, "studio": "studio:plain", "created_at": "2020-01-01T00:00:00.000Z"})
	insertRow(t, database, "scene", "scene:studio", map[string]any{"source_url": "b", "studio": "studio:tagged", "created_at": "2020-01-03T00:00:00.000Z"})
	insertRow(t, database, "scene", "scene:performer", map[string]any{"source_url": "c", "performers": []string{"performer:plain", "performer:tagged"}, "created_at": "2020-01-02T00:00:00.000Z"})
	insertRow(t, database, "scene", "scene:none", map[string]any{"source_url": "d", "tags": []string{"tag:other"}, "studio": "studio:plain", "performers": []string{"performer:plain"}})

	scenes, err := repo.ListScenes(ctx, EntityFilters{TagID: stringPointer("tag:target")}, nil, nil)
	if err != nil {
		t.Fatalf("list scenes: %v", err)
	}
	matches := tagMatchesByID(scenes, func(s *model.Scene) string { return s.ID }, func(s *model.Scene) *string { return s.TagMatch })
	assertMatches(t, matches, map[string]string{
		"scene:direct":    "direct",
		"scene:studio":    "inherited",
		"scene:performer": "inherited",
	})
	if scenes[0].ID != "scene:direct" {
		t.Fatalf("direct match must rank first, got %s", scenes[0].ID)
	}
	if scenes[1].ID != "scene:studio" {
		t.Fatalf("inherited matches keep created_at DESC order, got %s", scenes[1].ID)
	}

	unfiltered, err := repo.ListScenes(ctx, EntityFilters{}, nil, nil)
	if err != nil {
		t.Fatalf("list scenes: %v", err)
	}
	if len(unfiltered) != 4 || unfiltered[0].TagMatch != nil {
		t.Fatalf("unfiltered listing: %d scenes, tagMatch %v", len(unfiltered), unfiltered[0].TagMatch)
	}
}

func TestListScenesRelationsAndCounts(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	seedTaggedEntities(t, database)
	insertRow(t, database, "scene", "scene:one", map[string]any{
		"source_url": "a", "title": "Hello World", "studio": "studio:tagged",
		"performers": []string{"performer:tagged", "performer:plain"},
		"tags":       []string{"tag:other", "tag:target"},
	})
	insertRow(t, database, "scene", "scene:two", map[string]any{"source_url": "b", "title": "100% real", "performers": []string{"performer:plain"}})

	scene, err := repo.GetScene(ctx, "scene:one")
	if err != nil || scene == nil {
		t.Fatalf("get scene: %v", err)
	}
	if scene.Studio == nil || scene.Studio.ID != "studio:tagged" {
		t.Fatalf("studio not resolved: %+v", scene.Studio)
	}
	if len(scene.Performers) != 2 || scene.Performers[0].ID != "performer:tagged" || scene.Performers[1].ID != "performer:plain" {
		t.Fatalf("performers not resolved in credit order: %+v", scene.Performers)
	}
	if len(scene.Tags) != 2 || scene.Tags[0].ID != "tag:other" {
		t.Fatalf("tags not resolved in order: %+v", scene.Tags)
	}

	performer, err := repo.GetPerformer(ctx, "performer:plain")
	if err != nil || performer.SceneCount != 2 {
		t.Fatalf("performer scene count: %v %+v", err, performer)
	}
	tag, err := repo.GetTag(ctx, "tag:target")
	if err != nil || tag.SceneCount != 1 {
		t.Fatalf("tag scene count: %v %+v", err, tag)
	}

	searched, err := repo.ListScenes(ctx, EntityFilters{Search: stringPointer("%")}, nil, nil)
	if err != nil || len(searched) != 1 || searched[0].ID != "scene:two" {
		t.Fatalf("search must treat %% literally: %v %+v", err, searched)
	}
	byPerformer, err := repo.ListScenes(ctx, EntityFilters{PerformerID: stringPointer("performer:tagged")}, nil, nil)
	if err != nil || len(byPerformer) != 1 || byPerformer[0].ID != "scene:one" {
		t.Fatalf("performer filter: %v %+v", err, byPerformer)
	}

	studios, err := repo.ListStudios(ctx, EntityFilters{TagID: stringPointer("tag:target")}, nil, nil)
	if err != nil || len(studios) != 1 || studios[0].ID != "studio:tagged" || studios[0].SceneCount != 1 {
		t.Fatalf("studio tag filter: %v %+v", err, studios)
	}
	if _, err := repo.SetPerformerFavorite(ctx, "performer:plain", true); err != nil {
		t.Fatalf("set favorite: %v", err)
	}
	performer, _ = repo.GetPerformer(ctx, "performer:plain")
	if !performer.Favorite {
		t.Fatalf("favorite not stored")
	}
	if _, err := repo.SetStudioTags(ctx, "studio:plain", []string{"tag:target"}); err != nil {
		t.Fatalf("set studio tags: %v", err)
	}
	studio, _ := repo.GetStudio(ctx, "studio:plain")
	if len(studio.Tags) != 1 || studio.Tags[0].ID != "tag:target" {
		t.Fatalf("studio tags not replaced: %+v", studio.Tags)
	}
}

func TestListGalleriesInheritedTagMatch(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	seedTaggedEntities(t, database)
	insertRow(t, database, "gallery", "gallery:direct", map[string]any{"source_url": "a", "tags": []string{"tag:target"}})
	insertRow(t, database, "gallery", "gallery:studio", map[string]any{"source_url": "b", "studio": "studio:tagged"})
	insertRow(t, database, "gallery", "gallery:performer", map[string]any{"source_url": "c", "performers": []string{"performer:tagged"}})
	insertRow(t, database, "gallery", "gallery:none", map[string]any{"source_url": "d", "studio": "studio:plain"})

	galleries, err := repo.ListGalleries(ctx, EntityFilters{TagID: stringPointer("tag:target")}, nil, nil)
	if err != nil {
		t.Fatalf("list galleries: %v", err)
	}
	matches := tagMatchesByID(galleries, func(g *model.Gallery) string { return g.ID }, func(g *model.Gallery) *string { return g.TagMatch })
	assertMatches(t, matches, map[string]string{
		"gallery:direct":    "direct",
		"gallery:studio":    "inherited",
		"gallery:performer": "inherited",
	})
	if galleries[0].ID != "gallery:direct" {
		t.Fatalf("direct match must rank first, got %s", galleries[0].ID)
	}
}

func TestListImagesInheritsFromOwningGallery(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	seedTaggedEntities(t, database)
	insertRow(t, database, "gallery", "gallery:tagged", map[string]any{"source_url": "g1", "tags": []string{"tag:target"}, "studio": "studio:plain", "performers": []string{"performer:plain"}})
	insertRow(t, database, "gallery", "gallery:plain", map[string]any{"source_url": "g2"})
	insertRow(t, database, "image", "image:direct", map[string]any{"file_path": "1", "content": true, "tags": []string{"tag:target"}, "media": "gallery:plain"})
	insertRow(t, database, "image", "image:studio", map[string]any{"file_path": "2", "content": true, "studio": "studio:tagged"})
	insertRow(t, database, "image", "image:performer", map[string]any{"file_path": "3", "content": true, "performers": []string{"performer:tagged"}})
	insertRow(t, database, "image", "image:gallery", map[string]any{"file_path": "4", "content": true, "media": "gallery:tagged", "position": 1})
	insertRow(t, database, "image", "image:none", map[string]any{"file_path": "5", "content": true, "media": "gallery:plain"})
	insertRow(t, database, "image", "image:asset", map[string]any{"file_path": "6", "content": false, "tags": []string{"tag:target"}})

	images, err := repo.ListImages(ctx, EntityFilters{TagID: stringPointer("tag:target")}, nil, nil, nil)
	if err != nil {
		t.Fatalf("list images: %v", err)
	}
	matches := tagMatchesByID(images, func(i *model.Image) string { return i.ID }, func(i *model.Image) *string { return i.TagMatch })
	assertMatches(t, matches, map[string]string{
		"image:direct":    "direct",
		"image:studio":    "inherited",
		"image:performer": "inherited",
		"image:gallery":   "inherited",
	})

	byStudio, err := repo.ListImages(ctx, EntityFilters{StudioID: stringPointer("studio:plain")}, nil, nil, nil)
	if err != nil || len(byStudio) != 1 || byStudio[0].ID != "image:gallery" {
		t.Fatalf("studio filter via gallery: %v %+v", err, byStudio)
	}
	byPerformer, err := repo.ListImages(ctx, EntityFilters{PerformerID: stringPointer("performer:plain")}, nil, nil, nil)
	if err != nil || len(byPerformer) != 1 || byPerformer[0].ID != "image:gallery" {
		t.Fatalf("performer filter via gallery: %v %+v", err, byPerformer)
	}
	inGallery, err := repo.ListImages(ctx, EntityFilters{}, stringPointer("gallery:plain"), nil, nil)
	if err != nil || len(inGallery) != 2 {
		t.Fatalf("gallery filter: %v %+v", err, inGallery)
	}
	if inGallery[0].GalleryID == nil || *inGallery[0].GalleryID != "gallery:plain" {
		t.Fatalf("gallery id not set: %+v", inGallery[0])
	}

	gallery, err := repo.GetGallery(ctx, "gallery:tagged")
	if err != nil || len(gallery.Images) != 1 || gallery.Images[0].ID != "image:gallery" {
		t.Fatalf("gallery images: %v %+v", err, gallery)
	}
}

func TestBlocklistAndDisabledPluginExclusion(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	seedTaggedEntities(t, database)
	insertRow(t, database, "scene", "scene:blocked-tag", map[string]any{"source_url": "a", "tags": []string{"tag:other", "tag:target"}})
	insertRow(t, database, "scene", "scene:blocked-performer", map[string]any{"source_url": "b", "performers": []string{"performer:plain"}})
	insertRow(t, database, "scene", "scene:blocked-studio", map[string]any{"source_url": "c", "studio": "studio:plain"})
	insertRow(t, database, "scene", "scene:stub", map[string]any{"source_url": "d"})
	insertRow(t, database, "scene", "scene:scraped", map[string]any{"source_url": "e"})
	insertRow(t, database, "scene", "scene:kept", map[string]any{"source_url": "f", "tags": []string{"tag:other"}, "studio": "studio:tagged"})
	insertRow(t, database, "observation", "observation:stub", map[string]any{"target": "scene:stub", "plugin": "gone"})
	insertRow(t, database, "observation", "observation:scraped", map[string]any{"target": "scene:scraped", "plugin": "gone"})
	insertRow(t, database, "observation", "observation:kept", map[string]any{"target": "scene:kept", "plugin": "active"})
	insertRow(t, database, "stream", "stream:scraped", map[string]any{"media": "scene:scraped", "url": "http://x"})

	for _, block := range [][2]string{{"tag", "tag:target"}, {"performer", "performer:plain"}, {"studio", "studio:plain"}} {
		if _, err := repo.AddBlock(ctx, block[0], block[1], nil); err != nil {
			t.Fatalf("add block: %v", err)
		}
	}
	if _, err := repo.AddBlock(ctx, "tag", "tag:target", stringPointer("again")); err != nil {
		t.Fatalf("re-add block: %v", err)
	}
	entries, err := repo.ListBlocklist(ctx)
	if err != nil || len(entries) != 3 {
		t.Fatalf("blocklist must stay deduped per target: %v %+v", err, entries)
	}

	filters := EntityFilters{DisabledPlugins: []string{"gone"}}
	scenes, err := repo.ListScenes(ctx, filters, nil, nil)
	if err != nil {
		t.Fatalf("list scenes: %v", err)
	}
	if got := sceneIDs(scenes); !equalStrings(got, []string{"scene:kept", "scene:scraped"}) {
		t.Fatalf("list scenes exclusion: got %v", got)
	}
	random, err := repo.RandomScenes(ctx, 10, filters)
	if err != nil {
		t.Fatalf("random scenes: %v", err)
	}
	if got := sceneIDs(random); !equalStrings(got, []string{"scene:kept", "scene:scraped"}) {
		t.Fatalf("random scenes exclusion: got %v", got)
	}

	targets := repo.BlockedTargetIDSet(ctx)
	if len(targets) != 3 || !targets["studio:plain"] {
		t.Fatalf("blocked target set: %v", targets)
	}
	if _, err := repo.RemoveBlock(ctx, "tag:target"); err != nil {
		t.Fatalf("remove block: %v", err)
	}
	blocked := repo.BlockedSceneRIDs(ctx)
	if len(blocked) != 2 {
		t.Fatalf("blocked scenes after unblocking tag: %v", blocked)
	}
}

func sceneIDs(scenes []*model.Scene) []string {
	ids := make([]string, 0, len(scenes))
	for _, scene := range scenes {
		ids = append(ids, scene.ID)
	}
	sort.Strings(ids)
	return ids
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func TestCollections(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	insertRow(t, database, "scene", "scene:one", map[string]any{"source_url": "a", "title": "One", "poster_path": "one.jpg"})
	insertRow(t, database, "scene", "scene:two", map[string]any{"source_url": "b", "title": "Two"})
	insertRow(t, database, "collection", "collection:scraped", map[string]any{"name": "Scraped"})
	insertRow(t, database, "image", "image:asset", map[string]any{"file_path": "x", "content": false})

	created, err := repo.CreateCollection(ctx, "Mine")
	if err != nil || created.Origin != "user" {
		t.Fatalf("create collection: %v %+v", err, created)
	}
	for _, target := range []string{created.ID, "collection:scraped"} {
		if _, err := repo.AddToCollection(ctx, target, "scene:one"); err != nil {
			t.Fatalf("add to collection: %v", err)
		}
	}
	if _, err := repo.AddToCollection(ctx, created.ID, "scene:one"); err != nil {
		t.Fatalf("re-add must be idempotent: %v", err)
	}
	if _, err := repo.AddToCollection(ctx, created.ID, "scene:two"); err != nil {
		t.Fatalf("add second: %v", err)
	}
	if _, err := repo.AddToCollection(ctx, created.ID, "image:asset"); err == nil {
		t.Fatalf("asset images must be rejected")
	}

	ids, err := repo.CollectionIdsForMedia(ctx, "scene:one")
	if err != nil || !equalStrings(ids, []string{created.ID}) {
		t.Fatalf("collection ids for media must only list user-created: %v %v", err, ids)
	}

	userOnly, err := repo.ListCollections(ctx, nil, nil, stringPointer("user"), nil, nil)
	if err != nil || len(userOnly) != 1 || userOnly[0].ID != created.ID || userOnly[0].ItemCount != 2 {
		t.Fatalf("user collections: %v %+v", err, userOnly)
	}
	if userOnly[0].CoverPath == nil || *userOnly[0].CoverPath != "one.jpg" {
		t.Fatalf("cover from first member: %+v", userOnly[0].CoverPath)
	}
	scraped, err := repo.ListCollections(ctx, nil, nil, stringPointer("scraped"), nil, nil)
	if err != nil || len(scraped) != 1 || scraped[0].Origin != "scraped" {
		t.Fatalf("scraped collections: %v %+v", err, scraped)
	}
	all, err := repo.ListCollections(ctx, stringPointer("min"), nil, nil, nil, nil)
	if err != nil || len(all) != 1 {
		t.Fatalf("search collections: %v %+v", err, all)
	}

	if _, err := repo.ReorderCollection(ctx, created.ID, []string{"scene:two", "scene:one"}); err != nil {
		t.Fatalf("reorder: %v", err)
	}
	members, err := repo.CollectionMembers(ctx, created.ID)
	if err != nil || len(members) != 2 || members[0].MediaID != "scene:two" {
		t.Fatalf("members after reorder: %v %+v", err, members)
	}

	renamed, err := repo.RenameCollection(ctx, created.ID, "Renamed")
	if err != nil || renamed.Name != "Renamed" {
		t.Fatalf("rename: %v %+v", err, renamed)
	}
	if _, err := repo.DeleteCollection(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	remainingItems, _ := database.Int(ctx, "SELECT count(*) FROM collection_item WHERE collection = $id", db.Vars{"id": created.ID})
	if remainingItems != 0 {
		t.Fatalf("items must be deleted with their collection")
	}
}

func TestSceneMarkers(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	insertRow(t, database, "tag", "tag:existing", map[string]any{"name": "Existing", "aliases": []string{"alias"}})
	insertRow(t, database, "scene_marker", "scene_marker:global", map[string]any{"media": "scene:one", "seconds": 5, "label": "global"})

	personal, err := repo.CreateSceneMarker(ctx, "scene:one", 2, nil, stringPointer("alias"), nil)
	if err != nil {
		t.Fatalf("create marker: %v", err)
	}
	if !personal.Personal || personal.Tag == nil || personal.Tag.ID != "tag:existing" {
		t.Fatalf("marker must be personal and resolve tag by alias: %+v", personal)
	}
	if _, err := repo.CreateSceneMarker(ctx, "scene:one", 3, nil, stringPointer("Brand New"), nil); err != nil {
		t.Fatalf("create marker with new tag: %v", err)
	}
	tagCount, _ := database.Int(ctx, "SELECT count(*) FROM tag WHERE name = 'Brand New'", nil)
	if tagCount != 1 {
		t.Fatalf("new tag must be created once")
	}
	if _, err := repo.CreateSceneMarker(ctx, "scene:one", 3, nil, nil, nil); err == nil {
		t.Fatalf("marker without tag or label must fail")
	}

	markers, err := repo.SceneMarkers(ctx, "scene:one")
	if err != nil || len(markers) != 3 {
		t.Fatalf("scene markers: %v %+v", err, markers)
	}
	if markers[0].ID != personal.ID || markers[2].ID != "scene_marker:global" || markers[2].Personal {
		t.Fatalf("markers must be ordered by seconds with personal flags: %+v", markers)
	}
}

func TestPersonalLibraryUpserts(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)

	if _, err := repo.UpsertUserRating(ctx, model.UpsertUserRatingInput{Media: "scene:one", Rating: 4}); err != nil {
		t.Fatalf("rate: %v", err)
	}
	rating, err := repo.UpsertUserRating(ctx, model.UpsertUserRatingInput{Media: "scene:one", Rating: 8})
	if err != nil || rating.Rating != 8 {
		t.Fatalf("re-rate: %v %+v", err, rating)
	}
	ratings, _ := repo.ListUserRatings(ctx, nil, nil)
	if len(ratings) != 1 {
		t.Fatalf("rating must upsert, got %d rows", len(ratings))
	}

	first, _ := repo.AddToWatchlist(ctx, "scene:one")
	second, err := repo.AddToWatchlist(ctx, "scene:one")
	if err != nil || first.ID != second.ID {
		t.Fatalf("watchlist add must be idempotent: %v %+v %+v", err, first, second)
	}

	if _, err := repo.UpsertWatchHistory(ctx, model.UpsertWatchHistoryInput{Media: "scene:one", ProgressSeconds: 10}); err != nil {
		t.Fatalf("watch: %v", err)
	}
	completed := true
	entry, err := repo.UpsertWatchHistory(ctx, model.UpsertWatchHistoryInput{Media: "scene:one", ProgressSeconds: 20, Completed: &completed})
	if err != nil || entry.ProgressSeconds != 20 || !entry.Completed || entry.FinishedAt == nil {
		t.Fatalf("watch update: %v %+v", err, entry)
	}
	history, _ := repo.ListWatchHistory(ctx, nil, nil)
	if len(history) != 1 {
		t.Fatalf("watch history must stay one row per media, got %d", len(history))
	}
	rewound, err := repo.UpsertWatchHistory(ctx, model.UpsertWatchHistoryInput{Media: "scene:one", ProgressSeconds: 5})
	if err != nil || rewound.ProgressSeconds != 5 || rewound.MaxProgressSeconds != 20 {
		t.Fatalf("seeking back must keep the furthest position: %v %+v", err, rewound)
	}

	for range 2 {
		if _, err := repo.IncrementOCount(ctx, "scene:one"); err != nil {
			t.Fatalf("increment: %v", err)
		}
	}
	count, err := repo.DecrementOCount(ctx, "scene:one")
	if err != nil || count != 1 {
		t.Fatalf("decrement: %v %d", err, count)
	}

	for range 2 {
		if err := repo.LogSearch(ctx, "  Foo ", []string{"plugin"}, 3); err != nil {
			t.Fatalf("log search: %v", err)
		}
	}
	uses, _ := database.Int(ctx, "SELECT uses FROM search_term WHERE normalized_query = 'foo'", nil)
	if uses != 2 {
		t.Fatalf("search term uses: got %d", uses)
	}
}

func TestListFiltersBySource(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	insertRow(t, database, "scene", "scene:eporner", map[string]any{"source_url": "https://example.test/a", "title": "a"})
	insertRow(t, database, "scene", "scene:missav", map[string]any{"source_url": "https://example.test/b", "title": "b"})
	insertRow(t, database, "gallery", "gallery:pornpics", map[string]any{"source_url": "https://example.test/c", "title": "c"})
	insertRow(t, database, "observation", "observation:a", map[string]any{"target": "scene:eporner", "plugin": "eporner"})
	insertRow(t, database, "observation", "observation:b", map[string]any{"target": "scene:missav", "plugin": "missav"})
	insertRow(t, database, "observation", "observation:c", map[string]any{"target": "gallery:pornpics", "plugin": "pornpics"})

	scenes, err := repo.ListScenes(ctx, EntityFilters{Sources: []string{"eporner"}}, nil, nil)
	if err != nil || len(scenes) != 1 || scenes[0].ID != "scene:eporner" {
		t.Fatalf("scenes by source: %v %v", err, scenes)
	}
	all, _ := repo.ListScenes(ctx, EntityFilters{}, nil, nil)
	if len(all) != 2 {
		t.Fatalf("no sources must list everything, got %d", len(all))
	}
	galleries, err := repo.ListGalleries(ctx, EntityFilters{Sources: []string{"eporner"}}, nil, nil)
	if err != nil || len(galleries) != 0 {
		t.Fatalf("galleries by source: %v %d", err, len(galleries))
	}
}
