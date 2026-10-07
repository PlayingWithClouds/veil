package alike

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestBuildQueriesPerformerDriven(t *testing.T) {
	queries := BuildQueries(Target{
		Performers: []string{"Riley Reid", "Manuel Ferrara"},
		Studio:     "Blacked",
		Title:      "Poolside Seduction 1080p",
	})

	if len(queries) == 0 {
		t.Fatal("expected performer-driven queries")
	}
	// Performers lead; they are the strongest cross-site identifier.
	if queries[0] != "Riley Reid Manuel Ferrara" {
		t.Fatalf("first query should combine performers, got %q", queries[0])
	}
	// The title's meaningful words are also searched (junk tokens stripped), since
	// they can carry identifying info even when the full title differs.
	if !containsQuery(queries, "poolside seduction") {
		t.Fatalf("expected a title-token query alongside the performer queries, got %v", queries)
	}
}

func containsQuery(queries []string, want string) bool {
	for _, query := range queries {
		if query == want {
			return true
		}
	}
	return false
}

func TestBuildQueriesFallsBackToTitle(t *testing.T) {
	queries := BuildQueries(Target{Title: "Sunny Afternoon HD Video 1080p"})
	if len(queries) != 1 {
		t.Fatalf("expected a single title fallback query, got %d", len(queries))
	}
	// Junk/quality tokens are stripped from the fallback.
	if queries[0] != "sunny afternoon" {
		t.Fatalf("title fallback should drop junk tokens, got %q", queries[0])
	}
}

func TestNormalizeName(t *testing.T) {
	if got := normalizeName("Riley  Reid!"); got != "riley reid" {
		t.Fatalf("normalizeName = %q", got)
	}
	if got := normalizeName("O'Neil"); got != "o neil" {
		t.Fatalf("normalizeName = %q", got)
	}
}

func TestTitleTokensStripSiteAndJunk(t *testing.T) {
	tokens := titleTokens("Hot Scene - SpankBang 1080p Free")
	set := tokenSet(tokens)
	for _, banned := range []string{"spankbang", "1080p", "free", "scene"} {
		if _, ok := set[banned]; ok {
			t.Fatalf("token %q should have been stripped: %v", banned, tokens)
		}
	}
	if _, ok := set["hot"]; !ok {
		t.Fatalf("meaningful token dropped: %v", tokens)
	}
}

func TestDHashSimilarity(t *testing.T) {
	base := bandImage(0)
	same := bandImage(0)
	shifted := bandImage(15) // mild brightness shift, same structure
	inverted := invertedBands()

	baseHash := mustHash(t, base)
	if HammingSimilarity(baseHash, mustHash(t, same)) != 1 {
		t.Fatal("identical images must hash identically")
	}
	if sim := HammingSimilarity(baseHash, mustHash(t, shifted)); sim < 0.9 {
		t.Fatalf("a mild brightness shift should stay similar, got %.3f", sim)
	}
	if sim := HammingSimilarity(baseHash, mustHash(t, inverted)); sim > 0.5 {
		t.Fatalf("an inverted image should be dissimilar, got %.3f", sim)
	}
}

func TestHammingSimilarityZeroHashIsNeutral(t *testing.T) {
	if HammingSimilarity(0, 12345) != 0 {
		t.Fatal("a zero (uncomputed) hash must contribute no similarity")
	}
}

func mustHash(t *testing.T, img image.Image) uint64 {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode: %v", err)
	}
	hash, err := DecodePosterHash(buf.Bytes())
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	return hash
}

// bandPattern is a non-monotonic set of column brightnesses, so adjacent-pixel
// comparisons yield a mix of set and unset bits (not an all-zero/all-ones hash).
var bandPattern = []int{40, 210, 90, 250, 20, 160, 70, 200}

// bandImage paints vertical brightness bands from bandPattern, with an overall
// brightness offset of delta. The offset shifts brightness but preserves the
// relative order of neighbouring bands, so the difference hash is unchanged.
func bandImage(delta int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 64, 32))
	for x := 0; x < 64; x++ {
		level := bandPattern[(x/8)%len(bandPattern)] + delta
		value := uint8(clampByte(level))
		for y := 0; y < 32; y++ {
			img.Set(x, y, color.RGBA{value, value, value, 255})
		}
	}
	return img
}

// invertedBands reverses every band's brightness, flipping the comparisons.
func invertedBands() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 64, 32))
	for x := 0; x < 64; x++ {
		level := 255 - bandPattern[(x/8)%len(bandPattern)]
		value := uint8(clampByte(level))
		for y := 0; y < 32; y++ {
			img.Set(x, y, color.RGBA{value, value, value, 255})
		}
	}
	return img
}

func clampByte(value int) int {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return value
}
