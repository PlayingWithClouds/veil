package recommend

import "time"

// Every tunable number of the engine lives in this file so the relative
// strengths of signals, sources and penalties can be read side by side.

// Taste profile decay and signal windows.
const (
	// affinityHalfLife is how fast a signal fades: its weight halves every
	// half-life.
	affinityHalfLife = 30 * 24 * time.Hour
	// signalWindow bounds how far back events are read. After six half-lives
	// a signal keeps under 2% of its weight.
	signalWindow = 6 * affinityHalfLife
	// explicitDecayFloor is the share of its weight a rating or favourite
	// keeps however old: an explicit verdict doesn't wear off like a watch.
	explicitDecayFloor = 0.25
	// ignoredImpressionWindow bounds the "shown but not opened" signal and
	// feed fatigue.
	ignoredImpressionWindow = 30 * 24 * time.Hour
	// eventReadLimit caps the rows one signal query reads, newest first, so a
	// heavy user's log stays a bounded read.
	eventReadLimit = 5000
)

// Signal weights per interaction, strongest positive first. Negative weights
// express distaste.
const (
	weightO = 5.0
	// weightSessionEnd is a watch nothing else followed (see sessionGap): the
	// video was the one, whether or not it was watched to the end.
	weightSessionEnd = 4.0
	weightLike       = 3.0
	weightFavorite   = 2.0
	weightSave       = 1.5
	// A watch another video followed: it held attention but didn't end the
	// session, a little more when it was watched in full.
	weightCompletedWatch    = 1.0
	weightGalleryView       = 1.0
	weightPartialWatch      = 0.5
	weightChoice            = 0.5
	weightIgnoredImpression = -0.15
	weightSkip              = -1.0
	weightEarlyAbandon      = -1.5
	weightDislike           = -3.0
	// ignoredImpressionCap bounds how many ignored showings of one item count.
	ignoredImpressionCap = 3
	// affinityBlocked pins a blocked tag/performer/studio to the bottom of the
	// normalized affinity range, regardless of signals.
	affinityBlocked = -1.0
)

// Cam rooms (see camsignals.go). A room is live, not a fixed item, so its
// signals feed the taste through its tags, performer and site.
const (
	// weightCamWatch is a watch of a minute; longer watches grow with the log
	// of their length up to weightCamWatchCap.
	weightCamWatch    = 1.0
	weightCamWatchCap = 4.0
	// weightCamQuickLeave is a room left within camQuickLeaveSeconds.
	weightCamQuickLeave  = -0.5
	camQuickLeaveSeconds = 10
	// camTasteFactor is the share of a room's weight its tags, performer and
	// site take into the video taste: watching a room says less about which
	// videos to watch than watching a video.
	camTasteFactor = 0.5
	// camTasteTTL keeps the cam taste for a page of rooms and the next.
	camTasteTTL = time.Minute
)

// Ratings (UserState.rating, 0..10).
const (
	// ratingNeutral splits the rating: above leans like, below dislike; the
	// signal grows linearly to ±weightLike at 10 and 0.
	ratingNeutral = 5.0
	// ratingDislikeBelow excludes the item from recommendations.
	ratingDislikeBelow = 3.0
)

// Watch interpretation.
const (
	// Watches are judged by how long the video played (heatSpan events), not
	// by the position reached: viewers skip around to find what they like. A
	// watch that played less than earlyAbandonSeconds, or less than
	// earlyAbandonFraction of the runtime and under earlyAbandonCeiling, is an
	// early abandonment.
	earlyAbandonSeconds  = 10
	earlyAbandonFraction = 0.05
	earlyAbandonCeiling  = 60
	// A watch that played fullWatchSeconds, or this share of the runtime,
	// counts as watched in full, however it was skipped through.
	completedFraction = 0.8
	fullWatchSeconds  = 300
	// abandonGrace keeps a playback that just started from counting as abandoned.
	abandonGrace = 5 * time.Minute
	// sessionGap is how long nothing else may play after a watch for it to
	// have ended the session.
	sessionGap = 30 * time.Minute
	// sessionEndMinimumSeconds is how much a session's last watch must have
	// played to count as ending it, not as a look and leave.
	sessionEndMinimumSeconds = 60
)

// Candidate generation sizes.
const (
	// seedMinimumWeight is the decayed signal weight an item needs to seed
	// related candidates.
	seedMinimumWeight = 0.9
	seedLimit         = 12
	// seedEntityLimit caps the performers and rare tags of one seed that
	// propose candidates.
	seedEntityLimit       = 3
	topPerformerLimit     = 5
	topTagLimit           = 5
	topStudioLimit        = 3
	itemsPerEntity        = 20
	itemsPerSeedEntity    = 8
	adjacentTagLimit      = 4
	itemsPerAdjacentTag   = 10
	randomCandidateLimit  = 12
	recentSearchLimit     = 5
	recentSearchWindow    = 14 * 24 * time.Hour
	itemsPerSearch        = 24
	newestCandidateLimit  = 120
	similarSeedLimit      = 6
	similarCandidateLimit = 40
)

// Source priors: how much a candidate source is trusted before its content is
// scored, in [0, 1].
var sourcePriors = map[string]float64{
	sourceRelatedCompanion: 0.8,
	sourceRelatedSimilar:   0.75,
	sourceRelatedPerformer: 0.7,
	sourcePerformer:        0.65,
	sourceRelatedTag:       0.55,
	sourceRelatedStudio:    0.5,
	sourceStudio:           0.5,
	sourceTag:              0.45,
	sourceSearch:           0.4,
	sourceExploreAttribute: 0.35,
	sourceAdjacentTag:      0.3,
	sourceRandom:           0.2,
	sourceNewest:           0.1,
}

// Prior shaping.
const (
	// rankDecay lowers a candidate's prior with its position within its source.
	rankDecay = 0.05
	// seedWeightInfluence is how much a seed's strength scales the prior of
	// its related items (0 = not at all, 1 = fully proportional).
	seedWeightInfluence = 0.5
	// multiSourceBonus is added to the prior per extra source that found the
	// same item; newest corroborates nothing.
	multiSourceBonus = 0.1
)

// Ranking: score = affinity + prior + freshness − penalties.
const (
	rankWeightAffinity    = 1.0
	rankWeightSourcePrior = 0.6
	rankWeightFreshness   = 0.2
	// freshnessHalfLife is the age at which an item's freshness halves.
	freshnessHalfLife = 30 * 24 * time.Hour
	// undatedFreshness is the freshness of an item without a release date:
	// one half-life old, neither favoured nor held back.
	undatedFreshness = 0.5

	penaltyWatchedCompleted = 1.5
	penaltyWatchedPartial   = 0.4
	// Fatigue: each showing without an open beyond the free ones costs this.
	penaltyFatiguePerShow = 0.15
	fatigueFreeShows      = 1
	// scoreSteepness and scoreMidpoint map the open-ended score onto 0..1
	// (a logistic curve) for Match.Score.
	scoreSteepness = 2.0
	scoreMidpoint  = 0.5
)

// Rewatching: a video that ended a session did its job, so once it is not
// fresh in memory the watched penalty fades and turns into a small bonus.
const (
	rewatchFreeAfter    = 3 * 24 * time.Hour
	rewatchRampDuration = 11 * 24 * time.Hour
	rewatchBonus        = 0.3
)

// Moments: the tags of the segments around where a session-ending watch
// stopped, and around its replay peak, are what the user came for.
const (
	momentTagBoost      = 0.5
	momentSecondsBefore = 120.0
	momentSecondsAfter  = 20.0
	momentSceneLimit    = 100
)

// Replay heat, built from heatSpan events (and o positions).
const (
	// heatBucketSeconds is the resolution of a video's heat.
	heatBucketSeconds = 5.0
	// heatMinimumReplays is how often the busiest bucket must have played to
	// be a replay at all; heatPeakRatio how far above the mean of the played
	// buckets it must be.
	heatMinimumReplays = 2.0
	heatPeakRatio      = 1.5
	// heatRegionShare widens the peak to the neighbouring buckets at least
	// this share of it.
	heatRegionShare = 0.5
	// heatOWeight is the heat an o event adds to the heatOLeadSeconds before
	// its position.
	heatOWeight      = 3.0
	heatOLeadSeconds = 30.0
	// heatSpanReadLimit caps the spans read per video.
	heatSpanReadLimit = 2000
)

// Tag rarity: a tag on most items says little about taste. Its weight falls
// with the log of its share of the tagged items (inverse document frequency)
// from 1 for tags on rarityFullShare of them or fewer.
const (
	rarityFullShare = 0.05
	rarityFloor     = 0.05
	// rarityMinItems is how many tagged items the library needs before the
	// shares mean anything; below it every tag weighs 1.
	rarityMinItems = 50
	// rareTagMinimumWeight is the rarity weight a tag needs to relate items.
	rareTagMinimumWeight = 0.7
	// libraryStatsTTL is how long the library-wide aggregates (tag shares,
	// performer attributes) are reused.
	libraryStatsTTL = 10 * time.Minute
)

// Runtime preference.
const (
	// runtimeMinMass is the affinity mass a runtime band needs to reach ±1.
	runtimeMinMass = 8.0
)

// Performer attributes: looks liked on known performers are inherited, damped,
// by performers with no signal of their own.
const (
	attributeInheritance         = 0.5
	attributeMinMass             = 2.0
	attributeFullShare           = 0.15
	attributeFloor               = 0.2
	attributeCandidatePerformers = 8
	attributeItemsPerPerformer   = 5
	attributeCandidateMinimum    = 0.15
	attributeAffinityFloor       = 0.5
)

// Affinity match: how an item's features mix into one [-1, 1] value (the
// weights add up to 1).
const (
	matchWeightPerformer = 0.36
	matchWeightTag       = 0.27
	matchWeightStudio    = 0.18
	matchWeightSite      = 0.09
	matchWeightRuntime   = 0.10
)

// Re-ranking.
const (
	// pageSize is the span the diversity caps and quotas apply to.
	pageSize               = 24
	maxPerPerformerPerPage = 2
	maxPerStudioPerPage    = 3
	// maxPerSitePerPage keeps one site from filling a page.
	maxPerSitePerPage = 8
	maxSameSiteRun    = 2
	maxSameSourceRun  = 2
	// paceLookahead is how many slots ahead of an even spread over the page a
	// kind may run.
	paceLookahead = 2
)

// Source kinds: the groups of sources that share page slots.
const (
	kindPersonal = "personal"
	kindExplore  = "explore"
	kindSearch   = "search"
	// kindFiller is newest and anything unclassified: it only fills what the
	// other kinds leave open.
	kindFiller = "filler"
)

// kindQuotas is how many of a page's slots each kind may fill by default;
// learning moves them within kindQuotaMinimums and learnQuotaMaxFactor.
var kindQuotas = map[string]int{
	kindPersonal: 15,
	kindExplore:  6,
	kindSearch:   3,
}

// kindQuotaMinimums is the fewest slots a learned quota may leave a kind.
var kindQuotaMinimums = map[string]int{
	kindPersonal: 7,
	kindExplore:  3,
	kindSearch:   1,
}

// Learning from opens: per-source priors and per-kind quotas are nudged by how
// often each source's showings on recommendation surfaces were opened
// (Thompson sampling). Every number only bounds how far the data may move the
// hand-set values.
const (
	learnWindow = 60 * 24 * time.Hour
	// learnAttributionWindow is how long after a showing an open still counts
	// as its result; younger showings are not read yet.
	learnAttributionWindow = 6 * time.Hour
	// learnDownstreamWindow is how long after a showing a finish, o or like
	// adds to its reward.
	learnDownstreamWindow = 7 * 24 * time.Hour
	learnRewardOpen       = 0.6
	learnRewardDownstream = 0.4
	learnPriorRewards     = 30.0
	learnMinShows         = 300.0
	learnMinPooledRate    = 0.002
	learnMaxPooledRate    = 0.5
	learnMinMultiplier    = 0.5
	learnMaxMultiplier    = 1.5
	learnQuotaMaxFactor   = 1.5
)

// Output shaping.
const (
	// rankingTTL is how long a ranking serves later pages, so pagination stays
	// stable while impressions shift the scores; firstPageTTL is how long the
	// first page keeps the same ranking, so reloading Home doesn't reshuffle.
	rankingTTL   = 30 * time.Minute
	firstPageTTL = 10 * time.Minute
	// tasteEntryLimit caps the liked and the disliked entries per taste
	// dimension.
	tasteEntryLimit = 20
)

// Related items.
const (
	// A gallery and video linked to each other outrank a likely pair.
	relatedWeightLinked    = 1.5
	relatedWeightCompanion = 1.0
	relatedWeightSimilar   = 0.5
	relatedWeightPerformer = 0.4
	relatedWeightStudio    = 0.2
	relatedWeightTag       = 0.25
	// companionMinimumScore is the alike.Score a gallery and a video need to
	// count as shot together; companionDateSlackDays how far apart their
	// release dates may be.
	companionMinimumScore  = 0.5
	companionDateSlackDays = 3
	companionCandidates    = 50
)

// Shorts.
const (
	// shortMaxSeconds is the longest a short plays; verticalMaxSeconds the
	// longest vertical video that counts as a short.
	shortMaxSeconds    = 60.0
	shortMinSeconds    = 15.0
	markerShortSeconds = 30.0
	verticalMaxSeconds = 180
	shortsReadLimit    = 200
	momentVideoLimit   = 200
	// Base scores per kind of short: a saved clip is the user's own pick.
	shortBaseClip     = 0.9
	shortBaseMoment   = 0.8
	shortBaseSegment  = 0.5
	shortBaseVertical = 0.4
	// shortAffinityWeight scales the taste match of the short's video;
	// shortSkipPenalty costs a short skipped within shortSkipWindow.
	shortAffinityWeight = 0.5
	shortSkipPenalty    = 0.5
	shortSkipWindow     = 7 * 24 * time.Hour
	maxShortsPerVideo   = 2
)

// Embeddings.
const (
	// embeddingPageSize is how many vectors one scan query reads;
	// embeddingScanLimit bounds the vectors one similarity search compares.
	embeddingPageSize  = 2000
	embeddingScanLimit = 100000
	// similarMinimumCosine is the cosine similarity below which vectors say
	// nothing.
	similarMinimumCosine = 0.5
)
