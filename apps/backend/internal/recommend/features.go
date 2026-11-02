package recommend

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/schema"
)

// mediaFeatures is what ranking knows about a video or gallery.
type mediaFeatures struct {
	ref        schema.RecordRef
	title      string
	tags       []string
	performers []string
	studio     string
	// site is the plugin the item was first found on; empty for local files.
	site string
	// date is the release date as stored, publishedAt its parsed form (zero
	// when there is none: when it was stored says nothing about its age).
	date        string
	publishedAt time.Time
	// durationSeconds, width and height are known for videos only (0 when
	// unknown), imageCount for galleries.
	durationSeconds int
	width           int
	height          int
	imageCount      int
	// companion is the video's gallery or the gallery's video, when linked.
	companion string
}

// companionRef is the linked gallery of a video or video of a gallery.
func (features *mediaFeatures) companionRef() (schema.RecordRef, bool) {
	if features.companion == "" {
		return schema.RecordRef{}, false
	}
	if features.ref.Kind == schema.RecordKindVideo {
		return schema.RecordRef{Kind: schema.RecordKindGallery, Id: features.companion}, true
	}
	return schema.RecordRef{Kind: schema.RecordKindVideo, Id: features.companion}, true
}

// commonFeatureColumns are the feature columns both media tables have.
const commonFeatureColumns = `SELECT "media"."id", "media"."title", "media"."tags", "media"."performers",
	"media"."studio", "media"."date"`

// kindFeatureColumns are the feature columns of one media table.
var kindFeatureColumns = map[schema.RecordKind]string{
	schema.RecordKindVideo:   `, "media"."durationSeconds", "media"."width", "media"."height", "media"."gallery" AS "companion"`,
	schema.RecordKindGallery: `, "media"."imageCount", "media"."video" AS "companion"`,
}

// tableOf is the table of a media kind.
func tableOf(kind schema.RecordKind) string {
	if kind == schema.RecordKindGallery {
		return "Gallery"
	}
	return "Video"
}

// featureQuery is a query for the features of the media of kind matching
// filter (trashed and blocked media never match) and condition, followed by
// tail (ORDER BY, LIMIT). It returns the query and its parameters.
func featureQuery(kind schema.RecordKind, filter schema.Filter, condition, tail string) (string, db.Vars) {
	conditions, vars := library.FilterConditions(kind, filter)
	if condition != "" {
		conditions = append(conditions, condition)
	}
	query := commonFeatureColumns + kindFeatureColumns[kind] + library.MediaFrom(kind) +
		` WHERE ` + strings.Join(conditions, " AND ") + ` ` + tail
	return query, vars
}

// queryFeatures runs a featureQuery, binding extra parameters, and reads the
// sites of the items found. Rows come back in the query's order.
func (e *Engine) queryFeatures(ctx context.Context, kind schema.RecordKind, filter schema.Filter, condition, tail string, extra db.Vars) ([]*mediaFeatures, error) {
	query, vars := featureQuery(kind, filter, condition, tail)
	for name, value := range extra {
		vars[name] = value
	}
	rows, err := e.database.Query(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("load %s features: %w", kind, err)
	}
	out := make([]*mediaFeatures, 0, len(rows))
	for _, row := range rows {
		out = append(out, featuresFromRow(kind, row))
	}
	if err := e.loadSites(ctx, kind, out); err != nil {
		return nil, err
	}
	return out, nil
}

// featuresFromRow decodes one featureQuery row.
func featuresFromRow(kind schema.RecordKind, row db.Row) *mediaFeatures {
	return &mediaFeatures{
		ref:             schema.RecordRef{Kind: kind, Id: rowString(row, "id")},
		title:           rowString(row, "title"),
		tags:            rowStrings(row, "tags"),
		performers:      rowStrings(row, "performers"),
		studio:          rowString(row, "studio"),
		date:            rowString(row, "date"),
		publishedAt:     releaseDate(rowString(row, "date")),
		durationSeconds: int(rowFloat(row, "durationSeconds")),
		width:           rowInt(row, "width"),
		height:          rowInt(row, "height"),
		imageCount:      rowInt(row, "imageCount"),
		companion:       rowString(row, "companion"),
	}
}

// releaseDate parses the date part of a release date; zero when there is none.
func releaseDate(date string) time.Time {
	if len(date) < len("2006-01-02") {
		return time.Time{}
	}
	parsed, err := time.Parse("2006-01-02", date[:len("2006-01-02")])
	if err != nil {
		return time.Time{}
	}
	return parsed
}

// loadSites fills in the plugin each item was first found on. Source has no
// index on its subject, so this is one scan of it for the whole batch.
func (e *Engine) loadSites(ctx context.Context, kind schema.RecordKind, items []*mediaFeatures) error {
	if len(items) == 0 {
		return nil
	}
	byID := make(map[string]*mediaFeatures, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		byID[item.ref.Id] = item
		ids = append(ids, item.ref.Id)
	}
	// Newest first, so the earliest source is the one written last.
	rows, err := e.database.Query(ctx,
		`SELECT "subjectId", "plugin" FROM "Source"
		 WHERE "subjectKind" = $kind AND "deletedAt" IS NULL AND "subjectId" IN (SELECT value FROM json_each($ids))
		 ORDER BY "createdAt" DESC`,
		db.Vars{"kind": kind, "ids": ids})
	if err != nil {
		return fmt.Errorf("load sites: %w", err)
	}
	for _, row := range rows {
		byID[rowString(row, "subjectId")].site = rowString(row, "plugin")
	}
	return nil
}

// loadFeatures reads the features of the referenced videos and galleries,
// keyed by reference. Trashed and blocked items, and other kinds, are left out.
func (e *Engine) loadFeatures(ctx context.Context, refs []schema.RecordRef) (map[schema.RecordRef]*mediaFeatures, error) {
	idsByKind := map[schema.RecordKind][]string{}
	for _, ref := range refs {
		idsByKind[ref.Kind] = append(idsByKind[ref.Kind], ref.Id)
	}
	out := make(map[schema.RecordRef]*mediaFeatures, len(refs))
	for _, kind := range []schema.RecordKind{schema.RecordKindVideo, schema.RecordKindGallery} {
		if len(idsByKind[kind]) == 0 {
			continue
		}
		items, err := e.queryFeatures(ctx, kind, schema.Filter{}, `"media"."id" IN (SELECT value FROM json_each($ids))`, "", db.Vars{"ids": idsByKind[kind]})
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			out[item.ref] = item
		}
	}
	return out, nil
}

// entityNames reads the names of performers, studios and tags by id; trashed
// ones are left out.
func (e *Engine) entityNames(ctx context.Context, ids []string) (map[string]string, error) {
	out := map[string]string{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := e.database.Query(ctx,
		`SELECT "id", "name" FROM "Performer" WHERE "id" IN (SELECT value FROM json_each($ids)) AND "deletedAt" IS NULL
		 UNION ALL SELECT "id", "name" FROM "Studio" WHERE "id" IN (SELECT value FROM json_each($ids)) AND "deletedAt" IS NULL
		 UNION ALL SELECT "id", "name" FROM "Tag" WHERE "id" IN (SELECT value FROM json_each($ids)) AND "deletedAt" IS NULL`,
		db.Vars{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("entity names: %w", err)
	}
	for _, row := range rows {
		out[rowString(row, "id")] = rowString(row, "name")
	}
	return out, nil
}

// blockedEntities are the blocked performers, studios and tags.
type blockedEntities struct {
	performers map[string]bool
	studios    map[string]bool
	tags       map[string]bool
}

// loadBlocked reads the blocked performers, studios and tags.
func (e *Engine) loadBlocked(ctx context.Context) (*blockedEntities, error) {
	rows, err := e.database.Query(ctx,
		`SELECT "subjectKind", "subjectId" FROM "UserState"
		 WHERE "blocked" = 1 AND "deletedAt" IS NULL AND "subjectKind" IN ('performer', 'studio', 'tag')`, nil)
	if err != nil {
		return nil, fmt.Errorf("read blocked: %w", err)
	}
	blocked := &blockedEntities{performers: map[string]bool{}, studios: map[string]bool{}, tags: map[string]bool{}}
	for _, row := range rows {
		blocked.set(schema.RecordKind(rowString(row, "subjectKind")))[rowString(row, "subjectId")] = true
	}
	return blocked, nil
}

// set is the blocked set of one entity kind.
func (blocked *blockedEntities) set(kind schema.RecordKind) map[string]bool {
	switch kind {
	case schema.RecordKindPerformer:
		return blocked.performers
	case schema.RecordKindStudio:
		return blocked.studios
	}
	return blocked.tags
}
