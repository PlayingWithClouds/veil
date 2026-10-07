package ingest

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
)

// displayField is the searchable/orderable label column for a canonical table.
// Scenes and galleries carry `title`; performers and studios carry `name`.
func displayField(table string) string {
	switch table {
	case "performer", "studio":
		return "name"
	default:
		return "title"
	}
}

// ListByTable queries a typed canonical table with optional name/title search and
// pagination.
func (s *Service) ListByTable(ctx context.Context, table, search string, limit, offset int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 500
	}
	field := displayField(table)
	params := db.Vars{"limit": limit, "offset": offset}
	var query string
	if search != "" {
		query = fmt.Sprintf(`SELECT * FROM %s WHERE instr(lower(%s), lower($search)) > 0 ORDER BY %s LIMIT $limit OFFSET $offset`, table, field, field)
		params["search"] = search
	} else {
		query = fmt.Sprintf(`SELECT * FROM %s ORDER BY %s LIMIT $limit OFFSET $offset`, table, field)
	}
	return s.database.Query(ctx, query, params)
}

// GetByID retrieves a single record from its typed table by "table:id" string.
// Returns nil when the record does not exist.
func (s *Service) GetByID(ctx context.Context, id string) (map[string]any, error) {
	rid, err := db.ParseRecordID(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id %q: %w", id, err)
	}
	return s.database.Get(ctx, *rid)
}

// CountByTable returns the number of matching records in a typed table.
func (s *Service) CountByTable(ctx context.Context, table, search string) (int, error) {
	if search == "" {
		return s.database.Int(ctx, fmt.Sprintf(`SELECT count(*) FROM %s`, table), nil)
	}
	return s.database.Int(ctx,
		fmt.Sprintf(`SELECT count(*) FROM %s WHERE instr(lower(%s), lower($search)) > 0`, table, displayField(table)),
		db.Vars{"search": search})
}

// DeleteByID removes a record from its typed table.
func (s *Service) DeleteByID(ctx context.Context, id string) error {
	rid, err := db.ParseRecordID(id)
	if err != nil {
		return fmt.Errorf("invalid id %q: %w", id, err)
	}
	return s.database.Delete(ctx, *rid)
}

// UpdateRecord merges a partial field map into an existing canonical record.
// Fields without a column are dropped by the db layer; updated_at is bumped by
// the table trigger.
func (s *Service) UpdateRecord(ctx context.Context, id string, fields map[string]any) error {
	rid, err := db.ParseRecordID(id)
	if err != nil {
		return fmt.Errorf("invalid id %q: %w", id, err)
	}
	// created_at is set once at insert time and must never be overwritten.
	delete(fields, "id")
	delete(fields, "created_at")
	return s.database.Merge(ctx, *rid, fields)
}

// GetByIDWithObs fetches a canonical record and overlays observation-only fields
// (images, downloads) from the most recent observation.
func (s *Service) GetByIDWithObs(ctx context.Context, id string) (map[string]any, error) {
	m, err := s.GetByID(ctx, id)
	if err != nil || m == nil {
		return m, err
	}

	obs, obsErr := s.latestObservation(ctx, id)
	if obsErr != nil || obs == nil {
		return m, nil
	}
	for _, field := range []string{"images", "downloads"} {
		if v, ok := obs[field]; ok && !isEmpty(v) {
			m[field] = v
		}
	}
	return m, nil
}

// PrimarySource returns the plugin and source URL to re-scrape a canonical
// record: the record's own source_url plus the plugin from its most recent
// observation.
func (s *Service) PrimarySource(ctx context.Context, m map[string]any) (pluginName, sourceURL string) {
	sourceURL, _ = m["source_url"].(string)

	id := recordIDString(m)
	if _, err := db.ParseRecordID(id); err != nil {
		return "", sourceURL
	}
	plugins, err := s.database.Strings(ctx,
		`SELECT plugin FROM observation WHERE target = $target ORDER BY observed_at DESC, rowid DESC LIMIT 1`,
		db.Vars{"target": id})
	if err != nil || len(plugins) == 0 {
		return "", sourceURL
	}
	return plugins[0], sourceURL
}

// TableFromID extracts the table name from a "table:id" string.
func TableFromID(id string) string {
	parts := strings.SplitN(id, ":", 2)
	return parts[0]
}
