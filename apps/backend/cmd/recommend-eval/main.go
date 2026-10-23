// Command recommend-eval replays the Home recommendations over a veil.db
// history and prints how well each cutoff's ranking predicted what the user
// did next. It works on a temporary copy; the database is never modified.
//
//	recommend-eval -db apps/backend/data/veil.db
//	recommend-eval -db veil.db -variants default,no-learning,baseline-newest -json
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/recommendeval"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "recommend-eval:", err)
		os.Exit(1)
	}
}

// run parses the flags, evaluates and prints the report.
func run() error {
	databasePath := flag.String("db", "", "path to the veil.db to evaluate (read, never modified)")
	windowDays := flag.Float64("window-days", 7, "days after a cutoff in which the user's actions count")
	stepDays := flag.Float64("step-days", 2, "days between cutoffs")
	warmupDays := flag.Float64("warmup-days", 3, "history kept before the first cutoff, from the first recorded activity")
	fromText := flag.String("from", "", "first cutoff (RFC 3339 or YYYY-MM-DD); default: first activity + warmup")
	toText := flag.String("to", "", "last cutoff (RFC 3339 or YYYY-MM-DD); default: last activity - window")
	variants := flag.String("variants", "default,no-learning,baseline-newest,baseline-random",
		"comma-separated rankings: default, no-learning (hand-set priors and quotas), baseline-newest, baseline-random")
	repeats := flag.Int("repeats", 1, "rankings per cutoff to average (the engine samples)")
	seed := flag.Int64("seed", 1, "seed of the engine's sampling")
	asJSON := flag.Bool("json", false, "print JSON instead of tables")
	verbose := flag.Bool("verbose", false, "show the engine's log lines")
	flag.Parse()

	if !*verbose {
		log.SetOutput(io.Discard)
	}
	from, err := parseTime(*fromText)
	if err != nil {
		return fmt.Errorf("-from: %w", err)
	}
	to, err := parseTime(*toText)
	if err != nil {
		return fmt.Errorf("-to: %w", err)
	}
	report, err := recommendeval.Run(context.Background(), recommendeval.Config{
		DatabasePath: *databasePath,
		Window:       days(*windowDays),
		Step:         days(*stepDays),
		Warmup:       days(*warmupDays),
		From:         from,
		To:           to,
		Variants:     strings.Split(*variants, ","),
		Repeats:      *repeats,
		Seed:         *seed,
	})
	if err != nil {
		return err
	}
	if *asJSON {
		return report.WriteJSON(os.Stdout)
	}
	return report.WriteTable(os.Stdout)
}

// days converts a number of days to a duration.
func days(count float64) time.Duration {
	return time.Duration(count * float64(24*time.Hour))
}

// parseTime reads an RFC 3339 time or a date; the zero time when text is empty.
func parseTime(text string) (time.Time, error) {
	if text == "" {
		return time.Time{}, nil
	}
	if parsed, err := time.Parse(time.RFC3339, text); err == nil {
		return parsed, nil
	}
	return time.Parse(time.DateOnly, text)
}
