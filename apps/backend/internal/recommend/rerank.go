package recommend

import (
	"slices"
	"strings"
)

// rerankState tracks the page being filled: diversity counters, how many
// slots each source kind took, and the runs of items from the same source and
// the same site.
type rerankState struct {
	slot            int
	performerCounts map[string]int
	studioCounts    map[string]int
	siteCounts      map[string]int
	kindCounts      map[string]int
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

// resetPage clears the per-page diversity and quota counters.
func (state *rerankState) resetPage() {
	state.performerCounts = map[string]int{}
	state.studioCounts = map[string]int{}
	state.siteCounts = map[string]int{}
	state.kindCounts = map[string]int{}
}

// sourceKind groups a candidate source into the kind that shares page slots.
func sourceKind(source string) string {
	switch {
	case strings.HasPrefix(source, "related:"), strings.HasPrefix(source, "affinity:"), source == SourceSubscription:
		return kindPersonal
	case source == SourceTaste:
		return kindTaste
	case source == SourceCatalog, source == SourcePopular:
		return kindCatalog
	case strings.HasPrefix(source, "explore:"):
		return kindExplore
	case source == SourceSearch:
		return kindSearch
	}
	return kindFiller
}

// rerank turns the score-ordered list into the feed order. Each page of
// pageSize slots is shared out by source kind (kindQuotas), the kinds spread
// evenly over the page and ordered by score within; no page holds more than a
// few scenes of one performer, studio or site, and no source or site runs
// more than twice in a row. Slots a kind cannot fill go to the best remaining
// items overall. Items that don't fit a page are deferred, never dropped.
func rerank(scored []scoredItem) []Item {
	queue := slices.Clone(scored)
	state := newRerankState()
	out := make([]Item, 0, len(scored))
	for len(queue) > 0 {
		var picked scoredItem
		picked, queue = takeAt(queue, state.nextIndex(queue))
		state.record(picked)
		out = append(out, picked.Item)
	}
	return out
}

// takeAt removes and returns the item at index.
func takeAt(queue []scoredItem, index int) (scoredItem, []scoredItem) {
	item := queue[index]
	return item, slices.Delete(queue, index, index+1)
}

// nextIndex picks the item for the next slot. A kind still under its quota and
// on pace comes first, then any kind under its quota, then anything but
// search, and last of all search beyond its quota; each stage keeps the
// diversity constraints as far as the queue allows.
func (state *rerankState) nextIndex(queue []scoredItem) int {
	stages := []constraint{
		func(item scoredItem) bool { return state.underQuota(item) && state.onPace(item) },
		state.underQuota,
		func(item scoredItem) bool { return sourceKind(item.Source) != kindSearch },
	}
	for _, stage := range stages {
		if index := state.pickIndex(queue, stage); index >= 0 {
			return index
		}
	}
	if index := state.pickIndex(queue, nil); index >= 0 {
		return index
	}
	return 0
}

// underQuota reports whether the item's kind still has slots on this page.
func (state *rerankState) underQuota(item scoredItem) bool {
	kind := sourceKind(item.Source)
	return state.kindCounts[kind] < kindQuotas[kind]
}

// onPace reports whether taking the item keeps its kind from running ahead of
// an even spread of its quota over the page (paceLookahead slots ahead at
// most), so a strong kind does not take the page's first slots in one block.
func (state *rerankState) onPace(item scoredItem) bool {
	kind := sourceKind(item.Source)
	slotsSoFar := state.slot%pageSize + 1 + paceLookahead
	allowed := (kindQuotas[kind]*slotsSoFar + pageSize - 1) / pageSize
	return state.kindCounts[kind] < allowed
}

// pickIndex returns the index of the best-scored item that is admitted by
// stage (nil admits all) and keeps every constraint; -1 when none does.
// Constraints are dropped weakest first: the source run, then the site run,
// then the site cap; the performer/studio caps go last.
func (state *rerankState) pickIndex(queue []scoredItem, stage constraint) int {
	tiers := [][]constraint{
		{state.withinCaps, state.withinSiteCap, state.continuesSiteRun, state.continuesSourceRun},
		{state.withinCaps, state.withinSiteCap, state.continuesSiteRun},
		{state.withinCaps, state.withinSiteCap},
		{state.withinCaps},
	}
	for _, constraints := range tiers {
		if stage != nil {
			constraints = append([]constraint{stage}, constraints...)
		}
		if index := firstSatisfying(queue, constraints); index >= 0 {
			return index
		}
	}
	return -1
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
	state.kindCounts[sourceKind(item.Source)]++
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
