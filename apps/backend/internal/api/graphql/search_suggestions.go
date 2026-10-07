package graphql

import (
	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/suggest"
)

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
