package suggest

import "time"

// Every tunable number of the suggestion engine lives in this file so match
// quality, popularity and taste can be weighed against each other in one place.

// Match quality: how well the typed text matches a candidate, in (0, 1]. The
// tiers are spaced so the popularity and taste factors below can reorder
// neighbouring tiers (a popular word-prefix match beats an obscure prefix
// match) but never lift a typo match above a real prefix match: the best
// fuzzy score (0.2 × 1.5 × 1.25) stays below the worst word-prefix entity
// score (0.55 × 0.75 × 0.95).
const (
	// qualityExact: the whole name, alias or query equals the typed text.
	qualityExact = 1.0
	// qualityPrefix: the name, alias or query starts with the typed text.
	qualityPrefix = 0.8
	// qualityWordPrefix: a later word of the name starts with the typed text
	// ("reid" → "Riley Reid").
	qualityWordPrefix = 0.55
	// qualityFuzzy is a typo-tolerant prefix match with one edit; it is
	// divided by the edit count, so two edits score half.
	qualityFuzzy = 0.2
)

// Typo tolerance. Fuzzy matching only runs when the exact and prefix matches
// leave the list short, and only for entity names.
const (
	// fuzzyMinimumLength is the shortest typed text (in letters) that gets
	// typo tolerance; shorter prefixes match too much with one edit.
	fuzzyMinimumLength = 4
	// fuzzyTwoEditLength is the typed length from which two edits are allowed.
	fuzzyTwoEditLength = 7
)

// Score = quality × (1 + popularityWeight × popularity) × (1 + affinityWeight
// × affinity) × kind weight. popularity is in [0, 1], affinity in [-1, 1].
const (
	popularityWeight = 0.5
	affinityWeight   = 0.25
)

// kindWeights nudges whole kinds against each other: the user's own past
// searches first, entities next, mined title phrases last.
var kindWeights = map[Kind]float64{
	KindRecent:    1.1,
	KindTag:       1.0,
	KindPerformer: 1.0,
	KindStudio:    0.95,
	KindQuery:     0.8,
}

// Popularity: counts are log-scaled against a reference count that maps to
// popularity 1 (anything above it is capped).
const (
	// entityReferenceScenes is the scene count of a fully popular entity.
	entityReferenceScenes = 500
	// phraseReferenceCount is how many titles a fully popular phrase appears in.
	phraseReferenceCount = 200
	// historyReferenceCount is how often a fully popular past search was run.
	historyReferenceCount = 10
	// historyCountShare is how much of a past search's popularity comes from
	// its run count; the rest comes from how recently it ran.
	historyCountShare = 0.5
	// historyHalfLife is the age at which a past search's recency halves.
	historyHalfLife = 14 * 24 * time.Hour
)

// Output sizes.
const (
	// DefaultLimit is the number of suggestions when the caller gives none.
	DefaultLimit = 12
	// maxLimit caps what one call may ask for.
	maxLimit = 50
	// emptyRecentLimit is how many recent searches lead the empty-query list.
	emptyRecentLimit = 5
	// typedRecentLimit caps past searches among typed-query suggestions.
	typedRecentLimit = 4
	// queryCompletionLimit caps mined title phrases among typed-query
	// suggestions, so they don't crowd out entities.
	queryCompletionLimit = 5
	// tastePicksPerKind is how many top tags/performers/studios of the taste
	// profile are considered for the empty-query list.
	tastePicksPerKind = 8
	// historyScanLimit bounds how many past searches are read per call.
	historyScanLimit = 300
)

// Index: entity names and title phrases are held in memory, rebuilt in the
// background once stale, so a keystroke never scans the scene table.
const (
	// indexRefreshInterval is the age after which the index is rebuilt.
	indexRefreshInterval = 10 * time.Minute
	// indexBuildTimeout bounds one background rebuild.
	indexBuildTimeout = 2 * time.Minute
	// phraseSceneLimit is how many of the newest scene titles are mined.
	phraseSceneLimit = 50000
	// maxPhraseWords is the longest title n-gram offered as a completion.
	maxPhraseWords = 3
	// minimumPhraseCount drops phrases seen in fewer titles (one-off noise).
	minimumPhraseCount = 2
	// minimumPhraseLength drops phrases shorter than this many bytes.
	minimumPhraseLength = 3
	// profileCacheTTL is how long a computed taste profile is reused; the
	// profile aggregates the signal tables, too slow to redo per keystroke.
	profileCacheTTL = time.Minute
)
