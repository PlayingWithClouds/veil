package files

import (
	"context"
	"time"
)

// pacer holds a transfer back to a byte rate. The stream cache has no
// throttle of its own, but calls its progress callback synchronously from
// the copy loop (MP4) or after each segment (HLS), so waiting in the
// callback slows the transfer down.
type pacer struct {
	bytesPerSecond float64
	started        time.Time
}

// newPacer returns a pacer for kilobytesPerSecond; 0 or less = unlimited.
func newPacer(kilobytesPerSecond int) *pacer {
	return &pacer{bytesPerSecond: float64(kilobytesPerSecond) * 1024, started: time.Now()}
}

// wait blocks until received bytes are within the rate, or ctx ends.
func (p *pacer) wait(ctx context.Context, received int64) {
	if p.bytesPerSecond <= 0 {
		return
	}
	due := p.started.Add(time.Duration(float64(received) / p.bytesPerSecond * float64(time.Second)))
	delay := time.Until(due)
	if delay <= 0 {
		return
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// progressStep is the smallest progress change worth writing to the job.
const progressStep = 0.01

// steppedProgress passes progress on only in steps of progressStep, since
// the stream cache reports it on every read.
type steppedProgress struct {
	report   func(float64)
	reported float64
}

// set reports value (0..1) when it moved a step past the last report, or
// first reached 1.
func (p *steppedProgress) set(value float64) {
	reachedEnd := value >= 1 && p.reported < 1
	if value-p.reported < progressStep && !reachedEnd {
		return
	}
	p.reported = value
	p.report(value)
}
