package jobs

import (
	"time"

	"github.com/playingwithclouds/veil/internal/schema"
)

// Status is where a job is in its life: queued, running, or finished.
type Status = schema.JobStatus

const (
	StatusQueued    = schema.JobStatusQueued
	StatusRunning   = schema.JobStatusRunning
	StatusSucceeded = schema.JobStatusSucceeded
	StatusFailed    = schema.JobStatusFailed
	StatusCancelled = schema.JobStatusCancelled
)

// Kind is what a job does.
type Kind = schema.JobKind

const (
	KindScrape     = schema.JobKindScrape
	KindEnrich     = schema.JobKindEnrich
	KindDownload   = schema.JobKindDownload
	KindSpeedCheck = schema.JobKindSpeedCheck
)

// Job is a row of the "Job" table, so work survives crashes. It mirrors
// schema.Job with a decoded payload and parsed times.
type Job struct {
	ID         string         `json:"id,omitempty"`
	Kind       Kind           `json:"kind"`
	Status     Status         `json:"status"`
	PluginName string         `json:"plugin,omitempty"`
	Payload    map[string]any `json:"payload,omitempty"`
	Error      string         `json:"error,omitempty"`
	Attempts   int            `json:"attempts"`
	Priority   int            `json:"priority"`
	Progress   float64        `json:"progress"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	// RunAt is when the job becomes claimable; nil on enqueue means now.
	RunAt      *time.Time `json:"runAt,omitempty"`
	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	DedupeKey  string     `json:"dedupeKey,omitempty"`
}
