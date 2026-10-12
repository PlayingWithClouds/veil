// Package names compares people and company names that sites spell, cased and
// punctuate differently.
package names

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Key reduces a name to the form used to decide two spellings are the same
// person: lower case, accents removed, apostrophes dropped and every other run
// of punctuation or whitespace collapsed to one space ("Mia-Malkova",
// "MIA  MALKOVA" and "Mía Malkova" all give "mia malkova"). Empty when the
// name has no letters or digits.
func Key(name string) string {
	var key strings.Builder
	separatorPending := false
	for _, character := range norm.NFD.String(name) {
		switch {
		case unicode.Is(unicode.Mn, character), isApostrophe(character):
			continue
		case unicode.IsLetter(character), unicode.IsDigit(character):
			if separatorPending && key.Len() > 0 {
				key.WriteByte(' ')
			}
			separatorPending = false
			key.WriteRune(unicode.ToLower(character))
		default:
			separatorPending = true
		}
	}
	return key.String()
}

// isApostrophe reports whether the character is one of the apostrophes sites
// use inside names ("O'Reilly", "Summer’s"), which join the letters around them.
func isApostrophe(character rune) bool {
	return character == '\'' || character == '’' || character == '‘' || character == '`' || character == '´'
}
