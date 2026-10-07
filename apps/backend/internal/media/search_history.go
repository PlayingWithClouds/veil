package media

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
)

// LogSearch records one search and bumps the deduped search_term row it maps
// to. Both writes are recommendation signal only, so callers run this in the
// background and treat failures as non-fatal.
func (r *Repository) LogSearch(ctx context.Context, query string, plugins []string, resultCount int) error {
	normalized := normalizeQuery(query)
	if normalized == "" {
		return nil
	}
	if plugins == nil {
		plugins = []string{}
	}

	if err := r.createSearchHistory(ctx, query, normalized, plugins, resultCount); err != nil {
		return err
	}
	return r.bumpSearchTerm(ctx, normalized)
}

func (r *Repository) createSearchHistory(ctx context.Context, query, normalized string, plugins []string, resultCount int) error {
	_, err := r.database.Insert(ctx, "search_history", map[string]any{
		"query":            query,
		"normalized_query": normalized,
		"plugins":          plugins,
		"result_count":     resultCount,
	})
	if err != nil {
		return fmt.Errorf("create search history: %w", err)
	}
	return nil
}

// bumpSearchTerm increments the term's use count, creating it on first sight.
func (r *Repository) bumpSearchTerm(ctx context.Context, normalized string) error {
	_, err := r.database.Exec(ctx,
		`INSERT INTO search_term (id, normalized_query, uses, last_used) VALUES ($id, $normalized, 1, $now)
			ON CONFLICT (normalized_query) DO UPDATE SET uses = uses + 1, last_used = excluded.last_used`,
		db.Vars{"id": db.NewRecordID("search_term"), "normalized": normalized, "now": db.Now()})
	if err != nil {
		return fmt.Errorf("bump search term: %w", err)
	}
	return nil
}

func normalizeQuery(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}
