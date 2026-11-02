package library

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/schema"
)

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
	if kind != schema.RecordKindVideo && kind != schema.RecordKindGallery {
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
