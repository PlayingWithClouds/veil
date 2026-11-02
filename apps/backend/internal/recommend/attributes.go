package recommend

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/playingwithclouds/veil/internal/schema"
)

// performerAttributes holds the looks of the performers whose profile is
// known, as attribute keys ("hair:brown", "boobs:fake", "height:tall", ...),
// and how common each key is among them.
type performerAttributes struct {
	byPerformer map[string][]string
	frequency   map[string]int
}

// attributeWeight is how much an attribute says about taste: a common one
// (brown hair) less than a rare one. A nil index weighs 1.
func (looks *performerAttributes) attributeWeight(key string) float64 {
	if looks == nil {
		return 1
	}
	return shareWeight(looks.frequency[key], len(looks.byPerformer), attributeFullShare, attributeFloor)
}

// loadPerformerAttributes indexes the attributes of all profiled performers.
func (e *Engine) loadPerformerAttributes(ctx context.Context) (*performerAttributes, error) {
	rows, err := e.database.Query(ctx,
		`SELECT "id", "attributes" FROM "Performer" WHERE "deletedAt" IS NULL AND "attributes" != '{}'`, nil)
	if err != nil {
		return nil, fmt.Errorf("load performer attributes: %w", err)
	}
	looks := &performerAttributes{byPerformer: map[string][]string{}, frequency: map[string]int{}}
	for _, row := range rows {
		var attributes schema.PerformerAttributes
		if json.Unmarshal([]byte(rowString(row, "attributes")), &attributes) != nil {
			continue
		}
		keys := attributeKeys(attributes)
		if len(keys) == 0 {
			continue
		}
		looks.byPerformer[rowString(row, "id")] = keys
		for _, key := range keys {
			looks.frequency[key]++
		}
	}
	return looks, nil
}

// Height bands in centimetres: coarse on purpose, so performers a few
// centimetres apart share an attribute.
const (
	heightShortBelow = 160
	heightTallFrom   = 172
)

// cupPattern pulls the cup letters out of a bra size like "86DD-63-86 cm".
var cupPattern = regexp.MustCompile(`^\s*\d{2,3}\s*([A-Ka-k]{1,3})(?:[^A-Za-z]|$)`)

// attributeKeys derives the attribute keys of one performer; unknown and
// empty values yield no key.
func attributeKeys(attributes schema.PerformerAttributes) []string {
	keys := []string{}
	keys = appendCategory(keys, "hair", attributes.HairColor)
	keys = appendCategory(keys, "ethnicity", attributes.Ethnicity)
	keys = appendCategory(keys, "eyes", attributes.EyeColor)
	keys = appendCategory(keys, "country", attributes.Country)
	keys = appendCategory(keys, "boobs", attributes.FakeTits)
	keys = appendFlag(keys, "tattoos", attributes.Tattoos)
	keys = appendFlag(keys, "piercings", attributes.Piercings)
	keys = appendCategory(keys, "cup", stringPointer(cupBand(attributes.Measurements)))
	return appendCategory(keys, "height", stringPointer(heightBand(attributes.HeightCm)))
}

// appendCategory adds "kind:value" (lowercased) unless the value is empty or
// "unknown".
func appendCategory(keys []string, kind string, value *string) []string {
	if value == nil {
		return keys
	}
	normalized := strings.ToLower(strings.TrimSpace(*value))
	if normalized == "" || normalized == "unknown" {
		return keys
	}
	return append(keys, kind+":"+normalized)
}

// appendFlag adds "kind:yes" when the profile says "Yes: ..."; an empty field
// may only mean nothing was recorded, so it adds nothing.
func appendFlag(keys []string, kind string, value *string) []string {
	if value == nil {
		return keys
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(*value)), "yes") {
		return append(keys, kind+":yes")
	}
	return keys
}

// cupBand classifies a measurements string by its cup letters: small (A, B),
// medium (C, D), large (DD and up); empty when none is recognisable.
func cupBand(measurements *string) string {
	if measurements == nil {
		return ""
	}
	match := cupPattern.FindStringSubmatch(*measurements)
	if match == nil {
		return ""
	}
	cup := strings.ToUpper(match[1])
	switch {
	case cup == "A" || cup == "B" || cup == "AA":
		return "small"
	case cup == "C" || cup == "D":
		return "medium"
	case cup[0] >= 'D':
		return "large"
	}
	return ""
}

// heightBand classifies a height as short, average or tall; empty when unknown.
func heightBand(heightCentimetres *int32) string {
	if heightCentimetres == nil || *heightCentimetres <= 0 {
		return ""
	}
	if *heightCentimetres < heightShortBelow {
		return "short"
	}
	if *heightCentimetres >= heightTallFrom {
		return "tall"
	}
	return "average"
}

// attributeAffinities spreads the affinity of every known performer (blocked
// ones left out: blocking one is not a verdict on her looks) onto her
// attributes, weighted by how telling each is, and scales the result into
// [-1, 1] with at least attributeMinMass of evidence behind a full ±1.
func attributeAffinities(performers map[string]float64, blocked map[string]bool, looks *performerAttributes) map[string]float64 {
	out := map[string]float64{}
	if looks == nil {
		return out
	}
	for performerID, affinity := range performers {
		if blocked[performerID] {
			continue
		}
		for _, key := range looks.byPerformer[performerID] {
			out[key] += affinity * looks.attributeWeight(key)
		}
	}
	largest := attributeMinMass
	for _, value := range out {
		largest = math.Max(largest, math.Abs(value))
	}
	for key, value := range out {
		out[key] = value / largest
	}
	return out
}

// inheritedAffinity is the damped mean attribute affinity of a performer: what
// her looks alone say about the user's taste. 0 when nothing is known.
func (taste *profile) inheritedAffinity(performerID string) float64 {
	if taste.performerLooks == nil || len(taste.attributes) == 0 {
		return 0
	}
	keys := taste.performerLooks.byPerformer[performerID]
	if len(keys) == 0 {
		return 0
	}
	sum := 0.0
	for _, key := range keys {
		sum += taste.attributes[key]
	}
	return attributeInheritance * sum / float64(len(keys))
}

// castAffinity is the strongest affinity among an item's performers, or the
// lowest when none is positive. A performer the profile has no entry for
// inherits the damped affinity of her looks.
func castAffinity(performerIDs []string, taste *profile) float64 {
	highest, lowest := 0.0, 0.0
	for _, performerID := range performerIDs {
		affinity, known := taste.performers[performerID]
		if !known {
			affinity = taste.inheritedAffinity(performerID)
		}
		highest = math.Max(highest, affinity)
		lowest = math.Min(lowest, affinity)
	}
	if highest > 0 {
		return highest
	}
	return lowest
}

// lookalike is an unknown performer ranked by her inherited affinity.
type lookalike struct {
	performerID string
	affinity    float64
}

// lookalikes lists the performers the profile has no entry for whose looks
// match the user's taste, best first, up to limit.
func lookalikes(taste *profile, limit int) []lookalike {
	if taste.performerLooks == nil {
		return nil
	}
	found := []lookalike{}
	for performerID := range taste.performerLooks.byPerformer {
		if _, known := taste.performers[performerID]; known {
			continue
		}
		affinity := taste.inheritedAffinity(performerID)
		if affinity >= attributeCandidateMinimum {
			found = append(found, lookalike{performerID: performerID, affinity: affinity})
		}
	}
	sort.Slice(found, func(left, right int) bool {
		if found[left].affinity != found[right].affinity {
			return found[left].affinity > found[right].affinity
		}
		return found[left].performerID < found[right].performerID
	})
	if len(found) > limit {
		found = found[:limit]
	}
	return found
}
