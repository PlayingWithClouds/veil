package recommend

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// signal is one weighted user interaction with a scene.
type signal struct {
	sceneID string
	weight  float64
	at      time.Time
	// disliked marks an explicit thumbs-down, which also excludes the scene.
	disliked bool
	// weak marks a signal that says nothing about the scene's content (shown
	// but not clicked), so the runtime preference leaves it out.
	weak bool
	// sessionEnd marks a watch nothing else followed. stopSeconds is where
	// its playback was left and durationSeconds the runtime (0 when unknown).
	sessionEnd      bool
	stopSeconds     int
	durationSeconds int
}

// Queries returning (media, at) rows since $since; every row gets the same weight.
const (
	oEventQuery = `SELECT media, created_at AS at FROM o_event
		WHERE media LIKE 'scene:%' AND created_at >= $since`
	saveQuery = `SELECT media, created_at AS at FROM watchlist
		WHERE media LIKE 'scene:%' AND created_at >= $since
		UNION ALL
		SELECT collection_item.media, collection_item.added_at FROM collection_item
		JOIN collection ON collection.id = collection_item.collection
		WHERE collection.user_created = 1 AND collection_item.media LIKE 'scene:%'
		  AND collection_item.added_at >= $since`
	preferenceChosenQuery = `SELECT chosen AS media, created_at AS at FROM preference_event
		WHERE chosen LIKE 'scene:%' AND created_at >= $since`
	preferenceRejectedQuery = `SELECT rejected AS media, created_at AS at FROM preference_event
		WHERE rejected LIKE 'scene:%' AND created_at >= $since`
	// A click is a first visit (detail fetch) or a click logged from a
	// recommendation surface; one per scene.
	clickQuery = `SELECT media, max(at) AS at FROM (
		SELECT id AS media, detail_fetched_at AS at FROM scene WHERE detail_fetched_at >= $since
		UNION ALL
		SELECT media, shown_at FROM recommendation_impression WHERE kind = 'click' AND shown_at >= $since
	) GROUP BY media`
)

// ratingQuery reads the user's scene ratings since $since.
const ratingQuery = `SELECT media, rating, updated_at AS at FROM user_rating
	WHERE media LIKE 'scene:%' AND updated_at >= $since`

// watchQuery reads scene watches since $since, with the scene's runtime as the
// fallback duration.
const watchQuery = `SELECT watch_history.media, watch_history.played_seconds AS played,
	watch_history.progress_seconds AS stop,
	coalesce(watch_history.duration_seconds, scene.duration_seconds) AS duration,
	watch_history.started_at AS started, watch_history.updated_at AS at
	FROM watch_history LEFT JOIN scene ON scene.id = watch_history.media
	WHERE watch_history.media LIKE 'scene:%' AND watch_history.updated_at >= $since`

// collectSignals reads every scene interaction inside the signal window and
// converts each into a weighted signal. ignored holds the "shown but not
// clicked" counts, already read for fatigue.
func (e *Engine) collectSignals(ctx context.Context, now time.Time, ignored map[string]ignoredImpression) ([]signal, error) {
	since := db.FormatTime(now.Add(-signalWindow))
	fixedWeights := []struct {
		query  string
		weight float64
	}{
		{oEventQuery, weightOEvent},
		{saveQuery, weightSave},
		{preferenceChosenQuery, weightPreferenceChosen},
		{preferenceRejectedQuery, weightPreferenceRejected},
		{clickQuery, weightClick},
	}
	signals := ignoredImpressionSignals(ignored)
	for _, source := range fixedWeights {
		batch, err := e.fixedWeightSignals(ctx, source.query, since, source.weight)
		if err != nil {
			return nil, err
		}
		signals = append(signals, batch...)
	}
	ratings, err := e.ratingSignals(ctx, since)
	if err != nil {
		return nil, err
	}
	watches, err := e.watchSignals(ctx, since, now)
	if err != nil {
		return nil, err
	}
	signals = append(signals, ratings...)
	return append(signals, watches...), nil
}

// fixedWeightSignals runs a (media, at) query and gives every row the same weight.
func (e *Engine) fixedWeightSignals(ctx context.Context, query, since string, weight float64) ([]signal, error) {
	rows, err := e.database.Query(ctx, query, db.Vars{"since": since})
	if err != nil {
		return nil, fmt.Errorf("read signals: %w", err)
	}
	out := make([]signal, 0, len(rows))
	for _, row := range rows {
		out = append(out, signal{sceneID: rowString(row, "media"), weight: weight, at: rowTime(row, "at")})
	}
	return out, nil
}

// ratingSignals turns thumbs up/down ratings into like/dislike signals.
func (e *Engine) ratingSignals(ctx context.Context, since string) ([]signal, error) {
	rows, err := e.database.Query(ctx, ratingQuery, db.Vars{"since": since})
	if err != nil {
		return nil, fmt.Errorf("read rating signals: %w", err)
	}
	out := make([]signal, 0, len(rows))
	for _, row := range rows {
		rating := rowFloat(row, "rating")
		if rating == ratingNeutral {
			continue
		}
		entry := signal{sceneID: rowString(row, "media"), weight: weightLike, at: rowTime(row, "at")}
		if rating < ratingNeutral {
			entry.weight = weightDislike
			entry.disliked = true
		}
		out = append(out, entry)
	}
	return out, nil
}

// watchSignals turns watch history rows into session-ending, continued or
// abandoned watch signals.
func (e *Engine) watchSignals(ctx context.Context, since string, now time.Time) ([]signal, error) {
	rows, err := e.database.Query(ctx, watchQuery, db.Vars{"since": since})
	if err != nil {
		return nil, fmt.Errorf("read watch signals: %w", err)
	}
	watches := make([]watch, 0, len(rows))
	for _, row := range rows {
		watches = append(watches, watch{
			sceneID:         rowString(row, "media"),
			playedSeconds:   rowInt(row, "played"),
			stopSeconds:     rowInt(row, "stop"),
			durationSeconds: rowInt(row, "duration"),
			startedAt:       rowTime(row, "started"),
			lastActivityAt:  rowTime(row, "at"),
		})
	}
	followed := followedWatches(watches)
	out := make([]signal, 0, len(watches))
	for index, entry := range watches {
		sinceUpdate := now.Sub(entry.lastActivityAt)
		weight, counts := watchWeight(entry, followed[index], sinceUpdate)
		if !counts {
			continue
		}
		out = append(out, signal{
			sceneID:         entry.sceneID,
			weight:          weight,
			at:              entry.lastActivityAt,
			sessionEnd:      endsSession(entry, followed[index], sinceUpdate),
			stopSeconds:     entry.stopSeconds,
			durationSeconds: entry.durationSeconds,
		})
	}
	return out, nil
}

// watchWeight classifies one watch. A watch that ended the session (nothing
// else played within sessionGap) is the strongest watch signal, finished or
// not: the video did its job. A watch another video followed counts for little,
// more when it was watched in full. counts is false for a fresh playback that
// may still be starting. A watch whose session is still open counts as
// followed until sessionGap has passed.
func watchWeight(entry watch, followed bool, sinceUpdate time.Duration) (weight float64, counts bool) {
	if isEarlyAbandon(entry.playedSeconds, entry.durationSeconds) {
		if sinceUpdate < abandonGrace {
			return 0, false
		}
		return weightEarlyAbandon, true
	}
	if endsSession(entry, followed, sinceUpdate) {
		return weightSessionEnd, true
	}
	if watchedInFull(entry.playedSeconds, entry.durationSeconds) {
		return weightCompletedWatch, true
	}
	return weightPartialWatch, true
}

// endsSession reports whether the watch was the last of its session: nothing
// else played within sessionGap after it, it played long enough to be more
// than a look, and it was not an early abandonment.
func endsSession(entry watch, followed bool, sinceUpdate time.Duration) bool {
	if isEarlyAbandon(entry.playedSeconds, entry.durationSeconds) {
		return false
	}
	return !followed && sinceUpdate >= sessionGap && entry.playedSeconds >= sessionEndMinimumSeconds
}

// watchedFraction is the share of the runtime played; 0 when the runtime is
// unknown.
func watchedFraction(playedSeconds, durationSeconds int) float64 {
	if durationSeconds <= 0 {
		return 0
	}
	return float64(playedSeconds) / float64(durationSeconds)
}

// watchedInFull reports whether a scene played fullWatchSeconds or
// completedFraction of its runtime: a skimmer who spent five minutes on a long
// scene has watched it as much as one who played a short one to the end.
func watchedInFull(playedSeconds, durationSeconds int) bool {
	if playedSeconds >= fullWatchSeconds {
		return true
	}
	return watchedFraction(playedSeconds, durationSeconds) >= completedFraction
}

// isEarlyAbandon reports whether a scene played under earlyAbandonSeconds, or
// under earlyAbandonFraction of its runtime while also under
// earlyAbandonCeiling: a minute spent on a long scene is a look, not a reject.
func isEarlyAbandon(playedSeconds, durationSeconds int) bool {
	if playedSeconds < earlyAbandonSeconds {
		return true
	}
	if playedSeconds >= earlyAbandonCeiling {
		return false
	}
	return durationSeconds > 0 && watchedFraction(playedSeconds, durationSeconds) < earlyAbandonFraction
}

// ignoredImpressionSignals turns "shown but not clicked" counts into weak
// negative signals, capped per scene.
func ignoredImpressionSignals(ignored map[string]ignoredImpression) []signal {
	out := make([]signal, 0, len(ignored))
	for sceneID, impression := range ignored {
		count := min(impression.count, ignoredImpressionCap)
		out = append(out, signal{sceneID: sceneID, weight: weightIgnoredImpression * float64(count), at: impression.lastShown, weak: true})
	}
	return out
}
