package alike

import (
	"regexp"
	"strings"

	"github.com/playingwithclouds/veil/internal/names"
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

// bracketedGroup matches a parenthesised or square-bracketed part of a title.
var bracketedGroup = regexp.MustCompile(`\([^()]*\)|\[[^\[\]]*\]`)

// tokenize splits text into lower-case words as names.Key spells them: accents
// folded, apostrophes joined to their word ("Summer's" is "summers"), any other
// punctuation a separator.
func tokenize(text string) []string {
	return strings.Fields(names.Key(text))
}

// isNoiseToken reports whether a token says nothing about which video a title
// names: filler, a site name, or a bare number of at most four digits
// (resolution, age notice).
func isNoiseToken(token string) bool {
	if _, junk := junkTokens[token]; junk {
		return true
	}
	if _, site := siteNames[token]; site {
		return true
	}
	return len(token) <= 4 && isNumber(token)
}

// isNumber reports whether token is made of digits only.
func isNumber(token string) bool {
	for _, character := range token {
		if character < '0' || character > '9' {
			return false
		}
	}
	return token != ""
}

// stripNoiseGroups removes the bracketed parts of a title that hold only
// noise: "(18+)", "(1080)", "[Full HD]". Groups with real words, such as a
// cast list, stay.
func stripNoiseGroups(title string) string {
	return bracketedGroup.ReplaceAllStringFunc(title, func(group string) string {
		for _, token := range tokenize(group[1 : len(group)-1]) {
			if !isNoiseToken(token) {
				return group
			}
		}
		return " "
	})
}

// titleTokens returns the meaningful tokens of a title: noise groups, junk and
// site names removed, single letters dropped (single digits stay: they tell parts apart).
func titleTokens(title string) []string {
	tokens := make([]string, 0)
	for _, token := range tokenize(stripNoiseGroups(title)) {
		if len(token) < 2 && !isNumber(token) {
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
