package library

import (
	"context"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// defaultLimit and maxLimit bound a page of results.
const (
	defaultLimit = 40
	maxLimit     = 200
)

// MediaFilter narrows a list of videos or galleries. Empty fields don't filter.
type MediaFilter struct {
	Text        string
	PerformerID string
	StudioID    string
	TagID       string
	Limit       int
	Offset      int
}

// NameFilter narrows a list of performers, studios or tags.
type NameFilter struct {
	Text   string
	Limit  int
	Offset int
}

// Videos lists videos matching filter, newest first.
func (l *Library) Videos(ctx context.Context, filter MediaFilter) ([]schema.Video, error) {
	query, vars := mediaQuery(tableVideo, filter)
	return db.QueryAs[schema.Video](ctx, l.database, query, vars)
}

// Video loads one video; nil when it doesn't exist or is trashed.
func (l *Library) Video(ctx context.Context, id string) (*schema.Video, error) {
	return db.Get[schema.Video](ctx, l.database, tableVideo, id)
}

// Galleries lists galleries matching filter, newest first.
func (l *Library) Galleries(ctx context.Context, filter MediaFilter) ([]schema.Gallery, error) {
	query, vars := mediaQuery(tableGallery, filter)
	return db.QueryAs[schema.Gallery](ctx, l.database, query, vars)
}

// Gallery loads one gallery.
func (l *Library) Gallery(ctx context.Context, id string) (*schema.Gallery, error) {
	return db.Get[schema.Gallery](ctx, l.database, tableGallery, id)
}

// Images lists a gallery's images in order.
func (l *Library) Images(ctx context.Context, galleryID string) ([]schema.Image, error) {
	return db.QueryAs[schema.Image](ctx, l.database,
		`SELECT * FROM "Image" WHERE "gallery" = $gallery AND "deletedAt" IS NULL ORDER BY "position"`,
		db.Vars{"gallery": galleryID})
}

// Performers lists performers matching filter by name.
func (l *Library) Performers(ctx context.Context, filter NameFilter) ([]schema.Performer, error) {
	query, vars := nameQuery(tablePerformer, filter)
	return db.QueryAs[schema.Performer](ctx, l.database, query, vars)
}

// Performer loads one performer.
func (l *Library) Performer(ctx context.Context, id string) (*schema.Performer, error) {
	return db.Get[schema.Performer](ctx, l.database, tablePerformer, id)
}

// Studios lists studios matching filter by name.
func (l *Library) Studios(ctx context.Context, filter NameFilter) ([]schema.Studio, error) {
	query, vars := nameQuery(tableStudio, filter)
	return db.QueryAs[schema.Studio](ctx, l.database, query, vars)
}

// Studio loads one studio.
func (l *Library) Studio(ctx context.Context, id string) (*schema.Studio, error) {
	return db.Get[schema.Studio](ctx, l.database, tableStudio, id)
}

// Tags lists tags matching filter by name.
func (l *Library) Tags(ctx context.Context, filter NameFilter) ([]schema.Tag, error) {
	query, vars := nameQuery(tableTag, filter)
	return db.QueryAs[schema.Tag](ctx, l.database, query, vars)
}

// Tag loads one tag.
func (l *Library) Tag(ctx context.Context, id string) (*schema.Tag, error) {
	return db.Get[schema.Tag](ctx, l.database, tableTag, id)
}

// PerformersByID loads performers by id, in the order given; missing ones
// are skipped.
func (l *Library) PerformersByID(ctx context.Context, ids []string) ([]schema.Performer, error) {
	found, err := db.QueryAs[schema.Performer](ctx, l.database,
		`SELECT * FROM "Performer" WHERE "id" IN (SELECT value FROM json_each($ids)) AND "deletedAt" IS NULL`,
		db.Vars{"ids": ids})
	if err != nil {
		return nil, err
	}
	return inOrder(found, ids, func(performer schema.Performer) string { return performer.Id }), nil
}

// TagsByID loads tags by id, in the order given.
func (l *Library) TagsByID(ctx context.Context, ids []string) ([]schema.Tag, error) {
	found, err := db.QueryAs[schema.Tag](ctx, l.database,
		`SELECT * FROM "Tag" WHERE "id" IN (SELECT value FROM json_each($ids)) AND "deletedAt" IS NULL`,
		db.Vars{"ids": ids})
	if err != nil {
		return nil, err
	}
	return inOrder(found, ids, func(tag schema.Tag) string { return tag.Id }), nil
}

// CountMedia counts the live videos or galleries a performer, studio or tag
// appears in. column is "performers", "studio" or "tags".
func (l *Library) CountMedia(ctx context.Context, kind schema.RecordKind, column, id string) (int, error) {
	filter := MediaFilter{}
	switch column {
	case "performers":
		filter.PerformerID = id
	case "studio":
		filter.StudioID = id
	case "tags":
		filter.TagID = id
	}
	conditions, vars := mediaConditions(filter)
	return l.database.Int(ctx, `SELECT COUNT(*) FROM `+db.Quote(tableFor(kind))+` WHERE `+strings.Join(conditions, " AND "), vars)
}

// mediaQuery builds the SELECT for a filtered page of videos or galleries.
func mediaQuery(table string, filter MediaFilter) (string, db.Vars) {
	conditions, vars := mediaConditions(filter)
	vars["limit"], vars["offset"] = pageBounds(filter.Limit, filter.Offset)
	return `SELECT * FROM ` + db.Quote(table) + ` WHERE ` + strings.Join(conditions, " AND ") +
		` ORDER BY "createdAt" DESC LIMIT $limit OFFSET $offset`, vars
}

// mediaConditions turns a media filter into WHERE conditions. Performers and
// tags are JSON id lists, matched with json_each.
func mediaConditions(filter MediaFilter) ([]string, db.Vars) {
	conditions := []string{`"deletedAt" IS NULL`}
	vars := db.Vars{}
	if filter.Text != "" {
		conditions = append(conditions, `"title" LIKE $text`)
		vars["text"] = "%" + filter.Text + "%"
	}
	if filter.PerformerID != "" {
		conditions = append(conditions, `EXISTS (SELECT 1 FROM json_each("performers") WHERE value = $performer)`)
		vars["performer"] = filter.PerformerID
	}
	if filter.TagID != "" {
		conditions = append(conditions, `EXISTS (SELECT 1 FROM json_each("tags") WHERE value = $tag)`)
		vars["tag"] = filter.TagID
	}
	if filter.StudioID != "" {
		conditions = append(conditions, `"studio" = $studio`)
		vars["studio"] = filter.StudioID
	}
	return conditions, vars
}

// nameQuery builds the SELECT for a page of performers, studios or tags by name.
func nameQuery(table string, filter NameFilter) (string, db.Vars) {
	conditions := []string{`"deletedAt" IS NULL`}
	vars := db.Vars{}
	if filter.Text != "" {
		conditions = append(conditions, `("name" LIKE $text OR EXISTS (SELECT 1 FROM json_each("aliases") WHERE value LIKE $text))`)
		vars["text"] = "%" + filter.Text + "%"
	}
	vars["limit"], vars["offset"] = pageBounds(filter.Limit, filter.Offset)
	return `SELECT * FROM ` + db.Quote(table) + ` WHERE ` + strings.Join(conditions, " AND ") +
		` ORDER BY "name" COLLATE NOCASE LIMIT $limit OFFSET $offset`, vars
}

// pageBounds clamps a requested page to the allowed size.
func pageBounds(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// inOrder arranges found records in the order of ids.
func inOrder[T any](found []T, ids []string, idOf func(T) string) []T {
	byID := make(map[string]T, len(found))
	for _, record := range found {
		byID[idOf(record)] = record
	}
	ordered := make([]T, 0, len(found))
	for _, id := range ids {
		record, ok := byID[id]
		if ok {
			ordered = append(ordered, record)
		}
	}
	return ordered
}
