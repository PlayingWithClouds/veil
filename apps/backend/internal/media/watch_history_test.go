package media

import (
	"testing"
	"time"
)

// TestPlayedSinceLastSaveCountsSittings checks how a progress save adds to a
// scene's played time: the gap to the previous save within a sitting, at most
// one save interval (and no more than the position) for a sitting's first save.
func TestPlayedSinceLastSaveCountsSittings(t *testing.T) {
	now := time.Date(2026, 10, 28, 20, 0, 0, 0, time.UTC)
	cases := []struct {
		name         string
		previousSave time.Time
		position     int
		want         int
	}{
		{"first save of a scene", time.Time{}, 15, 15},
		{"first save right after a skip", time.Time{}, 1800, 15},
		{"first save after a few seconds", time.Time{}, 4, 4},
		{"next save while playing", now.Add(-15 * time.Second), 2400, 15},
		{"save on pause, shortly after", now.Add(-6 * time.Second), 30, 6},
		{"first save after a long pause", now.Add(-10 * time.Minute), 600, 15},
		{"previous save in the future", now.Add(time.Minute), 600, 15},
	}
	for _, testCase := range cases {
		got := playedSinceLastSave(testCase.previousSave, now, testCase.position)
		if got != testCase.want {
			t.Errorf("%s: played %d, want %d", testCase.name, got, testCase.want)
		}
	}
}
