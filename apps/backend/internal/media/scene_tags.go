package media

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
)

// ScenesCarryingTags returns, among sceneIDs, the ones carrying any of tagIDs
// (own tags, the studio's or a credited performer's), for dropping them from
// ranked feeds.
func (r *Repository) ScenesCarryingTags(ctx context.Context, sceneIDs, tagIDs []string) (map[string]bool, error) {
	carrying := map[string]bool{}
	if len(sceneIDs) == 0 || len(tagIDs) == 0 {
		return carrying, nil
	}
	vars := db.Vars{"ids": sceneIDs}
	var matches []string
	for index, tagID := range tagIDs {
		rid := parseFilterID(&tagID)
		if rid == nil {
			continue
		}
		name := fmt.Sprintf("tag%d", index)
		vars[name] = *rid
		matches = append(matches, inheritedTagConditionFor("scene", "$"+name))
	}
	if len(matches) == 0 {
		return carrying, nil
	}
	found, err := r.database.Strings(ctx,
		"SELECT id FROM scene WHERE id IN (SELECT value FROM json_each($ids)) AND ("+strings.Join(matches, " OR ")+")", vars)
	if err != nil {
		return nil, fmt.Errorf("scenes carrying tags: %w", err)
	}
	for _, id := range found {
		carrying[id] = true
	}
	return carrying, nil
}
