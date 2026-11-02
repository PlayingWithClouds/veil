package library

import (
	"context"
	"slices"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/schema"
)

// SuggestionNamedInTitle is the reason of a performer suggested because the
// video's title names them.
const SuggestionNamedInTitle = "Named in the title"

// PerformerSuggestion is a performer who may be in a video, and why.
type PerformerSuggestion struct {
	Performer schema.Performer
	Reason    string
}

// SuggestedPerformers lists the known performers a video's title names (by
// a name or alias of two words or more) that aren't credited on it. Only
// suggestions: a name in a title isn't proof, so nothing is credited.
// Blocked performers are left out.
func (l *Library) SuggestedPerformers(ctx context.Context, video *schema.Video) ([]PerformerSuggestion, error) {
	words := names.TextWords(video.Title)
	candidates, err := db.QueryAs[schema.Performer](ctx, l.database,
		`SELECT * FROM "Performer" AS "performer" WHERE "performer"."deletedAt" IS NULL
			AND "performer"."nameKey" IN (SELECT value FROM json_each($keys))
			AND NOT EXISTS (SELECT 1 FROM "UserState" AS "blocked"
				WHERE "blocked"."subjectKind" = 'performer' AND "blocked"."subjectId" = "performer"."id"
					AND "blocked"."blocked" = 1 AND "blocked"."deletedAt" IS NULL)`,
		db.Vars{"keys": wordRuns(words, 2, 4)})
	if err != nil {
		return nil, err
	}
	finder := names.NewFinder()
	byID := map[string]schema.Performer{}
	for _, performer := range candidates {
		finder.Add(performer.Id, performer.Name)
		byID[performer.Id] = performer
	}
	suggestions := []PerformerSuggestion{}
	for _, id := range finder.Find(video.Title) {
		if slices.Contains(video.Performers, id) {
			continue
		}
		suggestions = append(suggestions, PerformerSuggestion{Performer: byID[id], Reason: SuggestionNamedInTitle})
	}
	return suggestions, nil
}
