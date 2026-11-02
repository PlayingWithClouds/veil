package user

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// eventTable is the table events are logged in.
const eventTable = "Event"

// NewEvent is an event to log. Subject is optional (a search has none).
type NewEvent struct {
	Kind        schema.EventKind
	SubjectKind *schema.RecordKind
	SubjectID   *string
	// Data is kind-specific: a progress or finish event carries the position
	// ({"position": seconds}), a search its text ({"text": "…"}).
	Data      map[string]any
	SessionID *string
}

// Record logs an event and applies it to the subject's state: a view counts
// and stamps the last view, progress moves the resume point, finishing
// clears it.
func (s *Service) Record(ctx context.Context, event NewEvent) (*schema.Event, error) {
	if (event.SubjectKind == nil) != (event.SubjectID == nil) {
		return nil, fmt.Errorf("an event's subject needs both kind and id")
	}
	now := db.Now()
	var eventID string
	err := s.database.Tx(ctx, func(tx *db.DB) error {
		var err error
		eventID, err = tx.Insert(ctx, eventTable, s.eventRow(event, now))
		if err != nil || event.SubjectKind == nil {
			return err
		}
		return applyToState(ctx, tx, event, now)
	})
	if err != nil {
		return nil, err
	}
	return db.Get[schema.Event](ctx, s.database, eventTable, eventID)
}

// eventRow is the row an event is stored as.
func (s *Service) eventRow(event NewEvent, now string) db.Vars {
	row := db.Vars{"at": now, "kind": event.Kind}
	if event.SubjectKind != nil {
		row["subjectKind"] = *event.SubjectKind
		row["subjectId"] = *event.SubjectID
	}
	if event.Data != nil {
		row["data"] = event.Data
	}
	if s.deviceID != "" {
		row["device"] = s.deviceID
	}
	if event.SessionID != nil {
		row["sessionId"] = *event.SessionID
	}
	return row
}

// applyToState updates the subject's user state for views and playback.
func applyToState(ctx context.Context, tx *db.DB, event NewEvent, now string) error {
	switch event.Kind {
	case schema.EventKindView:
		return updateState(ctx, tx, event, `"viewCount" = "viewCount" + 1, "lastViewedAt" = $now`, db.Vars{"now": now})
	case schema.EventKindProgress:
		position, ok := asNumber(event.Data["position"])
		if !ok {
			return fmt.Errorf("a progress event needs data.position in seconds")
		}
		return updateState(ctx, tx, event, `"progress" = $position`, db.Vars{"position": position})
	case schema.EventKindFinish:
		return updateState(ctx, tx, event, `"progress" = NULL`, db.Vars{})
	}
	return nil
}

// asNumber reads a number decoded from JSON, which arrives as float64,
// json.Number or an integer depending on the decoder.
func asNumber(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil
	}
	return 0, false
}

// updateState applies a SET clause to the subject's state row, creating it
// first when needed.
func updateState(ctx context.Context, tx *db.DB, event NewEvent, assignments string, vars db.Vars) error {
	stateID, err := ensureState(ctx, tx, *event.SubjectKind, *event.SubjectID)
	if err != nil {
		return err
	}
	vars["id"] = stateID
	vars["updatedAt"] = db.Now()
	_, err = tx.Exec(ctx, `UPDATE "UserState" SET `+assignments+`, "updatedAt" = $updatedAt WHERE "id" = $id`, vars)
	return err
}

// EventQuery narrows the event log. Empty fields don't filter.
type EventQuery struct {
	Kinds       []schema.EventKind
	SubjectKind *schema.RecordKind
	SubjectID   *string
	// Since and Until bound the time (timestamps in db.TimeFormat).
	Since  string
	Until  string
	Limit  int
	Offset int
}

// Events lists logged events matching query, newest first.
func (s *Service) Events(ctx context.Context, query EventQuery) ([]schema.Event, error) {
	conditions, vars := eventConditions(query)
	vars["limit"] = query.Limit
	vars["offset"] = query.Offset
	return db.QueryAs[schema.Event](ctx, s.database,
		`SELECT * FROM "Event" WHERE `+strings.Join(conditions, " AND ")+` ORDER BY "at" DESC, rowid DESC LIMIT $limit OFFSET $offset`,
		vars)
}

// CountEvents counts the events matching query (paging ignored), e.g. a
// video's o-count.
func (s *Service) CountEvents(ctx context.Context, query EventQuery) (int, error) {
	conditions, vars := eventConditions(query)
	return s.database.Int(ctx, `SELECT COUNT(*) FROM "Event" WHERE `+strings.Join(conditions, " AND "), vars)
}

// eventConditions turns an event query into WHERE conditions.
func eventConditions(query EventQuery) ([]string, db.Vars) {
	conditions := []string{`"deletedAt" IS NULL`}
	vars := db.Vars{}
	if len(query.Kinds) > 0 {
		conditions = append(conditions, `"kind" IN (SELECT value FROM json_each($kinds))`)
		vars["kinds"] = query.Kinds
	}
	if query.SubjectKind != nil {
		conditions = append(conditions, `"subjectKind" = $subjectKind`)
		vars["subjectKind"] = *query.SubjectKind
	}
	if query.SubjectID != nil {
		conditions = append(conditions, `"subjectId" = $subjectId`)
		vars["subjectId"] = *query.SubjectID
	}
	if query.Since != "" {
		conditions = append(conditions, `"at" >= $since`)
		vars["since"] = query.Since
	}
	if query.Until != "" {
		conditions = append(conditions, `"at" < $until`)
		vars["until"] = query.Until
	}
	return conditions, vars
}

// RecentSearches lists the texts of the latest searches, newest first,
// without repeats.
func (s *Service) RecentSearches(ctx context.Context, limit int) ([]string, error) {
	return s.database.Strings(ctx,
		`SELECT json_extract("data", '$.text') AS "text" FROM "Event"
		 WHERE "kind" = $kind AND "deletedAt" IS NULL AND json_extract("data", '$.text') IS NOT NULL
		 GROUP BY "text" ORDER BY MAX("at") DESC LIMIT $limit`,
		db.Vars{"kind": schema.EventKindSearch, "limit": limit})
}

// ClearEvents forgets every event of kinds (all kinds when empty), e.g. to
// clear the search history. View counts and resume points stay.
func (s *Service) ClearEvents(ctx context.Context, kinds []schema.EventKind) (int, error) {
	if len(kinds) == 0 {
		deleted, err := s.database.Exec(ctx, `DELETE FROM "Event"`, nil)
		return int(deleted), err
	}
	deleted, err := s.database.Exec(ctx, `DELETE FROM "Event" WHERE "kind" IN (SELECT value FROM json_each($kinds))`, db.Vars{"kinds": kinds})
	return int(deleted), err
}
