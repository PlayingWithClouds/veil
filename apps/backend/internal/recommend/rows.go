package recommend

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// rowString reads a text column; "" when NULL.
func rowString(row db.Row, column string) string {
	value, present := row[column]
	if !present {
		return ""
	}
	if text, isText := value.(string); isText {
		return text
	}
	return fmt.Sprint(value)
}

// rowStrings reads a JSON list column; nil when NULL or not a list.
func rowStrings(row db.Row, column string) []string {
	var values []string
	if err := json.Unmarshal([]byte(rowString(row, column)), &values); err != nil {
		return nil
	}
	return values
}

// rowFloat reads a numeric column; 0 when NULL.
func rowFloat(row db.Row, column string) float64 {
	switch typed := row[column].(type) {
	case float64:
		return typed
	case int64:
		return float64(typed)
	case string:
		var parsed float64
		fmt.Sscan(typed, &parsed)
		return parsed
	}
	return 0
}

// rowInt reads an integer column; 0 when NULL.
func rowInt(row db.Row, column string) int {
	return db.AsInt(row[column])
}

// rowTime reads a timestamp column; zero when NULL or malformed.
func rowTime(row db.Row, column string) time.Time {
	return parseTime(rowString(row, column))
}

// parseTime reads a stored timestamp (db.TimeFormat, or RFC 3339 from other
// writers); zero when malformed.
func parseTime(text string) time.Time {
	parsed, err := time.Parse(db.TimeFormat, text)
	if err == nil {
		return parsed
	}
	parsed, err = time.Parse(time.RFC3339Nano, text)
	if err == nil {
		return parsed
	}
	return time.Time{}
}

// rowRef reads a record reference from a kind and an id column.
func rowRef(row db.Row, kindColumn, idColumn string) schema.RecordRef {
	return schema.RecordRef{Kind: schema.RecordKind(rowString(row, kindColumn)), Id: rowString(row, idColumn)}
}

// refKey is a record reference as one string, for CacheEntry keys.
func refKey(ref schema.RecordRef) string {
	return string(ref.Kind) + ":" + ref.Id
}

// decayFactor is the share of a signal's weight left after age.
func decayFactor(age, halfLife time.Duration) float64 {
	if age <= 0 {
		return 1
	}
	return math.Pow(0.5, float64(age)/float64(halfLife))
}

// clampFloat limits value to [low, high].
func clampFloat(value, low, high float64) float64 {
	return math.Min(math.Max(value, low), high)
}

// toSet builds a lookup set from values.
func toSet(values []string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		out[value] = true
	}
	return out
}

// setKeys lists a set's members, sorted so queries and draws are reproducible.
func setKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

// sortedKeys lists a map's keys in order.
func sortedKeys[Value any](values map[string]Value) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// containsAny reports whether any value is in set.
func containsAny(values []string, set map[string]bool) bool {
	for _, value := range values {
		if set[value] {
			return true
		}
	}
	return false
}

// topEntities returns up to limit ids with a positive value, highest first
// (ties by id, so the order is stable).
func topEntities(values map[string]float64, limit int) []string {
	ids := make([]string, 0, len(values))
	for id, value := range values {
		if value > 0 {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(left, right int) bool {
		if values[ids[left]] != values[ids[right]] {
			return values[ids[left]] > values[ids[right]]
		}
		return ids[left] < ids[right]
	})
	if len(ids) > limit {
		ids = ids[:limit]
	}
	return ids
}

// formatClock renders seconds as m:ss, or h:mm:ss from an hour.
func formatClock(seconds float64) string {
	total := int(seconds)
	hours, minutes, rest := total/3600, (total/60)%60, total%60
	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, rest)
	}
	return fmt.Sprintf("%d:%02d", minutes, rest)
}

// float32Pointer returns a pointer to value as float32.
func float32Pointer(value float64) *float32 {
	converted := float32(value)
	return &converted
}

// stringPointer returns a pointer to value.
func stringPointer(value string) *string {
	return &value
}

// veilEvidence is one evidence entry of the backend's own.
func veilEvidence(detail string) schema.Evidence {
	return schema.Evidence{Plugin: "veil", Detail: stringPointer(detail)}
}
