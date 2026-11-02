// Package limits keeps the user's usage limits (watch time and sessions per
// day or week) and measures usage from the event log.
//
// Windows follow the server's local clock: a day starts at midnight, a week
// on Monday at midnight.
//
// Watch time: progress and heatSpan events are heartbeats of playback. Per
// sessionId, the time between two consecutive heartbeats counts when they
// are at most maxHeartbeatGap apart; a longer gap is a pause and doesn't
// count. Heartbeats without a session are ignored.
//
// Sessions: the distinct sessionIds of the window's view events.
package limits

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// limitTable is the table limits are stored in.
const limitTable = "Limit"

// maxHeartbeatGap is the longest gap between heartbeats still counted as
// watching.
const maxHeartbeatGap = 60 * time.Second

// Service stores limits and measures usage against them.
type Service struct {
	database *db.DB
}

// New returns a limit service on database.
func New(database *db.DB) *Service {
	return &Service{database: database}
}

// NewLimit is a limit to add. Amount is minutes for watch time, a count for
// sessions.
type NewLimit struct {
	Kind    schema.LimitKind
	Window  schema.LimitWindow
	Amount  int
	Action  schema.LimitAction
	Enabled bool
}

// LimitEdit changes a limit; nil fields stay as they are.
type LimitEdit struct {
	Amount  *int
	Action  *schema.LimitAction
	Enabled *bool
}

// Create adds a limit.
func (s *Service) Create(ctx context.Context, limit NewLimit) (*schema.Limit, error) {
	if limit.Amount <= 0 {
		return nil, fmt.Errorf("a limit's amount must be positive")
	}
	id, err := s.database.Insert(ctx, limitTable, schema.Limit{
		Kind:    limit.Kind,
		Window:  limit.Window,
		Amount:  int32(limit.Amount),
		Action:  limit.Action,
		Enabled: limit.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// Update applies an edit to a limit.
func (s *Service) Update(ctx context.Context, id string, edit LimitEdit) (*schema.Limit, error) {
	fields := db.Vars{}
	if edit.Amount != nil {
		if *edit.Amount <= 0 {
			return nil, fmt.Errorf("a limit's amount must be positive")
		}
		fields["amount"] = *edit.Amount
	}
	if edit.Action != nil {
		fields["action"] = *edit.Action
	}
	if edit.Enabled != nil {
		fields["enabled"] = *edit.Enabled
	}
	if err := s.database.Update(ctx, limitTable, id, fields); err != nil {
		return nil, err
	}
	return s.Get(ctx, id)
}

// Delete moves a limit to the trash.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.database.Trash(ctx, limitTable, id)
}

// Get loads one limit; nil when there is none.
func (s *Service) Get(ctx context.Context, id string) (*schema.Limit, error) {
	return db.Get[schema.Limit](ctx, s.database, limitTable, id)
}

// Limits lists every limit, oldest first.
func (s *Service) Limits(ctx context.Context) ([]schema.Limit, error) {
	return db.QueryAs[schema.Limit](ctx, s.database,
		`SELECT * FROM "Limit" WHERE "deletedAt" IS NULL ORDER BY "createdAt", rowid`, nil)
}

// LimitStatus is one enabled limit with its usage in the current window.
type LimitStatus struct {
	Limit schema.Limit
	// Used is minutes for watch time, a count for sessions.
	Used    float64
	Reached bool
}

// Status measures every enabled limit against the current window.
func (s *Service) Status(ctx context.Context) ([]LimitStatus, error) {
	return s.StatusAt(ctx, time.Now())
}

// StatusAt measures every enabled limit against the window holding now.
func (s *Service) StatusAt(ctx context.Context, now time.Time) ([]LimitStatus, error) {
	limits, err := db.QueryAs[schema.Limit](ctx, s.database,
		`SELECT * FROM "Limit" WHERE "deletedAt" IS NULL AND "enabled" = 1 ORDER BY "createdAt", rowid`, nil)
	if err != nil {
		return nil, err
	}
	usages := map[schema.LimitWindow]Usage{}
	statuses := make([]LimitStatus, 0, len(limits))
	for _, limit := range limits {
		usage, measured := usages[limit.Window]
		if !measured {
			usage, err = s.Usage(ctx, limit.Window, now)
			if err != nil {
				return nil, err
			}
			usages[limit.Window] = usage
		}
		used := usedBy(limit.Kind, usage)
		statuses = append(statuses, LimitStatus{Limit: limit, Used: used, Reached: used >= float64(limit.Amount)})
	}
	return statuses, nil
}

// usedBy picks the measure a limit kind is checked against.
func usedBy(kind schema.LimitKind, usage Usage) float64 {
	if kind == schema.LimitKindSessions {
		return float64(usage.Sessions)
	}
	return usage.WatchMinutes
}
