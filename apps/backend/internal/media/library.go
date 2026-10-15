package media

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/ingest"
)

// DownloadedScenes returns the scenes stored locally, newest first: those
// with a completed download job, and cam recordings, which are written to the
// blob store directly and never go through a download job. Library-global.
func (r *Repository) DownloadedScenes(ctx context.Context) ([]*model.Scene, error) {
	targets, err := r.database.Strings(ctx,
		`SELECT scene FROM (
		   SELECT coalesce(target, json_extract(payload, '$.media')) AS scene,
		          coalesce(finished_at, updated_at) AS stored_at
		   FROM job
		   WHERE kind = 'download' AND status = 'completed'
		     AND coalesce(target, json_extract(payload, '$.media')) IS NOT NULL
		   UNION ALL
		   SELECT id AS scene, created_at AS stored_at FROM scene
		   WHERE substr(source_url, 1, length($recordingPrefix)) = $recordingPrefix
		 )
		 ORDER BY stored_at DESC`,
		db.Vars{"recordingPrefix": ingest.RecordingSourcePrefix})
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
