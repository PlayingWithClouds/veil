package recommend

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// sceneFeatures is what ranking knows about a scene.
type sceneFeatures struct {
	id         string
	title      string
	tags       []string
	performers []string
	studio     string
	// site is the plugin that first observed the scene.
	site        string
	publishedAt time.Time
}

// featureQuery reads the features of the scenes in $ids.
const featureQuery = `SELECT scene.id, scene.title, scene.tags, scene.performers, scene.studio,
	scene.date, scene.created_at,
	(SELECT plugin FROM observation WHERE observation.target = scene.id ORDER BY observed_at LIMIT 1) AS site
	FROM scene WHERE scene.id IN (SELECT value FROM json_each($ids))`

// loadFeatures reads the ranking features of the given scenes, keyed by id.
// Unknown ids are skipped.
func (e *Engine) loadFeatures(ctx context.Context, ids []string) (map[string]*sceneFeatures, error) {
	out := map[string]*sceneFeatures{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := e.database.Query(ctx, featureQuery, db.Vars{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("load scene features: %w", err)
	}
	for _, row := range rows {
		features := featuresFromRow(row)
		out[features.id] = features
	}
	return out, nil
}

// featuresFromRow decodes one featureQuery row.
func featuresFromRow(row db.Row) *sceneFeatures {
	return &sceneFeatures{
		id:          rowString(row, "id"),
		title:       rowString(row, "title"),
		tags:        rowStrings(row, "tags"),
		performers:  rowStrings(row, "performers"),
		studio:      rowString(row, "studio"),
		site:        rowString(row, "site"),
		publishedAt: publishedAt(row),
	}
}

// publishedAt prefers the release date and falls back to when the scene was
// first stored.
func publishedAt(row db.Row) time.Time {
	date := rowString(row, "date")
	if len(date) >= len("2006-01-02") {
		parsed, err := time.Parse("2006-01-02", date[:len("2006-01-02")])
		if err == nil {
			return parsed
		}
	}
	return rowTime(row, "created_at")
}
