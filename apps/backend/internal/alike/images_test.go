package alike

import "testing"

// TestSameSceneByImages merges on a shared cover or stills with near-equal
// runtimes, whatever the titles say, but not across runtimes, part numbers
// or casts, nor on a placeholder cover.
func TestSameSceneByImages(t *testing.T) {
	cover := uint64(0xF0F0F0F0F0F0F0F0)
	reencoded := cover ^ 0b111 // three bits off
	other := uint64(0x0F0F0F0F0F0F0F0F)
	frames := []uint64{0x1111111111111111, 0x2222222222222222, 0x4444444444444444}
	cases := []struct {
		name          string
		first, second Candidate
		firstPics     Pictures
		secondPics    Pictures
		coverUsable   bool
		want          bool
	}{
		{"same cover, reworded title", Candidate{Title: "Hot blonde rides", DurationSeconds: 1800}, Candidate{Title: "Blondine reitet", DurationSeconds: 1805}, Pictures{Cover: cover}, Pictures{Cover: reencoded}, true, true},
		{"placeholder cover", Candidate{Title: "A", DurationSeconds: 1800}, Candidate{Title: "B", DurationSeconds: 1800}, Pictures{Cover: cover}, Pictures{Cover: cover}, false, false},
		{"same stills", Candidate{Title: "A", DurationSeconds: 900}, Candidate{Title: "B", DurationSeconds: 900}, Pictures{Cover: cover, Frames: frames}, Pictures{Cover: other, Frames: frames[:2]}, true, true},
		{"one still is not enough", Candidate{Title: "A", DurationSeconds: 900}, Candidate{Title: "B", DurationSeconds: 900}, Pictures{Frames: frames[:1]}, Pictures{Frames: frames[:1]}, true, false},
		{"longer cut", Candidate{Title: "A", DurationSeconds: 1200}, Candidate{Title: "A", DurationSeconds: 1800}, Pictures{Cover: cover}, Pictures{Cover: cover}, true, false},
		{"other part", Candidate{Title: "Series part 4", DurationSeconds: 900}, Candidate{Title: "Series part 5", DurationSeconds: 900}, Pictures{Cover: cover}, Pictures{Cover: cover}, true, false},
		{"other cast", Candidate{Title: "A", DurationSeconds: 900, Performers: []string{"Anna"}}, Candidate{Title: "A", DurationSeconds: 900, Performers: []string{"Bea"}}, Pictures{Cover: cover}, Pictures{Cover: cover}, true, false},
		{"different covers", Candidate{Title: "A", DurationSeconds: 900}, Candidate{Title: "A", DurationSeconds: 900}, Pictures{Cover: cover}, Pictures{Cover: other}, true, false},
	}
	for _, testCase := range cases {
		got := SameSceneByImages(testCase.first, testCase.second, testCase.firstPics, testCase.secondPics, testCase.coverUsable)
		if got != testCase.want {
			t.Errorf("%s: got %v, want %v", testCase.name, got, testCase.want)
		}
	}
}
