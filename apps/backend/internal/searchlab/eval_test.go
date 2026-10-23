package searchlab

import (
	"math"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// TestNDCGIsOneForTheIdealOrder scores the ideal order 1 and a reversed one lower.
func TestNDCGIsOneForTheIdealOrder(t *testing.T) {
	ideal := []int{3, 2, 0}
	if got := ndcg([]int{3, 2, 0}, ideal, 10); math.Abs(got-1) > 1e-9 {
		t.Errorf("ideal order: %v", got)
	}
	if got := ndcg([]int{0, 2, 3}, ideal, 10); got >= 1 || got <= 0 {
		t.Errorf("reversed order: %v", got)
	}
	if got := ndcg([]int{0, 0}, []int{0, 0}, 10); got != 0 {
		t.Errorf("nothing relevant: %v", got)
	}
}

// TestSpearmanOfMonotoneSeries is 1 for agreeing and −1 for opposite orders.
func TestSpearmanOfMonotoneSeries(t *testing.T) {
	if got := pearson(ranks([]float64{1, 2, 3}), ranks([]float64{10, 20, 30})); math.Abs(got-1) > 1e-9 {
		t.Errorf("agreeing: %v", got)
	}
	if got := pearson(ranks([]float64{1, 2, 3}), ranks([]float64{3, 2, 1})); math.Abs(got+1) > 1e-9 {
		t.Errorf("opposite: %v", got)
	}
}

// TestLabelsRoundTrip writes grades to labels.tsv and reads them back,
// keeping unlabelled rows unlabelled.
func TestLabelsRoundTrip(t *testing.T) {
	directory := t.TempDir()
	snapshot := &Snapshot{Lists: []List{{QueryID: "q", Plugin: "site", Items: []plugins.DiscoveredItem{
		{Title: "first\twith tab", SourceURL: "https://site/1"},
		{Title: "second", SourceURL: "https://site/2"},
	}}}}
	labels := NewLabels()
	labels.Set("q", "https://site/1", GradePerfect)
	if err := SaveLabels(directory, snapshot, labels); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadLabels(directory)
	if err != nil {
		t.Fatal(err)
	}
	if grade, found := loaded.Grade("q", "https://site/1"); !found || grade != GradePerfect {
		t.Errorf("graded row: %v %v", grade, found)
	}
	if _, found := loaded.Grade("q", "https://site/2"); found {
		t.Error("unlabelled row came back graded")
	}
}

// TestTheoriesKeepEveryResult checks every theory ranks each result once and
// the text theory prefers matching titles.
func TestTheoriesKeepEveryResult(t *testing.T) {
	lists := []List{
		{Plugin: "a", Items: []plugins.DiscoveredItem{{Title: "blue", SourceURL: "a1"}, {Title: "red one", SourceURL: "a2"}}},
		{Plugin: "b", Items: []plugins.DiscoveredItem{{Title: "red two", SourceURL: "b1"}}},
	}
	query := Query{ID: "q", Text: "red"}
	for _, theory := range Theories {
		if got := len(theory.Rank(query, lists)); got != 3 {
			t.Errorf("%s ranked %d results, want 3", theory.Name, got)
		}
	}
	if first := rankText(query, lists)[0]; first.SourceURL == "a1" {
		t.Error("text theory put the non-matching title first")
	}
}
