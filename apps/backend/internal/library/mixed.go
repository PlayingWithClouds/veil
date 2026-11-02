package library

import (
	"context"
	"fmt"
	"math/rand/v2"
	"slices"
	"sort"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// mediaKinds maps the filter's media kinds to record kinds.
var mediaKinds = map[schema.MediaKind]schema.RecordKind{
	schema.MediaKindVideo:   schema.RecordKindVideo,
	schema.MediaKindGallery: schema.RecordKindGallery,
	schema.MediaKindAudio:   schema.RecordKindAudio,
	schema.MediaKindStory:   schema.RecordKindStory,
	schema.MediaKindPost:    schema.RecordKindPost,
}

// allMediaKinds is the order kinds are listed in when a filter names none.
var allMediaKinds = []schema.MediaKind{
	schema.MediaKindVideo, schema.MediaKindGallery, schema.MediaKindAudio, schema.MediaKindStory, schema.MediaKindPost,
}

// sortedRef is a record found for a mixed list with the value it sorts by.
type sortedRef struct {
	ref     schema.RecordRef
	sortKey any
	addedAt string
}

// Media lists a page of media of every kind the filter names (all when it
// names none), merged in the filter's order. Each kind is queried for the
// first offset+limit rows and the merge cut to the page.
func (l *Library) Media(ctx context.Context, query MediaQuery) ([]schema.RecordRef, error) {
	limit, offset := pageBounds(query.Limit, query.Offset)
	found := []sortedRef{}
	for _, mediaKind := range kindsOf(query.Filter) {
		refs, err := l.sortedRefs(ctx, mediaKinds[mediaKind], query.Filter, limit+offset)
		if err != nil {
			return nil, err
		}
		found = append(found, refs...)
	}
	orderRefs(found, query.Filter.Sort)
	return pageOf(found, limit, offset), nil
}

// CountAllMedia counts the media of every kind the filter names.
func (l *Library) CountAllMedia(ctx context.Context, filter schema.Filter) (int, error) {
	total := 0
	for _, mediaKind := range kindsOf(filter) {
		count, err := l.CountMedia(ctx, mediaKinds[mediaKind], filter)
		if err != nil {
			return 0, err
		}
		total += count
	}
	return total, nil
}

// kindsOf is the media kinds a filter asks for.
func kindsOf(filter schema.Filter) []schema.MediaKind {
	if len(filter.Kinds) == 0 {
		return allMediaKinds
	}
	return filter.Kinds
}

// sortedRefs lists up to limit records of kind matching filter, with their
// sort key.
func (l *Library) sortedRefs(ctx context.Context, kind schema.RecordKind, filter schema.Filter, limit int) ([]sortedRef, error) {
	conditions, vars := filterConditions(kind, filter)
	vars["limit"] = limit
	rows, err := l.database.Query(ctx,
		`SELECT "media"."id" AS "id", `+sortExpression(kind, filter.Sort)+` AS "sortKey", "media"."createdAt" AS "addedAt"`+
			mediaSelect(kind)+` WHERE `+strings.Join(conditions, " AND ")+
			` ORDER BY `+orderBy(kind, filter.Sort)+` LIMIT $limit`, vars)
	if err != nil {
		return nil, err
	}
	refs := make([]sortedRef, 0, len(rows))
	for _, row := range rows {
		refs = append(refs, sortedRef{
			ref:     schema.RecordRef{Kind: kind, Id: fmt.Sprint(row["id"])},
			sortKey: row["sortKey"],
			addedAt: fmt.Sprint(row["addedAt"]),
		})
	}
	return refs, nil
}

// sortExpression is the value a kind's rows are merged by.
func sortExpression(kind schema.RecordKind, order *schema.Sort) string {
	if order == nil || order.Field == schema.SortFieldRandom {
		return `"media"."createdAt"`
	}
	if order.Field == schema.SortFieldDuration {
		if hasDuration(kind) {
			return `"media"."durationSeconds"`
		}
		return `NULL`
	}
	column, known := sortColumns[order.Field]
	if !known {
		return `"media"."createdAt"`
	}
	return column
}

// orderRefs sorts merged records like each kind's query did: missing values
// last, then by the sort direction (newest added first by default).
func orderRefs(refs []sortedRef, order *schema.Sort) {
	if order != nil && order.Field == schema.SortFieldRandom {
		rand.Shuffle(len(refs), func(first, second int) { refs[first], refs[second] = refs[second], refs[first] })
		return
	}
	descending := order == nil || order.Descending
	sort.SliceStable(refs, func(first, second int) bool {
		return sortsBefore(refs[first], refs[second], descending)
	})
}

// sortsBefore reports whether first comes before second.
func sortsBefore(first, second sortedRef, descending bool) bool {
	if (first.sortKey == nil) != (second.sortKey == nil) {
		return second.sortKey == nil
	}
	comparison := compareValues(first.sortKey, second.sortKey)
	if comparison == 0 {
		return first.addedAt > second.addedAt
	}
	if descending {
		return comparison > 0
	}
	return comparison < 0
}

// compareValues compares two sort keys: numbers as numbers, the rest as
// case-insensitive text.
func compareValues(first, second any) int {
	firstNumber, firstIsNumber := asFloat(first)
	secondNumber, secondIsNumber := asFloat(second)
	if firstIsNumber && secondIsNumber {
		switch {
		case firstNumber < secondNumber:
			return -1
		case firstNumber > secondNumber:
			return 1
		}
		return 0
	}
	return strings.Compare(strings.ToLower(fmt.Sprint(first)), strings.ToLower(fmt.Sprint(second)))
}

// asFloat reads a numeric value SQLite returned.
func asFloat(value any) (float64, bool) {
	switch typed := value.(type) {
	case int64:
		return float64(typed), true
	case float64:
		return typed, true
	}
	return 0, false
}

// pageOf cuts one page out of merged records.
func pageOf(refs []sortedRef, limit, offset int) []schema.RecordRef {
	page := []schema.RecordRef{}
	for index := offset; index < len(refs) && index < offset+limit; index++ {
		page = append(page, refs[index].ref)
	}
	return page
}

// MediaRecordKinds are the record kinds of media.
var MediaRecordKinds = []schema.RecordKind{
	schema.RecordKindVideo, schema.RecordKindGallery, schema.RecordKindAudio, schema.RecordKindStory, schema.RecordKindPost,
}

// IsMedia reports whether kind is a kind of media.
func IsMedia(kind schema.RecordKind) bool {
	return slices.Contains(MediaRecordKinds, kind)
}

// Exists reports whether the record kind/id is stored and not trashed.
func (l *Library) Exists(ctx context.Context, kind schema.RecordKind, id string) (bool, error) {
	table := tableFor(kind)
	if table == "" {
		return false, fmt.Errorf("unknown record kind %q", kind)
	}
	count, err := l.database.Int(ctx, `SELECT COUNT(*) FROM `+db.Quote(table)+` WHERE "id" = $id AND "deletedAt" IS NULL`, db.Vars{"id": id})
	return count > 0, err
}
