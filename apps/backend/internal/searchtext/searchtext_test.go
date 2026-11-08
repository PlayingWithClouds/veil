package searchtext

import (
	"slices"
	"testing"
)

func TestParse(t *testing.T) {
	query := Parse(`angela "Emma Smile" -anal -"step mom" emma_smile "open`)
	if want := []string{"angela", "emma", "smile"}; !slices.Equal(query.Words, want) {
		t.Errorf("words = %q, want %q", query.Words, want)
	}
	if want := []string{"emma smile", "open"}; !slices.Equal(query.Phrases, want) {
		t.Errorf("phrases = %q, want %q", query.Phrases, want)
	}
	if want := []string{"anal", "step mom"}; !slices.Equal(query.Excluded, want) {
		t.Errorf("excluded = %q, want %q", query.Excluded, want)
	}
	if want := "angela Emma Smile emma_smile open"; query.SiteText() != want {
		t.Errorf("site text = %q, want %q", query.SiteText(), want)
	}
}

func TestMatches(t *testing.T) {
	cases := []struct {
		query string
		text  string
		words bool
		want  bool
	}{
		{`"angela white"`, "Angela White's best scene", false, true},
		{`"angela white"`, "White Angela", false, false},
		{`"angela white"`, "Angela Whitehouse", false, false},
		{`"emma smile"`, "EMMA-SMILE solo", false, true},
		{`angela -anal`, "Angela White anal", false, false},
		{`angela -"step mom"`, "Angela and her step-mom", false, false},
		{`angela white`, "Lena Paul", false, true},
		{`angela white`, "Lena Paul", true, false},
		{`mal`, "Mia Malkova", true, true},
		{`kova`, "Mia Malkova", true, false},
		{`Mía`, "mia malkova", true, true},
	}
	for _, test := range cases {
		got := Parse(test.query).Matches(test.text, test.words)
		if got != test.want {
			t.Errorf("Parse(%q).Matches(%q, %v) = %v, want %v", test.query, test.text, test.words, got, test.want)
		}
	}
}
