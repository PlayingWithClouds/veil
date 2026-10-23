package recommendeval

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/recommend"
)

// Target kinds the rankings are scored against.
const (
	TargetSessionEnd = "session-ending watch"
	TargetWatch      = "any watch"
	TargetClick      = "click"
)

// targetKinds lists the kinds in report order.
var targetKinds = []string{TargetSessionEnd, TargetWatch, TargetClick}

// watchRecord is one scene's playback as the history ended up holding it.
type watchRecord struct {
	sceneID         string
	startedAt       time.Time
	lastActivityAt  time.Time
	progressSeconds int
}

// history is everything the user did, read before the database is pruned:
// the future the rankings are scored against.
type history struct {
	watches []watchRecord
	// firstClick is when each scene was first opened: a logged click on a
	// recommendation surface or its first visit, whichever came first.
	firstClick map[string]time.Time
	// sceneCreatedAt is when each scene entered the library.
	sceneCreatedAt map[string]time.Time
	// endedSession marks, per watch index, the watches nothing else followed
	// (see recommend.SessionGap) and that reached the session-end minimum.
	endedSession []bool
	// firstActivity and lastActivity bound the recorded engagement.
	firstActivity time.Time
	lastActivity  time.Time
}

// loadHistory reads the user's engagement and the scenes' creation times.
func loadHistory(ctx context.Context, database *db.DB) (*history, error) {
	past := &history{firstClick: map[string]time.Time{}, sceneCreatedAt: map[string]time.Time{}}
	if err := past.loadWatches(ctx, database); err != nil {
		return nil, err
	}
	if err := past.loadClicks(ctx, database); err != nil {
		return nil, err
	}
	if err := past.loadSceneCreation(ctx, database); err != nil {
		return nil, err
	}
	past.endedSession = endedSessions(past.watches)
	past.findActivityBounds()
	return past, nil
}

// loadWatches reads the scene watches.
func (past *history) loadWatches(ctx context.Context, database *db.DB) error {
	rows, err := database.Query(ctx,
		`SELECT media, started_at, updated_at, max_progress_seconds FROM watch_history WHERE media LIKE 'scene:%'`, nil)
	if err != nil {
		return fmt.Errorf("read watches: %w", err)
	}
	for _, row := range rows {
		past.watches = append(past.watches, watchRecord{
			sceneID:         asString(row["media"]),
			startedAt:       asTime(row["started_at"]),
			lastActivityAt:  asTime(row["updated_at"]),
			progressSeconds: db.AsInt(row["max_progress_seconds"]),
		})
	}
	return nil
}

// loadClicks reads the logged clicks and first visits.
func (past *history) loadClicks(ctx context.Context, database *db.DB) error {
	rows, err := database.Query(ctx,
		`SELECT media, min(at) AS at FROM (
			SELECT id AS media, detail_fetched_at AS at FROM scene WHERE detail_fetched_at IS NOT NULL
			UNION ALL
			SELECT media, shown_at FROM recommendation_impression WHERE kind = 'click'
		) WHERE media LIKE 'scene:%' GROUP BY media`, nil)
	if err != nil {
		return fmt.Errorf("read clicks: %w", err)
	}
	for _, row := range rows {
		past.firstClick[asString(row["media"])] = asTime(row["at"])
	}
	return nil
}

// loadSceneCreation reads when every scene was stored.
func (past *history) loadSceneCreation(ctx context.Context, database *db.DB) error {
	rows, err := database.Query(ctx, `SELECT id, created_at FROM scene`, nil)
	if err != nil {
		return fmt.Errorf("read scene creation: %w", err)
	}
	for _, row := range rows {
		past.sceneCreatedAt[asString(row["id"])] = asTime(row["created_at"])
	}
	return nil
}

// findActivityBounds sets the first and last moment of recorded engagement.
func (past *history) findActivityBounds() {
	moments := []time.Time{}
	for _, entry := range past.watches {
		moments = append(moments, entry.startedAt, entry.lastActivityAt)
	}
	for _, clickedAt := range past.firstClick {
		moments = append(moments, clickedAt)
	}
	if len(moments) == 0 {
		return
	}
	sort.Slice(moments, func(left, right int) bool { return moments[left].Before(moments[right]) })
	past.firstActivity = moments[0]
	past.lastActivity = moments[len(moments)-1]
}

// endedSessions marks the watches that reached the session-end minimum and
// were followed by no other scene's start or playback within the session gap;
// the same rule as the engine's strongest watch signal. A watch the gap has
// not yet passed at the last recorded moment is undecided and not marked.
func endedSessions(watches []watchRecord) []bool {
	moments := make([]watchRecord, 0, 2*len(watches))
	var latest time.Time
	for _, entry := range watches {
		moments = append(moments,
			watchRecord{sceneID: entry.sceneID, lastActivityAt: entry.startedAt},
			watchRecord{sceneID: entry.sceneID, lastActivityAt: entry.lastActivityAt})
		if entry.lastActivityAt.After(latest) {
			latest = entry.lastActivityAt
		}
	}
	sort.Slice(moments, func(left, right int) bool { return moments[left].lastActivityAt.Before(moments[right].lastActivityAt) })
	ended := make([]bool, len(watches))
	for index, entry := range watches {
		until := entry.lastActivityAt.Add(recommend.SessionGap)
		decided := !until.After(latest)
		reached := entry.progressSeconds >= recommend.SessionEndMinimumSeconds
		ended[index] = decided && reached && !otherSceneWithin(moments, entry.sceneID, entry.lastActivityAt, until)
	}
	return ended
}

// otherSceneWithin reports whether a scene other than sceneID was active
// after from and up to until; moments is sorted by time.
func otherSceneWithin(moments []watchRecord, sceneID string, from, until time.Time) bool {
	first := sort.Search(len(moments), func(index int) bool { return moments[index].lastActivityAt.After(from) })
	for index := first; index < len(moments) && !moments[index].lastActivityAt.After(until); index++ {
		if moments[index].sceneID != sceneID {
			return true
		}
	}
	return false
}

// targetSet is the scenes the user turned to inside one window, split by what
// was reachable.
type targetSet struct {
	// reachable are the targets that already existed at the cutoff: the only
	// ones a ranking could have offered.
	reachable map[string]bool
	// unreachable counts the targets stored only after the cutoff.
	unreachable int
}

// targetsIn collects, per kind, the scenes first watched or opened in
// (cutoff, cutoff+window]. A scene watched or opened before the cutoff is no
// discovery and is left out.
func (past *history) targetsIn(cutoff time.Time, window time.Duration) map[string]*targetSet {
	end := cutoff.Add(window)
	targets := map[string]*targetSet{}
	for _, kind := range targetKinds {
		targets[kind] = &targetSet{reachable: map[string]bool{}}
	}
	for index, entry := range past.watches {
		if !entry.startedAt.After(cutoff) || entry.startedAt.After(end) {
			continue
		}
		past.addTarget(targets[TargetWatch], entry.sceneID, cutoff)
		if past.endedSession[index] {
			past.addTarget(targets[TargetSessionEnd], entry.sceneID, cutoff)
		}
	}
	for sceneID, clickedAt := range past.firstClick {
		if clickedAt.After(cutoff) && !clickedAt.After(end) {
			past.addTarget(targets[TargetClick], sceneID, cutoff)
		}
	}
	return targets
}

// addTarget files a scene as reachable when it existed at the cutoff.
func (past *history) addTarget(set *targetSet, sceneID string, cutoff time.Time) {
	createdAt, known := past.sceneCreatedAt[sceneID]
	if !known || createdAt.After(cutoff) {
		set.unreachable++
		return
	}
	set.reachable[sceneID] = true
}

// asString reads a text column.
func asString(value any) string {
	switch text := value.(type) {
	case string:
		return text
	case []byte:
		return string(text)
	}
	return ""
}

// asTime parses a stored timestamp; the zero time when it is missing or malformed.
func asTime(value any) time.Time {
	text := asString(value)
	if text == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return time.Time{}
	}
	return parsed.UTC()
}
