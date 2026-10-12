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
		{"site noise in brackets", Candidate{Title: "Flat Teen Zoe Bloom Masturbates And Cums Twice", DurationSeconds: 496}, Candidate{Title: "Flat teen Zoe Bloom Masturbates And Cums Twice (18+)", DurationSeconds: 501}, true},
		{"resolution tag in brackets", Candidate{Title: "Brandi Love And Zoe Bloom Mother Lover Society", DurationSeconds: 1492}, Candidate{Title: "Brandi Love And Zoe Bloom Mother Lover Society (1080)", DurationSeconds: 1500}, true},
		{"reworded title, same second", Candidate{Title: "Isla Summer's wet pussy is waiting for your dick", DurationSeconds: 620}, Candidate{Title: "Isla Summers wet pussy is waiting for your cock", DurationSeconds: 621}, true},
		{"cast list appended, same second", Candidate{Title: "Melissa needs more pussies", DurationSeconds: 4153}, Candidate{Title: "Melissa needs more pussies (Jenna Sativa, Melissa Moore)", DurationSeconds: 4153}, true},
		{"performer prefix and punctuation", Candidate{Title: "shy but honestly direct", DurationSeconds: 1881}, Candidate{Title: "[Full HD] - Zoe Bloom - Shy But Honestly Direct.", DurationSeconds: 1881}, true},
		{"reworded title, runtime not to the second", Candidate{Title: "Isla Summer's wet pussy is waiting for your dick", DurationSeconds: 620}, Candidate{Title: "Isla Summers wet pussy is waiting for your cock", DurationSeconds: 640}, false},
		{"series parts with different numbers", Candidate{Title: "Cumworld Ultimate Cumshot Compilation 14", DurationSeconds: 2215}, Candidate{Title: "The Ultimate Cumshot Compilation #15", DurationSeconds: 2215}, false},
		{"same part number", Candidate{Title: "Cumworld Ultimate Cumshot Compilation 14", DurationSeconds: 2215}, Candidate{Title: "The Ultimate Cumshot Compilation #14", DurationSeconds: 2215}, true},
		{"different ids", Candidate{Title: "FC2PPV 3635198 [Uncensored]", DurationSeconds: 4200}, Candidate{Title: "FC2PPV 4657517 [Uncensored]", DurationSeconds: 4200}, false},
		{"short generic titles, same second", Candidate{Title: "Cumshot compilation", DurationSeconds: 546}, Candidate{Title: "Cumshot Compilation - Simony Diamond", DurationSeconds: 546}, false},
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
