package recommend

import "slices"

// rerankState tracks the page being filled: diversity counters and the runs of
// items from the same source and the same site.
type rerankState struct {
	slot            int
	performerCounts map[string]int
	studioCounts    map[string]int
	siteCounts      map[string]int
	lastSource      string
	sameSourceRun   int
	lastSite        string
	sameSiteRun     int
}

// constraint is one re-ranking rule an item may break.
type constraint func(item scoredItem) bool

// newRerankState starts at the first slot of the first page.
func newRerankState() *rerankState {
	state := &rerankState{}
	state.resetPage()
	return state
}

// resetPage clears the per-page diversity counters.
func (state *rerankState) resetPage() {
	state.performerCounts = map[string]int{}
	state.studioCounts = map[string]int{}
	state.siteCounts = map[string]int{}
}

// rerank turns the score-ordered list into the feed order: every seventh slot
// goes to exploration, no page holds more than a few scenes of one performer,
// studio or site, and no source or site runs more than twice in a row. Items
// that don't fit a page are deferred, never dropped.
func rerank(scored []scoredItem) []Item {
	main, exploration := splitExploration(scored)
	state := newRerankState()
	out := make([]Item, 0, len(scored))
	for len(main)+len(exploration) > 0 {
		var picked scoredItem
		if len(main) == 0 || (state.isExplorationSlot() && len(exploration) > 0) {
			picked, exploration = takeAt(exploration, state.pickIndex(exploration))
		} else {
			picked, main = takeAt(main, state.pickIndex(main))
		}
		state.record(picked)
		out = append(out, picked.Item)
	}
	return out
}

// splitExploration separates exploration items, keeping each list's order.
func splitExploration(scored []scoredItem) (main, exploration []scoredItem) {
	for _, item := range scored {
		if isExploration(item.Source) {
			exploration = append(exploration, item)
			continue
		}
		main = append(main, item)
	}
	return main, exploration
}

// takeAt removes and returns the item at index.
func takeAt(queue []scoredItem, index int) (scoredItem, []scoredItem) {
	item := queue[index]
	return item, slices.Delete(queue, index, index+1)
}

// isExplorationSlot reports whether the next slot is reserved for exploration.
func (state *rerankState) isExplorationSlot() bool {
	return (state.slot+1)%explorationEvery == 0
}

// pickIndex returns the best-scored item that keeps every constraint. When
// none does (e.g. the feed holds a single site), constraints are dropped
// weakest first: the source run, then the site run, then the site cap; the
// performer/studio caps go last.
func (state *rerankState) pickIndex(queue []scoredItem) int {
	tiers := [][]constraint{
		{state.withinCaps, state.withinSiteCap, state.continuesSiteRun, state.continuesSourceRun},
		{state.withinCaps, state.withinSiteCap, state.continuesSiteRun},
		{state.withinCaps, state.withinSiteCap},
		{state.withinCaps},
	}
	for _, constraints := range tiers {
		if index := firstSatisfying(queue, constraints); index >= 0 {
			return index
		}
	}
	return 0
}

// firstSatisfying returns the index of the first item keeping every
// constraint, or -1.
func firstSatisfying(queue []scoredItem, constraints []constraint) int {
	for index, item := range queue {
		if satisfiesAll(item, constraints) {
			return index
		}
	}
	return -1
}

// satisfiesAll reports whether item keeps every constraint.
func satisfiesAll(item scoredItem, constraints []constraint) bool {
	for _, keeps := range constraints {
		if !keeps(item) {
			return false
		}
	}
	return true
}

// withinCaps reports whether the page still has room for the item's
// performers and studio.
func (state *rerankState) withinCaps(item scoredItem) bool {
	for _, performer := range item.features.performers {
		if state.performerCounts[performer] >= maxPerPerformerPerPage {
			return false
		}
	}
	studio := item.features.studio
	return studio == "" || state.studioCounts[studio] < maxPerStudioPerPage
}

// withinSiteCap reports whether the page still has room for the item's site.
func (state *rerankState) withinSiteCap(item scoredItem) bool {
	site := item.features.site
	return site == "" || state.siteCounts[site] < maxPerSitePerPage
}

// continuesSiteRun reports whether the item may follow the current run of
// same-site items.
func (state *rerankState) continuesSiteRun(item scoredItem) bool {
	site := item.features.site
	return site == "" || site != state.lastSite || state.sameSiteRun < maxSameSiteRun
}

// continuesSourceRun reports whether the item may follow the current run of
// same-source items.
func (state *rerankState) continuesSourceRun(item scoredItem) bool {
	return item.Source != state.lastSource || state.sameSourceRun < maxSameSourceRun
}

// record fills the slot with item and advances, starting a fresh page when
// this one is full.
func (state *rerankState) record(item scoredItem) {
	for _, performer := range item.features.performers {
		state.performerCounts[performer]++
	}
	if item.features.studio != "" {
		state.studioCounts[item.features.studio]++
	}
	if item.features.site != "" {
		state.siteCounts[item.features.site]++
	}
	state.recordRuns(item)
	state.slot++
	if state.slot%pageSize == 0 {
		state.resetPage()
	}
}

// recordRuns extends or restarts the same-source and same-site runs.
func (state *rerankState) recordRuns(item scoredItem) {
	if item.Source == state.lastSource {
		state.sameSourceRun++
	} else {
		state.lastSource = item.Source
		state.sameSourceRun = 1
	}
	if item.features.site == state.lastSite {
		state.sameSiteRun++
	} else {
		state.lastSite = item.features.site
		state.sameSiteRun = 1
	}
}
