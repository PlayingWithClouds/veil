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
