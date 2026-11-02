// Package user stores what the (single) user does: their state on each
// record (rating, favourite, notes, block, resume point, views) and the
// append-only event log that history, o-counts, recent searches and stats
// are read from.
package user

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// stateTable is the table user state lives in.
const stateTable = "UserState"

// Service reads and writes user state and events.
type Service struct {
	database *db.DB
	// deviceID is this install's Device row, stamped on every event.
	deviceID string
	// onEvent, when set, gets every recorded event (automation triggers).
	onEvent func(ctx context.Context, event schema.Event)
}

// SetOnEvent sets a function that receives every event once it's recorded.
func (s *Service) SetOnEvent(hook func(ctx context.Context, event schema.Event)) {
	s.onEvent = hook
}

// New returns the service; deviceID may be "" when devices aren't set up.
func New(database *db.DB, deviceID string) *Service {
	return &Service{database: database, deviceID: deviceID}
}

// State loads the user's state on a record; nil when there is none yet.
func (s *Service) State(ctx context.Context, kind schema.RecordKind, id string) (*schema.UserState, error) {
	return db.QueryOneAs[schema.UserState](ctx, s.database,
		`SELECT * FROM "UserState" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL`,
		db.Vars{"kind": kind, "id": id})
}

// Rate sets the record's rating (0..10); nil clears it.
func (s *Service) Rate(ctx context.Context, kind schema.RecordKind, id string, rating *float64) (*schema.UserState, error) {
	if rating != nil && (*rating < 0 || *rating > 10) {
		return nil, fmt.Errorf("a rating is between 0 and 10, not %v", *rating)
	}
	fields := db.Vars{"rating": nil}
	if rating != nil {
		fields["rating"] = *rating
	}
	return s.update(ctx, kind, id, fields)
}

// SetFavorite marks or unmarks the record as a favourite.
func (s *Service) SetFavorite(ctx context.Context, kind schema.RecordKind, id string, favorite bool) (*schema.UserState, error) {
	return s.update(ctx, kind, id, db.Vars{"favorite": favorite})
}

// SetBlocked hides the record from every list and feed, or shows it again.
// A blocked performer, studio or tag also hides the media it's credited on.
func (s *Service) SetBlocked(ctx context.Context, kind schema.RecordKind, id string, blocked bool) (*schema.UserState, error) {
	return s.update(ctx, kind, id, db.Vars{"blocked": blocked})
}

// SetNotes sets the user's notes on the record; blank clears them.
func (s *Service) SetNotes(ctx context.Context, kind schema.RecordKind, id, notes string) (*schema.UserState, error) {
	fields := db.Vars{"notes": nil}
	if strings.TrimSpace(notes) != "" {
		fields["notes"] = notes
	}
	return s.update(ctx, kind, id, fields)
}

// update writes fields to the record's state, creating it on first use.
func (s *Service) update(ctx context.Context, kind schema.RecordKind, id string, fields db.Vars) (*schema.UserState, error) {
	err := s.database.Tx(ctx, func(tx *db.DB) error {
		stateID, err := ensureState(ctx, tx, kind, id)
		if err != nil {
			return err
		}
		return tx.Update(ctx, stateTable, stateID, fields)
	})
	if err != nil {
		return nil, err
	}
	return s.State(ctx, kind, id)
}

// ensureState returns the id of the record's state row, inserting an empty
// one when it has none. It fails for a record that doesn't exist.
func ensureState(ctx context.Context, tx *db.DB, kind schema.RecordKind, id string) (string, error) {
	existing, err := tx.Strings(ctx,
		`SELECT "id" FROM "UserState" WHERE "subjectKind" = $kind AND "subjectId" = $id`,
		db.Vars{"kind": kind, "id": id})
	if err != nil || len(existing) > 0 {
		return firstOf(existing), err
	}
	if err := requireRecord(ctx, tx, kind, id); err != nil {
		return "", err
	}
	return tx.Insert(ctx, stateTable, db.Vars{
		"subjectKind": kind,
		"subjectId":   id,
		"favorite":    false,
		"blocked":     false,
		"viewCount":   0,
		"fields":      map[string]any{},
	})
}

// requireRecord fails unless the record kind/id is stored (trashed counts).
func requireRecord(ctx context.Context, tx *db.DB, kind schema.RecordKind, id string) error {
	table := provenance.Table(kind)
	if table == "" {
		return fmt.Errorf("a %s can't have user state yet", kind)
	}
	found, err := tx.Int(ctx, `SELECT COUNT(*) FROM `+db.Quote(table)+` WHERE "id" = $id`, db.Vars{"id": id})
	if err != nil {
		return err
	}
	if found == 0 {
		return fmt.Errorf("no %s %s", kind, id)
	}
	return nil
}

// firstOf returns the first value, or "".
func firstOf(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// StateFilter narrows a list of user states. Nil fields don't filter.
type StateFilter struct {
	Kind     *schema.RecordKind
	Favorite *bool
	Blocked  *bool
	Rated    *bool
	// Viewed lists what was viewed, most recent first (the history).
	Viewed *bool
}

// States lists user states matching filter, most recently changed first
// (most recently viewed first when filtering by Viewed).
func (s *Service) States(ctx context.Context, filter StateFilter, limit, offset int) ([]schema.UserState, error) {
	conditions := []string{`"deletedAt" IS NULL`}
	vars := db.Vars{"limit": limit, "offset": offset}
	if filter.Kind != nil {
		conditions = append(conditions, `"subjectKind" = $kind`)
		vars["kind"] = *filter.Kind
	}
	conditions = appendFlag(conditions, `"favorite" = 1`, filter.Favorite)
	conditions = appendFlag(conditions, `"blocked" = 1`, filter.Blocked)
	conditions = appendFlag(conditions, `"rating" IS NOT NULL`, filter.Rated)
	conditions = appendFlag(conditions, `"lastViewedAt" IS NOT NULL`, filter.Viewed)
	order := `"updatedAt" DESC`
	if filter.Viewed != nil && *filter.Viewed {
		order = `"lastViewedAt" DESC`
	}
	return db.QueryAs[schema.UserState](ctx, s.database,
		`SELECT * FROM "UserState" WHERE `+strings.Join(conditions, " AND ")+` ORDER BY `+order+`, rowid DESC LIMIT $limit OFFSET $offset`,
		vars)
}

// appendFlag adds condition when flag is true, its negation when false.
func appendFlag(conditions []string, condition string, flag *bool) []string {
	if flag == nil {
		return conditions
	}
	if *flag {
		return append(conditions, condition)
	}
	return append(conditions, "NOT ("+condition+")")
}
