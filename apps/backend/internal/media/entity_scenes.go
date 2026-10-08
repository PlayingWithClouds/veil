package media

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
)

// SceneTitle is a scene's id and title.
type SceneTitle struct {
	ID    string
	Title string
}

// UncreditedScenes returns the scenes among mediaIDs that have no credit of
// the entity's kind yet (no studio for a studio, no performers for a
// performer, no tags for a tag), with their titles. Non-scene ids are dropped.
func (r *Repository) UncreditedScenes(ctx context.Context, entityID string, mediaIDs []string) ([]SceneTitle, error) {
	target, err := db.ParseRecordID(entityID)
	if err != nil {
		return nil, fmt.Errorf("uncredited scenes: %w", err)
	}
	_, uncredited := sceneLinkConditions(target.Table)
	if uncredited == "" {
		return nil, fmt.Errorf("uncredited scenes: %s cannot be credited on scenes", target.Table)
	}
	if mediaIDs == nil {
		mediaIDs = []string{}
	}
	rows, err := r.database.Query(ctx,
		`SELECT scene.id, scene.title FROM scene
		 WHERE scene.id IN (SELECT value FROM json_each($ids)) AND `+uncredited,
		db.Vars{"ids": mediaIDs})
	if err != nil {
		return nil, fmt.Errorf("uncredited scenes: %w", err)
	}
	scenes := make([]SceneTitle, 0, len(rows))
	for _, row := range rows {
		scenes = append(scenes, SceneTitle{ID: rowString(row, "id"), Title: rowString(row, "title")})
	}
	return scenes, nil
}

// KeepLinked filters mediaIDs to the scenes credited to a studio, performer
// or tag (directly, not inherited).
func (r *Repository) KeepLinked(ctx context.Context, entityID string, mediaIDs []string) ([]string, error) {
	target, err := db.ParseRecordID(entityID)
	if err != nil {
		return nil, fmt.Errorf("keep linked scenes: %w", err)
	}
	linked, _ := sceneLinkConditions(target.Table)
	if linked == "" {
		return nil, fmt.Errorf("keep linked scenes: %s cannot be credited on scenes", target.Table)
	}
	if mediaIDs == nil {
		mediaIDs = []string{}
	}
	return r.database.Strings(ctx,
		`SELECT scene.id FROM scene WHERE scene.id IN (SELECT value FROM json_each($ids)) AND `+linked,
		db.Vars{"entity": target.String(), "ids": mediaIDs})
}

// CreditScenes credits a studio, performer or tag on the given scenes that
// still have no credit of its kind, and returns how many it credited. A
// performer also gets a performs_in credit per scene. Visiting a scene later
// replaces these with the credits its page lists.
func (r *Repository) CreditScenes(ctx context.Context, entityID string, sceneIDs []string) (int, error) {
	target, err := db.ParseRecordID(entityID)
	if err != nil {
		return 0, fmt.Errorf("credit scenes: %w", err)
	}
	assignment, err := creditAssignment(target.Table)
	if err != nil {
		return 0, err
	}
	_, uncredited := sceneLinkConditions(target.Table)
	credited := 0
	for _, sceneID := range sceneIDs {
		affected, err := r.database.Exec(ctx,
			`UPDATE scene SET `+assignment+` WHERE scene.id = $scene AND `+uncredited,
			db.Vars{"scene": sceneID, "entity": target.String()})
		if err != nil {
			return credited, fmt.Errorf("credit %s on %s: %w", entityID, sceneID, err)
		}
		if affected == 0 {
			continue
		}
		credited++
		if target.Table != "performer" {
			continue
		}
		if err := r.insertPerformerCredit(ctx, target.String(), sceneID); err != nil {
			return credited, err
		}
	}
	return credited, nil
}

// creditAssignment is the SET clause crediting $entity on a scene, by the
// entity's table.
func creditAssignment(table string) (string, error) {
	switch table {
	case "studio":
		return "studio = $entity", nil
	case "performer":
		return "performers = json_array($entity)", nil
	case "tag":
		return "tags = json_array($entity)", nil
	}
	return "", fmt.Errorf("credit scenes: %s cannot be credited on scenes", table)
}

// insertPerformerCredit records the performer's appearance in a scene once.
func (r *Repository) insertPerformerCredit(ctx context.Context, performerID, sceneID string) error {
	_, err := r.database.Exec(ctx,
		`INSERT INTO performs_in (id, performer, media) VALUES ($id, $performer, $media)
		 ON CONFLICT (performer, media) DO NOTHING`,
		db.Vars{"id": db.NewRecordID("performs_in"), "performer": performerID, "media": sceneID})
	if err != nil {
		return fmt.Errorf("performs_in credit %s on %s: %w", performerID, sceneID, err)
	}
	return nil
}
