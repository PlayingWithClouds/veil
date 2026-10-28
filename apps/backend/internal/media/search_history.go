package media

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// searchResultsKept is how many of a search's scenes are stored for the
// recommender, best first.
const searchResultsKept = 48

// searchResultRetention is how long a search's stored scenes are kept; the
// recommender only reads the last two weeks of searches.
const searchResultRetention = 30 * 24 * time.Hour

// LogSearch records one search with the scenes it returned (best first) and
// bumps the deduped search_term row it maps to. The writes are recommendation
// signal only, so callers run this in the background and treat failures as
// non-fatal.
func (r *Repository) LogSearch(ctx context.Context, query string, plugins []string, resultCount int, sceneIDs []string) error {
	normalized := normalizeQuery(query)
	if normalized == "" {
		return nil
	}
	if plugins == nil {
		plugins = []string{}
	}

	searchID, err := r.createSearchHistory(ctx, query, normalized, plugins, resultCount)
	if err != nil {
		return err
	}
	if err := r.storeSearchResults(ctx, searchID, sceneIDs); err != nil {
		return err
	}
	return r.bumpSearchTerm(ctx, normalized)
}

// createSearchHistory inserts the search_history row of one search and returns its id.
func (r *Repository) createSearchHistory(ctx context.Context, query, normalized string, plugins []string, resultCount int) (string, error) {
	id, err := r.database.Insert(ctx, "search_history", map[string]any{
		"query":            query,
		"normalized_query": normalized,
		"plugins":          plugins,
		"result_count":     resultCount,
	})
	if err != nil {
		return "", fmt.Errorf("create search history: %w", err)
	}
	return id.String(), nil
}

// storeSearchResults stores the first searchResultsKept scenes of a search in
// their order and drops the stored results of searches past retention.
func (r *Repository) storeSearchResults(ctx context.Context, searchID string, sceneIDs []string) error {
	if len(sceneIDs) > searchResultsKept {
		sceneIDs = sceneIDs[:searchResultsKept]
	}
	_, err := r.database.Exec(ctx,
		`INSERT OR IGNORE INTO search_result (search, scene, rank)
		 SELECT $search, value, key FROM json_each($scenes)`,
		db.Vars{"search": searchID, "scenes": sceneIDs})
	if err != nil {
		return fmt.Errorf("store search results: %w", err)
	}
	_, err = r.database.Exec(ctx, `DELETE FROM search_result WHERE created_at < $before`,
		db.Vars{"before": db.FormatTime(time.Now().Add(-searchResultRetention))})
	if err != nil {
		return fmt.Errorf("prune search results: %w", err)
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
