package alike

import "testing"

func TestSameScene(t *testing.T) {
	cases := []struct {
		name   string
		first  Candidate
		second Candidate
		want   bool
	}{
		{"same performers and runtime", Candidate{Performers: []string{"Ana Lee"}, Title: "x", DurationSeconds: 1200}, Candidate{Performers: []string{"ana  lee"}, Title: "y", DurationSeconds: 1210}, true},
		{"same title and runtime without performers", Candidate{Title: "Pool Party Fun", DurationSeconds: 900}, Candidate{Title: "pool party fun!", DurationSeconds: 905}, true},
		{"runtime far apart", Candidate{Performers: []string{"Ana Lee"}, DurationSeconds: 1200}, Candidate{Performers: []string{"Ana Lee"}, DurationSeconds: 600}, false},
		{"runtime unknown", Candidate{Performers: []string{"Ana Lee"}}, Candidate{Performers: []string{"Ana Lee"}, DurationSeconds: 600}, false},
		{"different performers and titles", Candidate{Performers: []string{"Ana Lee"}, Title: "One", DurationSeconds: 600}, Candidate{Performers: []string{"Bo Kim"}, Title: "Two", DurationSeconds: 600}, false},
		{"partial performer overlap", Candidate{Performers: []string{"Ana Lee", "Bo Kim"}, Title: "One", DurationSeconds: 600}, Candidate{Performers: []string{"Ana Lee"}, Title: "Two", DurationSeconds: 600}, false},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := SameScene(testCase.first, testCase.second); got != testCase.want {
				t.Errorf("SameScene = %v, want %v", got, testCase.want)
			}
			if got := SameScene(testCase.second, testCase.first); got != testCase.want {
				t.Errorf("SameScene is not symmetric: reversed = %v, want %v", got, testCase.want)
			}
		})
	}
}
