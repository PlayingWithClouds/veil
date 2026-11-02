package library

import (
	"context"
	"fmt"
	"slices"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// performerClearedValues are what a performer's fields become when the
// profile that filled them turns out to be someone else's.
var performerClearedValues = map[string]any{
	"details":    nil,
	"image":      nil,
	"aliases":    []string{},
	"links":      []string{},
	"attributes": map[string]any{},
}

// CreditSourceID returns the source that wrote a video's performers: a
// performer's profile when the credit came from listing their videos, the
// video's own source when its page credited them. "" when the user set the
// credits or nobody owns them.
func (l *Library) CreditSourceID(ctx context.Context, videoID string) (string, error) {
	owners, err := l.recorder.FieldSources(ctx, schema.RecordKindVideo, videoID)
	if err != nil {
		return "", err
	}
	ordinal, _ := provenance.Ordinal(schema.RecordKindVideo, "performers")
	for _, owner := range owners {
		if owner.FieldOrdinal != ordinal || owner.Locked || owner.Source == nil {
			continue
		}
		return *owner.Source, nil
	}
	return "", nil
}

// UncreditPerformer takes a performer off a video's credits as the user, so
// the sites can't put them back.
func (l *Library) UncreditPerformer(ctx context.Context, videoID, performerID string) error {
	video, err := l.Video(ctx, videoID)
	if err != nil || video == nil {
		return err
	}
	performers := slices.DeleteFunc(slices.Clone(video.Performers), func(id string) bool { return id == performerID })
	_, err = l.recorder.Update(ctx, schema.RecordKindVideo, videoID, db.Vars{"performers": performers}, provenance.User())
	return err
}

// SplitProfile moves a performer's profile on a site to a new performer of
// the same name: the site's page is about someone else. The videos only that
// profile credited them on are credited to the new performer instead, and
// the performer's fields that profile filled are cleared and released, so
// their other profiles fill them again. Returns the new performer's id.
func (l *Library) SplitProfile(ctx context.Context, performerID, sourceID string) (string, error) {
	source, err := db.Get[schema.Source](ctx, l.database, "Source", sourceID)
	if err != nil {
		return "", err
	}
	if source == nil || source.SubjectKind != schema.RecordKindPerformer || source.SubjectId != performerID {
		return "", fmt.Errorf("source %s isn't a profile of performer %s", sourceID, performerID)
	}
	otherID, err := l.newPerformerLike(ctx, performerID, source)
	if err != nil {
		return "", err
	}
	if _, err := l.database.Exec(ctx, `UPDATE "Source" SET "subjectId" = $other, "updatedAt" = $now WHERE "id" = $id`,
		db.Vars{"other": otherID, "id": sourceID, "now": db.Now()}); err != nil {
		return "", err
	}
	if err := l.moveProfileCredits(ctx, performerID, otherID, source); err != nil {
		return "", err
	}
	return otherID, l.releaseProfileFields(ctx, performerID, source)
}

// newPerformerLike creates the performer a split profile moves to, named as
// the profile names them, else as the performer.
func (l *Library) newPerformerLike(ctx context.Context, performerID string, source *schema.Source) (string, error) {
	performer, err := l.Performer(ctx, performerID)
	if err != nil || performer == nil {
		return "", fmt.Errorf("performer %s: %w", performerID, err)
	}
	name := performer.Name
	if source.Title != nil && *source.Title != "" {
		name = *source.Title
	}
	return l.database.Insert(ctx, tablePerformer, newNamedRecord(tablePerformer, name, names.Key(name)))
}

// moveProfileCredits credits the other performer, instead of this one, on
// the videos whose credits only the profile wrote (the user's and a video
// page's own credits stay).
func (l *Library) moveProfileCredits(ctx context.Context, performerID, otherID string, source *schema.Source) error {
	ordinal, _ := provenance.Ordinal(schema.RecordKindVideo, "performers")
	videoIDs, err := l.database.Strings(ctx,
		`SELECT "owner"."subjectId" FROM "FieldSource" AS "owner"
		JOIN "Video" AS "video" ON "video"."id" = "owner"."subjectId"
		WHERE "owner"."subjectKind" = 'video' AND "owner"."fieldOrdinal" = $ordinal
			AND "owner"."source" = $source AND "owner"."locked" = 0 AND "owner"."deletedAt" IS NULL
			AND $performer IN (SELECT value FROM json_each("video"."performers"))`,
		db.Vars{"ordinal": ordinal, "source": source.Id, "performer": performerID})
	if err != nil {
		return err
	}
	for _, videoID := range videoIDs {
		if err := l.replaceCredit(ctx, videoID, performerID, otherID, source); err != nil {
			return err
		}
	}
	return nil
}

// replaceCredit swaps one performer for another in a video's credits, as the profile's plugin.
func (l *Library) replaceCredit(ctx context.Context, videoID, performerID, otherID string, source *schema.Source) error {
	video, err := l.Video(ctx, videoID)
	if err != nil || video == nil {
		return err
	}
	performers := []string{}
	for _, id := range video.Performers {
		if id == performerID {
			id = otherID
		}
		if !slices.Contains(performers, id) {
			performers = append(performers, id)
		}
	}
	_, err = l.recorder.Update(ctx, schema.RecordKindVideo, videoID, db.Vars{"performers": performers}, provenance.Plugin(source.Plugin, source.Id))
	return err
}

// releaseProfileFields clears the performer's fields the profile owns and
// releases them, so the next profile that describes them takes them over.
func (l *Library) releaseProfileFields(ctx context.Context, performerID string, source *schema.Source) error {
	owners, err := l.recorder.FieldSources(ctx, schema.RecordKindPerformer, performerID)
	if err != nil {
		return err
	}
	cleared := db.Vars{}
	for _, owner := range owners {
		column := provenance.Column(schema.RecordKindPerformer, owner.FieldOrdinal)
		value, clearable := performerClearedValues[column]
		if owner.Locked || owner.Source == nil || *owner.Source != source.Id || !clearable {
			continue
		}
		cleared[column] = value
	}
	if len(cleared) == 0 {
		return nil
	}
	if _, err := l.recorder.Update(ctx, schema.RecordKindPerformer, performerID, cleared, provenance.Plugin(source.Plugin, source.Id)); err != nil {
		return err
	}
	for column := range cleared {
		if err := l.recorder.SetLocked(ctx, schema.RecordKindPerformer, performerID, column, false); err != nil {
			return err
		}
	}
	return nil
}
