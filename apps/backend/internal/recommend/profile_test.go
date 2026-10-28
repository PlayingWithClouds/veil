package recommend

import (
	"context"
	"testing"
	"time"
)

func TestDecayFactorHalvesEveryHalfLife(t *testing.T) {
	assertClose(t, "now", decayFactor(0, AffinityHalfLife), 1)
	assertClose(t, "one half-life", decayFactor(AffinityHalfLife, AffinityHalfLife), 0.5)
	assertClose(t, "two half-lives", decayFactor(2*AffinityHalfLife, AffinityHalfLife), 0.25)
	assertClose(t, "future", decayFactor(-time.Hour, AffinityHalfLife), 1)
}

func TestWatchWeightClassifiesWatches(t *testing.T) {
	cases := []struct {
		name        string
		played      int
		duration    int
		followed    bool
		sinceUpdate time.Duration
		wantWeight  float64
		wantCounts  bool
	}{
		{"ended the session mid-video", 300, 1200, false, time.Hour, weightSessionEnd, true},
		{"ended the session at the end", 1200, 1200, false, time.Hour, weightSessionEnd, true},
		{"played in full, then another", 1200, 1200, true, time.Hour, weightCompletedWatch, true},
		{"played 80% of a short one, then another", 200, 240, true, time.Hour, weightCompletedWatch, true},
		{"skimmed five minutes of a long one, then another", 300, 3600, true, time.Hour, weightCompletedWatch, true},
		{"partial, then another", 200, 1200, true, time.Hour, weightPartialWatch, true},
		{"session still open", 200, 1200, false, 10 * time.Minute, weightPartialWatch, true},
		{"last, but under a minute", 40, 600, false, time.Hour, weightPartialWatch, true},
		{"under ten seconds", 8, 600, false, time.Hour, weightEarlyAbandon, true},
		{"under five percent", 20, 600, true, time.Hour, weightEarlyAbandon, true},
		{"a minute and a half of a long one", 90, 3600, true, time.Hour, weightPartialWatch, true},
		{"unknown runtime, then another", 20, 0, true, time.Hour, weightPartialWatch, true},
		{"just started", 2, 600, false, time.Minute, 0, false},
	}
	for _, testCase := range cases {
		entry := watch{playedSeconds: testCase.played, durationSeconds: testCase.duration}
		weight, counts := watchWeight(entry, testCase.followed, testCase.sinceUpdate)
		if weight != testCase.wantWeight || counts != testCase.wantCounts {
			t.Errorf("%s: got (%v, %v), want (%v, %v)", testCase.name, weight, counts, testCase.wantWeight, testCase.wantCounts)
		}
	}
}

func TestFollowedWatchesFindsTheSessionsLastWatch(t *testing.T) {
	base := time.Date(2026, 10, 1, 20, 0, 0, 0, time.UTC)
	minutes := func(count int) time.Time { return base.Add(time.Duration(count) * time.Minute) }
	watches := []watch{
		// Followed by "second", which starts ten minutes after it stopped.
		{sceneID: "scene:first", startedAt: minutes(0), lastActivityAt: minutes(20)},
		// Last of the evening.
		{sceneID: "scene:second", startedAt: minutes(30), lastActivityAt: minutes(50)},
		// The next day: too late to follow "second".
		{sceneID: "scene:tomorrow", startedAt: minutes(24 * 60), lastActivityAt: minutes(24*60 + 15)},
		// A row reused by a replay: it was first started long ago, but its
		// last activity follows "early" within the gap.
		{sceneID: "scene:early", startedAt: minutes(-600), lastActivityAt: minutes(-590)},
		{sceneID: "scene:replayed", startedAt: minutes(-2000), lastActivityAt: minutes(-570)},
	}
	got := followedWatches(watches)
	want := []bool{true, false, false, true, false}
	for index := range want {
		if got[index] != want[index] {
			t.Errorf("%s: followed = %v, want %v", watches[index].sceneID, got[index], want[index])
		}
	}
}

func TestProfileWeighsDecaysAndBlocks(t *testing.T) {
	ctx := context.Background()
	engine, database, repo := newTestEngine(t)
	insertScene(t, database, "scene:oed", map[string]any{"performers": []string{"performer:oed"}, "tags": []string{"tag:a", "tag:b"}})
	insertScene(t, database, "scene:liked", map[string]any{"performers": []string{"performer:liked"}})
	insertScene(t, database, "scene:disliked", map[string]any{"performers": []string{"performer:disliked"}})
	insertScene(t, database, "scene:abandoned", map[string]any{"performers": []string{"performer:abandoned"}})
	insertRow(t, database, "o_event", "o_event:1", map[string]any{"media": "scene:oed", "created_at": ago(0)})
	// A like two half-lives old keeps a quarter of its weight.
	insertRow(t, database, "user_rating", "user_rating:1", map[string]any{"media": "scene:liked", "rating": 10, "updated_at": ago(2 * AffinityHalfLife)})
	insertRow(t, database, "user_rating", "user_rating:2", map[string]any{"media": "scene:disliked", "rating": 1, "updated_at": ago(0)})
	insertRow(t, database, "watch_history", "watch_history:1", map[string]any{
		"media": "scene:abandoned", "progress_seconds": 5, "max_progress_seconds": 5, "duration_seconds": 600, "updated_at": ago(time.Hour),
	})
	if _, err := repo.AddBlock(ctx, "tag", "tag:blocked", nil); err != nil {
		t.Fatalf("block: %v", err)
	}

	profile, err := engine.Profile(ctx)
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	// Normalized by the strongest performer (the o-event, weight 5).
	assertClose(t, "o-event performer", profile.Performers["performer:oed"], 1)
	assertClose(t, "decayed like", profile.Performers["performer:liked"], weightLike*0.25/weightOEvent)
	assertClose(t, "dislike", profile.Performers["performer:disliked"], weightDislike/weightOEvent)
	assertClose(t, "early abandon", profile.Performers["performer:abandoned"], weightEarlyAbandon*decayFactor(time.Hour, AffinityHalfLife)/weightOEvent)
	assertClose(t, "blocked tag", profile.Tags["tag:blocked"], affinityBlocked)
	assertClose(t, "shared tag weight", profile.Tags["tag:a"], 1)
	if !profile.Disliked["scene:disliked"] {
		t.Error("thumbs-down scene must be marked disliked")
	}
	if len(profile.Seeds) != 1 || profile.Seeds[0].SceneID != "scene:oed" {
		t.Errorf("only the o-event scene is strong enough to seed, got %+v", profile.Seeds)
	}
}

func TestAdjacentTagsCoOccurWithTopTags(t *testing.T) {
	profile := newProfile(1)
	// Five top tags; tag:near is the sixth strongest, just outside them.
	profile.Tags = map[string]float64{
		"tag:top": 1, "tag:top2": 0.9, "tag:top3": 0.8, "tag:top4": 0.7, "tag:top5": 0.6,
		"tag:near": 0.1, "tag:disliked": -0.5,
	}
	features := map[string]*sceneFeatures{
		"scene:liked":   {id: "scene:liked", tags: []string{"tag:top", "tag:near", "tag:disliked"}},
		"scene:unliked": {id: "scene:unliked", tags: []string{"tag:top", "tag:far"}},
	}
	weights := map[string]float64{"scene:liked": 2, "scene:unliked": -1}

	adjacent := adjacentTags(profile, weights, features)
	if len(adjacent) != 1 || adjacent[0] != "tag:near" {
		t.Errorf("adjacent = %v, want [tag:near]", adjacent)
	}
}
