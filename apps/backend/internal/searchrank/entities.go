package searchrank

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/playingwithclouds/veil/internal/names"
)

// EntityKind says what a recognised name in the query stands for.
type EntityKind string

const (
	KindPerformer EntityKind = "performer"
	KindStudio    EntityKind = "studio"
	KindTag       EntityKind = "tag"
)

// maximumEntityWords is the longest name looked for in a query.
const maximumEntityWords = 5

// maximumFuzzyQueryWords is the longest query whose words are tried as a
// near-miss spelling of a name: longer queries are titles, not names.
const maximumFuzzyQueryWords = 6

// possessive matches the 's ending of "Mia Khalifa's", which titles add to a name.
var possessive = regexp.MustCompile(`['’]s\b`)

// Entity is a library record the query may name.
type Entity struct {
	Kind EntityKind
	// ID is the record id ("performer:x").
	ID   string
	Name string
}

// EntityMatch is an entity recognised in the query.
type EntityMatch struct {
	Entity
	// Start and End delimit the query words (End exclusive) the name covers.
	Start, End int
	// ViaAlias is true when the query used one of the record's aliases rather
	// than its name; Fuzzy when it used neither but a near-miss spelling.
	ViaAlias bool
	Fuzzy    bool
	// Keys are the comparison forms (names.Key) of the record's name and
	// aliases, which a title naming the entity may use.
	Keys []string
}

// EntityIndex finds the performers, studios and tags a query names, by their
// spelling-insensitive names and aliases (names.Key). Build it with AddName
// for every record, then AddAlias; a name always wins over another record's
// alias.
type EntityIndex struct {
	byKey   map[string]indexedName
	records map[string]*indexedRecord
}

// indexedName is the record a name key resolves to.
type indexedName struct {
	record  *indexedRecord
	isAlias bool
}

// indexedRecord is one record and every key it answers to.
type indexedRecord struct {
	entity Entity
	keys   []string
}

// NewEntityIndex returns an index that knows no names.
func NewEntityIndex() *EntityIndex {
	return &EntityIndex{byKey: map[string]indexedName{}, records: map[string]*indexedRecord{}}
}

// AddName teaches the index that the entity's name stands for it. A name
// already taken by another record stays with that one.
func (index *EntityIndex) AddName(entity Entity) {
	key := names.Key(entity.Name)
	if key == "" {
		return
	}
	record := index.recordFor(entity)
	record.keys = append(record.keys, key)
	if _, taken := index.byKey[key]; !taken {
		index.byKey[key] = indexedName{record: record}
	}
}

// AddAlias teaches the index that an alias of an entity (added with AddName)
// stands for it, unless another record owns that spelling already.
func (index *EntityIndex) AddAlias(entityID, alias string) {
	key := names.Key(alias)
	record := index.records[entityID]
	if key == "" || record == nil {
		return
	}
	record.keys = append(record.keys, key)
	if _, taken := index.byKey[key]; !taken {
		index.byKey[key] = indexedName{record: record, isAlias: true}
	}
}

// recordFor returns the record of an entity, creating it on first sight.
func (index *EntityIndex) recordFor(entity Entity) *indexedRecord {
	if record, found := index.records[entity.ID]; found {
		return record
	}
	record := &indexedRecord{entity: entity}
	index.records[entity.ID] = record
	return record
}

// Find returns the entities the query names, left to right and never
// overlapping; where several names start at the same word the longest wins
// (a performer before a studio before a tag on a tie). A single-word name is
// recognised for studios and tags anywhere, for performers only when it is the
// whole query ("Anna" in a longer query is just a word). Words no exact name
// accounts for are then tried as a near-miss spelling of a longer name
// (FuzzyEntityMinimumLength): such a match is flagged Fuzzy and displaces the
// exact matches it overlaps.
func (index *EntityIndex) Find(query string) []EntityMatch {
	words := strings.Fields(names.Key(query))
	exact := index.findExact(words)
	if coveredWords(exact) == len(words) || len(words) > maximumFuzzyQueryWords {
		return exact
	}
	return mergeMatches(exact, index.findFuzzy(words))
}

// coveredWords counts the query words the matches cover.
func coveredWords(matches []EntityMatch) int {
	covered := 0
	for _, match := range matches {
		covered += match.End - match.Start
	}
	return covered
}

// findExact scans the words for names spelled exactly (modulo Key).
func (index *EntityIndex) findExact(words []string) []EntityMatch {
	var found []EntityMatch
	for start := 0; start < len(words); {
		match, matched := index.longestAt(words, start)
		if !matched {
			start++
			continue
		}
		found = append(found, match)
		start = match.End
	}
	return found
}

// longestAt returns the longest recognised name starting at words[start].
func (index *EntityIndex) longestAt(words []string, start int) (EntityMatch, bool) {
	for length := min(maximumEntityWords, len(words)-start); length >= 1; length-- {
		name, found := index.byKey[strings.Join(words[start:start+length], " ")]
		if !found || !index.recognisable(name.record.entity.Kind, length, len(words)) {
			continue
		}
		return EntityMatch{
			Entity:   name.record.entity,
			Start:    start,
			End:      start + length,
			ViaAlias: name.isAlias,
			Keys:     slices.Clone(name.record.keys),
		}, true
	}
	return EntityMatch{}, false
}

// recognisable applies the single-word rule: a performer's one-word name is
// too often just a word unless the query is nothing else.
func (index *EntityIndex) recognisable(kind EntityKind, nameWords, queryWords int) bool {
	if nameWords > 1 || kind != KindPerformer {
		return true
	}
	return queryWords == 1
}

// findFuzzy looks, longest span first, for runs of query words that are a
// typo or two away from a known name: two words or more, or the whole query
// when it is one word. Spans that are names themselves are left to the exact
// search.
func (index *EntityIndex) findFuzzy(words []string) []EntityMatch {
	var found []EntityMatch
	minimumSpan := 2
	if len(words) == 1 {
		minimumSpan = 1
	}
	for length := min(maximumEntityWords, len(words)); length >= minimumSpan; length-- {
		for start := 0; start+length <= len(words); start++ {
			if overlapsAny(found, start, start+length) {
				continue
			}
			span := strings.Join(words[start:start+length], " ")
			if _, exact := index.byKey[span]; exact {
				continue
			}
			if match, near := index.nearestName(span, len(words)); near {
				match.Start, match.End = start, start+length
				found = append(found, match)
			}
		}
	}
	return found
}

// overlapsAny reports whether the word range [start, end) overlaps a match.
func overlapsAny(matches []EntityMatch, start, end int) bool {
	for _, match := range matches {
		if start < match.End && match.Start < end {
			return true
		}
	}
	return false
}

// nearestName returns the known name closest to the span within the typo
// allowance of its length, the closest real name before an alias and the
// alphabetically first key on a tie, so the answer does not depend on map
// order.
func (index *EntityIndex) nearestName(span string, queryWords int) (EntityMatch, bool) {
	allowance := fuzzyEntityAllowance(span)
	if allowance == 0 {
		return EntityMatch{}, false
	}
	spanRunes := []rune(span)
	var best *indexedName
	bestDistance, bestKey := allowance+1, ""
	for key, name := range index.byKey {
		if !index.recognisable(name.record.entity.Kind, strings.Count(key, " ")+1, queryWords) {
			continue
		}
		distance, near := nearKey(spanRunes, key, allowance)
		if !near || distance > bestDistance || (distance == bestDistance && !preferName(name, key, best, bestKey)) {
			continue
		}
		candidate := name
		best, bestDistance, bestKey = &candidate, distance, key
	}
	if best == nil {
		return EntityMatch{}, false
	}
	return EntityMatch{
		Entity:   best.record.entity,
		ViaAlias: best.isAlias,
		Fuzzy:    true,
		Keys:     slices.Clone(best.record.keys),
	}, true
}

// mergeMatches combines exact and fuzzy matches in query order; an exact match
// overlapping a fuzzy one is dropped (the fuzzy span is the longer reading).
func mergeMatches(exact, fuzzy []EntityMatch) []EntityMatch {
	merged := slices.Clone(fuzzy)
	for _, match := range exact {
		if !overlapsAny(fuzzy, match.Start, match.End) {
			merged = append(merged, match)
		}
	}
	slices.SortFunc(merged, func(first, second EntityMatch) int { return first.Start - second.Start })
	return merged
}

// fuzzyEntityAllowance is the typos tolerated in a name of that many letters
// (spaces included): none below FuzzyEntityMinimumLength, one from there, two
// from twice that.
func fuzzyEntityAllowance(span string) int {
	length := utf8.RuneCountInString(span)
	switch {
	case length < FuzzyEntityMinimumLength:
		return 0
	case length < 2*FuzzyEntityMinimumLength:
		return 1
	}
	return 2
}

// nearKey reports the edit distance of a key to the span when it is within
// the allowance and starts with the same letter.
func nearKey(span []rune, key string, allowance int) (int, bool) {
	keyLength := utf8.RuneCountInString(key)
	if abs(len(span)-keyLength) > allowance {
		return 0, false
	}
	first, _ := utf8.DecodeRuneInString(key)
	if first != span[0] {
		return 0, false
	}
	distance := editDistance(span, []rune(key), allowance)
	return distance, distance > 0 && distance <= allowance
}

// preferName breaks a distance tie between two fuzzy candidates in a way that
// does not depend on map order: a real name over an alias, then the
// alphabetically first key.
func preferName(candidate indexedName, candidateKey string, current *indexedName, currentKey string) bool {
	if current == nil {
		return true
	}
	if candidate.isAlias != current.isAlias {
		return !candidate.isAlias
	}
	return candidateKey < currentKey
}

// Coverage is the share of the query's words that matches of the given kinds
// cover, in [0, 1].
func Coverage(matches []EntityMatch, wordCount int, kinds ...EntityKind) float64 {
	if wordCount == 0 {
		return 0
	}
	covered := 0
	for _, match := range matches {
		if slices.Contains(kinds, match.Kind) {
			covered += match.End - match.Start
		}
	}
	return min(1, float64(covered)/float64(wordCount))
}

// Rewrite spells out the entities that were found through an alias or a
// near-miss spelling by their real names, leaving the rest of the query as it
// was. Empty when no match needed it: the query already says what it means.
func Rewrite(query string, matches []EntityMatch) string {
	words := strings.Fields(names.Key(query))
	var rewritten []string
	changed := false
	position := 0
	for _, match := range matches {
		rewritten = append(rewritten, words[position:match.Start]...)
		if match.ViaAlias || match.Fuzzy {
			rewritten = append(rewritten, strings.Fields(names.Key(match.Name))...)
			changed = true
		} else {
			rewritten = append(rewritten, words[match.Start:match.End]...)
		}
		position = match.End
	}
	if !changed {
		return ""
	}
	rewritten = append(rewritten, words[position:]...)
	return strings.Join(rewritten, " ")
}

// MentionedIn reports whether the text names the entity (its name or any
// alias) as whole words, ignoring case, accents and punctuation.
func (match EntityMatch) MentionedIn(text string) bool {
	padded := " " + names.Key(possessive.ReplaceAllString(text, "")) + " "
	for _, key := range match.Keys {
		if strings.Contains(padded, " "+key+" ") {
			return true
		}
	}
	return false
}
