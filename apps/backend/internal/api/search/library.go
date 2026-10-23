package search

import (
	"context"
	"log"
	"sort"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/searchrank"
)

// libraryEmitter hands library hits on: annotate adds the personal and entity
// terms, send scores and streams the hit.
type libraryEmitter struct {
	blockedScenes, blockedTargets map[string]bool
	annotate                      func(hit *searchrank.Hit)
	send                          func(source, id string, hit searchrank.Hit)
}

// emit sends a stored record as a hit unless it is blocked. Returns whether
// it was sent.
func (emitter libraryEmitter) emit(mediaType plugins.MediaType, record map[string]any) bool {
	id := recordID(record)
	if emitter.blockedScenes[id] || emitter.blockedTargets[id] {
		return false
	}
	hit := libraryHit(mediaType, id, record)
	emitter.annotate(&hit)
	emitter.send(searchrank.LibrarySource, id, hit)
	return true
}

// sendLibraryHits looks the query up in the library, one media type at a
// time, and hands each match that is not blocked to the emitter in relevance
// order (so a list's place is its relevance). Each of the queries is tried in
// turn (the query as typed, then its rewrite with entity names spelled out); a
// record already sent for an earlier one is not sent again. Returns the ids
// of the scenes sent.
func (h *Handler) sendLibraryHits(ctx context.Context, queries []string, limit int, emitter libraryEmitter) map[string]bool {
	sentScenes := map[string]bool{}
	for _, mediaType := range []plugins.MediaType{plugins.MediaTypeScene, plugins.MediaTypeGallery, plugins.MediaTypePerformer, plugins.MediaTypeStudio} {
		sent := map[string]bool{}
		for _, query := range queries {
			records, err := h.ingestSvc.SearchLibrary(ctx, string(mediaType), query, limit)
			if err != nil {
				log.Printf("search db(%s, %q): %v", mediaType, query, err)
				continue
			}
			for _, record := range records {
				id := recordID(record)
				if sent[id] {
					continue
				}
				sent[id] = true
				if emitter.emit(mediaType, record) && mediaType == plugins.MediaTypeScene {
					sentScenes[id] = true
				}
			}
		}
	}
	return sentScenes
}

// sendFuzzyScenes is the typo-tolerant fallback for scenes: candidates that
// share many three-letter pieces with the query are kept when their words are
// within a typo of the query's (text match of at least
// searchrank.FuzzyMinimumTextMatch) and sent best match first, up to limit,
// skipping scenes already sent. Their text match is lower than an exact match's
// by searchrank.FuzzyCreditFactor, so they rank below exact results.
func (h *Handler) sendFuzzyScenes(ctx context.Context, query string, limit int, alreadySent map[string]bool, emitter libraryEmitter) {
	candidates, err := h.ingestSvc.SearchSceneCandidates(ctx, query)
	if err != nil {
		log.Printf("search fuzzy %q: %v", query, err)
		return
	}
	type scored struct {
		record map[string]any
		match  float64
	}
	var kept []scored
	for _, record := range candidates {
		if alreadySent[recordID(record)] {
			continue
		}
		match := searchrank.MatchHit(query, libraryHit(plugins.MediaTypeScene, recordID(record), record))
		if match >= searchrank.FuzzyMinimumTextMatch {
			kept = append(kept, scored{record: record, match: match})
		}
	}
	sort.SliceStable(kept, func(first, second int) bool { return kept[first].match > kept[second].match })
	for index := 0; index < len(kept) && index < limit; index++ {
		emitter.emit(plugins.MediaTypeScene, kept[index].record)
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
