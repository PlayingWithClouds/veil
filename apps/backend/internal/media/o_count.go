package media

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
)

// IncrementOCount records one o-event for a scene and returns the new total.
// History is preserved (one row per event), so the count is a COUNT().
func (r *Repository) IncrementOCount(ctx context.Context, mediaID string) (int, error) {
	media, err := parseMediaID(mediaID)
	if err != nil {
		return 0, err
	}
	if _, err := r.database.Insert(ctx, "o_event", map[string]any{"media": *media}); err != nil {
		return 0, fmt.Errorf("increment o count: %w", err)
	}
	return r.oCount(ctx, *media)
}

// DecrementOCount removes the most recent o-event for a scene and returns the
// new total. A no-op (returns 0) when there is nothing to undo.
func (r *Repository) DecrementOCount(ctx context.Context, mediaID string) (int, error) {
	media, err := parseMediaID(mediaID)
	if err != nil {
		return 0, err
	}
	latest, err := r.database.Strings(ctx,
		`SELECT id FROM o_event WHERE media = $media ORDER BY created_at DESC, rowid DESC LIMIT 1`,
		db.Vars{"media": *media})
	if err != nil {
		return 0, fmt.Errorf("find latest o event: %w", err)
	}
	if len(latest) == 0 {
		return 0, nil
	}
	if _, err := r.database.Exec(ctx, `DELETE FROM o_event WHERE id = $id`, db.Vars{"id": latest[0]}); err != nil {
		return 0, fmt.Errorf("decrement o count: %w", err)
	}
	return r.oCount(ctx, *media)
}

// OCount returns the o-count for a scene.
func (r *Repository) OCount(ctx context.Context, mediaID string) (int, error) {
	media, err := parseMediaID(mediaID)
	if err != nil {
		return 0, err
	}
	return r.oCount(ctx, *media)
}

// ProfileOCounts returns every o-count keyed by media record id.
func (r *Repository) ProfileOCounts(ctx context.Context) (map[string]int, error) {
	rows, err := r.database.Query(ctx,
		`SELECT media, count(*) AS count FROM o_event GROUP BY media`, nil)
	if err != nil {
		return nil, fmt.Errorf("profile o counts: %w", err)
	}
	out := make(map[string]int, len(rows))
	for _, row := range rows {
		out[mString(row, "media")] = mInt(row, "count")
	}
	return out, nil
}

func (r *Repository) oCount(ctx context.Context, media db.RecordID) (int, error) {
	count, err := r.database.Int(ctx,
		`SELECT count(*) FROM o_event WHERE media = $media`,
		db.Vars{"media": media})
	if err != nil {
		return 0, fmt.Errorf("o count: %w", err)
	}
	return count, nil
}

// parseMediaID parses a media record id argument.
func parseMediaID(mediaID string) (*db.RecordID, error) {
	media, err := db.ParseRecordID(mediaID)
	if err != nil {
		return nil, fmt.Errorf("invalid media id: %w", err)
	}
	return media, nil
}
