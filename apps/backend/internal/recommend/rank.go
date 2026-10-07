package recommend

import (
	"math"
	"sort"
	"time"
)

// scoredItem is a ranked item plus the features re-ranking needs.
type scoredItem struct {
	Item
	features *sceneFeatures
}

// scoreCandidates scores every candidate that has features and returns them
// best first.
func scoreCandidates(candidates []*candidate, features map[string]*sceneFeatures, state *snapshot) []scoredItem {
	out := make([]scoredItem, 0, len(candidates))
	for _, proposed := range candidates {
		scene, found := features[proposed.sceneID]
		if !found {
			continue
		}
		item := Item{SceneID: proposed.sceneID, Source: proposed.source, Reason: proposed.reason, Score: score(proposed, scene, state)}
		out = append(out, scoredItem{Item: item, features: scene})
	}
	sortByScore(out)
	return out
}

// sortByScore orders items best first, ties by scene id so the order is stable.
func sortByScore(items []scoredItem) {
	sort.SliceStable(items, func(left, right int) bool {
		if items[left].Score != items[right].Score {
			return items[left].Score > items[right].Score
		}
		return items[left].SceneID < items[right].SceneID
	})
}

// score is the transparent ranking formula: affinity match, source prior and
// freshness, minus the penalties.
func score(proposed *candidate, scene *sceneFeatures, state *snapshot) float64 {
	prior := proposed.prior + multiSourceBonus*float64(proposed.sourceCount-1)
	total := rankWeightAffinity * affinityMatch(scene, state.profile)
	total += rankWeightSourcePrior * prior
	total += rankWeightFreshness * freshness(scene.publishedAt, state.now)
	return total - penalties(proposed.sceneID, state)
}

// affinityMatch mixes the scene's performer, tag, studio and site affinities
// into one value in [-1, 1].
func affinityMatch(scene *sceneFeatures, profile *Profile) float64 {
	match := matchWeightPerformer * strongestAffinity(scene.performers, profile.Performers)
	match += matchWeightTag * tagAffinity(scene.tags, profile.Tags)
	match += matchWeightStudio * profile.Studios[scene.studio]
	match += matchWeightSite * profile.Sites[scene.site]
	return match
}

// strongestAffinity is the highest affinity among ids, or the lowest when none
// is positive: one loved performer carries a scene, an all-disliked cast sinks it.
func strongestAffinity(ids []string, affinities map[string]float64) float64 {
	highest, lowest := 0.0, 0.0
	for _, id := range ids {
		highest = math.Max(highest, affinities[id])
		lowest = math.Min(lowest, affinities[id])
	}
	if highest > 0 {
		return highest
	}
	return lowest
}

// tagAffinity sums the tag affinities damped by the tag count (mirroring how
// the profile spreads weight over tags), clamped to [-1, 1].
func tagAffinity(tags []string, affinities map[string]float64) float64 {
	if len(tags) == 0 {
		return 0
	}
	sum := 0.0
	for _, tag := range tags {
		sum += affinities[tag]
	}
	return math.Max(-1, math.Min(1, sum/math.Sqrt(float64(len(tags)))))
}

// penalties adds up what counts against a scene: already watched, shown
// repeatedly without a click, or related to a disliked scene.
func penalties(sceneID string, state *snapshot) float64 {
	total := watchedPenalty(state.watched[sceneID])
	total += fatiguePenalty(state.ignored[sceneID].count)
	if state.dislikedNeighbors[sceneID] {
		total += penaltyDislikedNeighbor
	}
	return total
}

// watchedPenalty costs a completed watch more than a partial one.
func watchedPenalty(state watchState) float64 {
	switch state {
	case watchCompleted:
		return penaltyWatchedCompleted
	case watchPartial:
		return penaltyWatchedPartial
	}
	return 0
}

// fatiguePenalty grows with every ignored showing past the free ones.
func fatiguePenalty(ignoredShows int) float64 {
	return penaltyFatiguePerShow * float64(max(0, ignoredShows-fatigueFreeShows))
}

// freshness is 1 for a scene published now, halving every freshnessHalfLife.
func freshness(published, now time.Time) float64 {
	return decayFactor(now.Sub(published), freshnessHalfLife)
}
