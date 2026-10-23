package graphql

import (
	"context"
	"log"
	"strings"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/searchrank"
)

// recordSearchEvents files the impressions that carry a search query with the
// search history, grouped by query, so the search ranking learns from them.
// Best-effort: a failure is logged, the impressions themselves are stored.
func (r *Resolver) recordSearchEvents(ctx context.Context, inputs []*model.ImpressionInput) {
	if r.searchHistory == nil {
		return
	}
	eventsByQuery := map[string][]searchrank.Event{}
	for _, input := range inputs {
		if input.Query == nil || strings.TrimSpace(*input.Query) == "" {
			continue
		}
		event := searchrank.Event{Media: input.MediaID, Plugin: searchrank.LibrarySource, Position: derefInt(input.Position, 0)}
		if input.Source != nil && *input.Source != "" {
			event.Plugin = *input.Source
		}
		if input.Clicked != nil {
			event.Clicked = *input.Clicked
		}
		eventsByQuery[*input.Query] = append(eventsByQuery[*input.Query], event)
	}
	for query, events := range eventsByQuery {
		if err := r.searchHistory.Record(ctx, query, events); err != nil {
			log.Printf("record search events: %v", err)
		}
	}
}
