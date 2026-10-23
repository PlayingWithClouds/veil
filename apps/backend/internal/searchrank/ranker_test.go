package searchrank

import (
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// sceneHit builds a scene hit of a site list.
func sceneHit(source, title, url string, durationSeconds int) Hit {
	return Hit{Source: source, Item: plugins.DiscoveredItem{
		Title: title, MediaType: plugins.MediaTypeScene, SourceURL: url, DurationSeconds: durationSeconds,
	}}
}

// TestTextMatchOrdersByRelevance checks the order of a full title start,
// a late phrase, a performer-only match, a partial match and no match.
func TestTextMatchOrdersByRelevance(t *testing.T) {
	words := QueryWords("mia malkova")
	startsWith := TextMatch(words, TextFields{Title: "Mia Malkova in the garden"})
	inside := TextMatch(words, TextFields{Title: "The garden with Mia Malkova"})
	performerOnly := TextMatch(words, TextFields{Title: "Garden party", Performers: "Mia Malkova"})
	partial := TextMatch(words, TextFields{Title: "Mia in the garden"})
	none := TextMatch(words, TextFields{Title: "Garden party"})

	ordered := []float64{startsWith, inside, performerOnly, partial, none}
	for index := 1; index < len(ordered); index++ {
		if ordered[index-1] <= ordered[index] {
			t.Fatalf("scores not descending: %v", ordered)
		}
	}
	if startsWith != 1 || none != 0 {
		t.Fatalf("bounds: startsWith %v, none %v", startsWith, none)
	}
}

// TestTextMatchTreatsLastWordAsPrefix checks a query still being typed.
func TestTextMatchTreatsLastWordAsPrefix(t *testing.T) {
	typing := TextMatch(QueryWords("garden par"), TextFields{Title: "Garden party"})
	complete := TextMatch(QueryWords("garden party"), TextFields{Title: "Garden party"})
	if typing < 0.9 || typing >= complete {
		t.Fatalf("typing = %v, complete = %v", typing, complete)
	}
}

// TestSiteRankOrdersWithinAList checks that equally relevant results keep the
// order their site gave them, and that later pages score below earlier ones.
func TestSiteRankOrdersWithinAList(t *testing.T) {
	ranker := New("garden", 0)
	first := ranker.Add(sceneHit("alpha", "Garden one", "https://a/1", 600))
	second := ranker.Add(sceneHit("alpha", "Garden two", "https://a/2", 700))
	if first.Score <= second.Score {
		t.Fatalf("first %v should beat second %v", first.Score, second.Score)
	}

	secondPage := New("garden", 20).Add(sceneHit("alpha", "Garden three", "https://a/3", 800))
	if secondPage.Score >= first.Score {
		t.Fatalf("page two result %v should score below page one %v", secondPage.Score, first.Score)
	}
}

// TestFusionAcrossLists checks that a result better placed in its own list
// beats a worse placed one from another site, and relevance beats rank.
func TestFusionAcrossLists(t *testing.T) {
	ranker := New("blue lagoon", 0)
	ranker.Add(sceneHit("alpha", "Unrelated", "https://a/1", 100))
	topOfBeta := ranker.Add(sceneHit("beta", "Blue Lagoon", "https://b/1", 200))
	secondOfAlpha := ranker.Add(sceneHit("alpha", "Blue Lagoon", "https://a/2", 300))
	if topOfBeta.Score <= secondOfAlpha.Score {
		t.Fatalf("list leader %v should beat runner-up %v", topOfBeta.Score, secondOfAlpha.Score)
	}
	unrelatedFirst := New("blue lagoon", 0).Add(sceneHit("alpha", "Unrelated", "https://a/9", 400))
	if unrelatedFirst.Score >= secondOfAlpha.Score {
		t.Fatalf("text match should outweigh list place: %v vs %v", unrelatedFirst.Score, secondOfAlpha.Score)
	}
}

// TestDuplicatesCollapse checks that a cross-site copy is not sent again when
// it scores no better, and replaces the sent copy when it does.
func TestDuplicatesCollapse(t *testing.T) {
	ranker := New("garden party", 0)
	original := ranker.Add(sceneHit("alpha", "Garden Party Scene", "https://a/1", 1200))
	if !original.Emit {
		t.Fatal("first copy must be sent")
	}

	if copy := ranker.Add(sceneHit("beta", "Garden Party Scene", "https://b/1", 1205)); copy.Emit {
		t.Fatalf("an equal copy must not be sent again: %+v", copy)
	}
}

// TestBetterDuplicateReplaces checks that a copy with better text replaces the
// one already sent, carrying the fused score of both.
func TestBetterDuplicateReplaces(t *testing.T) {
	ranker := New("garden party", 0)
	first := ranker.Add(sceneHit("alpha", "Garden Party Scene", "https://a/1", 1200))
	// Same video, runtime within 5 %, near-identical title, but with a rating
	// that lifts its quality term.
	better := sceneHit("beta", "Garden Party Scene", "https://b/1", 1205)
	better.Item.Rating = 9
	verdict := ranker.Add(better)

	if !verdict.Emit || verdict.Replaces != "https://a/1" {
		t.Fatalf("verdict = %+v, want replacement of the alpha copy", verdict)
	}
	if verdict.Score <= first.Score {
		t.Fatalf("fused copy score %v should exceed the single one %v", verdict.Score, first.Score)
	}
}

// TestDuplicateNeedsKnownRuntime checks that scenes without a runtime are
// never merged.
func TestDuplicateNeedsKnownRuntime(t *testing.T) {
	ranker := New("garden", 0)
	ranker.Add(sceneHit("alpha", "Garden Party Scene", "https://a/1", 0))
	if verdict := ranker.Add(sceneHit("beta", "Garden Party Scene", "https://b/1", 0)); !verdict.Emit || verdict.Replaces != "" {
		t.Fatalf("scenes without runtime must stay separate, got %+v", verdict)
	}
}

// TestSameRecordCollapses checks that a library hit and the plugin hit of the
// same stored record count as one result.
func TestSameRecordCollapses(t *testing.T) {
	ranker := New("garden", 0)
	library := sceneHit(LibrarySource, "Garden", "https://a/1", 0)
	library.ID = "scene:one"
	library.Performers = []string{"Someone"}
	ranker.Add(library)

	plugin := sceneHit("alpha", "Garden", "https://a/1", 0)
	plugin.ID = "scene:one"
	if verdict := ranker.Add(plugin); verdict.Emit {
		t.Fatalf("same record sent twice: %+v", verdict)
	}
}

// TestQuality checks the shape of the quality term.
func TestQuality(t *testing.T) {
	if Quality(0, 0) != NeutralQuality {
		t.Fatal("no numbers must be neutral")
	}
	if Quality(10, 10_000_000) != 1 {
		t.Fatalf("best listing = %v, want 1", Quality(10, 10_000_000))
	}
	if Quality(5, 0) != 0.5 {
		t.Fatalf("rating only = %v, want 0.5", Quality(5, 0))
	}
	if Quality(0, 1000) >= Quality(0, 1_000_000) {
		t.Fatal("more views must rate higher")
	}
}

// TestPersonalStaysBounded checks the affinity mix is clamped and that dislike
// is negative.
func TestPersonalStaysBounded(t *testing.T) {
	loved := Personal(Affinities{Performers: []float64{1}, Tags: []float64{1, 1, 1, 1}, Studio: 1, Site: 1})
	if loved <= 0.9 || loved > 1 {
		t.Fatalf("loved = %v", loved)
	}
	disliked := Personal(Affinities{Performers: []float64{-1, -0.5}})
	if disliked >= 0 {
		t.Fatalf("disliked = %v", disliked)
	}
	if Personal(Affinities{}) != 0 {
		t.Fatal("no signal must be zero")
	}
}

// TestPersonalNeverOutweighsRelevance checks the weights keep taste small: a
// fully loved but irrelevant result stays below a relevant neutral one.
func TestPersonalNeverOutweighsRelevance(t *testing.T) {
	relevant := New("garden", 0).Add(sceneHit("alpha", "Garden", "https://a/1", 100))
	irrelevant := sceneHit("alpha", "Something else", "https://a/2", 100)
	irrelevant.Personal = 1
	if got := New("garden", 0).Add(irrelevant); got.Score >= relevant.Score {
		t.Fatalf("taste %v outweighed relevance %v", got.Score, relevant.Score)
	}
}

// TestVariantMatchesScoreBelowTheQueryAsTyped checks a result matching the
// rewritten query is found, but one matching what was typed wins.
func TestVariantMatchesScoreBelowTheQueryAsTyped(t *testing.T) {
	typed := New("mia k", 0).Add(sceneHit("alpha", "Mia K at the pool", "https://a/1", 100))

	withVariant := New("mia k", 0)
	withVariant.SetVariants("mia khalifa")
	spelledOut := withVariant.Add(sceneHit("alpha", "Mia Khalifa at the pool", "https://a/2", 200))
	withoutVariant := New("mia k", 0).Add(sceneHit("alpha", "Mia Khalifa at the pool", "https://a/2", 200))

	if spelledOut.Score < withoutVariant.Score {
		t.Fatalf("variant must not lower a score: %v < %v", spelledOut.Score, withoutVariant.Score)
	}
	if spelledOut.Score >= typed.Score {
		t.Fatalf("variant match %v should score below the typed query's %v", spelledOut.Score, typed.Score)
	}
}

// TestEntityCreditLiftsCreditedResults checks the entity term adds
// WeightEntity times the credit.
func TestEntityCreditLiftsCreditedResults(t *testing.T) {
	plain := sceneHit("alpha", "Garden", "https://a/1", 100)
	credited := sceneHit("alpha", "Garden", "https://a/2", 200)
	credited.EntityCredit = 1
	base := New("garden", 0).Add(plain).Score
	lifted := New("garden", 0).Add(credited).Score
	if diff := lifted - base; diff < WeightEntity-0.001 || diff > WeightEntity+0.001 {
		t.Fatalf("entity term = %v, want %v", diff, WeightEntity)
	}
}

// TestLearnedClickMemoryAndSourceShare checks a remembered scene is lifted by
// WeightClickMemory times its memory and a favoured site's rank share scales.
func TestLearnedClickMemoryAndSourceShare(t *testing.T) {
	remembered := sceneHit("alpha", "Garden", "https://a/1", 100)
	remembered.ID = "scene:one"
	ranker := New("garden", 0)
	ranker.SetLearned(Learned{ClickMemory: map[string]float64{"scene:one": 0.5}})
	lifted := ranker.Add(remembered).Score
	base := New("garden", 0).Add(remembered).Score
	if diff := lifted - base; diff < 0.5*WeightClickMemory-0.001 || diff > 0.5*WeightClickMemory+0.001 {
		t.Fatalf("memory term = %v", diff)
	}

	favoured := New("garden", 0)
	favoured.SetLearned(Learned{SourceShares: map[string]float64{"alpha": 1.5}})
	boosted := favoured.Add(sceneHit("alpha", "Garden", "https://a/3", 300)).Score
	plain := New("garden", 0).Add(sceneHit("alpha", "Garden", "https://a/3", 300)).Score
	if diff := boosted - plain; diff < 0.5*WeightSiteRank-0.001 || diff > 0.5*WeightSiteRank+0.001 {
		t.Fatalf("source share term = %v", diff)
	}
}

// TestPersonalWeightIsSetPerQuery checks the same affinity moves a broad
// query's results further than a specific one's.
func TestPersonalWeightIsSetPerQuery(t *testing.T) {
	liked := sceneHit("alpha", "Garden", "https://a/1", 100)
	liked.Personal = 1
	broad := New("garden", 0)
	broad.SetPersonalWeight(PersonalWeight(0))
	specific := New("garden", 0)
	specific.SetPersonalWeight(PersonalWeight(1))
	broadScore, specificScore := broad.Add(liked).Score, specific.Add(liked).Score
	base := New("garden", 0).Add(sceneHit("alpha", "Garden", "https://a/1", 100)).Score
	if !(broadScore-base > specificScore-base && specificScore-base > 0) {
		t.Fatalf("broad +%v, specific +%v", broadScore-base, specificScore-base)
	}
}

// TestSeparateListsKeepTheirOwnPlaces checks an entity's page list and the
// keyword list of one plugin each start at first place.
func TestSeparateListsKeepTheirOwnPlaces(t *testing.T) {
	ranker := New("garden", 0)
	ranker.Add(sceneHit("alpha", "Garden one", "https://a/1", 100))
	page := sceneHit("alpha", "Garden two", "https://a/2", 200)
	page.List = "page:performer:x"
	pageScore := ranker.Add(page).Score
	keyword := New("garden", 0).Add(sceneHit("alpha", "Garden two", "https://a/2", 200)).Score
	if pageScore != keyword {
		t.Fatalf("page list score %v should equal a first place's %v", pageScore, keyword)
	}
}
