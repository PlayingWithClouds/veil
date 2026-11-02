package limits

import (
	"context"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Usage is what the user did in one window.
type Usage struct {
	Since        time.Time
	Until        time.Time
	WatchMinutes float64
	Sessions     int
}

// Usage measures watch time and sessions in the window holding now.
func (s *Service) Usage(ctx context.Context, window schema.LimitWindow, now time.Time) (Usage, error) {
	since := windowStart(window, now)
	usage := Usage{Since: since, Until: now}
	vars := db.Vars{"since": db.FormatTime(since), "until": db.FormatTime(now)}
	watched, err := s.watchedTime(ctx, vars)
	if err != nil {
		return usage, err
	}
	usage.WatchMinutes = watched.Minutes()
	usage.Sessions, err = s.database.Int(ctx,
		`SELECT COUNT(DISTINCT "sessionId") FROM "Event" WHERE "kind" = $view AND "sessionId" IS NOT NULL
		 AND "at" >= $since AND "at" < $until AND "deletedAt" IS NULL`,
		db.Vars{"view": schema.EventKindView, "since": vars["since"], "until": vars["until"]})
	return usage, err
}

// windowStart is the start of the day or week holding now, in now's zone.
func windowStart(window schema.LimitWindow, now time.Time) time.Time {
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if window != schema.LimitWindowWeek {
		return midnight
	}
	daysSinceMonday := (int(now.Weekday()) + 6) % 7
	return midnight.AddDate(0, 0, -daysSinceMonday)
}

// heartbeat is one playback event of a session.
type heartbeat struct {
	SessionID string `json:"sessionId"`
	At        string `json:"at"`
}

// watchedTime sums the counted gaps between heartbeats per session (see the
// package doc).
func (s *Service) watchedTime(ctx context.Context, vars db.Vars) (time.Duration, error) {
	vars["progress"] = schema.EventKindProgress
	vars["heatSpan"] = schema.EventKindHeatSpan
	heartbeats, err := db.QueryAs[heartbeat](ctx, s.database,
		`SELECT "sessionId", "at" FROM "Event" WHERE "kind" IN ($progress, $heatSpan) AND "sessionId" IS NOT NULL
		 AND "at" >= $since AND "at" < $until AND "deletedAt" IS NULL ORDER BY "sessionId", "at"`, vars)
	if err != nil {
		return 0, err
	}
	var total time.Duration
	for index := 1; index < len(heartbeats); index++ {
		total += countedGap(heartbeats[index-1], heartbeats[index])
	}
	return total, nil
}

// countedGap is the watch time between two consecutive heartbeats: their
// distance when they share a session and are close enough, else nothing.
func countedGap(previous, current heartbeat) time.Duration {
	if previous.SessionID != current.SessionID {
		return 0
	}
	previousAt, previousErr := time.Parse(db.TimeFormat, previous.At)
	currentAt, currentErr := time.Parse(db.TimeFormat, current.At)
	if previousErr != nil || currentErr != nil {
		return 0
	}
	gap := currentAt.Sub(previousAt)
	if gap > maxHeartbeatGap {
		return 0
	}
	return gap
}
