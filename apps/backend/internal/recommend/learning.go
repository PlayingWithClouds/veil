package recommend

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

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
	// clicks is the decayed number of opened showings, for the log only.
	clicks float64
}

// learnedPolicy is what one ranking learned from the past ones: a factor on
// each source's hand-set prior and the page quotas per kind. A nil policy is
// the hand-set one.
type learnedPolicy struct {
	multipliers map[string]float64
	quotas      map[string]int
	// pooledRate is the reward per showing over all sources.
	pooledRate float64
	// rates is each source's observed reward per showing, for the log only.
	rates map[string]float64
	// examination describes the fitted position model, for the log only.
	examination string
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

// Position-bias model of the learned priors. A showing deep in a list is
// examined less than one at the top, so a source that mostly fills deep
// positions would look worse than it is. Rewards are therefore divided by the
// examination propensity of the position they were shown at (inverse
// propensity weighting), estimated per surface from the logged positions.
const (
	// positionBucketCount is the number of position buckets; bucket i holds
	// the positions below positionBucketBounds[i], the last one the rest.
	positionBucketCount = 7
	// unknownPositionBucket marks showings logged before positions were sent.
	unknownPositionBucket = -1
	// positionPriorSlope shapes the examination curve assumed before the
	// data says otherwise: 1 / (1 + slope * bucket).
	positionPriorSlope = 0.5
	// positionPriorStrength is the weight of that curve in the estimate, in
	// pseudo-rewards: the data overrides it once a bucket has earned about
	// this many rewards.
	positionPriorStrength = 5.0
	// positionMinExposure floors a bucket's examination propensity (before
	// 1/propensity weighting), so a sparse bucket cannot multiply a reward by
	// more than 1/positionMinExposure.
	positionMinExposure = 0.25
	// positionFitIterations is how often source quality and examination are
	// re-estimated from each other.
	positionFitIterations = 8
)

// positionBucketBounds are the upper bounds (exclusive) of the first
// positionBucketCount-1 buckets: positions 0 | 1-2 | 3-6 | 7-14 | 15-30 |
// 31-62 | 63+.
var positionBucketBounds = []int{1, 3, 7, 15, 31, 63}

// positionBucket is the bucket a 0-based position falls into.
func positionBucket(position int) int {
	for bucket, bound := range positionBucketBounds {
		if position < bound {
			return bucket
		}
	}
	return len(positionBucketBounds)
}

// positionBucketSQL renders positionBucket for the column holding a
// position: the unknown bucket for NULL.
func positionBucketSQL(column string) string {
	var cases strings.Builder
	fmt.Fprintf(&cases, "CASE WHEN %s IS NULL THEN %d", column, unknownPositionBucket)
	for bucket, bound := range positionBucketBounds {
		fmt.Fprintf(&cases, " WHEN %s < %d THEN %d", column, bound, bucket)
	}
	fmt.Fprintf(&cases, " ELSE %d END", len(positionBucketBounds))
	return cases.String()
}

// outcomeQuery counts, per source, day, surface and position bucket, the
// showings from learnWindow that are old enough to have an outcome, the opened
// ones and their reward. Only showings on the recommendation surfaces count
// (see recommendationSurfaces): a scene shown on a performer page or in
// search results was not a choice of the engine. A showing is opened when a
// logged click or a first visit of the scene follows within the attribution
// window and no later showing of the scene (on any surface) came in between
// (that one gets the credit). Of the opened ones, a watch that ended the
// session (see watchWeight) or a like within the downstream window adds to
// the reward.
var outcomeQuery = strings.Replace(outcomeQueryTemplate, "$positionBucket", positionBucketSQL("shown.position"), 1)

const outcomeQueryTemplate = `SELECT source, substr(shown_at, 1, 10) AS day, surface, bucket,
	count(*) AS shows, sum(opened) AS clicks,
	sum(opened * min(1.0, $rewardClick + $rewardSessionEnd * ended_session + $rewardLiked * liked)) AS reward
	FROM (
	  SELECT shown.source, shown.shown_at, shown.surface, $positionBucket AS bucket,
	    EXISTS (SELECT 1 FROM (
	        SELECT media, shown_at AS at FROM recommendation_impression WHERE kind = 'click'
	        UNION ALL
	        SELECT id, detail_fetched_at FROM scene WHERE detail_fetched_at IS NOT NULL
	      ) AS engagement
	      WHERE engagement.media = shown.media AND engagement.at >= shown.shown_at
	        AND engagement.at <= strftime('%Y-%m-%dT%H:%M:%fZ', shown.shown_at, $attribution)
	        AND NOT EXISTS (SELECT 1 FROM recommendation_impression AS later
	          WHERE later.media = shown.media AND later.kind = 'shown'
	            AND later.shown_at > shown.shown_at AND later.shown_at <= engagement.at)) AS opened,
	    EXISTS (SELECT 1 FROM watch_history
	      WHERE watch_history.media = shown.media
	        AND watch_history.max_progress_seconds >= $sessionEndMinimum
	        AND watch_history.updated_at >= shown.shown_at
	        AND watch_history.updated_at <= strftime('%Y-%m-%dT%H:%M:%fZ', shown.shown_at, $downstream)
	        AND watch_history.updated_at <= $sessionDecided
	        AND NOT EXISTS (SELECT 1 FROM watch_history AS next
	          WHERE next.media != watch_history.media
	            AND ((next.updated_at > watch_history.updated_at
	                  AND next.updated_at <= strftime('%Y-%m-%dT%H:%M:%fZ', watch_history.updated_at, $sessionGap))
	              OR (next.started_at > watch_history.updated_at
	                  AND next.started_at <= strftime('%Y-%m-%dT%H:%M:%fZ', watch_history.updated_at, $sessionGap))))) AS ended_session,
	    EXISTS (SELECT 1 FROM user_rating
	      WHERE user_rating.media = shown.media AND user_rating.rating > $neutral
	        AND user_rating.updated_at >= shown.shown_at
	        AND user_rating.updated_at <= strftime('%Y-%m-%dT%H:%M:%fZ', shown.shown_at, $downstream)) AS liked
	  FROM recommendation_impression AS shown
	  WHERE shown.kind = 'shown' AND shown.source != ''
	    AND shown.surface IN (SELECT value FROM json_each($surfaces))
	    AND shown.shown_at >= $since AND shown.shown_at <= $until
	) GROUP BY source, day, surface, bucket`

// outcomeCell is where showings are counted for the position model: a
// source's showings at one position bucket of one surface.
type outcomeCell struct {
	source  string
	surface string
	bucket  int
}

// sourceOutcomes reads the decayed showings and rewards of every source, the
// rewards corrected for the examination of the positions they were shown at.
// The fitted examination model comes back for the log.
func (e *Engine) sourceOutcomes(ctx context.Context, now time.Time) (map[string]outcomeCounts, *examinationModel, error) {
	rows, err := e.database.Query(ctx, outcomeQuery, db.Vars{
		"rewardClick":       learnRewardClick,
		"rewardSessionEnd":  learnRewardSessionEnd,
		"sessionEndMinimum": sessionEndMinimumSeconds,
		"sessionGap":        fmt.Sprintf("+%d seconds", int(sessionGap.Seconds())),
		"sessionDecided":    db.FormatTime(now.Add(-sessionGap)),
		"rewardLiked":       learnRewardLiked,
		"neutral":           ratingNeutral,
		"attribution":       fmt.Sprintf("+%d seconds", int(learnAttributionWindow.Seconds())),
		"downstream":        fmt.Sprintf("+%d seconds", int(learnDownstreamWindow.Seconds())),
		"surfaces":          recommendationSurfaces,
		"since":             db.FormatTime(now.Add(-learnWindow)),
		"until":             db.FormatTime(now.Add(-learnAttributionWindow)),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("read source outcomes: %w", err)
	}
	cells := map[outcomeCell]outcomeCounts{}
	for _, row := range rows {
		day, err := time.Parse("2006-01-02", rowString(row, "day"))
		if err != nil {
			continue
		}
		// A day's showings are dated at its middle.
		weight := decayFactor(now.Sub(day.Add(12*time.Hour)), AffinityHalfLife)
		cell := outcomeCell{source: rowString(row, "source"), surface: canonicalSurface(rowString(row, "surface")), bucket: rowInt(row, "bucket")}
		counts := cells[cell]
		counts.shows += weight * float64(rowInt(row, "shows"))
		counts.clicks += weight * float64(rowInt(row, "clicks"))
		counts.reward += weight * rowFloat(row, "reward")
		cells[cell] = counts
	}
	model := fitExamination(cells)
	return model.correctedBySource(cells), model, nil
}

// learnPolicy reads the outcomes and draws this ranking's policy; nil (the
// hand-set values) while there is too little data.
func (e *Engine) learnPolicy(ctx context.Context, now time.Time) (*learnedPolicy, error) {
	if e.learningDisabled {
		return nil, nil
	}
	outcomes, model, err := e.sourceOutcomes(ctx, now)
	if err != nil {
		return nil, err
	}
	policy := learnFrom(outcomes, e.random)
	if policy != nil {
		policy.examination = model.describe()
	}
	e.logPolicy(policy, now)
	return policy, nil
}

// learnFrom draws one Thompson sample per source and per kind. Each has a
// Beta posterior over its reward per showing, starting from the pooled rate
// with learnPriorClicks of strength, so a source without data keeps its
// hand-set prior; the sample relative to the pooled rate, clamped to
// [learnMinMultiplier, learnMaxMultiplier], scales the prior. The kinds'
// samples shift kindQuotas the same way (allocateQuotas). Nil when the
// showings total less than learnMinShows.
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
	policy := &learnedPolicy{multipliers: map[string]float64{}, pooledRate: pooledRate, rates: map[string]float64{}}
	for _, source := range sortedKeys(bySource) {
		counts := bySource[source]
		policy.multipliers[source] = sampledMultiplier(counts, pooledRate, random)
		policy.rates[source] = counts.reward / counts.shows
	}
	targets := map[string]float64{}
	for _, kind := range sortedKeys(kindQuotas) {
		targets[kind] = float64(kindQuotas[kind]) * sampledMultiplier(byKind[kind], pooledRate, random)
	}
	policy.quotas = allocateQuotas(targets, pageSize)
	return policy
}

// plus adds two counts.
func (counts outcomeCounts) plus(other outcomeCounts) outcomeCounts {
	return outcomeCounts{shows: counts.shows + other.shows, reward: counts.reward + other.reward, clicks: counts.clicks + other.clicks}
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
// the prior of learnPriorClicks pseudo-rewards at the pooled rate, plus what
// the source earned (reward as successes, the rest of its showings as
// failures).
func betaPosterior(counts outcomeCounts, pooledRate float64) (alpha, beta float64) {
	priorShows := learnPriorClicks / pooledRate
	alpha = learnPriorClicks + counts.reward
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

// allocateQuotas turns the kinds' wanted slots into whole slots summing to
// total: scaled to fit, each kept within kindQuotaMinimums and
// learnQuotaMaxFactor times its default, the surplus of a clamped kind
// spread over the others, then rounded by largest remainder.
func allocateQuotas(targets map[string]float64, total int) map[string]int {
	minimums, maximums := map[string]float64{}, map[string]float64{}
	for kind, quota := range kindQuotas {
		minimums[kind] = float64(kindQuotaMinimums[kind])
		maximums[kind] = math.Floor(float64(quota) * learnQuotaMaxFactor)
	}
	shares := fitShares(targets, minimums, maximums, float64(total))
	return roundShares(shares, maximums, total)
}

// fitShares scales targets to sum to total within the bounds: clamp the kinds
// that leave their bounds (the side that matters for the sum first), and
// share what is left over among the rest.
func fitShares(targets, minimums, maximums map[string]float64, total float64) map[string]float64 {
	fixed := map[string]float64{}
	for {
		remaining, free := total, 0.0
		for _, value := range fixed {
			remaining -= value
		}
		for kind, target := range targets {
			if _, isFixed := fixed[kind]; !isFixed {
				free += target
			}
		}
		scaled := map[string]float64{}
		surplus := 0.0
		for kind, target := range targets {
			if _, isFixed := fixed[kind]; isFixed {
				continue
			}
			scaled[kind] = target * remaining / free
			surplus += clampFloat(scaled[kind], minimums[kind], maximums[kind]) - scaled[kind]
		}
		violators := violatingKinds(scaled, minimums, maximums, surplus)
		if len(violators) == 0 {
			for kind, value := range scaled {
				fixed[kind] = value
			}
			return fixed
		}
		for _, kind := range violators {
			fixed[kind] = clampFloat(scaled[kind], minimums[kind], maximums[kind])
		}
	}
}

// violatingKinds lists the kinds whose scaled share leaves its bounds, the
// side that decides the sum first: below the minimum when clamping adds slots
// overall (surplus > 0), above the maximum when it takes some away.
func violatingKinds(scaled, minimums, maximums map[string]float64, surplus float64) []string {
	var below, above []string
	for _, kind := range sortedKeys(scaled) {
		if scaled[kind] < minimums[kind] {
			below = append(below, kind)
		}
		if scaled[kind] > maximums[kind] {
			above = append(above, kind)
		}
	}
	if surplus > 0 && len(below) > 0 {
		return below
	}
	if surplus < 0 && len(above) > 0 {
		return above
	}
	return append(below, above...)
}

// roundShares rounds shares that sum to total down and hands the remaining
// slots to the largest fractions, never past a kind's maximum.
func roundShares(shares, maximums map[string]float64, total int) map[string]int {
	out := map[string]int{}
	assigned := 0
	for kind, value := range shares {
		out[kind] = int(math.Floor(value + 1e-9))
		assigned += out[kind]
	}
	kinds := sortedKeys(shares)
	sort.SliceStable(kinds, func(left, right int) bool {
		return shares[kinds[left]]-math.Floor(shares[kinds[left]]) > shares[kinds[right]]-math.Floor(shares[kinds[right]])
	})
	for assigned < total {
		progressed := false
		for _, kind := range kinds {
			if assigned < total && float64(out[kind]) < maximums[kind] {
				out[kind]++
				assigned++
				progressed = true
			}
		}
		if !progressed {
			break
		}
	}
	return out
}

// logPolicy prints the learned multipliers and quotas, at most once per
// learnLogInterval.
func (e *Engine) logPolicy(policy *learnedPolicy, now time.Time) {
	if policy == nil || now.Sub(e.lastPolicyLog) < learnLogInterval {
		return
	}
	e.lastPolicyLog = now
	var sources []string
	for _, source := range sortedKeys(policy.multipliers) {
		sources = append(sources, fmt.Sprintf("%s=x%.2f(%.2f%%)", source, policy.multipliers[source], 100*policy.rates[source]))
	}
	var quotas []string
	for _, kind := range sortedKeys(policy.quotas) {
		quotas = append(quotas, fmt.Sprintf("%s=%d", kind, policy.quotas[kind]))
	}
	log.Printf("recommend: learned priors (pooled reward rate %.2f%%): %s; quotas: %s; position examination: %s",
		100*policy.pooledRate, strings.Join(sources, " "), strings.Join(quotas, " "), policy.examination)
}

// clampFloat limits value to [low, high].
func clampFloat(value, low, high float64) float64 {
	return math.Min(math.Max(value, low), high)
}

// sortedKeys lists a map's keys in order, so draws and logs are reproducible.
func sortedKeys[Value any](values map[string]Value) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
