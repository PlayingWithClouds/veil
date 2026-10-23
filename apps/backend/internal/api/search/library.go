package search

import (
	"context"
	"log"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/recommend"
	"github.com/playingwithclouds/veil/internal/searchrank"
)

// sendLibraryHits looks the query up in the library, one media type at a
// time, and hands each match that is not blocked to send in relevance order
// (so a list's place is its relevance).
func (h *Handler) sendLibraryHits(
	ctx context.Context,
	query string,
	limit int,
	blockedScenes, blockedTargets map[string]bool,
	profile *recommend.Profile,
	send func(source, id string, hit searchrank.Hit),
) {
	for _, mediaType := range []plugins.MediaType{plugins.MediaTypeScene, plugins.MediaTypeGallery, plugins.MediaTypePerformer, plugins.MediaTypeStudio} {
		records, err := h.ingestSvc.SearchLibrary(ctx, string(mediaType), query, limit)
		if err != nil {
			log.Printf("search db(%s, %q): %v", mediaType, query, err)
			continue
		}
		for _, record := range records {
			id := recordID(record)
			if blockedScenes[id] || blockedTargets[id] {
				continue
			}
			hit := libraryHit(mediaType, id, record)
			hit.Personal = h.personalAffinity(ctx, profile, hit)
			send(searchrank.LibrarySource, id, hit)
		}
	}
}

// libraryHit describes a stored record as a search hit. Performers and
// studios label with `name`; scenes and galleries with `title`. Scenes also
// carry the runtime, date and site numbers that clients sort by.
func libraryHit(mediaType plugins.MediaType, id string, record map[string]any) searchrank.Hit {
	title, _ := record["title"].(string)
	if title == "" {
		title, _ = record["name"].(string)
	}
	item := plugins.DiscoveredItem{Title: title, MediaType: mediaType}
	item.SourceURL, _ = record["source_url"].(string)
	item.ExternalID, _ = record["external_id"].(string)
	item.PosterPath, _ = record["poster_path"].(string)
	item.Date, _ = record["date"].(string)
	item.DurationSeconds = db.AsInt(record["duration_seconds"])
	item.ViewCount = db.AsInt(record["view_count"])
	item.Rating, _ = record["rating"].(float64)

	hit := searchrank.Hit{Source: searchrank.LibrarySource, ID: id, Item: item}
	hit.Performers = splitLines(record["search_performers"])
	hit.Tags = splitLines(record["search_tags"])
	hit.Studio, _ = record["search_studio"].(string)
	return hit
}

// splitLines splits the index's one-name-per-line text into names.
func splitLines(value any) []string {
	text, _ := value.(string)
	if text == "" {
		return nil
	}
	return strings.Split(text, "\n")
}
