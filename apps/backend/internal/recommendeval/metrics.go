package recommendeval

// Ranking cut-offs the hit rate and recall are measured at, and the depth of
// the ranking requested (for MRR).
const (
	SmallCutoff = 24
	LargeCutoff = 96
	listDepth   = 300
)

// listScore is how one ranking did against one target set.
type listScore struct {
	// hitsSmall and hitsLarge are the targets inside the first SmallCutoff
	// and LargeCutoff positions.
	hitsSmall, hitsLarge int
	// reciprocalRank is 1 over the position of the first target, 0 when none
	// is ranked.
	reciprocalRank float64
}

// scoreList counts the reachable targets among the ranked scenes.
func scoreList(ranking []rankedScene, targets map[string]bool) listScore {
	var score listScore
	for position, entry := range ranking {
		if !targets[entry.sceneID] {
			continue
		}
		if score.reciprocalRank == 0 {
			score.reciprocalRank = 1 / float64(position+1)
		}
		if position < SmallCutoff {
			score.hitsSmall++
		}
		if position < LargeCutoff {
			score.hitsLarge++
		}
	}
	return score
}

// coverage is the spread of a ranking's first SmallCutoff scenes.
type coverage struct {
	sites       int
	tags        int
	sourceCount map[string]int
	size        int
}

// measureCoverage counts the distinct sites and tags and the sources of the
// first SmallCutoff scenes.
func measureCoverage(ranking []rankedScene, traits map[string]sceneTraits) coverage {
	top := ranking
	if len(top) > SmallCutoff {
		top = top[:SmallCutoff]
	}
	sites, tags := map[string]bool{}, map[string]bool{}
	measured := coverage{sourceCount: map[string]int{}, size: len(top)}
	for _, entry := range top {
		sceneTraits := traits[entry.sceneID]
		if sceneTraits.site != "" {
			sites[sceneTraits.site] = true
		}
		for _, tag := range sceneTraits.tags {
			tags[tag] = true
		}
		if entry.source != "" {
			measured.sourceCount[entry.source]++
		}
	}
	measured.sites, measured.tags = len(sites), len(tags)
	return measured
}

// topSceneIDs lists the scene ids of the first SmallCutoff positions.
func topSceneIDs(ranking []rankedScene) []string {
	count := min(SmallCutoff, len(ranking))
	ids := make([]string, 0, count)
	for _, entry := range ranking[:count] {
		ids = append(ids, entry.sceneID)
	}
	return ids
}
