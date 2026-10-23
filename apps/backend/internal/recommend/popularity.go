package recommend

import (
	"context"
	"fmt"
	"sort"
)

// popularityIndex holds what each site's scenes report as views and ratings,
// sorted, so a scene can be placed among its site's other scenes.
type popularityIndex struct {
	views   map[string][]float64
	ratings map[string][]float64
}

// popularityQuery reads the views and rating of every scene that has either,
// with the site (first observing plugin) that reported them.
const popularityQuery = `SELECT
	(SELECT plugin FROM observation WHERE observation.target = scene.id ORDER BY observed_at LIMIT 1) AS site,
	scene.view_count, scene.rating
	FROM scene WHERE scene.view_count > 0 OR scene.rating IS NOT NULL`

// loadPopularity reads the per-site distributions of views and ratings.
func (e *Engine) loadPopularity(ctx context.Context) (*popularityIndex, error) {
	rows, err := e.database.Query(ctx, popularityQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("load popularity: %w", err)
	}
	index := &popularityIndex{views: map[string][]float64{}, ratings: map[string][]float64{}}
	for _, row := range rows {
		site := rowString(row, "site")
		if site == "" {
			continue
		}
		if views := rowInt(row, "view_count"); views > 0 {
			index.views[site] = append(index.views[site], float64(views))
		}
		if row["rating"] != nil {
			index.ratings[site] = append(index.ratings[site], rowFloat(row, "rating"))
		}
	}
	for _, values := range index.views {
		sort.Float64s(values)
	}
	for _, values := range index.ratings {
		sort.Float64s(values)
	}
	return index, nil
}

// of is the scene's standing among its site's scenes in [0, 1]: the views
// percentile blended with the rating percentile, or whichever of the two the
// scene and the site's sample support, else popularityNeutral. Safe on a nil
// index.
func (index *popularityIndex) of(scene *sceneFeatures) float64 {
	if index == nil {
		return popularityNeutral
	}
	viewStanding, hasViews := percentile(index.views[scene.site], float64(scene.viewCount), scene.viewCount > 0)
	ratingStanding, hasRating := percentile(index.ratings[scene.site], scene.rating, scene.rated)
	switch {
	case hasViews && hasRating:
		return popularityViewShare*viewStanding + (1-popularityViewShare)*ratingStanding
	case hasViews:
		return viewStanding
	case hasRating:
		return ratingStanding
	}
	return popularityNeutral
}

// percentile places value among the sorted sample: the share of the sample
// below it, ties counting half. ok is false when the scene has no value or
// the sample is too small to compare with.
func percentile(sample []float64, value float64, present bool) (standing float64, ok bool) {
	if !present || len(sample) < popularityMinSample {
		return 0, false
	}
	below := sort.SearchFloat64s(sample, value)
	notAbove := sort.Search(len(sample), func(position int) bool { return sample[position] > value })
	return float64(below+notAbove) / 2 / float64(len(sample)), true
}
