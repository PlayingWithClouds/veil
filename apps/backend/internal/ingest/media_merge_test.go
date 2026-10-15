package ingest

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// insertScene inserts a scene row with a runtime, the way a listing stub has one.
func insertScene(t *testing.T, ctx context.Context, database *db.DB, sourceURL, title string, seconds int) string {
	t.Helper()
	id, err := database.Insert(ctx, "scene", map[string]any{"source_url": sourceURL, "title": title, "duration_seconds": seconds})
	if err != nil {
		t.Fatalf("insert scene: %v", err)
	}
	return id.String()
}

// TestMergeDuplicateScenesKeepsEverything checks that two sites' copies of one
// video become one scene that holds the streams, history, rating and the other
// site's address, and that the survivor is the copy the user used.
func TestMergeDuplicateScenesKeepsEverything(t *testing.T) {
	service, database, ctx := openTestService(t)
	stub := insertScene(t, ctx, database, "https://a.test/video/1", "Tanya And Maya Share A Pool Party", 1200)
	used := insertScene(t, ctx, database, "https://b.test/watch/9", "Tanya and Maya share a pool party", 1195)
	other := insertScene(t, ctx, database, "https://c.test/watch/3", "A completely different video", 1200)
	for _, query := range []string{
		`INSERT INTO watch_history (id, media, progress_seconds) VALUES ('watch_history:one', $used, 300)`,
		`INSERT INTO user_rating (id, media, rating) VALUES ('user_rating:one', $used, 8)`,
		`INSERT INTO stream (id, media, url) VALUES ('stream:one', $used, 'https://cdn.test/used.mp4')`,
		`INSERT INTO stream (id, media, url) VALUES ('stream:two', $stub, 'https://cdn.test/stub.mp4')`,
	} {
		if _, err := database.Exec(ctx, query, db.Vars{"used": used, "stub": stub}); err != nil {
			t.Fatalf("seed %q: %v", query, err)
		}
	}

	merged, err := service.MergeDuplicateMedia(ctx)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if merged != 1 {
		t.Fatalf("merged = %d, want 1", merged)
	}

	if scenes := countRows(t, ctx, database, `SELECT count(*) FROM scene`, nil); scenes != 2 {
		t.Fatalf("scenes = %d, want 2", scenes)
	}
	if gone := countRows(t, ctx, database, `SELECT count(*) FROM scene WHERE id = $id`, db.Vars{"id": stub}); gone != 0 {
		t.Fatalf("the unused copy is still there")
	}
	if streams := countRows(t, ctx, database, `SELECT count(*) FROM stream WHERE media = $id`, db.Vars{"id": used}); streams != 2 {
		t.Fatalf("survivor streams = %d, want 2", streams)
	}
	if history := countRows(t, ctx, database, `SELECT count(*) FROM watch_history WHERE media = $id`, db.Vars{"id": used}); history != 1 {
		t.Fatalf("survivor history = %d, want 1", history)
	}
	if otherKept := countRows(t, ctx, database, `SELECT count(*) FROM scene WHERE id = $id`, db.Vars{"id": other}); otherKept != 1 {
		t.Fatalf("a different video was merged")
	}

	listed := ingestScene(t, ctx, service, "a", &plugins.Scene{Type: plugins.MediaTypeScene, SourceURL: "https://a.test/video/1", Title: "Tanya And Maya Share A Pool Party"})
	if listed.IsNew || listed.CanonicalID != used {
		t.Fatalf("merged copy's page resolved to %s (new=%v), want %s", listed.CanonicalID, listed.IsNew, used)
	}

	again, err := service.MergeDuplicateMedia(ctx)
	if err != nil || again != 0 {
		t.Fatalf("second pass merged %d (err %v), want 0", again, err)
	}
}

// TestNewSceneJoinsItsTwinOnIngest checks that a listing of a video the library
// already has from another site adds no scene: the result is the existing one.
func TestNewSceneJoinsItsTwinOnIngest(t *testing.T) {
	service, database, ctx := openTestService(t)
	first := ingestScene(t, ctx, service, "alpha", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: "https://a.test/video/1", Title: "Sunset Beach Fun With Friends", Duration: 900,
	})
	second := ingestScene(t, ctx, service, "beta", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: "https://b.test/watch/2", Title: "Sunset beach fun with friends", Duration: 902,
		Downloads: []plugins.Download{{URL: "https://cdn.test/b.mp4"}},
	})
	if second.IsNew || second.CanonicalID != first.CanonicalID {
		t.Fatalf("second copy = %+v, want the first scene %s", second, first.CanonicalID)
	}
	if scenes := countRows(t, ctx, database, `SELECT count(*) FROM scene`, nil); scenes != 1 {
		t.Fatalf("scenes = %d, want 1", scenes)
	}
	if streams := countRows(t, ctx, database, `SELECT count(*) FROM stream WHERE media = $id`, db.Vars{"id": first.CanonicalID}); streams != 1 {
		t.Fatalf("streams on the scene = %d, want 1", streams)
	}

	different := ingestScene(t, ctx, service, "beta", &plugins.Scene{
		Type: plugins.MediaTypeScene, SourceURL: "https://b.test/watch/3", Title: "Mountain hike in the rain", Duration: 900,
	})
	if !different.IsNew {
		t.Fatalf("an unrelated video joined a scene")
	}
}

// TestMergeDuplicateGalleries checks that the same image set from two sites is
// merged only when the counts agree, and keeps the images once.
func TestMergeDuplicateGalleries(t *testing.T) {
	service, database, ctx := openTestService(t)
	kept, err := database.Insert(ctx, "gallery", map[string]any{"source_url": "https://a.test/g/1", "title": "Katie Morgan poolside bikini photoshoot", "image_count": 40})
	if err != nil {
		t.Fatalf("insert gallery: %v", err)
	}
	copyID, err := database.Insert(ctx, "gallery", map[string]any{"source_url": "https://b.test/g/7", "title": "Katie Morgan Poolside Bikini Photoshoot", "image_count": 40})
	if err != nil {
		t.Fatalf("insert gallery: %v", err)
	}
	other, err := database.Insert(ctx, "gallery", map[string]any{"source_url": "https://b.test/g/8", "title": "Katie Morgan poolside bikini photoshoot", "image_count": 25})
	if err != nil {
		t.Fatalf("insert gallery: %v", err)
	}
	for _, image := range []struct{ gallery, path string }{{kept.String(), "a1"}, {copyID.String(), "b1"}} {
		if _, err := database.Insert(ctx, "image", map[string]any{"media": image.gallery, "file_path": image.path, "content": true}); err != nil {
			t.Fatalf("insert image: %v", err)
		}
	}

	merged, err := service.MergeDuplicateMedia(ctx)
	if err != nil || merged != 1 {
		t.Fatalf("merged %d (err %v), want 1", merged, err)
	}
	if galleries := countRows(t, ctx, database, `SELECT count(*) FROM gallery`, nil); galleries != 2 {
		t.Fatalf("galleries = %d, want 2", galleries)
	}
	if left := countRows(t, ctx, database, `SELECT count(*) FROM gallery WHERE id = $other`, db.Vars{"other": other.String()}); left != 1 {
		t.Fatalf("the different-size gallery went")
	}
	if images := countRows(t, ctx, database, `SELECT count(*) FROM image WHERE content`, nil); images != 1 {
		t.Fatalf("images = %d, want the set once", images)
	}
}

// TestMergeDuplicateCollections checks that same-named scraper collections that
// share a member merge, and that user playlists are left alone.
func TestMergeDuplicateCollections(t *testing.T) {
	service, database, ctx := openTestService(t)
	scene := insertScene(t, ctx, database, "https://a.test/video/1", "Member", 600)
	extra := insertScene(t, ctx, database, "https://a.test/video/2", "Another", 700)
	big, err := database.Insert(ctx, "collection", map[string]any{"name": "Pool Party Series", "source_url": "https://a.test/series/1"})
	if err != nil {
		t.Fatalf("insert collection: %v", err)
	}
	small, err := database.Insert(ctx, "collection", map[string]any{"name": "pool-party series", "source_url": "https://b.test/series/1"})
	if err != nil {
		t.Fatalf("insert collection: %v", err)
	}
	playlist, err := database.Insert(ctx, "collection", map[string]any{"name": "Pool Party Series", "user_created": true})
	if err != nil {
		t.Fatalf("insert collection: %v", err)
	}
	for _, member := range []struct{ collection, media string }{
		{big.String(), scene}, {big.String(), extra}, {small.String(), scene}, {playlist.String(), scene},
	} {
		if _, err := database.Insert(ctx, "collection_item", map[string]any{"collection": member.collection, "media": member.media}); err != nil {
			t.Fatalf("insert member: %v", err)
		}
	}

	merged, err := service.MergeDuplicateMedia(ctx)
	if err != nil || merged != 1 {
		t.Fatalf("merged %d (err %v), want 1", merged, err)
	}
	if collections := countRows(t, ctx, database, `SELECT count(*) FROM collection WHERE id != 'collection:favorites'`, nil); collections != 2 {
		t.Fatalf("collections = %d, want 2", collections)
	}
	if members := countRows(t, ctx, database, `SELECT count(*) FROM collection_item WHERE collection = $id`, db.Vars{"id": big.String()}); members != 2 {
		t.Fatalf("survivor members = %d, want 2", members)
	}
}
