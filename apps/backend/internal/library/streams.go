package library

import (
	"context"
	"net/url"
	"path"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// replaceStreams swaps the streams a source offers for the ones it lists now.
// Streams are the site's, so the old rows go for good.
func (l *Library) replaceStreams(ctx context.Context, kind schema.RecordKind, id, sourceID string, downloads []plugins.Download) error {
	return l.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM "Stream" WHERE "source" = $source`, db.Vars{"source": sourceID}); err != nil {
			return err
		}
		for _, download := range downloads {
			if download.URL == "" {
				continue
			}
			if _, err := tx.Insert(ctx, tableStream, streamRow(kind, id, sourceID, download)); err != nil {
				return err
			}
		}
		return nil
	})
}

// streamRow maps a plugin download to a stream row.
func streamRow(kind schema.RecordKind, id, sourceID string, download plugins.Download) schema.Stream {
	row := schema.Stream{
		SubjectKind: kind,
		SubjectId:   id,
		Source:      sourceID,
		Url:         download.URL,
		MimeType:    mimeTypeOf(download),
		Quality:     optional(download.Quality),
		Headers:     map[string]string{},
	}
	if download.SpeedBps > 0 {
		kilobits := int32(download.SpeedBps * 8 / 1000)
		row.SpeedKbps = &kilobits
	}
	return row
}

// mimeTypeOf guesses a download's media type from its format or URL.
func mimeTypeOf(download plugins.Download) string {
	format := strings.ToLower(download.Format)
	if format == "" {
		format = urlExtension(download.URL)
	}
	switch format {
	case "hls", "m3u8":
		return "application/x-mpegURL"
	case "webm":
		return "video/webm"
	case "mp4", "":
		return "video/mp4"
	}
	return "video/" + format
}

// urlExtension is the lower-case file extension of a URL's path, without the dot.
func urlExtension(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(strings.ToLower(path.Ext(parsed.Path)), ".")
}

// Streams lists the streams of a record, fastest measured first.
func (l *Library) Streams(ctx context.Context, kind schema.RecordKind, id string) ([]schema.Stream, error) {
	return db.QueryAs[schema.Stream](ctx, l.database,
		`SELECT * FROM "Stream" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL
		 ORDER BY "speedKbps" IS NULL, "speedKbps" DESC, "createdAt"`,
		db.Vars{"kind": string(kind), "id": id})
}
