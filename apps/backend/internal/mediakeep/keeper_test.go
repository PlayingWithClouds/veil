package mediakeep

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/storage"
)

// TestReplaceable keeps blob posters and other sites' posters.
func TestReplaceable(t *testing.T) {
	listed := "https://b2.trafficdeposit.com/pivi/2/vid/new/small.jpg"
	cases := []struct {
		name    string
		current string
		want    bool
	}{
		{"empty", "", true},
		{"same site, other host", "https://b1.trafficdeposit.com/pivi/1/vid/old/small.jpg", true},
		{"already kept", "http://localhost:8080/api/blob/posters/sxyprn-1.jpg", false},
		{"other site", "https://static.eporner.com/thumbs/1.jpg", false},
	}
	for _, testCase := range cases {
		if got := replaceable(testCase.current, listed); got != testCase.want {
			t.Errorf("%s: replaceable = %v, want %v", testCase.name, got, testCase.want)
		}
	}
	if replaceable("", "") {
		t.Error("an empty listed poster must not replace anything")
	}
}

// TestPosterExtension ignores the query and falls back to .jpg.
func TestPosterExtension(t *testing.T) {
	if got := posterExtension("https://cdn.example.com/a/0.webp?x=1"); got != ".webp" {
		t.Errorf("got %q", got)
	}
	if got := posterExtension("https://cdn.example.com/a/thumb"); got != ".jpg" {
		t.Errorf("got %q", got)
	}
}

// TestKeepStoresPosterAndRefreshesPreview copies an expiring poster into the
// blob store and swaps in the listing's fresh preview URL.
func TestKeepStoresPosterAndRefreshesPreview(t *testing.T) {
	ctx := context.Background()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "image/jpeg")
		writer.Write([]byte("jpeg bytes"))
	}))
	defer server.Close()

	directory := t.TempDir()
	database, err := db.Open(ctx, filepath.Join(directory, "veil.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	store, err := storage.New(filepath.Join(directory, "blobs"), "http://localhost:8080")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(ctx,
		`INSERT INTO scene (id, title, source_url, poster_path, preview_video) VALUES ('scene:a', 'A', 'https://sxyprn.com/post/a.html', $poster, $preview)`,
		db.Vars{"poster": server.URL + "/old/small.jpg", "preview": server.URL + "/old/vidthumb.mp4"}); err != nil {
		t.Fatal(err)
	}

	keeper := &Keeper{database: database, store: store, lookup: func(string) (*plugins.Plugin, bool) { return nil, false }}
	item := plugins.DiscoveredItem{
		ExternalID: "sxyprn-a", SourceURL: "https://sxyprn.com/post/a.html",
		PosterPath: server.URL + "/new/small.jpg", PreviewVideo: server.URL + "/new/vidthumb.mp4",
	}
	if err := keeper.keep(ctx, stub{sceneID: "scene:a", item: item}); err != nil {
		t.Fatal(err)
	}

	rows, err := database.Query(ctx, `SELECT poster_path, preview_video FROM scene WHERE id = 'scene:a'`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := rows[0]["poster_path"]; got != "http://localhost:8080/api/blob/posters/sxyprn-a.jpg" {
		t.Errorf("poster_path = %v", got)
	}
	if got := rows[0]["preview_video"]; got != item.PreviewVideo {
		t.Errorf("preview_video = %v", got)
	}
	if !store.Exists(ctx, "posters/sxyprn-a.jpg") {
		t.Error("poster blob missing")
	}
}
