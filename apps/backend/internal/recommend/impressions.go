package recommend

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// ignoredImpression counts the showings of one scene that no click or
// playback followed.
type ignoredImpression struct {
	count     int
	lastShown time.Time
}

// ignoredImpressionQuery counts, per scene, the showings since $since that the
// user never engaged with afterwards: no logged click, first visit or playback.
const ignoredImpressionQuery = `SELECT impression.media, count(*) AS ignored, max(impression.shown_at) AS last_shown
	FROM recommendation_impression AS impression
	WHERE impression.kind = 'shown' AND impression.shown_at >= $since
	  AND NOT EXISTS (SELECT 1 FROM recommendation_impression AS click
	    WHERE click.media = impression.media AND click.kind = 'click' AND click.shown_at >= impression.shown_at)
	  AND NOT EXISTS (SELECT 1 FROM scene
	    WHERE scene.id = impression.media AND scene.detail_fetched_at >= impression.shown_at)
	  AND NOT EXISTS (SELECT 1 FROM watch_history
	    WHERE watch_history.media = impression.media AND watch_history.updated_at >= impression.shown_at)
	GROUP BY impression.media`

// RecordImpressions stores a batch of shown/clicked feed items and prunes rows
// past the retention window. Returns how many were stored.
func (e *Engine) RecordImpressions(ctx context.Context, impressions []Impression) (int, error) {
	if len(impressions) > ImpressionBatchLimit {
		return 0, fmt.Errorf("record impressions: at most %d per batch, got %d", ImpressionBatchLimit, len(impressions))
	}
	for _, impression := range impressions {
		if _, err := db.ParseRecordID(impression.SceneID); err != nil {
			return 0, fmt.Errorf("record impressions: invalid media id %q", impression.SceneID)
		}
	}
	now := e.now()
	err := e.database.Tx(ctx, func(tx *db.DB) error {
		for _, impression := range impressions {
			if err := insertImpression(ctx, tx, impression, now); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx, `DELETE FROM recommendation_impression WHERE shown_at < $cutoff`,
			db.Vars{"cutoff": db.FormatTime(now.Add(-impressionRetention))})
		return err
	})
	if err != nil {
		return 0, fmt.Errorf("record impressions: %w", err)
	}
	return len(impressions), nil
}

// insertImpression writes one impression row.
func insertImpression(ctx context.Context, tx *db.DB, impression Impression, now time.Time) error {
	kind := "shown"
	if impression.Clicked {
		kind = "click"
	}
	_, err := tx.Exec(ctx,
		`INSERT INTO recommendation_impression (id, media, kind, source, surface, position, shown_at)
		 VALUES ($id, $media, $kind, $source, $surface, $position, $shownAt)`,
		db.Vars{
			"id":       db.NewRecordID("recommendation_impression"),
			"media":    impression.SceneID,
			"kind":     kind,
			"source":   impression.Source,
			"surface":  impression.Surface,
			"position": impression.Position,
			"shownAt":  db.FormatTime(now),
		})
	return err
}

// ignoredImpressions reads the "shown but not clicked" counts per scene.
func (e *Engine) ignoredImpressions(ctx context.Context, now time.Time) (map[string]ignoredImpression, error) {
	rows, err := e.database.Query(ctx, ignoredImpressionQuery,
		db.Vars{"since": db.FormatTime(now.Add(-ignoredImpressionWindow))})
	if err != nil {
		return nil, fmt.Errorf("read ignored impressions: %w", err)
	}
	out := make(map[string]ignoredImpression, len(rows))
	for _, row := range rows {
		out[rowString(row, "media")] = ignoredImpression{count: rowInt(row, "ignored"), lastShown: rowTime(row, "last_shown")}
	}
	return out, nil
}
