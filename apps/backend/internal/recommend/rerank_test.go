package recommend

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// syntheticItem builds a scored item with one performer.
func syntheticItem(sceneID, source, performer string, score float64) scoredItem {
	return scoredItem{
		Item:     Item{SceneID: sceneID, Source: source, Score: score},
		features: &sceneFeatures{id: sceneID, performers: []string{performer}},
	}
}

func TestRerankCapsPerformersPerPage(t *testing.T) {
	scored := []scoredItem{}
	for index := range 6 {
		scored = append(scored, syntheticItem(fmt.Sprintf("scene:star%d", index), SourcePerformer, "performer:star", float64(10-index)))
	}
	for index := range 6 {
		scored = append(scored, syntheticItem(fmt.Sprintf("scene:other%d", index), SourceTag, fmt.Sprintf("performer:other%d", index), float64(4-index)))
	}

	ranked := rerank(scored)
	if len(ranked) != len(scored) {
		t.Fatalf("re-ranking must not drop items: %d of %d", len(ranked), len(scored))
	}
	stars := 0
	for _, item := range ranked[:8] {
		if item.Source == SourcePerformer {
			stars++
		}
	}
	if stars > maxPerPerformerPerPage {
		t.Errorf("%d scenes of one performer in the first slots, cap is %d", stars, maxPerPerformerPerPage)
	}
	if ranked[0].SceneID != "scene:star0" {
		t.Errorf("the best item still leads, got %s", ranked[0].SceneID)
	}
}

// kindItems builds count items of one source on distinct performers, scored
// from top downwards, spread over several sites so no diversity cap binds.
func kindItems(prefix, source string, count int, top float64) []scoredItem {
	items := []scoredItem{}
	for index := range count {
		item := siteItem(fmt.Sprintf("scene:%s%d", prefix, index), source, fmt.Sprintf("site%d", index%7), top-float64(index)*0.01)
		items = append(items, item)
	}
	return items
}

// kindCountsOf counts the items of ranked per source kind.
func kindCountsOf(ranked []Item) map[string]int {
	counts := map[string]int{}
	for _, item := range ranked {
		counts[sourceKind(item.Source)]++
	}
	return counts
}

func TestRerankSharesEveryPageByKind(t *testing.T) {
	scored := kindItems("related", SourceRelatedSite, 40, 9)
	scored = append(scored, kindItems("taste", SourceTaste, 40, 5)...)
	scored = append(scored, kindItems("catalog", SourceCatalog, 40, 4)...)
	scored = append(scored, kindItems("explore", SourceAdjacentTag, 40, 3)...)
	scored = append(scored, kindItems("search", SourceSearch, 40, 2)...)
	scored = append(scored, kindItems("newest", SourceNewest, 40, 1)...)
	sortByScore(scored)

	ranked := rerank(scored)
	for pageNumber := range 2 {
		counts := kindCountsOf(ranked[pageNumber*pageSize : (pageNumber+1)*pageSize])
		for kind, quota := range kindQuotas {
			if counts[kind] != quota {
				t.Errorf("page %d: %d %s items, quota is %d", pageNumber, counts[kind], kind, quota)
			}
		}
		if counts[kindFiller] != 0 {
			t.Errorf("page %d: %d filler items while every kind has enough", pageNumber, counts[kindFiller])
		}
	}
	personalInFirstSlots := kindCountsOf(ranked[:4])[kindPersonal]
	if personalInFirstSlots > 2 {
		t.Errorf("the strongest kind took %d of the first 4 slots, it must spread over the page", personalInFirstSlots)
	}
	if ranked[0].SceneID != "scene:related0" {
		t.Errorf("the best item still leads, got %s", ranked[0].SceneID)
	}
}

func TestRerankGivesAnExhaustedKindsSlotsToTheBestRemaining(t *testing.T) {
	scored := kindItems("related", SourceRelatedSite, 3, 9)
	scored = append(scored, kindItems("taste", SourceTaste, 30, 1)...)
	// Filler of several sources, so the source-run rule does not pull taste in.
	scored = append(scored, kindItems("newest", SourceNewest, 10, 5)...)
	scored = append(scored, kindItems("other", "other:a", 10, 5)...)
	scored = append(scored, kindItems("another", "other:b", 10, 5)...)
	sortByScore(scored)

	counts := kindCountsOf(rerank(scored)[:pageSize])
	if counts[kindPersonal] != 3 || counts[kindTaste] != kindQuotas[kindTaste] || counts[kindFiller] != pageSize-3-kindQuotas[kindTaste] {
		t.Errorf("page kinds = %v: want the 3 personal, the taste quota, and the best of the rest (newest) after them", counts)
	}
}

func TestRerankKeepsSearchBeyondItsQuotaForLast(t *testing.T) {
	scored := kindItems("search", SourceSearch, 10, 9)
	scored = append(scored, kindItems("newest", SourceNewest, 5, 1)...)
	sortByScore(scored)

	ranked := rerank(scored)
	counts := kindCountsOf(ranked[:kindQuotas[kindSearch]+5])
	if counts[kindSearch] != kindQuotas[kindSearch] || counts[kindFiller] != 5 {
		t.Errorf("first slots hold %v: want the search quota, then newest before more search", counts)
	}
	if len(ranked) != len(scored) {
		t.Errorf("re-ranking must not drop items: %d of %d", len(ranked), len(scored))
	}
}

func TestSourceKindGroupsEverySource(t *testing.T) {
	want := map[string]string{
		SourceRelatedSite: kindPersonal, SourceRelatedPerformer: kindPersonal, SourceRelatedTag: kindPersonal,
		SourceRelatedTitle: kindPersonal, SourceSubscription: kindPersonal, SourcePerformer: kindPersonal,
		SourceStudio: kindPersonal, SourceTag: kindPersonal, SourceRelatedWalk: kindPersonal,
		SourceTaste: kindTaste, SourceCatalog: kindCatalog, SourcePopular: kindCatalog,
		SourceAdjacentTag: kindExplore, SourceRandom: kindExplore, SourceExploreSite: kindExplore,
		SourceExploreAttribute: kindExplore,
		SourceSearch:           kindSearch, SourceNewest: kindFiller, "unknown": kindFiller,
	}
	for source, kind := range want {
		if got := sourceKind(source); got != kind {
			t.Errorf("sourceKind(%q) = %q, want %q", source, got, kind)
		}
	}
	for source := range sourcePriors {
		if _, known := want[source]; !known {
			t.Errorf("source %q has a prior but no kind in this test", source)
		}
	}
}

func TestRerankBreaksLongSourceRuns(t *testing.T) {
	scored := []scoredItem{
		syntheticItem("scene:a", SourceSearch, "performer:a", 5),
		syntheticItem("scene:b", SourceSearch, "performer:b", 4),
		syntheticItem("scene:c", SourceSearch, "performer:c", 3),
		syntheticItem("scene:d", SourceNewest, "performer:d", 1),
	}
	ranked := rerank(scored)
	if ranked[2].SceneID != "scene:d" {
		t.Errorf("a third same-source item in a row must yield, got %s", ranked[2].SceneID)
	}
}

// siteItem builds a scored item from site with its own performer.
func siteItem(sceneID, source, site string, score float64) scoredItem {
	item := syntheticItem(sceneID, source, "performer:"+sceneID, score)
	item.features.site = site
	return item
}

func TestRerankSpreadsSites(t *testing.T) {
	scored := []scoredItem{}
	// A fresh scrape: one site outscores everything else.
	for index := range 20 {
		scored = append(scored, siteItem(fmt.Sprintf("scene:fresh%d", index), SourceCatalog, "freshsite", float64(100-index)))
	}
	for index := range 10 {
		source := SourceTag
		if index%2 == 0 {
			source = SourceStudio
		}
		scored = append(scored, siteItem(fmt.Sprintf("scene:a%d", index), source, "sitea", float64(10-index)))
		scored = append(scored, siteItem(fmt.Sprintf("scene:b%d", index), source, "siteb", float64(10-index)))
	}

	ranked := rerank(scored)
	sites := map[string]string{}
	for _, item := range scored {
		sites[item.SceneID] = item.features.site
	}
	run, perSite := 0, map[string]int{}
	for index, item := range ranked[:pageSize] {
		site := sites[item.SceneID]
		perSite[site]++
		run++
		if index == 0 || sites[ranked[index-1].SceneID] != site {
			run = 1
		}
		if run > maxSameSiteRun {
			t.Errorf("slot %d: %d scenes of %s in a row, limit is %d", index, run, site, maxSameSiteRun)
		}
	}
	if perSite["freshsite"] > maxPerSitePerPage {
		t.Errorf("%d scenes of one site on the first page, cap is %d", perSite["freshsite"], maxPerSitePerPage)
	}
	if ranked[0].SceneID != "scene:fresh0" {
		t.Errorf("the best item still leads, got %s", ranked[0].SceneID)
	}
}

func TestRerankKeepsSingleSiteFeedWhole(t *testing.T) {
	scored := []scoredItem{}
	for index := range 30 {
		scored = append(scored, siteItem(fmt.Sprintf("scene:only%d", index), SourceNewest, "onlysite", float64(30-index)))
	}
	ranked := rerank(scored)
	if len(ranked) != len(scored) {
		t.Fatalf("re-ranking must not drop items: %d of %d", len(ranked), len(scored))
	}
	if ranked[0].SceneID != "scene:only0" || ranked[1].SceneID != "scene:only1" {
		t.Errorf("a single-site feed keeps score order, got %s, %s", ranked[0].SceneID, ranked[1].SceneID)
	}
}

func TestImpressionsDriveFatigueUntilClicked(t *testing.T) {
	ctx := context.Background()
	engine, database, _ := newTestEngine(t)
	insertScene(t, database, "scene:shown", map[string]any{})
	shown := []Impression{{SceneID: "scene:shown", Source: SourceNewest, Surface: "feed"}}
	for range 3 {
		if _, err := engine.RecordImpressions(ctx, shown); err != nil {
			t.Fatalf("record: %v", err)
		}
	}

	ignored, err := engine.ignoredImpressions(ctx, testNow)
	if err != nil || ignored["scene:shown"].count != 3 {
		t.Fatalf("ignored = %+v (%v)", ignored, err)
	}
	assertClose(t, "fatigue", fatiguePenalty(3), 2*penaltyFatiguePerShow)

	engine.now = func() time.Time { return testNow.Add(time.Minute) }
	if _, err := engine.RecordImpressions(ctx, []Impression{{SceneID: "scene:shown", Clicked: true}}); err != nil {
		t.Fatalf("record click: %v", err)
	}
	ignored, _ = engine.ignoredImpressions(ctx, testNow)
	if _, found := ignored["scene:shown"]; found {
		t.Error("a click must clear earlier showings")
	}
}

func TestRecordImpressionsValidates(t *testing.T) {
	ctx := context.Background()
	engine, _, _ := newTestEngine(t)
	if _, err := engine.RecordImpressions(ctx, []Impression{{SceneID: "not-an-id"}}); err == nil {
		t.Error("an invalid media id must be rejected")
	}
	tooMany := make([]Impression, ImpressionBatchLimit+1)
	if _, err := engine.RecordImpressions(ctx, tooMany); err == nil {
		t.Error("an oversized batch must be rejected")
	}
}
