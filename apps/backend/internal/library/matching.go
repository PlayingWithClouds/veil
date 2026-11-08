package library

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// searchedMedia are the kinds a search lists that a filter narrows.
var searchedMedia = []schema.RecordKind{
	schema.RecordKindVideo, schema.RecordKindGallery, schema.RecordKindAudio, schema.RecordKindStory,
}

// Matching keeps the referenced media that filter matches (and that aren't
// trashed or blocked), in their order. Other kinds of records pass through.
func (l *Library) Matching(ctx context.Context, refs []schema.RecordRef, filter schema.Filter) ([]schema.RecordRef, error) {
	idsByKind := map[schema.RecordKind][]string{}
	for _, ref := range refs {
		idsByKind[ref.Kind] = append(idsByKind[ref.Kind], ref.Id)
	}
	matched := map[schema.RecordRef]bool{}
	for kind, ids := range idsByKind {
		if err := l.matchKind(ctx, kind, ids, filter, matched); err != nil {
			return nil, err
		}
	}
	kept := make([]schema.RecordRef, 0, len(refs))
	for _, ref := range refs {
		if matched[ref] {
			kept = append(kept, ref)
		}
	}
	return kept, nil
}

// matchKind marks the ids of one kind that filter matches; a kind the
// filter doesn't apply to is marked whole.
func (l *Library) matchKind(ctx context.Context, kind schema.RecordKind, ids []string, filter schema.Filter, matched map[schema.RecordRef]bool) error {
	if !slices.Contains(searchedMedia, kind) {
		for _, id := range ids {
			matched[schema.RecordRef{Kind: kind, Id: id}] = true
		}
		return nil
	}
	conditions, vars := FilterConditions(kind, filter)
	vars["ids"] = ids
	found, err := l.database.Strings(ctx,
		`SELECT "media"."id"`+MediaFrom(kind)+` WHERE `+strings.Join(conditions, " AND ")+
			` AND "media"."id" IN (SELECT value FROM json_each($ids))`, vars)
	if err != nil {
		return fmt.Errorf("match %s: %w", kind, err)
	}
	for _, id := range found {
		matched[schema.RecordRef{Kind: kind, Id: id}] = true
	}
	return nil
}

// MatchingSearch keeps the referenced media whose text has every quoted
// phrase of the search query and none of its excluded words, in their order.
// Bare words aren't checked: the sites searched for them their own way.
// Other kinds of records pass through.
func (l *Library) MatchingSearch(ctx context.Context, refs []schema.RecordRef, query string) ([]schema.RecordRef, error) {
	idsByKind := map[schema.RecordKind][]string{}
	for _, ref := range refs {
		idsByKind[ref.Kind] = append(idsByKind[ref.Kind], ref.Id)
	}
	matched := map[schema.RecordRef]bool{}
	for kind, ids := range idsByKind {
		if err := l.matchSearchKind(ctx, kind, ids, query, matched); err != nil {
			return nil, err
		}
	}
	kept := make([]schema.RecordRef, 0, len(refs))
	for _, ref := range refs {
		if matched[ref] {
			kept = append(kept, ref)
		}
	}
	return kept, nil
}

// matchSearchKind marks the ids of one kind whose text the query's phrases
// and exclusions allow; a kind searches don't narrow is marked whole.
func (l *Library) matchSearchKind(ctx context.Context, kind schema.RecordKind, ids []string, query string, matched map[schema.RecordRef]bool) error {
	if !slices.Contains(searchedMedia, kind) {
		for _, id := range ids {
			matched[schema.RecordRef{Kind: kind, Id: id}] = true
		}
		return nil
	}
	found, err := l.database.Strings(ctx, `SELECT "media"."id" FROM `+db.Quote(tableFor(kind))+` AS "media"
		WHERE "media"."id" IN (SELECT value FROM json_each($ids)) AND veil_search(`+searchedText+`, $query, 0) = 1`,
		db.Vars{"ids": ids, "query": query})
	if err != nil {
		return fmt.Errorf("match search %s: %w", kind, err)
	}
	for _, id := range found {
		matched[schema.RecordRef{Kind: kind, Id: id}] = true
	}
	return nil
}
