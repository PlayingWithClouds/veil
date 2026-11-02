package files

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// get requests path from handler with an optional Range header.
func get(handler http.Handler, path, byteRange string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	if byteRange != "" {
		request.Header.Set("Range", byteRange)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

// TestHandlerServesRangesInsideLocations serves a library file with range
// requests and refuses rows whose path, or symlink, leaves the location.
func TestHandlerServesRangesInsideLocations(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	root, folder := env.libraryRoot(t, nil)
	content := patterned(4096, 1)
	writeFile(t, filepath.Join(folder, "clip.mp4"), content)
	env.scan(t, root.Id)
	clip := addedAt(t, env.added, "media/clip.mp4")
	handler := env.service.Handler()

	if url := env.service.FileURL(ctx, clip); url != "http://veil.test/api/file/"+clip.Id {
		t.Errorf("FileURL = %s", url)
	}
	response := get(handler, FileRoutePrefix+clip.Id, "bytes=0-3")
	if response.Code != http.StatusPartialContent || !bytes.Equal(response.Body.Bytes(), content[:4]) {
		t.Errorf("range: %d %v", response.Code, response.Body.Bytes())
	}
	if response.Header().Get("Content-Type") != "video/mp4" {
		t.Errorf("content type = %s", response.Header().Get("Content-Type"))
	}

	outside := filepath.Join(t.TempDir(), "secret.mp4")
	writeFile(t, outside, []byte("secret"))
	if err := os.Symlink(outside, filepath.Join(folder, "link.mp4")); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"../" + filepath.Base(filepath.Dir(outside)) + "/secret.mp4", "media/link.mp4"} {
		escaping := clip
		escaping.Id = db.NewID()
		escaping.Path = path
		if _, err := env.database.Insert(ctx, tableMediaFile, escaping); err != nil {
			t.Fatal(err)
		}
		if response := get(handler, FileRoutePrefix+escaping.Id, ""); response.Code != http.StatusNotFound {
			t.Errorf("%s served with %d", path, response.Code)
		}
	}
}

// TestHandlerServesBlobStoreFiles serves a file of the blob store through
// the store, and FileURL points at its blob URL.
func TestHandlerServesBlobStoreFiles(t *testing.T) {
	ctx := context.Background()
	env := newTestEnv(t)
	video := env.newVideo(t, "Downloaded")
	content := []byte("blob content")
	key := "stream-cache/mp4/abc.mp4"
	if err := env.store.Put(ctx, key, bytes.NewReader(content), int64(len(content)), "video/mp4"); err != nil {
		t.Fatal(err)
	}
	location, err := env.service.downloadLocation(ctx)
	if err != nil {
		t.Fatal(err)
	}
	file := schema.MediaFile{SubjectKind: video.Kind, SubjectId: video.Id, Location: location.Id, Path: key, SizeBytes: int64(len(content)), MimeType: "video/mp4"}
	file.Id = db.NewID()
	if err := env.service.insertFile(ctx, file); err != nil {
		t.Fatal(err)
	}
	if url := env.service.FileURL(ctx, file); url != env.store.DirectURL(key) {
		t.Errorf("FileURL = %s", url)
	}
	response := get(env.service.Handler(), FileRoutePrefix+file.Id, "")
	if response.Code != http.StatusOK || response.Body.String() != string(content) {
		t.Errorf("blob: %d %q", response.Code, response.Body.String())
	}
}
