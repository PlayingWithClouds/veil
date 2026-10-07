package jobs

import (
	"context"
	"log"
	"time"
)

// Handler is a function that executes a job.
type Handler func(ctx context.Context, j *Job)

// Worker pulls jobs from the queue and executes them.
type Worker struct {
	queue   *Queue
	handler Handler
	kinds   []Kind
	workers int
	vpnGate func() bool // if non-nil, drain is skipped when it returns false
}

func NewWorker(queue *Queue, kinds []Kind, concurrency int, handler Handler) *Worker {
	return &Worker{
		queue:   queue,
		handler: handler,
		kinds:   kinds,
		workers: concurrency,
	}
}

// SetVpnGate installs a function that gates job execution.
// When the function returns false, the worker skips claiming new jobs.
func (w *Worker) SetVpnGate(fn func() bool) { w.vpnGate = fn }

// Start launches the worker pool. Blocks until ctx is cancelled.
func (w *Worker) Start(ctx context.Context) {
	for i := 0; i < w.workers; i++ {
		go w.loop(ctx)
	}
	<-ctx.Done()
}

func (w *Worker) loop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.queue.Notify():
			if w.vpnGate == nil || w.vpnGate() {
				w.drain(ctx)
			} else {
				log.Println("worker: VPN not connected, skipping job claim")
			}
		case <-time.After(10 * time.Second):
			if w.vpnGate == nil || w.vpnGate() {
				w.drain(ctx)
			}
		}
	}
}

func (w *Worker) drain(ctx context.Context) {
	for {
		j, err := w.queue.Claim(ctx, w.kinds)
		if err != nil {
			log.Printf("worker: claim error: %v", err)
			return
		}
		if j == nil {
			return // nothing left
		}
		w.handler(ctx, j)
	}
}
