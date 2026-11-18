package library

import (
	"context"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// storedUploader converts a plugin's uploader to the stored value; nil for a
// plugin that reported none or one without a name.
func storedUploader(uploader *plugins.Uploader) *schema.Uploader {
	if uploader == nil || strings.TrimSpace(uploader.Name) == "" {
		return nil
	}
	return &schema.Uploader{
		Name:       strings.TrimSpace(uploader.Name),
		Url:        optional(strings.TrimSpace(uploader.URL)),
		ExternalId: optional(strings.TrimSpace(uploader.ExternalID)),
		Image:      assetOf(uploader.ImagePath),
	}
}

// setUploader remembers who uploaded the video a source describes. A scene
// that names no uploader leaves what an earlier read stored.
func (l *Library) setUploader(ctx context.Context, sourceID string, uploader *plugins.Uploader) error {
	stored := storedUploader(uploader)
	if stored == nil {
		return nil
	}
	return l.database.Update(ctx, tableSource, sourceID, db.Vars{"uploader": stored})
}

// SetUploaderWhereMissing gives the plugin's sources of the listed videos
// this uploader unless they already name one: a video listed on an
// uploader's page is theirs even when its card doesn't say so.
func (l *Library) SetUploaderWhereMissing(ctx context.Context, plugin string, videoIDs []string, uploader plugins.Uploader) error {
	stored := storedUploader(&uploader)
	if stored == nil || len(videoIDs) == 0 {
		return nil
	}
	_, err := l.database.Exec(ctx, `UPDATE "Source" SET "uploader" = $uploader, "updatedAt" = $now
		WHERE "plugin" = $plugin AND "subjectKind" = 'video' AND "uploader" IS NULL AND "deletedAt" IS NULL
			AND "subjectId" IN (SELECT value FROM json_each($ids))`,
		db.Vars{"uploader": stored, "plugin": plugin, "ids": videoIDs, "now": db.Now()})
	return err
}
