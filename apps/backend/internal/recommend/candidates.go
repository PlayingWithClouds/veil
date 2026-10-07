package recommend

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// candidate is a scene some source proposes, before scoring.
type candidate struct {
	sceneID string
	source  string
	// prior is the source's trust in this scene, in [0, 1].
	prior float64
	// sourceCount is how many sources proposed the scene.
	sourceCount int
	reason      Reason
}

// candidatePool merges candidates by scene, keeping the strongest source's
// prior and reason and counting how many sources agreed.
type candidatePool struct {
	byScene map[string]*candidate
	order   []string
}

// newCandidatePool returns an empty pool.
func newCandidatePool() *candidatePool {
	return &candidatePool{byScene: map[string]*candidate{}}
}

// add merges one candidate into the pool.
func (pool *candidatePool) add(proposed candidate) {
	existing, found := pool.byScene[proposed.sceneID]
	if !found {
		proposed.sourceCount = 1
		pool.byScene[proposed.sceneID] = &proposed
		pool.order = append(pool.order, proposed.sceneID)
		return
	}
	if existing.source != proposed.source {
		existing.sourceCount++
	}
	if proposed.prior > existing.prior {
		existing.prior = proposed.prior
		existing.source = proposed.source
		existing.reason = proposed.reason
	}
}

// list returns the pooled candidates in first-seen order.
func (pool *candidatePool) list() []*candidate {
	out := make([]*candidate, 0, len(pool.order))
	for _, sceneID := range pool.order {
		out = append(out, pool.byScene[sceneID])
	}
	return out
}

// candidateGenerator proposes candidates from one source.
type candidateGenerator func(ctx context.Context, state *snapshot) ([]candidate, error)

// generateCandidates runs every source and pools the results.
func (e *Engine) generateCandidates(ctx context.Context, state *snapshot) (*candidatePool, error) {
	generators := []candidateGenerator{
		e.relatedCandidates,
		e.subscriptionCandidates,
		e.affinityCandidates,
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
			pool.add(proposed)
		}
	}
	return pool, nil
}

// rankedPrior lowers a source prior by the item's position within the source.
func rankedPrior(prior float64, rank int) float64 {
	return prior / (1 + float64(rank)*rankDecay)
}

// sceneCandidates turns a ranked id list from one source into candidates.
func sceneCandidates(sceneIDs []string, source string, prior float64, reason Reason) []candidate {
	out := make([]candidate, 0, len(sceneIDs))
	for rank, sceneID := range sceneIDs {
		out = append(out, candidate{sceneID: sceneID, source: source, prior: rankedPrior(prior, rank), reason: reason})
	}
	return out
}

// sceneIDs runs a single-column id query.
func (e *Engine) sceneIDs(ctx context.Context, query string, vars db.Vars) ([]string, error) {
	ids, err := e.database.Strings(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("candidate query: %w", err)
	}
	return ids, nil
}

// relatedCandidates proposes the related scenes of the seeds, the site's own
// related list first. A stronger seed lends its related scenes a higher prior.
func (e *Engine) relatedCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	seeds := state.profile.Seeds
	if len(seeds) == 0 {
		return nil, nil
	}
	seedsByID := make(map[string]Seed, len(seeds))
	seedIDs := make([]string, 0, len(seeds))
	for _, seed := range seeds {
		seedsByID[seed.SceneID] = seed
		seedIDs = append(seedIDs, seed.SceneID)
	}
	rows, err := e.database.Query(ctx,
		`SELECT scene, related, source, rank FROM scene_related
		 WHERE scene IN (SELECT value FROM json_each($seeds)) AND rank < $perSource
		 ORDER BY rank`,
		db.Vars{"seeds": seedIDs, "perSource": relatedPerSeedSource})
	if err != nil {
		return nil, fmt.Errorf("related candidates: %w", err)
	}
	out := make([]candidate, 0, len(rows))
	for _, row := range rows {
		out = append(out, relatedCandidate(row, seedsByID[rowString(row, "scene")], seeds[0].Weight))
	}
	return out, nil
}

// relatedCandidate builds one related-scene candidate; strongestSeed scales
// the seed's influence.
func relatedCandidate(row db.Row, seed Seed, strongestSeed float64) candidate {
	source := "related:" + rowString(row, "source")
	seedFactor := 1 - seedWeightInfluence + seedWeightInfluence*seed.Weight/strongestSeed
	return candidate{
		sceneID: rowString(row, "related"),
		source:  source,
		prior:   rankedPrior(sourcePriors[source], rowInt(row, "rank")) * seedFactor,
		reason:  Reason{Kind: ReasonRelated, EntityID: seed.SceneID, EntityName: seed.Title},
	}
}

// subscriptionCandidates proposes scenes that search subscriptions found since
// they were last marked seen, newest first.
func (e *Engine) subscriptionCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	rows, err := e.database.Query(ctx,
		`SELECT item.media, subscription.id AS subscription, subscription.query
		 FROM search_subscription_item AS item
		 JOIN search_subscription AS subscription ON subscription.id = item.subscription
		 WHERE subscription.enabled = 1 AND item.media LIKE 'scene:%'
		   AND item.first_seen_at > coalesce(subscription.seen_at, '')
		 ORDER BY item.first_seen_at DESC LIMIT $limit`,
		db.Vars{"limit": subscriptionCandidateLimit})
	if err != nil {
		return nil, fmt.Errorf("subscription candidates: %w", err)
	}
	out := make([]candidate, 0, len(rows))
	for rank, row := range rows {
		out = append(out, candidate{
			sceneID: rowString(row, "media"),
			source:  SourceSubscription,
			prior:   rankedPrior(sourcePriors[SourceSubscription], rank),
			reason:  Reason{Kind: ReasonSubscription, EntityID: rowString(row, "subscription"), EntityName: rowString(row, "query")},
		})
	}
	return out, nil
}

// Newest scenes carrying one entity ($entity), by entity kind.
var entitySceneQueries = map[string]string{
	ReasonPerformer: `SELECT id FROM scene WHERE EXISTS (SELECT 1 FROM json_each(scene.performers) WHERE value = $entity)
		ORDER BY created_at DESC LIMIT $limit`,
	ReasonStudio: `SELECT id FROM scene WHERE studio = $entity ORDER BY created_at DESC LIMIT $limit`,
	ReasonTag: `SELECT id FROM scene WHERE EXISTS (SELECT 1 FROM json_each(scene.tags) WHERE value = $entity)
		ORDER BY created_at DESC LIMIT $limit`,
}

// affinityCandidates proposes scenes of the top-affinity performers, studios
// and tags, each entity's prior scaled by its affinity.
func (e *Engine) affinityCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	dimensions := []struct {
		kind       string
		source     string
		affinities map[string]float64
		limit      int
	}{
		{ReasonPerformer, SourcePerformer, state.profile.Performers, topPerformerLimit},
		{ReasonStudio, SourceStudio, state.profile.Studios, topStudioLimit},
		{ReasonTag, SourceTag, state.profile.Tags, topTagLimit},
	}
	out := []candidate{}
	for _, dimension := range dimensions {
		found, err := e.entityCandidates(ctx, dimension.kind, dimension.source, dimension.affinities, dimension.limit)
		if err != nil {
			return nil, err
		}
		out = append(out, found...)
	}
	return out, nil
}

// entityCandidates proposes the newest scenes of the top entities of one
// dimension.
func (e *Engine) entityCandidates(ctx context.Context, kind, source string, affinities map[string]float64, limit int) ([]candidate, error) {
	out := []candidate{}
	for _, entityID := range TopEntities(affinities, limit) {
		ids, err := e.sceneIDs(ctx, entitySceneQueries[kind], db.Vars{"entity": entityID, "limit": scenesPerAffinityEntity})
		if err != nil {
			return nil, err
		}
		reason := Reason{Kind: kind, EntityID: entityID}
		out = append(out, sceneCandidates(ids, source, sourcePriors[source]*affinities[entityID], reason)...)
	}
	return out, nil
}

// adjacentTagCandidates proposes a random sample of scenes from tags adjacent
// to the top tags: the exploration share that widens the profile.
func (e *Engine) adjacentTagCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	out := []candidate{}
	for _, tagID := range state.profile.AdjacentTags {
		ids, err := e.sceneIDs(ctx,
			`SELECT id FROM scene WHERE EXISTS (SELECT 1 FROM json_each(scene.tags) WHERE value = $tag)
			 ORDER BY random() LIMIT $limit`,
			db.Vars{"tag": tagID, "limit": scenesPerAdjacentTag})
		if err != nil {
			return nil, err
		}
		reason := Reason{Kind: ReasonExplore, EntityID: tagID}
		out = append(out, sceneCandidates(ids, SourceAdjacentTag, sourcePriors[SourceAdjacentTag], reason)...)
	}
	return out, nil
}

// randomCandidates proposes a few random scenes once there is a profile to
// escape from; a cold start stays on searches and newest.
func (e *Engine) randomCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	if state.profile.Empty() {
		return nil, nil
	}
	ids, err := e.sceneIDs(ctx, `SELECT id FROM scene ORDER BY random() LIMIT $limit`, db.Vars{"limit": randomCandidateLimit})
	if err != nil {
		return nil, err
	}
	return sceneCandidates(ids, SourceRandom, sourcePriors[SourceRandom], Reason{Kind: ReasonRandom}), nil
}

// searchCandidates proposes the unvisited stubs recent searches stored: the
// scenes first stored around the time each search ran.
func (e *Engine) searchCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	rows, err := e.database.Query(ctx,
		`SELECT query, max(created_at) AS at FROM search_history
		 WHERE created_at >= $since GROUP BY normalized_query ORDER BY at DESC LIMIT $limit`,
		db.Vars{"since": db.FormatTime(state.now.Add(-recentSearchWindow)), "limit": recentSearchLimit})
	if err != nil {
		return nil, fmt.Errorf("recent searches: %w", err)
	}
	out := []candidate{}
	for _, row := range rows {
		ids, err := e.searchResultIDs(ctx, rowTime(row, "at"))
		if err != nil {
			return nil, err
		}
		reason := Reason{Kind: ReasonSearch, EntityName: rowString(row, "query")}
		out = append(out, sceneCandidates(ids, SourceSearch, sourcePriors[SourceSearch], reason)...)
	}
	return out, nil
}

// searchResultIDs lists the unvisited scenes first stored around a search.
func (e *Engine) searchResultIDs(ctx context.Context, searchedAt time.Time) ([]string, error) {
	return e.sceneIDs(ctx,
		`SELECT id FROM scene
		 WHERE created_at >= $from AND created_at <= $to AND detail_fetched_at IS NULL
		 ORDER BY created_at LIMIT $limit`,
		db.Vars{
			"from":  db.FormatTime(searchedAt.Add(-searchResultsBefore)),
			"to":    db.FormatTime(searchedAt.Add(searchResultsAfter)),
			"limit": scenesPerSearch,
		})
}

// newestCandidates proposes the most recently stored scenes: the cold-start
// fallback and the feed's tail.
func (e *Engine) newestCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	ids, err := e.sceneIDs(ctx, `SELECT id FROM scene ORDER BY created_at DESC LIMIT $limit`, db.Vars{"limit": newestCandidateLimit})
	if err != nil {
		return nil, err
	}
	return sceneCandidates(ids, SourceNewest, sourcePriors[SourceNewest], Reason{Kind: ReasonNewest}), nil
}
