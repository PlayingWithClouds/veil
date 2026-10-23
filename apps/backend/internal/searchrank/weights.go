// Package searchrank scores and orders live search results. The library's own
// matches and every site's results are scored the same way, fused across
// sites, and cross-site copies of one video are collapsed to the best one.
//
// A result's score is the weighted sum
//
//	score = WeightText*text + WeightSiteRank*fusedRank + WeightQuality*(quality-0.5) + WeightPersonal*personal
//
// where text, fusedRank, quality and personal are described on the weights
// below. The server sends the score with every result; clients order by it.
package searchrank

// Weights of the score's terms. They live here together so the balance of the
// ranking can be read and tuned in one place: text relevance dominates (a full
// match is worth more than the whole spread of site ranks), the site's own
// order breaks ties between equally relevant results, and quality and taste
// only nudge.
const (
	// WeightText scales the text match in [0, 1] of the query against the
	// title, performers, studio and tags (see TextMatch).
	WeightText = 3.0

	// WeightSiteRank scales the fused rank: the reciprocal-rank-fusion share of
	// every list that carries the result, each list normalized so its first
	// place is 1 (see rankShare). A video two sites both list, high on both,
	// can reach about 2.
	WeightSiteRank = 1.5

	// WeightQuality scales the centred quality in [-0.5, 0.5]: the mean of the
	// site's rating (0..10, as a fraction) and the log of its view count,
	// 0.5 when the listing shows neither.
	WeightQuality = 0.4

	// WeightPersonal scales the personal affinity in [-1, 1] taken from the
	// recommendation profile. Kept small so search stays predictable: taste
	// reorders near-equal results and never lifts an irrelevant one.
	WeightPersonal = 0.3
)

// Reciprocal rank fusion and quality shaping.
const (
	// RankFusionConstant is k in 1/(k+rank). A site list holds about 20 results
	// per page, so a small k keeps first place clearly ahead of twentieth
	// (1.0 against 0.37) where the usual 60 would flatten them.
	RankFusionConstant = 10.0

	// ViewCountCeiling is the view count that earns the full view score; the
	// score is log-scaled up to it.
	ViewCountCeiling = 10_000_000.0

	// NeutralQuality is the quality of a result whose listing carries no rating
	// or view count: neither rewarded nor punished.
	NeutralQuality = 0.5
)

// Text match shaping, see TextMatch.
const (
	// coverageShare is how much of the text score comes from how many query
	// words the result matches; the rest is the whole query appearing as a
	// phrase in the title.
	coverageShare = 0.75

	// phraseAtStartBonus and phraseInsideBonus are the phrase shares for a
	// title that starts with, or merely contains at a word boundary, the whole
	// query. They add up with coverageShare to 1 only for a title starting with
	// the query.
	phraseAtStartBonus = 0.25
	phraseInsideBonus  = 0.15
)

// Credit a query word earns by where, and how fully, it is found. A word is
// found fully when it equals a word of the field, partially when it is the
// start of one (the user may still be typing) or, for titles only, inside one.
const (
	titleWordCredit   = 1.0
	titlePrefixCredit = 0.85
	titleInsideCredit = 0.5

	performerWordCredit   = 0.95
	performerPrefixCredit = 0.8

	studioWordCredit   = 0.8
	studioPrefixCredit = 0.7

	tagWordCredit   = 0.55
	tagPrefixCredit = 0.45
)

// Weights of the personal affinity's dimensions, mirroring how the
// recommendation ranking mixes the same profile (strongest performer, tags
// damped by their count, studio, site).
const (
	PersonalWeightPerformer = 0.4
	PersonalWeightTag       = 0.3
	PersonalWeightStudio    = 0.2
	PersonalWeightSite      = 0.1
)
