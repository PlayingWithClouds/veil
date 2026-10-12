package names

import "testing"

// TestKey checks that spellings of one name share a key and different names don't.
func TestKey(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"Mia Malkova", "mia malkova"},
		{"MIA  MALKOVA", "mia malkova"},
		{"mia-malkova", "mia malkova"},
		{"  Mia.Malkova ", "mia malkova"},
		{"Mía Malkova", "mia malkova"},
		{"Isla Summer's", "isla summers"},
		{"Isla Summer’s", "isla summers"},
		{"O'Reilly", "oreilly"},
		{"Zoe Bloom 2", "zoe bloom 2"},
		{"???", ""},
		{"", ""},
	}
	for _, testCase := range cases {
		if got := Key(testCase.name); got != testCase.want {
			t.Errorf("Key(%q) = %q, want %q", testCase.name, got, testCase.want)
		}
	}
}
