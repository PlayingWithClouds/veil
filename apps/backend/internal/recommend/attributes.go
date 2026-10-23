package recommend

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
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
func (attributes *performerAttributes) attributeWeight(key string) float64 {
	if attributes == nil {
		return 1
	}
	return shareWeight(attributes.frequency[key], len(attributes.byPerformer), attributeFullShare, attributeFloor)
}

// performerAttributeQuery reads the looks columns of every performer that
// has at least one filled in.
const performerAttributeQuery = `SELECT id, ethnicity, eye_color, hair_color, height_cm,
	measurements, fake_tits, tattoos, piercings, country FROM performer
	WHERE ethnicity IS NOT NULL OR eye_color IS NOT NULL OR hair_color IS NOT NULL
	   OR height_cm IS NOT NULL OR measurements IS NOT NULL OR fake_tits IS NOT NULL
	   OR tattoos IS NOT NULL OR piercings IS NOT NULL`

// loadPerformerAttributes indexes the attributes of all profiled performers.
func (e *Engine) loadPerformerAttributes(ctx context.Context) (*performerAttributes, error) {
	rows, err := e.database.Query(ctx, performerAttributeQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("load performer attributes: %w", err)
	}
	attributes := &performerAttributes{byPerformer: map[string][]string{}, frequency: map[string]int{}}
	for _, row := range rows {
		keys := attributeKeys(row)
		if len(keys) == 0 {
			continue
		}
		attributes.byPerformer[rowString(row, "id")] = keys
		for _, key := range keys {
			attributes.frequency[key]++
		}
	}
	return attributes, nil
}

// Height bands in centimetres and the cup bands: coarse on purpose, so
// performers a few centimetres apart share an attribute.
const (
	heightShortBelow = 160
	heightTallFrom   = 172
)

// cupPattern pulls the cup letters out of a bra size like "86DD-63-86 cm".
var cupPattern = regexp.MustCompile(`^\s*\d{2,3}\s*([A-Ka-k]{1,3})(?:[^A-Za-z]|$)`)

// attributeKeys derives the attribute keys of one performer row; unknown and
// empty values yield no key.
func attributeKeys(row db.Row) []string {
	keys := []string{}
	keys = appendCategory(keys, "hair", rowString(row, "hair_color"))
	keys = appendCategory(keys, "ethnicity", rowString(row, "ethnicity"))
	keys = appendCategory(keys, "eyes", rowString(row, "eye_color"))
	keys = appendCategory(keys, "country", rowString(row, "country"))
	keys = appendCategory(keys, "boobs", rowString(row, "fake_tits"))
	keys = appendFlag(keys, "tattoos", rowString(row, "tattoos"))
	keys = appendFlag(keys, "piercings", rowString(row, "piercings"))
	keys = appendCategory(keys, "cup", cupBand(rowString(row, "measurements")))
	return appendCategory(keys, "height", heightBand(rowInt(row, "height_cm")))
}

// appendCategory adds "kind:value" (lowercased) unless the value is empty or "unknown".
func appendCategory(keys []string, kind, value string) []string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || value == "unknown" {
		return keys
	}
	return append(keys, kind+":"+value)
}

// appendFlag adds "kind:yes" when the profile says "Yes: ..."; an empty
// field may only mean nothing was recorded, so it adds nothing.
func appendFlag(keys []string, kind, value string) []string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(value)), "yes") {
		return append(keys, kind+":yes")
	}
	return keys
}

// cupBand classifies a measurements string by its cup letters: small (A, B),
// medium (C, D), large (DD and up); empty when none is recognisable.
func cupBand(measurements string) string {
	match := cupPattern.FindStringSubmatch(measurements)
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

// heightBand classifies a height in centimetres as short, average or tall;
// empty when unknown.
func heightBand(heightCentimetres int) string {
	switch {
	case heightCentimetres <= 0:
		return ""
	case heightCentimetres < heightShortBelow:
		return "short"
	case heightCentimetres >= heightTallFrom:
		return "tall"
	}
	return "average"
}

// attributeAffinities spreads the affinity of every known performer (blocked
// ones left out: blocking one is not a verdict on her looks) onto her
// attributes, weighted by how telling each is, and scales the result into
// [-1, 1] with at least attributeMinMass of evidence behind a full ±1.
func attributeAffinities(performers map[string]float64, blocked map[string]bool, attributes *performerAttributes) map[string]float64 {
	out := map[string]float64{}
	if attributes == nil {
		return out
	}
	for performerID, affinity := range performers {
		if blocked[performerID] {
			continue
		}
		for _, key := range attributes.byPerformer[performerID] {
			out[key] += affinity * attributes.attributeWeight(key)
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
func (profile *Profile) inheritedAffinity(performerID string) float64 {
	if profile.attributes == nil || len(profile.Attributes) == 0 {
		return 0
	}
	keys := profile.attributes.byPerformer[performerID]
	if len(keys) == 0 {
		return 0
	}
	sum := 0.0
	for _, key := range keys {
		sum += profile.Attributes[key]
	}
	return attributeInheritance * sum / float64(len(keys))
}

// castAffinity is the strongest affinity among a scene's performers, or the
// lowest when none is positive. A performer the profile has no entry for
// inherits the damped affinity of her attributes.
func castAffinity(performerIDs []string, profile *Profile) float64 {
	highest, lowest := 0.0, 0.0
	for _, performerID := range performerIDs {
		affinity, known := profile.Performers[performerID]
		if !known {
			affinity = profile.inheritedAffinity(performerID)
		}
		highest = math.Max(highest, affinity)
		lowest = math.Min(lowest, affinity)
	}
	if highest > 0 {
		return highest
	}
	return lowest
}

// attributePerformer is an unknown performer ranked by her inherited affinity.
type attributePerformer struct {
	performerID string
	affinity    float64
}

// attributePerformers lists the performers the profile has no entry for whose
// attributes match the user's taste, best first, up to limit.
func attributePerformers(profile *Profile, limit int) []attributePerformer {
	if profile.attributes == nil {
		return nil
	}
	found := []attributePerformer{}
	for performerID := range profile.attributes.byPerformer {
		if _, known := profile.Performers[performerID]; known {
			continue
		}
		affinity := profile.inheritedAffinity(performerID)
		if affinity >= attributeCandidateMinimum {
			found = append(found, attributePerformer{performerID: performerID, affinity: affinity})
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

// attributeCandidates proposes the newest scenes of unknown performers who
// look like the ones the user likes. This is exploration (explore:attribute,
// kindExplore): the performer has no signal of her own, only her looks vouch
// for her.
func (e *Engine) attributeCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	performers := attributePerformers(state.profile, attributeCandidatePerformers)
	if len(performers) == 0 {
		return nil, nil
	}
	performerIDs := make([]string, 0, len(performers))
	for _, performer := range performers {
		performerIDs = append(performerIDs, performer.performerID)
	}
	scenesByPerformer, err := e.newestScenesByEntity(ctx, "performers", performerIDs, attributeScenesPerPerformer)
	if err != nil {
		return nil, err
	}
	out := []candidate{}
	for _, performer := range performers {
		reason := Reason{Kind: ReasonPerformer, EntityID: performer.performerID}
		match := min(1, performer.affinity/attributeInheritance)
		prior := sourcePriors[SourceExploreAttribute] * (attributeAffinityFloor + (1-attributeAffinityFloor)*match)
		out = append(out, sceneCandidates(scenesByPerformer[performer.performerID], SourceExploreAttribute, prior, reason)...)
	}
	return out, nil
}

// newestScenesByEntity lists, per performer or tag id, its newest scenes (at
// most perEntity) in one query. column is the scene's JSON list column
// ("performers" or "tags").
func (e *Engine) newestScenesByEntity(ctx context.Context, column string, entityIDs []string, perEntity int) (map[string][]string, error) {
	rows, err := e.database.Query(ctx,
		`SELECT id, entity FROM (
			SELECT scene.id AS id, linked.value AS entity,
			       row_number() OVER (PARTITION BY linked.value ORDER BY scene.created_at DESC) AS position
			FROM scene, json_each(scene.`+column+`) AS linked
			WHERE linked.value IN (SELECT value FROM json_each($entities))
		) WHERE position <= $limit ORDER BY entity, position`,
		db.Vars{"entities": entityIDs, "limit": perEntity})
	if err != nil {
		return nil, fmt.Errorf("newest scenes by %s: %w", column, err)
	}
	out := map[string][]string{}
	for _, row := range rows {
		entityID := rowString(row, "entity")
		out[entityID] = append(out[entityID], rowString(row, "id"))
	}
	return out, nil
}
