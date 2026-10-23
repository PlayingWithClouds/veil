package recommend

import (
	"context"
	"fmt"
	"sort"

	"github.com/playingwithclouds/veil/internal/db"
)

// SourceRelatedWalk proposes scenes found by walking the related graph a few
// hops from the seeds (mixed with shared-performer and shared-rare-tag links).
const SourceRelatedWalk = "related:walk"

// SourceExploreAttribute proposes scenes of performers with no signal of their
// own whose looks match the performers the user likes. It is exploration
// (kindExplore): nothing the user did points at the performer, only her
// resemblance to someone they liked.
const SourceExploreAttribute = "explore:attribute"

// walkNode is mass a scene collected during one hop, and the seed the
// strongest share of it came from.
type walkNode struct {
	mass       float64
	origin     string
	originPass float64
}

// walkEdge is a link along which a scene passes mass on. forward marks a
// scene_related edge walked in its own direction.
type walkEdge struct {
	from    string
	to      string
	weight  float64
	forward bool
}

// walkCandidates proposes scenes up to walkHops hops from the seeds: a short
// personalized random walk. Each hop a scene passes walkDamping of its mass on
// along its edges, split in proportion to the edge weights, and a scene is
// counted where it is first reached. What the direct related lists already
// propose (see relatedCandidates) is left out; the walk adds what is further
// out, with the mass it collected as its strength. Each hop expands at most
// walkFrontierLimit scenes with a handful of queries.
func (e *Engine) walkCandidates(ctx context.Context, state *snapshot) ([]candidate, error) {
	seeds := state.profile.Seeds
	if len(seeds) == 0 {
		return nil, nil
	}
	frontier := map[string]*walkNode{}
	visited := map[string]bool{}
	credits := map[string]string{}
	for _, seed := range seeds {
		frontier[seed.SceneID] = &walkNode{mass: seed.Weight / seeds[0].Weight, origin: seed.SceneID}
		visited[seed.SceneID] = true
		credits[seed.SceneID] = seed.Credit
	}
	reached := map[string]*walkNode{}
	direct := map[string]bool{}
	for hop := 1; hop <= walkHops && len(frontier) > 0; hop++ {
		edges, err := e.walkEdges(ctx, frontier, state.profile.rarity)
		if err != nil {
			return nil, err
		}
		if hop == 1 {
			direct = forwardTargets(edges)
		}
		layer := spreadMass(frontier, edges, visited)
		for sceneID, node := range layer {
			visited[sceneID] = true
			if !direct[sceneID] {
				reached[sceneID] = node
			}
		}
		frontier = strongestNodes(layer, walkFrontierLimit)
	}
	return walkResults(reached, credits), nil
}

// forwardTargets collects the scenes the forward related edges point at.
func forwardTargets(edges []walkEdge) map[string]bool {
	out := map[string]bool{}
	for _, edge := range edges {
		if edge.forward {
			out[edge.to] = true
		}
	}
	return out
}

// spreadMass passes each frontier scene's damped mass along its edges, split
// by edge weight, and returns the mass every not yet visited scene received.
func spreadMass(frontier map[string]*walkNode, edges []walkEdge, visited map[string]bool) map[string]*walkNode {
	totals := map[string]float64{}
	for _, edge := range edges {
		totals[edge.from] += edge.weight
	}
	layer := map[string]*walkNode{}
	for _, edge := range edges {
		if visited[edge.to] {
			continue
		}
		source := frontier[edge.from]
		passed := source.mass * walkDamping * edge.weight / totals[edge.from]
		node := layer[edge.to]
		if node == nil {
			node = &walkNode{}
			layer[edge.to] = node
		}
		node.mass += passed
		if passed > node.originPass {
			node.originPass = passed
			node.origin = source.origin
		}
	}
	return layer
}

// strongestNodes keeps the limit nodes with the most mass (ties by id).
func strongestNodes(nodes map[string]*walkNode, limit int) map[string]*walkNode {
	ids := sortedByMass(nodes)
	if len(ids) > limit {
		ids = ids[:limit]
	}
	out := make(map[string]*walkNode, len(ids))
	for _, sceneID := range ids {
		out[sceneID] = nodes[sceneID]
	}
	return out
}

// sortedByMass lists the node ids, most mass first, ties by id.
func sortedByMass(nodes map[string]*walkNode) []string {
	ids := make([]string, 0, len(nodes))
	for sceneID := range nodes {
		ids = append(ids, sceneID)
	}
	sort.Slice(ids, func(left, right int) bool {
		if nodes[ids[left]].mass != nodes[ids[right]].mass {
			return nodes[ids[left]].mass > nodes[ids[right]].mass
		}
		return ids[left] < ids[right]
	})
	return ids
}

// walkResults turns the scenes the walk reached into candidates, the one with
// the most mass first. The prior scales with the mass the way a seed's
// strength scales its related scenes' prior.
func walkResults(reached map[string]*walkNode, credits map[string]string) []candidate {
	ids := sortedByMass(reached)
	if len(ids) > walkCandidateLimit {
		ids = ids[:walkCandidateLimit]
	}
	out := make([]candidate, 0, len(ids))
	for rank, sceneID := range ids {
		node := reached[sceneID]
		massFactor := 1 - seedWeightInfluence + seedWeightInfluence*node.mass/reached[ids[0]].mass
		out = append(out, candidate{
			sceneID: sceneID,
			source:  SourceRelatedWalk,
			prior:   rankedPrior(sourcePriors[SourceRelatedWalk]*massFactor, rank),
			reason:  Reason{Kind: ReasonRelated, EntityID: node.origin, EntityName: credits[node.origin]},
		})
	}
	return out
}

// walkEdges collects the edges leaving the frontier scenes: their related
// lists (own direction at full weight, the scenes listing them at
// walkReverseWeight), and shared performers and shared rare tags.
func (e *Engine) walkEdges(ctx context.Context, frontier map[string]*walkNode, rarity *tagRarity) ([]walkEdge, error) {
	sceneIDs := make([]string, 0, len(frontier))
	for sceneID := range frontier {
		sceneIDs = append(sceneIDs, sceneID)
	}
	sort.Strings(sceneIDs)
	related, err := e.relatedEdges(ctx, sceneIDs)
	if err != nil {
		return nil, err
	}
	shared, err := e.sharedEntityEdges(ctx, sceneIDs, rarity)
	if err != nil {
		return nil, err
	}
	return append(related, shared...), nil
}

// relatedEdges reads the related lists of the scenes and, reversed, the
// strongest lists naming them.
func (e *Engine) relatedEdges(ctx context.Context, sceneIDs []string) ([]walkEdge, error) {
	forward, err := e.database.Query(ctx,
		`SELECT scene AS anchor, related AS other, source, rank FROM scene_related
		 WHERE scene IN (SELECT value FROM json_each($ids)) AND rank < $perSource`,
		db.Vars{"ids": sceneIDs, "perSource": relatedPerSeedSource})
	if err != nil {
		return nil, fmt.Errorf("walk related edges: %w", err)
	}
	reverse, err := e.database.Query(ctx,
		`SELECT related AS anchor, scene AS other, source, rank FROM scene_related
		 WHERE related IN (SELECT value FROM json_each($ids)) AND rank < $perSource
		 ORDER BY rank LIMIT $limit`,
		db.Vars{"ids": sceneIDs, "perSource": relatedPerSeedSource, "limit": len(sceneIDs) * relatedPerSeedSource})
	if err != nil {
		return nil, fmt.Errorf("walk reverse related edges: %w", err)
	}
	edges := make([]walkEdge, 0, len(forward)+len(reverse))
	for _, row := range forward {
		edges = append(edges, walkEdge{from: rowString(row, "anchor"), to: rowString(row, "other"), weight: relatedEdgeWeight(row), forward: true})
	}
	for _, row := range reverse {
		edges = append(edges, walkEdge{from: rowString(row, "anchor"), to: rowString(row, "other"), weight: walkReverseWeight * relatedEdgeWeight(row)})
	}
	return edges, nil
}

// relatedEdgeWeight is a related edge's source prior lowered by its rank.
func relatedEdgeWeight(row db.Row) float64 {
	return rankedPrior(sourcePriors["related:"+rowString(row, "source")], rowInt(row, "rank"))
}

// sharedEntityEdges links each scene to the newest scenes of its performers
// and of its rarest tags (rarity at least walkTagMinRarity).
func (e *Engine) sharedEntityEdges(ctx context.Context, sceneIDs []string, rarity *tagRarity) ([]walkEdge, error) {
	rows, err := e.database.Query(ctx,
		`SELECT id, tags, performers FROM scene WHERE id IN (SELECT value FROM json_each($ids))`,
		db.Vars{"ids": sceneIDs})
	if err != nil {
		return nil, fmt.Errorf("walk entities: %w", err)
	}
	performersOf := map[string][]string{}
	tagsOf := map[string][]string{}
	for _, row := range rows {
		sceneID := rowString(row, "id")
		performersOf[sceneID] = rowStrings(row, "performers")
		tagsOf[sceneID] = rarestTags(rowStrings(row, "tags"), rarity)
	}
	edges, err := e.entityEdges(ctx, "performers", performersOf, walkPerformerEdgeWeight)
	if err != nil {
		return nil, err
	}
	tagEdges, err := e.entityEdges(ctx, "tags", tagsOf, walkTagEdgeWeight)
	if err != nil {
		return nil, err
	}
	return append(edges, tagEdges...), nil
}

// rarestTags keeps the (at most walkEntitiesPerScene) tags rare enough to
// link scenes, rarest first.
func rarestTags(tags []string, rarity *tagRarity) []string {
	rare := []string{}
	for _, tag := range tags {
		if rarity.weight(tag) >= walkTagMinRarity {
			rare = append(rare, tag)
		}
	}
	sort.SliceStable(rare, func(left, right int) bool { return rarity.weight(rare[left]) > rarity.weight(rare[right]) })
	if len(rare) > walkEntitiesPerScene {
		rare = rare[:walkEntitiesPerScene]
	}
	return rare
}

// entityEdges links every scene to the newest scenes of its entities
// (performers or tags, per entitiesOf) with the given base weight.
func (e *Engine) entityEdges(ctx context.Context, column string, entitiesOf map[string][]string, baseWeight float64) ([]walkEdge, error) {
	distinct := map[string]bool{}
	for _, entities := range entitiesOf {
		for _, entity := range entities {
			distinct[entity] = true
		}
	}
	if len(distinct) == 0 {
		return nil, nil
	}
	scenesOf, err := e.newestScenesByEntity(ctx, column, setKeys(distinct), walkScenesPerEntity)
	if err != nil {
		return nil, err
	}
	edges := []walkEdge{}
	for sceneID, entities := range entitiesOf {
		for _, entity := range entities {
			for position, other := range scenesOf[entity] {
				if other != sceneID {
					edges = append(edges, walkEdge{from: sceneID, to: other, weight: rankedPrior(baseWeight, position)})
				}
			}
		}
	}
	return edges, nil
}
