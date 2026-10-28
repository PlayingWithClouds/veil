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
	weightOEvent = 5.0
	// weightSessionEnd is a watch nothing else followed (see sessionGap): the
	// video was the one, whether or not it was watched to the end.
	weightSessionEnd       = 4.0
	weightLike             = 3.0
	weightSave             = 1.5
	weightPreferenceChosen = 1.5
	// A watch another video followed: it held attention but didn't end the
	// session, a little more when it was watched to the end.
	weightCompletedWatch     = 1.0
	weightPartialWatch       = 0.5
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
	// sessionGap is how long nothing else may play after a watch for it to
	// have ended the session.
	sessionGap = 30 * time.Minute
	// sessionEndMinimumSeconds is how much a session's last watch must have
	// played to count as ending it, not as a look and leave.
	sessionEndMinimumSeconds = 60
)

// Candidate generation sizes.
const (
	// seedMinimumWeight is the decayed signal weight a scene needs to seed
	// related-scene candidates (a watch another video followed, fresh, or a
	// session-ending watch from up to two months ago).
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
	// The feeder's finds (feed_discovery) count for this long after they
	// were last listed; catalog finds are capped per plugin.
	feedDiscoveryWindow   = 14 * 24 * time.Hour
	catalogPerPluginLimit = 40
	// Taste and explore finds are capped per entity, so the newest run's
	// entity cannot fill the whole pool.
	tastePerEntityLimit       = 20
	tasteCandidateLimit       = 160
	exploreSiteCandidateLimit = 90
	explorePerEntityLimit     = 30
	popularPerSiteLimit       = 40
	catalogCandidateLimit     = 2000
	popularCandidateLimit     = 200
	// lowPoolThreshold is how few unwatched, not yet shown candidates a
	// ranking may have before it asks the feeder for a top-up.
	lowPoolThreshold = 100
	// categoryOverfetch widens a tag category's query so scoring has room to
	// reorder it.
	categoryOverfetch = 3
)

// Source priors: how much a candidate source is trusted before its content is
// scored, in [0, 1]. The site's own related list is collaborative filtering
// computed by the site over its whole audience, so it ranks highest.
var sourcePriors = map[string]float64{
	SourceRelatedSite:      1.0,
	SourceSubscription:     0.8,
	SourceRelatedPerformer: 0.7,
	SourcePerformer:        0.65,
	SourceTaste:            0.6,
	SourceRelatedTag:       0.55,
	SourceStudio:           0.5,
	SourceTag:              0.45,
	SourceRelatedTitle:     0.4,
	SourceSearch:           0.4,
	SourceAdjacentTag:      0.3,
	SourceCatalog:          0.3,
	SourcePopular:          0.3,
	SourceExploreSite:      0.3,
	SourceRandom:           0.2,
	SourceNewest:           0.1,
	// SourceRelatedWalk: scenes two or three hops from a seed. It sits below
	// the direct lists (the site's own and the library-built ones) because
	// every hop dilutes the evidence.
	SourceRelatedWalk: 0.5,
	// SourceExploreAttribute: scenes of performers the user has no signal on
	// yet whose looks match the performers they like.
	SourceExploreAttribute: 0.35,
}

// Prior shaping.
const (
	// rankDecay lowers a candidate's prior with its position within its source.
	rankDecay = 0.05
	// seedWeightInfluence is how much a seed's strength scales the prior of
	// its related scenes (0 = not at all, 1 = fully proportional).
	seedWeightInfluence = 0.5
	// multiSourceBonus is added to the prior per extra source that found the
	// same scene; newest proposes whatever was stored last and corroborates
	// nothing.
	multiSourceBonus = 0.1
	// tasteAffinityFloor is the share of the taste prior a find keeps however
	// thin the profile: a taste find's entity has positive affinity by
	// construction, and with few signals the affinity alone (~0.2) would put
	// every find below the catalog.
	tasteAffinityFloor = 0.5
)

// Ranking: score = affinity + prior + freshness + popularity − penalties.
const (
	rankWeightAffinity    = 1.0
	rankWeightSourcePrior = 0.6
	rankWeightFreshness   = 0.2
	// rankWeightPopularity scales the site-popularity term in [0, 1]; it
	// decides between candidates the profile has no opinion on.
	rankWeightPopularity = 0.35
	// freshnessHalfLife is the age at which a scene's freshness halves.
	freshnessHalfLife = 30 * 24 * time.Hour
	// undatedFreshnessDefault is an undated scene's freshness when no ranked
	// scene has a date to compare with (one half-life old).
	undatedFreshnessDefault = 0.5

	penaltyWatchedCompleted = 1.5
	penaltyWatchedPartial   = 0.4
	// Fatigue: each showing without a click beyond the free ones costs this.
	penaltyFatiguePerShow = 0.15
	fatigueFreeShows      = 1
	// penaltyDislikedNeighbor hits scenes a disliked scene lists as related.
	penaltyDislikedNeighbor = 0.5
)

// Rewatching: a video that ended a session did its job, so once it is not
// fresh in memory the watched penalty fades and turns into a small bonus.
// A watch another video followed keeps the full penalty.
const (
	// rewatchFreeAfter is how long after a session-ending watch the full
	// penalty stands.
	rewatchFreeAfter = 3 * 24 * time.Hour
	// rewatchRampDuration is how long the penalty then takes to give way
	// completely (after rewatchFreeAfter + rewatchRampDuration the scene
	// carries only the bonus).
	rewatchRampDuration = 11 * 24 * time.Hour
	// rewatchBonus is the score a fully lifted session ender gains instead
	// of a penalty.
	rewatchBonus = 0.3
	// rewatchRowMinAge is how long ago a session ender was last watched to
	// be offered on the "Watch again" shelf.
	rewatchRowMinAge = 7 * 24 * time.Hour
	// rewatchRowLimit caps the shelf's scenes.
	rewatchRowLimit = 30
	// rewatchRowPosition is where the shelf sits among the rows (0 = first).
	rewatchRowPosition = 1
)

// Moments: the tags of the markers around where a session-ending watch
// stopped, and around the scene's replay peak, are what the user came for.
const (
	// momentTagBoost is the share of a scene's weight its moment tags get on
	// top of the weight every tag of the scene already carries.
	momentTagBoost = 0.5
	// momentSecondsBefore and momentSecondsAfter bound the markers near the
	// stop point: the last minutes watched explain why it ended there.
	momentSecondsBefore = 120.0
	momentSecondsAfter  = 20.0
	// momentHeatMarginSeconds widens the replay peak's bucket on both sides.
	momentHeatMarginSeconds = 30.0
	// momentHeatPeakRatio is how far above the bucket average the busiest
	// heat bucket must be to count as a peak.
	momentHeatPeakRatio = 2.0
	// momentSceneLimit bounds how many session-ending scenes (the most
	// recent) are read for markers and heat.
	momentSceneLimit = 100
)

// Tag rarity: a tag on most scenes says little about taste. Its weight falls
// with the log of its share of the tagged scenes (inverse document frequency)
// from 1 for tags on rarityFullShare of them or fewer.
const (
	rarityFullShare = 0.05
	rarityFloor     = 0.05
	// rarityMinScenes is how many tagged scenes the library needs before the
	// shares mean anything; below it every tag weighs 1.
	rarityMinScenes = 50
	// libraryStatsTTL is how long the library-wide aggregates (tag shares,
	// performer attributes) are reused.
	libraryStatsTTL = 10 * time.Minute
)

// Runtime preference: the user's signals are spread over runtime bands (see
// runtimeBandSeconds), smoothed between neighbours.
const (
	// runtimeMinMass is the affinity mass a band needs to reach ±1: with
	// less evidence than a couple of session-ending watches the preference
	// stays proportionally weak.
	runtimeMinMass = 8.0
)

// Performer attributes: looks (hair, ethnicity, ...) liked on known
// performers are inherited, damped, by performers with no signal of their own.
const (
	// attributeInheritance damps an unknown performer's inherited affinity.
	attributeInheritance = 0.5
	// attributeMinMass is the evidence an attribute needs to reach ±1: two
	// fully liked performers sharing it.
	attributeMinMass = 2.0
	// attributeFullShare and attributeFloor shape the weight of an attribute
	// by how many profiled performers share it (hair:Brown says less than
	// hair:Red); values are categories, so the scale is coarser than a tag's.
	attributeFullShare = 0.15
	attributeFloor     = 0.2
	// attributeCandidatePerformers is how many unknown performers the
	// attribute source follows, attributeScenesPerPerformer how many scenes
	// of each, and attributeCandidateMinimum the inherited affinity a
	// performer needs to be followed.
	attributeCandidatePerformers = 8
	attributeScenesPerPerformer  = 5
	attributeCandidateMinimum    = 0.15
	// attributeAffinityFloor is the share of the source prior a find keeps
	// however faint the match (the inherited affinity is damped twice, by
	// attributeInheritance and by thin evidence).
	attributeAffinityFloor = 0.5
)

// Walk: a short personalized random walk over the related graph (scene_related,
// both directions) mixed with shared-performer and shared-rare-tag edges.
const (
	// walkHops is how many hops the walk goes from the seeds; the first hop
	// is what relatedCandidates already proposes.
	walkHops = 3
	// walkFrontierLimit caps the scenes expanded per hop (the strongest by
	// mass), so the queries stay small however dense the graph is.
	walkFrontierLimit = 100
	// walkDamping is the share of its mass a scene passes on per hop.
	walkDamping = 0.6
	// walkReverseWeight scales edges walked against their direction (scenes
	// that list this one as related).
	walkReverseWeight = 0.5
	// walkPerformerEdgeWeight and walkTagEdgeWeight are the edge weights of a
	// shared performer and a shared rare tag, against the related edges'
	// source priors (site 1.0 ... title 0.4).
	walkPerformerEdgeWeight = 0.6
	walkTagEdgeWeight       = 0.3
	// walkTagMinRarity is the rarity weight a tag needs to link scenes.
	walkTagMinRarity = 0.7
	// walkEntitiesPerScene caps the tags a scene links through, walkScenesPerEntity
	// the newest scenes taken per performer or tag.
	walkEntitiesPerScene = 3
	walkScenesPerEntity  = 6
	// walkCandidateLimit caps the scenes the walk proposes.
	walkCandidateLimit = 60
)

// Popularity: how a scene's standing on its own site is read. Views and
// ratings are compared to the site's other scenes (sites count views and
// rate on very different scales), then blended.
const (
	// popularityViewShare is the views' share of the blend when a scene has
	// both views and a rating.
	popularityViewShare = 0.7
	// popularityNeutral is the popularity of a scene the site gave neither
	// views nor a rating, or whose site has too few to compare with.
	popularityNeutral = 0.5
	// popularityMinSample is how many scenes of a site must carry a value
	// before percentiles against them mean anything.
	popularityMinSample = 20
)

// Affinity match: how a scene's features mix into one [-1, 1] value (the
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
)

// Source kinds: the groups of sources that share page slots.
const (
	// kindPersonal is what the user's own signals point at: the related
	// lists of watched scenes, top performers/studios/tags, subscriptions.
	kindPersonal = "personal"
	// kindTaste is the feeder's finds on the site pages of top entities.
	kindTaste = "taste"
	// kindCatalog is the feeder's catalog and most-viewed listings.
	kindCatalog = "catalog"
	// kindExplore is everything meant to widen the profile.
	kindExplore = "explore"
	// kindSearch is the unvisited stubs of recent searches.
	kindSearch = "search"
	// kindFiller is newest and anything unclassified: it only fills what the
	// other kinds leave open.
	kindFiller = "filler"
)

// kindQuotas is how many of a page's pageSize slots each kind may fill by
// default (learning shifts them, see learning.go), in the order the page is
// described. Filler has no quota. A kind with fewer items
// than slots hands the rest to the best remaining items overall, and search
// is the last kind to take such leftovers.
var kindQuotas = map[string]int{
	kindPersonal: 8,
	kindTaste:    5,
	kindCatalog:  5,
	kindExplore:  4,
	kindSearch:   2,
}

// Learning from clicks: per-source priors and per-kind quotas are nudged by
// how often each source's showings were opened (Thompson sampling, see
// learning.go). Every number here only bounds how far the data may move the
// hand-set values above.
const (
	// learnWindow is how far back showings are read; they decay with
	// AffinityHalfLife inside it. It stays under impressionRetention.
	learnWindow = 60 * 24 * time.Hour
	// learnAttributionWindow is how long after a showing an open of the same
	// scene (logged click or first visit) still counts as its result. Showings
	// younger than this are not read yet: their outcome is still open.
	learnAttributionWindow = 6 * time.Hour
	// learnDownstreamWindow is how long after an opened showing a watch that
	// ended the session or a like still adds to its reward.
	learnDownstreamWindow = 7 * 24 * time.Hour
	// A showing's reward is 0 when ignored, learnRewardClick when opened, plus
	// a bonus for a watch that ended the session and for a like, capped at 1 so a reward
	// never outweighs the showing it belongs to.
	learnRewardClick      = 0.6
	learnRewardSessionEnd = 0.25
	learnRewardLiked      = 0.25
	// learnPriorClicks is the strength of the Beta prior in pseudo-rewards:
	// every source starts as Beta(learnPriorClicks, learnPriorClicks·(1−r)/r)
	// around the pooled reward rate r, i.e. as if it had already been shown
	// learnPriorClicks/r times at the average rate. 30 means a source needs
	// roughly 30 rewards of evidence of its own before the data outweighs
	// the prior; at the current ~1% click rate that is ~1600 showings.
	learnPriorClicks = 30.0
	// learnMinShows is the decayed number of showings (all sources) below
	// which nothing is learned and the hand-set values stand.
	learnMinShows = 300.0
	// learnMinPooledRate floors the pooled reward rate so a window with
	// almost no opens does not give the prior an absurd pseudo-count.
	learnMinPooledRate = 0.002
	// A sampled rate is read relative to the pooled rate and clamped to
	// [learnMinMultiplier, learnMaxMultiplier]; that is the factor applied to
	// a source's hand-set prior.
	learnMinMultiplier = 0.5
	learnMaxMultiplier = 1.5
	// learnQuotaMaxFactor bounds a kind's learned quota above, as a multiple
	// of its default in kindQuotas (rounded down).
	learnQuotaMaxFactor = 1.5
	// learnLogInterval rate-limits the log line with the learned values.
	learnLogInterval = 30 * time.Minute
)

// kindQuotaMinimums is the fewest slots a learned quota may leave a kind, so
// the bandit cannot starve a kind it knows little about. Exploration keeps
// the most relative to its default: it exists to widen the data.
var kindQuotaMinimums = map[string]int{
	kindPersonal: 4,
	kindTaste:    2,
	kindCatalog:  2,
	kindExplore:  3,
	kindSearch:   1,
}

// paceLookahead is how many slots ahead of an even spread over the page a
// kind may run, so the best items of the strongest kinds still lead the page
// while no kind takes it in one block.
const paceLookahead = 2

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
