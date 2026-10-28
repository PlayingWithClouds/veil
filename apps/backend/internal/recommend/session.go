package recommend

import (
	"sort"
	"time"
)

// watch is one watch history row: a scene's playback, how long it played and
// when it was first started and last played.
type watch struct {
	sceneID string
	// playedSeconds is how long the scene played over all sittings.
	playedSeconds int
	// stopSeconds is where playback was when the scene was last left.
	stopSeconds     int
	durationSeconds int
	startedAt       time.Time
	lastActivityAt  time.Time
}

// activity is one moment some scene was playing.
type activity struct {
	at      time.Time
	sceneID string
}

// followedWatches reports, per watch, whether another scene started or played
// within sessionGap after the watch's last activity. A row is reused when a
// scene is played again, so both its start and its last activity count.
func followedWatches(watches []watch) []bool {
	moments := make([]activity, 0, 2*len(watches))
	for _, entry := range watches {
		moments = append(moments,
			activity{at: entry.startedAt, sceneID: entry.sceneID},
			activity{at: entry.lastActivityAt, sceneID: entry.sceneID})
	}
	sort.Slice(moments, func(left, right int) bool { return moments[left].at.Before(moments[right].at) })

	followed := make([]bool, len(watches))
	for index, entry := range watches {
		followed[index] = otherActivityWithin(moments, entry.sceneID, entry.lastActivityAt, entry.lastActivityAt.Add(sessionGap))
	}
	return followed
}

// otherActivityWithin reports whether a scene other than sceneID was playing
// after from and up to until; moments is sorted by time.
func otherActivityWithin(moments []activity, sceneID string, from, until time.Time) bool {
	first := sort.Search(len(moments), func(index int) bool { return moments[index].at.After(from) })
	for index := first; index < len(moments) && !moments[index].at.After(until); index++ {
		if moments[index].sceneID != sceneID {
			return true
		}
	}
	return false
}
