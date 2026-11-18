package alike

import "testing"

// TestFullLengthOf finds the long version of a clip by its title words and
// cast, and refuses other parts, other casts, clips and same-length copies.
func TestFullLengthOf(t *testing.T) {
	clip := Candidate{Title: "Anna Bell hot pool party teaser", DurationSeconds: 120, Performers: []string{"Anna Bell"}}
	cases := []struct {
		name      string
		candidate Candidate
		want      bool
	}{
		{"full scene", Candidate{Title: "Anna Bell hot pool party full scene", DurationSeconds: 2400, Performers: []string{"Anna Bell", "Max"}}, true},
		{"uncredited full scene", Candidate{Title: "Hot pool party with Anna Bell", DurationSeconds: 1900}, true},
		{"another clip", Candidate{Title: "Anna Bell hot pool party", DurationSeconds: 300}, false},
		{"other cast", Candidate{Title: "Anna Bell hot pool party", DurationSeconds: 2400, Performers: []string{"Bea"}}, false},
		{"other title", Candidate{Title: "Office overtime", DurationSeconds: 2400, Performers: []string{"Anna Bell"}}, false},
	}
	for _, testCase := range cases {
		if got := FullLengthOf(clip, testCase.candidate); got != testCase.want {
			t.Errorf("%s: got %v, want %v", testCase.name, got, testCase.want)
		}
	}
	numbered := Candidate{Title: "Pool party part 2 trailer", DurationSeconds: 90}
	if FullLengthOf(numbered, Candidate{Title: "Pool party part 3", DurationSeconds: 2400}) {
		t.Error("another part counted as the full length")
	}
}
