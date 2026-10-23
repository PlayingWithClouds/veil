package searchrank

import (
	"strings"

	"github.com/playingwithclouds/veil/internal/names"
)

// TextFields is the text a result offers a query to match against. Each field
// holds names or words separated by spaces; empty fields are skipped.
type TextFields struct {
	Title      string
	Performers string
	Studio     string
	Tags       string
}

// QueryWords splits a query into its comparable words (lower case, accents
// removed, punctuation dropped).
func QueryWords(query string) []string {
	return strings.Fields(names.Key(query))
}

// TextMatch rates in [0, 1] how well the fields answer the query words. Each
// word earns the best credit any field gives it (a title word beats a
// performer, studio or tag, a whole word beats a prefix); the mean over the
// words is the coverage, and the whole query appearing in the title as a
// phrase adds a bonus, more at its start. Zero when there are no words.
func TextMatch(queryWords []string, fields TextFields) float64 {
	if len(queryWords) == 0 {
		return 0
	}
	title := strings.Fields(names.Key(fields.Title))
	performers := strings.Fields(names.Key(fields.Performers))
	studio := strings.Fields(names.Key(fields.Studio))
	tags := strings.Fields(names.Key(fields.Tags))

	coverage := 0.0
	for _, word := range queryWords {
		coverage += wordCredit(word, title, performers, studio, tags)
	}
	coverage /= float64(len(queryWords))
	return min(1, coverageShare*coverage+phraseBonus(queryWords, title))
}

// wordCredit is the best credit a query word earns across the fields.
func wordCredit(word string, title, performers, studio, tags []string) float64 {
	best := 0.0
	best = max(best, fieldCredit(word, title, titleWordCredit, titlePrefixCredit, titleInsideCredit))
	best = max(best, fieldCredit(word, performers, performerWordCredit, performerPrefixCredit, 0))
	best = max(best, fieldCredit(word, studio, studioWordCredit, studioPrefixCredit, 0))
	best = max(best, fieldCredit(word, tags, tagWordCredit, tagPrefixCredit, 0))
	return best
}

// fieldCredit is the credit of a word against one field's words: wordCredit
// for an equal word, prefixCredit for a word starting with it, insideCredit for
// one containing it (zero disables that level). A word that matches in none of
// those ways but is a typo or two away from a field word earns the whole-word
// credit scaled by FuzzyCreditFactor, so an exact match always scores higher.
func fieldCredit(word string, fieldWords []string, wholeCredit, prefixCredit, insideCredit float64) float64 {
	best := 0.0
	for _, fieldWord := range fieldWords {
		switch {
		case fieldWord == word:
			return wholeCredit
		case strings.HasPrefix(fieldWord, word):
			best = max(best, prefixCredit)
		case insideCredit > 0 && strings.Contains(fieldWord, word):
			best = max(best, insideCredit)
		}
	}
	if best > 0 {
		return best
	}
	for _, fieldWord := range fieldWords {
		if WithinTypos(word, fieldWord) {
			return wholeCredit * FuzzyCreditFactor
		}
	}
	return 0
}

// phraseBonus rewards a title that holds the query words as consecutive words:
// the larger bonus when it starts with them. Single-word queries have no
// phrase, so they earn nothing here.
func phraseBonus(queryWords []string, title []string) float64 {
	if len(queryWords) < 2 {
		return singleWordBonus(queryWords, title)
	}
	for start := 0; start+len(queryWords) <= len(title); start++ {
		if !startsWith(title[start:], queryWords) {
			continue
		}
		if start == 0 {
			return phraseAtStartBonus
		}
		return phraseInsideBonus
	}
	return 0
}

// singleWordBonus gives a one-word query the phrase bonus for a title that
// starts with that word (as a whole or the start of the word), or has it as a
// whole word later.
func singleWordBonus(queryWords []string, title []string) float64 {
	if len(queryWords) == 0 || len(title) == 0 {
		return 0
	}
	word := queryWords[0]
	if strings.HasPrefix(title[0], word) {
		return phraseAtStartBonus
	}
	for _, titleWord := range title[1:] {
		if titleWord == word {
			return phraseInsideBonus
		}
	}
	return 0
}

// startsWith reports whether words begins with prefix, the last prefix word
// matching as a word start so a query still being typed counts.
func startsWith(words, prefix []string) bool {
	last := len(prefix) - 1
	for index, word := range prefix {
		if index == last {
			return strings.HasPrefix(words[index], word)
		}
		if words[index] != word {
			return false
		}
	}
	return true
}
