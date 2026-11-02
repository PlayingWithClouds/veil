package jobs

import (
	"context"
	"errors"
	"testing"
	"time"
)

// waitForStatus polls until the job has status, failing after a few seconds.
func waitForStatus(t *testing.T, ctx context.Context, queue *Queue, id string, status Status) *Job {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		job, err := queue.Get(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if job != nil && job.Status == status {
			return job
		}
		time.Sleep(20 * time.Millisecond)
	}
	job, _ := queue.Get(ctx, id)
	t.Fatalf("job %s never became %s: %+v", id, status, job)
	return nil
}

// TestRunnerRunsRetriesAndCancels runs a job to success, retries a failing
// one until it succeeds, and cancels a running one.
func TestRunnerRunsRetriesAndCancels(t *testing.T) {
	queue, ctx := openTestQueue(t)
	runContext, stop := context.WithCancel(ctx)
	defer stop()
	runner := NewRunner(queue)
	failures := 1
	started := make(chan struct{})
	runner.Register(KindScrape, Limits{Concurrency: 2}, func(ctx context.Context, job *Job, progress func(float64)) error {
		progress(0.5)
		switch job.Payload["mode"] {
		case "flaky":
			if failures > 0 {
				failures--
				return errors.New("try again")
			}
		case "slow":
			close(started)
			<-ctx.Done()
			return ctx.Err()
		}
		return nil
	})
	runner.Start(runContext)

	plain := enqueue(t, ctx, queue, Job{Kind: KindScrape})
	flaky := enqueue(t, ctx, queue, Job{Kind: KindScrape, Payload: map[string]any{"mode": "flaky"}})
	if job := waitForStatus(t, ctx, queue, plain, StatusSucceeded); job.Progress != 1 {
		t.Errorf("progress = %v, want 1", job.Progress)
	}
	if job := waitForStatus(t, ctx, queue, flaky, StatusSucceeded); job.Attempts != 2 {
		t.Errorf("attempts = %d, want 2", job.Attempts)
	}

	slow := enqueue(t, ctx, queue, Job{Kind: KindScrape, Payload: map[string]any{"mode": "slow"}})
	<-started
	if err := runner.Cancel(ctx, slow); err != nil {
		t.Fatal(err)
	}
	waitForStatus(t, ctx, queue, slow, StatusCancelled)
	time.Sleep(50 * time.Millisecond)
	if job, _ := queue.Get(ctx, slow); job.Status != StatusCancelled {
		t.Errorf("cancelled job ended as %s", job.Status)
	}
}

// TestRetryDelay grows the delay by the multiplier up to the cap.
func TestRetryDelay(t *testing.T) {
	limits := Limits{RetryInitial: time.Second, RetryMultiplier: 2, RetryMax: 5 * time.Second}
	for attempts, want := range map[int]time.Duration{1: time.Second, 2: 2 * time.Second, 3: 4 * time.Second, 4: 5 * time.Second} {
		if got := limits.retryDelay(attempts); got != want {
			t.Errorf("retryDelay(%d) = %v, want %v", attempts, got, want)
		}
	}
}
