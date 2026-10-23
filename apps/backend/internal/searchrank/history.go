package searchrank

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
)

// Event is one result the search page showed or the user opened, with the
// query it appeared for.
type Event struct {
	Media string
	// Plugin is the source of the result: LibrarySource or a plugin name.
	Plugin   string
	Clicked  bool
	Position int
}

// Learned is what the search history says about a query.
type Learned struct {
	// ClickMemory maps a scene id to a value in (0, 1]: how strongly the user
	// opened it from the same or a similar query lately.
	ClickMemory map[string]float64
	// SourceShares maps a source (plugin or LibrarySource) to the factor its
	// rank shares are multiplied by, in [1-PluginShareSpread, 1+PluginShareSpread].
	// Sources without history are absent and count as 1.
	SourceShares map[string]float64
}

// History stores search events and turns them into Learned values. Events are
// kept HistoryRetentionDays; both lookups read only that window and decay
// older events with half-life ClickHalfLifeDays.
type History struct {
	database *db.DB
	now      func() time.Time
}

// NewHistory creates the history on the database's search_event table.
func NewHistory(database *db.DB) *History {
	return &History{database: database, now: time.Now}
}

// QueryTokens reduces a query to its distinct words, sorted, which is how
// events are filed and compared.
func QueryTokens(query string) []string {
	tokens := strings.Fields(names.Key(query))
	slices.Sort(tokens)
	return slices.Compact(tokens)
}

// Record files the events under the query and drops events past retention.
// A query without words records nothing.
func (history *History) Record(ctx context.Context, query string, events []Event) error {
	tokens := QueryTokens(query)
	if len(tokens) == 0 || len(events) == 0 {
		return nil
	}
	now := history.now()
	err := history.database.Tx(ctx, func(tx *db.DB) error {
		for _, event := range events {
			if err := insertEvent(ctx, tx, tokens, event, now); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx, `DELETE FROM search_event WHERE created_at < $cutoff`,
			db.Vars{"cutoff": db.FormatTime(now.AddDate(0, 0, -HistoryRetentionDays))})
		return err
	})
	if err != nil {
		return fmt.Errorf("record search events: %w", err)
	}
	return nil
}

// insertEvent writes one search_event row.
func insertEvent(ctx context.Context, tx *db.DB, tokens []string, event Event, now time.Time) error {
	kind := "shown"
	if event.Clicked {
		kind = "click"
	}
	_, err := tx.Exec(ctx,
		`INSERT INTO search_event (id, tokens, media, plugin, kind, position, created_at)
		 VALUES ($id, $tokens, $media, $plugin, $kind, $position, $createdAt)`,
		db.Vars{
			"id":        db.NewRecordID("search_event"),
			"tokens":    tokens,
			"media":     event.Media,
			"plugin":    event.Plugin,
			"kind":      kind,
			"position":  event.Position,
			"createdAt": db.FormatTime(now),
		})
	return err
}

// overlapCondition keeps search_event rows whose query shares a word (or the
// start of one, three letters at least) with the $tokens being searched.
const overlapCondition = `EXISTS (SELECT 1 FROM json_each(search_event.tokens) AS recorded, json_each($tokens) AS wanted
	WHERE recorded.value = wanted.value
	   OR (length(wanted.value) >= 3 AND recorded.value LIKE wanted.value || '%')
	   OR (length(recorded.value) >= 3 AND wanted.value LIKE recorded.value || '%'))`

// Learn reads the history for a query: click memory per scene and rank-share
// factors per source. Empty (not an error) when the history has nothing.
func (history *History) Learn(ctx context.Context, query string) (Learned, error) {
	learned := Learned{ClickMemory: map[string]float64{}, SourceShares: map[string]float64{}}
	tokens := QueryTokens(query)
	if len(tokens) == 0 {
		return learned, nil
	}
	now := history.now()
	vars := db.Vars{"tokens": tokens, "since": db.FormatTime(now.AddDate(0, 0, -HistoryRetentionDays)), "limit": HistoryRowLimit}
	clicks, err := history.database.Query(ctx,
		`SELECT media, tokens, created_at FROM search_event
		 WHERE kind = 'click' AND created_at >= $since AND `+overlapCondition+`
		 ORDER BY created_at DESC LIMIT $limit`, vars)
	if err != nil {
		return learned, fmt.Errorf("read search clicks: %w", err)
	}
	learned.ClickMemory = ClickMemory(rowsToClicks(clicks), tokens, now)

	counts, err := history.database.Query(ctx,
		`SELECT tokens, plugin, kind, substr(created_at, 1, 10) AS day, count(*) AS total FROM search_event
		 WHERE created_at >= $since AND `+overlapCondition+`
		 GROUP BY tokens, plugin, kind, day ORDER BY day DESC LIMIT $limit`, vars)
	if err != nil {
		return learned, fmt.Errorf("read search rates: %w", err)
	}
	learned.SourceShares = SourceShares(rowsToCounts(counts), tokens, now)
	return learned, nil
}

// Click is an earlier click: which scene, for which query, when.
type Click struct {
	Media  string
	Tokens []string
	At     time.Time
}

// ClickMemory sums, per scene, the decayed similarity of the earlier queries
// it was opened from to the current one, capped at 1. Queries less similar
// than ClickMemoryMinimumSimilarity do not count.
func ClickMemory(clicks []Click, tokens []string, now time.Time) map[string]float64 {
	memory := map[string]float64{}
	for _, click := range clicks {
		similarity := TokenSimilarity(tokens, click.Tokens)
		if similarity < ClickMemoryMinimumSimilarity {
			continue
		}
		memory[click.Media] = min(1, memory[click.Media]+similarity*decay(now.Sub(click.At)))
	}
	return memory
}

// Count is a tally of events of one query and source on one day.
type Count struct {
	Tokens  []string
	Plugin  string
	Clicked bool
	At      time.Time
	Total   int
}

// SourceShares turns the tallies into rank-share factors per source. Each
// tally is weighted by the share of the current query's words its query
// shares and by its age. A source's click rate (decayed clicks over decayed
// showings) is smoothed toward the overall rate with PluginRatePriorShown
// pseudo-showings, and its factor is that rate over the overall rate, kept
// within PluginShareSpread of 1. Nothing is returned while the history holds
// fewer than PluginRateMinimumShown showings or no click at all.
func SourceShares(counts []Count, tokens []string, now time.Time) map[string]float64 {
	clicks, shown := map[string]float64{}, map[string]float64{}
	totalClicks, totalShown := 0.0, 0.0
	for _, count := range counts {
		weight := overlap(tokens, count.Tokens) * decay(now.Sub(count.At)) * float64(count.Total)
		if count.Clicked {
			clicks[count.Plugin] += weight
			totalClicks += weight
		} else {
			shown[count.Plugin] += weight
			totalShown += weight
		}
	}
	shares := map[string]float64{}
	if totalShown < PluginRateMinimumShown || totalClicks == 0 {
		return shares
	}
	overall := totalClicks / totalShown
	for plugin := range shown {
		rate := (clicks[plugin] + overall*PluginRatePriorShown) / (shown[plugin] + PluginRatePriorShown)
		shares[plugin] = max(1-PluginShareSpread, min(1+PluginShareSpread, rate/overall))
	}
	return shares
}

// TokenSimilarity is the share of the words two queries have in common among
// all their words (Jaccard), counting a word that starts the other (three
// letters at least) as the same: "khal" is "khalifa" being typed.
func TokenSimilarity(first, second []string) float64 {
	shared := sharedTokens(first, second)
	union := len(first) + len(second) - shared
	if union <= 0 {
		return 0
	}
	return float64(shared) / float64(union)
}

// overlap is the share of the wanted words that the other query has too.
func overlap(wanted, other []string) float64 {
	if len(wanted) == 0 {
		return 0
	}
	return float64(sharedTokens(wanted, other)) / float64(len(wanted))
}

// sharedTokens counts the words of first that second has too.
func sharedTokens(first, second []string) int {
	shared := 0
	for _, token := range first {
		if slices.ContainsFunc(second, func(other string) bool { return sameToken(token, other) }) {
			shared++
		}
	}
	return shared
}

// sameToken reports whether two query words are equal or one starts the other.
func sameToken(first, second string) bool {
	if first == second {
		return true
	}
	shorter, longer := first, second
	if len(shorter) > len(longer) {
		shorter, longer = longer, shorter
	}
	return len(shorter) >= 3 && strings.HasPrefix(longer, shorter)
}

// decay is the weight of an event of that age: 1 now, 1/2 after
// ClickHalfLifeDays.
func decay(age time.Duration) float64 {
	days := max(0, age.Hours()/24)
	return math.Pow(0.5, days/ClickHalfLifeDays)
}

// rowsToClicks reads the click rows.
func rowsToClicks(rows []db.Row) []Click {
	clicks := make([]Click, 0, len(rows))
	for _, row := range rows {
		media, _ := row["media"].(string)
		clicks = append(clicks, Click{Media: media, Tokens: rowTokens(row), At: rowTime(row, "created_at")})
	}
	return clicks
}

// rowsToCounts reads the per-day tally rows.
func rowsToCounts(rows []db.Row) []Count {
	counts := make([]Count, 0, len(rows))
	for _, row := range rows {
		plugin, _ := row["plugin"].(string)
		kind, _ := row["kind"].(string)
		day, _ := row["day"].(string)
		at, _ := time.Parse("2006-01-02", day)
		counts = append(counts, Count{
			Tokens: rowTokens(row), Plugin: plugin, Clicked: kind == "click",
			At: at.Add(12 * time.Hour), Total: db.AsInt(row["total"]),
		})
	}
	return counts
}

// rowTokens reads the decoded JSON word list of a row.
func rowTokens(row db.Row) []string {
	items, _ := row["tokens"].([]any)
	tokens := make([]string, 0, len(items))
	for _, item := range items {
		if token, ok := item.(string); ok {
			tokens = append(tokens, token)
		}
	}
	return tokens
}

// rowTime reads a stored timestamp, the zero time when unreadable.
func rowTime(row db.Row, key string) time.Time {
	text, _ := row[key].(string)
	parsed, err := time.Parse(db.TimeFormat, text)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
