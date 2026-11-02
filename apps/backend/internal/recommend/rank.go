package recommend

import (
	"math"
	"sort"
	"time"

	"github.com/playingwithclouds/veil/internal/schema"
)

// scoredItem is a ranked candidate plus the features re-ranking needs.
type scoredItem struct {
	*candidate
	features *mediaFeatures
	score    float64
	// rewatch marks a session ender whose watched penalty has given way: a
	// favourite worth seeing again.
	rewatch bool
}

// scoreCandidates scores every eligible candidate (one with features, so not
// trashed or blocked, and not disliked) and returns them best first.
func scoreCandidates(candidates []*candidate, features map[schema.RecordRef]*mediaFeatures, state *snapshot) []scoredItem {
	out := make([]scoredItem, 0, len(candidates))
	for _, proposed := range candidates {
		item, found := features[proposed.item]
		if !found || state.taste.disliked[proposed.item] {
			continue
		}
		total, rewatch := score(proposed, item, state)
		out = append(out, scoredItem{candidate: proposed, features: item, score: total, rewatch: rewatch})
	}
	sortByScore(out)
	return out
}

// sortByScore orders items best first, ties by id so the order is stable.
func sortByScore(items []scoredItem) {
	sort.SliceStable(items, func(left, right int) bool {
		if items[left].score != items[right].score {
			return items[left].score > items[right].score
		}
		return items[left].item.Id < items[right].item.Id
	})
}

// score is the transparent ranking formula: affinity match, source prior and
// freshness, minus the penalties. rewatch reports whether the watched penalty
// turned into a bonus.
func score(proposed *candidate, item *mediaFeatures, state *snapshot) (total float64, rewatch bool) {
	prior := proposed.prior*state.policy.multiplier(proposed.source) + multiSourceBonus*float64(proposed.sourceCount-1)
	total = rankWeightAffinity * affinityMatch(item, state.taste)
	total += rankWeightSourcePrior * prior
	total += rankWeightFreshness * freshness(item.publishedAt, state.now)
	watched := watchedPenalty(state.watched[proposed.item], state.taste.sessionEnds[proposed.item], state.now)
	total -= watched + fatiguePenalty(state.collected.ignored[proposed.item].count)
	return total, watched < 0
}

// affinityMatch mixes the item's performer, tag, studio, site and runtime
// affinities into one value in [-1, 1].
func affinityMatch(item *mediaFeatures, taste *profile) float64 {
	match := matchWeightPerformer * castAffinity(item.performers, taste)
	match += matchWeightTag * tagAffinity(item.tags, taste.tags, taste.rarity)
	match += matchWeightStudio * taste.studios[item.studio]
	match += matchWeightSite * taste.sites[item.site]
	match += matchWeightRuntime * taste.runtimeAffinity(item.durationSeconds)
	return match
}

// tagAffinity sums the tag affinities, each weighted by how rare the tag is
// and damped by the tag count (mirroring how the profile spreads weight over
// tags), clamped to [-1, 1]. A blocked tag keeps its full weight.
func tagAffinity(tags []string, affinities map[string]float64, rarity *tagRarity) float64 {
	if len(tags) == 0 {
		return 0
	}
	sum := 0.0
	for _, tag := range tags {
		affinity := affinities[tag]
		if affinity > affinityBlocked {
			affinity *= rarity.weight(tag)
		}
		sum += affinity
	}
	return clampFloat(sum/math.Sqrt(float64(len(tags))), -1, 1)
}

// watchedPenalty costs a completed watch more than a partial one. A watch that
// ended a session (endedSessionAt, zero when none did) is a favourite: its
// penalty fades with time and turns into a small bonus (a negative penalty).
func watchedPenalty(state watchState, endedSessionAt, now time.Time) float64 {
	penalty := 0.0
	switch state {
	case watchCompleted:
		penalty = penaltyWatchedCompleted
	case watchPartial:
		penalty = penaltyWatchedPartial
	}
	if penalty == 0 || endedSessionAt.IsZero() {
		return penalty
	}
	lift := rewatchLift(now.Sub(endedSessionAt))
	return penalty*(1-lift) - rewatchBonus*lift
}

// rewatchLift is how far a session ender's watched penalty has given way, in
// [0, 1]: 0 until rewatchFreeAfter, then linear over rewatchRampDuration.
func rewatchLift(sinceWatched time.Duration) float64 {
	return clampFloat(float64(sinceWatched-rewatchFreeAfter)/float64(rewatchRampDuration), 0, 1)
}

// fatiguePenalty grows with every ignored showing past the free ones.
func fatiguePenalty(ignoredShows int) float64 {
	return penaltyFatiguePerShow * float64(max(0, ignoredShows-fatigueFreeShows))
}

// freshness is 1 for an item released now, halving every freshnessHalfLife;
// undatedFreshness without a release date.
func freshness(published, now time.Time) float64 {
	if published.IsZero() {
		return undatedFreshness
	}
	return decayFactor(now.Sub(published), freshnessHalfLife)
}

// normalizedScore maps an open-ended ranking score onto 0..1.
func normalizedScore(total float64) float64 {
	return 1 / (1 + math.Exp(-scoreSteepness*(total-scoreMidpoint)))
}
