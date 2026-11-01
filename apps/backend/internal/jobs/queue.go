package jobs

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// jobTable is the table jobs are stored in.
const jobTable = "Job"

// Queue is a crash-recoverable job queue backed by the "Job" table.
// A single in-process mutex plus a write transaction prevents claim races.
type Queue struct {
	mu         sync.Mutex
	database   *db.DB
	notify     chan struct{}
	maxRetries int
}

// NewQueue returns a queue over database; a job fails for good after
// maxRetries attempts (3 when not positive).
func NewQueue(database *db.DB, maxRetries int) *Queue {
	if maxRetries <= 0 {
		maxRetries = 3
	}
	return &Queue{
		database:   database,
		notify:     make(chan struct{}, 1),
		maxRetries: maxRetries,
	}
}

// Enqueue adds a job, reusing a queued or running job with the same
// DedupeKey. It returns the id of the job (existing or newly created).
func (q *Queue) Enqueue(ctx context.Context, j Job) (string, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var jobID string
	isNew := false
	err := q.database.Tx(ctx, func(tx *db.DB) error {
		existingID, err := reuseOrClearDedupe(ctx, tx, j.DedupeKey)
		if err != nil || existingID != "" {
			jobID = existingID
			return err
		}
		created, err := tx.Insert(ctx, jobTable, newJobFields(j))
		if err != nil {
			return fmt.Errorf("enqueue job: %w", err)
		}
		jobID = created
		isNew = true
		return nil
	})
	if err != nil {
		return "", err
	}
	if isNew {
		q.signal()
	}
	return jobID, nil
}

// reuseOrClearDedupe returns the id of a queued or running job holding
// dedupeKey. A finished job with the key still holds the unique dedupe index,
// so it is deleted to let a re-run be queued.
func reuseOrClearDedupe(ctx context.Context, tx *db.DB, dedupeKey string) (string, error) {
	if dedupeKey == "" {
		return "", nil
	}
	vars := db.Vars{"key": dedupeKey, "queued": StatusQueued, "running": StatusRunning}
	existing, err := tx.Strings(ctx,
		`SELECT "id" FROM "Job" WHERE "dedupeKey" = $key AND "status" IN ($queued, $running) LIMIT 1`, vars)
	if err != nil {
		return "", fmt.Errorf("lookup dedupe job: %w", err)
	}
	if len(existing) > 0 {
		return existing[0], nil
	}
	if _, err := tx.Exec(ctx,
		`DELETE FROM "Job" WHERE "dedupeKey" = $key AND "status" NOT IN ($queued, $running)`, vars); err != nil {
		return "", fmt.Errorf("clear finished dedupe jobs: %w", err)
	}
	return "", nil
}

// newJobFields builds the insert row for a fresh queued job. An empty
// dedupeKey stays NULL, since "" would collide on the unique index.
func newJobFields(j Job) db.Vars {
	runAt := time.Now()
	if j.RunAt != nil {
		runAt = *j.RunAt
	}
	fields := db.Vars{
		"kind":     j.Kind,
		"status":   StatusQueued,
		"payload":  j.Payload,
		"attempts": 0,
		"priority": j.Priority,
		"progress": 0,
		"runAt":    db.FormatTime(runAt),
	}
	if j.PluginName != "" {
		fields["plugin"] = j.PluginName
	}
	if j.DedupeKey != "" {
		fields["dedupeKey"] = j.DedupeKey
	}
	if j.Error != "" {
		fields["error"] = j.Error
	}
	return fields
}

// signal wakes one waiting worker without blocking.
func (q *Queue) signal() {
	select {
	case q.notify <- struct{}{}:
	default:
	}
}

// Claim atomically takes the next queued job of one of kinds: highest
// priority first, then oldest, skipping jobs whose runAt is still ahead.
func (q *Queue) Claim(ctx context.Context, kinds []Kind) (*Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var claimed *Job
	err := q.database.Tx(ctx, func(tx *db.DB) error {
		nextIDs, err := tx.Strings(ctx,
			`SELECT "id" FROM "Job"
			WHERE "status" = $queued
				AND "kind" IN (SELECT value FROM json_each($kinds))
				AND "runAt" <= $now
			ORDER BY "priority" DESC, "createdAt", rowid
			LIMIT 1`,
			db.Vars{"queued": StatusQueued, "kinds": kinds, "now": db.Now()})
		if err != nil || len(nextIDs) == 0 {
			return err
		}
		vars := db.Vars{"id": nextIDs[0], "running": StatusRunning, "now": db.Now()}
		if _, err := tx.Exec(ctx,
			`UPDATE "Job" SET "status" = $running, "attempts" = "attempts" + 1, "startedAt" = $now, "updatedAt" = $now
			 WHERE "id" = $id`, vars); err != nil {
			return err
		}
		claimed, err = db.QueryOneAs[Job](ctx, tx, `SELECT * FROM "Job" WHERE "id" = $id`, vars)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("claim job: %w", err)
	}
	return claimed, nil
}

// Complete marks a job as done by removing it from the queue.
func (q *Queue) Complete(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}
	return q.database.Delete(ctx, jobTable, id)
}

// MarkCompleted marks a job as succeeded without deleting it.
func (q *Queue) MarkCompleted(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}
	return q.database.Update(ctx, jobTable, id, db.Vars{"status": StatusSucceeded, "finishedAt": db.Now()})
}

// Reorder assigns descending priorities so earlier ids are claimed first. Used
// to reorder queued downloads. Higher priority wins in Claim.
func (q *Queue) Reorder(ctx context.Context, ids []string) error {
	total := len(ids)
	for index, id := range ids {
		if err := q.database.Update(ctx, jobTable, id, db.Vars{"priority": total - index}); err != nil {
			return err
		}
	}
	return nil
}

// Defer reschedules a job to run after delay without counting it as a failed
// attempt. Used to hold a job back on a soft condition (e.g. active playback).
func (q *Queue) Defer(ctx context.Context, id string, delay time.Duration) error {
	if id == "" {
		return nil
	}
	return q.database.Update(ctx, jobTable, id, db.Vars{
		"status": StatusQueued,
		"runAt":  db.FormatTime(time.Now().Add(delay)),
	})
}

// Fail retries the job if attempts remain, otherwise marks it failed.
// retryAfter delays when the job becomes claimable again (0 = immediate).
func (q *Queue) Fail(ctx context.Context, j *Job, jobErr error, retryAfter time.Duration) error {
	if j == nil || j.ID == "" {
		return nil
	}
	if j.Attempts >= q.maxRetries {
		return q.database.Update(ctx, jobTable, j.ID, db.Vars{
			"status":     StatusFailed,
			"error":      jobErr.Error(),
			"finishedAt": db.Now(),
		})
	}
	updates := db.Vars{
		"status": StatusQueued,
		"error":  jobErr.Error(),
		"runAt":  db.FormatTime(time.Now().Add(retryAfter)),
	}
	if err := q.database.Update(ctx, jobTable, j.ID, updates); err != nil {
		return err
	}
	if retryAfter <= 0 {
		q.signal()
	}
	return nil
}

// ResetStuck moves running jobs back to the queue on startup (crash recovery).
func (q *Queue) ResetStuck(ctx context.Context) error {
	if _, err := q.database.Exec(ctx, `UPDATE "Job" SET "status" = $queued WHERE "status" = $running`,
		db.Vars{"queued": StatusQueued, "running": StatusRunning}); err != nil {
		return fmt.Errorf("reset stuck jobs: %w", err)
	}
	log.Println("jobs: reset stuck running jobs to queued")
	return nil
}

// DB exposes the underlying database for direct queries.
func (q *Queue) DB() *db.DB {
	return q.database
}

// Notify returns a channel that receives a signal when new jobs are enqueued.
func (q *Queue) Notify() <-chan struct{} {
	return q.notify
}

// Delete removes a single job by id.
func (q *Queue) Delete(ctx context.Context, id string) error {
	return q.database.Delete(ctx, jobTable, id)
}

// Retry queues a failed or cancelled job again with its attempts reset.
func (q *Queue) Retry(ctx context.Context, id string) error {
	if _, err := q.database.Exec(ctx,
		`UPDATE "Job" SET "status" = $queued, "attempts" = 0, "error" = NULL, "runAt" = $now, "updatedAt" = $now
		 WHERE "id" = $id AND "status" IN ($failed, $cancelled)`,
		db.Vars{"id": id, "queued": StatusQueued, "failed": StatusFailed, "cancelled": StatusCancelled, "now": db.Now()}); err != nil {
		return fmt.Errorf("retry job: %w", err)
	}
	q.signal()
	return nil
}

// DeleteByKind removes all jobs of the given kind and returns the count deleted.
func (q *Queue) DeleteByKind(ctx context.Context, kind Kind) (int, error) {
	deleted, err := q.database.Exec(ctx, `DELETE FROM "Job" WHERE "kind" = $kind`, db.Vars{"kind": kind})
	return int(deleted), err
}

// DownloadURLs returns the blob URLs of finished downloads still in the job
// list: the user's downloaded videos. The URL is the payload's download_url.
func (q *Queue) DownloadURLs(ctx context.Context) ([]string, error) {
	return q.database.Strings(ctx,
		`SELECT json_extract("payload", '$.download_url') FROM "Job"
		 WHERE "kind" = $kind AND json_extract("payload", '$.download_url') IS NOT NULL`,
		db.Vars{"kind": KindDownload})
}

// missingDownloadError is the error a download gets once its file is gone.
const missingDownloadError = "downloaded file is missing, download again"

// ForgetMissingDownloads marks finished downloads whose file no longer exists
// as failed and drops their blob URL, so nothing hands out a dead link and the
// download can be retried. It returns how many it changed.
func (q *Queue) ForgetMissingDownloads(ctx context.Context, exists func(downloadURL string) bool) (int, error) {
	urls, err := q.DownloadURLs(ctx)
	if err != nil {
		return 0, err
	}
	var missing []string
	for _, downloadURL := range urls {
		if !exists(downloadURL) {
			missing = append(missing, downloadURL)
		}
	}
	if len(missing) == 0 {
		return 0, nil
	}
	changed, err := q.database.Exec(ctx,
		`UPDATE "Job" SET "status" = $failed, "error" = $error,
		   "payload" = json_remove("payload", '$.download_url')
		 WHERE "kind" = $kind AND json_extract("payload", '$.download_url') IN (SELECT value FROM json_each($urls))`,
		db.Vars{"kind": KindDownload, "failed": StatusFailed, "error": missingDownloadError, "urls": missing})
	return int(changed), err
}

// Clear removes all jobs and returns the count deleted.
func (q *Queue) Clear(ctx context.Context) (int, error) {
	deleted, err := q.database.Exec(ctx, `DELETE FROM "Job"`, nil)
	return int(deleted), err
}

// All returns all jobs, newest first.
func (q *Queue) All(ctx context.Context) ([]Job, error) {
	return db.QueryAs[Job](ctx, q.database, `SELECT * FROM "Job" ORDER BY "createdAt" DESC, rowid DESC`, nil)
}
