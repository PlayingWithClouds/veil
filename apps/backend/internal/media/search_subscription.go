package media

import (
	"context"
	"errors"
	"fmt"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// DefaultSearchSubscriptionHours is the re-run interval when none is given.
const DefaultSearchSubscriptionHours = 6

// ErrSearchSubscriptionNotFound is returned for an unknown subscription id.
var ErrSearchSubscriptionNotFound = errors.New("search subscription not found")

// SubscribeSearch saves query as a search subscription. created is false when
// the (normalized) query was already subscribed; the existing one is returned.
func (r *Repository) SubscribeSearch(ctx context.Context, query string, sources []string, intervalHours int) (subscription *model.SearchSubscription, created bool, err error) {
	normalized := normalizeQuery(query)
	if normalized == "" {
		return nil, false, fmt.Errorf("subscribe search: empty query")
	}
	if intervalHours <= 0 {
		intervalHours = DefaultSearchSubscriptionHours
	}
	if sources == nil {
		sources = []string{}
	}
	affected, err := r.database.Exec(ctx,
		`INSERT INTO search_subscription (id, query, normalized_query, sources, interval_hours)
		 VALUES ($id, $query, $normalized, $sources, $interval)
		 ON CONFLICT (normalized_query) WHERE kind = 'search' DO NOTHING`,
		db.Vars{
			"id":         db.NewRecordID("search_subscription"),
			"query":      query,
			"normalized": normalized,
			"sources":    sources,
			"interval":   intervalHours,
		})
	if err != nil {
		return nil, false, fmt.Errorf("subscribe search: %w", err)
	}
	subscription, err = r.searchSubscriptionWhere(ctx, "kind = 'search' AND normalized_query = $normalized", db.Vars{"normalized": normalized})
	return subscription, affected > 0, err
}

// SearchSubscription returns one subscription, or ErrSearchSubscriptionNotFound.
func (r *Repository) SearchSubscription(ctx context.Context, id string) (*model.SearchSubscription, error) {
	return r.searchSubscriptionWhere(ctx, "id = $id", db.Vars{"id": id})
}

// SearchSubscriptions lists every subscription, newest first.
func (r *Repository) SearchSubscriptions(ctx context.Context) ([]*model.SearchSubscription, error) {
	return r.searchSubscriptionsWhere(ctx, "1 = 1 ORDER BY created_at DESC", nil)
}

// DueSearchSubscriptions lists enabled subscriptions whose next run is due.
func (r *Repository) DueSearchSubscriptions(ctx context.Context) ([]*model.SearchSubscription, error) {
	return r.searchSubscriptionsWhere(ctx,
		"enabled = 1 AND (next_run_at IS NULL OR next_run_at <= $now) ORDER BY next_run_at",
		db.Vars{"now": db.Now()})
}

// UpdateSearchSubscription changes the interval and/or enabled flag. A new
// interval reschedules the next run relative to the last one.
func (r *Repository) UpdateSearchSubscription(ctx context.Context, id string, intervalHours *int, enabled *bool) (*model.SearchSubscription, error) {
	if intervalHours != nil && *intervalHours <= 0 {
		return nil, fmt.Errorf("interval must be at least one hour")
	}
	if intervalHours != nil {
		_, err := r.database.Exec(ctx,
			`UPDATE search_subscription
			    SET interval_hours = $interval,
			        next_run_at = CASE WHEN last_run_at IS NULL THEN next_run_at
			          ELSE strftime('%Y-%m-%dT%H:%M:%fZ', last_run_at, '+' || $interval || ' hours') END
			  WHERE id = $id`,
			db.Vars{"id": id, "interval": *intervalHours})
		if err != nil {
			return nil, fmt.Errorf("update search subscription: %w", err)
		}
	}
	if enabled != nil {
		if _, err := r.database.Exec(ctx, `UPDATE search_subscription SET enabled = $enabled WHERE id = $id`,
			db.Vars{"id": id, "enabled": *enabled}); err != nil {
			return nil, fmt.Errorf("update search subscription: %w", err)
		}
	}
	return r.SearchSubscription(ctx, id)
}

// DeleteSearchSubscription removes a subscription and its feed. Deleting an
// unknown id is not an error.
func (r *Repository) DeleteSearchSubscription(ctx context.Context, id string) error {
	return r.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM search_subscription_item WHERE subscription = $id`, db.Vars{"id": id}); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `DELETE FROM search_subscription WHERE id = $id`, db.Vars{"id": id})
		return err
	})
}

// RecordSearchSubscriptionRun adds a run's results to the feed and schedules
// the next run. The first run's results are marked seen, so only later
// arrivals count as new. runErr (may be nil) is kept as last_error.
func (r *Repository) RecordSearchSubscriptionRun(ctx context.Context, id string, mediaIDs []string, runErr error) error {
	var lastError any
	if runErr != nil {
		lastError = runErr.Error()
	}
	now := db.Now()
	return r.database.Tx(ctx, func(tx *db.DB) error {
		for _, mediaID := range mediaIDs {
			_, err := tx.Exec(ctx,
				`INSERT INTO search_subscription_item (id, subscription, media, first_seen_at)
				 VALUES ($id, $subscription, $media, $now)
				 ON CONFLICT (subscription, media) DO NOTHING`,
				db.Vars{"id": db.NewRecordID("search_subscription_item"), "subscription": id, "media": mediaID, "now": now})
			if err != nil {
				return fmt.Errorf("record search subscription item: %w", err)
			}
		}
		_, err := tx.Exec(ctx,
			`UPDATE search_subscription
			    SET last_run_at = $now,
			        next_run_at = strftime('%Y-%m-%dT%H:%M:%fZ', $now, '+' || interval_hours || ' hours'),
			        last_error = $lastError,
			        seen_at = ifnull(seen_at, $now)
			  WHERE id = $id`,
			db.Vars{"id": id, "now": now, "lastError": lastError})
		return err
	})
}

// MarkSearchSubscriptionSeen clears the subscription's new-scene count.
func (r *Repository) MarkSearchSubscriptionSeen(ctx context.Context, id string) (*model.SearchSubscription, error) {
	if _, err := r.database.Exec(ctx, `UPDATE search_subscription SET seen_at = $now WHERE id = $id`,
		db.Vars{"id": id, "now": db.Now()}); err != nil {
		return nil, fmt.Errorf("mark search subscription seen: %w", err)
	}
	return r.SearchSubscription(ctx, id)
}

// SearchSubscriptionSceneCounts counts the scenes in a subscription's feed
// and those first seen after it was last marked seen, skipping the same
// scenes SearchSubscriptionScenes hides.
func (r *Repository) SearchSubscriptionSceneCounts(ctx context.Context, id string, disabledPlugins []string) (total, unseen int, err error) {
	row, err := r.database.QueryRow(ctx,
		`SELECT count(*) AS total,
		        count(*) FILTER (WHERE item.first_seen_at > ifnull(subscription.seen_at, '')) AS unseen
		 FROM search_subscription_item item
		 JOIN search_subscription subscription ON subscription.id = item.subscription
		 JOIN scene ON scene.id = item.media
		 WHERE item.subscription = $subscription
		   AND scene.id NOT IN (SELECT value FROM json_each($hidden))`,
		db.Vars{"subscription": id, "hidden": r.excludedSceneIDs(ctx, EntityFilters{DisabledPlugins: disabledPlugins})})
	if err != nil || row == nil {
		return 0, 0, err
	}
	return db.AsInt(row["total"]), db.AsInt(row["unseen"]), nil
}

// SearchSubscriptionScenes pages through the scenes a subscription has found,
// newest first, hiding blocked scenes and those of inactive plugins.
func (r *Repository) SearchSubscriptionScenes(ctx context.Context, id string, limit, offset int, disabledPlugins []string) ([]*model.Scene, error) {
	if limit <= 0 {
		limit = 48
	}
	rows, err := r.database.Query(ctx,
		`SELECT scene.* FROM search_subscription_item item
		 JOIN scene ON scene.id = item.media
		 WHERE item.subscription = $subscription
		   AND scene.id NOT IN (SELECT value FROM json_each($hidden))
		 ORDER BY item.first_seen_at DESC, item.rowid
		 LIMIT $limit OFFSET $offset`,
		db.Vars{
			"subscription": id,
			"hidden":       r.excludedSceneIDs(ctx, EntityFilters{DisabledPlugins: disabledPlugins}),
			"limit":        limit,
			"offset":       offset,
		})
	if err != nil {
		return nil, fmt.Errorf("search subscription scenes: %w", err)
	}
	out := make([]*model.Scene, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.sceneFromMap(ctx, row))
	}
	return out, nil
}

// searchSubscriptionWhere returns the single subscription matching condition.
func (r *Repository) searchSubscriptionWhere(ctx context.Context, condition string, vars db.Vars) (*model.SearchSubscription, error) {
	subscriptions, err := r.searchSubscriptionsWhere(ctx, condition+" LIMIT 1", vars)
	if err != nil {
		return nil, err
	}
	if len(subscriptions) == 0 {
		return nil, ErrSearchSubscriptionNotFound
	}
	return subscriptions[0], nil
}

// searchSubscriptionsWhere selects subscriptions by a trailing SQL condition.
func (r *Repository) searchSubscriptionsWhere(ctx context.Context, condition string, vars db.Vars) ([]*model.SearchSubscription, error) {
	rows, err := r.database.Query(ctx,
		`SELECT * FROM search_subscription WHERE `+condition, vars)
	if err != nil {
		return nil, fmt.Errorf("search subscriptions: %w", err)
	}
	out := make([]*model.SearchSubscription, 0, len(rows))
	for _, row := range rows {
		subscription := searchSubscriptionFromRow(row)
		subscription.Target = r.subscriptionTarget(ctx, row)
		out = append(out, subscription)
	}
	return out, nil
}

// searchSubscriptionFromRow maps a subscription row to the GraphQL model,
// without its target.
func searchSubscriptionFromRow(row db.Row) *model.SearchSubscription {
	sources := []string{}
	if list, ok := row["sources"].([]any); ok {
		for _, source := range list {
			if name, ok := source.(string); ok {
				sources = append(sources, name)
			}
		}
	}
	return &model.SearchSubscription{
		ID:            rowString(row, "id"),
		Kind:          subscriptionKindFromColumn(rowString(row, "kind")),
		Query:         rowString(row, "query"),
		Sources:       sources,
		IntervalHours: db.AsInt(row["interval_hours"]),
		Enabled:       db.AsBool(row["enabled"]),
		LastRunAt:     rowStringPointer(row, "last_run_at"),
		NextRunAt:     rowStringPointer(row, "next_run_at"),
		LastError:     rowStringPointer(row, "last_error"),
		CreatedAt:     rowString(row, "created_at"),
	}
}

// rowString returns a text column, or "" when it is NULL.
func rowString(row db.Row, column string) string {
	value, _ := row[column].(string)
	return value
}

// rowStringPointer returns a text column, or nil when it is NULL or empty.
func rowStringPointer(row db.Row, column string) *string {
	value, ok := row[column].(string)
	if !ok || value == "" {
		return nil
	}
	return &value
}
