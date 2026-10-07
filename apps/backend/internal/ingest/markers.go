package ingest

import (
	"context"
	"log"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// syncSceneMarkersLogged writes plugin-emitted markers as global (personal = 0)
// scene_marker rows created by the plugin. Best-effort: failures are logged,
// not propagated.
func (s *Service) syncSceneMarkersLogged(ctx context.Context, canonicalID, pluginName string, markers []plugins.SceneMarkerInput) {
	if len(markers) == 0 {
		return
	}
	sceneRID, err := db.ParseRecordID(canonicalID)
	if err != nil {
		log.Printf("ingest: invalid scene id %q: %v", canonicalID, err)
		return
	}

	pluginRID := s.pluginRecordID(ctx, pluginName)
	for _, marker := range markers {
		if err := s.upsertGlobalMarker(ctx, *sceneRID, pluginRID, marker); err != nil {
			log.Printf("ingest: marker on %s at %.1fs: %v", canonicalID, marker.Seconds, err)
		}
	}
}

// upsertGlobalMarker creates one global marker unless an equivalent already
// exists (same scene, same second, same tag/label).
func (s *Service) upsertGlobalMarker(ctx context.Context, sceneRID, pluginRID db.RecordID, marker plugins.SceneMarkerInput) error {
	tagName := strings.TrimSpace(marker.Tag)
	label := strings.TrimSpace(marker.Label)
	if tagName == "" && label == "" {
		return nil
	}

	var tagRID *db.RecordID
	if tagName != "" {
		id, err := s.findOrCreateNamed(ctx, "tag", tagName)
		if err != nil {
			return err
		}
		tagRID = &id
	}

	exists, err := s.markerExists(ctx, sceneRID, tagRID, label, marker.Seconds)
	if err != nil || exists {
		return err
	}

	row := map[string]any{
		"media":      sceneRID,
		"personal":   false,
		"seconds":    marker.Seconds,
		"created_by": pluginRID,
	}
	if tagRID != nil {
		row["tag"] = *tagRID
	}
	if label != "" {
		row["label"] = label
	}
	if marker.EndSeconds != nil {
		row["end_seconds"] = *marker.EndSeconds
	}
	_, err = s.database.Insert(ctx, "scene_marker", row)
	return err
}

// markerExists checks for a global marker with the same identity.
func (s *Service) markerExists(ctx context.Context, sceneRID db.RecordID, tagRID *db.RecordID, label string, seconds float64) (bool, error) {
	condition := "media = $media AND personal = 0 AND seconds = $seconds"
	params := db.Vars{"media": sceneRID, "seconds": seconds}
	if tagRID != nil {
		condition += " AND tag = $tag"
		params["tag"] = *tagRID
	} else {
		condition += " AND label = $label"
		params["label"] = label
	}
	count, err := s.database.Int(ctx, "SELECT count(*) FROM scene_marker WHERE "+condition, params)
	return count > 0, err
}
