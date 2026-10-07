package recommend

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// watchState is how far the user got through a scene.
type watchState int

const (
	watchNone watchState = iota
	watchPartial
	watchCompleted
)

// snapshot is everything one ranking pass reads about the user, taken once so
// every candidate is scored against the same state.
type snapshot struct {
	now     time.Time
	profile *Profile
	ignored map[string]ignoredImpression
	watched map[string]watchState
	// excluded holds blocked scenes and unfetched stubs of inactive plugins.
	excluded          map[string]bool
	dislikedNeighbors map[string]bool
}

// watchStateQuery reads every scene watch with its furthest position.
const watchStateQuery = `SELECT watch_history.media, watch_history.completed,
	watch_history.max_progress_seconds AS progress,
	coalesce(watch_history.duration_seconds, scene.duration_seconds) AS duration
	FROM watch_history LEFT JOIN scene ON scene.id = watch_history.media
	WHERE watch_history.media LIKE 'scene:%'`

// profileAt builds the taste profile as of now, returning the ignored
// impressions it read along the way.
func (e *Engine) profileAt(ctx context.Context, now time.Time) (*Profile, map[string]ignoredImpression, error) {
	ignored, err := e.ignoredImpressions(ctx, now)
	if err != nil {
		return nil, nil, err
	}
	signals, err := e.collectSignals(ctx, now, ignored)
	if err != nil {
		return nil, nil, err
	}
	features, err := e.loadFeatures(ctx, signalSceneIDs(signals))
	if err != nil {
		return nil, nil, err
	}
	return buildProfile(signals, features, e.repo.BlockedTargetIDSet(ctx), now), ignored, nil
}

// takeSnapshot reads the profile and the per-scene state ranking needs.
func (e *Engine) takeSnapshot(ctx context.Context, disabledPlugins []string) (*snapshot, error) {
	now := e.now()
	profile, ignored, err := e.profileAt(ctx, now)
	if err != nil {
		return nil, err
	}
	watched, err := e.watchStates(ctx)
	if err != nil {
		return nil, err
	}
	neighbors, err := e.dislikedNeighbors(ctx, profile.Disliked)
	if err != nil {
		return nil, err
	}
	return &snapshot{
		now:               now,
		profile:           profile,
		ignored:           ignored,
		watched:           watched,
		excluded:          recordIDSet(e.repo.ExcludedSceneIDs(ctx, disabledPlugins)),
		dislikedNeighbors: neighbors,
	}, nil
}

// eligible drops candidates that must never be recommended: blocked,
// inactive-plugin stubs and disliked scenes.
func (state *snapshot) eligible(candidates []*candidate) []*candidate {
	out := make([]*candidate, 0, len(candidates))
	for _, proposed := range candidates {
		if state.excluded[proposed.sceneID] || state.profile.Disliked[proposed.sceneID] {
			continue
		}
		out = append(out, proposed)
	}
	return out
}

// watchStates reads how far each watched scene got.
func (e *Engine) watchStates(ctx context.Context) (map[string]watchState, error) {
	rows, err := e.database.Query(ctx, watchStateQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("read watch states: %w", err)
	}
	out := make(map[string]watchState, len(rows))
	for _, row := range rows {
		out[rowString(row, "media")] = watchStateOf(rowInt(row, "progress"), rowInt(row, "duration"), db.AsBool(row["completed"]))
	}
	return out, nil
}

// watchStateOf classifies one watch as completed or partial.
func watchStateOf(progressSeconds, durationSeconds int, completed bool) watchState {
	if completed || watchedFraction(progressSeconds, durationSeconds) >= completedFraction {
		return watchCompleted
	}
	return watchPartial
}

// dislikedNeighbors reads the scenes that disliked scenes list as related.
func (e *Engine) dislikedNeighbors(ctx context.Context, disliked map[string]bool) (map[string]bool, error) {
	if len(disliked) == 0 {
		return map[string]bool{}, nil
	}
	ids, err := e.database.Strings(ctx,
		`SELECT related FROM scene_related WHERE scene IN (SELECT value FROM json_each($disliked))`,
		db.Vars{"disliked": setKeys(disliked)})
	if err != nil {
		return nil, fmt.Errorf("read disliked neighbors: %w", err)
	}
	return toSet(ids), nil
}

// signalSceneIDs lists the distinct scenes the signals are about.
func signalSceneIDs(signals []signal) []string {
	seen := map[string]bool{}
	for _, entry := range signals {
		seen[entry.sceneID] = true
	}
	return setKeys(seen)
}

// setKeys lists a set's members.
func setKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	return out
}

// recordIDSet turns record ids into a lookup set of "table:id" strings.
func recordIDSet(ids []db.RecordID) map[string]bool {
	out := make(map[string]bool, len(ids))
	for _, id := range ids {
		out[id.String()] = true
	}
	return out
}
