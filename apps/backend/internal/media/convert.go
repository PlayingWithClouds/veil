package media

import (
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// recordIDString renders a record id pointer as "table:id", or "" if nil.
func recordIDString(id *db.RecordID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

// recordIDPtrString renders an optional record link as a string pointer.
func recordIDPtrString(id *db.RecordID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// recordIDStrings renders a slice of record links as their string ids.
func recordIDStrings(ids []db.RecordID) []string {
	out := make([]string, 0, len(ids))
	for i := range ids {
		out = append(out, ids[i].String())
	}
	return out
}

// dateString renders an optional datetime as a date-only string (YYYY-MM-DD).
func dateString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format("2006-01-02")
	return &s
}

// timePtrString renders an optional datetime as an RFC3339 string pointer.
func timePtrString(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.UTC().Format(time.RFC3339)
	return &s
}

const (
	defaultPageLimit = 24
	maxPageLimit     = 100
)

func pageLimit(limit *int) int {
	if limit == nil || *limit <= 0 {
		return defaultPageLimit
	}
	if *limit > maxPageLimit {
		return maxPageLimit
	}
	return *limit
}

func pageOffset(offset *int) int {
	if offset == nil || *offset < 0 {
		return 0
	}
	return *offset
}
