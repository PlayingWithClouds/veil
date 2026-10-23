package searchlab

import (
	"sort"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/searchrank"
)

// Seed confidence: how sure a keyword result is what the query names, without
// any labels. A title carrying the whole query is sure; a title carrying part
// of it is believable only when its uploader has several such results (a
// series or a performer's own channel), which a word that just happens to
// match rarely does.
const (
	phraseConfidence   = 1.0
	uploaderConfidence = 0.7
	// uploaderMinimum is how many part-matching results an uploader needs.
	uploaderMinimum = 2
	// nicheThreshold: a query with fewer whole-phrase hits than this counts as
	// niche, and only niche queries are expanded.
	nicheThreshold = 10
	// seedLimit is how many confident hits a niche query expands from.
	seedLimit = 3
)

// Seed is a keyword result believed to be what the query names.
type Seed struct {
	Plugin     string
	Item       plugins.DiscoveredItem
	Confidence float64
	// place orders seeds of equal confidence: earlier in its site list first.
	place int
}

// SeedConfidences returns each keyword result's confidence (by ItemKey), 0
// for results that are not believed to be it.
func SeedConfidences(query Query, lists []List) map[string]float64 {
	words := searchrank.QueryWords(query.Text)
	partialByUploader := map[string]int{}
	for _, list := range keywordLists(lists) {
		for _, item := range list.Items {
			if uploader := uploaderOf(item); uploader != "" && sharesWord(words, item.Title) {
				partialByUploader[uploader]++
			}
		}
	}
	confidences := map[string]float64{}
	for _, list := range keywordLists(lists) {
		for _, item := range list.Items {
			confidences[ItemKey(item)] = itemConfidence(words, item, partialByUploader)
		}
	}
	return confidences
}

// itemConfidence is one result's seed confidence.
func itemConfidence(words []string, item plugins.DiscoveredItem, partialByUploader map[string]int) float64 {
	if containsPhrase(searchrank.QueryWords(item.Title), words) {
		return phraseConfidence
	}
	uploader := uploaderOf(item)
	if uploader != "" && sharesWord(words, item.Title) && partialByUploader[uploader] >= uploaderMinimum {
		return uploaderConfidence
	}
	return 0
}

// IsNiche reports whether few results carry the whole query, so the query
// is worth expanding from its confident hits.
func IsNiche(query Query, lists []List) bool {
	words := searchrank.QueryWords(query.Text)
	phraseHits := 0
	for _, list := range keywordLists(lists) {
		for _, item := range list.Items {
			if containsPhrase(searchrank.QueryWords(item.Title), words) {
				phraseHits++
			}
		}
	}
	return phraseHits < nicheThreshold
}

// Seeds returns the query's most confident keyword results, best first.
func Seeds(query Query, lists []List) []Seed {
	confidences := SeedConfidences(query, lists)
	seeds := []Seed{}
	for _, list := range keywordLists(lists) {
		for place, item := range list.Items {
			if confidence := confidences[ItemKey(item)]; confidence > 0 {
				seeds = append(seeds, Seed{Plugin: list.Plugin, Item: item, Confidence: confidence, place: place})
			}
		}
	}
	sort.SliceStable(seeds, func(left, right int) bool {
		if seeds[left].Confidence != seeds[right].Confidence {
			return seeds[left].Confidence > seeds[right].Confidence
		}
		return seeds[left].place < seeds[right].place
	})
	if len(seeds) > seedLimit {
		seeds = seeds[:seedLimit]
	}
	return seeds
}

// uploaderOf is the channel or uploader the listing names, empty when none.
func uploaderOf(item plugins.DiscoveredItem) string {
	if item.Studio == nil {
		return ""
	}
	if item.Studio.SourceURL != "" {
		return item.Studio.SourceURL
	}
	return item.Studio.Name
}

// sharesWord reports whether the title carries any query word as a word.
func sharesWord(words []string, title string) bool {
	titleWords := map[string]bool{}
	for _, word := range searchrank.QueryWords(title) {
		titleWords[word] = true
	}
	for _, word := range words {
		if titleWords[word] {
			return true
		}
	}
	return false
}

// containsPhrase reports whether words appear in title as one run.
func containsPhrase(title, words []string) bool {
	if len(words) == 0 {
		return false
	}
	for start := 0; start+len(words) <= len(title); start++ {
		matched := true
		for offset, word := range words {
			if title[start+offset] != word {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}
