package cams

import (
	"context"
	"slices"
	"strings"
	"unicode"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// maxVideoRooms bounds the rooms a video names, each checked live on its site.
const maxVideoRooms = 3

// videoRoomCandidatesQuery selects the rooms that may be a video's: the one it
// was recorded from, those linked to its performers, and those whose username
// is somewhere in its title (narrowed by videoNamesRoom).
const videoRoomCandidatesQuery = `
	SELECT * FROM "CamModel" WHERE "deletedAt" IS NULL AND (
		"id" = $recordedFrom
		OR "performer" IN (SELECT value FROM json_each($performers))
		OR instr(lower($title), lower("username")) > 0)
	ORDER BY "lastSeenOnlineAt" DESC`

// VideoRooms returns the cam rooms of the person in a video, with their live
// state checked now: the room it was recorded from, rooms linked to its
// performers, and rooms its title names by username together with their
// site (a cam show archive's title, or its studio, names the site).
func (s *Service) VideoRooms(ctx context.Context, video schema.Video) ([]schema.CamModel, error) {
	recordedFrom := ""
	if video.RecordedFrom != nil {
		recordedFrom = *video.RecordedFrom
	}
	candidates, err := db.QueryAs[schema.CamModel](ctx, s.services.Database, videoRoomCandidatesQuery, db.Vars{
		"recordedFrom": recordedFrom, "performers": video.Performers, "title": video.Title,
	})
	if err != nil {
		return nil, err
	}
	credits := video.Title + " " + s.studioName(ctx, video.Studio)
	rooms := []schema.CamModel{}
	for _, candidate := range candidates {
		if len(rooms) >= maxVideoRooms {
			break
		}
		if candidate.Id != recordedFrom && !linkedToCredits(candidate, video.Performers) && !s.videoNamesRoom(credits, candidate) {
			continue
		}
		rooms = append(rooms, s.withLiveChecked(ctx, candidate))
	}
	return rooms, nil
}

// linkedToCredits reports whether the room is linked to one of the performers.
func linkedToCredits(camModel schema.CamModel, performers []string) bool {
	return camModel.Performer != nil && slices.Contains(performers, *camModel.Performer)
}

// videoNamesRoom reports whether credits (a title and studio) name the room:
// its username as a whole word, and its site by name, so a tube video titled
// "Sandy" isn't taken for the room "sandy".
func (s *Service) videoNamesRoom(credits string, camModel schema.CamModel) bool {
	if !containsWord(credits, camModel.Username) {
		return false
	}
	plugin, ok := s.services.Sites.Plugin(camModel.Plugin)
	if !ok {
		return false
	}
	return strings.Contains(strings.ToLower(credits), strings.ToLower(plugin.Meta.DisplayName))
}

// containsWord reports whether word is one of text's words, ignoring case.
// Words are runs of letters, digits and underscores, as usernames are.
func containsWord(text, word string) bool {
	words := strings.FieldsFunc(strings.ToLower(text), func(character rune) bool {
		return !unicode.IsLetter(character) && !unicode.IsDigit(character) && character != '_'
	})
	return slices.Contains(words, strings.ToLower(word))
}

// studioName is the name of the studio with id, "" without one.
func (s *Service) studioName(ctx context.Context, id *string) string {
	if id == nil {
		return ""
	}
	names, err := s.services.Database.Strings(ctx,
		`SELECT "name" FROM "Studio" WHERE "id" = $id`, db.Vars{"id": *id})
	if err != nil || len(names) == 0 {
		return ""
	}
	return names[0]
}

// withLiveChecked asks the room's site whether it's live now and returns the
// room as stored after the answer; the stored room when that fails.
func (s *Service) withLiveChecked(ctx context.Context, camModel schema.CamModel) schema.CamModel {
	s.checkOnline(ctx, &camModel)
	refreshed, err := s.load(ctx, camModel.Id)
	if err != nil {
		return camModel
	}
	return *refreshed
}
