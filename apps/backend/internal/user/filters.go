package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// savedFilterTable is the table saved filters live in.
const savedFilterTable = "SavedFilter"

// SavedFilters lists the saved filters, pinned first, then by name.
func (s *Service) SavedFilters(ctx context.Context) ([]schema.SavedFilter, error) {
	return db.QueryAs[schema.SavedFilter](ctx, s.database,
		`SELECT * FROM "SavedFilter" WHERE "deletedAt" IS NULL ORDER BY "pinned" DESC, "name" COLLATE NOCASE`, nil)
}

// SavedFilter loads one saved filter; nil when it doesn't exist.
func (s *Service) SavedFilter(ctx context.Context, id string) (*schema.SavedFilter, error) {
	return db.Get[schema.SavedFilter](ctx, s.database, savedFilterTable, id)
}

// SaveFilter stores a named filter: a new one when id is empty, else it
// replaces the saved filter with that id.
func (s *Service) SaveFilter(ctx context.Context, id, name string, filter schema.Filter, pinned bool) (*schema.SavedFilter, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("a saved filter needs a name")
	}
	if filter.Kinds == nil {
		filter.Kinds = []schema.MediaKind{}
	}
	fields := db.Vars{"name": name, "filter": filter, "pinned": pinned}
	if id == "" {
		created, err := s.database.Insert(ctx, savedFilterTable, fields)
		if err != nil {
			return nil, err
		}
		return s.SavedFilter(ctx, created)
	}
	existing, err := s.SavedFilter(ctx, id)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("no saved filter %s", id)
	}
	if err := s.database.Update(ctx, savedFilterTable, id, fields); err != nil {
		return nil, err
	}
	return s.SavedFilter(ctx, id)
}

// DeleteSavedFilter trashes a saved filter.
func (s *Service) DeleteSavedFilter(ctx context.Context, id string) error {
	return s.database.Trash(ctx, savedFilterTable, id)
}

// PurgeTrash removes saved filters trashed before cutoff for good.
func (s *Service) PurgeTrash(ctx context.Context, cutoff time.Time) (int, error) {
	purged, err := s.database.Exec(ctx, `DELETE FROM "SavedFilter" WHERE "deletedAt" IS NOT NULL AND "deletedAt" < $cutoff`,
		db.Vars{"cutoff": db.FormatTime(cutoff)})
	return int(purged), err
}
