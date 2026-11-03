package pipeline

import "testing"

// TestPluginWanted checks that an empty list wants every plugin and a
// non-empty one only those it names.
func TestPluginWanted(t *testing.T) {
	cases := []struct {
		only   []string
		plugin string
		want   bool
	}{
		{nil, "iafd", true},
		{[]string{}, "iafd", true},
		{[]string{"iafd"}, "iafd", true},
		{[]string{"iafd", "babepedia"}, "babepedia", true},
		{[]string{"iafd"}, "babepedia", false},
	}
	for _, testCase := range cases {
		if got := pluginWanted(testCase.only, testCase.plugin); got != testCase.want {
			t.Errorf("pluginWanted(%v, %q) = %v, want %v", testCase.only, testCase.plugin, got, testCase.want)
		}
	}
}
