package media

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

func (r *Repository) imageFromMap(ctx context.Context, m map[string]any) *model.Image {
	image := &model.Image{
		ID:         mID(m),
		FilePath:   mString(m, "file_path"),
		Title:      mStringPtr(m, "title"),
		Details:    mStringPtr(m, "details"),
		Date:       mStringPtr(m, "date"),
		SourceURL:  mStringPtr(m, "source_url"),
		ExternalID: mStringPtr(m, "external_id"),
		Width:      mIntPtr(m, "width"),
		Height:     mIntPtr(m, "height"),
		Position:   mIntPtr(m, "position"),
		Rating:     mFloatPtr(m, "rating"),
		Organized:  mBool(m, "organized"),
		Studio:     r.studioLite(ctx, asRecordID(m["studio"])),
		Performers: r.performersByIDs(ctx, mRIDList(m, "performers")),
		Tags:       r.tagsByIDs(ctx, mRIDList(m, "tags")),
		TagMatch:   tagMatchFromMap(m),
		CreatedAt:  mString(m, "created_at"),
		UpdatedAt:  mString(m, "updated_at"),
	}
	if owner := asRecordID(m["media"]); owner != nil && owner.Table == "gallery" {
		ownerID := owner.String()
		image.GalleryID = &ownerID
	}
	return image
}

// ListImages lists content images (gallery pages and standalone images). Asset
// rows (posters, logos) never surface here. Tag filtering matches the image's
// own tags, its studio/performers, or its owning gallery; direct matches first.
func (r *Repository) ListImages(ctx context.Context, filters EntityFilters, galleryID *string, limit, offset *int) ([]*model.Image, error) {
	vars := db.Vars{"limit": pageLimit(limit), "offset": pageOffset(offset)}
	conditions := []string{"content = 1"}
	if filters.Search != nil && *filters.Search != "" {
		vars["q"] = *filters.Search
		conditions = append(conditions, searchCondition("title"))
	}
	if rid := parseFilterID(filters.StudioID); rid != nil {
		vars["studio"] = *rid
		conditions = append(conditions, "(image.studio = $studio OR "+owningGalleryMatches("owner.studio = $studio")+")")
	}
	if rid := parseFilterID(filters.PerformerID); rid != nil {
		vars["performer"] = *rid
		conditions = append(conditions, "("+jsonContains("image.performers", "$performer")+" OR "+
			owningGalleryMatches(jsonContains("owner.performers", "$performer"))+")")
	}
	if rid := parseFilterID(galleryID); rid != nil {
		vars["gallery"] = *rid
		conditions = append(conditions, "media = $gallery")
	}
	tagFiltered := false
	if rid := parseFilterID(filters.TagID); rid != nil {
		vars["tag"] = *rid
		tagFiltered = true
		conditions = append(conditions, "("+inheritedTagCondition("image")+" OR "+
			owningGalleryMatches(jsonContains("owner.tags", "$tag"))+")")
	}
	projection, order := "*", "created_at DESC"
	if tagFiltered {
		projection = inheritedTagProjection("image")
		order = "direct_match DESC, " + order
	}
	query := fmt.Sprintf("SELECT %s FROM image WHERE %s ORDER BY %s LIMIT $limit OFFSET $offset",
		projection, joinAnd(conditions), order)
	rows, err := r.database.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("list images: %w", err)
	}
	out := make([]*model.Image, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.imageFromMap(ctx, row))
	}
	return out, nil
}

// owningGalleryMatches matches images whose owning gallery (aliased owner)
// satisfies condition.
func owningGalleryMatches(condition string) string {
	return "EXISTS (SELECT 1 FROM gallery AS owner WHERE owner.id = image.media AND " + condition + ")"
}

func (r *Repository) GetImage(ctx context.Context, id string) (*model.Image, error) {
	m, err := r.getByID(ctx, id)
	if err != nil || m == nil {
		return nil, err
	}
	return r.imageFromMap(ctx, m), nil
}

// SetImageTags replaces a content image's tags with exactly the given tag ids.
func (r *Repository) SetImageTags(ctx context.Context, imageID string, tagIDs []string) (*model.Image, error) {
	if err := r.setTagsField(ctx, imageID, tagIDs); err != nil {
		return nil, err
	}
	return r.GetImage(ctx, imageID)
}
