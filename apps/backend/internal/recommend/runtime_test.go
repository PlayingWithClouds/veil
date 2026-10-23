package recommend

import "testing"

func TestRuntimeSharesInterpolateOnTheLogScale(t *testing.T) {
	lower, upper, ok := runtimeShares(60)
	if !ok || lower != 0 || upper != 0 {
		t.Errorf("under the first centre = (%d, %v, %v), want wholly band 0", lower, upper, ok)
	}
	lower, upper, _ = runtimeShares(int(runtimeBandSeconds[2]))
	if lower != 2 || upper > 0.001 {
		t.Errorf("on a centre = (%d, %v), want wholly that band", lower, upper)
	}
	lower, upper, _ = runtimeShares(10 * 3600)
	if lower != runtimeBands-2 || upper != 1 {
		t.Errorf("beyond the last centre = (%d, %v), want wholly the last band", lower, upper)
	}
	// The geometric middle of two centres splits evenly.
	lower, upper, _ = runtimeShares(509)
	if lower != 1 || upper < 0.49 || upper > 0.51 {
		t.Errorf("between 360 and 720 s = (%d, %v), want about half", lower, upper)
	}
	if _, _, ok := runtimeShares(0); ok {
		t.Error("an unknown runtime has no band")
	}
}

func TestProfileLearnsRuntimeFromSessionEndersAndDislikes(t *testing.T) {
	features := map[string]*sceneFeatures{
		"scene:loved":    {id: "scene:loved", durationSeconds: 1500},
		"scene:loved2":   {id: "scene:loved2", durationSeconds: 1500},
		"scene:disliked": {id: "scene:disliked", durationSeconds: 180},
		"scene:unknown":  {id: "scene:unknown"},
	}
	signals := []signal{
		{sceneID: "scene:loved", weight: weightSessionEnd, at: testNow},
		{sceneID: "scene:loved2", weight: weightSessionEnd, at: testNow},
		{sceneID: "scene:disliked", weight: weightDislike, at: testNow},
		{sceneID: "scene:unknown", weight: weightSessionEnd, at: testNow},
	}
	profile := buildProfile(signals, features, nil, testNow)

	assertClose(t, "loved runtime", profile.runtimeAffinity(1500), 1)
	if disliked := profile.runtimeAffinity(180); disliked >= 0 {
		t.Errorf("a disliked runtime must be negative, got %v", disliked)
	}
	assertClose(t, "unknown runtime is neutral", profile.runtimeAffinity(0), 0)
	if between := profile.runtimeAffinity(1000); between <= 0 || between >= 1 {
		t.Errorf("a runtime between bands interpolates, got %v", between)
	}
}

func TestRuntimePreferenceStaysWeakOnThinEvidence(t *testing.T) {
	features := map[string]*sceneFeatures{"scene:one": {id: "scene:one", durationSeconds: 1500}}
	signals := []signal{{sceneID: "scene:one", weight: weightClick, at: testNow}}
	profile := buildProfile(signals, features, nil, testNow)
	if affinity := profile.runtimeAffinity(1500); affinity <= 0 || affinity > 0.1 {
		t.Errorf("one click should leave a faint preference, got %v", affinity)
	}
}

func TestRuntimeIgnoresShownButNotClicked(t *testing.T) {
	features := map[string]*sceneFeatures{"scene:skipped": {id: "scene:skipped", durationSeconds: 5400}}
	signals := []signal{{sceneID: "scene:skipped", weight: weightIgnoredImpression * 3, at: testNow, weak: true}}
	profile := buildProfile(signals, features, nil, testNow)
	assertClose(t, "ignored impressions teach nothing about runtime", profile.runtimeAffinity(5400), 0)
}

func TestAffinityMatchAddsRuntimeAndStaysNeutralWithoutOne(t *testing.T) {
	profile := newProfile(1)
	profile.runtime[3] = 1
	with := affinityMatch(&sceneFeatures{durationSeconds: int(runtimeBandSeconds[3])}, profile)
	without := affinityMatch(&sceneFeatures{}, profile)
	assertClose(t, "runtime term", with-without, matchWeightRuntime)
	assertClose(t, "no runtime", without, 0)
}
