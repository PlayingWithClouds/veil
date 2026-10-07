package alike

import (
	"regexp"
	"strings"
)

// Site names, quality markers and generic adult-content filler that carry no
// signal for identifying a specific scene. Dropped from title token sets so a
// title bonus rewards only meaningful words.
var junkTokens = map[string]struct{}{
	// Quality / format noise.
	"hd": {}, "sd": {}, "uhd": {}, "4k": {}, "1080p": {}, "720p": {}, "480p": {},
	"2160p": {}, "full": {}, "hq": {}, "mp4": {}, "video": {}, "vid": {}, "clip": {},
	"scene": {}, "movie": {}, "watch": {}, "online": {}, "stream": {}, "download": {},
	// Generic adult filler.
	"porn": {}, "porno": {}, "xxx": {}, "sex": {}, "free": {}, "hardcore": {},
	"adult": {}, "tube": {}, "com": {}, "www": {},
}

// Known source-site names, stripped from titles that append " - SpankBang",
// " | TXXX", etc.
var siteNames = map[string]struct{}{
	"spankbang": {}, "txxx": {}, "upornia": {}, "hclips": {}, "hdzog": {},
	"vjav": {}, "voyeurhit": {}, "thegay": {}, "shemalez": {}, "fetishshrine": {},
	"hotmovs": {}, "tnaflix": {}, "xhamster": {}, "eporner": {}, "hqporner": {},
	"missav": {}, "pornpics": {}, "pornhub": {}, "xvideos": {}, "youporn": {},
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// tokenize lowercases and splits on any run of non-alphanumeric characters.
func tokenize(text string) []string {
	lowered := strings.ToLower(text)
	parts := nonAlphanumeric.Split(lowered, -1)
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			tokens = append(tokens, part)
		}
	}
	return tokens
}

// titleTokens returns the meaningful tokens of a title: junk and site names
// removed, single-character tokens dropped.
func titleTokens(title string) []string {
	tokens := make([]string, 0)
	for _, token := range tokenize(title) {
		if len(token) < 2 {
			continue
		}
		if _, junk := junkTokens[token]; junk {
			continue
		}
		if _, site := siteNames[token]; site {
			continue
		}
		tokens = append(tokens, token)
	}
	return tokens
}

// normalizeName collapses a performer/studio name to a canonical form: lowercased
// tokens joined by single spaces, with junk removed. Used to compare name sets
// across sites that spell or punctuate names differently.
func normalizeName(name string) string {
	tokens := make([]string, 0)
	for _, token := range tokenize(name) {
		if _, junk := junkTokens[token]; junk {
			continue
		}
		tokens = append(tokens, token)
	}
	return strings.Join(tokens, " ")
}

// nameSet builds a set of normalized names, skipping any that normalize to empty.
func nameSet(names []string) map[string]struct{} {
	set := make(map[string]struct{}, len(names))
	for _, name := range names {
		normalized := normalizeName(name)
		if normalized == "" {
			continue
		}
		set[normalized] = struct{}{}
	}
	return set
}

// tokenSet builds a set from a token slice.
func tokenSet(tokens []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		set[token] = struct{}{}
	}
	return set
}

// jaccard is the intersection-over-union of two sets, in 0..1. Two empty sets
// have no similarity (0), not full similarity.
func jaccard(a, b map[string]struct{}) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	intersection := 0
	for key := range a {
		if _, ok := b[key]; ok {
			intersection++
		}
	}
	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}
