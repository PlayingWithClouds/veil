package library

import (
	"context"

	"github.com/playingwithclouds/veil/internal/alike"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// AlikeCandidate describes a stored video for the alike matchers (title,
// cast names, runtime, date).
func (l *Library) AlikeCandidate(ctx context.Context, video *schema.Video) (alike.Candidate, error) {
	return l.alikeCandidate(ctx, video)
}

// SimilarRuntime lists the live videos whose runtime is within a few percent
// of the video's, oldest first: the ones worth comparing it with.
func (l *Library) SimilarRuntime(ctx context.Context, video *schema.Video) ([]schema.Video, error) {
	if video.DurationSeconds == nil {
		return []schema.Video{}, nil
	}
	return l.similarRuntime(ctx, video)
}

// Copies lists the live videos on the other side of a video's undismissed
// alike pairs: probably the same video from other sites, hidden from lists.
func (l *Library) Copies(ctx context.Context, videoID string) ([]schema.Video, error) {
	return db.QueryAs[schema.Video](ctx, l.database,
		`SELECT * FROM "Video" WHERE "deletedAt" IS NULL AND "id" IN (
			SELECT CASE WHEN "firstId" = $id THEN "secondId" ELSE "firstId" END FROM "Alike"
			WHERE "subjectKind" = 'video' AND "dismissed" = 0 AND "deletedAt" IS NULL
			  AND ("firstId" = $id OR "secondId" = $id))
		 ORDER BY "createdAt"`,
		db.Vars{"id": videoID})
}

// VideosMissingPictures lists live videos with a cover or preview stills
// whose pictures aren't hashed yet: no cache entry under keyPrefix + id.
func (l *Library) VideosMissingPictures(ctx context.Context, keyPrefix string, limit int) ([]schema.Video, error) {
	return db.QueryAs[schema.Video](ctx, l.database,
		`SELECT * FROM "Video" AS "video" WHERE "video"."deletedAt" IS NULL
		 AND ("video"."cover" IS NOT NULL OR COALESCE("video"."previewFrames", '[]') != '[]')
		 AND NOT EXISTS (SELECT 1 FROM "CacheEntry" WHERE "key" = $prefix || "video"."id")
		 ORDER BY "video"."createdAt" DESC LIMIT $limit`,
		db.Vars{"prefix": keyPrefix, "limit": limit})
}
