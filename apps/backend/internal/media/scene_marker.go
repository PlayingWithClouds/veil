package media

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// sceneMarkerRecord mirrors the `scene_marker` table.
type sceneMarkerRecord struct {
	ID         *db.RecordID `json:"id,omitempty"`
	Media      *db.RecordID `json:"media,omitempty"`
	Personal   bool         `json:"personal"`
	Tag        *db.RecordID `json:"tag,omitempty"`
	Label      *string      `json:"label,omitempty"`
	Seconds    float64      `json:"seconds"`
	EndSeconds *float64     `json:"end_seconds,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
}

func (r *Repository) markerToModel(ctx context.Context, m sceneMarkerRecord) *model.SceneMarker {
	marker := &model.SceneMarker{
		ID:         recordIDString(m.ID),
		Label:      m.Label,
		Seconds:    m.Seconds,
		EndSeconds: m.EndSeconds,
		Personal:   m.Personal,
		CreatedAt:  m.CreatedAt.UTC().Format(time.RFC3339),
	}
	if m.Tag != nil {
		tags := r.tagsByIDs(ctx, []db.RecordID{*m.Tag})
		if len(tags) > 0 {
			marker.Tag = tags[0]
		}
	}
	return marker
}

// SceneMarkers returns a scene's global and personal markers, earliest
// timestamp first.
func (r *Repository) SceneMarkers(ctx context.Context, mediaID string) ([]*model.SceneMarker, error) {
	media, err := parseMediaID(mediaID)
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryAs[sceneMarkerRecord](ctx, r.database,
		`SELECT * FROM scene_marker WHERE media = $media ORDER BY seconds ASC`,
		db.Vars{"media": *media})
	if err != nil {
		return nil, fmt.Errorf("scene markers: %w", err)
	}
	out := make([]*model.SceneMarker, 0, len(rows))
	for _, row := range rows {
		out = append(out, r.markerToModel(ctx, row))
	}
	return out, nil
}

// CreateSceneMarker adds a personal (user-created) marker on a scene. At least
// one of tagName/label must be set; tagName resolves alias-aware, creating the
// tag when it doesn't exist yet.
func (r *Repository) CreateSceneMarker(ctx context.Context, mediaID string, seconds float64, endSeconds *float64, tagName, label *string) (*model.SceneMarker, error) {
	media, err := parseMediaID(mediaID)
	if err != nil {
		return nil, err
	}

	name := ""
	if tagName != nil {
		name = strings.TrimSpace(*tagName)
	}
	labelValue := ""
	if label != nil {
		labelValue = strings.TrimSpace(*label)
	}
	if name == "" && labelValue == "" {
		return nil, fmt.Errorf("marker needs a tag or a label")
	}

	fields := map[string]any{
		"media":    *media,
		"personal": true,
		"seconds":  seconds,
	}
	if endSeconds != nil {
		if *endSeconds <= seconds {
			return nil, fmt.Errorf("marker end must be after its start")
		}
		fields["end_seconds"] = *endSeconds
	}
	if labelValue != "" {
		fields["label"] = labelValue
	}
	if name != "" {
		tagRID, tagErr := r.findOrCreateTag(ctx, name)
		if tagErr != nil {
			return nil, tagErr
		}
		fields["tag"] = tagRID
	}

	id, err := r.database.Insert(ctx, "scene_marker", fields)
	if err != nil {
		return nil, fmt.Errorf("create scene marker: %w", err)
	}
	row, err := db.QueryOneAs[sceneMarkerRecord](ctx, r.database,
		`SELECT * FROM scene_marker WHERE id = $id`, db.Vars{"id": id})
	if err != nil {
		return nil, fmt.Errorf("create scene marker: %w", err)
	}
	if row == nil {
		return nil, fmt.Errorf("create scene marker: no record returned")
	}
	return r.markerToModel(ctx, *row), nil
}

// DeleteSceneMarker removes a marker by id.
func (r *Repository) DeleteSceneMarker(ctx context.Context, markerID string) (bool, error) {
	id, err := db.ParseRecordID(markerID)
	if err != nil {
		return false, fmt.Errorf("invalid marker id: %w", err)
	}
	if err := r.database.Delete(ctx, *id); err != nil {
		return false, fmt.Errorf("delete scene marker: %w", err)
	}
	return true, nil
}

// findOrCreateTag resolves a tag by name or alias, creating it when absent.
func (r *Repository) findOrCreateTag(ctx context.Context, name string) (db.RecordID, error) {
	existing, err := r.database.Strings(ctx,
		`SELECT id FROM tag
			WHERE name = $name OR EXISTS (SELECT 1 FROM json_each(tag.aliases) WHERE value = $name)
			LIMIT 1`,
		db.Vars{"name": name})
	if err != nil {
		return db.RecordID{}, fmt.Errorf("find tag: %w", err)
	}
	if len(existing) > 0 {
		parsed, err := db.ParseRecordID(existing[0])
		if err != nil {
			return db.RecordID{}, fmt.Errorf("find tag: %w", err)
		}
		return *parsed, nil
	}
	created, err := r.database.Insert(ctx, "tag", map[string]any{"name": name})
	if err != nil {
		return db.RecordID{}, fmt.Errorf("create tag: %w", err)
	}
	return created, nil
}
