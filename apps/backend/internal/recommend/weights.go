package recommend

import "time"

// Every tunable number of the engine lives in this file so the relative
// strengths of signals, sources and penalties can be read side by side.

// Taste profile decay and signal windows.
const (
	// AffinityHalfLife is how fast a signal fades: its weight halves every
	// half-life.
	AffinityHalfLife = 30 * 24 * time.Hour
	// signalWindow bounds how far back signals are read. After six half-lives
	// a signal keeps under 2% of its weight.
	signalWindow = 6 * AffinityHalfLife
	// ignoredImpressionWindow bounds the "shown but not clicked" signal and
	// feed fatigue.
	ignoredImpressionWindow = 30 * 24 * time.Hour
	// impressionRetention is how long impression rows are kept at all.
	impressionRetention = 90 * 24 * time.Hour
)

// Signal weights per scene interaction, strongest positive first. Negative
// weights express distaste.
const (
	weightOEvent             = 5.0
	weightLike               = 3.0
	weightCompletedWatch     = 2.0
	weightSave               = 1.5
	weightPreferenceChosen   = 1.5
	weightPartialWatch       = 1.0
	weightClick              = 0.5
	weightIgnoredImpression  = -0.15
	weightPreferenceRejected = -0.5
	weightEarlyAbandon       = -1.5
	weightDislike            = -3.0
	// ignoredImpressionCap bounds how many ignored showings of one scene count.
	ignoredImpressionCap = 3
	// affinityBlocked pins a blocked tag/performer/studio to the bottom of the
	// normalized affinity range, regardless of signals.
	affinityBlocked = -1.0
)

// Watch interpretation.
const (
	// ratingNeutral splits the 1–10 user rating: above is a like, below a dislike.
	ratingNeutral = 5.0
	// A watch that stopped before either threshold is an early abandonment.
	earlyAbandonSeconds  = 10
	earlyAbandonFraction = 0.05
	// A watch that reached this share of the runtime counts as completed.
	completedFraction = 0.8
	// abandonGrace keeps a playback that just started from counting as abandoned.
	abandonGrace = 5 * time.Minute
)

// Candidate generation sizes.
const (
	// seedMinimumWeight is the decayed signal weight a scene needs to seed
	// related-scene candidates (a completed watch from up to a month ago).
	seedMinimumWeight = 1.0
	seedSceneLimit    = 12
	// relatedPerSeedSource caps each seed's edges per scene_related source.
	relatedPerSeedSource       = 24
	topPerformerLimit          = 5
	topTagLimit                = 5
	topStudioLimit             = 3
	scenesPerAffinityEntity    = 20
	adjacentTagLimit           = 4
	scenesPerAdjacentTag       = 10
	randomCandidateLimit       = 12
	subscriptionCandidateLimit = 60
	recentSearchLimit          = 8
	recentSearchWindow         = 14 * 24 * time.Hour
	scenesPerSearch            = 24
	newestCandidateLimit       = 120
	// categoryOverfetch widens a tag category's query so scoring has room to
	// reorder it.
	categoryOverfetch = 3
	// A search's stubs are the scenes first stored in this window around the
	// search_history row (logged when the search finishes, or just before the
	// background stub writes of the GraphQL search).
	searchResultsBefore = 3 * time.Minute
	searchResultsAfter  = time.Minute
)

// Source priors: how much a candidate source is trusted before its content is
// scored, in [0, 1]. The site's own related list is collaborative filtering
// computed by the site over its whole audience, so it ranks highest.
var sourcePriors = map[string]float64{
	SourceRelatedSite:      1.0,
	SourceSubscription:     0.8,
	SourceRelatedPerformer: 0.7,
	SourcePerformer:        0.65,
	SourceRelatedTag:       0.55,
	SourceStudio:           0.5,
	SourceTag:              0.45,
	SourceRelatedTitle:     0.4,
	SourceSearch:           0.4,
	SourceAdjacentTag:      0.3,
	SourceRandom:           0.2,
	SourceNewest:           0.1,
}

// Prior shaping.
const (
	// rankDecay lowers a candidate's prior with its position within its source.
	rankDecay = 0.05
	// seedWeightInfluence is how much a seed's strength scales the prior of
	// its related scenes (0 = not at all, 1 = fully proportional).
	seedWeightInfluence = 0.5
	// multiSourceBonus is added to the prior per extra source that found the
	// same scene.
	multiSourceBonus = 0.1
)

// Ranking: score = affinity + prior + freshness − penalties.
const (
	rankWeightAffinity    = 1.0
	rankWeightSourcePrior = 0.6
	rankWeightFreshness   = 0.2
	// freshnessHalfLife is the age at which a scene's freshness halves.
	freshnessHalfLife = 30 * 24 * time.Hour

	penaltyWatchedCompleted = 1.5
	penaltyWatchedPartial   = 0.4
	// Fatigue: each showing without a click beyond the free ones costs this.
	penaltyFatiguePerShow = 0.15
	fatigueFreeShows      = 1
	// penaltyDislikedNeighbor hits scenes a disliked scene lists as related.
	penaltyDislikedNeighbor = 0.5
)

// Affinity match: how a scene's features mix into one [-1, 1] value.
const (
	matchWeightPerformer = 0.4
	matchWeightTag       = 0.3
	matchWeightStudio    = 0.2
	matchWeightSite      = 0.1
)

// Re-ranking.
const (
	// pageSize is the span the diversity caps apply to.
	pageSize               = 24
	maxPerPerformerPerPage = 2
	maxPerStudioPerPage    = 3
	// maxPerSitePerPage keeps one plugin's fresh scrape from filling a page
	// (a third of it at most, so a page spans at least three sites).
	maxPerSitePerPage = 8
	// maxSameSiteRun is how many scenes from one plugin may follow each other.
	maxSameSiteRun = 2
	// maxSameSourceRun is how many items from one source may follow each other.
	maxSameSourceRun = 2
	// explorationEvery reserves every seventh slot (~15%) for exploration.
	explorationEvery = 7
)

// Output shaping.
const (
	// feedCacheTTL is how long a ranked feed serves later pages (offset > 0),
	// so pagination stays stable while impressions shift the scores.
	feedCacheTTL = 30 * time.Minute
	// feedFirstPageTTL is how long the first page (offset 0) keeps serving the
	// same ranking, so reloading Home doesn't reshuffle it. An explicit refresh
	// re-ranks regardless.
	feedFirstPageTTL = 10 * time.Minute
	// minRowItems drops rows too thin to be worth a shelf.
	minRowItems = 3
	// pairCooldown keeps scenes out of the A/B pair for a while after they
	// were compared.
	pairCooldown = 7 * 24 * time.Hour
	// ImpressionBatchLimit is the most impressions one call may record.
	ImpressionBatchLimit = 500
)
