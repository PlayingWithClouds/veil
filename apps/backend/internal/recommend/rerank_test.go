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

func TestRerankCapsPerformersAndReservesExploration(t *testing.T) {
	scored := []scoredItem{}
	for index := range 6 {
		scored = append(scored, syntheticItem(fmt.Sprintf("scene:star%d", index), SourcePerformer, "performer:star", float64(10-index)))
	}
	for index := range 6 {
		scored = append(scored, syntheticItem(fmt.Sprintf("scene:other%d", index), SourceTag, fmt.Sprintf("performer:other%d", index), float64(4-index)))
	}
	scored = append(scored,
		syntheticItem("scene:explore0", SourceAdjacentTag, "performer:x0", -5),
		syntheticItem("scene:explore1", SourceRandom, "performer:x1", -6))

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
	if !isExploration(ranked[explorationEvery-1].Source) {
		t.Errorf("slot %d is reserved for exploration, got %s", explorationEvery-1, ranked[explorationEvery-1].Source)
	}
	if ranked[0].SceneID != "scene:star0" {
		t.Errorf("the best item still leads, got %s", ranked[0].SceneID)
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
		scored = append(scored, siteItem(fmt.Sprintf("scene:fresh%d", index), SourceNewest, "freshsite", float64(100-index)))
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
