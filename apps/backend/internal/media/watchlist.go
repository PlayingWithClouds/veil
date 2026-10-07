package media

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// watchlistRecord mirrors the `watchlist` table. Maps to model.WatchlistItem.
type watchlistRecord struct {
	ID        *db.RecordID `json:"id,omitempty"`
	Media     *db.RecordID `json:"media,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
}

func (w watchlistRecord) ToModel() *model.WatchlistItem {
	return &model.WatchlistItem{
		ID:        recordIDString(w.ID),
		Media:     recordIDString(w.Media),
		CreatedAt: w.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// AddToWatchlist adds a media item to the watchlist, returning the existing
// entry if already present (idempotent).
func (r *Repository) AddToWatchlist(ctx context.Context, mediaID string) (*model.WatchlistItem, error) {
	media, err := parseMediaID(mediaID)
	if err != nil {
		return nil, err
	}
	_, err = r.database.Exec(ctx,
		`INSERT INTO watchlist (id, media) VALUES ($id, $media) ON CONFLICT (media) DO NOTHING`,
		db.Vars{"id": db.NewRecordID("watchlist"), "media": *media})
	if err != nil {
		return nil, fmt.Errorf("add to watchlist: %w", err)
	}
	row, err := db.QueryOneAs[watchlistRecord](ctx, r.database,
		`SELECT * FROM watchlist WHERE media = $media`, db.Vars{"media": *media})
	if err != nil {
		return nil, fmt.Errorf("get watchlist entry: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("add to watchlist: no record returned")
	}
	return row.ToModel(), nil
}

// RemoveFromWatchlist removes a media item from the watchlist.
func (r *Repository) RemoveFromWatchlist(ctx context.Context, mediaID string) (bool, error) {
	media, err := parseMediaID(mediaID)
	if err != nil {
		return false, err
	}
	_, err = r.database.Exec(ctx, `DELETE FROM watchlist WHERE media = $media`, db.Vars{"media": *media})
	if err != nil {
		return false, fmt.Errorf("remove from watchlist: %w", err)
	}
	return true, nil
}

// ListWatchlist returns the watchlist, newest first.
func (r *Repository) ListWatchlist(ctx context.Context) ([]*model.WatchlistItem, error) {
	rows, err := db.QueryAs[watchlistRecord](ctx, r.database,
		`SELECT * FROM watchlist ORDER BY created_at DESC`, nil)
	if err != nil {
		return nil, fmt.Errorf("list watchlist: %w", err)
	}
	out := make([]*model.WatchlistItem, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToModel())
	}
	return out, nil
}
