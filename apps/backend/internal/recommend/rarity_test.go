package recommend

import (
	"context"
	"fmt"
	"testing"
)

func TestTagRarityWeighsGenericTagsDown(t *testing.T) {
	rarity := &tagRarity{taggedScenes: 1000, sceneCounts: map[string]int{
		"tag:everywhere": 900, "tag:common": 200, "tag:rare": 20, "tag:single": 1,
	}}
	if weight := rarity.weight("tag:single"); weight != 1 {
		t.Errorf("single-scene tag = %v, want 1", weight)
	}
	if weight := rarity.weight("tag:rare"); weight != 1 {
		t.Errorf("tag under rarityFullShare = %v, want 1", weight)
	}
	if weight := rarity.weight("tag:unseen"); weight != 1 {
		t.Errorf("tag on no scene = %v, want 1", weight)
	}
	common, everywhere := rarity.weight("tag:common"), rarity.weight("tag:everywhere")
	if !(1 > common && common > everywhere && everywhere >= rarityFloor) {
		t.Errorf("weights must fall with the share: common %v, everywhere %v", common, everywhere)
	}
	if everywhere > 0.1 {
		t.Errorf("a tag on 90%% of the scenes should barely count, got %v", everywhere)
	}
}

func TestTagRarityIsNeutralWithoutEnoughScenes(t *testing.T) {
	var missing *tagRarity
	if weight := missing.weight("tag:a"); weight != 1 {
		t.Errorf("nil rarity = %v, want 1", weight)
	}
	small := &tagRarity{taggedScenes: rarityMinScenes - 1, sceneCounts: map[string]int{"tag:a": rarityMinScenes - 1}}
	if weight := small.weight("tag:a"); weight != 1 {
		t.Errorf("small library = %v, want 1", weight)
	}
}

func TestProfileSpreadsLessWeightOntoGenericTags(t *testing.T) {
	rarity := &tagRarity{taggedScenes: 1000, sceneCounts: map[string]int{"tag:generic": 800, "tag:niche": 10}}
	features := map[string]*sceneFeatures{"scene:a": {id: "scene:a", tags: []string{"tag:generic", "tag:niche"}}}
	signals := []signal{{sceneID: "scene:a", weight: weightLike, at: testNow}}

	profile := buildProfileWith(signals, features, nil, testNow, profileInputs{rarity: rarity})
	assertClose(t, "niche tag is the strongest, so 1", profile.Tags["tag:niche"], 1)
	if profile.Tags["tag:generic"] > 0.2 {
		t.Errorf("generic tag affinity = %v, want it near zero", profile.Tags["tag:generic"])
	}
	plain := buildProfile(signals, features, nil, testNow)
	assertClose(t, "without rarity both tags weigh the same", plain.Tags["tag:generic"], plain.Tags["tag:niche"])
}

func TestTagAffinityWeighsByRarityButNotBlocked(t *testing.T) {
	rarity := &tagRarity{taggedScenes: 1000, sceneCounts: map[string]int{"tag:generic": 800, "tag:niche": 10, "tag:blocked": 900}}
	affinities := map[string]float64{"tag:generic": 1, "tag:niche": 1, "tag:blocked": affinityBlocked}

	generic := tagAffinity([]string{"tag:generic"}, affinities, rarity)
	niche := tagAffinity([]string{"tag:niche"}, affinities, rarity)
	if !(niche > 5*generic) {
		t.Errorf("niche %v must dwarf generic %v", niche, generic)
	}
	assertClose(t, "blocked tag keeps full weight", tagAffinity([]string{"tag:blocked"}, affinities, rarity), affinityBlocked)
	assertClose(t, "no rarity, one tag", tagAffinity([]string{"tag:generic"}, affinities, nil), 1)
	assertClose(t, "no tags", tagAffinity(nil, affinities, rarity), 0)
}

func TestLibraryStatsCountsTaggedScenesAndIsReused(t *testing.T) {
	ctx := context.Background()
	engine, database, _ := newTestEngine(t)
	for index := 0; index < 3; index++ {
		insertScene(t, database, fmt.Sprintf("scene:tagged%d", index), map[string]any{"tags": []string{"tag:a", "tag:b"}})
	}
	insertScene(t, database, "scene:untagged", map[string]any{})
	insertScene(t, database, "scene:other", map[string]any{"tags": []string{"tag:a"}})

	stats, err := engine.libraryStats(ctx)
	if err != nil {
		t.Fatalf("library stats: %v", err)
	}
	if stats.rarity.taggedScenes != 4 || stats.rarity.sceneCounts["tag:a"] != 4 || stats.rarity.sceneCounts["tag:b"] != 3 {
		t.Errorf("counts = %+v", stats.rarity)
	}
	insertScene(t, database, "scene:later", map[string]any{"tags": []string{"tag:a"}})
	again, err := engine.libraryStats(ctx)
	if err != nil {
		t.Fatalf("library stats: %v", err)
	}
	if again != stats {
		t.Error("stats must be reused inside libraryStatsTTL")
	}
}
