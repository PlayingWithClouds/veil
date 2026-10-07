package suggest

import (
	"strings"
	"unicode"
)

// matchQuality ranks how a candidate matched the typed text; higher is better.
type matchQuality int

const (
	matchNone matchQuality = iota
	matchFuzzy
	matchWordPrefix
	matchPrefix
	matchExact
)

// stopWords never start or end a mined title phrase ("in the", "with").
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "of": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "with": true,
	"by": true, "from": true, "is": true, "it": true, "her": true, "his": true,
	"he": true, "she": true, "my": true, "me": true, "you": true, "your": true,
	"gets": true, "get": true, "into": true, "up": true, "while": true,
}

// NormalizeQuery is a search's identity in the history: trimmed, whitespace
// collapsed, case-folded.
func NormalizeQuery(query string) string {
	return strings.ToLower(strings.Join(strings.Fields(query), " "))
}

// matchText is the form names, aliases, titles and typed text are compared
// in: lower-case words without punctuation, joined by single spaces.
func matchText(text string) string {
	return strings.Join(words(text), " ")
}

// words splits text into lower-case words of letters, digits and inner
// apostrophes ("Mom's" stays one word, "Riley-Reid" becomes two).
func words(text string) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), isWordSeparator)
	out := fields[:0]
	for _, field := range fields {
		trimmed := strings.Trim(field, "'’")
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// isWordSeparator reports whether a rune splits words.
func isWordSeparator(character rune) bool {
	if character == '\'' || character == '’' {
		return false
	}
	return !unicode.IsLetter(character) && !unicode.IsDigit(character)
}

// wordSuffixes lists the text from each later word on ("riley ann reid" →
// "ann reid", "reid"), the keys a word-prefix match looks up.
func wordSuffixes(text string) []string {
	out := []string{}
	for index := 0; index < len(text); index++ {
		if text[index] == ' ' {
			out = append(out, text[index+1:])
		}
	}
	return out
}

// textMatch rates how typed matches candidate, both in matchText form, without
// typo tolerance.
func textMatch(typed, candidate string) matchQuality {
	if candidate == typed {
		return matchExact
	}
	if strings.HasPrefix(candidate, typed) {
		return matchPrefix
	}
	for _, suffix := range wordSuffixes(candidate) {
		if strings.HasPrefix(suffix, typed) {
			return matchWordPrefix
		}
	}
	return matchNone
}

// allowedEdits is the typo budget for typed text of the given rune length.
func allowedEdits(length int) int {
	if length < fuzzyMinimumLength {
		return 0
	}
	if length < fuzzyTwoEditLength {
		return 1
	}
	return 2
}

// fuzzyPrefixDistance returns the fewest edits turning typed into some prefix
// of candidate, or maxEdits+1 when that takes more than maxEdits. rows is a
// scratch buffer of at least 2×(len(candidate)+1) ints, reused across calls.
func fuzzyPrefixDistance(typed, candidate []rune, maxEdits int, rows []int) int {
	// Only candidate prefixes up to len(typed)+maxEdits can be within budget.
	width := min(len(candidate), len(typed)+maxEdits)
	previous := rows[:width+1]
	current := rows[width+1 : 2*(width+1)]
	for column := range previous {
		previous[column] = column
	}
	for row := 1; row <= len(typed); row++ {
		current[0] = row
		rowMinimum := row
		for column := 1; column <= width; column++ {
			current[column] = editCell(previous, current, column, typed[row-1] == candidate[column-1])
			rowMinimum = min(rowMinimum, current[column])
		}
		if rowMinimum > maxEdits {
			return maxEdits + 1
		}
		previous, current = current, previous
	}
	// previous now holds the last row: typed against every candidate prefix.
	best := maxEdits + 1
	for _, distance := range previous {
		best = min(best, distance)
	}
	return best
}

// editCell is one Levenshtein cell: the cheapest of substitution (free when
// the runes are equal), deletion and insertion.
func editCell(previous, current []int, column int, equal bool) int {
	substitution := previous[column-1]
	if !equal {
		substitution++
	}
	return min(substitution, previous[column]+1, current[column-1]+1)
}
