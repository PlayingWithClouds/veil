package media

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
)

func TestSceneHeatAccumulatesAcrossSessionsAndPicksTheScrubbedMoment(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	insertRow(t, database, "scene", "scene:heat1", map[string]any{"title": "Heat", "source_url": "https://site/heat1", "duration_seconds": 1000})

	session := model.RecordSceneHeatInput{
		SceneID:         "scene:heat1",
		DurationSeconds: 1000,
		Spans:           []*model.HeatSpanInput{{FromSeconds: 200, ToSeconds: 300}},
		Scrubs:          []float64{250},
	}
	for range 3 {
		if err := repo.RecordSceneHeat(ctx, session); err != nil {
			t.Fatalf("record: %v", err)
		}
	}

	heat, err := repo.SceneHeat(ctx, "scene:heat1")
	if err != nil {
		t.Fatalf("heat: %v", err)
	}
	if len(heat.Buckets) != heatBuckets {
		t.Fatalf("got %d buckets, want %d", len(heat.Buckets), heatBuckets)
	}
	if heat.Buckets[25] != 1 {
		t.Errorf("scrubbed bucket = %v, want the peak, 1", heat.Buckets[25])
	}
	if heat.BestMomentSeconds == nil || *heat.BestMomentSeconds != 255 {
		t.Errorf("best moment = %v, want 255", heat.BestMomentSeconds)
	}
	if heat.ThumbnailSeconds != nil {
		t.Errorf("thumbnail seconds = %v before one was set", *heat.ThumbnailSeconds)
	}
}

func TestSceneThumbnailReplacesThePosterInTheScene(t *testing.T) {
	ctx := context.Background()
	repo, database := newTestRepository(t)
	insertRow(t, database, "scene", "scene:thumb1", map[string]any{"title": "Thumb", "source_url": "https://site/thumb1", "poster_path": "https://site/poster.jpg"})

	if err := repo.SetSceneThumbnail(ctx, "scene:thumb1", "/api/blob/images/scene-thumbnails/scene-thumb1-90.jpg", 90); err != nil {
		t.Fatalf("set thumbnail: %v", err)
	}

	scene, err := repo.GetScene(ctx, "scene:thumb1")
	if err != nil || scene == nil {
		t.Fatalf("get scene: %v %v", scene, err)
	}
	if scene.PosterPath == nil || *scene.PosterPath != "/api/blob/images/scene-thumbnails/scene-thumb1-90.jpg" {
		t.Errorf("poster = %v, want the thumbnail", scene.PosterPath)
	}
	heat, err := repo.SceneHeat(ctx, "scene:thumb1")
	if err != nil {
		t.Fatalf("heat: %v", err)
	}
	if heat.ThumbnailSeconds == nil || *heat.ThumbnailSeconds != 90 {
		t.Errorf("thumbnail seconds = %v, want 90", heat.ThumbnailSeconds)
	}
}
