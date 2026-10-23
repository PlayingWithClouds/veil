package searchrank

import "unicode/utf8"

// FuzzyAllowance is how many typos a query word of that length tolerates:
// none for short words, one for ordinary ones, two for long ones.
func FuzzyAllowance(word string) int {
	length := utf8.RuneCountInString(word)
	switch {
	case length < FuzzyMinimumWordLength:
		return 0
	case length < FuzzyLongWordLength:
		return 1
	}
	return 2
}

// WithinTypos reports whether fieldWord is the query word with a typo or two:
// a letter added, dropped, changed or swapped with its neighbour, up to
// FuzzyAllowance(queryWord) of them. Both start with the same letter, which
// is rarely the one mistyped and cuts the false matches ("date" for "mate").
// Identical words are not typos of each other.
func WithinTypos(queryWord, fieldWord string) bool {
	allowance := FuzzyAllowance(queryWord)
	if allowance == 0 || queryWord == fieldWord {
		return false
	}
	query, field := []rune(queryWord), []rune(fieldWord)
	if query[0] != field[0] || abs(len(query)-len(field)) > allowance {
		return false
	}
	return editDistance(query, field, allowance) <= allowance
}

// EditDistance is the optimal-string-alignment distance of two words: the
// fewest insertions, deletions, substitutions and adjacent swaps turning one
// into the other.
func EditDistance(first, second string) int {
	return editDistance([]rune(first), []rune(second), max(utf8.RuneCountInString(first), utf8.RuneCountInString(second)))
}

// editDistance computes the optimal-string-alignment distance, giving up
// (returning limit+1) once every cell of a row exceeds limit.
func editDistance(first, second []rune, limit int) int {
	width := len(second) + 1
	twoBack := make([]int, width)
	previous := make([]int, width)
	current := make([]int, width)
	for column := range previous {
		previous[column] = column
	}
	for row := 1; row <= len(first); row++ {
		current[0] = row
		rowBest := current[0]
		for column := 1; column <= len(second); column++ {
			cost := 1
			if first[row-1] == second[column-1] {
				cost = 0
			}
			value := min(previous[column]+1, current[column-1]+1, previous[column-1]+cost)
			swapped := row > 1 && column > 1 && first[row-1] == second[column-2] && first[row-2] == second[column-1]
			if swapped {
				value = min(value, twoBack[column-2]+1)
			}
			current[column] = value
			rowBest = min(rowBest, value)
		}
		if rowBest > limit {
			return limit + 1
		}
		twoBack, previous, current = previous, current, twoBack
	}
	return previous[len(second)]
}

// abs is the absolute value of an int.
func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
