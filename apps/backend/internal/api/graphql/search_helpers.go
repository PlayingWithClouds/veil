package graphql

import "strconv"

// yearFromDate extracts the leading year from a YYYY-MM-DD date, 0 if absent.
func yearFromDate(date string) int {
	if len(date) < 4 {
		return 0
	}
	year, err := strconv.Atoi(date[:4])
	if err != nil {
		return 0
	}
	return year
}
