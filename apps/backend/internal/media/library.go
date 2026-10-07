package media

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
)

// DownloadedScenes returns scenes that have a completed download job, newest
// download first. Downloads are library-global.
func (r *Repository) DownloadedScenes(ctx context.Context) ([]*model.Scene, error) {
	targets, err := r.database.Strings(ctx,
		`SELECT coalesce(target, json_extract(payload, '$.media')) AS scene FROM job
		 WHERE kind = 'download' AND status = 'completed'
		   AND coalesce(target, json_extract(payload, '$.media')) IS NOT NULL
		 ORDER BY coalesce(finished_at, updated_at) DESC`,
		nil)
	if err != nil {
		return nil, fmt.Errorf("downloaded scenes: %w", err)
	}
	seen := make(map[string]bool, len(targets))
	out := make([]*model.Scene, 0, len(targets))
	for _, id := range targets {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		scene, err := r.GetScene(ctx, id)
		if err != nil || scene == nil {
			continue
		}
		out = append(out, scene)
	}
	return out, nil
}
