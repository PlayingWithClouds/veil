package ingest

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
)

// SearchLibrary returns up to limit records of a canonical table matching the
// query, most relevant first. Scenes are ranked by bm25 over the scene_search
// index (title, performers, studio, tags), with plain title substring matches
// filling what the index leaves free; the other tables rank an exact name or
// title first, then prefixes, then word prefixes, then any substring. Scene
// rows carry search_performers, search_studio and search_tags (the index text)
// for callers that score the match themselves.
func (s *Service) SearchLibrary(ctx context.Context, table, query string, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 20
	}
	if table != "scene" {
		return s.searchByName(ctx, table, query, limit)
	}
	matched, err := s.searchSceneIndex(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	if len(matched) >= limit {
		return matched, nil
	}
	return s.fillWithTitleMatches(ctx, matched, query, limit)
}

// searchSceneIndex runs the full-text query against scene_search, best bm25
// first. Title matches weigh most, then performers, studio and tags.
func (s *Service) searchSceneIndex(ctx context.Context, query string, limit int) ([]map[string]any, error) {
	match := FullTextQuery(query)
	if match == "" {
		return nil, nil
	}
	return s.database.Query(ctx, `SELECT scene.*,
			scene_search.performers AS search_performers,
			scene_search.studio AS search_studio,
			scene_search.tags AS search_tags
		FROM scene_search JOIN scene ON scene.id = scene_search.scene_id
		WHERE scene_search MATCH $match
		ORDER BY bm25(scene_search, 0.0, 10.0, 4.0, 3.0, 1.0)
		LIMIT $limit`, db.Vars{"match": match, "limit": limit})
}

// fillWithTitleMatches appends scenes whose title contains the query text
// verbatim (mid-word matches the index cannot see), skipping ones already
// matched, up to limit.
func (s *Service) fillWithTitleMatches(ctx context.Context, matched []map[string]any, query string, limit int) ([]map[string]any, error) {
	found := make([]string, 0, len(matched))
	for _, record := range matched {
		found = append(found, recordIDString(record))
	}
	extra, err := s.database.Query(ctx, `SELECT * FROM scene
		WHERE instr(lower(title), lower($search)) > 0
			AND id NOT IN (SELECT value FROM json_each($found))
		ORDER BY length(title), title LIMIT $limit`,
		db.Vars{"search": query, "found": found, "limit": limit - len(matched)})
	if err != nil {
		return nil, err
	}
	return append(matched, extra...), nil
}

// searchByName lists records of table whose display field contains the query,
// exact matches first, then prefix, word-prefix and mid-word matches, shorter
// names before longer ones.
func (s *Service) searchByName(ctx context.Context, table, query string, limit int) ([]map[string]any, error) {
	field := displayField(table)
	statement := fmt.Sprintf(`SELECT * FROM %[1]s WHERE instr(lower(%[2]s), lower($search)) > 0
		ORDER BY CASE
			WHEN lower(%[2]s) = lower($search) THEN 0
			WHEN instr(lower(%[2]s), lower($search)) = 1 THEN 1
			WHEN instr(lower(%[2]s), ' ' || lower($search)) > 0 THEN 2
			ELSE 3 END, length(%[2]s), %[2]s
		LIMIT $limit`, table, field)
	return s.database.Query(ctx, statement, db.Vars{"search": query, "limit": limit})
}

// FullTextQuery turns the user's text into an FTS5 MATCH expression: every
// word must match (quoted, so no FTS syntax leaks through) and the last one
// may still be being typed, so it matches as a prefix. Empty when the text has
// no words.
func FullTextQuery(text string) string {
	words := strings.Fields(names.Key(text))
	if len(words) == 0 {
		return ""
	}
	terms := make([]string, 0, len(words))
	for _, word := range words {
		terms = append(terms, `"`+word+`"`)
	}
	terms[len(terms)-1] += "*"
	return strings.Join(terms, " ")
}
