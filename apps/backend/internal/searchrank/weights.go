// Package searchrank scores and orders live search results. The library's own
// matches and every site's results are scored the same way, fused across
// sites, and cross-site copies of one video are collapsed to the best one.
//
// A result's score is the weighted sum
//
//	score = WeightText*text + WeightSiteRank*fusedRank + WeightQuality*(quality-0.5)
//	      + personalWeight*personal + WeightEntity*entity + WeightClickMemory*memory
//
// where text, fusedRank, quality, personal, entity and memory are described on
// the weights below. fusedRank's per-site shares are scaled by what the user's
// clicks say about the site for this query (see History), and personalWeight
// is WeightPersonal scaled by how broad the query is (see PersonalWeight). The server sends the score with every result; clients order by it.
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
	// recommendation profile, for a query of middling specificity. Kept small
	// so search stays predictable: taste reorders near-equal results and never
	// lifts an irrelevant one. PersonalWeight moves it with the query.
	WeightPersonal = 0.3

	// WeightEntity scales the entity credit in [0, 1]: how well a result
	// belongs to the performers, studios and tags recognised in the query (see
	// EntityCredit). A scene credited to the performer the query names earns
	// two thirds of a perfect text match on top of its text score.
	WeightEntity = 2.0

	// WeightClickMemory scales the click memory in [0, 1]: scenes opened
	// before from the same or a similar query (see History.ClickMemory).
	WeightClickMemory = 1.2
)

// Entity credit shaping, see EntityCredit.
const (
	// EntityMentionCredit is the credit of a result that carries no stored
	// credit for the entity (a fresh stub) but names it in its title.
	EntityMentionCredit = 0.7

	// FuzzyEntityFactor scales the credit for an entity that was only found
	// by a near-miss spelling of the query, never by its exact name or alias.
	FuzzyEntityFactor = 0.6

	// EntityPerformerWeight, EntityStudioWeight and EntityTagWeight are the
	// shares of the recognised entities in the mean entity credit: a
	// performer says most about what the user wants, a tag least.
	EntityPerformerWeight = 1.0
	EntityStudioWeight    = 0.8
	EntityTagWeight       = 0.5
)

// Typo tolerance, see FuzzyAllowance and fieldCredit.
const (
	// FuzzyCreditFactor scales the credit of a query word that matches a field
	// word only within a typo or two, so an exact match always scores higher.
	FuzzyCreditFactor = 0.6

	// FuzzyMinimumWordLength is the shortest query word that tolerates a typo:
	// a short word one letter off is usually another word.
	FuzzyMinimumWordLength = 4

	// FuzzyLongWordLength is the word length from which two typos are allowed.
	FuzzyLongWordLength = 8

	// FuzzyMinimumTextMatch is the text match a trigram candidate needs to be
	// kept as a typo-tolerant library result.
	FuzzyMinimumTextMatch = 0.4

	// SynonymTextFactor scales the text match against an expansion of the
	// query (an alias spelled out as the entity's name, a corrected spelling):
	// close to the original, but a result matching what was typed wins a tie.
	SynonymTextFactor = 0.92
)

// Query specificity and the personal weight it sets, see Specificity.
const (
	// PersonalBroadFactor and PersonalSpecificFactor scale WeightPersonal for
	// the broadest and the most specific queries; in between it is linear in
	// the specificity. So the personal term ranges from 0.075 for "mia
	// khalifa in a hotel room" to 0.6 for "anal", never beyond twice the base.
	PersonalBroadFactor    = 2.0
	PersonalSpecificFactor = 0.25
)

// wordSpecificity is the specificity of a query by its word count (index 0 is
// one word), the last value holding for longer queries: a single word is a
// topic, five words are nearly a title.
var wordSpecificity = []float64{0.0, 0.2, 0.45, 0.65, 0.85}

const (
	// SpecificEntityBonus is added, scaled by the share of the query's words
	// that a recognised performer or studio covers: naming a person or a
	// channel narrows the query however short it is.
	SpecificEntityBonus = 0.5

	// BroadTagPenalty is subtracted, scaled by the share of the words that
	// are tag names: "blonde teen" asks for a topic, not a video.
	BroadTagPenalty = 0.3

	// LibraryFewMatches and LibraryManyMatches bound the library match count
	// that moves the specificity: at or below the first the query is narrow,
	// at or above the second broad, log-scaled between.
	LibraryFewMatches  = 5
	LibraryManyMatches = 200

	// LibraryBroadPenalty and LibraryNarrowBonus are the specificity change at
	// the broad and the narrow end of the library count.
	LibraryBroadPenalty = 0.3
	LibraryNarrowBonus  = 0.1
)

// What search clicks teach, see History.
const (
	// HistoryRetentionDays is how long search events are kept and read.
	HistoryRetentionDays = 90

	// ClickHalfLifeDays is the half-life of a click's weight, in the memory
	// and in the plugin rates.
	ClickHalfLifeDays = 30.0

	// ClickMemoryMinimumSimilarity is the least overlap (shared query words
	// over all words of both queries) for an earlier query to count as similar.
	ClickMemoryMinimumSimilarity = 0.5

	// PluginRatePriorShown is the pseudo-count of showings that pulls a
	// plugin's click rate toward the average, so a few clicks do not swing it.
	PluginRatePriorShown = 15.0

	// PluginRateMinimumShown is the weighted number of showings, over all
	// plugins, below which the history is too thin to move any plugin's share.
	PluginRateMinimumShown = 10.0

	// PluginShareSpread bounds a plugin's rank-share multiplier to
	// [1 - spread, 1 + spread].
	PluginShareSpread = 0.5

	// HistoryRowLimit bounds the rows read per lookup.
	HistoryRowLimit = 5000
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

// Entity recognition, see EntityIndex.
const (
	// FuzzyEntityMinimumLength is the shortest query (in letters, spaces
	// included) whose near-miss spelling is taken for an entity's name: one
	// typo is allowed from there, two from twice that length.
	FuzzyEntityMinimumLength = 6
)

// How the search handler widens a query.
const (
	// FuzzyFallbackBelow is the number of exact library scene matches under
	// which the trigram index is asked for typo-tolerant candidates too.
	FuzzyFallbackBelow = 5

	// MaxEntityPages is how many recognised performers and studios get their
	// site page listed beside the keyword search.
	MaxEntityPages = 2
)
