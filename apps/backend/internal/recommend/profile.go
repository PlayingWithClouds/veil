package recommend

import (
	"math"
	"sort"
	"strings"
	"time"
)

// Profile is the user's time-decayed taste: an affinity in [-1, 1] per tag,
// performer, studio and site, normalized per dimension so the strongest
// entity of each is ±1.
type Profile struct {
	Tags       map[string]float64
	Performers map[string]float64
	Studios    map[string]float64
	Sites      map[string]float64
	// Seeds are the scenes with the strongest positive signal, strongest first.
	Seeds []Seed
	// Disliked holds the scenes rated thumbs-down.
	Disliked map[string]bool
	// AdjacentTags co-occur with the top tags on liked scenes without being
	// top tags themselves: the exploration frontier.
	AdjacentTags []string
	// SignalCount is how many signals the profile was built from.
	SignalCount int
}

// Seed is a scene whose related scenes are worth recommending.
type Seed struct {
	SceneID string
	Title   string
	Weight  float64
}

// Empty reports whether there is no signal yet (cold start).
func (profile *Profile) Empty() bool {
	return profile.SignalCount == 0
}

// newProfile returns a profile with empty affinity maps.
func newProfile(signalCount int) *Profile {
	return &Profile{
		Tags:        map[string]float64{},
		Performers:  map[string]float64{},
		Studios:     map[string]float64{},
		Sites:       map[string]float64{},
		Disliked:    map[string]bool{},
		SignalCount: signalCount,
	}
}

// decayFactor is the share of a signal's weight left after age.
func decayFactor(age, halfLife time.Duration) float64 {
	if age <= 0 {
		return 1
	}
	return math.Pow(0.5, float64(age)/float64(halfLife))
}

// buildProfile turns signals into per-entity affinities. features describes
// the signalled scenes; blocked holds the blocked entity ids ("table:id").
func buildProfile(signals []signal, features map[string]*sceneFeatures, blocked map[string]bool, now time.Time) *Profile {
	profile := newProfile(len(signals))
	sceneWeights := decayedSceneWeights(signals, now)
	for sceneID, weight := range sceneWeights {
		if scene, found := features[sceneID]; found {
			profile.addScene(scene, weight)
		}
	}
	profile.normalize()
	profile.applyBlocked(blocked)
	profile.Disliked = dislikedScenes(signals)
	profile.Seeds = seedsFrom(sceneWeights, features)
	profile.AdjacentTags = adjacentTags(profile, sceneWeights, features)
	return profile
}

// decayedSceneWeights sums each scene's signals after time decay.
func decayedSceneWeights(signals []signal, now time.Time) map[string]float64 {
	out := map[string]float64{}
	for _, entry := range signals {
		out[entry.sceneID] += entry.weight * decayFactor(now.Sub(entry.at), AffinityHalfLife)
	}
	return out
}

// dislikedScenes collects the scenes with a thumbs-down.
func dislikedScenes(signals []signal) map[string]bool {
	out := map[string]bool{}
	for _, entry := range signals {
		if entry.disliked {
			out[entry.sceneID] = true
		}
	}
	return out
}

// addScene spreads a scene's weight onto its entities. Tags share the weight
// (damped by the square root of their count) so a heavily tagged scene
// doesn't swamp the tag dimension; performers, studio and site take it whole.
func (profile *Profile) addScene(scene *sceneFeatures, weight float64) {
	for _, performer := range scene.performers {
		profile.Performers[performer] += weight
	}
	if len(scene.tags) > 0 {
		tagShare := weight / math.Sqrt(float64(len(scene.tags)))
		for _, tag := range scene.tags {
			profile.Tags[tag] += tagShare
		}
	}
	if scene.studio != "" {
		profile.Studios[scene.studio] += weight
	}
	if scene.site != "" {
		profile.Sites[scene.site] += weight
	}
}

// normalize scales each dimension into [-1, 1] by its largest magnitude.
func (profile *Profile) normalize() {
	for _, affinities := range profile.dimensions() {
		normalizeAffinities(affinities)
	}
}

// dimensions lists the affinity maps.
func (profile *Profile) dimensions() []map[string]float64 {
	return []map[string]float64{profile.Tags, profile.Performers, profile.Studios, profile.Sites}
}

// normalizeAffinities divides every value by the largest magnitude.
func normalizeAffinities(affinities map[string]float64) {
	largest := 0.0
	for _, value := range affinities {
		largest = math.Max(largest, math.Abs(value))
	}
	if largest == 0 {
		return
	}
	for key, value := range affinities {
		affinities[key] = value / largest
	}
}

// applyBlocked pins blocked tags, performers and studios to affinityBlocked.
func (profile *Profile) applyBlocked(blocked map[string]bool) {
	for id := range blocked {
		table, _, _ := strings.Cut(id, ":")
		switch table {
		case "tag":
			profile.Tags[id] = affinityBlocked
		case "performer":
			profile.Performers[id] = affinityBlocked
		case "studio":
			profile.Studios[id] = affinityBlocked
		}
	}
}

// seedsFrom picks the scenes strong enough to seed related candidates,
// strongest first.
func seedsFrom(sceneWeights map[string]float64, features map[string]*sceneFeatures) []Seed {
	seeds := []Seed{}
	for sceneID, weight := range sceneWeights {
		scene, found := features[sceneID]
		if !found || weight < seedMinimumWeight {
			continue
		}
		seeds = append(seeds, Seed{SceneID: sceneID, Title: scene.title, Weight: weight})
	}
	sort.Slice(seeds, func(left, right int) bool {
		if seeds[left].Weight != seeds[right].Weight {
			return seeds[left].Weight > seeds[right].Weight
		}
		return seeds[left].SceneID < seeds[right].SceneID
	})
	if len(seeds) > seedSceneLimit {
		seeds = seeds[:seedSceneLimit]
	}
	return seeds
}

// adjacentTags weighs the tags that appear next to a top tag on positively
// signalled scenes, excluding the top tags and disliked tags.
func adjacentTags(profile *Profile, sceneWeights map[string]float64, features map[string]*sceneFeatures) []string {
	topTags := toSet(TopEntities(profile.Tags, topTagLimit))
	coOccurrence := map[string]float64{}
	for sceneID, weight := range sceneWeights {
		scene, found := features[sceneID]
		if !found || weight <= 0 || !containsAny(scene.tags, topTags) {
			continue
		}
		addCoOccurrence(coOccurrence, scene.tags, topTags, weight)
	}
	for tag := range coOccurrence {
		if profile.Tags[tag] < 0 {
			delete(coOccurrence, tag)
		}
	}
	return TopEntities(coOccurrence, adjacentTagLimit)
}

// addCoOccurrence credits weight to every tag not in exclude.
func addCoOccurrence(coOccurrence map[string]float64, tags []string, exclude map[string]bool, weight float64) {
	for _, tag := range tags {
		if !exclude[tag] {
			coOccurrence[tag] += weight
		}
	}
}

// TopEntities returns up to limit ids with a positive value, highest first
// (ties by id, so the order is stable).
func TopEntities(values map[string]float64, limit int) []string {
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

// containsAny reports whether any value is in set.
func containsAny(values []string, set map[string]bool) bool {
	for _, value := range values {
		if set[value] {
			return true
		}
	}
	return false
}

// toSet builds a lookup set from values.
func toSet(values []string) map[string]bool {
	out := make(map[string]bool, len(values))
	for _, value := range values {
		out[value] = true
	}
	return out
}
