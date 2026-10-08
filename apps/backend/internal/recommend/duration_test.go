package recommend

import "testing"

// TestDurationRangeContains checks open and half-open bounds and unknown runtimes.
func TestDurationRangeContains(t *testing.T) {
	cases := []struct {
		name            string
		window          DurationRange
		durationSeconds int
		want            bool
	}{
		{"open keeps unknown", DurationRange{}, 0, true},
		{"bounded drops unknown", DurationRange{MaxSeconds: 600}, 0, false},
		{"under max", DurationRange{MaxSeconds: 600}, 599, true},
		{"over max", DurationRange{MaxSeconds: 600}, 601, false},
		{"below min", DurationRange{MinSeconds: 1800}, 1799, false},
		{"above min", DurationRange{MinSeconds: 1800}, 3000, true},
		{"inside window", DurationRange{MinSeconds: 600, MaxSeconds: 1800}, 900, true},
	}
	for _, testCase := range cases {
		if got := testCase.window.contains(testCase.durationSeconds); got != testCase.want {
			t.Errorf("%s: contains(%d) = %v, want %v", testCase.name, testCase.durationSeconds, got, testCase.want)
		}
	}
}
