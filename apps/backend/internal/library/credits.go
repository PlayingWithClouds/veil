package library

import (
	"context"
	"slices"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// CreditPerformer adds a performer to a video's credits as the plugin whose
// list of the performer's videos showed it, from that list's source. A
// credit already there stays; credits the user set or another site owns
// aren't touched (provenance).
func (l *Library) CreditPerformer(ctx context.Context, videoID, performerID, plugin, sourceID string) error {
	video, err := l.Video(ctx, videoID)
	if err != nil || video == nil {
		return err
	}
	if slices.Contains(video.Performers, performerID) {
		return nil
	}
	performers := append(slices.Clone(video.Performers), performerID)
	_, err = l.recorder.Update(ctx, schema.RecordKindVideo, videoID, db.Vars{"performers": performers}, provenance.Plugin(plugin, sourceID))
	return err
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
