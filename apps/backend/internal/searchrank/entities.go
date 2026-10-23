package searchrank

import (
	"slices"
	"strings"

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
// whole query ("Anna" in a longer query is just a word). When nothing matches
// exactly, a near-miss spelling of a name covering the whole query is returned
// as a Fuzzy match.
func (index *EntityIndex) Find(query string) []EntityMatch {
	words := strings.Fields(names.Key(query))
	found := index.findExact(words)
	if len(found) == 0 {
		return index.findFuzzy(words)
	}
	return found
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

// findFuzzy looks for one name within a typo of the whole query: same
// length class, same first letter, at most FuzzyEntityAllowance edits. Names
// of fewer than FuzzyEntityMinimumLength letters are never guessed.
func (index *EntityIndex) findFuzzy(words []string) []EntityMatch {
	query := strings.Join(words, " ")
	allowance := fuzzyEntityAllowance(query)
	if allowance == 0 || len(words) == 0 {
		return nil
	}
	var best *indexedName
	bestDistance := allowance + 1
	bestKey := ""
	for key, name := range index.byKey {
		if !index.recognisable(name.record.entity.Kind, len(strings.Fields(key)), len(words)) {
			continue
		}
		distance, near := nearKey(query, key, allowance)
		if !near || distance > bestDistance || (distance == bestDistance && !preferName(name, key, best, bestKey)) {
			continue
		}
		candidate := name
		best, bestDistance, bestKey = &candidate, distance, key
	}
	if best == nil {
		return nil
	}
	return []EntityMatch{{
		Entity:   best.record.entity,
		Start:    0,
		End:      len(words),
		ViaAlias: best.isAlias,
		Fuzzy:    true,
		Keys:     slices.Clone(best.record.keys),
	}}
}

// fuzzyEntityAllowance is the typos tolerated in a whole-query name: one from
// FuzzyEntityMinimumLength letters, two from twice that.
func fuzzyEntityAllowance(query string) int {
	length := len([]rune(query))
	switch {
	case length < FuzzyEntityMinimumLength:
		return 0
	case length < 2*FuzzyEntityMinimumLength:
		return 1
	}
	return 2
}

// nearKey reports the edit distance of a key to the query when it is within
// the allowance and starts with the same letter.
func nearKey(query, key string, allowance int) (int, bool) {
	queryRunes, keyRunes := []rune(query), []rune(key)
	if queryRunes[0] != keyRunes[0] || abs(len(queryRunes)-len(keyRunes)) > allowance || query == key {
		return 0, false
	}
	distance := editDistance(queryRunes, keyRunes, allowance)
	return distance, distance <= allowance
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
	padded := " " + names.Key(text) + " "
	for _, key := range match.Keys {
		if strings.Contains(padded, " "+key+" ") {
			return true
		}
	}
	return false
}
