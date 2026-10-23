// Package searchlab is a small, fixed search benchmark: a few queries run
// through real plugins, frozen in a snapshot, graded by hand, and ranking
// theories scored against the grades. It lets ideas about search and labelling
// be compared on the same data instead of argued.
package searchlab

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// File names inside a lab directory.
const (
	ConfigFile   = "queries.json"
	SnapshotFile = "snapshot.json"
	LabelsFile   = "labels.tsv"
)

// Query is one benchmark query.
type Query struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	// Note says what makes the query easy or hard.
	Note string `json:"note,omitempty"`
}

// Config is what a lab fetches: the queries, the plugins and how many results
// per plugin and query.
type Config struct {
	Plugins []string `json:"plugins"`
	Limit   int      `json:"limit"`
	Queries []Query  `json:"queries"`
}

// Snapshot is the frozen plugin output every theory is scored on.
type Snapshot struct {
	FetchedAt time.Time `json:"fetched_at"`
	Lists     []List    `json:"lists"`
}

// Kinds of lists: what the plugin was asked.
const (
	// KindKeyword is the plugin's keyword search for the query.
	KindKeyword = ""
	// KindRelated is the site's related list of a confident hit (see expand).
	KindRelated = "related"
)

// List is one plugin's results for one query, in the site's order.
type List struct {
	QueryID string                   `json:"query_id"`
	Plugin  string                   `json:"plugin"`
	Items   []plugins.DiscoveredItem `json:"items"`
	// Error is set when the plugin failed; Items is then empty.
	Error string `json:"error,omitempty"`
	// Kind is KindKeyword or KindRelated; a related list names its Seed (the
	// hit's page) and how confident the hit was.
	Kind           string  `json:"kind,omitempty"`
	Seed           string  `json:"seed,omitempty"`
	SeedTitle      string  `json:"seed_title,omitempty"`
	SeedConfidence float64 `json:"seed_confidence,omitempty"`
}

// keywordLists keeps the keyword search lists: what a theory that does not
// expand gets to rank.
func keywordLists(lists []List) []List {
	out := []List{}
	for _, list := range lists {
		if list.Kind == KindKeyword {
			out = append(out, list)
		}
	}
	return out
}

// LoadConfig reads queries.json from a lab directory.
func LoadConfig(directory string) (*Config, error) {
	config := &Config{}
	if err := readJSON(filepath.Join(directory, ConfigFile), config); err != nil {
		return nil, err
	}
	if config.Limit <= 0 {
		config.Limit = 20
	}
	return config, nil
}

// LoadSnapshot reads snapshot.json from a lab directory.
func LoadSnapshot(directory string) (*Snapshot, error) {
	snapshot := &Snapshot{}
	if err := readJSON(filepath.Join(directory, SnapshotFile), snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

// SaveSnapshot writes snapshot.json into a lab directory.
func SaveSnapshot(directory string, snapshot *Snapshot) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomically(filepath.Join(directory, SnapshotFile), data)
}

// Query returns the config's query with the id, or false.
func (config *Config) Query(id string) (Query, bool) {
	for _, query := range config.Queries {
		if query.ID == id {
			return query, true
		}
	}
	return Query{}, false
}

// ListsFor returns the snapshot's lists of one query, in config plugin order.
func (snapshot *Snapshot) ListsFor(queryID string) []List {
	out := []List{}
	for _, list := range snapshot.Lists {
		if list.QueryID == queryID {
			out = append(out, list)
		}
	}
	return out
}

// ItemKey identifies a result across the snapshot, labels and theories.
func ItemKey(item plugins.DiscoveredItem) string {
	return item.SourceURL
}

// readJSON decodes a JSON file into target.
func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// writeAtomically replaces path with data, so a crash never leaves half a file.
func writeAtomically(path string, data []byte) error {
	temporary := path + ".tmp"
	if err := os.WriteFile(temporary, data, 0o644); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}
