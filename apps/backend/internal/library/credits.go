package library

import (
	"context"
	"fmt"
	"slices"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// CreditPerformer adds a performer to a video's credits as the plugin whose
// list of the performer's videos showed it, from that list's source, and
// reports whether it did. A credit already there stays; credits the user set
// or another site owns aren't touched (provenance).
func (l *Library) CreditPerformer(ctx context.Context, videoID, performerID, plugin, sourceID string) (bool, error) {
	video, err := l.Video(ctx, videoID)
	if err != nil || video == nil {
		return false, err
	}
	if slices.Contains(video.Performers, performerID) {
		return false, nil
	}
	performers := append(slices.Clone(video.Performers), performerID)
	changed, err := l.recorder.Update(ctx, schema.RecordKindVideo, videoID, db.Vars{"performers": performers}, provenance.Plugin(plugin, sourceID))
	return len(changed) > 0, err
}

// textAttributes are the performer attributes held as text, the ones
// PerformerAttributeValues lists.
var textAttributes = []string{
	"gender", "country", "ethnicity", "eyeColor", "hairColor", "fakeTits", "tattoos", "piercings",
}

// PerformerAttributeValues lists the values one text attribute ("hairColor")
// has across the library's performers, most common first.
func (l *Library) PerformerAttributeValues(ctx context.Context, attribute string, limit int) ([]string, error) {
	if !slices.Contains(textAttributes, attribute) {
		return nil, fmt.Errorf("%q isn't a text attribute", attribute)
	}
	return l.database.Strings(ctx,
		`SELECT json_extract("attributes", $path) AS "value" FROM "Performer"
		WHERE "deletedAt" IS NULL AND "value" IS NOT NULL AND trim("value") != ''
		GROUP BY "value" ORDER BY count(*) DESC, "value" LIMIT $limit`,
		db.Vars{"path": "$." + attribute, "limit": limit})
}

// IsFavorite reports whether the user marked a record a favourite.
func (l *Library) IsFavorite(ctx context.Context, kind schema.RecordKind, id string) (bool, error) {
	row, err := l.database.QueryRow(ctx,
		`SELECT "favorite" FROM "UserState" WHERE "subjectKind" = $kind AND "subjectId" = $id`,
		db.Vars{"kind": kind, "id": id})
	if err != nil || row == nil {
		return false, err
	}
	favorite, _ := row["favorite"].(int64)
	return favorite == 1, nil
}
