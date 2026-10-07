package jobs

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// Queue is a crash-recoverable job queue backed by the job table.
// A single in-process mutex plus a write transaction prevents claim races.
type Queue struct {
	mu         sync.Mutex
	database   *db.DB
	notify     chan struct{}
	maxRetries int
}

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

// Enqueue adds a job, skipping if a pending/running job with the same dedupeKey exists.
// Returns the string ID of the job (existing or newly created).
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
		created, err := tx.Insert(ctx, "job", newJobFields(j))
		if err != nil {
			return fmt.Errorf("enqueue job: %w", err)
		}
		jobID = created.String()
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

// reuseOrClearDedupe returns the id of a pending/running job holding dedupeKey.
// A terminal (completed/failed) job with the key still holds the unique dedupe
// index, so it is deleted to let a re-run be queued.
func reuseOrClearDedupe(ctx context.Context, tx *db.DB, dedupeKey string) (string, error) {
	if dedupeKey == "" {
		return "", nil
	}
	vars := db.Vars{"key": dedupeKey}
	existing, err := tx.Strings(ctx,
		`SELECT id FROM job WHERE dedupe_key = $key AND status IN ('pending', 'running') LIMIT 1`, vars)
	if err != nil {
		return "", fmt.Errorf("lookup dedupe job: %w", err)
	}
	if len(existing) > 0 {
		return existing[0], nil
	}
	if _, err := tx.Exec(ctx,
		`DELETE FROM job WHERE dedupe_key = $key AND status NOT IN ('pending', 'running')`, vars); err != nil {
		return "", fmt.Errorf("clear terminal dedupe jobs: %w", err)
	}
	return "", nil
}

// newJobFields builds the insert row for a fresh pending job. Unset optional
// columns are omitted so they stay NULL (an empty dedupe_key would collide on
// the unique index).
func newJobFields(j Job) map[string]any {
	fields := map[string]any{
		"kind":        string(j.Kind),
		"status":      string(StatusPending),
		"plugin_name": j.PluginName,
		"payload":     j.Payload,
		"attempts":    0,
	}
	if j.DedupeKey != "" {
		fields["dedupe_key"] = j.DedupeKey
	}
	if j.RunAt != nil {
		fields["run_at"] = *j.RunAt
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

// Claim atomically takes the next pending job: highest priority first, then
// oldest, skipping jobs whose run_at is still in the future.
func (q *Queue) Claim(ctx context.Context, kinds []Kind) (*Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	kindNames := make([]string, len(kinds))
	for index, kind := range kinds {
		kindNames[index] = string(kind)
	}

	var claimed *Job
	err := q.database.Tx(ctx, func(tx *db.DB) error {
		nextIDs, err := tx.Strings(ctx,
			`SELECT id FROM job
			WHERE status = 'pending'
				AND kind IN (SELECT value FROM json_each($kinds))
				AND (run_at IS NULL OR run_at <= $now)
			ORDER BY priority DESC, created_at, rowid
			LIMIT 1`,
			db.Vars{"kinds": kindNames, "now": db.Now()})
		if err != nil || len(nextIDs) == 0 {
			return err
		}
		vars := db.Vars{"id": nextIDs[0]}
		if _, err := tx.Exec(ctx,
			`UPDATE job SET status = 'running', attempts = attempts + 1 WHERE id = $id`, vars); err != nil {
			return err
		}
		claimed, err = db.QueryOneAs[Job](ctx, tx, `SELECT * FROM job WHERE id = $id`, vars)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("claim job: %w", err)
	}
	return claimed, nil
}

// Complete marks a job as done and removes it from the queue.
func (q *Queue) Complete(ctx context.Context, id *db.RecordID) error {
	if id == nil {
		return nil
	}
	return q.database.Delete(ctx, *id)
}

// MarkCompleted marks a job as completed without deleting it.
func (q *Queue) MarkCompleted(ctx context.Context, id *db.RecordID) error {
	if id == nil {
		return nil
	}
	return q.database.Merge(ctx, *id, map[string]any{"status": StatusCompleted})
}

// Reorder assigns descending priorities so earlier ids are claimed first. Used
// to reorder pending downloads. Higher priority wins in Claim.
func (q *Queue) Reorder(ctx context.Context, ids []string) error {
	total := len(ids)
	for index, id := range ids {
		recordID, err := db.ParseRecordID(id)
		if err != nil {
			continue
		}
		if err := q.database.Merge(ctx, *recordID, map[string]any{"priority": total - index}); err != nil {
			return err
		}
	}
	return nil
}

// Defer reschedules a job to run after delay without counting it as a failed
// attempt. Used to hold a job back on a soft condition (e.g. active playback).
func (q *Queue) Defer(ctx context.Context, id *db.RecordID, delay time.Duration) error {
	if id == nil {
		return nil
	}
	return q.database.Merge(ctx, *id, map[string]any{
		"status": StatusPending,
		"run_at": time.Now().Add(delay),
	})
}

// Fail retries the job if attempts remain, otherwise marks it failed.
// retryAfter delays when the job becomes claimable again (0 = immediate).
func (q *Queue) Fail(ctx context.Context, j *Job, jobErr error, retryAfter time.Duration) error {
	if j == nil || j.ID == nil {
		return nil
	}
	if j.Attempts >= q.maxRetries {
		return q.database.Merge(ctx, *j.ID, map[string]any{
			"status": StatusFailed,
			"error":  jobErr.Error(),
		})
	}
	updates := map[string]any{
		"status": StatusPending,
		"error":  jobErr.Error(),
	}
	if retryAfter > 0 {
		updates["run_at"] = time.Now().Add(retryAfter)
	}
	if err := q.database.Merge(ctx, *j.ID, updates); err != nil {
		return err
	}
	if retryAfter == 0 {
		q.signal()
	}
	return nil
}

// ResetStuck moves running jobs back to pending on startup (crash recovery).
func (q *Queue) ResetStuck(ctx context.Context) error {
	if _, err := q.database.Exec(ctx, `UPDATE job SET status = 'pending' WHERE status = 'running'`, nil); err != nil {
		return fmt.Errorf("reset stuck jobs: %w", err)
	}
	log.Println("jobs: reset stuck running jobs to pending")
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

// Delete removes a single job by ID. Accepts both "job:xyz" and bare "xyz".
func (q *Queue) Delete(ctx context.Context, id string) error {
	if !strings.Contains(id, ":") {
		id = "job:" + id
	}
	recordID, err := db.ParseRecordID(id)
	if err != nil {
		return fmt.Errorf("delete job: parse id %q: %w", id, err)
	}
	return q.database.Delete(ctx, *recordID)
}

// Retry queues a failed job again with its attempts reset. Accepts both
// "job:xyz" and bare "xyz".
func (q *Queue) Retry(ctx context.Context, id string) error {
	if !strings.Contains(id, ":") {
		id = "job:" + id
	}
	if _, err := q.database.Exec(ctx,
		`UPDATE job SET status = 'pending', attempts = 0, error = NULL, run_at = NULL
		 WHERE id = $id AND status IN ('failed', 'cancelled', 'canceled')`,
		db.Vars{"id": id}); err != nil {
		return fmt.Errorf("retry job: %w", err)
	}
	q.signal()
	return nil
}

// DeleteByKind removes all jobs of the given kind and returns the count deleted.
func (q *Queue) DeleteByKind(ctx context.Context, kind Kind) (int, error) {
	deleted, err := q.database.Exec(ctx, `DELETE FROM job WHERE kind = $kind`, db.Vars{"kind": string(kind)})
	return int(deleted), err
}

// DownloadURLs returns the blob URLs of finished downloads still in the job
// list: the user's downloaded videos.
func (q *Queue) DownloadURLs(ctx context.Context) ([]string, error) {
	return q.database.Strings(ctx,
		`SELECT json_extract(payload, '$.download_url') FROM job
		 WHERE kind = $kind AND json_extract(payload, '$.download_url') IS NOT NULL`,
		db.Vars{"kind": string(KindDownload)})
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
		`UPDATE job SET status = 'failed', error = $error,
		   payload = json_remove(payload, '$.download_url')
		 WHERE kind = $kind AND json_extract(payload, '$.download_url') IN (SELECT value FROM json_each($urls))`,
		db.Vars{"kind": string(KindDownload), "error": missingDownloadError, "urls": missing})
	return int(changed), err
}

// Clear removes all jobs and returns the count deleted.
func (q *Queue) Clear(ctx context.Context) (int, error) {
	deleted, err := q.database.Exec(ctx, `DELETE FROM job`, nil)
	return int(deleted), err
}

// All returns all jobs (for GraphQL query).
func (q *Queue) All(ctx context.Context) ([]Job, error) {
	return db.QueryAs[Job](ctx, q.database, `SELECT * FROM job ORDER BY created_at DESC`, nil)
}
