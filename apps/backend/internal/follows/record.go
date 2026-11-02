package follows

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// followableTables are the record kinds a follow can target, with their
// tables.
var followableTables = map[schema.RecordKind]string{
	schema.RecordKindPerformer: "Performer",
	schema.RecordKindStudio:    "Studio",
	schema.RecordKindTag:       "Tag",
	schema.RecordKindCategory:  "Category",
	schema.RecordKindCamModel:  "CamModel",
}

// nameColumns are the columns that name a followable record.
var nameColumns = map[schema.RecordKind]string{
	schema.RecordKindCamModel: "displayName",
}

// FilterFor is the filter that follows a record: media crediting the
// performer or studio, carrying the tag, or in the category.
func FilterFor(ref schema.RecordRef) (schema.Filter, error) {
	ids := []string{ref.Id}
	switch ref.Kind {
	case schema.RecordKindPerformer:
		return schema.Filter{PerformerIds: ids}, nil
	case schema.RecordKindStudio:
		return schema.Filter{StudioIds: ids}, nil
	case schema.RecordKindTag:
		return schema.Filter{IncludeTagIds: ids}, nil
	case schema.RecordKindCategory:
		return schema.Filter{CategoryIds: ids}, nil
	case schema.RecordKindCamModel:
		// A room is watched for going live, not filtered for media.
		return schema.Filter{}, nil
	}
	return schema.Filter{}, fmt.Errorf("a %s can't be followed", ref.Kind)
}

// FollowOf returns the follow of a record, nil when it isn't followed.
func (s *Service) FollowOf(ctx context.Context, ref schema.RecordRef) (*schema.Follow, error) {
	return s.followOf(ctx, ref)
}

// withoutTarget drops the target's own condition from a follow's filter.
// Listing cards carry no credits, so media listed on the target's page
// can't match it until their page is read.
func withoutTarget(filter schema.Filter, kind schema.RecordKind) schema.Filter {
	switch kind {
	case schema.RecordKindPerformer:
		filter.PerformerIds = nil
	case schema.RecordKindStudio:
		filter.StudioIds = nil
	case schema.RecordKindTag:
		filter.IncludeTagIds = nil
	case schema.RecordKindCategory:
		filter.CategoryIds = nil
	}
	return filter
}

// FollowRecord follows a performer, studio, tag or category under its name,
// or returns the follow it already has.
func (s *Service) FollowRecord(ctx context.Context, ref schema.RecordRef, intervalHours int, notify bool) (*schema.Follow, error) {
	filter, err := FilterFor(ref)
	if err != nil {
		return nil, err
	}
	existing, err := s.followOf(ctx, ref)
	if err != nil || existing != nil {
		return existing, err
	}
	name, err := s.recordName(ctx, ref)
	if err != nil {
		return nil, err
	}
	return s.Follow(ctx, NewFollow{Name: name, Filter: filter, TargetKind: &ref.Kind, TargetID: &ref.Id, IntervalHours: intervalHours, Notify: notify})
}

// FollowSearch follows a search text across every site.
func (s *Service) FollowSearch(ctx context.Context, text string, intervalHours int, notify bool) (*schema.Follow, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, fmt.Errorf("a search follow needs a text")
	}
	return s.Follow(ctx, NewFollow{Name: trimmed, Filter: schema.Filter{Text: &trimmed}, IntervalHours: intervalHours, Notify: notify})
}

// followOf returns the follow targeting ref, or nil.
func (s *Service) followOf(ctx context.Context, ref schema.RecordRef) (*schema.Follow, error) {
	return db.QueryOneAs[schema.Follow](ctx, s.database,
		`SELECT * FROM "Follow" WHERE "targetKind" = $kind AND "targetId" = $id AND "deletedAt" IS NULL`,
		db.Vars{"kind": ref.Kind, "id": ref.Id})
}

// recordName loads the name of a followable record.
func (s *Service) recordName(ctx context.Context, ref schema.RecordRef) (string, error) {
	column, named := nameColumns[ref.Kind]
	if !named {
		column = "name"
	}
	names, err := s.database.Strings(ctx,
		`SELECT `+db.Quote(column)+` FROM `+db.Quote(followableTables[ref.Kind])+` WHERE "id" = $id AND "deletedAt" IS NULL`,
		db.Vars{"id": ref.Id})
	if err != nil {
		return "", err
	}
	if len(names) == 0 {
		return "", fmt.Errorf("no %s %s", ref.Kind, ref.Id)
	}
	return names[0], nil
}
