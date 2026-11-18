package dedupe

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// fakeImages serves generated pictures by URL.
type fakeImages map[string][]byte

// Load returns the picture stored under raw.
func (images fakeImages) Load(_ context.Context, raw string) ([]byte, error) {
	data, found := images[raw]
	if !found {
		return nil, fmt.Errorf("no image %s", raw)
	}
	return data, nil
}

// picture draws a gradient whose direction depends on seed, as PNG bytes.
func picture(t *testing.T, seed int) []byte {
	t.Helper()
	canvas := image.NewGray(image.Rect(0, 0, 64, 36))
	for y := range 36 {
		for x := range 64 {
			value := (x*seed*7 + y*(11-seed)*5) % 256
			canvas.SetGray(x, y, color.Gray{Y: uint8(value)})
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, canvas); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

// scene is a scene page from a site.
func scene(externalID, title, cover string, duration int) []*plugins.ScrapeResult {
	return []*plugins.ScrapeResult{{Type: plugins.MediaTypeScene, Scene: &plugins.Scene{
		ExternalID: externalID, SourceURL: "https://site.test/" + externalID, Title: title, Duration: duration, PosterPath: cover,
	}}}
}

// TestPassMergesSameCover stores one scene from two sites under unrelated
// titles with the same cover, and a different scene of the same length:
// the pass merges the copies and leaves the other alone.
func TestPassMergesSameCover(t *testing.T) {
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	lib := library.New(database)
	images := fakeImages{
		"https://a.test/cover.jpg": picture(t, 3),
		"https://b.test/cover.jpg": picture(t, 3),
		"https://c.test/cover.jpg": picture(t, 8),
	}
	first, err := lib.IngestFound(ctx, "site-a", scene("a1", "Hot blonde by the pool", "https://a.test/cover.jpg", 1800))
	if err != nil {
		t.Fatal(err)
	}
	second, err := lib.IngestFound(ctx, "site-b", scene("b1", "Blondine am Pool", "https://b.test/cover.jpg", 1803))
	if err != nil {
		t.Fatal(err)
	}
	other, err := lib.IngestFound(ctx, "site-c", scene("c1", "Something else", "https://c.test/cover.jpg", 1800))
	if err != nil {
		t.Fatal(err)
	}
	if err := New(lib, cache.New(ctx, database), images).Pass(ctx); err != nil {
		t.Fatal(err)
	}
	if survivor, _ := lib.Video(ctx, first.Id); survivor == nil {
		t.Fatal("the first copy is gone")
	}
	if merged, _ := lib.Video(ctx, second.Id); merged != nil {
		t.Error("the second copy wasn't merged")
	}
	if kept, _ := lib.Video(ctx, other.Id); kept == nil {
		t.Error("an unrelated video of the same length was merged")
	}
	sources, err := lib.Sources(ctx, schema.RecordKindVideo, first.Id)
	if err != nil || len(sources) != 2 {
		t.Errorf("survivor sources = %d, err %v; want both sites", len(sources), err)
	}
}
