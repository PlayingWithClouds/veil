package recommend

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// Learning needs to know which source an item shown on Home came from, but
// impressions only name the item. So every page Recommendations serves is
// remembered in CacheEntry ("recommend:served:<kind>:<id>" → source, kept for
// learnWindow), and an impression is credited to the source its item was
// last served from.

// servedPrefix starts the CacheEntry keys of served items.
const servedPrefix = "recommend:served:"

// sampler is the random source of Thompson sampling: *rand.Rand, or a seeded
// one in tests.
type sampler interface {
	Float64() float64
	NormFloat64() float64
}

// outcomeCounts is what one source (or kind) earned: its showings and the
// reward they brought, both decayed by age.
type outcomeCounts struct {
	shows  float64
	reward float64
}

// plus adds two counts.
func (counts outcomeCounts) plus(other outcomeCounts) outcomeCounts {
	return outcomeCounts{shows: counts.shows + other.shows, reward: counts.reward + other.reward}
}

// learnedPolicy is what one ranking learned from the past ones: a factor on
// each source's hand-set prior and the page quotas per kind. A nil policy is
// the hand-set one.
type learnedPolicy struct {
	multipliers map[string]float64
	quotas      map[string]int
}

// multiplier is the factor on the source's hand-set prior; 1 when nothing
// was learned about it.
func (policy *learnedPolicy) multiplier(source string) float64 {
	if policy == nil {
		return 1
	}
	factor, learned := policy.multipliers[source]
	if !learned {
		return 1
	}
	return factor
}

// pageQuotas is how many page slots each kind may fill.
func (policy *learnedPolicy) pageQuotas() map[string]int {
	if policy == nil {
		return kindQuotas
	}
	return policy.quotas
}

// rememberServed records the source each served item came from, for learning.
func (e *Engine) rememberServed(ctx context.Context, items []scoredItem) error {
	if len(items) == 0 {
		return nil
	}
	served := make([]map[string]string, 0, len(items))
	for _, item := range items {
		served = append(served, map[string]string{"key": refKey(item.item), "source": item.source})
	}
	_, err := e.database.Exec(ctx,
		`INSERT INTO "CacheEntry" ("key", "value", "expiresAt")
		 SELECT $prefix || json_extract(value, '$.key'), json_extract(value, '$.source'), $expires FROM json_each($served) WHERE 1
		 ON CONFLICT ("key") DO UPDATE SET "value" = excluded."value", "expiresAt" = excluded."expiresAt"`,
		db.Vars{"prefix": servedPrefix, "served": served, "expires": e.now().Add(learnWindow).Unix()})
	if err != nil {
		return fmt.Errorf("remember served items: %w", err)
	}
	return nil
}

// outcomeQuery counts, per source and day, the showings on the recommendation
// surfaces old enough to have an outcome, and their reward: learnRewardOpen
// when a choice or view of the item followed within the attribution window,
// plus learnRewardDownstream when a finish, o or like followed within the
// downstream window.
const outcomeQuery = `SELECT "served"."value" AS "source", substr("shown"."at", 1, 10) AS "day", COUNT(*) AS "shows",
	SUM("shown"."opened" * MIN(1.0, $rewardOpen + $rewardDownstream * "shown"."rewarded")) AS "reward"
	FROM (
	  SELECT "impression"."at", "impression"."subjectKind", "impression"."subjectId",
	    EXISTS (SELECT 1 FROM "Event" AS "opened"
	      WHERE "opened"."subjectKind" = "impression"."subjectKind" AND "opened"."subjectId" = "impression"."subjectId"
	        AND "opened"."kind" IN ('choice', 'view') AND "opened"."deletedAt" IS NULL
	        AND "opened"."at" >= "impression"."at"
	        AND "opened"."at" <= strftime('%Y-%m-%dT%H:%M:%fZ', "impression"."at", $attribution)) AS "opened",
	    EXISTS (SELECT 1 FROM "Event" AS "outcome"
	      WHERE "outcome"."subjectKind" = "impression"."subjectKind" AND "outcome"."subjectId" = "impression"."subjectId"
	        AND ("outcome"."kind" IN ('finish', 'o') OR ("outcome"."kind" = 'like' AND json_extract("outcome"."data", '` + pathValue + `') > 0))
	        AND "outcome"."deletedAt" IS NULL AND "outcome"."at" >= "impression"."at"
	        AND "outcome"."at" <= strftime('%Y-%m-%dT%H:%M:%fZ', "impression"."at", $downstream)) AS "rewarded"
	  FROM "Event" AS "impression"
	  WHERE "impression"."kind" = 'impression' AND "impression"."deletedAt" IS NULL
	    AND "impression"."at" >= $since AND "impression"."at" <= $until
	    AND json_extract("impression"."data", '` + pathSurface + `') IN (SELECT value FROM json_each($surfaces))
	) AS "shown"
	JOIN "CacheEntry" AS "served" ON "served"."key" = $prefix || "shown"."subjectKind" || ':' || "shown"."subjectId"
	GROUP BY "source", "day"`

// sourceOutcomes reads the decayed showings and rewards of every source.
func (e *Engine) sourceOutcomes(ctx context.Context, now time.Time) (map[string]outcomeCounts, error) {
	rows, err := e.database.Query(ctx, outcomeQuery, db.Vars{
		"rewardOpen":       learnRewardOpen,
		"rewardDownstream": learnRewardDownstream,
		"attribution":      fmt.Sprintf("+%d seconds", int(learnAttributionWindow.Seconds())),
		"downstream":       fmt.Sprintf("+%d seconds", int(learnDownstreamWindow.Seconds())),
		"since":            db.FormatTime(now.Add(-learnWindow)),
		"until":            db.FormatTime(now.Add(-learnAttributionWindow)),
		"surfaces":         recommendationSurfaces,
		"prefix":           servedPrefix,
	})
	if err != nil {
		return nil, fmt.Errorf("read source outcomes: %w", err)
	}
	out := map[string]outcomeCounts{}
	for _, row := range rows {
		day, err := time.Parse("2006-01-02", rowString(row, "day"))
		if err != nil {
			continue
		}
		// A day's showings are dated at its middle.
		weight := decayFactor(now.Sub(day.Add(12*time.Hour)), affinityHalfLife)
		source := rowString(row, "source")
		out[source] = out[source].plus(outcomeCounts{shows: weight * rowFloat(row, "shows"), reward: weight * rowFloat(row, "reward")})
	}
	return out, nil
}

// learnPolicy reads the outcomes and draws this ranking's policy; nil (the
// hand-set values) while there is too little data.
func (e *Engine) learnPolicy(ctx context.Context, now time.Time) (*learnedPolicy, error) {
	outcomes, err := e.sourceOutcomes(ctx, now)
	if err != nil {
		return nil, err
	}
	return learnFrom(outcomes, e.random), nil
}

// learnFrom draws one Thompson sample per source and per kind. Each has a
// Beta posterior over its reward per showing, starting from the pooled rate
// with learnPriorRewards of strength, so a source without data keeps its
// hand-set prior; the sample relative to the pooled rate, clamped, scales the
// prior, and the kinds' samples scale their quotas. Nil when the showings
// total less than learnMinShows.
func learnFrom(bySource map[string]outcomeCounts, random sampler) *learnedPolicy {
	var total outcomeCounts
	byKind := map[string]outcomeCounts{}
	for source, counts := range bySource {
		total = total.plus(counts)
		kind := sourceKind(source)
		byKind[kind] = byKind[kind].plus(counts)
	}
	if total.shows < learnMinShows {
		return nil
	}
	pooledRate := clampFloat(total.reward/total.shows, learnMinPooledRate, learnMaxPooledRate)
	policy := &learnedPolicy{multipliers: map[string]float64{}, quotas: map[string]int{}}
	for _, source := range sortedKeys(bySource) {
		policy.multipliers[source] = sampledMultiplier(bySource[source], pooledRate, random)
	}
	for _, kind := range sortedKeys(kindQuotas) {
		policy.quotas[kind] = learnedQuota(kind, sampledMultiplier(byKind[kind], pooledRate, random))
	}
	return policy
}

// learnedQuota scales a kind's default quota, kept between its minimum and
// learnQuotaMaxFactor times the default. Slots the quotas leave open go to
// the best remaining items.
func learnedQuota(kind string, factor float64) int {
	quota := math.Round(float64(kindQuotas[kind]) * factor)
	maximum := math.Floor(float64(kindQuotas[kind]) * learnQuotaMaxFactor)
	return int(clampFloat(quota, float64(kindQuotaMinimums[kind]), maximum))
}

// sampledMultiplier draws from the Beta posterior of counts and returns the
// draw relative to the pooled rate, clamped. No showings, no draw: 1.
func sampledMultiplier(counts outcomeCounts, pooledRate float64, random sampler) float64 {
	if counts.shows <= 0 {
		return 1
	}
	alpha, beta := betaPosterior(counts, pooledRate)
	return clampFloat(sampleBeta(alpha, beta, random)/pooledRate, learnMinMultiplier, learnMaxMultiplier)
}

// betaPosterior is the Beta(alpha, beta) of a source's reward per showing:
// the prior of learnPriorRewards pseudo-rewards at the pooled rate, plus what
// the source earned.
func betaPosterior(counts outcomeCounts, pooledRate float64) (alpha, beta float64) {
	priorShows := learnPriorRewards / pooledRate
	alpha = learnPriorRewards + counts.reward
	beta = priorShows*(1-pooledRate) + math.Max(counts.shows-counts.reward, 0)
	return alpha, beta
}

// sampleBeta draws from Beta(alpha, beta) as X/(X+Y) of two gamma draws.
func sampleBeta(alpha, beta float64, random sampler) float64 {
	x := sampleGamma(alpha, random)
	y := sampleGamma(beta, random)
	if x+y == 0 {
		return alpha / (alpha + beta)
	}
	return x / (x + y)
}

// sampleGamma draws from Gamma(shape, 1) with Marsaglia and Tsang's method;
// a shape under 1 is drawn one higher and scaled down.
func sampleGamma(shape float64, random sampler) float64 {
	if shape < 1 {
		return sampleGamma(shape+1, random) * math.Pow(random.Float64(), 1/shape)
	}
	d := shape - 1.0/3
	c := 1 / math.Sqrt(9*d)
	for {
		normal := random.NormFloat64()
		v := 1 + c*normal
		if v <= 0 {
			continue
		}
		v = v * v * v
		u := random.Float64()
		if u < 1-0.0331*normal*normal*normal*normal || math.Log(u) < 0.5*normal*normal+d*(1-v+math.Log(v)) {
			return d * v
		}
	}
}
