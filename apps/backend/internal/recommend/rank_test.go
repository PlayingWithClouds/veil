package recommend

import "testing"

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
