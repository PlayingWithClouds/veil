package jobs

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// openTestQueue returns a queue over a fresh database.
func openTestQueue(t *testing.T) (*Queue, context.Context) {
	t.Helper()
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return NewQueue(database, 3), ctx
}

// enqueue adds job and returns its id, failing the test on error.
func enqueue(t *testing.T, ctx context.Context, queue *Queue, job Job) string {
	t.Helper()
	id, err := queue.Enqueue(ctx, job)
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	return id
}

// claim takes the next job of kinds, failing the test on error.
func claim(t *testing.T, ctx context.Context, queue *Queue, kinds ...Kind) *Job {
	t.Helper()
	job, err := queue.Claim(ctx, kinds)
	if err != nil {
		t.Fatalf("claim: %v", err)
	}
	return job
}

// TestEnqueueDedupeReusesActiveJob reuses a queued or running job with the
// same key and re-queues once it finished.
func TestEnqueueDedupeReusesActiveJob(t *testing.T) {
	queue, ctx := openTestQueue(t)
	job := Job{Kind: KindScrape, PluginName: "p", Payload: map[string]any{"url": "u"}, DedupeKey: "scrape:p:u"}

	firstID := enqueue(t, ctx, queue, job)
	if secondID := enqueue(t, ctx, queue, job); secondID != firstID {
		t.Fatalf("queued duplicate: got %s, want %s", secondID, firstID)
	}

	claimed := claim(t, ctx, queue, KindScrape)
	if claimed == nil || claimed.ID != firstID || claimed.PluginName != "p" {
		t.Fatalf("claimed %+v, want %s", claimed, firstID)
	}
	if runningID := enqueue(t, ctx, queue, job); runningID != firstID {
		t.Fatalf("running duplicate: got %s, want %s", runningID, firstID)
	}

	if err := queue.MarkCompleted(ctx, claimed.ID); err != nil {
		t.Fatalf("mark completed: %v", err)
	}
	rerunID := enqueue(t, ctx, queue, job)
	if rerunID == firstID {
		t.Fatalf("completed job was reused instead of re-queued")
	}
	all, err := queue.All(ctx)
	if err != nil {
		t.Fatalf("all: %v", err)
	}
	if len(all) != 1 || all[0].Status != StatusQueued {
		t.Fatalf("jobs after re-run: %+v", all)
	}
}

// TestEnqueueWithoutDedupeKeyNeverCollides gives every keyless job its own row.
func TestEnqueueWithoutDedupeKeyNeverCollides(t *testing.T) {
	queue, ctx := openTestQueue(t)
	firstID := enqueue(t, ctx, queue, Job{Kind: KindDownload})
	secondID := enqueue(t, ctx, queue, Job{Kind: KindDownload})
	if firstID == secondID {
		t.Fatalf("jobs without dedupe key share id %s", firstID)
	}
}

// TestClaimOrdersByPriorityThenAgeAndFiltersKinds claims the highest
// priority first, then the oldest, and only the asked kinds.
func TestClaimOrdersByPriorityThenAgeAndFiltersKinds(t *testing.T) {
	queue, ctx := openTestQueue(t)
	oldestID := enqueue(t, ctx, queue, Job{Kind: KindDownload, PluginName: "a"})
	newerID := enqueue(t, ctx, queue, Job{Kind: KindDownload, PluginName: "b"})
	urgentID := enqueue(t, ctx, queue, Job{Kind: KindDownload, PluginName: "c"})
	otherKindID := enqueue(t, ctx, queue, Job{Kind: KindScrape, PluginName: "d"})

	if err := queue.Reorder(ctx, []string{urgentID}); err != nil {
		t.Fatalf("reorder: %v", err)
	}

	for _, wantID := range []string{urgentID, oldestID, newerID} {
		job := claim(t, ctx, queue, KindDownload)
		if job == nil || job.ID != wantID {
			t.Fatalf("claimed %+v, want %s", job, wantID)
		}
		if job.Status != StatusRunning || job.Attempts != 1 || job.StartedAt == nil {
			t.Fatalf("claimed job state: %+v", job)
		}
	}
	if job := claim(t, ctx, queue, KindDownload); job != nil {
		t.Fatalf("expected empty download queue, got %s", job.ID)
	}
	if job := claim(t, ctx, queue, KindScrape, KindEnrich); job == nil || job.ID != otherKindID {
		t.Fatalf("scrape claim got %+v, want %s", job, otherKindID)
	}
}

// TestClaimRespectsRunAt skips jobs scheduled for later.
func TestClaimRespectsRunAt(t *testing.T) {
	queue, ctx := openTestQueue(t)
	future := time.Now().Add(time.Hour)
	enqueue(t, ctx, queue, Job{Kind: KindScrape, RunAt: &future})
	readyID := enqueue(t, ctx, queue, Job{Kind: KindScrape})

	job := claim(t, ctx, queue, KindScrape)
	if job == nil || job.ID != readyID {
		t.Fatalf("claimed %+v, want ready job %s", job, readyID)
	}
	if job := claim(t, ctx, queue, KindScrape); job != nil {
		t.Fatalf("future job claimed early: %s", job.ID)
	}

	// Deferring into the past makes a job claimable again.
	if err := queue.Defer(ctx, job.ID, -time.Second); err != nil {
		t.Fatalf("defer: %v", err)
	}
	if again := claim(t, ctx, queue, KindScrape); again == nil || again.ID != readyID {
		t.Fatalf("deferred job not reclaimed: %+v", again)
	}
}

// TestFailRetriesThenFails retries until maxRetries, then marks failed.
func TestFailRetriesThenFails(t *testing.T) {
	queue, ctx := openTestQueue(t)
	jobID := enqueue(t, ctx, queue, Job{Kind: KindEnrich})
	jobErr := errors.New("boom")

	for attempt := 1; attempt <= 3; attempt++ {
		job := claim(t, ctx, queue, KindEnrich)
		if job == nil || job.ID != jobID || job.Attempts != attempt {
			t.Fatalf("attempt %d: claimed %+v", attempt, job)
		}
		if err := queue.Fail(ctx, job, jobErr, 0); err != nil {
			t.Fatalf("fail: %v", err)
		}
	}
	if job := claim(t, ctx, queue, KindEnrich); job != nil {
		t.Fatalf("exhausted job claimed again")
	}
	all, err := queue.All(ctx)
	if err != nil {
		t.Fatalf("all: %v", err)
	}
	if all[0].Status != StatusFailed || all[0].Error != "boom" || all[0].FinishedAt == nil {
		t.Fatalf("final job state: %+v", all[0])
	}
}

// TestResetStuckRequeuesRunningJobs puts running jobs back after a crash.
func TestResetStuckRequeuesRunningJobs(t *testing.T) {
	queue, ctx := openTestQueue(t)
	jobID := enqueue(t, ctx, queue, Job{Kind: KindScrape, Payload: map[string]any{"url": "x"}})
	if job := claim(t, ctx, queue, KindScrape); job == nil {
		t.Fatalf("nothing claimed")
	}
	if err := queue.ResetStuck(ctx); err != nil {
		t.Fatalf("reset stuck: %v", err)
	}
	job := claim(t, ctx, queue, KindScrape)
	if job == nil || job.ID != jobID {
		t.Fatalf("stuck job not requeued: %+v", job)
	}
	if job.Payload["url"] != "x" {
		t.Fatalf("payload lost: %+v", job.Payload)
	}
}

// TestRetryRequeuesFailedJob re-queues a failed job and leaves running ones.
func TestRetryRequeuesFailedJob(t *testing.T) {
	queue, ctx := openTestQueue(t)
	jobID := enqueue(t, ctx, queue, Job{Kind: KindDownload})
	for attempt := 0; attempt < queue.maxRetries; attempt++ {
		job := claim(t, ctx, queue, KindDownload)
		queue.Fail(ctx, job, errors.New("boom"), 0)
	}
	if err := queue.Retry(ctx, jobID); err != nil {
		t.Fatalf("retry: %v", err)
	}
	all, _ := queue.All(ctx)
	if all[0].Status != StatusQueued || all[0].Attempts != 0 || all[0].Error != "" {
		t.Fatalf("retried job state: %+v", all[0])
	}
	if job := claim(t, ctx, queue, KindDownload); job == nil {
		t.Fatal("a retried job must be claimable")
	}
	if err := queue.Retry(ctx, jobID); err != nil {
		t.Fatalf("retry running: %v", err)
	}
	all, _ = queue.All(ctx)
	if all[0].Status != StatusRunning {
		t.Fatalf("retry must leave a running job alone: %+v", all[0])
	}
}

// TestDeleteByIDKindAndClear removes one job, a kind, then the rest.
func TestDeleteByIDKindAndClear(t *testing.T) {
	queue, ctx := openTestQueue(t)
	jobID := enqueue(t, ctx, queue, Job{Kind: KindScrape})
	if err := queue.Delete(ctx, jobID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	enqueue(t, ctx, queue, Job{Kind: KindScrape})
	enqueue(t, ctx, queue, Job{Kind: KindDownload})
	deleted, err := queue.DeleteByKind(ctx, KindScrape)
	if err != nil || deleted != 1 {
		t.Fatalf("delete by kind: deleted=%d err=%v", deleted, err)
	}
	cleared, err := queue.Clear(ctx)
	if err != nil || cleared != 1 {
		t.Fatalf("clear: cleared=%d err=%v", cleared, err)
	}
}

// TestDownloadURLsListsFinishedDownloads lists only downloads with a file.
func TestDownloadURLsListsFinishedDownloads(t *testing.T) {
	queue, ctx := openTestQueue(t)
	finishedID := enqueue(t, ctx, queue, Job{Kind: KindDownload, DedupeKey: "a"})
	enqueue(t, ctx, queue, Job{Kind: KindDownload, DedupeKey: "b"})
	enqueue(t, ctx, queue, Job{Kind: KindScrape})
	payload := db.Vars{"payload": map[string]any{"download_url": "http://localhost/api/blob/stream-cache/mp4/x.mp4"}}
	if err := queue.DB().Update(ctx, jobTable, finishedID, payload); err != nil {
		t.Fatalf("update: %v", err)
	}

	urls, err := queue.DownloadURLs(ctx)
	if err != nil {
		t.Fatalf("download urls: %v", err)
	}
	if len(urls) != 1 || urls[0] != "http://localhost/api/blob/stream-cache/mp4/x.mp4" {
		t.Errorf("download urls = %v", urls)
	}
}

// TestForgetMissingDownloads fails downloads whose file is gone.
func TestForgetMissingDownloads(t *testing.T) {
	queue, ctx := openTestQueue(t)
	for key, downloadURL := range map[string]string{"kept": "http://h/api/blob/kept.mp4", "lost": "http://h/api/blob/lost.mp4"} {
		jobID := enqueue(t, ctx, queue, Job{Kind: KindDownload, DedupeKey: key})
		fields := db.Vars{"status": StatusSucceeded, "payload": map[string]any{"title": key, "download_url": downloadURL}}
		if err := queue.DB().Update(ctx, jobTable, jobID, fields); err != nil {
			t.Fatalf("update: %v", err)
		}
	}

	forgotten, err := queue.ForgetMissingDownloads(ctx, func(downloadURL string) bool {
		return downloadURL == "http://h/api/blob/kept.mp4"
	})
	if err != nil || forgotten != 1 {
		t.Fatalf("forgotten = %d, %v", forgotten, err)
	}
	urls, _ := queue.DownloadURLs(ctx)
	if len(urls) != 1 || urls[0] != "http://h/api/blob/kept.mp4" {
		t.Errorf("remaining download urls = %v", urls)
	}
	failed, _ := queue.DB().Strings(ctx, `SELECT json_extract("payload", '$.title') FROM "Job" WHERE "status" = 'failed'`, nil)
	if len(failed) != 1 || failed[0] != "lost" {
		t.Errorf("failed downloads = %v", failed)
	}
}
