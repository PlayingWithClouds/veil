package ingest

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// maxStoredComments caps the comments kept per scene and plugin.
const maxStoredComments = 200

// syncSceneCommentsLogged replaces the scene's comments from this plugin with
// the ones its scrape found. A scrape without comments leaves the stored ones,
// since a page that failed to render them says nothing. Best-effort: failures
// are logged, not propagated.
func (s *Service) syncSceneCommentsLogged(ctx context.Context, canonicalID, pluginName string, comments []plugins.SceneComment) {
	if len(comments) == 0 {
		return
	}
	sceneRID, err := db.ParseRecordID(canonicalID)
	if err != nil {
		log.Printf("ingest: invalid scene id %q: %v", canonicalID, err)
		return
	}
	if err := s.replaceSceneComments(ctx, *sceneRID, pluginName, comments); err != nil {
		log.Printf("ingest: comments on %s: %v", canonicalID, err)
	}
}

// replaceSceneComments swaps the stored comments of (scene, plugin) for these,
// in their order, dropping empty ones and replies to comments not kept.
func (s *Service) replaceSceneComments(ctx context.Context, scene db.RecordID, pluginName string, comments []plugins.SceneComment) error {
	return s.database.Tx(ctx, func(tx *db.DB) error {
		_, err := tx.Exec(ctx, `DELETE FROM scene_comment WHERE scene = $scene AND plugin = $plugin`,
			db.Vars{"scene": scene, "plugin": pluginName})
		if err != nil {
			return fmt.Errorf("clear comments: %w", err)
		}
		for position, comment := range comments {
			if position >= maxStoredComments {
				break
			}
			text := strings.TrimSpace(comment.Text)
			if text == "" {
				continue
			}
			_, err := tx.Exec(ctx,
				`INSERT INTO scene_comment (scene, plugin, position, author, text, posted_at, likes, reply_to)
				 VALUES ($scene, $plugin, $position, $author, $text, $posted, $likes, $replyTo)`,
				db.Vars{
					"scene": scene, "plugin": pluginName, "position": position,
					"author": nullIfEmpty(strings.TrimSpace(comment.Author)), "text": text,
					"posted": nullIfEmpty(comment.Date), "likes": comment.Likes,
					"replyTo": earlierPosition(comment.ReplyTo, position),
				})
			if err != nil {
				return fmt.Errorf("insert comment %d: %w", position, err)
			}
		}
		return nil
	})
}

// earlierPosition keeps a reply's target only when it is an earlier comment.
func earlierPosition(replyTo *int, position int) any {
	if replyTo == nil || *replyTo < 0 || *replyTo >= position {
		return nil
	}
	return *replyTo
}

// nullIfEmpty stores "" as NULL.
func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
