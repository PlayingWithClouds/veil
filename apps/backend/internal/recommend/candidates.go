package recommend

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Candidate source names. Every recommended item carries the one it was
// served from; learning credits its opens to it.
const (
	sourceRelatedCompanion = "related:companion"
	sourceRelatedSimilar   = "related:similar"
	sourceRelatedPerformer = "related:performer"
	sourceRelatedStudio    = "related:studio"
	sourceRelatedTag       = "related:tag"
	sourcePerformer        = "affinity:performer"
	sourceStudio           = "affinity:studio"
	sourceTag              = "affinity:tag"
	sourceExploreAttribute = "explore:attribute"
	sourceAdjacentTag      = "explore:adjacent"
	sourceRandom           = "explore:random"
	sourceSearch           = "search"
	sourceNewest           = "newest"
)

// candidate is an item some source proposes, before scoring.
type candidate struct {
	item   schema.RecordRef
	source string
	// prior is the source's hand-set trust in this item, in [0, 1]; scoring
	// scales it by what the source earned (learnedPolicy).
	prior float64
	// sourceCount is how many sources proposed the item.
	sourceCount int
	reason      reason
}

// candidatePool merges candidates by item, keeping the strongest source's
// prior and reason and counting how many sources agreed.
type candidatePool struct {
	byItem map[schema.RecordRef]*candidate
	order  []schema.RecordRef
}

// newCandidatePool returns an empty pool.
func newCandidatePool() *candidatePool {
	return &candidatePool{byItem: map[schema.RecordRef]*candidate{}}
}

// add merges one candidate into the pool.
func (pool *candidatePool) add(proposed candidate) {
	existing, found := pool.byItem[proposed.item]
	if !found {
		proposed.sourceCount = 1
		pool.byItem[proposed.item] = &proposed
		pool.order = append(pool.order, proposed.item)
		return
	}
	if corroborates(existing.source, proposed.source) {
		existing.sourceCount++
	}
	if proposed.prior > existing.prior {
		existing.prior = proposed.prior
		existing.source = proposed.source
		existing.reason = proposed.reason
	}
}

// corroborates reports whether a second source proposing an item is
// independent evidence for it: a different source, and neither is newest,
// which lists whatever was stored last.
func corroborates(existingSource, proposedSource string) bool {
	if existingSource == sourceNewest || proposedSource == sourceNewest {
		return false
	}
	return existingSource != proposedSource
}

// list returns the pooled candidates in first-seen order.
func (pool *candidatePool) list() []*candidate {
	out := make([]*candidate, 0, len(pool.order))
	for _, ref := range pool.order {
		out = append(out, pool.byItem[ref])
	}
	return out
}

// candidateGenerator proposes candidates from one source.
type candidateGenerator func(ctx context.Context, state *snapshot) ([]candidate, error)

// generateCandidates runs every source and pools the results of the
// snapshot's kinds; a source may still propose others (a gallery seed's
// video).
func (e *Engine) generateCandidates(ctx context.Context, state *snapshot) (*candidatePool, error) {
	generators := []candidateGenerator{
		e.relatedCandidates,
		e.similarCandidates,
		e.affinityCandidates,
		e.lookalikeCandidates,
		e.adjacentTagCandidates,
		e.randomCandidates,
		e.searchCandidates,
		e.newestCandidates,
	}
	pool := newCandidatePool()
	for _, generate := range generators {
		found, err := generate(ctx, state)
		if err != nil {
			return nil, err
		}
		for _, proposed := range found {
			if slices.Contains(state.kinds, proposed.item.Kind) {
				pool.add(proposed)
			}
		}
	}
	return pool, nil
}

// rankedPrior lowers a source prior by the item's position within the source.
func rankedPrior(prior float64, rank int) float64 {
	return prior / (1 + float64(rank)*rankDecay)
}

// itemCandidates turns a ranked item list from one source into candidates.
func itemCandidates(items []schema.RecordRef, source string, prior float64, why reason) []candidate {
	out := make([]candidate, 0, len(items))
	for rank, item := range items {
		out = append(out, candidate{item: item, source: source, prior: rankedPrior(prior, rank), reason: why})
	}
	return out
}

// entityColumns tells entityItems how to reach each credit of a media row:
// the join that lists it and the expression that is the entity id.
var entityColumns = map[string]struct{ join, value string }{
	"performers": {join: `, json_each("media"."performers") AS "linked"`, value: `"linked".value`},
	"tags":       {join: `, json_each("media"."tags") AS "linked"`, value: `"linked".value`},
	"studio":     {join: ``, value: `"media"."studio"`},
}

// entityItemsQuery lists, per entity of $entities, its media in an order, at
// most $limit each. %s are the media table, the credit join, the entity
// expression (twice) and the order.
const entityItemsQuery = `SELECT "id", "entity", "createdAt" FROM (
	SELECT "media"."id" AS "id", %[3]s AS "entity", "media"."createdAt" AS "createdAt",
		ROW_NUMBER() OVER (PARTITION BY %[3]s ORDER BY %[4]s) AS "position"
	FROM %[1]s AS "media"%[2]s
	WHERE "media"."deletedAt" IS NULL AND %[3]s IN (SELECT value FROM json_each($entities))
) WHERE "position" <= $limit`

// entityItems lists, per performer, tag or studio id (column "performers",
// "tags" or "studio"), its newest items of kinds, at most perEntity; random
// picks a random sample instead of the newest. Credits aren't indexed, so
// this is one scan per media table.
func (e *Engine) entityItems(ctx context.Context, kinds []schema.RecordKind, column string, entityIDs []string, perEntity int, random bool) (map[string][]schema.RecordRef, error) {
	if len(entityIDs) == 0 {
		return map[string][]schema.RecordRef{}, nil
	}
	order := `"media"."createdAt" DESC`
	if random {
		order = `RANDOM()`
	}
	found := map[string][]datedRef{}
	for _, kind := range kinds {
		query := fmt.Sprintf(entityItemsQuery, db.Quote(tableOf(kind)), entityColumns[column].join, entityColumns[column].value, order)
		rows, err := e.database.Query(ctx, query, db.Vars{"entities": entityIDs, "limit": perEntity})
		if err != nil {
			return nil, fmt.Errorf("items by %s: %w", column, err)
		}
		for _, row := range rows {
			entityID := rowString(row, "entity")
			found[entityID] = append(found[entityID], datedRef{ref: schema.RecordRef{Kind: kind, Id: rowString(row, "id")}, at: rowString(row, "createdAt")})
		}
	}
	out := make(map[string][]schema.RecordRef, len(found))
	for entityID, items := range found {
		out[entityID] = newestRefs(items, perEntity, random)
	}
	return out, nil
}

// datedRef is an item with when it was stored.
type datedRef struct {
	ref schema.RecordRef
	at  string
}

// newestRefs orders the items newest first (kept as found when random) and
// keeps at most limit.
func newestRefs(items []datedRef, limit int, random bool) []schema.RecordRef {
	if !random {
		sort.SliceStable(items, func(left, right int) bool { return items[left].at > items[right].at })
	}
	out := make([]schema.RecordRef, 0, min(limit, len(items)))
	for _, item := range items[:min(limit, len(items))] {
		out = append(out, item.ref)
	}
	return out
}

// relatedCandidates proposes, per seed, its linked gallery or video and the
// newest items sharing its performers, studio or rare tags. A stronger seed
// lends its related items a higher prior.
func (e *Engine) relatedCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	seeds := state.taste.seeds
	if len(seeds) == 0 {
		return nil, nil
	}
	performers, studios, tags := seedEntities(state)
	byPerformer, err := e.entityItems(ctx, state.kinds, "performers", setKeys(performers), itemsPerSeedEntity, false)
	if err != nil {
		return nil, err
	}
	byStudio, err := e.entityItems(ctx, state.kinds, "studio", setKeys(studios), itemsPerSeedEntity, false)
	if err != nil {
		return nil, err
	}
	byTag, err := e.entityItems(ctx, state.kinds, "tags", setKeys(tags), itemsPerSeedEntity, false)
	if err != nil {
		return nil, err
	}
	out := []candidate{}
	for _, current := range seeds {
		features := state.features[current.item]
		factor := seedFactor(current, seeds[0].weight)
		why := reason{kind: reasonRelated, entityID: current.item.Id, entityName: current.title}
		if companion, linked := features.companionRef(); linked {
			out = append(out, candidate{item: companion, source: sourceRelatedCompanion, prior: sourcePriors[sourceRelatedCompanion] * factor, reason: why})
		}
		out = append(out, seedCandidates(features.performers[:min(seedEntityLimit, len(features.performers))], byPerformer, sourceRelatedPerformer, factor, why)...)
		out = append(out, seedCandidates([]string{features.studio}, byStudio, sourceRelatedStudio, factor, why)...)
		out = append(out, seedCandidates(rareTags(features.tags, state.taste.rarity), byTag, sourceRelatedTag, factor, why)...)
	}
	return out, nil
}

// seedEntities collects the performers, studios and rare tags of the seeds.
func seedEntities(state *snapshot) (performers, studios, tags map[string]bool) {
	performers, studios, tags = map[string]bool{}, map[string]bool{}, map[string]bool{}
	for _, current := range state.taste.seeds {
		features := state.features[current.item]
		for _, performer := range features.performers[:min(seedEntityLimit, len(features.performers))] {
			performers[performer] = true
		}
		if features.studio != "" {
			studios[features.studio] = true
		}
		for _, tag := range rareTags(features.tags, state.taste.rarity) {
			tags[tag] = true
		}
	}
	return performers, studios, tags
}

// seedFactor scales a seed's related priors by its strength relative to the
// strongest seed.
func seedFactor(current seed, strongest float64) float64 {
	return 1 - seedWeightInfluence + seedWeightInfluence*current.weight/strongest
}

// seedCandidates turns the items of a seed's entities into candidates of one
// source; the seed itself is left out by the pool's eligibility later.
func seedCandidates(entityIDs []string, itemsByEntity map[string][]schema.RecordRef, source string, factor float64, why reason) []candidate {
	out := []candidate{}
	for _, entityID := range entityIDs {
		out = append(out, itemCandidates(itemsByEntity[entityID], source, sourcePriors[source]*factor, why)...)
	}
	return out
}

// rareTags lists the tags rare enough to relate items, rarest first, at most
// seedEntityLimit.
func rareTags(tags []string, rarity *tagRarity) []string {
	rare := []string{}
	for _, tag := range tags {
		if rarity.weight(tag) >= rareTagMinimumWeight {
			rare = append(rare, tag)
		}
	}
	sort.SliceStable(rare, func(left, right int) bool { return rarity.weight(rare[left]) > rarity.weight(rare[right]) })
	return rare[:min(seedEntityLimit, len(rare))]
}

// similarCandidates proposes the items whose embeddings are closest to the
// strongest seeds'; nothing when no seed has vectors.
func (e *Engine) similarCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	seeds := state.taste.seeds[:min(similarSeedLimit, len(state.taste.seeds))]
	if len(seeds) == 0 {
		return nil, nil
	}
	refs := make([]schema.RecordRef, 0, len(seeds))
	titles := map[schema.RecordRef]string{}
	for _, current := range seeds {
		refs = append(refs, current.item)
		titles[current.item] = current.title
	}
	found, err := e.similarTo(ctx, refs, similarCandidateLimit)
	if err != nil {
		return nil, err
	}
	out := make([]candidate, 0, len(found))
	for _, item := range found {
		why := reason{kind: reasonRelated, entityID: item.owner.Id, entityName: titles[item.owner]}
		out = append(out, candidate{item: item.item, source: sourceRelatedSimilar, prior: sourcePriors[sourceRelatedSimilar] * item.similarity, reason: why})
	}
	return out, nil
}

// affinityCandidates proposes the newest items of the top-affinity
// performers, studios and tags, each entity's prior scaled by its affinity.
func (e *Engine) affinityCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	dimensions := []struct {
		column     string
		kind       string
		source     string
		affinities map[string]float64
		limit      int
	}{
		{"performers", reasonPerformer, sourcePerformer, state.taste.performers, topPerformerLimit},
		{"studio", reasonStudio, sourceStudio, state.taste.studios, topStudioLimit},
		{"tags", reasonTag, sourceTag, state.taste.tags, topTagLimit},
	}
	out := []candidate{}
	for _, dimension := range dimensions {
		top := topEntities(dimension.affinities, dimension.limit)
		itemsByEntity, err := e.entityItems(ctx, state.kinds, dimension.column, top, itemsPerEntity, false)
		if err != nil {
			return nil, err
		}
		for _, entityID := range top {
			prior := sourcePriors[dimension.source] * dimension.affinities[entityID]
			out = append(out, itemCandidates(itemsByEntity[entityID], dimension.source, prior, reason{kind: dimension.kind, entityID: entityID})...)
		}
	}
	return out, nil
}

// lookalikeCandidates proposes the newest items of unknown performers who
// look like the ones the user likes: exploration, only her looks vouch for her.
func (e *Engine) lookalikeCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	performers := lookalikes(state.taste, attributeCandidatePerformers)
	if len(performers) == 0 {
		return nil, nil
	}
	performerIDs := make([]string, 0, len(performers))
	for _, performer := range performers {
		performerIDs = append(performerIDs, performer.performerID)
	}
	itemsByPerformer, err := e.entityItems(ctx, state.kinds, "performers", performerIDs, attributeItemsPerPerformer, false)
	if err != nil {
		return nil, err
	}
	out := []candidate{}
	for _, performer := range performers {
		match := min(1, performer.affinity/attributeInheritance)
		prior := sourcePriors[sourceExploreAttribute] * (attributeAffinityFloor + (1-attributeAffinityFloor)*match)
		why := reason{kind: reasonLookalike, entityID: performer.performerID}
		out = append(out, itemCandidates(itemsByPerformer[performer.performerID], sourceExploreAttribute, prior, why)...)
	}
	return out, nil
}

// adjacentTagCandidates proposes a random sample of items from tags adjacent
// to the top tags: the exploration share that widens the profile.
func (e *Engine) adjacentTagCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	itemsByTag, err := e.entityItems(ctx, state.kinds, "tags", state.taste.adjacentTags, itemsPerAdjacentTag, true)
	if err != nil {
		return nil, err
	}
	out := []candidate{}
	for _, tagID := range state.taste.adjacentTags {
		why := reason{kind: reasonExplore, entityID: tagID}
		out = append(out, itemCandidates(itemsByTag[tagID], sourceAdjacentTag, sourcePriors[sourceAdjacentTag], why)...)
	}
	return out, nil
}

// mediaUnion lists the items of kinds together as (kind, id, title,
// createdAt), trashed ones left out.
func mediaUnion(kinds []schema.RecordKind) string {
	selects := make([]string, 0, len(kinds))
	for _, kind := range kinds {
		selects = append(selects, fmt.Sprintf(`SELECT '%s' AS "kind", "id", "title", "createdAt" FROM %s WHERE "deletedAt" IS NULL`,
			kind, db.Quote(tableOf(kind))))
	}
	return strings.Join(selects, " UNION ALL ")
}

// unionRefs lists the items of the snapshot's kinds matching condition (on
// mediaUnion's columns) in order, at most $limit.
func (e *Engine) unionRefs(ctx context.Context, state *snapshot, condition, order string, vars db.Vars) ([]schema.RecordRef, error) {
	rows, err := e.database.Query(ctx,
		`SELECT "kind", "id" FROM (`+mediaUnion(state.kinds)+`) WHERE `+condition+` ORDER BY `+order+` LIMIT $limit`, vars)
	if err != nil {
		return nil, fmt.Errorf("candidate query: %w", err)
	}
	out := make([]schema.RecordRef, 0, len(rows))
	for _, row := range rows {
		out = append(out, rowRef(row, "kind", "id"))
	}
	return out, nil
}

// randomCandidates proposes a few random items once there is a profile to
// escape from; a cold start stays on searches and newest.
func (e *Engine) randomCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	if state.taste.empty() {
		return nil, nil
	}
	refs, err := e.unionRefs(ctx, state, `1`, `RANDOM()`, db.Vars{"limit": randomCandidateLimit})
	if err != nil {
		return nil, err
	}
	return itemCandidates(refs, sourceRandom, sourcePriors[sourceRandom], reason{kind: reasonRandom}), nil
}

// searchCandidates proposes the items whose titles match the recent searches,
// newest first per search.
func (e *Engine) searchCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	texts, err := e.recentSearches(ctx, state.now)
	if err != nil {
		return nil, err
	}
	out := []candidate{}
	for _, text := range texts {
		refs, err := e.unionRefs(ctx, state, `"title" LIKE $pattern`, `"createdAt" DESC`,
			db.Vars{"pattern": "%" + text + "%", "limit": itemsPerSearch})
		if err != nil {
			return nil, err
		}
		out = append(out, itemCandidates(refs, sourceSearch, sourcePriors[sourceSearch], reason{kind: reasonSearch, entityName: text})...)
	}
	return out, nil
}

// recentSearches lists the distinct texts searched within recentSearchWindow,
// newest first.
func (e *Engine) recentSearches(ctx context.Context, now time.Time) ([]string, error) {
	texts, err := e.database.Strings(ctx,
		`SELECT json_extract("data", '`+pathText+`') AS "text" FROM "Event"
		 WHERE "kind" = 'search' AND "at" >= $since AND "deletedAt" IS NULL AND json_extract("data", '`+pathText+`') IS NOT NULL
		 GROUP BY "text" ORDER BY MAX("at") DESC LIMIT $limit`,
		db.Vars{"since": db.FormatTime(now.Add(-recentSearchWindow)), "limit": recentSearchLimit})
	if err != nil {
		return nil, fmt.Errorf("recent searches: %w", err)
	}
	out := []string{}
	for _, text := range texts {
		if trimmed := strings.TrimSpace(text); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out, nil
}

// newestCandidates proposes the most recently stored items: the cold-start
// fallback and the feed's tail.
func (e *Engine) newestCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	refs, err := e.unionRefs(ctx, state, `1`, `"createdAt" DESC`, db.Vars{"limit": newestCandidateLimit})
	if err != nil {
		return nil, err
	}
	return itemCandidates(refs, sourceNewest, sourcePriors[sourceNewest], reason{kind: reasonNewest}), nil
}
