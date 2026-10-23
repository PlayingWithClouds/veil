package searchrank

import "testing"

// TestEntityCreditWeighsKindsAndMentions checks stored credit beats a title
// mention, a missing performer weighs more than a missing tag, and a fuzzy
// entity counts less.
func TestEntityCreditWeighsKindsAndMentions(t *testing.T) {
	index := testIndex()
	matches := index.Find("mia khalifa brazzers")
	credited := map[string]bool{"performer:mia": true, "studio:brazzers": true}
	if got := EntityCredit(matches, credited, "", "Anything"); got != 1 {
		t.Fatalf("fully credited = %v", got)
	}
	performerOnly := EntityCredit(matches, map[string]bool{"performer:mia": true}, "", "Anything")
	studioOnly := EntityCredit(matches, map[string]bool{"studio:brazzers": true}, "", "Anything")
	if !(performerOnly > studioOnly && studioOnly > 0 && performerOnly < 1) {
		t.Fatalf("performer only %v, studio only %v", performerOnly, studioOnly)
	}
	mentioned := EntityCredit(matches[:1], nil, "", "Mia Khalifa at the pool")
	if mentioned != EntityMentionCredit {
		t.Fatalf("mention = %v", mentioned)
	}
	if own := EntityCredit(matches[:1], nil, "performer:mia", "Mia Khalifa"); own != 1 {
		t.Fatalf("the entity's own record = %v", own)
	}
	fuzzy := index.Find("mia khalfia")
	if got := EntityCredit(fuzzy, credited, "", ""); got != FuzzyEntityFactor {
		t.Fatalf("fuzzy credit = %v", got)
	}
	if got := EntityCredit(nil, credited, "", ""); got != 0 {
		t.Fatalf("no matches = %v", got)
	}
}
