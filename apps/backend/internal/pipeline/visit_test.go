package pipeline

import "testing"

// TestMayRetryDetailWaitsBetweenTries checks that a visited scene without
// streams is fetched again at most once per interval.
func TestMayRetryDetailWaitsBetweenTries(t *testing.T) {
	orchestrator := &Orchestrator{}
	if !orchestrator.mayRetryDetail("scene:one") {
		t.Fatalf("first retry should be due")
	}
	if orchestrator.mayRetryDetail("scene:one") {
		t.Fatalf("second retry inside the interval should wait")
	}
	if !orchestrator.mayRetryDetail("scene:two") {
		t.Fatalf("another scene has its own interval")
	}
}
