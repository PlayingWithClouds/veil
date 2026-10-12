package names

import (
	"slices"
	"testing"
)

// TestFinderFind checks matching of names inside titles.
func TestFinderFind(t *testing.T) {
	finder := NewFinder()
	finder.Add("performer:isla", "Isla Summer")
	finder.Add("performer:cassie", "Cassie Del Isla")
	finder.Add("performer:del", "Del Isla")
	finder.Add("performer:mia", "Mia")
	finder.Add("performer:zoe", "Zoe Bloom")
	finder.Add("performer:zoe", "Zoey Bloom")

	cases := []struct {
		title string
		want  []string
	}{
		{"Isla Summer's wet pussy", []string{"performer:isla"}},
		{"[Full HD] - ZOE BLOOM - Shy But Honestly Direct.", []string{"performer:zoe"}},
		{"Zoey-Bloom and isla summer", []string{"performer:zoe", "performer:isla"}},
		{"Cassie Del Isla Anal", []string{"performer:cassie"}},
		{"Mia is a single word, never matched", []string{}},
		{"Isla alone and Summer apart", []string{}},
		{"Zoe Bloom Zoe Bloom twice", []string{"performer:zoe"}},
		{"", []string{}},
	}
	for _, testCase := range cases {
		if got := finder.Find(testCase.title); !slices.Equal(got, testCase.want) {
			t.Errorf("Find(%q) = %v, want %v", testCase.title, got, testCase.want)
		}
	}
}
