package library

import (
	"context"
	"log"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// tableSiteRelated holds the videos sites list as related on a video's page.
const tableSiteRelated = "SiteRelated"

// replaceSiteRelated swaps the related videos a source's page lists for the
// ones it lists now, each stored as a stub like a search result. An item
// that can't be stored is left out.
func (l *Library) replaceSiteRelated(ctx context.Context, plugin, sourceID string, items []plugins.DiscoveredItem) error {
	videoIDs := make([]string, 0, len(items))
	for _, item := range items {
		if item.MediaType != "" && item.MediaType != plugins.MediaTypeScene {
			continue
		}
		item.MediaType = plugins.MediaTypeScene
		ref, err := l.IngestListed(ctx, plugin, item)
		if err != nil {
			log.Printf("store %s related %s: %v", plugin, item.SourceURL, err)
			continue
		}
		videoIDs = append(videoIDs, ref.Id)
	}
	return l.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM "SiteRelated" WHERE "source" = $source`, db.Vars{"source": sourceID}); err != nil {
			return err
		}
		for position, videoID := range videoIDs {
			row := db.Vars{"source": sourceID, "video": videoID, "position": position}
			if _, err := tx.Insert(ctx, tableSiteRelated, row); err != nil {
				return err
			}
		}
		return nil
	})
}

// SiteRelated lists a page of the videos the sites list as related on the
// video's pages: by their place in a site's list, the sites taking turns. The
// video itself and trashed, blocked or hidden videos are left out.
func (l *Library) SiteRelated(ctx context.Context, videoID string, limit, offset int) ([]schema.Video, error) {
	conditions, vars := filterConditions(schema.RecordKindVideo, schema.Filter{})
	conditions = append(conditions, `"media"."id" != $videoId`)
	vars["videoId"] = videoID
	vars["limit"], vars["offset"] = pageBounds(limit, offset)
	return db.QueryAs[schema.Video](ctx, l.database, `SELECT "media".*`+mediaSelect(schema.RecordKindVideo)+`
		JOIN (SELECT "related"."video" AS "video", MIN("related"."position") AS "position", MIN("source"."plugin") AS "plugin"
			FROM "SiteRelated" AS "related" JOIN "Source" AS "source" ON "source"."id" = "related"."source"
			WHERE "source"."subjectKind" = 'video' AND "source"."subjectId" = $videoId
				AND "source"."deletedAt" IS NULL AND "related"."deletedAt" IS NULL
			GROUP BY "related"."video") AS "listed" ON "listed"."video" = "media"."id"
		WHERE `+strings.Join(conditions, " AND ")+`
		ORDER BY "listed"."position", "listed"."plugin" LIMIT $limit OFFSET $offset`, vars)
}
