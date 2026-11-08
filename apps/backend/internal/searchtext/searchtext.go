// Package searchtext reads a search text the way the user means it: bare
// words, "quoted phrases" that must appear as written, and -excluded words
// or -"phrases". Words are compared as names.TextWords spells them, so case,
// accents, punctuation and a possessive 's don't matter.
package searchtext

import (
	"strings"

	"github.com/playingwithclouds/veil/internal/names"
)

// Query is a parsed search text; Words, Phrases and Excluded are folded
// (words joined by single spaces).
type Query struct {
	Words    []string
	Phrases  []string
	Excluded []string
	// siteTerms are the words and phrases as typed, in order, for asking sites.
	siteTerms []string
}

// Parse splits text into bare words, quoted phrases and excluded words or
// phrases. A quote left open runs to the end of the text.
func Parse(text string) Query {
	var query Query
	rest := text
	for {
		rest = strings.TrimSpace(rest)
		if rest == "" {
			return query
		}
		excluded := false
		if strings.HasPrefix(rest, "-") && len(rest) > 1 {
			excluded = true
			rest = rest[1:]
		}
		var term string
		quoted := strings.HasPrefix(rest, `"`)
		if quoted {
			term, rest = cutQuoted(rest[1:])
		} else {
			term, rest = cutWord(rest)
		}
		query.add(term, quoted, excluded)
	}
}

// cutQuoted returns the phrase up to the closing quote (or the end) and what follows it.
func cutQuoted(text string) (string, string) {
	phrase, rest, _ := strings.Cut(text, `"`)
	return phrase, rest
}

// cutWord returns the text up to the next space and what follows it.
func cutWord(text string) (string, string) {
	end := strings.IndexAny(text, " \t\n")
	if end == -1 {
		return text, ""
	}
	return text[:end], text[end:]
}

// add files one term under what it is; a term without letters or digits is dropped.
func (q *Query) add(term string, quoted, excluded bool) {
	folded := fold(term)
	if folded == "" {
		return
	}
	switch {
	case excluded:
		q.Excluded = append(q.Excluded, folded)
	case quoted:
		q.Phrases = append(q.Phrases, folded)
		q.siteTerms = append(q.siteTerms, strings.TrimSpace(term))
	default:
		// "emma_smile" folds to two words; each is required on its own.
		q.Words = append(q.Words, strings.Fields(folded)...)
		q.siteTerms = append(q.siteTerms, term)
	}
}

// fold spells a term as its words, joined by single spaces.
func fold(term string) string {
	return strings.Join(names.TextWords(term), " ")
}

// SiteText is what the sites are asked: the words and phrases without
// quotes, exclusions left out (sites don't agree on either syntax).
func (q Query) SiteText() string {
	return strings.Join(q.siteTerms, " ")
}

// Strict reports whether the query has phrases or exclusions, which the
// sites' own results have to be checked against.
func (q Query) Strict() bool {
	return len(q.Phrases) > 0 || len(q.Excluded) > 0
}

// Matches reports whether text has every phrase and no excluded word or
// phrase, and, with words, every bare word as the start of one of its words
// ("mal" finds "Malkova").
func (q Query) Matches(text string, words bool) bool {
	padded := " " + fold(text) + " "
	for _, phrase := range q.Phrases {
		if !strings.Contains(padded, " "+phrase+" ") {
			return false
		}
	}
	for _, excluded := range q.Excluded {
		if strings.Contains(padded, " "+excluded+" ") {
			return false
		}
	}
	if !words {
		return true
	}
	for _, word := range q.Words {
		if !strings.Contains(padded, " "+word) {
			return false
		}
	}
	return true
}
