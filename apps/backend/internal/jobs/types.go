package jobs

import (
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type Kind string

const (
	KindScrape     Kind = "scrape"
	KindEnrich     Kind = "enrich"
	KindDownload   Kind = "download"
	KindSpeedCheck Kind = "speed-check"
)

// Job is persisted in the job table so work survives crashes.
type Job struct {
	ID         *db.RecordID   `json:"id,omitempty"`
	Kind       Kind           `json:"kind"`
	Status     Status         `json:"status"`
	PluginName string         `json:"plugin_name"`
	Payload    map[string]any `json:"payload"`
	Error      string         `json:"error,omitempty"`
	Attempts   int            `json:"attempts"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	RunAt      *time.Time     `json:"run_at,omitempty"` // nil = run immediately; set by backoff
	DedupeKey  string         `json:"dedupe_key,omitempty"`
}

// StringID returns the full "table:id" string representation.
func (j Job) StringID() string {
	if j.ID == nil {
		return ""
	}
	return j.ID.String()
}
