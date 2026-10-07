package suggest

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// pastSearch is one recent_search row.
type pastSearch struct {
	query          string
	matchQuery     string
	count          int
	lastSearchedAt time.Time
}

// RecordSearch remembers a submitted search: a new row on first sight, else
// one more run with the latest spelling as display text. Blank queries are
// ignored and report false.
func (s *Service) RecordSearch(ctx context.Context, query string) (bool, error) {
	normalized := NormalizeQuery(query)
	if normalized == "" {
		return false, nil
	}
	_, err := s.database.Exec(ctx,
		`INSERT INTO recent_search (id, query, normalized_query, search_count, last_searched_at)
		 VALUES ($id, $query, $normalized, 1, $now)
		 ON CONFLICT (normalized_query) DO UPDATE SET
		   query = excluded.query,
		   search_count = search_count + 1,
		   last_searched_at = excluded.last_searched_at`,
		db.Vars{
			"id":         db.NewRecordID("recent_search"),
			"query":      strings.Join(strings.Fields(query), " "),
			"normalized": normalized,
			"now":        db.FormatTime(s.now()),
		})
	if err != nil {
		return false, fmt.Errorf("record search: %w", err)
	}
	return true, nil
}

// ForgetSearch removes a query from the recent searches, reporting whether
// it was there.
func (s *Service) ForgetSearch(ctx context.Context, query string) (bool, error) {
	removed, err := s.database.Exec(ctx,
		`DELETE FROM recent_search WHERE normalized_query = $normalized`,
		db.Vars{"normalized": NormalizeQuery(query)})
	if err != nil {
		return false, fmt.Errorf("forget search: %w", err)
	}
	return removed > 0, nil
}

// pastSearches reads the most recent searches, newest first.
func (s *Service) pastSearches(ctx context.Context, limit int) ([]pastSearch, error) {
	rows, err := s.database.Query(ctx,
		`SELECT query, search_count, last_searched_at FROM recent_search
		 ORDER BY last_searched_at DESC LIMIT $limit`,
		db.Vars{"limit": limit})
	if err != nil {
		return nil, fmt.Errorf("recent searches: %w", err)
	}
	out := make([]pastSearch, 0, len(rows))
	for _, row := range rows {
		lastSearchedAt, _ := time.Parse(db.TimeFormat, rowString(row, "last_searched_at"))
		out = append(out, pastSearch{
			query:          rowString(row, "query"),
			matchQuery:     matchText(rowString(row, "query")),
			count:          db.AsInt(row["search_count"]),
			lastSearchedAt: lastSearchedAt,
		})
	}
	return out, nil
}

// popularity mixes how often and how recently the search ran, in [0, 1].
func (search pastSearch) popularity(now time.Time) float64 {
	countPart := logPopularity(search.count, historyReferenceCount)
	age := now.Sub(search.lastSearchedAt)
	recencyPart := 1.0
	if age > 0 {
		recencyPart = math.Pow(0.5, float64(age)/float64(historyHalfLife))
	}
	return historyCountShare*countPart + (1-historyCountShare)*recencyPart
}

// logPopularity maps a count onto [0, 1] logarithmically, reaching 1 at reference.
func logPopularity(count, reference int) float64 {
	if count <= 0 {
		return 0
	}
	return min(1, math.Log1p(float64(count))/math.Log1p(float64(reference)))
}
