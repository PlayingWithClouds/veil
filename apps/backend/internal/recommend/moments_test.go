package recommend

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestMomentTagsFindMarkersNearTheStopPoint(t *testing.T) {
	enders := []sessionEnder{{sceneID: "scene:a", stopSeconds: 600, durationSeconds: 1200}}
	markers := []marker{
		{sceneID: "scene:a", tagID: "tag:just-before", start: 540, end: 540},
		{sceneID: "scene:a", tagID: "tag:spanning", start: 400, end: 700},
		{sceneID: "scene:a", tagID: "tag:long-before", start: 100, end: 100},
		{sceneID: "scene:a", tagID: "tag:after", start: 700, end: 700},
		{sceneID: "scene:b", tagID: "tag:other-scene", start: 600, end: 600},
	}
	got := momentTags(enders, markers, nil)
	want := map[string][]string{"scene:a": {"tag:just-before", "tag:spanning"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("moment tags = %v, want %v", got, want)
	}
}

func TestMomentTagsAddTheReplayPeak(t *testing.T) {
	// Duration 1000 s: a bucket is 10 s. Bucket 80 (800-810 s) is the peak.
	heat := make([]float64, heatBucketCount)
	heat[80] = 50
	heat[10] = 1
	enders := []sessionEnder{{sceneID: "scene:a", stopSeconds: 200, durationSeconds: 1000}}
	markers := []marker{
		{sceneID: "scene:a", tagID: "tag:peak", start: 790, end: 790},
		{sceneID: "scene:a", tagID: "tag:elsewhere", start: 500, end: 500},
	}
	got := momentTags(enders, markers, map[string][]float64{"scene:a": heat})
	if want := []string{"tag:peak"}; !reflect.DeepEqual(got["scene:a"], want) {
		t.Errorf("moment tags = %v, want %v", got["scene:a"], want)
	}
}

func TestHeatPeakNeedsToStandOut(t *testing.T) {
	flat := make([]float64, heatBucketCount)
	for bucket := range flat {
		flat[bucket] = 1
	}
	if _, found := heatPeak(flat, 1000); found {
		t.Error("a flat heat has no peak")
	}
	if _, found := heatPeak(make([]float64, heatBucketCount), 1000); found {
		t.Error("an empty heat has no peak")
	}
	peaked := make([]float64, heatBucketCount)
	peaked[0] = 10
	if _, found := heatPeak(peaked, 0); found {
		t.Error("without a runtime the buckets have no seconds")
	}
}

func TestSessionEndersKeepsTheLatestPerSceneAndSceneLimit(t *testing.T) {
	signals := []signal{
		{sceneID: "scene:a", sessionEnd: true, at: testNow.Add(-2 * time.Hour), stopSeconds: 100},
		{sceneID: "scene:a", sessionEnd: true, at: testNow.Add(-time.Hour), stopSeconds: 300},
		{sceneID: "scene:b", sessionEnd: false, at: testNow},
	}
	got := sessionEnders(signals)
	if len(got) != 1 || got[0].sceneID != "scene:a" || got[0].stopSeconds != 300 {
		t.Errorf("session enders = %+v", got)
	}
}

func TestProfileBoostsMomentTagsOnlyForPositiveScenes(t *testing.T) {
	features := map[string]*sceneFeatures{
		"scene:a": {id: "scene:a", tags: []string{"tag:overall"}},
		"scene:b": {id: "scene:b", tags: []string{"tag:overall"}},
	}
	signals := []signal{
		{sceneID: "scene:a", weight: weightSessionEnd, at: testNow},
		{sceneID: "scene:b", weight: weightDislike, at: testNow},
	}
	moments := map[string][]string{"scene:a": {"tag:moment"}, "scene:b": {"tag:disliked-moment"}}

	profile := buildProfileWith(signals, features, nil, testNow, profileInputs{moments: moments})
	if profile.Tags["tag:moment"] <= 0 {
		t.Errorf("moment tag of a session ender must gain affinity, got %v", profile.Tags["tag:moment"])
	}
	if _, boosted := profile.Tags["tag:disliked-moment"]; boosted {
		t.Error("a disliked scene's moments must not be boosted")
	}
	// weightSessionEnd * momentTagBoost against the overall tag's weightSessionEnd + weightDislike.
	if profile.Tags["tag:moment"] <= profile.Tags["tag:overall"] {
		t.Errorf("moment %v should beat the overall tag %v here", profile.Tags["tag:moment"], profile.Tags["tag:overall"])
	}
}

func TestLoadMomentsReadsMarkersAndHeatFromTheDatabase(t *testing.T) {
	ctx := context.Background()
	engine, database, _ := newTestEngine(t)
	insertScene(t, database, "scene:a", map[string]any{"duration_seconds": 1000})
	insertRow(t, database, "scene_marker", "scene_marker:1", map[string]any{"media": "scene:a", "tag": "tag:stop", "seconds": 560})
	insertRow(t, database, "scene_marker", "scene_marker:2", map[string]any{"media": "scene:a", "tag": "tag:span", "seconds": 780, "end_seconds": 900})
	insertRow(t, database, "scene_marker", "scene_marker:3", map[string]any{"media": "scene:a", "label": "no tag", "seconds": 590})
	insertRow(t, database, "scene_marker", "scene_marker:4", map[string]any{"media": "scene:a", "tag": "tag:far", "seconds": 100})
	for bucket := 0; bucket < heatBucketCount; bucket++ {
		fields := map[string]any{"scene": "scene:a", "bucket": bucket, "watched": 1.0, "scrubbed": 0.0}
		if bucket == 80 {
			fields["watched"] = 40.0
		}
		if _, err := database.Exec(ctx, `INSERT INTO scene_heat (scene, bucket, watched, scrubbed) VALUES ($scene, $bucket, $watched, $scrubbed)`, fields); err != nil {
			t.Fatalf("insert heat: %v", err)
		}
	}

	got, err := engine.loadMoments(ctx, []sessionEnder{{sceneID: "scene:a", stopSeconds: 600, durationSeconds: 1000}})
	if err != nil {
		t.Fatalf("load moments: %v", err)
	}
	if want := []string{"tag:span", "tag:stop"}; !reflect.DeepEqual(got["scene:a"], want) {
		t.Errorf("moments = %v, want %v", got["scene:a"], want)
	}
}

func TestProfileUsesMarkersNearWhereTheSessionEnded(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertSessionEnder(t, engine, "scene:a", 2*time.Hour, map[string]any{"tags": []string{"tag:overall"}})
	insertRow(t, database, "scene_marker", "scene_marker:1", map[string]any{"media": "scene:a", "tag": "tag:moment", "seconds": 560})
	insertRow(t, database, "scene_marker", "scene_marker:2", map[string]any{"media": "scene:a", "tag": "tag:early", "seconds": 30})

	profile, err := engine.Profile(context.Background())
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if profile.Tags["tag:moment"] <= 0 || profile.Tags["tag:early"] != 0 {
		t.Errorf("tags = %v: the marker at the stop point must count, the early one not", profile.Tags)
	}
}
