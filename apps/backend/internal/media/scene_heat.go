package media

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// RecordSceneHeat adds one viewing session's playback spans and scrub targets to the scene's heat.
func (r *Repository) RecordSceneHeat(ctx context.Context, input model.RecordSceneHeatInput) error {
	scene, err := parseMediaID(input.SceneID)
	if err != nil {
		return err
	}
	increments := heatIncrements(input.DurationSeconds, input.Spans, input.Scrubs)
	return r.database.Tx(ctx, func(tx *db.DB) error {
		for _, increment := range increments {
			_, err := tx.Exec(ctx,
				`INSERT INTO scene_heat (scene, bucket, watched, scrubbed) VALUES ($scene, $bucket, $watched, $scrubbed)
				 ON CONFLICT (scene, bucket) DO UPDATE SET watched = watched + excluded.watched, scrubbed = scrubbed + excluded.scrubbed`,
				db.Vars{"scene": *scene, "bucket": increment.bucket, "watched": increment.watched, "scrubbed": increment.scrubbed})
			if err != nil {
				return fmt.Errorf("record scene heat: %w", err)
			}
		}
		return nil
	})
}

// SceneHeat returns the scene's replay graph, its best moment and the moment its thumbnail was taken at.
func (r *Repository) SceneHeat(ctx context.Context, sceneID string) (*model.SceneHeat, error) {
	scene, err := parseMediaID(sceneID)
	if err != nil {
		return nil, err
	}
	watched, scrubbed, err := r.heatTotals(ctx, *scene)
	if err != nil {
		return nil, err
	}
	heat := &model.SceneHeat{Buckets: []float64{}}
	if buckets := replayBuckets(watched, scrubbed); buckets != nil {
		heat.Buckets = buckets
	}
	row, err := r.database.QueryRow(ctx,
		`SELECT thumbnail_seconds,
		        COALESCE(duration_seconds, (SELECT duration_seconds FROM watch_history WHERE media = scene.id LIMIT 1)) AS duration
		   FROM scene WHERE id = $scene`, db.Vars{"scene": *scene})
	if err != nil {
		return nil, fmt.Errorf("scene heat: %w", err)
	}
	if row == nil {
		return heat, nil
	}
	heat.ThumbnailSeconds = mFloatPtr(row, "thumbnail_seconds")
	duration := mFloatPtr(row, "duration")
	if duration == nil {
		return heat, nil
	}
	markers, err := r.database.Query(ctx, `SELECT seconds FROM scene_marker WHERE media = $scene`, db.Vars{"scene": *scene})
	if err != nil {
		return nil, fmt.Errorf("scene heat markers: %w", err)
	}
	markerSeconds := make([]float64, 0, len(markers))
	for _, marker := range markers {
		if seconds := mFloatPtr(marker, "seconds"); seconds != nil {
			markerSeconds = append(markerSeconds, *seconds)
		}
	}
	heat.BestMomentSeconds = bestMomentSeconds(scrubbed, markerSeconds, *duration)
	return heat, nil
}

// heatTotals reads the stored per-bucket totals as two dense slices of heatBuckets values.
func (r *Repository) heatTotals(ctx context.Context, scene db.RecordID) (watched, scrubbed []float64, err error) {
	watched = make([]float64, heatBuckets)
	scrubbed = make([]float64, heatBuckets)
	rows, err := r.database.Query(ctx, `SELECT bucket, watched, scrubbed FROM scene_heat WHERE scene = $scene`, db.Vars{"scene": scene})
	if err != nil {
		return nil, nil, fmt.Errorf("scene heat: %w", err)
	}
	for _, row := range rows {
		bucket := db.AsInt(row["bucket"])
		if bucket < 0 || bucket >= heatBuckets {
			continue
		}
		if value := mFloatPtr(row, "watched"); value != nil {
			watched[bucket] = *value
		}
		if value := mFloatPtr(row, "scrubbed"); value != nil {
			scrubbed[bucket] = *value
		}
	}
	return watched, scrubbed, nil
}

// SetSceneThumbnail points the scene's listing thumbnail at an already stored frame taken at atSeconds.
func (r *Repository) SetSceneThumbnail(ctx context.Context, sceneID, thumbnailPath string, atSeconds float64) error {
	scene, err := parseMediaID(sceneID)
	if err != nil {
		return err
	}
	_, err = r.database.Exec(ctx,
		`UPDATE scene SET thumbnail_path = $path, thumbnail_seconds = $seconds WHERE id = $scene`,
		db.Vars{"scene": *scene, "path": thumbnailPath, "seconds": atSeconds})
	if err != nil {
		return fmt.Errorf("set scene thumbnail: %w", err)
	}
	return nil
}
