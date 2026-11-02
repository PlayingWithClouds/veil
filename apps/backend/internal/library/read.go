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

// NameFilter narrows a list of performers, studios or tags.
type NameFilter struct {
	Text   string
	Limit  int
	Offset int
}

// Videos lists a page of videos matching the query's filter, in its order.
func (l *Library) Videos(ctx context.Context, query MediaQuery) ([]schema.Video, error) {
	statement, vars := mediaQuery(schema.RecordKindVideo, query)
	return db.QueryAs[schema.Video](ctx, l.database, statement, vars)
}

// Video loads one video; nil when it doesn't exist or is trashed.
func (l *Library) Video(ctx context.Context, id string) (*schema.Video, error) {
	return db.Get[schema.Video](ctx, l.database, tableVideo, id)
}

// Galleries lists a page of galleries matching the query's filter.
func (l *Library) Galleries(ctx context.Context, query MediaQuery) ([]schema.Gallery, error) {
	statement, vars := mediaQuery(schema.RecordKindGallery, query)
	return db.QueryAs[schema.Gallery](ctx, l.database, statement, vars)
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
	query, vars := nameQuery(schema.RecordKindPerformer, filter)
	return db.QueryAs[schema.Performer](ctx, l.database, query, vars)
}

// Performer loads one performer.
func (l *Library) Performer(ctx context.Context, id string) (*schema.Performer, error) {
	return db.Get[schema.Performer](ctx, l.database, tablePerformer, id)
}

// Studios lists studios matching filter by name.
func (l *Library) Studios(ctx context.Context, filter NameFilter) ([]schema.Studio, error) {
	query, vars := nameQuery(schema.RecordKindStudio, filter)
	return db.QueryAs[schema.Studio](ctx, l.database, query, vars)
}

// Studio loads one studio.
func (l *Library) Studio(ctx context.Context, id string) (*schema.Studio, error) {
	return db.Get[schema.Studio](ctx, l.database, tableStudio, id)
}

// Tags lists tags matching filter by name.
func (l *Library) Tags(ctx context.Context, filter NameFilter) ([]schema.Tag, error) {
	query, vars := nameQuery(schema.RecordKindTag, filter)
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

// CountMedia counts the videos or galleries matching filter (blocked and
// trashed ones excluded).
func (l *Library) CountMedia(ctx context.Context, kind schema.RecordKind, filter schema.Filter) (int, error) {
	conditions, vars := filterConditions(kind, filter)
	return l.database.Int(ctx, `SELECT COUNT(*)`+mediaSelect(kind)+` WHERE `+strings.Join(conditions, " AND "), vars)
}

// TagCount is a tag and how many media of a filter's results carry it.
type TagCount struct {
	schema.Tag
	Count int `json:"count"`
}

// TagCounts lists the tags most common among the videos or galleries
// matching filter, most common first, at most limit. Tags the filter already
// requires are left out: every result carries them.
func (l *Library) TagCounts(ctx context.Context, kind schema.RecordKind, filter schema.Filter, limit int) ([]TagCount, error) {
	conditions, vars := filterConditions(kind, filter)
	conditions = append(conditions, `"tag"."deletedAt" IS NULL`,
		`"tag"."id" NOT IN (SELECT value FROM json_each($requiredTagIds))`)
	vars["requiredTagIds"] = nonNil(filter.IncludeTagIds)
	vars["limit"], _ = pageBounds(limit, 0)
	return db.QueryAs[TagCount](ctx, l.database, `SELECT "tag".*, COUNT(*) AS "count"`+mediaSelect(kind)+`
		JOIN json_each("media"."tags") AS "credit"
		JOIN "Tag" AS "tag" ON "tag"."id" = "credit".value
		WHERE `+strings.Join(conditions, " AND ")+`
		GROUP BY "tag"."id" ORDER BY "count" DESC, "tag"."name" LIMIT $limit`, vars)
}

// nonNil returns ids, or an empty list for nil, so it binds as a JSON array.
func nonNil(ids []string) []string {
	if ids == nil {
		return []string{}
	}
	return ids
}

// mediaQuery builds the SELECT for a filtered page of videos or galleries.
func mediaQuery(kind schema.RecordKind, query MediaQuery) (string, db.Vars) {
	conditions, vars := filterConditions(kind, query.Filter)
	vars["limit"], vars["offset"] = pageBounds(query.Limit, query.Offset)
	return `SELECT "media".*` + mediaSelect(kind) + ` WHERE ` + strings.Join(conditions, " AND ") +
		` ORDER BY ` + orderBy(kind, query.Filter.Sort) + ` LIMIT $limit OFFSET $offset`, vars
}

// nameQuery builds the SELECT for a page of performers, studios or tags by
// name. Blocked ones are left out.
func nameQuery(kind schema.RecordKind, filter NameFilter) (string, db.Vars) {
	table := tableFor(kind)
	conditions := []string{`"deletedAt" IS NULL`, `NOT EXISTS (SELECT 1 FROM "UserState" AS "blocked"
		WHERE "blocked"."subjectKind" = $kind AND "blocked"."subjectId" = ` + db.Quote(table) + `."id"
			AND "blocked"."blocked" = 1 AND "blocked"."deletedAt" IS NULL)`}
	vars := db.Vars{"kind": string(kind)}
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
