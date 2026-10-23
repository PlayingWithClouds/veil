package recommend

import (
	"context"
	"math"
	"math/rand"
	"testing"
	"time"
)

// seededRandom is a reproducible random source.
func seededRandom(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}

// averageMultiplier is the mean learned multiplier of source over many draws.
func averageMultiplier(outcomes map[string]outcomeCounts, source string) float64 {
	const draws = 300
	sum := 0.0
	for seed := range int64(draws) {
		policy := learnFrom(outcomes, seededRandom(seed))
		sum += policy.multiplier(source)
	}
	return sum / draws
}

// averageQuota is the mean learned quota of kind over many draws.
func averageQuota(outcomes map[string]outcomeCounts, kind string) float64 {
	const draws = 300
	sum := 0
	for seed := range int64(draws) {
		sum += learnFrom(outcomes, seededRandom(seed)).quotas[kind]
	}
	return float64(sum) / draws
}

// pooledOutcomes is a library whose sources click at 2%: enough showings to
// learn from, with related:site, catalog and newest listed on top of it.
func pooledOutcomes(extra map[string]outcomeCounts) map[string]outcomeCounts {
	outcomes := map[string]outcomeCounts{
		SourceTag:          {shows: 1000, reward: 20},
		SourceAdjacentTag:  {shows: 1000, reward: 20},
		SourcePerformer:    {shows: 1000, reward: 20},
		SourceRelatedTitle: {shows: 1000, reward: 20},
	}
	for source, counts := range extra {
		outcomes[source] = counts
	}
	return outcomes
}

func TestBetaPosteriorStartsAtThePooledRate(t *testing.T) {
	alpha, beta := betaPosterior(outcomeCounts{}, 0.02)
	assertClose(t, "prior mean", alpha/(alpha+beta), 0.02)
	assertClose(t, "prior strength in rewards", alpha, learnPriorClicks)

	alpha, beta = betaPosterior(outcomeCounts{shows: 1000, reward: 100}, 0.02)
	if mean := alpha / (alpha + beta); mean <= 0.02 || mean >= 0.1 {
		t.Errorf("posterior mean %v, want between the prior (0.02) and the data (0.1)", mean)
	}
}

func TestSampleBetaStaysInRangeAndCentresOnItsMean(t *testing.T) {
	random := seededRandom(1)
	for _, shape := range [][2]float64{{0.5, 40}, {30, 1500}, {3, 3}} {
		sum := 0.0
		const draws = 4000
		for range draws {
			value := sampleBeta(shape[0], shape[1], random)
			if value < 0 || value > 1 || math.IsNaN(value) {
				t.Fatalf("Beta(%v, %v) drew %v", shape[0], shape[1], value)
			}
			sum += value
		}
		want := shape[0] / (shape[0] + shape[1])
		if mean := sum / draws; math.Abs(mean-want) > 0.1*want {
			t.Errorf("Beta(%v, %v) mean %v, want about %v", shape[0], shape[1], mean, want)
		}
	}
}

func TestLearningNeedsEnoughShowings(t *testing.T) {
	if policy := learnFrom(map[string]outcomeCounts{SourceTag: {shows: learnMinShows - 1, reward: 50}}, seededRandom(1)); policy != nil {
		t.Error("too few showings must leave the hand-set policy (nil)")
	}
	var none *learnedPolicy
	assertClose(t, "nil multiplier", none.multiplier(SourceTag), 1)
	if got := none.pageQuotas()[kindPersonal]; got != kindQuotas[kindPersonal] {
		t.Errorf("nil policy quota %d, want the default %d", got, kindQuotas[kindPersonal])
	}
}

func TestMultipliersStayWithinBounds(t *testing.T) {
	outcomes := pooledOutcomes(map[string]outcomeCounts{
		SourceRelatedSite: {shows: 5000, reward: 2500},
		SourceCatalog:     {shows: 5000, reward: 0},
		SourceNewest:      {shows: 3, reward: 3},
	})
	for seed := range int64(100) {
		policy := learnFrom(outcomes, seededRandom(seed))
		for source, factor := range policy.multipliers {
			if factor < learnMinMultiplier || factor > learnMaxMultiplier {
				t.Fatalf("seed %d: %s multiplier %v outside [%v, %v]", seed, source, factor, learnMinMultiplier, learnMaxMultiplier)
			}
		}
	}
}

func TestClickedSourceGainsAndIgnoredSourceLoses(t *testing.T) {
	outcomes := pooledOutcomes(map[string]outcomeCounts{
		SourceRelatedSite: {shows: 1000, reward: 80},
		SourceCatalog:     {shows: 1000, reward: 2},
	})
	if average := averageMultiplier(outcomes, SourceRelatedSite); average < 1.2 {
		t.Errorf("a source clicked at 4x the pool averages x%.2f, want a clear gain", average)
	}
	if average := averageMultiplier(outcomes, SourceCatalog); average > 0.8 {
		t.Errorf("a source clicked at a tenth of the pool averages x%.2f, want a clear loss", average)
	}
}

func TestSourceWithoutDataStaysNearItsDefault(t *testing.T) {
	outcomes := pooledOutcomes(map[string]outcomeCounts{SourceStudio: {shows: 4, reward: 0}})
	assertClose(t, "unseen source", learnFrom(outcomes, seededRandom(1)).multiplier(SourceRandom), 1)
	if average := averageMultiplier(outcomes, SourceStudio); average < 0.9 || average > 1.1 {
		t.Errorf("a source with 4 showings averages x%.2f, want about 1", average)
	}
}

func TestLearningIsReproducibleWithTheSameRandomSource(t *testing.T) {
	outcomes := pooledOutcomes(map[string]outcomeCounts{SourceRelatedSite: {shows: 500, reward: 20}})
	first := learnFrom(outcomes, seededRandom(7))
	second := learnFrom(outcomes, seededRandom(7))
	for source, factor := range first.multipliers {
		if second.multipliers[source] != factor {
			t.Errorf("%s drew %v then %v from the same seed", source, factor, second.multipliers[source])
		}
	}
}

// assertQuotasValid fails unless the quotas sum to the page size and keep
// every kind between its minimum and its maximum.
func assertQuotasValid(t *testing.T, label string, quotas map[string]int) {
	t.Helper()
	sum := 0
	for kind, defaultQuota := range kindQuotas {
		quota := quotas[kind]
		sum += quota
		if quota < kindQuotaMinimums[kind] {
			t.Errorf("%s: %s quota %d under its minimum %d", label, kind, quota, kindQuotaMinimums[kind])
		}
		if maximum := int(math.Floor(float64(defaultQuota) * learnQuotaMaxFactor)); quota > maximum {
			t.Errorf("%s: %s quota %d over its maximum %d", label, kind, quota, maximum)
		}
	}
	if sum != pageSize {
		t.Errorf("%s: quotas %v sum to %d, want %d", label, quotas, sum, pageSize)
	}
}

func TestQuotasWithoutDataAreTheDefaults(t *testing.T) {
	targets := map[string]float64{}
	for kind, quota := range kindQuotas {
		targets[kind] = float64(quota)
	}
	quotas := allocateQuotas(targets, pageSize)
	for kind, quota := range kindQuotas {
		if quotas[kind] != quota {
			t.Errorf("%s quota %d, want the default %d", kind, quotas[kind], quota)
		}
	}
}

func TestQuotasKeepTheirMinimaAndSum(t *testing.T) {
	extremes := []map[string]float64{
		{kindPersonal: 100, kindTaste: 0.1, kindCatalog: 0.1, kindExplore: 0.1, kindSearch: 0.1},
		{kindPersonal: 0.1, kindTaste: 100, kindCatalog: 100, kindExplore: 0.1, kindSearch: 0.1},
		{kindPersonal: 4, kindTaste: 7.5, kindCatalog: 7.5, kindExplore: 6, kindSearch: 3},
		{kindPersonal: 4, kindTaste: 2.5, kindCatalog: 2.5, kindExplore: 2, kindSearch: 1},
	}
	for index, targets := range extremes {
		assertQuotasValid(t, "extreme "+string(rune('a'+index)), allocateQuotas(targets, pageSize))
	}
	random := seededRandom(3)
	for draw := range 500 {
		targets := map[string]float64{}
		for kind, quota := range kindQuotas {
			targets[kind] = float64(quota) * (learnMinMultiplier + random.Float64()*(learnMaxMultiplier-learnMinMultiplier))
		}
		assertQuotasValid(t, "random draw "+string(rune('0'+draw%10)), allocateQuotas(targets, pageSize))
	}
}

func TestLearnedQuotasFollowTheClicks(t *testing.T) {
	outcomes := pooledOutcomes(map[string]outcomeCounts{
		SourceRelatedSite: {shows: 2000, reward: 200},
		SourceCatalog:     {shows: 2000, reward: 4},
		SourceSearch:      {shows: 2000, reward: 4},
	})
	for seed := range int64(100) {
		assertQuotasValid(t, "learned", learnFrom(outcomes, seededRandom(seed)).quotas)
	}
	if personal := averageQuota(outcomes, kindPersonal); personal <= float64(kindQuotas[kindPersonal]) {
		t.Errorf("personal averages %.1f slots, want more than its default %d", personal, kindQuotas[kindPersonal])
	}
	if catalog := averageQuota(outcomes, kindCatalog); catalog >= float64(kindQuotas[kindCatalog]) {
		t.Errorf("catalog averages %.1f slots, want fewer than its default %d", catalog, kindQuotas[kindCatalog])
	}
}

func TestRerankFollowsLearnedQuotas(t *testing.T) {
	scored := kindItems("related", SourceRelatedSite, 40, 9)
	scored = append(scored, kindItems("catalog", SourceCatalog, 40, 4)...)
	scored = append(scored, kindItems("taste", SourceTaste, 40, 3)...)
	scored = append(scored, kindItems("explore", SourceAdjacentTag, 40, 3)...)
	scored = append(scored, kindItems("search", SourceSearch, 40, 2)...)
	sortByScore(scored)
	quotas := map[string]int{kindPersonal: 4, kindCatalog: 12, kindTaste: 2, kindExplore: 3, kindSearch: 3}

	counts := kindCountsOf(rerankWithQuotas(scored, quotas)[:pageSize])
	for kind, quota := range quotas {
		if counts[kind] != quota {
			t.Errorf("page kinds = %v, want the learned quotas %v", counts, quotas)
			break
		}
	}
}

func TestScoreScalesTheSourcePriorByItsMultiplier(t *testing.T) {
	proposed := &candidate{sceneID: "scene:a", source: SourceCatalog, prior: 0.3, sourceCount: 1}
	scene := &sceneFeatures{}
	plain := &snapshot{now: testNow, profile: newProfile(0), ignored: map[string]ignoredImpression{}, watched: map[string]watchState{}, popularity: &popularityIndex{}}
	boosted := *plain
	boosted.policy = &learnedPolicy{multipliers: map[string]float64{SourceCatalog: 1.5}}

	gain := score(proposed, scene, &boosted, 0.5) - score(proposed, scene, plain, 0.5)
	assertClose(t, "score gain", gain, rankWeightSourcePrior*0.3*0.5)
}

func TestSourceOutcomesAttributeOpensToTheLatestShowing(t *testing.T) {
	engine, database, _ := newTestEngine(t)
	ctx := context.Background()
	// All showings sit on one day three days back, so they share one decay weight.
	day := testNow.Add(-3 * 24 * time.Hour).Truncate(24 * time.Hour)
	at := func(offset time.Duration) string { return ago(testNow.Sub(day.Add(offset))) }

	for _, id := range []string{"scene:clicked", "scene:reshown", "scene:ignored", "scene:visited", "scene:late", "scene:young", "scene:followed", "scene:next"} {
		insertScene(t, database, id, map[string]any{})
	}
	counter := 0
	impression := func(media, kind, source string, offset time.Duration) {
		counter++
		insertRow(t, database, "recommendation_impression", "recommendation_impression:"+string(rune('a'+counter)),
			map[string]any{"media": media, "kind": kind, "source": source, "shown_at": at(offset)})
	}
	visit := func(media string, offset time.Duration) {
		if _, err := database.Exec(ctx, `UPDATE scene SET detail_fetched_at = $at WHERE id = $id`,
			map[string]any{"at": at(offset), "id": media}); err != nil {
			t.Fatalf("visit %s: %v", media, err)
		}
	}

	// Opened through a logged click.
	impression("scene:clicked", "shown", "src:click", 1*time.Hour)
	impression("scene:clicked", "click", "src:click", 1*time.Hour+time.Minute)
	// Shown twice, then opened: only the second showing gets the credit.
	impression("scene:reshown", "shown", "src:first", 2*time.Hour)
	impression("scene:reshown", "shown", "src:second", 3*time.Hour)
	visit("scene:reshown", 3*time.Hour+time.Minute)
	// Never opened.
	impression("scene:ignored", "shown", "src:ignored", 4*time.Hour)
	// Opened by a first visit, then watched until the session ended and liked: full reward.
	impression("scene:visited", "shown", "src:visit", 5*time.Hour)
	visit("scene:visited", 5*time.Hour+time.Minute)
	insertRow(t, database, "watch_history", "watch_history:visited", map[string]any{"media": "scene:visited", "max_progress_seconds": 600,
		"started_at": at(5*time.Hour + 2*time.Minute), "updated_at": at(6 * time.Hour)})
	// Opened and watched to the end, but another video followed: the open only.
	impression("scene:followed", "shown", "src:followed", 8*time.Hour)
	visit("scene:followed", 8*time.Hour+time.Minute)
	insertRow(t, database, "watch_history", "watch_history:followed", map[string]any{"media": "scene:followed", "completed": true, "max_progress_seconds": 600,
		"started_at": at(8*time.Hour + 2*time.Minute), "updated_at": at(9 * time.Hour)})
	insertRow(t, database, "watch_history", "watch_history:next", map[string]any{"media": "scene:next", "max_progress_seconds": 30,
		"started_at": at(9*time.Hour + 5*time.Minute), "updated_at": at(9*time.Hour + 6*time.Minute)})
	insertRow(t, database, "user_rating", "user_rating:visited", map[string]any{"media": "scene:visited", "rating": 9.0, "updated_at": at(6 * time.Hour)})
	// Opened only after the attribution window: not this showing's doing.
	impression("scene:late", "shown", "src:late", 1*time.Hour)
	visit("scene:late", 1*time.Hour+learnAttributionWindow+time.Hour)
	// Shown a moment ago: the outcome is still open, so it is not counted.
	insertRow(t, database, "recommendation_impression", "recommendation_impression:young",
		map[string]any{"media": "scene:young", "kind": "shown", "source": "src:young", "shown_at": ago(time.Hour)})

	outcomes, err := engine.sourceOutcomes(ctx, testNow)
	if err != nil {
		t.Fatalf("source outcomes: %v", err)
	}
	weight := decayFactor(testNow.Sub(day.Add(12*time.Hour)), AffinityHalfLife)
	want := map[string]outcomeCounts{
		"src:click":    {shows: weight, clicks: weight, reward: weight * learnRewardClick},
		"src:first":    {shows: weight},
		"src:second":   {shows: weight, clicks: weight, reward: weight * learnRewardClick},
		"src:ignored":  {shows: weight},
		"src:visit":    {shows: weight, clicks: weight, reward: weight},
		"src:late":     {shows: weight},
		"src:followed": {shows: weight, clicks: weight, reward: weight * learnRewardClick},
	}
	if len(outcomes) != len(want) {
		t.Errorf("outcomes %v, want sources %v only", outcomes, want)
	}
	for source, expected := range want {
		got := outcomes[source]
		assertClose(t, source+" shows", got.shows, expected.shows)
		assertClose(t, source+" clicks", got.clicks, expected.clicks)
		assertClose(t, source+" reward", got.reward, expected.reward)
	}
}
