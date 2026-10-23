package recommendeval

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"
)

// Report is the outcome of one evaluation run.
type Report struct {
	Database   string          `json:"database"`
	WindowDays float64         `json:"windowDays"`
	StepDays   float64         `json:"stepDays"`
	Repeats    int             `json:"repeats"`
	Cutoffs    []CutoffReport  `json:"cutoffs"`
	Variants   []VariantReport `json:"variants"`
}

// CutoffReport is what was available and wanted at one cutoff.
type CutoffReport struct {
	Cutoff time.Time `json:"cutoff"`
	// Scenes is how many scenes the library held at the cutoff.
	Scenes  int                     `json:"scenes"`
	Targets map[string]TargetCounts `json:"targets"`
	// Hits is, per variant and target kind, the targets found in the top
	// SmallCutoff and LargeCutoff (averaged over repeats) and the
	// reciprocal rank.
	Hits map[string]map[string]CutoffHits `json:"hits"`
}

// TargetCounts is how many scenes the user turned to in the window.
type TargetCounts struct {
	Reachable   int `json:"reachable"`
	Unreachable int `json:"unreachable"`
}

// CutoffHits is one ranking's score at one cutoff for one target kind.
type CutoffHits struct {
	HitsAt24 float64 `json:"hitsAt24"`
	HitsAt96 float64 `json:"hitsAt96"`
	// AnyAt24 and AnyAt96 are the share of runs that found at least one
	// target.
	AnyAt24        float64 `json:"anyAt24"`
	AnyAt96        float64 `json:"anyAt96"`
	ReciprocalRank float64 `json:"reciprocalRank"`
}

// VariantReport averages one variant over the cutoffs.
type VariantReport struct {
	Name  string       `json:"name"`
	Kinds []KindReport `json:"kinds"`
	// SitesInTop24 and TagsInTop24 are the mean distinct counts in the first
	// 24 scenes.
	SitesInTop24 float64 `json:"sitesInTop24"`
	TagsInTop24  float64 `json:"tagsInTop24"`
	// SourceShare is each candidate source's share of the first 24 positions.
	SourceShare map[string]float64 `json:"sourceShare"`
}

// KindReport averages one target kind over the cutoffs that had targets of
// it. HitRate is the share of those cutoffs with at least one target in the
// top, Recall the mean share of a cutoff's targets found there.
type KindReport struct {
	Kind        string  `json:"kind"`
	Cutoffs     int     `json:"cutoffs"`
	Reachable   int     `json:"reachable"`
	Unreachable int     `json:"unreachable"`
	HitRateAt24 float64 `json:"hitRateAt24"`
	HitRateAt96 float64 `json:"hitRateAt96"`
	RecallAt24  float64 `json:"recallAt24"`
	RecallAt96  float64 `json:"recallAt96"`
	MRR         float64 `json:"mrr"`
}

// WriteJSON prints the report as indented JSON.
func (report *Report) WriteJSON(writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// WriteTable prints the report as aligned tables.
func (report *Report) WriteTable(writer io.Writer) error {
	fmt.Fprintf(writer, "database %s\n%d cutoffs every %.1f days, %.1f day window, %d run(s) per cutoff\n\n",
		report.Database, len(report.Cutoffs), report.StepDays, report.WindowDays, report.Repeats)
	if err := report.writeCutoffs(writer); err != nil {
		return err
	}
	for _, variant := range report.Variants {
		if err := variant.writeTable(writer); err != nil {
			return err
		}
	}
	return nil
}

// writeCutoffs prints the cutoffs with what the user turned to after each.
func (report *Report) writeCutoffs(writer io.Writer) error {
	table := tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	fmt.Fprintf(table, "cutoff\tscenes")
	for _, kind := range targetKinds {
		fmt.Fprintf(table, "\t%s (reachable/all)", kind)
	}
	fmt.Fprintln(table)
	for _, cutoff := range report.Cutoffs {
		fmt.Fprintf(table, "%s\t%d", cutoff.Cutoff.Format("2006-01-02 15:04"), cutoff.Scenes)
		for _, kind := range targetKinds {
			counts := cutoff.Targets[kind]
			fmt.Fprintf(table, "\t%d/%d", counts.Reachable, counts.Reachable+counts.Unreachable)
		}
		fmt.Fprintln(table)
	}
	fmt.Fprintln(table)
	return table.Flush()
}

// writeTable prints one variant's metrics.
func (variant VariantReport) writeTable(writer io.Writer) error {
	fmt.Fprintf(writer, "== %s ==\n", variant.Name)
	table := tabwriter.NewWriter(writer, 0, 4, 2, ' ', 0)
	fmt.Fprintln(table, "target\tcutoffs\ttargets\thit@24\thit@96\trecall@24\trecall@96\tMRR")
	for _, kind := range variant.Kinds {
		fmt.Fprintf(table, "%s\t%d\t%d (+%d unreachable)\t%.2f\t%.2f\t%.3f\t%.3f\t%.4f\n",
			kind.Kind, kind.Cutoffs, kind.Reachable, kind.Unreachable,
			kind.HitRateAt24, kind.HitRateAt96, kind.RecallAt24, kind.RecallAt96, kind.MRR)
	}
	if err := table.Flush(); err != nil {
		return err
	}
	fmt.Fprintf(writer, "coverage of the top 24: %.1f sites, %.1f tags\n", variant.SitesInTop24, variant.TagsInTop24)
	if len(variant.SourceShare) > 0 {
		fmt.Fprintf(writer, "sources in the top 24: %s\n", formatShares(variant.SourceShare))
	}
	fmt.Fprintln(writer)
	return nil
}

// formatShares renders source shares, largest first.
func formatShares(shares map[string]float64) string {
	names := sortedKeys(shares)
	for left := range names {
		for right := left + 1; right < len(names); right++ {
			if shares[names[right]] > shares[names[left]] {
				names[left], names[right] = names[right], names[left]
			}
		}
	}
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, fmt.Sprintf("%s %.0f%%", name, 100*shares[name]))
	}
	return strings.Join(parts, ", ")
}
