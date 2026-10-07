package suggest

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// Library size of the large fixture: roughly a heavily used install.
const (
	largeSceneCount     = 50000
	largePerformerCount = 15000
	largeTagCount       = 3000
	largeStudioCount    = 1500
	largeVocabularySize = 3000
)

// syllables build the fixture's made-up words and names.
var syllables = []string{"ka", "ri", "lo", "me", "sa", "ta", "ni", "ve", "zo", "ul", "an", "br", "st", "el", "or", "ph", "qu", "ex"}

// fakeWord returns a deterministic pseudo word of two to four syllables.
func fakeWord(random *rand.Rand) string {
	var builder strings.Builder
	for range 2 + random.IntN(3) {
		builder.WriteString(syllables[random.IntN(len(syllables))])
	}
	return builder.String()
}

// seedLargeLibrary fills the database with a large random library in one
// transaction and returns the vocabulary the titles were drawn from.
func seedLargeLibrary(t testing.TB, database *db.DB) []string {
	t.Helper()
	random := rand.New(rand.NewPCG(1, 2))
	vocabulary := make([]string, largeVocabularySize)
	for index := range vocabulary {
		vocabulary[index] = fakeWord(random)
	}
	err := database.Tx(context.Background(), func(tx *db.DB) error {
		if err := insertLargeEntities(tx, random); err != nil {
			return err
		}
		return insertLargeScenes(tx, random, vocabulary)
	})
	if err != nil {
		t.Fatalf("seed large library: %v", err)
	}
	return vocabulary
}

// insertLargeEntities stores the fixture's tags, performers and studios.
func insertLargeEntities(tx *db.DB, random *rand.Rand) error {
	tables := []struct {
		table string
		count int
	}{{"tag", largeTagCount}, {"performer", largePerformerCount}, {"studio", largeStudioCount}}
	for _, entry := range tables {
		for number := range entry.count {
			name := fmt.Sprintf("%s %s %d", fakeWord(random), fakeWord(random), number)
			_, err := tx.Exec(context.Background(),
				fmt.Sprintf(`INSERT INTO %s (id, name, aliases) VALUES ($id, $name, $aliases)`, entry.table),
				db.Vars{"id": fmt.Sprintf("%s:%d", entry.table, number), "name": name, "aliases": []string{fakeWord(random)}})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// insertLargeScenes stores the fixture's scenes with titles drawn from a
// skewed vocabulary, a few tags, performers and a studio each.
func insertLargeScenes(tx *db.DB, random *rand.Rand, vocabulary []string) error {
	for number := range largeSceneCount {
		titleWords := make([]string, 4+random.IntN(6))
		for index := range titleWords {
			// Squaring skews the draw towards the start, like real word frequencies.
			draw := random.Float64()
			titleWords[index] = vocabulary[int(draw*draw*float64(len(vocabulary)))]
		}
		_, err := tx.Exec(context.Background(),
			`INSERT INTO scene (id, source_url, title, tags, performers, studio) VALUES ($id, $url, $title, $tags, $performers, $studio)`,
			db.Vars{
				"id":         fmt.Sprintf("scene:%d", number),
				"url":        fmt.Sprintf("https://example.test/%d", number),
				"title":      strings.Join(titleWords, " "),
				"tags":       []string{fmt.Sprintf("tag:%d", random.IntN(largeTagCount)), fmt.Sprintf("tag:%d", random.IntN(largeTagCount))},
				"performers": []string{fmt.Sprintf("performer:%d", random.IntN(largePerformerCount))},
				"studio":     fmt.Sprintf("studio:%d", random.IntN(largeStudioCount)),
			})
		if err != nil {
			return err
		}
	}
	return nil
}

// largeQueries mixes short prefixes, whole words, typos and misses.
func largeQueries(vocabulary []string) []string {
	queries := []string{"", "k", "ka", "kar", "ri lo", "zzzz", "karilo"}
	for _, word := range vocabulary[:20] {
		queries = append(queries, word[:3], word, word[:len(word)-1]+"x", word+" "+vocabulary[1][:2])
	}
	return queries
}

func TestSuggestOnLargeLibraryIsFast(t *testing.T) {
	if testing.Short() {
		t.Skip("large fixture")
	}
	service, database, _, _ := newTestService(t)
	vocabulary := seedLargeLibrary(t, database)
	buildStarted := time.Now()
	service.Warm(context.Background())
	t.Logf("index build: %v", time.Since(buildStarted))

	queries := largeQueries(vocabulary)
	slowest := time.Duration(0)
	for _, query := range queries {
		started := time.Now()
		list := suggestions(t, service, query, DefaultLimit)
		slowest = max(slowest, time.Since(started))
		if len(list) > DefaultLimit {
			t.Errorf("query %q returned %d suggestions", query, len(list))
		}
	}
	t.Logf("slowest of %d queries: %v", len(queries), slowest)
	// Generous bound so a loaded CI machine doesn't flake; the target is < 20 ms.
	if slowest > 200*time.Millisecond {
		t.Errorf("slowest query took %v", slowest)
	}
	if list := suggestions(t, service, vocabulary[0][:3], DefaultLimit); len(list) == 0 {
		t.Error("a common prefix must yield suggestions")
	}
}

func BenchmarkSuggest(b *testing.B) {
	service, database, _, _ := newTestService(b)
	vocabulary := seedLargeLibrary(b, database)
	service.Warm(context.Background())
	queries := largeQueries(vocabulary)
	b.ResetTimer()
	for iteration := 0; b.Loop(); iteration++ {
		if _, err := service.Suggest(context.Background(), queries[iteration%len(queries)], DefaultLimit); err != nil {
			b.Fatal(err)
		}
	}
}
