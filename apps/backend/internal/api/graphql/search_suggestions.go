package graphql

import (
	"strings"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/sitesuggest"
	"github.com/playingwithclouds/veil/internal/suggest"
)

// defaultSourceSuggestionLimit is how many site completions a query returns by default.
const defaultSourceSuggestionLimit = 8

// searchSuggestionModels converts engine suggestions; empty optional fields
// become null.
func searchSuggestionModels(suggestions []suggest.Suggestion) []*model.SearchSuggestion {
	out := make([]*model.SearchSuggestion, 0, len(suggestions))
	for _, suggestion := range suggestions {
		out = append(out, &model.SearchSuggestion{
			Kind:     model.SearchSuggestionKind(suggestion.Kind),
			Text:     suggestion.Text,
			EntityID: strPtrIfSet(suggestion.EntityID),
			ImageURL: strPtrIfSet(suggestion.ImageURL),
			Detail:   strPtrIfSet(suggestion.Detail),
		})
	}
	return out
}

// sourceSuggestionModels converts the sites' completions; their lower-case
// kinds become the shared enum.
func sourceSuggestionModels(suggestions []sitesuggest.Suggestion) []*model.SourceSuggestion {
	out := make([]*model.SourceSuggestion, 0, len(suggestions))
	for _, suggestion := range suggestions {
		out = append(out, &model.SourceSuggestion{
			Kind:    model.SearchSuggestionKind(strings.ToUpper(suggestion.Kind)),
			Text:    suggestion.Text,
			Sources: suggestion.Sources,
		})
	}
	return out
}
