package graphql

import (
	"testing"

	"github.com/playingwithclouds/veil/internal/updater"
)

func TestSiteFeedShort(t *testing.T) {
	resolver := &Resolver{feeder: &updater.Feeder{}}
	cases := []struct {
		name      string
		sources   []string
		offset    int
		refresh   bool
		itemCount int
		want      bool
	}{
		{"refresh that didn't fill the page", []string{"eporner"}, 0, true, 5, true},
		{"refresh with a full page", []string{"eporner"}, 0, true, 24, false},
		{"plain load, empty", []string{"eporner"}, 0, false, 0, true},
		{"plain load, some", []string{"eporner"}, 0, false, 5, false},
		{"later page", []string{"eporner"}, 24, true, 0, false},
		{"no site filter", nil, 0, true, 0, false},
	}
	for _, testCase := range cases {
		got := resolver.siteFeedShort(testCase.sources, testCase.offset, testCase.refresh, testCase.itemCount, 24)
		if got != testCase.want {
			t.Errorf("%s: got %v, want %v", testCase.name, got, testCase.want)
		}
	}
	if (&Resolver{}).siteFeedShort([]string{"eporner"}, 0, true, 0, 24) {
		t.Error("short without a feeder")
	}
}
