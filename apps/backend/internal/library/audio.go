package library

import (
	"context"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// upsertAudio stores a plugin's audio like upsertVideo does a scene: its
// streams, segments, transcripts, scripts and comments come with it.
func (l *Library) upsertAudio(ctx context.Context, plugin string, audio *plugins.Audio, fetched bool) (string, error) {
	site, err := identityOf(schema.RecordKindAudio, plugin, audio.ExternalID, audio.SourceURL)
	if err != nil {
		return "", err
	}
	fields, err := l.audioFields(ctx, plugin, audio, fetched)
	if err != nil {
		return "", err
	}
	id, sourceID, err := l.upsertMedia(ctx, mediaWrite{
		kind: schema.RecordKindAudio, plugin: plugin, site: site, fields: fields,
		label: audio.Title, image: assetOf(audio.CoverPath),
	})
	if err != nil {
		return "", err
	}
	return id, l.storeExtras(ctx, pageExtras{
		kind: schema.RecordKindAudio, id: id, sourceID: sourceID, fetched: fetched,
		downloads: audio.Downloads, markers: audio.Markers, subtitles: audio.Subtitles,
		scripts: audio.Scripts, comments: audio.Comments,
	})
}

// audioFields are the columns an audio fills; a listing leaves credits alone.
func (l *Library) audioFields(ctx context.Context, plugin string, audio *plugins.Audio, fetched bool) (db.Vars, error) {
	fields := db.Vars{}
	setText(fields, "title", audio.Title)
	setText(fields, "details", audio.Details)
	setText(fields, "date", isoDate(audio.Date))
	setAsset(fields, "cover", audio.CoverPath)
	if audio.Duration > 0 {
		fields["durationSeconds"] = float64(audio.Duration)
	}
	if err := l.setStudio(ctx, fields, plugin, audio.Studio); err != nil {
		return nil, err
	}
	if !fetched {
		return fields, nil
	}
	return fields, l.setCredits(ctx, fields, plugin, audio.Performers, audio.Tags)
}

// Audios lists a page of audio matching the query's filter.
func (l *Library) Audios(ctx context.Context, query MediaQuery) ([]schema.Audio, error) {
	statement, vars := mediaQuery(schema.RecordKindAudio, query)
	return db.QueryAs[schema.Audio](ctx, l.database, statement, vars)
}

// Audio loads one audio; nil when it doesn't exist or is trashed.
func (l *Library) Audio(ctx context.Context, id string) (*schema.Audio, error) {
	return db.Get[schema.Audio](ctx, l.database, tableAudio, id)
}
