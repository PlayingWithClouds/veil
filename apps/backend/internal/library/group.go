package library

import (
	"context"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// tableGroup is the table site groups live in.
const tableGroup = "Group"

// upsertGroup stores a site's own grouping of videos (a movie, a series, a
// compilation) and its members in order. Members are named by page URL: one
// the library knows becomes that video, an unknown one a stub video that's
// read when opened.
func (l *Library) upsertGroup(ctx context.Context, plugin string, collection *plugins.Collection) (string, error) {
	site, err := identityOf(schema.RecordKindGroup, plugin, collection.ExternalID, collection.SourceURL)
	if err != nil {
		return "", err
	}
	fields := db.Vars{}
	setText(fields, "title", collection.Name)
	setText(fields, "details", collection.Details)
	setAsset(fields, "cover", collection.CoverPath)
	id, err := l.storeGroup(ctx, plugin, site, fields, collection)
	if err != nil {
		return "", err
	}
	if len(collection.MemberURLs) == 0 {
		return id, nil
	}
	return id, l.replaceGroupItems(ctx, plugin, id, collection.MemberURLs)
}

// storeGroup creates the group on first sight, else updates the fields its
// source owns.
func (l *Library) storeGroup(ctx context.Context, plugin string, site siteIdentity, fields db.Vars, collection *plugins.Collection) (string, error) {
	source, err := l.findSource(ctx, site)
	if err != nil {
		return "", err
	}
	if source != nil {
		_, err = l.recorder.Update(ctx, schema.RecordKindGroup, source.SubjectId, fields, provenance.Plugin(plugin, source.Id))
		return source.SubjectId, err
	}
	filled := columnsOf(fields)
	if _, hasTitle := fields["title"]; !hasTitle {
		fields["title"] = site.URL
	}
	id, err := l.database.Insert(ctx, tableGroup, fields)
	if err != nil {
		return "", err
	}
	sourceID, err := l.addSource(ctx, schema.RecordKindGroup, id, site, collection.Name, assetOf(collection.CoverPath))
	if err != nil {
		return "", err
	}
	if err := l.markFetched(ctx, sourceID); err != nil {
		return "", err
	}
	return id, l.recorder.Claim(ctx, schema.RecordKindGroup, id, filled, provenance.Plugin(plugin, sourceID))
}

// replaceGroupItems swaps a group's members for the ones its page lists.
func (l *Library) replaceGroupItems(ctx context.Context, plugin, groupID string, memberURLs []string) error {
	videoIDs := []string{}
	for _, url := range memberURLs {
		videoID, err := l.videoForURL(ctx, plugin, url)
		if err != nil {
			return err
		}
		videoIDs = append(videoIDs, videoID)
	}
	return l.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM "GroupItem" WHERE "group" = $group`, db.Vars{"group": groupID}); err != nil {
			return err
		}
		seen := map[string]bool{}
		for position, videoID := range videoIDs {
			if seen[videoID] {
				continue
			}
			seen[videoID] = true
			if _, err := tx.Insert(ctx, "GroupItem", db.Vars{"group": groupID, "video": videoID, "position": position}); err != nil {
				return err
			}
		}
		return nil
	})
}

// videoForURL returns the video a page URL belongs to, storing a stub when
// the library doesn't know the page yet.
func (l *Library) videoForURL(ctx context.Context, plugin, url string) (string, error) {
	known, err := l.database.Strings(ctx,
		`SELECT "subjectId" FROM "Source" WHERE "url" = $url AND "subjectKind" = $kind AND "deletedAt" IS NULL LIMIT 1`,
		db.Vars{"url": url, "kind": schema.RecordKindVideo})
	if err != nil || len(known) > 0 {
		return firstString(known), err
	}
	return l.upsertVideo(ctx, plugin, &plugins.Scene{SourceURL: url}, false)
}

// firstString returns the first value, or "".
func firstString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// Groups lists site groups, newest first, optionally by title text or one
// of their videos.
func (l *Library) Groups(ctx context.Context, text, videoID string, limit, offset int) ([]schema.Group, error) {
	conditions := `"deletedAt" IS NULL`
	vars := db.Vars{}
	if text != "" {
		conditions += ` AND "title" LIKE $text`
		vars["text"] = "%" + text + "%"
	}
	if videoID != "" {
		conditions += ` AND "id" IN (SELECT "group" FROM "GroupItem" WHERE "video" = $video AND "deletedAt" IS NULL)`
		vars["video"] = videoID
	}
	vars["limit"], vars["offset"] = pageBounds(limit, offset)
	return db.QueryAs[schema.Group](ctx, l.database,
		`SELECT * FROM "Group" WHERE `+conditions+` ORDER BY "createdAt" DESC LIMIT $limit OFFSET $offset`, vars)
}

// Group loads one group; nil when it doesn't exist or is trashed.
func (l *Library) Group(ctx context.Context, id string) (*schema.Group, error) {
	return db.Get[schema.Group](ctx, l.database, tableGroup, id)
}

// Recordings lists the videos recorded from a cam room, newest first.
func (l *Library) Recordings(ctx context.Context, camModelID string) ([]schema.Video, error) {
	return db.QueryAs[schema.Video](ctx, l.database,
		`SELECT * FROM "Video" WHERE "recordedFrom" = $room AND "deletedAt" IS NULL ORDER BY "createdAt" DESC`,
		db.Vars{"room": camModelID})
}

// GroupVideos lists a group's videos in the group's order.
func (l *Library) GroupVideos(ctx context.Context, groupID string) ([]schema.Video, error) {
	return db.QueryAs[schema.Video](ctx, l.database,
		`SELECT "Video".* FROM "GroupItem" JOIN "Video" ON "Video"."id" = "GroupItem"."video"
		 WHERE "GroupItem"."group" = $group AND "GroupItem"."deletedAt" IS NULL AND "Video"."deletedAt" IS NULL
		 ORDER BY "GroupItem"."position"`,
		db.Vars{"group": groupID})
}
