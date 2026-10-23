// Command search-lab runs the search benchmark in a lab directory (default
// apps/backend/searchlab): fetch freezes real plugin results for the queries,
// label serves a page to grade them by hand, eval scores every ranking theory
// against the grades.
//
//	go run ./cmd/search-lab fetch [-plugins DIR]
//	go run ./cmd/search-lab label [-addr 127.0.0.1:8099]
//	go run ./cmd/search-lab eval [-json]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/playingwithclouds/veil/internal/netdns"
	"github.com/playingwithclouds/veil/internal/searchlab"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	command, arguments := os.Args[1], os.Args[2:]
	flags := flag.NewFlagSet(command, flag.ExitOnError)
	directory := flags.String("dir", "searchlab", "lab directory (queries.json, snapshot.json, labels.tsv)")
	pluginDir := flags.String("plugins", "data/plugins", "installed plugin bundles (fetch)")
	address := flags.String("addr", "127.0.0.1:8099", "address of the labelling page (label)")
	asJSON := flags.Bool("json", false, "print the report as JSON (eval)")
	flags.Parse(arguments)

	config, err := searchlab.LoadConfig(*directory)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	switch command {
	case "fetch":
		runFetch(*directory, *pluginDir, config)
	case "label":
		runLabel(*directory, *address, config)
	case "eval":
		runEval(*directory, config, *asJSON)
	default:
		usage()
	}
}

// runFetch freezes fresh plugin results and lists them in labels.tsv,
// keeping the grades of results that are still there.
func runFetch(directory, pluginDir string, config *searchlab.Config) {
	netdns.Configure(os.Getenv("DNS_SERVERS"), os.Getenv("DNS_SERVERS_FILE"))
	solver := strings.TrimSpace(os.Getenv("FLARESOLVERR_URL")) != ""
	if !solver {
		log.Print("FLARESOLVERR_URL is not set: plugins behind Cloudflare (missav) will fail")
	}
	snapshot, err := searchlab.Fetch(context.Background(), config, pluginDir, solver)
	if err != nil {
		log.Fatalf("fetch: %v", err)
	}
	labels, err := searchlab.LoadLabels(directory)
	if err != nil {
		log.Fatalf("labels: %v", err)
	}
	if err := searchlab.SaveSnapshot(directory, snapshot); err != nil {
		log.Fatalf("save snapshot: %v", err)
	}
	if err := searchlab.SaveLabels(directory, snapshot, labels); err != nil {
		log.Fatalf("save labels: %v", err)
	}
}

// runLabel serves the labelling page.
func runLabel(directory, address string, config *searchlab.Config) {
	snapshot, labels := loadSnapshotAndLabels(directory)
	log.Fatal(searchlab.Serve(address, directory, config, snapshot, labels))
}

// runEval prints the evaluation.
func runEval(directory string, config *searchlab.Config, asJSON bool) {
	snapshot, labels := loadSnapshotAndLabels(directory)
	report := searchlab.Evaluate(config, snapshot, labels)
	if !asJSON {
		report.Print(os.Stdout, config)
		return
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(report)
}

// loadSnapshotAndLabels reads the frozen results and the grades.
func loadSnapshotAndLabels(directory string) (*searchlab.Snapshot, *searchlab.Labels) {
	snapshot, err := searchlab.LoadSnapshot(directory)
	if err != nil {
		log.Fatalf("snapshot (run fetch first): %v", err)
	}
	labels, err := searchlab.LoadLabels(directory)
	if err != nil {
		log.Fatalf("labels: %v", err)
	}
	return snapshot, labels
}

// usage explains the commands and exits.
func usage() {
	fmt.Fprintln(os.Stderr, "usage: search-lab fetch|label|eval [-dir searchlab] [-plugins data/plugins] [-addr 127.0.0.1:8099] [-json]")
	os.Exit(2)
}
