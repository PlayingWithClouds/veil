package recommend

import (
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// rowString reads a text column, "" when absent.
func rowString(row db.Row, key string) string {
	switch value := row[key].(type) {
	case string:
		return value
	case []byte:
		return string(value)
	}
	return ""
}

// rowStrings reads a JSON array column of strings.
func rowStrings(row db.Row, key string) []string {
	values, isList := row[key].([]any)
	if !isList {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if text, isText := value.(string); isText {
			out = append(out, text)
		}
	}
	return out
}

// rowFloat reads a numeric column, 0 when absent.
func rowFloat(row db.Row, key string) float64 {
	switch value := row[key].(type) {
	case float64:
		return value
	case int64:
		return float64(value)
	}
	return 0
}

// rowInt reads an integer column, 0 when absent.
func rowInt(row db.Row, key string) int {
	return db.AsInt(row[key])
}

// rowTime reads a stored timestamp column, the zero time when absent or
// malformed.
func rowTime(row db.Row, key string) time.Time {
	parsed, err := time.Parse(db.TimeFormat, rowString(row, key))
	if err != nil {
		return time.Time{}
	}
	return parsed
}
