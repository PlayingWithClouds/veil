package alike

import "testing"

// TestSamePerson merges a shared name only with a tie (birthdate, a shared
// video, the same picture) and never across stated differences.
func TestSamePerson(t *testing.T) {
	picture := uint64(0xABCDEF0123456789)
	cases := []struct {
		name          string
		first, second PersonFacts
		shared        int
		want          bool
	}{
		{"name alone", PersonFacts{}, PersonFacts{}, 0, false},
		{"same birthdate", PersonFacts{Birthdate: "1995-04-12"}, PersonFacts{Birthdate: "1995-04-12T00:00:00Z"}, 0, true},
		{"other birthdate", PersonFacts{Birthdate: "1995-04-12"}, PersonFacts{Birthdate: "1990-01-01"}, 3, false},
		{"shared video", PersonFacts{}, PersonFacts{Birthdate: "1995-04-12"}, 1, true},
		{"other gender", PersonFacts{Gender: "female"}, PersonFacts{Gender: "male"}, 1, false},
		{"same picture", PersonFacts{PictureHash: picture}, PersonFacts{PictureHash: picture ^ 1}, 0, true},
	}
	for _, testCase := range cases {
		if got := SamePerson(testCase.first, testCase.second, testCase.shared); got != testCase.want {
			t.Errorf("%s: got %v, want %v", testCase.name, got, testCase.want)
		}
	}
}

// TestSameStudio ties studios by host, network or logo, and keeps different
// hosts apart.
func TestSameStudio(t *testing.T) {
	if !SameStudio(StudioFacts{URL: "https://www.brazzers.com/"}, StudioFacts{URL: "http://brazzers.com/videos"}) {
		t.Error("same host not merged")
	}
	if SameStudio(StudioFacts{URL: "https://a.com", Parent: "n"}, StudioFacts{URL: "https://b.com", Parent: "n"}) {
		t.Error("different hosts merged")
	}
	if !SameStudio(StudioFacts{Parent: "network"}, StudioFacts{Parent: "network"}) {
		t.Error("same network not merged")
	}
	if SameStudio(StudioFacts{}, StudioFacts{}) {
		t.Error("name alone merged")
	}
}
