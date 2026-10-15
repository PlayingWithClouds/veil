package media

import (
	"context"
	"testing"
)

func TestDownloadedScenesFollowsPayloadMediaLink(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	insertRow(t, database, "scene", "scene:kept", map[string]any{"source_url": "https://example.test/kept", "title": "kept"})
	insertRow(t, database, "scene", "scene:queued", map[string]any{"source_url": "https://example.test/queued", "title": "queued"})
	insertRow(t, database, "job", "job:done", map[string]any{
		"kind": "download", "status": "completed", "plugin_name": "eporner",
		"payload": map[string]any{"media": "scene:kept", "download_url": "http://localhost/api/blob/x.mp4"},
	})
	insertRow(t, database, "job", "job:running", map[string]any{
		"kind": "download", "status": "running", "plugin_name": "eporner",
		"payload": map[string]any{"media": "scene:queued"},
	})

	scenes, err := repo.DownloadedScenes(ctx)
	if err != nil {
		t.Fatalf("downloaded scenes: %v", err)
	}
	if len(scenes) != 1 || scenes[0].ID != "scene:kept" {
		t.Fatalf("only the completed download's scene belongs in the library: %+v", scenes)
	}
}

// TestDownloadedScenesIncludesRecordings lists cam recordings (stored without a
// download job) with the downloads, newest first.
func TestDownloadedScenesIncludesRecordings(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	insertRow(t, database, "scene", "scene:downloaded", map[string]any{"source_url": "https://example.test/kept", "title": "downloaded"})
	insertRow(t, database, "job", "job:done", map[string]any{
		"kind": "download", "status": "completed", "plugin_name": "eporner", "finished_at": "2026-10-15T10:00:00.000Z",
		"payload": map[string]any{"media": "scene:downloaded"},
	})
	insertRow(t, database, "scene", "scene:recording", map[string]any{
		"source_url": "veil-recording:abc", "title": "model — live", "created_at": "2026-10-15T12:00:00.000Z",
	})
	insertRow(t, database, "scene", "scene:stub", map[string]any{"source_url": "https://example.test/stub", "title": "stub"})

	scenes, err := repo.DownloadedScenes(ctx)
	if err != nil {
		t.Fatalf("downloaded scenes: %v", err)
	}
	if len(scenes) != 2 || scenes[0].ID != "scene:recording" || scenes[1].ID != "scene:downloaded" {
		t.Fatalf("want the recording, then the download: %+v", scenes)
	}
}
