package files

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestSnapshot archives a source's page and its video's cover into the blob
// store and points the source's snapshot at the page.
func TestSnapshot(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/cover.jpg" {
			writer.Header().Set("Content-Type", "image/jpeg")
			writer.Write([]byte("jpeg bytes"))
			return
		}
		writer.Header().Set("Content-Type", "text/html")
		writer.Write([]byte("<html>video page</html>"))
	}))
	defer server.Close()
	video := env.newVideo(t, "Pool day")
	if _, err := env.database.Exec(ctx, `UPDATE "Video" SET "cover" = $cover WHERE "id" = $id`,
		db.Vars{"cover": schema.Asset{Uri: server.URL + "/cover.jpg"}, "id": video.Id}); err != nil {
		t.Fatal(err)
	}
	sourceID, err := env.database.Insert(ctx, tableSource, db.Vars{
		"subjectKind": video.Kind, "subjectId": video.Id, "plugin": "site", "externalId": "v1", "url": server.URL + "/video/1",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := env.service.Snapshot(ctx, sourceID); err != nil {
		t.Fatal(err)
	}
	env.runJob(t, schema.JobKindSnapshot, env.service.SnapshotTask())
	source, _ := env.service.source(ctx, sourceID)
	pageKey := "archive/" + sourceID + "/page.html"
	if source.Snapshot == nil || source.Snapshot.Uri != env.store.DirectURL(pageKey) || *source.Snapshot.MimeType != "text/html" {
		t.Fatalf("snapshot = %+v", source.Snapshot)
	}
	if page := readBlob(t, env, pageKey); page != "<html>video page</html>" {
		t.Errorf("page = %q", page)
	}
	if cover := readBlob(t, env, "archive/"+sourceID+"/cover.jpg"); cover != "jpeg bytes" {
		t.Errorf("cover = %q", cover)
	}
	locations, _ := env.devices.Locations(ctx, "")
	if !hasKind(locations, schema.StorageKindArchive) {
		t.Errorf("locations = %+v, want an archive", locations)
	}
}

// readBlob reads a stored blob as text.
func readBlob(t *testing.T, env *testEnv, key string) string {
	t.Helper()
	content, err := env.store.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("blob %s: %v", key, err)
	}
	defer content.Close()
	data, _ := io.ReadAll(content)
	return string(data)
}

// hasKind reports whether one of locations is of kind.
func hasKind(locations []schema.StorageLocation, kind schema.StorageKind) bool {
	for _, location := range locations {
		if location.Kind == kind {
			return true
		}
	}
	return false
}
