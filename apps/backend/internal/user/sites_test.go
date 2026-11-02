package user

import "testing"

func TestScorePreferences(t *testing.T) {
	counts := []SitePreference{
		{Plugin: "rarely", Shown: 400, Chosen: 4},
		{Plugin: "often", Shown: 400, Chosen: 40},
		{Plugin: "once", Shown: 2, Chosen: 1},
	}
	scorePreferences(counts)
	if counts[0].Plugin != "often" {
		t.Fatalf("most picked site first, got %+v", counts)
	}
	if counts[len(counts)-1].Plugin != "rarely" {
		t.Fatalf("least picked site last, got %+v", counts)
	}
	for _, count := range counts {
		if count.Plugin == "once" && count.Score > counts[0].Score {
			t.Fatalf("a site shown twice outranks one picked often: %+v", counts)
		}
	}
}

func TestScorePreferencesWithoutPicks(t *testing.T) {
	counts := []SitePreference{{Plugin: "shown", Shown: 10}}
	scorePreferences(counts)
	if counts[0].Score != 1 {
		t.Fatalf("no picks yet scores neutral, got %v", counts[0].Score)
	}
}
