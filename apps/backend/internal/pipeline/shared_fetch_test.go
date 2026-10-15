package pipeline

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSharedDetailFetchMakesLateVisitsWait(t *testing.T) {
	orchestrator := &Orchestrator{}
	release := make(chan struct{})
	var runs atomic.Int32
	var finished atomic.Bool
	fetch := func(ctx context.Context, id string) error {
		runs.Add(1)
		select {
		case <-release:
		case <-ctx.Done():
			return ctx.Err()
		}
		finished.Store(true)
		return nil
	}

	// The prefetch gives up early; its fetch must keep running for the page.
	prefetchContext, cancelPrefetch := context.WithCancel(context.Background())
	prefetchDone := make(chan error, 1)
	go func() { prefetchDone <- orchestrator.sharedDetailFetch(prefetchContext, "scene:a", fetch) }()
	waitFor(t, func() bool { return orchestrator.fetchingDetail("scene:a") })
	cancelPrefetch()
	<-prefetchDone

	pageDone := make(chan error, 1)
	go func() { pageDone <- orchestrator.sharedDetailFetch(context.Background(), "scene:a", fetch) }()
	time.Sleep(20 * time.Millisecond)
	if finished.Load() {
		t.Fatal("fetch finished before it was released")
	}
	close(release)
	if err := <-pageDone; err != nil {
		t.Fatalf("page visit: %v", err)
	}
	if !finished.Load() || runs.Load() != 1 {
		t.Errorf("finished %v after %d runs, want one finished run", finished.Load(), runs.Load())
	}
}

// waitFor polls condition for up to a second.
func waitFor(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for !condition() {
		if time.Now().After(deadline) {
			t.Fatal("condition never held")
		}
		time.Sleep(time.Millisecond)
	}
}
