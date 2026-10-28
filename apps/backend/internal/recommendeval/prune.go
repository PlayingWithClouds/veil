package recommendeval

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// pruneStatements rewrite the database to what was known at $cutoff. Order
// matters: the scenes first visited after the cutoff lose their tags and
// performers before their visit stamp is cleared, because a stub carries
// neither (they come with the detail page) and keeping them would make
// exactly the scenes the user went on to open look richer than the rest.
var pruneStatements = []string{
	// Taste and engagement events after the cutoff have not happened yet.
	`DELETE FROM o_event WHERE created_at > $cutoff`,
	`DELETE FROM user_rating WHERE created_at > $cutoff`,
	`DELETE FROM watchlist WHERE created_at > $cutoff`,
	`DELETE FROM collection_item WHERE added_at > $cutoff`,
	`DELETE FROM preference_event WHERE created_at > $cutoff`,
	`DELETE FROM recommendation_impression WHERE shown_at > $cutoff`,
	`DELETE FROM search_history WHERE created_at > $cutoff`,
	`DELETE FROM search_result WHERE created_at > $cutoff`,
	// A watch under way at the cutoff keeps what could have been played by
	// then (never more than the wall clock allows) and counts as last played
	// at the cutoff; a watch begun later is dropped.
	`DELETE FROM watch_history WHERE started_at > $cutoff`,
	`UPDATE watch_history SET
		max_progress_seconds = max(0, min(max_progress_seconds, cast((julianday($cutoff) - julianday(started_at)) * 86400 AS INTEGER))),
		progress_seconds = max(0, min(progress_seconds, cast((julianday($cutoff) - julianday(started_at)) * 86400 AS INTEGER))),
		played_seconds = max(0, min(played_seconds, cast((julianday($cutoff) - julianday(started_at)) * 86400 AS INTEGER))),
		completed = 0, finished_at = NULL, updated_at = $cutoff
		WHERE updated_at > $cutoff`,
	// What the sites and the feeder had stored by then.
	`DELETE FROM feed_discovery WHERE found_at > $cutoff`,
	`DELETE FROM search_subscription_item WHERE first_seen_at > $cutoff`,
	`UPDATE search_subscription SET seen_at = $cutoff WHERE seen_at > $cutoff`,
	`DELETE FROM scene_related WHERE created_at > $cutoff`,
	`DELETE FROM scene_related WHERE scene IN (SELECT id FROM scene WHERE created_at > $cutoff)
		OR related IN (SELECT id FROM scene WHERE created_at > $cutoff)`,
	`DELETE FROM scene WHERE created_at > $cutoff`,
	`UPDATE scene SET tags = '[]', performers = '[]' WHERE detail_fetched_at > $cutoff`,
	`UPDATE scene SET detail_fetched_at = NULL WHERE detail_fetched_at > $cutoff`,
}

// pruneTo rewrites the database to its state at cutoff. It only deletes, so
// calling it with earlier and earlier cutoffs on the same database works.
func pruneTo(ctx context.Context, database *db.DB, cutoff time.Time) error {
	for _, statement := range pruneStatements {
		if _, err := database.Exec(ctx, statement, db.Vars{"cutoff": db.FormatTime(cutoff)}); err != nil {
			return fmt.Errorf("prune to %s: %w", db.FormatTime(cutoff), err)
		}
	}
	return nil
}
