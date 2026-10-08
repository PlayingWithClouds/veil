package media

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// savedFilterRecord mirrors the `saved_filter` table. Entity references are stored
// as plain string ids so the preset round-trips into GraphQL filter args directly.
type savedFilterRecord struct {
	ID          *db.RecordID `json:"id,omitempty"`
	Name        string       `json:"name"`
	Search      *string      `json:"search,omitempty"`
	StudioID    *string      `json:"studio_id,omitempty"`
	PerformerID *string      `json:"performer_id,omitempty"`
	TagID       *string      `json:"tag_id,omitempty"`
	MinRating   *float64     `json:"min_rating,omitempty"`
	MinDuration *int         `json:"min_duration,omitempty"`
	MaxDuration *int         `json:"max_duration,omitempty"`
	DateFrom    *string      `json:"date_from,omitempty"`
	DateTo      *string      `json:"date_to,omitempty"`
	Sort        *string      `json:"sort,omitempty"`
	Sources     []string     `json:"sources,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}

func (f savedFilterRecord) toModel() *model.SavedFilter {
	return &model.SavedFilter{
		ID:   recordIDString(f.ID),
		Name: f.Name,
		Filter: &model.SceneFilter{
			Search:      f.Search,
			StudioID:    f.StudioID,
			PerformerID: f.PerformerID,
			TagID:       f.TagID,
			MinRating:   f.MinRating,
			MinDuration: f.MinDuration,
			MaxDuration: f.MaxDuration,
			DateFrom:    f.DateFrom,
			DateTo:      f.DateTo,
			Sort:        f.Sort,
			Sources:     f.Sources,
		},
		CreatedAt: f.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// SavedFilterInput carries the filter fields to persist. Mirrors the GraphQL
// SceneFilterInput.
type SavedFilterInput struct {
	Search      *string
	StudioID    *string
	PerformerID *string
	TagID       *string
	MinRating   *float64
	MinDuration *int
	MaxDuration *int
	DateFrom    *string
	DateTo      *string
	Sort        *string
	Sources     []string
}

// ListSavedFilters returns the saved filter presets, newest first.
func (r *Repository) ListSavedFilters(ctx context.Context) ([]*model.SavedFilter, error) {
	rows, err := db.QueryAs[savedFilterRecord](ctx, r.database,
		`SELECT * FROM saved_filter ORDER BY created_at DESC`, nil)
	if err != nil {
		return nil, fmt.Errorf("list saved filters: %w", err)
	}
	out := make([]*model.SavedFilter, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.toModel())
	}
	return out, nil
}

// CreateSavedFilter persists a named filter preset. Empty strings are stored as
// NULL so they round-trip as unset filter fields.
func (r *Repository) CreateSavedFilter(ctx context.Context, name string, filter SavedFilterInput) (*model.SavedFilter, error) {
	content := map[string]any{"name": name}
	setIfString(content, "search", filter.Search)
	setIfString(content, "studio_id", filter.StudioID)
	setIfString(content, "performer_id", filter.PerformerID)
	setIfString(content, "tag_id", filter.TagID)
	setIfString(content, "date_from", filter.DateFrom)
	setIfString(content, "date_to", filter.DateTo)
	setIfString(content, "sort", filter.Sort)
	if len(filter.Sources) > 0 {
		content["sources"] = filter.Sources
	}
	if filter.MinRating != nil {
		content["min_rating"] = *filter.MinRating
	}
	if filter.MinDuration != nil {
		content["min_duration"] = *filter.MinDuration
	}
	if filter.MaxDuration != nil {
		content["max_duration"] = *filter.MaxDuration
	}
	id, err := r.database.Insert(ctx, "saved_filter", content)
	if err != nil {
		return nil, fmt.Errorf("create saved filter: %w", err)
	}
	row, err := db.QueryOneAs[savedFilterRecord](ctx, r.database,
		`SELECT * FROM saved_filter WHERE id = $id`, db.Vars{"id": id})
	if err != nil {
		return nil, fmt.Errorf("create saved filter: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("create saved filter: no record returned")
	}
	return row.toModel(), nil
}

// setIfString adds a non-empty string pointer to the content map, leaving the
// column NULL otherwise.
func setIfString(content map[string]any, key string, value *string) {
	if value != nil && *value != "" {
		content[key] = *value
	}
}

// DeleteSavedFilter removes a saved filter preset by id.
func (r *Repository) DeleteSavedFilter(ctx context.Context, filterID string) (bool, error) {
	id, err := db.ParseRecordID(filterID)
	if err != nil {
		return false, fmt.Errorf("invalid filter id: %w", err)
	}
	if err := r.database.Delete(ctx, *id); err != nil {
		return false, fmt.Errorf("delete saved filter: %w", err)
	}
	return true, nil
}
