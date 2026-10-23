// Package recommendeval replays the Home recommendations over the history of
// a database: at each cutoff it rebuilds the state as of that moment on a
// copy, ranks with the real engine (clock set to the cutoff) and checks the
// ranking against what the user did in the following window.
//
// What a ranking is scored against: the scenes first watched ("any watch"),
// first watched and ended the session ("session-ending watch": at least
// recommend.SessionEndMinimumSeconds played and nothing else within
// recommend.SessionGap) or first opened ("click") inside the window.
// Scenes stored only after the cutoff could not have been recommended and are
// counted apart ("unreachable") instead of as misses.
//
// The reconstruction is an approximation: tables without timestamps (the
// blocklist, plugin settings, tag and performer records) are taken as they are
// now, and a watch under way at the cutoff keeps the playback the wall clock
// allows (see pruneStatements).
package recommendeval

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// Config selects what to evaluate.
type Config struct {
	// DatabasePath is the veil.db to read; it is never modified.
	DatabasePath string
	// Window is how far after a cutoff the user's actions count.
	Window time.Duration
	// Step is the spacing of the cutoffs.
	Step time.Duration
	// From and To bound the cutoffs; zero values default to the first
	// recorded activity plus Warmup, and the last minus Window.
	From, To time.Time
	// Warmup is how much history the first cutoff keeps when From is unset.
	Warmup time.Duration
	// Variants lists the rankings to compare (see the Variant constants).
	Variants []string
	// Repeats ranks that many times per cutoff (the engine samples) and
	// averages.
	Repeats int
	// Seed makes the sampling reproducible.
	Seed int64
}

// Run evaluates the configured rankings on a temporary copy of the database.
func Run(ctx context.Context, config Config) (*Report, error) {
	if err := validate(config); err != nil {
		return nil, err
	}
	directory, err := os.MkdirTemp("", "recommend-eval-")
	if err != nil {
		return nil, fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(directory)
	copyPath := filepath.Join(directory, "veil.db")
	if err := copyDatabase(ctx, config.DatabasePath, copyPath); err != nil {
		return nil, err
	}
	database, err := db.Open(ctx, copyPath)
	if err != nil {
		return nil, fmt.Errorf("open copy: %w", err)
	}
	defer database.Close()
	past, err := loadHistory(ctx, database)
	if err != nil {
		return nil, err
	}
	cutoffs, err := chooseCutoffs(config, past)
	if err != nil {
		return nil, err
	}
	return evaluate(ctx, config, database, past, cutoffs)
}

// validate rejects configurations that cannot run.
func validate(config Config) error {
	if config.DatabasePath == "" {
		return fmt.Errorf("no database path")
	}
	if config.Window <= 0 || config.Step <= 0 {
		return fmt.Errorf("window and step must be positive")
	}
	if len(config.Variants) == 0 {
		return fmt.Errorf("no variants")
	}
	return nil
}

// chooseCutoffs spaces cutoffs Step apart between the bounds.
func chooseCutoffs(config Config, past *history) ([]time.Time, error) {
	if past.firstActivity.IsZero() {
		return nil, fmt.Errorf("the database has no watches or clicks to evaluate against")
	}
	from, to := config.From, config.To
	if from.IsZero() {
		from = past.firstActivity.Add(config.Warmup)
	}
	if to.IsZero() {
		to = past.lastActivity.Add(-config.Window)
	}
	var cutoffs []time.Time
	for cutoff := from.UTC(); !cutoff.After(to); cutoff = cutoff.Add(config.Step) {
		cutoffs = append(cutoffs, cutoff)
	}
	if len(cutoffs) == 0 {
		return nil, fmt.Errorf("no cutoff between %s and %s: activity spans %s to %s",
			from.Format(time.DateOnly), to.Format(time.DateOnly), past.firstActivity.Format(time.DateOnly), past.lastActivity.Format(time.DateOnly))
	}
	return cutoffs, nil
}

// variantTotals accumulates one variant's coverage over cutoffs and repeats.
type variantTotals struct {
	rankings    int
	sites, tags float64
	sourceShare map[string]float64
}

// evaluate walks the cutoffs from the latest to the earliest, so pruning the
// copy further each time never needs a fresh one.
func evaluate(ctx context.Context, config Config, database *db.DB, past *history, cutoffs []time.Time) (*Report, error) {
	clock := &evaluationClock{}
	repeats := max(config.Repeats, 1)
	rankers := map[string][]ranker{}
	totals := map[string]*variantTotals{}
	for _, name := range config.Variants {
		totals[name] = &variantTotals{sourceShare: map[string]float64{}}
		for repeat := 0; repeat < repeats; repeat++ {
			built, err := newRanker(name, database, clock, config.Seed+int64(repeat))
			if err != nil {
				return nil, err
			}
			rankers[name] = append(rankers[name], built)
		}
	}
	report := &Report{
		Database: config.DatabasePath, WindowDays: config.Window.Hours() / 24, StepDays: config.Step.Hours() / 24, Repeats: repeats,
	}
	for index := len(cutoffs) - 1; index >= 0; index-- {
		clock.current = cutoffs[index]
		if err := pruneTo(ctx, database, cutoffs[index]); err != nil {
			return nil, err
		}
		result, err := evaluateCutoff(ctx, config, database, past, cutoffs[index], rankers, totals)
		if err != nil {
			return nil, err
		}
		report.Cutoffs = append(report.Cutoffs, *result)
	}
	sort.Slice(report.Cutoffs, func(left, right int) bool { return report.Cutoffs[left].Cutoff.Before(report.Cutoffs[right].Cutoff) })
	report.dropEmptyCutoffs()
	for _, name := range config.Variants {
		report.Variants = append(report.Variants, summarize(name, report.Cutoffs, totals[name]))
	}
	return report, nil
}

// evaluateCutoff ranks every variant at one cutoff and scores it against the
// window after it.
func evaluateCutoff(ctx context.Context, config Config, database *db.DB, past *history, cutoff time.Time,
	rankers map[string][]ranker, totals map[string]*variantTotals) (*CutoffReport, error) {
	targets := past.targetsIn(cutoff, config.Window)
	result := &CutoffReport{Cutoff: cutoff, Targets: map[string]TargetCounts{}, Hits: map[string]map[string]CutoffHits{}}
	for kind, set := range targets {
		result.Targets[kind] = TargetCounts{Reachable: len(set.reachable), Unreachable: set.unreachable}
	}
	scenes, err := database.Int(ctx, `SELECT count(*) FROM scene`, nil)
	if err != nil {
		return nil, fmt.Errorf("count scenes: %w", err)
	}
	result.Scenes = scenes
	if !result.hasTargets() {
		return result, nil
	}
	for _, name := range config.Variants {
		hits, err := scoreVariant(ctx, database, rankers[name], targets, totals[name])
		if err != nil {
			return nil, fmt.Errorf("%s at %s: %w", name, cutoff.Format(time.DateOnly), err)
		}
		result.Hits[name] = hits
	}
	return result, nil
}

// scoreVariant ranks once per repeat and averages the scores per target kind.
func scoreVariant(ctx context.Context, database *db.DB, repeats []ranker, targets map[string]*targetSet,
	totals *variantTotals) (map[string]CutoffHits, error) {
	sum := map[string]CutoffHits{}
	for _, repeat := range repeats {
		ranking, err := repeat.rank(ctx, listDepth)
		if err != nil {
			return nil, err
		}
		if err := totals.add(ctx, database, ranking); err != nil {
			return nil, err
		}
		for kind, set := range targets {
			score := scoreList(ranking, set.reachable)
			entry := sum[kind]
			entry.HitsAt24 += float64(score.hitsSmall)
			entry.HitsAt96 += float64(score.hitsLarge)
			entry.ReciprocalRank += score.reciprocalRank
			entry.AnyAt24 += indicator(score.hitsSmall > 0)
			entry.AnyAt96 += indicator(score.hitsLarge > 0)
			sum[kind] = entry
		}
	}
	for kind, entry := range sum {
		count := float64(len(repeats))
		sum[kind] = CutoffHits{
			HitsAt24: entry.HitsAt24 / count, HitsAt96: entry.HitsAt96 / count,
			AnyAt24: entry.AnyAt24 / count, AnyAt96: entry.AnyAt96 / count, ReciprocalRank: entry.ReciprocalRank / count,
		}
	}
	return sum, nil
}

// add measures a ranking's coverage into the totals.
func (totals *variantTotals) add(ctx context.Context, database *db.DB, ranking []rankedScene) error {
	traits, err := loadTraits(ctx, database, topSceneIDs(ranking))
	if err != nil {
		return err
	}
	measured := measureCoverage(ranking, traits)
	totals.rankings++
	totals.sites += float64(measured.sites)
	totals.tags += float64(measured.tags)
	for source, count := range measured.sourceCount {
		totals.sourceShare[source] += float64(count) / float64(max(measured.size, 1))
	}
	return nil
}

// hasTargets reports whether anything reachable was wanted in the window.
func (result *CutoffReport) hasTargets() bool {
	for _, counts := range result.Targets {
		if counts.Reachable > 0 {
			return true
		}
	}
	return false
}

// dropEmptyCutoffs removes the cutoffs nothing reachable followed; they
// carry no information about the rankings.
func (report *Report) dropEmptyCutoffs() {
	kept := report.Cutoffs[:0]
	for _, cutoff := range report.Cutoffs {
		if cutoff.hasTargets() {
			kept = append(kept, cutoff)
		}
	}
	report.Cutoffs = kept
}

// summarize averages a variant's per-cutoff scores into a report.
func summarize(name string, cutoffs []CutoffReport, totals *variantTotals) VariantReport {
	variant := VariantReport{Name: name, SourceShare: map[string]float64{}}
	for _, kind := range targetKinds {
		variant.Kinds = append(variant.Kinds, summarizeKind(name, kind, cutoffs))
	}
	if totals.rankings > 0 {
		variant.SitesInTop24 = totals.sites / float64(totals.rankings)
		variant.TagsInTop24 = totals.tags / float64(totals.rankings)
		for source, share := range totals.sourceShare {
			variant.SourceShare[source] = share / float64(totals.rankings)
		}
	}
	return variant
}

// summarizeKind averages one target kind over the cutoffs that had reachable
// targets of it.
func summarizeKind(name, kind string, cutoffs []CutoffReport) KindReport {
	summary := KindReport{Kind: kind}
	for _, cutoff := range cutoffs {
		counts := cutoff.Targets[kind]
		summary.Unreachable += counts.Unreachable
		if counts.Reachable == 0 {
			continue
		}
		hits := cutoff.Hits[name][kind]
		summary.Cutoffs++
		summary.Reachable += counts.Reachable
		summary.HitRateAt24 += hits.AnyAt24
		summary.HitRateAt96 += hits.AnyAt96
		summary.RecallAt24 += hits.HitsAt24 / float64(counts.Reachable)
		summary.RecallAt96 += hits.HitsAt96 / float64(counts.Reachable)
		summary.MRR += hits.ReciprocalRank
	}
	if summary.Cutoffs == 0 {
		return summary
	}
	count := float64(summary.Cutoffs)
	summary.HitRateAt24 /= count
	summary.HitRateAt96 /= count
	summary.RecallAt24 /= count
	summary.RecallAt96 /= count
	summary.MRR /= count
	return summary
}

// indicator is 1 for true, 0 for false.
func indicator(condition bool) float64 {
	if condition {
		return 1
	}
	return 0
}
