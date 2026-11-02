package jobs

import (
	"context"
	"log"
	"math"
	"sync"
	"time"
)

// pollInterval is how often idle workers look for jobs that became due
// (runAt reached) without a new enqueue waking them.
const pollInterval = 10 * time.Second

// Task does the work of one job. progress records how far it is (0..1). A
// returned error fails the attempt; the job is retried with backoff until
// the queue's retries are used up.
type Task func(ctx context.Context, job *Job, progress func(float64)) error

// Limits bound how a kind of job runs.
type Limits struct {
	// Concurrency is how many jobs of the kind run at once (at least 1).
	Concurrency int
	// RetryInitial is the delay before the first retry; each further retry
	// multiplies it by RetryMultiplier, up to RetryMax.
	RetryInitial    time.Duration
	RetryMultiplier float64
	RetryMax        time.Duration
}

// retryDelay is the delay before retrying after attempts failed attempts.
func (l Limits) retryDelay(attempts int) time.Duration {
	multiplier := l.RetryMultiplier
	if multiplier < 1 {
		multiplier = 1
	}
	delay := time.Duration(float64(l.RetryInitial) * math.Pow(multiplier, float64(attempts-1)))
	if l.RetryMax > 0 && delay > l.RetryMax {
		return l.RetryMax
	}
	return delay
}

// registration is a task with its limits.
type registration struct {
	task   Task
	limits Limits
	wake   chan struct{}
}

// Runner runs queued jobs with the task registered for their kind. Kinds
// without a task stay queued.
type Runner struct {
	queue *Queue
	tasks map[Kind]*registration
	// running maps the id of each running job to the cancel of its context.
	running sync.Map
}

// NewRunner returns a runner over queue. Register tasks before Start.
func NewRunner(queue *Queue) *Runner {
	return &Runner{queue: queue, tasks: map[Kind]*registration{}}
}

// Register sets the task that runs jobs of kind.
func (r *Runner) Register(kind Kind, limits Limits, task Task) {
	if limits.Concurrency < 1 {
		limits.Concurrency = 1
	}
	r.tasks[kind] = &registration{task: task, limits: limits, wake: make(chan struct{}, 1)}
}

// Start runs workers for every registered kind until ctx ends.
func (r *Runner) Start(ctx context.Context) {
	for kind, registered := range r.tasks {
		for range registered.limits.Concurrency {
			go r.work(ctx, kind, registered)
		}
	}
	go r.forwardWakeups(ctx)
}

// forwardWakeups passes each enqueue signal on to every kind's workers, so
// a new job of any kind is picked up at once.
func (r *Runner) forwardWakeups(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.queue.Notify():
			for _, registered := range r.tasks {
				wake(registered.wake)
			}
		}
	}
}

// wake signals a channel without blocking.
func wake(channel chan struct{}) {
	select {
	case channel <- struct{}{}:
	default:
	}
}

// work is one worker of kind: it runs jobs until none is due, then waits.
func (r *Runner) work(ctx context.Context, kind Kind, registered *registration) {
	for {
		r.drain(ctx, kind, registered)
		select {
		case <-ctx.Done():
			return
		case <-registered.wake:
		case <-time.After(pollInterval):
		}
	}
}

// drain claims and runs jobs of kind until none is due.
func (r *Runner) drain(ctx context.Context, kind Kind, registered *registration) {
	for ctx.Err() == nil {
		job, err := r.queue.Claim(ctx, []Kind{kind})
		if err != nil {
			log.Printf("jobs: claim %s: %v", kind, err)
			return
		}
		if job == nil {
			return
		}
		r.run(ctx, job, registered)
	}
}

// run executes one claimed job and records how it ended.
func (r *Runner) run(ctx context.Context, job *Job, registered *registration) {
	jobContext, cancel := context.WithCancel(ctx)
	r.running.Store(job.ID, cancel)
	defer func() {
		r.running.Delete(job.ID)
		cancel()
	}()
	progress := func(value float64) {
		if err := r.queue.SetProgress(ctx, job.ID, value); err != nil {
			log.Printf("jobs: progress %s: %v", job.ID, err)
		}
	}
	taskErr := registered.task(jobContext, job, progress)
	if jobContext.Err() != nil && ctx.Err() == nil {
		// Cancelled through Cancel: the queue already says so.
		return
	}
	if taskErr == nil {
		r.logError(job, r.queue.Succeed(ctx, job.ID))
		return
	}
	log.Printf("jobs: %s %s failed (attempt %d): %v", job.Kind, job.ID, job.Attempts, taskErr)
	r.logError(job, r.queue.Fail(ctx, job, taskErr, registered.limits.retryDelay(job.Attempts)))
}

// logError logs a failure to record a job's outcome.
func (r *Runner) logError(job *Job, err error) {
	if err != nil {
		log.Printf("jobs: record outcome of %s: %v", job.ID, err)
	}
}

// Cancel stops a job: a queued one never runs, a running one has its
// context cancelled.
func (r *Runner) Cancel(ctx context.Context, id string) error {
	wasRunning, err := r.queue.Cancel(ctx, id)
	if err != nil || !wasRunning {
		return err
	}
	cancel, running := r.running.Load(id)
	if running {
		cancel.(context.CancelFunc)()
	}
	return nil
}
