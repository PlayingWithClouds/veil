package media

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

func (r *Repository) galleryFromMap(ctx context.Context, m map[string]any) *model.Gallery {
	return &model.Gallery{
		ID:         mID(m),
		ExternalID: mString(m, "external_id"),
		SourceURL:  mString(m, "source_url"),
		Title:      mString(m, "title"),
		Details:    mStringPtr(m, "details"),
		Date:       mStringPtr(m, "date"),
		CoverPath:  mStringPtr(m, "cover_path"),
		ImageCount: mInt(m, "image_count"),
		Rating:     mFloatPtr(m, "rating"),
		Organized:  mBool(m, "organized"),
		Studio:     r.studioLite(ctx, asRecordID(m["studio"])),
		Performers: r.performersByIDs(ctx, mRIDList(m, "performers")),
		Tags:       r.tagsByIDs(ctx, mRIDList(m, "tags")),
		Images:     r.galleryImages(ctx, asRecordID(m["id"])),
		TagMatch:   tagMatchFromMap(m),
	}
}

func (r *Repository) galleryImages(ctx context.Context, galleryID *db.RecordID) []*model.Image {
	out := []*model.Image{}
	if galleryID == nil {
		return out
	}
	rows, err := r.database.Query(ctx,
		"SELECT * FROM image WHERE media = $id AND content = 1 ORDER BY position ASC",
		db.Vars{"id": *galleryID})
	if err != nil {
		return out
	}
	for _, row := range rows {
		out = append(out, r.imageFromMap(ctx, row))
	}
	return out
}

func (r *Repository) ListGalleries(ctx context.Context, filters EntityFilters, limit, offset *int) ([]*model.Gallery, error) {
	vars := db.Vars{"limit": pageLimit(limit), "offset": pageOffset(offset)}
	var conditions []string
	if filters.Search != nil && *filters.Search != "" {
		vars["q"] = *filters.Search
		conditions = append(conditions, searchCondition("title"))
	}
	if rid := parseFilterID(filters.StudioID); rid != nil {
		vars["studio"] = *rid
		conditions = append(conditions, "studio = $studio")
	}
	if rid := parseFilterID(filters.PerformerID); rid != nil {
		vars["performer"] = *rid
		conditions = append(conditions, jsonContains("gallery.performers", "$performer"))
	}
	tagFiltered := false
	if rid := parseFilterID(filters.TagID); rid != nil {
		vars["tag"] = *rid
		tagFiltered = true
		conditions = append(conditions, inheritedTagCondition("gallery"))
	}
	if len(filters.Sources) > 0 {
		vars["sources"] = filters.Sources
		conditions = append(conditions, sourceCondition("gallery"))
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + joinAnd(conditions)
	}
	projection, order := "*", "created_at DESC"
	if tagFiltered {
		projection = inheritedTagProjection("gallery")
		order = "direct_match DESC, " + order
	}
	query := fmt.Sprintf("SELECT %s FROM gallery %s ORDER BY %s LIMIT $limit OFFSET $offset", projection, where, order)
	rows, err := r.database.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("list galleries: %w", err)
	}
	out := make([]*model.Gallery, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.galleryFromMap(ctx, row))
	}
	return out, nil
}

func (r *Repository) GetGallery(ctx context.Context, id string) (*model.Gallery, error) {
	m, err := r.getByID(ctx, id)
	if err != nil || m == nil {
		return nil, err
	}
	return r.galleryFromMap(ctx, m), nil
}
