package media

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// userRatingRecord mirrors the `user_rating` table. Maps to model.UserRating.
type userRatingRecord struct {
	ID        *db.RecordID `json:"id,omitempty"`
	Media     *db.RecordID `json:"media,omitempty"`
	Rating    float64      `json:"rating"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

func (u userRatingRecord) ToModel() *model.UserRating {
	return &model.UserRating{
		ID:        recordIDString(u.ID),
		Media:     recordIDString(u.Media),
		Rating:    u.Rating,
		CreatedAt: u.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// UpsertUserRating creates or updates the rating for a media item.
func (r *Repository) UpsertUserRating(ctx context.Context, in model.UpsertUserRatingInput) (*model.UserRating, error) {
	media, err := parseMediaID(in.Media)
	if err != nil {
		return nil, err
	}
	_, err = r.database.Exec(ctx,
		`INSERT INTO user_rating (id, media, rating) VALUES ($id, $media, $rating)
			ON CONFLICT (media) DO UPDATE SET rating = excluded.rating`,
		db.Vars{"id": db.NewRecordID("user_rating"), "media": *media, "rating": in.Rating})
	if err != nil {
		return nil, fmt.Errorf("upsert user rating: %w", err)
	}
	row, err := r.getUserRatingRecord(ctx, *media)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, fmt.Errorf("upsert user rating: no record returned")
	}
	return row.ToModel(), nil
}

// UserRating returns the rating for a media item, or nil.
func (r *Repository) UserRating(ctx context.Context, mediaID string) (*model.UserRating, error) {
	media, err := parseMediaID(mediaID)
	if err != nil {
		return nil, err
	}
	rec, err := r.getUserRatingRecord(ctx, *media)
	if err != nil || rec == nil {
		return nil, err
	}
	return rec.ToModel(), nil
}

// ListUserRatings returns ratings ordered by most recently updated.
func (r *Repository) ListUserRatings(ctx context.Context, limit, offset *int) ([]*model.UserRating, error) {
	rows, err := db.QueryAs[userRatingRecord](ctx, r.database,
		`SELECT * FROM user_rating ORDER BY updated_at DESC LIMIT $limit OFFSET $offset`,
		db.Vars{"limit": pageLimit(limit), "offset": pageOffset(offset)})
	if err != nil {
		return nil, fmt.Errorf("list user ratings: %w", err)
	}
	out := make([]*model.UserRating, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.ToModel())
	}
	return out, nil
}

// DeleteUserRating removes a media item's rating. Returns true always.
func (r *Repository) DeleteUserRating(ctx context.Context, mediaID string) (bool, error) {
	media, err := parseMediaID(mediaID)
	if err != nil {
		return false, err
	}
	_, err = r.database.Exec(ctx, `DELETE FROM user_rating WHERE media = $media`, db.Vars{"media": *media})
	if err != nil {
		return false, fmt.Errorf("delete user rating: %w", err)
	}
	return true, nil
}

// ProfileRatings returns every rating keyed by media record id.
func (r *Repository) ProfileRatings(ctx context.Context) (map[string]float64, error) {
	rows, err := db.QueryAs[userRatingRecord](ctx, r.database, `SELECT * FROM user_rating`, nil)
	if err != nil {
		return nil, fmt.Errorf("profile ratings: %w", err)
	}
	out := make(map[string]float64, len(rows))
	for _, row := range rows {
		out[recordIDString(row.Media)] = row.Rating
	}
	return out, nil
}

func (r *Repository) getUserRatingRecord(ctx context.Context, media db.RecordID) (*userRatingRecord, error) {
	row, err := db.QueryOneAs[userRatingRecord](ctx, r.database,
		`SELECT * FROM user_rating WHERE media = $media`, db.Vars{"media": media})
	if err != nil {
		return nil, fmt.Errorf("get user rating: %w", err)
	}
	return row, nil
}
