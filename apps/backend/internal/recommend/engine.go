// Package recommend is the recommendation engine: candidate generation from
// cheap local sources, a transparent weighted ranking against a time-decayed
// taste profile, and diversity-aware re-ranking, for videos and galleries.
// It also finds related items, builds the shorts feed and summarizes taste.
//
// Everything it knows about the user comes from the event log and UserState
// (see events.go for the event data it reads). The taste profile is computed
// on demand rather than cached in a table: the signal queries are bounded by
// a time window and a row cap, so there is no cache to invalidate on every
// like or watch. The ranked feed itself is memoized for pagination.
//
// Blocked media, media crediting a blocked performer, studio or tag, and
// trashed media are never returned (library.FilterConditions decides).
package recommend

import (
	"context"
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Engine ranks recommendations. Safe for concurrent use.
type Engine struct {
	database *db.DB
	// now is the clock, swappable in tests.
	now func() time.Time
	// random drives the Thompson sampling of learned priors; used under mutex
	// (rankings are built one at a time), swappable in tests.
	random sampler
	// stats caches the library-wide aggregates (tag rarity, performer looks).
	stats statsCache

	mutex sync.Mutex
	// rankings holds one ranking per set of kinds (rankingKey): the mixed
	// feed of Home and the galleries-only feed of the gallery pages.
	rankings map[string]*ranking
}

// ranking is one full ranking pass: the feed order.
type ranking struct {
	builtAt time.Time
	items   []scoredItem
}

// New returns an engine over the database.
func New(database *db.DB) *Engine {
	return &Engine{database: database, now: time.Now, random: rand.New(rand.NewSource(time.Now().UnixNano())), rankings: map[string]*ranking{}}
}

// rankedKinds is what a ranking proposes when the filter names no kind.
var rankedKinds = []schema.RecordKind{schema.RecordKindVideo, schema.RecordKindGallery}

// Recommendations returns one page of the ranked feed of videos and
// galleries, or of the ones filter.Kinds names: a feed of one kind is ranked
// on its own, so its candidates aren't crowded out by the other's. The first
// page re-ranks once the ranking is older than firstPageTTL (a reload shows
// the same feed); later pages read the same ranking while it is younger than
// rankingTTL, so pages neither overlap nor skip as impressions shift the
// scores. Items blocked or trashed since the ranking was built are left out,
// and so are those filter doesn't match; pages count over what the filter
// leaves.
func (e *Engine) Recommendations(ctx context.Context, filter schema.Filter, limit, offset int) ([]schema.Match, error) {
	kinds := rankingKinds(filter)
	if len(kinds) == 0 {
		return []schema.Match{}, nil
	}
	maxAge := rankingTTL
	if offset == 0 {
		maxAge = firstPageTTL
	}
	current, err := e.currentRanking(ctx, kinds, maxAge)
	if err != nil {
		return nil, err
	}
	eligible, err := e.stillEligible(ctx, current.items, filter)
	if err != nil {
		return nil, err
	}
	page := pageOf(eligible, limit, offset)
	if err := e.rememberServed(ctx, page); err != nil {
		return nil, err
	}
	return recommendationMatches(page), nil
}

// rankingKinds is the ranked kinds filter asks for, in rankedKinds order:
// all of them when it names none, none when it names only kinds that aren't
// ranked.
func rankingKinds(filter schema.Filter) []schema.RecordKind {
	if len(filter.Kinds) == 0 {
		return rankedKinds
	}
	kinds := []schema.RecordKind{}
	for _, kind := range rankedKinds {
		if slices.Contains(filter.Kinds, schema.MediaKind(kind)) {
			kinds = append(kinds, kind)
		}
	}
	return kinds
}

// rankingKey names the ranking of a set of kinds.
func rankingKey(kinds []schema.RecordKind) string {
	names := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		names = append(names, string(kind))
	}
	return strings.Join(names, ",")
}

// currentRanking returns the memoized ranking of kinds, rebuilding it when it
// is older than maxAge.
func (e *Engine) currentRanking(ctx context.Context, kinds []schema.RecordKind, maxAge time.Duration) (*ranking, error) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	key := rankingKey(kinds)
	if memoized := e.rankings[key]; memoized != nil && e.now().Sub(memoized.builtAt) < maxAge {
		return memoized, nil
	}
	built, err := e.buildRanking(ctx, kinds)
	if err != nil {
		return nil, err
	}
	e.rankings[key] = built
	return built, nil
}

// buildRanking runs the full pipeline over items of kinds: snapshot,
// candidates, scoring, duplicates, reasons, re-ranking.
func (e *Engine) buildRanking(ctx context.Context, kinds []schema.RecordKind) (*ranking, error) {
	state, err := e.takeSnapshot(ctx, kinds)
	if err != nil {
		return nil, err
	}
	pool, err := e.generateCandidates(ctx, state)
	if err != nil {
		return nil, err
	}
	features, err := e.loadFeatures(ctx, pool.order)
	if err != nil {
		return nil, err
	}
	scored, err := e.dropDuplicates(ctx, scoreCandidates(pool.list(), features, state))
	if err != nil {
		return nil, err
	}
	if err := e.nameReasons(ctx, scored); err != nil {
		return nil, err
	}
	return &ranking{builtAt: state.now, items: rerank(scored, state.policy.pageQuotas())}, nil
}

// nameReasons fills in the names of the performers, studios and tags the
// reasons refer to.
func (e *Engine) nameReasons(ctx context.Context, items []scoredItem) error {
	ids := map[string]bool{}
	for _, item := range items {
		if item.reason.named() {
			ids[item.reason.entityID] = true
		}
	}
	names, err := e.entityNames(ctx, setKeys(ids))
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.reason.named() {
			item.reason.entityName = names[item.reason.entityID]
		}
	}
	return nil
}

// stillEligible drops the items trashed or blocked since the ranking was
// built, and those filter leaves out, keeping the order.
func (e *Engine) stillEligible(ctx context.Context, items []scoredItem, filter schema.Filter) ([]scoredItem, error) {
	refs := make([]schema.RecordRef, 0, len(items))
	for _, item := range items {
		refs = append(refs, item.item)
	}
	eligible, err := e.eligibleRefs(ctx, refs, filter)
	if err != nil {
		return nil, err
	}
	kept := make([]scoredItem, 0, len(items))
	for _, item := range items {
		if eligible[item.item] {
			kept = append(kept, item)
		}
	}
	return kept, nil
}

// eligibleRefs reports which of the referenced videos and galleries may be
// shown: not trashed, not blocked, crediting nothing blocked, and matching
// filter.
func (e *Engine) eligibleRefs(ctx context.Context, refs []schema.RecordRef, filter schema.Filter) (map[schema.RecordRef]bool, error) {
	idsByKind := map[schema.RecordKind][]string{}
	for _, ref := range refs {
		idsByKind[ref.Kind] = append(idsByKind[ref.Kind], ref.Id)
	}
	out := make(map[schema.RecordRef]bool, len(refs))
	for kind, ids := range idsByKind {
		if !isMediaKind(kind) {
			continue
		}
		conditions, vars := library.FilterConditions(kind, filter)
		vars["ids"] = ids
		found, err := e.database.Strings(ctx,
			`SELECT "media"."id"`+library.MediaFrom(kind)+` WHERE `+strings.Join(conditions, " AND ")+
				` AND "media"."id" IN (SELECT value FROM json_each($ids))`, vars)
		if err != nil {
			return nil, fmt.Errorf("check eligible %s: %w", kind, err)
		}
		for _, id := range found {
			out[schema.RecordRef{Kind: kind, Id: id}] = true
		}
	}
	return out, nil
}

// pageOf slices one page out of the feed.
func pageOf(items []scoredItem, limit, offset int) []scoredItem {
	if offset < 0 || offset >= len(items) || limit <= 0 {
		return []scoredItem{}
	}
	return items[offset:min(offset+limit, len(items))]
}

// recommendationMatches turns ranked items into matches with their reasons.
func recommendationMatches(items []scoredItem) []schema.Match {
	out := make([]schema.Match, 0, len(items))
	for _, item := range items {
		evidence := []schema.Evidence{}
		if item.rewatch {
			evidence = append(evidence, veilEvidence("Watch again"))
		}
		evidence = append(evidence, veilEvidence(item.reason.text()))
		out = append(out, schema.Match{Item: item.item, Score: float32Pointer(normalizedScore(item.score)), Evidence: evidence})
	}
	return out
}
