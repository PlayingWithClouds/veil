package media

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// watchHistoryRecord mirrors the `watch_history` table. Maps to model.WatchHistory.
type watchHistoryRecord struct {
	ID              *db.RecordID `json:"id,omitempty"`
	Media           *db.RecordID `json:"media,omitempty"`
	StartedAt       time.Time    `json:"started_at"`
	FinishedAt      *time.Time   `json:"finished_at,omitempty"`
	ProgressSeconds int          `json:"progress_seconds"`
	// MaxProgressSeconds is the furthest position reached; seeking back
	// doesn't lower it.
	MaxProgressSeconds int       `json:"max_progress_seconds"`
	DurationSeconds    *int      `json:"duration_seconds,omitempty"`
	Completed          bool      `json:"completed"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ToModel converts the row to its GraphQL shape.
func (w watchHistoryRecord) ToModel() *model.WatchHistory {
	return &model.WatchHistory{
		ID:                 recordIDString(w.ID),
		Media:              recordIDString(w.Media),
		StartedAt:          w.StartedAt.UTC().Format(time.RFC3339),
		FinishedAt:         timePtrString(w.FinishedAt),
		ProgressSeconds:    w.ProgressSeconds,
		MaxProgressSeconds: w.MaxProgressSeconds,
		DurationSeconds:    w.DurationSeconds,
		Completed:          w.Completed,
		UpdatedAt:          w.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// UpsertWatchHistory creates or updates the single row per media item so
// playback position is preserved across sessions.
func (r *Repository) UpsertWatchHistory(ctx context.Context, in model.UpsertWatchHistoryInput) (*model.WatchHistory, error) {
	media, err := parseMediaID(in.Media)
	if err != nil {
		return nil, err
	}

	completed := false
	if in.Completed != nil {
		completed = *in.Completed
	}
	fields := map[string]any{
		"progress_seconds": in.ProgressSeconds,
		"duration_seconds": in.DurationSeconds,
		"completed":        completed,
	}
	if completed {
		fields["finished_at"] = db.Now()
	}

	id, err := r.saveWatchHistory(ctx, *media, fields)
	if err != nil {
		return nil, fmt.Errorf("upsert watch history: %w", err)
	}
	if err := r.raiseMaxProgress(ctx, id, in.ProgressSeconds); err != nil {
		return nil, fmt.Errorf("upsert watch history: %w", err)
	}
	row, err := db.QueryOneAs[watchHistoryRecord](ctx, r.database,
		`SELECT * FROM watch_history WHERE id = $id`, db.Vars{"id": id})
	if err != nil {
		return nil, fmt.Errorf("upsert watch history: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("upsert watch history: no record returned")
	}
	return row.ToModel(), nil
}

// saveWatchHistory merges fields into the media item's existing row, or
// inserts a new one, and returns the row id.
func (r *Repository) saveWatchHistory(ctx context.Context, media db.RecordID, fields map[string]any) (db.RecordID, error) {
	existing, err := r.getWatchHistoryRecord(ctx, media)
	if err != nil {
		return db.RecordID{}, err
	}
	if existing != nil && existing.ID != nil {
		return *existing.ID, r.database.Merge(ctx, *existing.ID, fields)
	}
	fields["media"] = media
	return r.database.Insert(ctx, "watch_history", fields)
}

// raiseMaxProgress moves the row's furthest-reached position up to progress,
// never down.
func (r *Repository) raiseMaxProgress(ctx context.Context, id db.RecordID, progress int) error {
	_, err := r.database.Exec(ctx,
		`UPDATE watch_history SET max_progress_seconds = max(max_progress_seconds, $progress) WHERE id = $id`,
		db.Vars{"id": id, "progress": progress})
	return err
}

// ListWatchHistory returns the history ordered by most recently updated.
func (r *Repository) ListWatchHistory(ctx context.Context, limit, offset *int) ([]*model.WatchHistory, error) {
	rows, err := db.QueryAs[watchHistoryRecord](ctx, r.database,
		`SELECT * FROM watch_history ORDER BY updated_at DESC LIMIT $limit OFFSET $offset`,
		db.Vars{"limit": pageLimit(limit), "offset": pageOffset(offset)})
	if err != nil {
		return nil, fmt.Errorf("list watch history: %w", err)
	}
	return watchHistoryModels(rows), nil
}

// WatchHistoryEntry returns the entry for a media item, or nil.
func (r *Repository) WatchHistoryEntry(ctx context.Context, mediaID string) (*model.WatchHistory, error) {
	media, err := parseMediaID(mediaID)
	if err != nil {
		return nil, err
	}
	rec, err := r.getWatchHistoryRecord(ctx, *media)
	if err != nil || rec == nil {
		return nil, err
	}
	return rec.ToModel(), nil
}

// DeleteWatchHistory removes a media item's entry. Returns true always.
func (r *Repository) DeleteWatchHistory(ctx context.Context, mediaID string) (bool, error) {
	media, err := parseMediaID(mediaID)
	if err != nil {
		return false, err
	}
	_, err = r.database.Exec(ctx, `DELETE FROM watch_history WHERE media = $media`, db.Vars{"media": *media})
	if err != nil {
		return false, fmt.Errorf("delete watch history: %w", err)
	}
	return true, nil
}

// ContinueWatching returns in-progress (not completed) entries, newest first.
// mediaType filters by the media record's table: "movie" or, for series,
// "episode".
func (r *Repository) ContinueWatching(ctx context.Context, mediaType *string, limit *int) ([]*model.WatchHistory, error) {
	vars := db.Vars{"limit": pageLimit(limit)}
	query := `SELECT * FROM watch_history WHERE completed = 0 ORDER BY updated_at DESC LIMIT $limit`
	if table := continueWatchingTable(mediaType); table != "" {
		vars["table"] = table
		query = `SELECT * FROM watch_history WHERE completed = 0 AND media LIKE $table || ':%' ORDER BY updated_at DESC LIMIT $limit`
	}
	rows, err := db.QueryAs[watchHistoryRecord](ctx, r.database, query, vars)
	if err != nil {
		return nil, fmt.Errorf("continue watching: %w", err)
	}
	return watchHistoryModels(rows), nil
}

// WatchedMediaIDs returns the set of media record ids with any history.
func (r *Repository) WatchedMediaIDs(ctx context.Context) (map[string]bool, error) {
	mediaIDs, err := r.database.Strings(ctx, `SELECT media FROM watch_history`, nil)
	if err != nil {
		return nil, fmt.Errorf("watched media ids: %w", err)
	}
	out := make(map[string]bool, len(mediaIDs))
	for _, mediaID := range mediaIDs {
		out[mediaID] = true
	}
	return out, nil
}

func (r *Repository) getWatchHistoryRecord(ctx context.Context, media db.RecordID) (*watchHistoryRecord, error) {
	row, err := db.QueryOneAs[watchHistoryRecord](ctx, r.database,
		`SELECT * FROM watch_history WHERE media = $media LIMIT 1`, db.Vars{"media": media})
	if err != nil {
		return nil, fmt.Errorf("get watch history: %w", err)
	}
	return row, nil
}

func watchHistoryModels(rows []watchHistoryRecord) []*model.WatchHistory {
	out := make([]*model.WatchHistory, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToModel())
	}
	return out
}

func continueWatchingTable(mediaType *string) string {
	if mediaType == nil {
		return ""
	}
	switch *mediaType {
	case "movie", "movies":
		return "movie"
	case "series", "episode", "episodes":
		return "episode"
	default:
		return ""
	}
}
