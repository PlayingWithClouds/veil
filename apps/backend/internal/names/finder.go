package names

import (
	"regexp"
	"strings"
)

// possessive matches the 's ending of "Isla Summer's", which titles add to a name.
var possessive = regexp.MustCompile(`['’]s\b`)

// minimumFinderWords is the fewest words a name needs to be looked for inside
// text: a single word ("Anna", "Mia") is too often just a word.
const minimumFinderWords = 2

// Finder finds known names inside free text such as a video title, matching
// word by word as Key spells them, so case, accents and punctuation don't matter.
type Finder struct {
	// byFirstWord indexes entries by the first word of their name.
	byFirstWord map[string][]finderEntry
}

// finderEntry is one name a Finder looks for.
type finderEntry struct {
	id    string
	words []string
}

// NewFinder returns a Finder that knows no names.
func NewFinder() *Finder {
	return &Finder{byFirstWord: map[string][]finderEntry{}}
}

// Add teaches the Finder that name stands for id. A record may be added under
// several names (its name and aliases). Names of a single word are ignored.
func (f *Finder) Add(id, name string) {
	words := strings.Fields(Key(name))
	if len(words) < minimumFinderWords {
		return
	}
	f.byFirstWord[words[0]] = append(f.byFirstWord[words[0]], finderEntry{id: id, words: words})
}

// Find returns the ids of the known names that text contains, each once, in
// order of appearance. Where several names start at the same word the longest
// wins, and matches never overlap ("Cassie Del Isla" is one performer, not
// also "Del Isla"). A possessive 's after a name does not hide it.
func (f *Finder) Find(text string) []string {
	words := strings.Fields(Key(possessive.ReplaceAllString(text, "")))
	found := []string{}
	seen := map[string]bool{}
	for position := 0; position < len(words); {
		match, matched := f.longestAt(words, position)
		if !matched {
			position++
			continue
		}
		if !seen[match.id] {
			seen[match.id] = true
			found = append(found, match.id)
		}
		position += len(match.words)
	}
	return found
}

// longestAt returns the longest known name that starts at words[position].
func (f *Finder) longestAt(words []string, position int) (finderEntry, bool) {
	var best finderEntry
	matched := false
	for _, entry := range f.byFirstWord[words[position]] {
		if len(entry.words) <= len(best.words) || position+len(entry.words) > len(words) {
			continue
		}
		if wordsEqual(words[position:position+len(entry.words)], entry.words) {
			best = entry
			matched = true
		}
	}
	return best, matched
}

// wordsEqual reports whether two word lists are the same.
func wordsEqual(first, second []string) bool {
	for index := range first {
		if first[index] != second[index] {
			return false
		}
	}
	return len(first) == len(second)
}
