package recommend

import (
	"context"
	"time"

	"github.com/playingwithclouds/veil/internal/schema"
)

// snapshot is everything one ranking pass reads about the user, taken once so
// every candidate is scored against the same state.
type snapshot struct {
	now       time.Time
	taste     *profile
	collected *signals
	// features describes the signalled items (the seeds among them).
	features map[schema.RecordRef]*mediaFeatures
	watched  map[schema.RecordRef]watchState
	// policy is what this ranking learned from past opens; nil = hand-set.
	policy *learnedPolicy
}

// profileAt builds the taste profile as of now, with the signals and the
// features of the signalled items it was built from.
func (e *Engine) profileAt(ctx context.Context, now time.Time) (*profile, *signals, map[schema.RecordRef]*mediaFeatures, error) {
	collected, err := e.collectSignals(ctx, now)
	if err != nil {
		return nil, nil, nil, err
	}
	features, err := e.loadFeatures(ctx, signalItems(collected))
	if err != nil {
		return nil, nil, nil, err
	}
	stats, err := e.libraryStats(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	blocked, err := e.loadBlocked(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	moments, err := e.loadMoments(ctx, sessionEnders(collected.items))
	if err != nil {
		return nil, nil, nil, err
	}
	inputs := profileInputs{features: features, blocked: blocked, stats: stats, moments: moments}
	return buildProfile(collected, inputs, now), collected, features, nil
}

// takeSnapshot reads the profile and the per-item state ranking needs.
func (e *Engine) takeSnapshot(ctx context.Context) (*snapshot, error) {
	now := e.now()
	taste, collected, features, err := e.profileAt(ctx, now)
	if err != nil {
		return nil, err
	}
	watched, err := e.loadWatchStates(ctx, collected.watches)
	if err != nil {
		return nil, err
	}
	policy, err := e.learnPolicy(ctx, now)
	if err != nil {
		return nil, err
	}
	return &snapshot{now: now, taste: taste, collected: collected, features: features, watched: watched, policy: policy}, nil
}
