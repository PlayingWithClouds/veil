package recommend

import (
	"math"
	"sort"
	"time"

	"github.com/playingwithclouds/veil/internal/schema"
)

// profile is the user's time-decayed taste: an affinity in [-1, 1] per tag,
// performer, studio and site, normalized per dimension so the strongest
// entity of each is ±1.
type profile struct {
	tags       map[string]float64
	performers map[string]float64
	studios    map[string]float64
	sites      map[string]float64
	// seeds are the items with the strongest positive signal, strongest first.
	seeds []seed
	// disliked holds the items rated or liked down.
	disliked map[schema.RecordRef]bool
	// adjacentTags co-occur with the top tags on liked items without being
	// top tags themselves: the exploration frontier.
	adjacentTags []string
	// signalCount is how many signals the profile was built from.
	signalCount int
	// attributes are the looks of the performers the user likes, as
	// affinities in [-1, 1] per attribute key.
	attributes map[string]float64
	// sessionEnds is when each session-ending video was last watched.
	sessionEnds map[schema.RecordRef]time.Time
	// camRooms is the affinity per cam room id, camGenders per room gender.
	camRooms   map[string]float64
	camGenders map[string]float64
	// camTags holds the room tags no library tag is spelled like, by
	// names.CompactKey, on the scale of tags.
	camTags map[string]float64

	runtime          [runtimeBands]float64
	rarity           *tagRarity
	performerLooks   *performerAttributes
	blockedPerformer map[string]bool
}

// seed is an item whose related items are worth recommending.
type seed struct {
	item   schema.RecordRef
	title  string
	weight float64
}

// empty reports whether there is no signal yet (cold start).
func (taste *profile) empty() bool {
	return taste.signalCount == 0
}

// profileInputs is what the profile reads beyond the signals themselves.
type profileInputs struct {
	features map[schema.RecordRef]*mediaFeatures
	blocked  *blockedEntities
	stats    *libraryStats
	// moments holds, per session-ending video, the tags of the segments near
	// where playback stopped and near the replay peak.
	moments map[schema.RecordRef][]string
	// cams describes the signalled cam rooms.
	cams *camInputs
}

// newProfile returns a profile with empty affinity maps.
func newProfile(signalCount int) *profile {
	return &profile{
		tags:             map[string]float64{},
		performers:       map[string]float64{},
		studios:          map[string]float64{},
		sites:            map[string]float64{},
		disliked:         map[schema.RecordRef]bool{},
		attributes:       map[string]float64{},
		sessionEnds:      map[schema.RecordRef]time.Time{},
		camRooms:         map[string]float64{},
		camGenders:       map[string]float64{},
		camTags:          map[string]float64{},
		blockedPerformer: map[string]bool{},
		signalCount:      signalCount,
	}
}

// buildProfile turns signals into per-entity affinities.
func buildProfile(collected *signals, inputs profileInputs, now time.Time) *profile {
	taste := newProfile(len(collected.items) + len(collected.entities) + len(collected.cams))
	if inputs.stats != nil {
		taste.rarity = inputs.stats.rarity
		taste.performerLooks = inputs.stats.attributes
	}
	itemWeights := decayedItemWeights(collected.items, now, false)
	for ref, weight := range itemWeights {
		item, found := inputs.features[ref]
		if !found {
			continue
		}
		taste.addItem(item, weight)
		taste.addMomentTags(inputs.moments[ref], weight)
	}
	for ref, weight := range decayedItemWeights(collected.items, now, true) {
		if item, found := inputs.features[ref]; found {
			taste.addRuntime(item.durationSeconds, weight)
		}
	}
	taste.addEntitySignals(collected.entities, now)
	taste.addCamSignals(collected.cams, inputs.cams, now)
	taste.normalize()
	taste.normalizeRuntime()
	taste.applyBlocked(inputs.blocked)
	taste.attributes = attributeAffinities(taste.performers, taste.blockedPerformer, taste.performerLooks)
	taste.disliked = dislikedItems(collected.items)
	taste.seeds = seedsFrom(itemWeights, inputs.features, taste.disliked)
	taste.adjacentTags = adjacentTags(taste, itemWeights, inputs.features)
	taste.sessionEnds = lastSessionEnds(collected.items)
	return taste
}

// decayedItemWeights sums each item's signals after time decay; contentOnly
// leaves out the signals that say nothing about content (shown, not opened).
func decayedItemWeights(items []signal, now time.Time, contentOnly bool) map[schema.RecordRef]float64 {
	out := map[schema.RecordRef]float64{}
	for _, entry := range items {
		if contentOnly && entry.weak {
			continue
		}
		out[entry.item] += entry.weight * signalDecay(entry, now)
	}
	return out
}

// signalDecay is the share of its weight a signal keeps at now.
func signalDecay(entry signal, now time.Time) float64 {
	decay := decayFactor(now.Sub(entry.at), affinityHalfLife)
	if entry.explicit {
		return math.Max(decay, explicitDecayFloor)
	}
	return decay
}

// dislikedItems collects the items with a thumbs-down or low rating.
func dislikedItems(items []signal) map[schema.RecordRef]bool {
	out := map[schema.RecordRef]bool{}
	for _, entry := range items {
		if entry.disliked {
			out[entry.item] = true
		}
	}
	return out
}

// addItem spreads an item's weight onto its entities. Tags share the weight
// (damped by the square root of their count, scaled by how rare each tag is)
// so a heavily tagged item or a generic tag doesn't swamp the tag dimension;
// performers, studio and site take it whole.
func (taste *profile) addItem(item *mediaFeatures, weight float64) {
	for _, performer := range item.performers {
		taste.performers[performer] += weight
	}
	if len(item.tags) > 0 {
		tagShare := weight / math.Sqrt(float64(len(item.tags)))
		for _, tag := range item.tags {
			taste.tags[tag] += tagShare * taste.rarity.weight(tag)
		}
	}
	if item.studio != "" {
		taste.studios[item.studio] += weight
	}
	if item.site != "" {
		taste.sites[item.site] += weight
	}
}

// addMomentTags gives the tags of a positively signalled video's moments an
// extra share of its weight, on top of what every tag of it already got.
func (taste *profile) addMomentTags(tags []string, weight float64) {
	if len(tags) == 0 || weight <= 0 {
		return
	}
	tagShare := momentTagBoost * weight / math.Sqrt(float64(len(tags)))
	for _, tag := range tags {
		taste.tags[tag] += tagShare * taste.rarity.weight(tag)
	}
}

// addEntitySignals adds the verdicts on performers, studios and tags
// themselves, decayed like explicit verdicts.
func (taste *profile) addEntitySignals(entities []entitySignal, now time.Time) {
	for _, entry := range entities {
		decay := math.Max(decayFactor(now.Sub(entry.at), affinityHalfLife), explicitDecayFloor)
		taste.dimension(entry.kind)[entry.id] += entry.weight * decay
	}
}

// dimension is the affinity map of an entity kind.
func (taste *profile) dimension(kind schema.RecordKind) map[string]float64 {
	switch kind {
	case schema.RecordKindPerformer:
		return taste.performers
	case schema.RecordKindStudio:
		return taste.studios
	}
	return taste.tags
}

// normalize scales each dimension into [-1, 1] by its largest magnitude. The
// room tags without a library tag share the tags' scale.
func (taste *profile) normalize() {
	normalizeAffinities(taste.tags, taste.camTags)
	for _, affinities := range []map[string]float64{taste.performers, taste.studios, taste.sites, taste.camRooms, taste.camGenders} {
		normalizeAffinities(affinities)
	}
}

// normalizeAffinities divides every value of the maps by the largest
// magnitude among them.
func normalizeAffinities(maps ...map[string]float64) {
	largest := 0.0
	for _, affinities := range maps {
		for _, value := range affinities {
			largest = math.Max(largest, math.Abs(value))
		}
	}
	if largest == 0 {
		return
	}
	for _, affinities := range maps {
		for key, value := range affinities {
			affinities[key] = value / largest
		}
	}
}

// applyBlocked pins blocked tags, performers and studios to affinityBlocked.
func (taste *profile) applyBlocked(blocked *blockedEntities) {
	if blocked == nil {
		return
	}
	for _, kind := range []schema.RecordKind{schema.RecordKindPerformer, schema.RecordKindStudio, schema.RecordKindTag} {
		for id := range blocked.set(kind) {
			taste.dimension(kind)[id] = affinityBlocked
		}
	}
	taste.blockedPerformer = blocked.performers
}

// seedsFrom picks the items strong enough to seed related candidates,
// strongest first.
func seedsFrom(itemWeights map[schema.RecordRef]float64, features map[schema.RecordRef]*mediaFeatures, disliked map[schema.RecordRef]bool) []seed {
	seeds := []seed{}
	for ref, weight := range itemWeights {
		item, found := features[ref]
		if !found || weight < seedMinimumWeight || disliked[ref] {
			continue
		}
		seeds = append(seeds, seed{item: ref, title: item.title, weight: weight})
	}
	sort.Slice(seeds, func(left, right int) bool {
		if seeds[left].weight != seeds[right].weight {
			return seeds[left].weight > seeds[right].weight
		}
		return seeds[left].item.Id < seeds[right].item.Id
	})
	if len(seeds) > seedLimit {
		seeds = seeds[:seedLimit]
	}
	return seeds
}

// adjacentTags weighs the tags that appear next to a top tag on positively
// signalled items, excluding the top tags and disliked tags.
func adjacentTags(taste *profile, itemWeights map[schema.RecordRef]float64, features map[schema.RecordRef]*mediaFeatures) []string {
	topTags := toSet(topEntities(taste.tags, topTagLimit))
	coOccurrence := map[string]float64{}
	for ref, weight := range itemWeights {
		item, found := features[ref]
		if !found || weight <= 0 || !containsAny(item.tags, topTags) {
			continue
		}
		for _, tag := range item.tags {
			if !topTags[tag] && taste.tags[tag] >= 0 {
				coOccurrence[tag] += weight
			}
		}
	}
	return topEntities(coOccurrence, adjacentTagLimit)
}

// lastSessionEnds is, per item with a session-ending watch, when it was last
// watched.
func lastSessionEnds(items []signal) map[schema.RecordRef]time.Time {
	out := map[schema.RecordRef]time.Time{}
	for _, entry := range items {
		if entry.sessionEnd && entry.at.After(out[entry.item]) {
			out[entry.item] = entry.at
		}
	}
	return out
}
