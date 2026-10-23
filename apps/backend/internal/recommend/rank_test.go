package recommend

import (
	"context"
	"testing"
	"time"
)

func TestUndatedFreshnessIsTheDatedMedian(t *testing.T) {
	features := map[string]*sceneFeatures{
		"scene:new":     {publishedAt: testNow},
		"scene:old":     {publishedAt: testNow.Add(-2 * freshnessHalfLife)},
		"scene:middle":  {publishedAt: testNow.Add(-freshnessHalfLife)},
		"scene:undated": {},
	}
	assertClose(t, "median", undatedFreshness(features, testNow), 0.5)
	assertClose(t, "no dated scene", undatedFreshness(map[string]*sceneFeatures{"scene:undated": {}}, testNow), undatedFreshnessDefault)
}

func TestNewestCorroboratesNothing(t *testing.T) {
	pool := newCandidatePool()
	pool.add(candidate{sceneID: "scene:a", source: SourceCatalog, prior: 0.3})
	pool.add(candidate{sceneID: "scene:a", source: SourceNewest, prior: 0.1})
	pool.add(candidate{sceneID: "scene:b", source: SourceCatalog, prior: 0.3})
	pool.add(candidate{sceneID: "scene:b", source: SourceSearch, prior: 0.4})
	if count := pool.byScene["scene:a"].sourceCount; count != 1 {
		t.Errorf("catalog + newest count %d sources, want 1", count)
	}
	if count := pool.byScene["scene:b"].sourceCount; count != 2 {
		t.Errorf("catalog + search count %d sources, want 2", count)
	}
}

func TestUndatedSceneGetsTheNeutralFreshness(t *testing.T) {
	assertClose(t, "undated", sceneFreshness(&sceneFeatures{}, testNow, 0.7), 0.7)
	assertClose(t, "dated", sceneFreshness(&sceneFeatures{publishedAt: testNow.Add(-freshnessHalfLife)}, testNow, 0.7), 0.5)
}

// ascendingSample is the values 1..count as floats.
func ascendingSample(count int) []float64 {
	sample := make([]float64, count)
	for index := range sample {
		sample[index] = float64(index + 1)
	}
	return sample
}

func TestPercentilePlacesAValueInItsSample(t *testing.T) {
	sample := ascendingSample(30)
	top, _ := percentile(sample, 30, true)
	bottom, _ := percentile(sample, 1, true)
	assertClose(t, "top", top, 59.0/60)
	assertClose(t, "bottom", bottom, 1.0/60)
	tied, _ := percentile([]float64{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5}, 5, true)
	assertClose(t, "all tied", tied, 0.5)
	if _, ok := percentile(sample[:popularityMinSample-1], 3, true); ok {
		t.Error("a sample below popularityMinSample must not rank anything")
	}
	if _, ok := percentile(sample, 0, false); ok {
		t.Error("a scene without a value has no percentile")
	}
}

func TestPopularityBlendsViewsAndRatingPerSite(t *testing.T) {
	index := &popularityIndex{
		views:   map[string][]float64{"site": ascendingSample(30), "tiny": ascendingSample(3)},
		ratings: map[string][]float64{"site": ascendingSample(30)},
	}
	viewed := &sceneFeatures{site: "site", viewCount: 30}
	rated := &sceneFeatures{site: "site", rating: 1, rated: true}
	both := &sceneFeatures{site: "site", viewCount: 30, rating: 1, rated: true}
	assertClose(t, "views only", index.of(viewed), 59.0/60)
	assertClose(t, "rating only", index.of(rated), 1.0/60)
	assertClose(t, "blend", index.of(both), popularityViewShare*59.0/60+(1-popularityViewShare)*1.0/60)
	assertClose(t, "no data", index.of(&sceneFeatures{site: "site"}), popularityNeutral)
	assertClose(t, "thin site", index.of(&sceneFeatures{site: "tiny", viewCount: 3}), popularityNeutral)
	assertClose(t, "other site", index.of(&sceneFeatures{site: "elsewhere", viewCount: 30}), popularityNeutral)
	var missing *popularityIndex
	assertClose(t, "no index", missing.of(viewed), popularityNeutral)
}

func TestPopularSceneOutscoresAnUnknownOneWithoutAffinity(t *testing.T) {
	state := &snapshot{now: testNow, profile: newProfile(0), popularity: &popularityIndex{views: map[string][]float64{"site": ascendingSample(30)}}}
	proposed := &candidate{sceneID: "scene:a", source: SourceCatalog, prior: 0.3, sourceCount: 1}
	popular := &sceneFeatures{site: "site", viewCount: 30}
	obscure := &sceneFeatures{site: "site", viewCount: 1}
	gap := score(proposed, popular, state, 0.5) - score(proposed, obscure, state, 0.5)
	assertClose(t, "gap", gap, rankWeightPopularity*(59.0-1.0)/60)
}

func TestLoadPopularityGroupsByFirstObservingPlugin(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertScene(t, database, "scene:viewed", map[string]any{"view_count": 120})
	insertScene(t, database, "scene:rated", map[string]any{"rating": 7.5})
	insertScene(t, database, "scene:plain", map[string]any{})
	for index, target := range []string{"scene:viewed", "scene:rated", "scene:plain"} {
		insertRow(t, database, "observation", "observation:"+target[len("scene:"):], map[string]any{
			"target": target, "plugin": "eporner", "observed_at": ago(time.Duration(index+1) * time.Hour),
		})
	}
	index, err := engine.loadPopularity(context.Background())
	if err != nil {
		t.Fatalf("load popularity: %v", err)
	}
	if len(index.views["eporner"]) != 1 || index.views["eporner"][0] != 120 {
		t.Errorf("views = %v", index.views)
	}
	if len(index.ratings["eporner"]) != 1 || index.ratings["eporner"][0] != 7.5 {
		t.Errorf("ratings = %v", index.ratings)
	}
}
