package graphql

import (
	"context"

	"github.com/playingwithclouds/veil/internal/schema"
)

// defaultFollowInterval is how often a follow runs unless told otherwise.
const defaultFollowInterval = 24

// followInterval is the requested interval in hours, else the default.
func followInterval(hours *int) int {
	if hours == nil || *hours <= 0 {
		return defaultFollowInterval
	}
	return *hours
}

// notifyOrDefault is the requested notify flag; follows notify by default.
func notifyOrDefault(notify *bool) bool {
	return notify == nil || *notify
}

// stepsOf copies the API's steps.
func stepsOf(steps []*schema.Step) []schema.Step {
	copied := make([]schema.Step, 0, len(steps))
	for _, step := range steps {
		if step != nil {
			copied = append(copied, *step)
		}
	}
	return copied
}

// notificationStream forwards new notifications to a subscriber until its
// context ends.
func (r *Resolver) notificationStream(ctx context.Context) <-chan *schema.Notification {
	added, cancel := r.notify.Subscribe()
	forwarded := make(chan *schema.Notification, 1)
	go func() {
		defer cancel()
		defer close(forwarded)
		for {
			select {
			case <-ctx.Done():
				return
			case notification, open := <-added:
				if !open {
					return
				}
				forwarded <- &notification
			}
		}
	}()
	return forwarded
}
