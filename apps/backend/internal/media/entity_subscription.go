package media

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// Entity subscriptions follow a studio (channel), performer or tag. They share
// the search_subscription table and feed with saved searches: kind names the
// followed record's table and entity holds its id.

// ErrSubscriptionTargetNotFound is returned when a followed record is unknown.
var ErrSubscriptionTargetNotFound = errors.New("subscription target not found")

// EntityOrigin says where a followed studio, performer or tag comes from, so
// the scheduler can ask the right plugins for its new scenes.
type EntityOrigin struct {
	Name      string
	SourceURL string
	// Observers are the plugins that ingested the record itself (from its own
	// page or a listing), most recent first.
	Observers []string
	// SceneSources are the plugins whose scenes credit the record, most scenes first.
	SceneSources []string
}

// SubscribeEntity follows a studio, performer or tag; kind must match the
// target's record type. created is false when the target was already
// followed; the existing subscription is returned.
func (r *Repository) SubscribeEntity(ctx context.Context, kind model.SubscriptionKind, targetID string, intervalHours int) (subscription *model.SearchSubscription, created bool, err error) {
	target, err := r.subscriptionTargetRecord(ctx, targetID)
	if err != nil {
		return nil, false, fmt.Errorf("subscribe: %w", err)
	}
	if target.Table != subscriptionKindColumn(kind) {
		return nil, false, fmt.Errorf("subscribe: %s is not a %s", targetID, strings.ToLower(string(kind)))
	}
	if intervalHours <= 0 {
		intervalHours = DefaultSearchSubscriptionHours
	}
	name := mString(target.Row, "name")
	affected, err := r.database.Exec(ctx,
		`INSERT INTO search_subscription (id, kind, entity, query, normalized_query, interval_hours)
		 VALUES ($id, $kind, $entity, $query, $normalized, $interval)
		 ON CONFLICT (entity) WHERE entity IS NOT NULL DO NOTHING`,
		db.Vars{
			"id":         db.NewRecordID("search_subscription"),
			"kind":       target.Table,
			"entity":     target.ID,
			"query":      name,
			"normalized": normalizeQuery(name),
			"interval":   intervalHours,
		})
	if err != nil {
		return nil, false, fmt.Errorf("subscribe: %w", err)
	}
	subscription, err = r.SubscriptionForTarget(ctx, target.ID)
	return subscription, affected > 0, err
}

// SubscriptionForTarget returns the subscription following a studio,
// performer or tag, or ErrSearchSubscriptionNotFound.
func (r *Repository) SubscriptionForTarget(ctx context.Context, targetID string) (*model.SearchSubscription, error) {
	return r.searchSubscriptionWhere(ctx, "entity = $entity", db.Vars{"entity": targetID})
}

// EntityOrigin loads a followed record's name, source URL and the plugins it
// and its credited scenes came from.
func (r *Repository) EntityOrigin(ctx context.Context, entityID string) (*EntityOrigin, error) {
	target, err := r.subscriptionTargetRecord(ctx, entityID)
	if err != nil {
		return nil, err
	}
	observers, err := r.database.Strings(ctx,
		`SELECT plugin FROM observation WHERE target = $entity
		 GROUP BY plugin ORDER BY max(observed_at) DESC`,
		db.Vars{"entity": target.ID})
	if err != nil {
		return nil, fmt.Errorf("entity origin: %w", err)
	}
	linked, _ := sceneLinkConditions(target.Table)
	sceneSources, err := r.database.Strings(ctx,
		`SELECT observation.plugin FROM scene
		 JOIN observation ON observation.target = scene.id
		 WHERE `+linked+`
		 GROUP BY observation.plugin ORDER BY count(DISTINCT scene.id) DESC`,
		db.Vars{"entity": target.ID})
	if err != nil {
		return nil, fmt.Errorf("entity origin: %w", err)
	}
	return &EntityOrigin{
		Name:         mString(target.Row, "name"),
		SourceURL:    mString(target.Row, "source_url"),
		Observers:    observers,
		SceneSources: sceneSources,
	}, nil
}

// LinkedSceneIDs lists the library scenes credited to a studio, performer or
// tag (directly, not inherited).
func (r *Repository) LinkedSceneIDs(ctx context.Context, entityID string) ([]string, error) {
	target, err := db.ParseRecordID(entityID)
	if err != nil {
		return nil, fmt.Errorf("linked scenes: %w", err)
	}
	linked, _ := sceneLinkConditions(target.Table)
	if linked == "" {
		return nil, fmt.Errorf("linked scenes: %s cannot be followed", target.Table)
	}
	return r.database.Strings(ctx, `SELECT scene.id FROM scene WHERE `+linked, db.Vars{"entity": target.String()})
}

// KeepLinkedOrUncredited filters mediaIDs to the scenes credited to the
// entity, plus scenes with no credit of that kind yet (search stubs carry
// no studio, performers or tags until their page is visited). Non-scene ids
// are dropped.
func (r *Repository) KeepLinkedOrUncredited(ctx context.Context, entityID string, mediaIDs []string) ([]string, error) {
	target, err := db.ParseRecordID(entityID)
	if err != nil {
		return nil, fmt.Errorf("filter linked scenes: %w", err)
	}
	linked, uncredited := sceneLinkConditions(target.Table)
	if linked == "" {
		return nil, fmt.Errorf("filter linked scenes: %s cannot be followed", target.Table)
	}
	if mediaIDs == nil {
		mediaIDs = []string{}
	}
	return r.database.Strings(ctx,
		`SELECT scene.id FROM scene
		 WHERE scene.id IN (SELECT value FROM json_each($ids))
		   AND (`+linked+` OR `+uncredited+`)`,
		db.Vars{"entity": target.String(), "ids": mediaIDs})
}

// subscriptionTargetRow is a followed record with its parsed id.
type subscriptionTargetRow struct {
	ID    string
	Table string
	Row   map[string]any
}

// subscriptionTargetRecord loads a studio, performer or tag by id.
func (r *Repository) subscriptionTargetRecord(ctx context.Context, targetID string) (*subscriptionTargetRow, error) {
	parsed, err := db.ParseRecordID(targetID)
	if err != nil {
		return nil, fmt.Errorf("invalid target id %q: %w", targetID, err)
	}
	if linked, _ := sceneLinkConditions(parsed.Table); linked == "" {
		return nil, fmt.Errorf("%s records cannot be followed", parsed.Table)
	}
	row, err := r.database.Get(ctx, *parsed)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, fmt.Errorf("%w: %s", ErrSubscriptionTargetNotFound, targetID)
	}
	return &subscriptionTargetRow{ID: parsed.String(), Table: parsed.Table, Row: row}, nil
}

// subscriptionTarget describes the record an entity subscription follows,
// nil for searches. A deleted record keeps the name stored at subscribe time.
func (r *Repository) subscriptionTarget(ctx context.Context, row db.Row) *model.SubscriptionTarget {
	entity := rowString(row, "entity")
	if entity == "" {
		return nil
	}
	target := &model.SubscriptionTarget{ID: entity, Name: rowString(row, "query")}
	record, err := r.getByID(ctx, entity)
	if err != nil || record == nil {
		return target
	}
	target.Name = mString(record, "name")
	target.ImageURL = mStringPtr(record, "image_path")
	return target
}

// sceneLinkConditions returns SQL conditions on scene for "credited to
// $entity" and "has no credit of this kind", by the entity's table. Both are
// empty for tables that cannot be followed.
func sceneLinkConditions(table string) (linked, uncredited string) {
	switch table {
	case "studio":
		return "scene.studio = $entity", "scene.studio IS NULL"
	case "performer":
		return jsonContains("scene.performers", "$entity"), "json_array_length(scene.performers) = 0"
	case "tag":
		return jsonContains("scene.tags", "$entity"), "json_array_length(scene.tags) = 0"
	}
	return "", ""
}

// subscriptionKindColumn maps a GraphQL kind to the kind column ("studio", ...),
// which for entity kinds is also the followed record's table.
func subscriptionKindColumn(kind model.SubscriptionKind) string {
	return strings.ToLower(string(kind))
}

// subscriptionKindFromColumn maps the kind column to the GraphQL kind; rows
// from before kinds existed are searches.
func subscriptionKindFromColumn(column string) model.SubscriptionKind {
	kind := model.SubscriptionKind(strings.ToUpper(column))
	if !kind.IsValid() {
		return model.SubscriptionKindSearch
	}
	return kind
}
