package media

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// subscriptionFeedQuery merges the subscriptions' items into one row per
// scene: a scene found by several subscriptions belongs to the first finder,
// and is new only if it is new there. Scenes found in the same run are
// ordered by release date.
const subscriptionFeedQuery = `
SELECT scene.*, feed.found_at AS feed_found_at, feed.subscription AS feed_subscription, feed.is_new AS feed_is_new
FROM (
  SELECT item.media, item.first_seen_at AS found_at, item.subscription,
         item.first_seen_at > ifnull(subscription.seen_at, '') AS is_new,
         row_number() OVER (PARTITION BY item.media ORDER BY item.first_seen_at, item.rowid) AS finder_rank
  FROM search_subscription_item item
  JOIN search_subscription subscription ON subscription.id = item.subscription
  WHERE ($subscription = '' OR subscription.id = $subscription)
    AND (json_array_length($kinds) = 0 OR subscription.kind IN (SELECT value FROM json_each($kinds)))
) feed
JOIN scene ON scene.id = feed.media
WHERE feed.finder_rank = 1
  AND scene.id NOT IN (SELECT value FROM json_each($hidden))
  AND ($newOnly = 0 OR feed.is_new)
  AND ($unwatchedOnly = 0 OR NOT EXISTS (SELECT 1 FROM watch_history history WHERE history.media = scene.id))
ORDER BY feed.found_at DESC, ifnull(scene.date, '') DESC, scene.id
LIMIT $limit OFFSET $offset`

// SubscriptionFeed pages through the scenes the subscriptions have found,
// newest find first, hiding blocked scenes and those of inactive plugins.
func (r *Repository) SubscriptionFeed(ctx context.Context, filter *model.SubscriptionFeedFilter, limit, offset int, disabledPlugins []string) ([]*model.SubscriptionFeedItem, error) {
	if limit <= 0 {
		limit = 48
	}
	vars := subscriptionFeedVars(filter)
	vars["hidden"] = r.excludedSceneIDs(ctx, EntityFilters{DisabledPlugins: disabledPlugins})
	vars["limit"] = limit
	vars["offset"] = offset
	rows, err := r.database.Query(ctx, subscriptionFeedQuery, vars)
	if err != nil {
		return nil, fmt.Errorf("subscription feed: %w", err)
	}
	items := make([]*model.SubscriptionFeedItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &model.SubscriptionFeedItem{
			Scene:          r.sceneFromMap(ctx, row),
			FoundAt:        rowString(row, "feed_found_at"),
			SubscriptionID: rowString(row, "feed_subscription"),
			IsNew:          db.AsBool(row["feed_is_new"]),
		})
	}
	return items, nil
}

// subscriptionFeedVars binds the feed filter; a nil filter matches everything.
func subscriptionFeedVars(filter *model.SubscriptionFeedFilter) db.Vars {
	vars := db.Vars{"subscription": "", "kinds": []string{}, "newOnly": false, "unwatchedOnly": false}
	if filter == nil {
		return vars
	}
	if filter.SubscriptionID != nil {
		vars["subscription"] = *filter.SubscriptionID
	}
	kinds := make([]string, 0, len(filter.Kinds))
	for _, kind := range filter.Kinds {
		kinds = append(kinds, subscriptionKindColumn(kind))
	}
	vars["kinds"] = kinds
	vars["newOnly"] = filter.NewOnly != nil && *filter.NewOnly
	vars["unwatchedOnly"] = filter.UnwatchedOnly != nil && *filter.UnwatchedOnly
	return vars
}
