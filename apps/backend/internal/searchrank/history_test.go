package searchrank

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// openHistory opens a history on a fresh database.
func openHistory(t *testing.T) (*History, *db.DB) {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return NewHistory(database), database
}

// TestTokenSimilarityToleratesWordStarts checks equal queries are 1, a word
// being typed counts as the full word and unrelated queries are 0.
func TestTokenSimilarityToleratesWordStarts(t *testing.T) {
	if got := TokenSimilarity(QueryTokens("mia khalifa"), QueryTokens("Khalifa  MIA")); got != 1 {
		t.Fatalf("same words = %v", got)
	}
	if got := TokenSimilarity(QueryTokens("mia khal"), QueryTokens("mia khalifa")); got != 1 {
		t.Fatalf("word start = %v", got)
	}
	if got := TokenSimilarity(QueryTokens("mia khalifa"), QueryTokens("mia khalifa hotel")); got < 0.6 || got > 0.7 {
		t.Fatalf("superset = %v", got)
	}
	if got := TokenSimilarity(QueryTokens("garden"), QueryTokens("hotel")); got != 0 {
		t.Fatalf("unrelated = %v", got)
	}
}

// TestClickMemoryDecaysAndNeedsSimilarity checks fresh clicks of the same query
// count fully, old ones half per half-life, dissimilar queries not at all, and
// the memory never exceeds 1.
func TestClickMemoryDecaysAndNeedsSimilarity(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	tokens := QueryTokens("blue lagoon")
	half := now.AddDate(0, 0, -int(ClickHalfLifeDays))
	memory := ClickMemory([]Click{
		{Media: "scene:fresh", Tokens: tokens, At: now},
		{Media: "scene:old", Tokens: tokens, At: half},
		{Media: "scene:other", Tokens: QueryTokens("garden party"), At: now},
		{Media: "scene:repeat", Tokens: tokens, At: now},
		{Media: "scene:repeat", Tokens: tokens, At: now},
		{Media: "scene:similar", Tokens: QueryTokens("blue lagoon night"), At: now},
	}, tokens, now)
	if memory["scene:fresh"] != 1 {
		t.Fatalf("fresh = %v", memory["scene:fresh"])
	}
	if old := memory["scene:old"]; old < 0.49 || old > 0.51 {
		t.Fatalf("old = %v", old)
	}
	if _, found := memory["scene:other"]; found {
		t.Fatal("a dissimilar query must not count")
	}
	if memory["scene:repeat"] != 1 {
		t.Fatalf("repeat should be capped, got %v", memory["scene:repeat"])
	}
	if similar := memory["scene:similar"]; similar <= 0 || similar >= 1 {
		t.Fatalf("similar = %v", similar)
	}
}

// TestSourceSharesFavourClickedPlugins checks a plugin whose results get
// clicked gets a share above 1, an ignored one below, both bounded, and thin
// history changes nothing.
func TestSourceSharesFavourClickedPlugins(t *testing.T) {
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	tokens := QueryTokens("blue lagoon")
	counts := []Count{
		{Tokens: tokens, Plugin: "alpha", At: now, Total: 20},
		{Tokens: tokens, Plugin: "alpha", Clicked: true, At: now, Total: 8},
		{Tokens: tokens, Plugin: "beta", At: now, Total: 20},
		{Tokens: tokens, Plugin: "beta", Clicked: true, At: now, Total: 1},
		{Tokens: tokens, Plugin: "gamma", At: now, Total: 20},
	}
	shares := SourceShares(counts, tokens, now)
	if !(shares["alpha"] > 1 && shares["beta"] < 1 && shares["gamma"] < shares["beta"]) {
		t.Fatalf("shares = %v", shares)
	}
	for plugin, share := range shares {
		if share < 1-PluginShareSpread || share > 1+PluginShareSpread {
			t.Fatalf("%s share %v out of bounds", plugin, share)
		}
	}
	few := []Count{{Tokens: tokens, Plugin: "alpha", At: now, Total: 4}, {Tokens: tokens, Plugin: "alpha", Clicked: true, At: now, Total: 2}}
	if got := SourceShares(few, tokens, now); len(got) != 0 {
		t.Fatalf("thin history = %v", got)
	}
	unrelated := []Count{{Tokens: QueryTokens("garden"), Plugin: "alpha", At: now, Total: 50}, {Tokens: QueryTokens("garden"), Plugin: "alpha", Clicked: true, At: now, Total: 20}}
	if got := SourceShares(unrelated, tokens, now); len(got) != 0 {
		t.Fatalf("other queries must not count: %v", got)
	}
}

// TestHistoryRecordsAndLearns stores events through the database and reads
// them back: a clicked scene is remembered for a similar query only, and a
// plugin with clicks gets a larger share.
func TestHistoryRecordsAndLearns(t *testing.T) {
	history, database := openHistory(t)
	ctx := context.Background()
	var shown []Event
	for index := 0; index < 12; index++ {
		shown = append(shown, Event{Media: "scene:a" + string(rune('a'+index)), Plugin: "alpha", Position: index})
		shown = append(shown, Event{Media: "scene:b" + string(rune('a'+index)), Plugin: "beta", Position: index})
	}
	shown = append(shown, Event{Media: "scene:aa", Plugin: "alpha", Clicked: true}, Event{Media: "scene:ab", Plugin: "alpha", Clicked: true})
	if err := history.Record(ctx, "Blue Lagoon", shown); err != nil {
		t.Fatalf("record: %v", err)
	}

	learned, err := history.Learn(ctx, "lagoon blue")
	if err != nil {
		t.Fatalf("learn: %v", err)
	}
	if learned.ClickMemory["scene:aa"] < 0.99 || learned.ClickMemory["scene:ba"] != 0 {
		t.Fatalf("click memory = %v", learned.ClickMemory)
	}
	if !(learned.SourceShares["alpha"] > 1 && learned.SourceShares["beta"] < 1) {
		t.Fatalf("source shares = %v", learned.SourceShares)
	}

	prefix, err := history.Learn(ctx, "blue lago")
	if err != nil || prefix.ClickMemory["scene:aa"] < 0.99 {
		t.Fatalf("a query still being typed should recall: %v %v", prefix.ClickMemory, err)
	}
	other, err := history.Learn(ctx, "garden party")
	if err != nil || len(other.ClickMemory) != 0 || len(other.SourceShares) != 0 {
		t.Fatalf("unrelated query learned %+v (%v)", other, err)
	}

	if _, err := database.Exec(ctx, `UPDATE search_event SET created_at = '2000-01-01T00:00:00.000Z' WHERE media = 'scene:aa'`, nil); err != nil {
		t.Fatalf("age event: %v", err)
	}
	if err := history.Record(ctx, "blue lagoon", []Event{{Media: "scene:new", Plugin: "alpha"}}); err != nil {
		t.Fatalf("record: %v", err)
	}
	remaining, err := database.Int(ctx, `SELECT count(*) FROM search_event WHERE media = 'scene:aa'`, nil)
	if err != nil || remaining != 0 {
		t.Fatalf("events past retention should be dropped: %d %v", remaining, err)
	}
}
