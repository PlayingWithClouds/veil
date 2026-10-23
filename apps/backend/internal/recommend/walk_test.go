package recommend

import (
	"context"
	"testing"
	"time"
)

// insertRelated links scene to related with the given source and rank.
func insertRelated(t *testing.T, engine *Engine, scene, related, source string, rank int) {
	t.Helper()
	insertRow(t, engine.database, "scene_related", "scene_related:"+scene+">"+related, map[string]any{
		"scene": scene, "related": related, "source": source, "rank": rank,
	})
}

// walkCandidatesBySceneID runs the walk and keys its candidates by scene.
func walkCandidatesBySceneID(t *testing.T, engine *Engine) map[string]candidate {
	t.Helper()
	ctx := context.Background()
	state, err := engine.takeSnapshot(ctx, nil)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	found, err := engine.walkCandidates(ctx, state)
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	out := map[string]candidate{}
	for _, proposed := range found {
		out[proposed.sceneID] = proposed
	}
	return out
}

func TestWalkReachesScenesBeyondTheDirectRelatedList(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertSessionEnder(t, engine, "scene:seed", 2*time.Hour, map[string]any{"performers": []string{"performer:seed"}})
	insertRow(t, database, "performer", "performer:seed", map[string]any{"name": "Seed Performer"})
	for _, id := range []string{"scene:direct", "scene:second", "scene:third", "scene:fourth", "scene:lister"} {
		insertScene(t, database, id, map[string]any{})
	}
	insertRelated(t, engine, "scene:seed", "scene:direct", "site", 0)
	insertRelated(t, engine, "scene:direct", "scene:second", "site", 0)
	insertRelated(t, engine, "scene:second", "scene:third", "site", 0)
	insertRelated(t, engine, "scene:third", "scene:fourth", "site", 0)
	// A scene whose own list names the seed: reached against the edge's direction.
	insertRelated(t, engine, "scene:lister", "scene:seed", "site", 0)

	found := walkCandidatesBySceneID(t, engine)
	for _, id := range []string{"scene:second", "scene:third", "scene:lister"} {
		if _, reached := found[id]; !reached {
			t.Errorf("%s must be reached by the walk, got %v", id, found)
		}
	}
	for _, id := range []string{"scene:seed", "scene:direct", "scene:fourth"} {
		if _, reached := found[id]; reached {
			t.Errorf("%s must not be proposed (seed, direct list, or beyond %d hops)", id, walkHops)
		}
	}
	second, third := found["scene:second"], found["scene:third"]
	if second.prior <= third.prior {
		t.Errorf("a nearer scene carries more mass: second %v, third %v", second.prior, third.prior)
	}
	if second.source != SourceRelatedWalk || second.reason.Kind != ReasonRelated || second.reason.EntityID != "scene:seed" || second.reason.EntityName != "Seed Performer" {
		t.Errorf("second = %+v", second)
	}
	if second.prior >= sourcePriors[SourceRelatedSite] {
		t.Errorf("walk prior %v must stay below related:site", second.prior)
	}
}

func TestWalkLinksScenesThroughSharedPerformersAndRareTags(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertSessionEnder(t, engine, "scene:seed", 2*time.Hour, map[string]any{
		"performers": []string{"performer:shared"}, "tags": []string{"tag:niche"},
	})
	insertScene(t, database, "scene:same-performer", map[string]any{"performers": []string{"performer:shared"}})
	insertScene(t, database, "scene:same-tag", map[string]any{"tags": []string{"tag:niche"}})
	insertScene(t, database, "scene:unrelated", map[string]any{"performers": []string{"performer:other"}, "tags": []string{"tag:other"}})

	found := walkCandidatesBySceneID(t, engine)
	if _, reached := found["scene:same-performer"]; !reached {
		t.Errorf("a scene by the seed's performer must be reached: %v", found)
	}
	if _, reached := found["scene:same-tag"]; !reached {
		t.Errorf("a scene with the seed's rare tag must be reached: %v", found)
	}
	if _, reached := found["scene:unrelated"]; reached {
		t.Error("an unlinked scene must not be reached")
	}
}

func TestWalkIgnoresGenericTags(t *testing.T) {
	rarity := &tagRarity{taggedScenes: 1000, sceneCounts: map[string]int{"tag:generic": 900, "tag:niche": 10}}
	got := rarestTags([]string{"tag:generic", "tag:niche", "tag:unseen"}, rarity)
	if len(got) != 2 || got[0] == "tag:generic" || got[1] == "tag:generic" {
		t.Errorf("rarest tags = %v, want the generic tag left out", got)
	}
}

func TestSpreadMassSplitsByEdgeWeightAndSkipsVisited(t *testing.T) {
	frontier := map[string]*walkNode{
		"scene:a": {mass: 1, origin: "scene:a"},
		"scene:b": {mass: 0.5, origin: "scene:b"},
	}
	edges := []walkEdge{
		{from: "scene:a", to: "scene:x", weight: 3},
		{from: "scene:a", to: "scene:y", weight: 1},
		{from: "scene:a", to: "scene:seen", weight: 1},
		{from: "scene:b", to: "scene:x", weight: 1},
	}
	layer := spreadMass(frontier, edges, map[string]bool{"scene:seen": true})
	if _, reached := layer["scene:seen"]; reached {
		t.Error("a visited scene gets nothing")
	}
	assertClose(t, "x mass", layer["scene:x"].mass, walkDamping*(1*3.0/5+0.5))
	assertClose(t, "y mass", layer["scene:y"].mass, walkDamping*1.0/5)
	if layer["scene:x"].origin != "scene:a" {
		t.Errorf("x's strongest share (0.36 against 0.3) came from a's seed, got %q", layer["scene:x"].origin)
	}
}

func TestStrongestNodesCapsTheFrontier(t *testing.T) {
	nodes := map[string]*walkNode{"scene:a": {mass: 3}, "scene:b": {mass: 2}, "scene:c": {mass: 2}, "scene:d": {mass: 1}}
	kept := strongestNodes(nodes, 2)
	if len(kept) != 2 || kept["scene:a"] == nil || kept["scene:b"] == nil {
		t.Errorf("kept = %v, want the two strongest with ties by id", kept)
	}
}

func TestFeedServesWalkedScenesAsRelatedWalk(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	insertSessionEnder(t, engine, "scene:seed", 2*time.Hour, map[string]any{})
	insertScene(t, database, "scene:direct", map[string]any{})
	insertScene(t, database, "scene:second", map[string]any{})
	insertRelated(t, engine, "scene:seed", "scene:direct", "site", 0)
	insertRelated(t, engine, "scene:direct", "scene:second", "site", 0)

	item := itemBySceneID(feed(t, engine, nil), "scene:second")
	if item == nil || item.Source != SourceRelatedWalk || item.Reason.Kind != ReasonRelated {
		t.Errorf("second-hop scene = %+v, want related:walk", item)
	}
}
