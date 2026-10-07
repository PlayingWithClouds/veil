package alike

import "testing"

func TestScorePerformersOutweighTitle(t *testing.T) {
	target := Target{
		Performers: []string{"Riley Reid", "Manuel Ferrara"},
		Title:      "Amazing Passionate Encounter",
	}
	// Same performers, an entirely different (SEO-spam) title.
	sameScene := Candidate{
		Performers: []string{"Manuel Ferrara", "Riley Reid"},
		Title:      "HOT Teen Fucked Hard 1080p FREE PORN",
	}
	// A different scene that happens to echo the target's title words.
	otherScene := Candidate{
		Performers: []string{"Someone Else"},
		Title:      "Amazing Passionate Encounter",
	}

	same := Score(target, sameScene)
	other := Score(target, otherScene)

	if same <= other {
		t.Fatalf("matching performers with a different title (%.3f) should outrank a title-only match (%.3f)", same, other)
	}
	if same < 0.4 {
		t.Fatalf("same-performer match scored too low: %.3f", same)
	}
}

func TestScoreDurationPrefersFullLength(t *testing.T) {
	target := Target{
		Performers:      []string{"Riley Reid"},
		DurationSeconds: 120, // current scene is a short teaser
	}
	fullLength := Candidate{
		Performers:      []string{"Riley Reid"},
		DurationSeconds: 1800,
	}
	anotherTeaser := Candidate{
		Performers:      []string{"Riley Reid"},
		DurationSeconds: 60,
	}

	full := Score(target, fullLength)
	teaser := Score(target, anotherTeaser)

	if full <= teaser {
		t.Fatalf("full-length copy (%.3f) should outrank a shorter teaser (%.3f)", full, teaser)
	}
}

func TestScoreUnrelatedIsLow(t *testing.T) {
	target := Target{
		Performers:      []string{"Riley Reid"},
		DurationSeconds: 1200,
		Title:           "Sunny Afternoon",
	}
	unrelated := Candidate{
		Performers:      []string{"Nobody Here"},
		DurationSeconds: 1200,
		Title:           "Completely Different Thing",
	}

	if score := Score(target, unrelated); score > 0.4 {
		t.Fatalf("unrelated candidate scored too high: %.3f", score)
	}
}

func TestScorePosterLiftsTitleMismatch(t *testing.T) {
	// Stage-1 conditions: no performers/duration yet, titles differ. A matching
	// poster hash must still lift the score above a non-matching poster.
	target := Target{Title: "Some Title", PosterHash: 0xFF00FF00FF00FF00}
	sameFrame := Candidate{Title: "Totally Other Words", PosterHash: 0xFF00FF00FF00FF00}
	otherFrame := Candidate{Title: "Totally Other Words", PosterHash: 0x00FF00FF00FF00FF}

	if Score(target, sameFrame) <= Score(target, otherFrame) {
		t.Fatal("matching poster hash should raise the score")
	}
}

func TestScoreEmptyIsZero(t *testing.T) {
	if score := Score(Target{}, Candidate{}); score != 0 {
		t.Fatalf("no shared signals should score 0, got %.3f", score)
	}
}

func TestDurationScore(t *testing.T) {
	if durationScore(0, 100) != 0 {
		t.Fatal("unknown target duration must be neutral")
	}
	if durationScore(100, 0) != 0 {
		t.Fatal("unknown candidate duration must be neutral")
	}
	if durationScore(600, 600) != 1 {
		t.Fatal("equal duration must score 1")
	}
	if durationScore(120, 1800) != 1 {
		t.Fatal("a longer full-length candidate must score 1")
	}
	if score := durationScore(600, 300); score >= 1 || score <= 0 {
		t.Fatalf("a half-length candidate must score between 0 and 1, got %.3f", score)
	}
}
