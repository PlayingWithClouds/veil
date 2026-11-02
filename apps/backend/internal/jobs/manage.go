package jobs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// finishedStatuses are the statuses a job ends in.
var finishedStatuses = []Status{StatusSucceeded, StatusFailed, StatusCancelled}

// Query narrows a list of jobs. Empty fields don't filter.
type Query struct {
	Statuses []Status
	Kinds    []Kind
	Limit    int
	Offset   int
}

// List returns jobs matching query, newest first.
func (q *Queue) List(ctx context.Context, query Query) ([]Job, error) {
	conditions := []string{"1 = 1"}
	vars := db.Vars{"limit": query.Limit, "offset": query.Offset}
	if len(query.Statuses) > 0 {
		conditions = append(conditions, `"status" IN (SELECT value FROM json_each($statuses))`)
		vars["statuses"] = query.Statuses
	}
	if len(query.Kinds) > 0 {
		conditions = append(conditions, `"kind" IN (SELECT value FROM json_each($kinds))`)
		vars["kinds"] = query.Kinds
	}
	return db.QueryAs[Job](ctx, q.database,
		`SELECT * FROM "Job" WHERE `+strings.Join(conditions, " AND ")+` ORDER BY "createdAt" DESC, rowid DESC LIMIT $limit OFFSET $offset`,
		vars)
}

// Get loads one job; nil when it doesn't exist.
func (q *Queue) Get(ctx context.Context, id string) (*Job, error) {
	return db.QueryOneAs[Job](ctx, q.database, `SELECT * FROM "Job" WHERE "id" = $id`, db.Vars{"id": id})
}

// Succeed marks a running job as done. A job cancelled while it ran stays
// cancelled.
func (q *Queue) Succeed(ctx context.Context, id string) error {
	_, err := q.database.Exec(ctx,
		`UPDATE "Job" SET "status" = $succeeded, "progress" = 1, "error" = NULL, "finishedAt" = $now, "updatedAt" = $now
		 WHERE "id" = $id AND "status" = $running`,
		db.Vars{"id": id, "succeeded": StatusSucceeded, "running": StatusRunning, "now": db.Now()})
	return err
}

// Cancel stops a queued or running job from (further) running. It reports
// whether the job was running, so its runner can interrupt it.
func (q *Queue) Cancel(ctx context.Context, id string) (bool, error) {
	job, err := q.Get(ctx, id)
	if err != nil {
		return false, err
	}
	if job == nil {
		return false, fmt.Errorf("no job %s", id)
	}
	if job.Status != StatusQueued && job.Status != StatusRunning {
		return false, nil
	}
	err = q.database.Update(ctx, jobTable, id, db.Vars{"status": StatusCancelled, "finishedAt": db.Now()})
	return job.Status == StatusRunning, err
}

// SetProgress records how far a running job is, from 0 to 1.
func (q *Queue) SetProgress(ctx context.Context, id string, progress float64) error {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	_, err := q.database.Exec(ctx, `UPDATE "Job" SET "progress" = $progress WHERE "id" = $id`,
		db.Vars{"id": id, "progress": progress})
	return err
}

// PurgeFinished deletes jobs that finished before cutoff and returns how
// many. Downloads stay: their row is the record of the downloaded file.
func (q *Queue) PurgeFinished(ctx context.Context, cutoff time.Time) (int, error) {
	deleted, err := q.database.Exec(ctx,
		`DELETE FROM "Job" WHERE "status" IN (SELECT value FROM json_each($finished))
		 AND "kind" != $download AND COALESCE("finishedAt", "updatedAt") < $cutoff`,
		db.Vars{"finished": finishedStatuses, "download": KindDownload, "cutoff": db.FormatTime(cutoff)})
	return int(deleted), err
}
