package automations

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// scheduleInterval is how often RunSchedules checks the schedules; cron's
// finest step is a minute.
const scheduleInterval = time.Minute

// parseCron parses a standard five-field cron expression (minute hour
// day-of-month month day-of-week, or a descriptor like "@daily"). FireDue
// reads it in the zone of the time it is given: the server's local time.
func parseCron(expression string) (cron.Schedule, error) {
	schedule, err := cron.ParseStandard(expression)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression %q: %w", expression, err)
	}
	return schedule, nil
}

// RunSchedules fires due schedule automations every minute until ctx ends.
func (s *Service) RunSchedules(ctx context.Context) {
	ticker := time.NewTicker(scheduleInterval)
	defer ticker.Stop()
	for {
		if _, err := s.FireDue(ctx, s.now()); err != nil {
			log.Printf("automations: schedules: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// FireDue queues a run of every enabled schedule automation whose next
// time after its last firing has come, and returns how many it queued. A
// schedule missed while the server was down fires once on the next check.
func (s *Service) FireDue(ctx context.Context, now time.Time) (int, error) {
	automations, err := s.enabledWithTrigger(ctx, schema.TriggerKindSchedule)
	if err != nil {
		return 0, err
	}
	fired := 0
	for _, automation := range automations {
		due, err := s.isDue(ctx, automation, now)
		if err != nil || !due {
			continue
		}
		s.markScheduled(automation.Id, now)
		if _, err := s.Enqueue(ctx, Request{AutomationID: automation.Id, TriggerData: map[string]any{"scheduledAt": db.FormatTime(now)}}); err != nil {
			return fired, err
		}
		fired++
	}
	return fired, nil
}

// isDue reports whether a schedule automation's next time has come.
func (s *Service) isDue(ctx context.Context, automation schema.Automation, now time.Time) (bool, error) {
	if automation.Trigger.Cron == nil {
		return false, nil
	}
	schedule, err := parseCron(*automation.Trigger.Cron)
	if err != nil {
		return false, err
	}
	last, err := s.lastFired(ctx, automation)
	if err != nil {
		return false, err
	}
	// cron reads the expression in the zone of the time it is given.
	return !schedule.Next(last.In(now.Location())).After(now), nil
}

// lastFired is when an automation last fired: the latest of its last
// scheduled firing, its last run and its last edit, so enabling or
// rescheduling it doesn't fire it at once.
func (s *Service) lastFired(ctx context.Context, automation schema.Automation) (time.Time, error) {
	latest := parseTime(automation.UpdatedAt)
	s.scheduleMutex.Lock()
	scheduled := s.lastScheduled[automation.Id]
	s.scheduleMutex.Unlock()
	latest = later(latest, scheduled)
	started, err := s.database.Strings(ctx,
		`SELECT MAX("startedAt") FROM "AutomationRun" WHERE "automation" = $automation`,
		db.Vars{"automation": automation.Id})
	if err != nil || len(started) == 0 {
		return latest, err
	}
	return later(latest, parseTime(started[0])), nil
}

// markScheduled remembers that an automation fired at when.
func (s *Service) markScheduled(automationID string, when time.Time) {
	s.scheduleMutex.Lock()
	defer s.scheduleMutex.Unlock()
	s.lastScheduled[automationID] = when
}

// parseTime reads a stored timestamp; the zero time when it can't.
func parseTime(stored string) time.Time {
	parsed, err := time.Parse(db.TimeFormat, stored)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

// later returns the later of two times.
func later(first, second time.Time) time.Time {
	if second.After(first) {
		return second
	}
	return first
}
