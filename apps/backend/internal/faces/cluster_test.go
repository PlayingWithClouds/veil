package faces

import (
	"testing"

	"github.com/playingwithclouds/veil/internal/embeddings"
)

// identity is a unit axis standing for one person's face.
func identity(axis int) []float32 {
	vector := make([]float32, 8)
	vector[axis] = 1
	return vector
}

// noisy is a face of the person on axis, turned by a little of another axis
// so no two sightings are equal.
func noisy(axis, other int, amount float32) []float32 {
	vector := identity(axis)
	vector[other] += amount
	return vector
}

// at returns a frame time.
func at(seconds float64) *float64 {
	return &seconds
}

// TestClusterGroupsPeopleAndKeepsFramesApart groups two people's faces,
// largest first, and never puts two faces of one frame into one person.
func TestClusterGroupsPeopleAndKeepsFramesApart(t *testing.T) {
	sightings := []Sighting{
		{EmbeddingID: "a1", Vector: noisy(0, 5, 0.2), AtSeconds: at(10)},
		{EmbeddingID: "b1", Vector: noisy(1, 6, 0.2), AtSeconds: at(10)},
		{EmbeddingID: "a2", Vector: noisy(0, 6, 0.3), AtSeconds: at(20)},
		{EmbeddingID: "a3", Vector: noisy(0, 7, 0.1), AtSeconds: at(5)},
		{EmbeddingID: "b2", Vector: noisy(1, 5, 0.3), AtSeconds: at(30)},
	}
	people := Cluster(sightings, linkage)
	if len(people) != 2 {
		t.Fatalf("people = %d, want 2", len(people))
	}
	if ids(people[0]) != "a3 a1 a2" || ids(people[1]) != "b1 b2" {
		t.Errorf("people = %q, %q", ids(people[0]), ids(people[1]))
	}
	if similarity, _ := embeddings.Cosine(people[0].Centroid, identity(0)); similarity < 0.95 {
		t.Errorf("centroid similarity = %f", similarity)
	}

	alike := []Sighting{
		{EmbeddingID: "x", Vector: identity(0), AtSeconds: at(1)},
		{EmbeddingID: "y", Vector: identity(0), AtSeconds: at(1)},
	}
	if people := Cluster(alike, linkage); len(people) != 2 {
		t.Errorf("two faces of one frame became %d people", len(people))
	}
}

// TestCandidatesAndAssign ranks performers by their closest known face and
// assigns each person and performer at most once.
func TestCandidatesAndAssign(t *testing.T) {
	ada := Person{Centroid: identity(0)}
	bea := Person{Centroid: identity(1)}
	picture := "blob://ada"
	references := []Reference{
		{PerformerID: "ada", Vector: noisy(0, 2, 0.3), Picture: &picture},
		{PerformerID: "ada", Vector: noisy(0, 2, 1.5)},
		{PerformerID: "bea", Vector: noisy(1, 0, 0.5)},
		{PerformerID: "cy", Vector: identity(3)},
	}
	adaCandidates := Candidates(ada, references, []string{"bea"})
	if len(adaCandidates) != 2 || adaCandidates[0].PerformerID != "ada" || !adaCandidates[0].Likely || adaCandidates[0].Picture == nil {
		t.Fatalf("candidates = %+v", adaCandidates)
	}
	if adaCandidates[1].PerformerID != "bea" || adaCandidates[1].Likely || !adaCandidates[1].Credited {
		t.Errorf("second candidate = %+v", adaCandidates[1])
	}
	guesses := Assign([][]Candidate{adaCandidates, Candidates(bea, references, nil)})
	if len(guesses) != 2 || guesses[0].PerformerID != "ada" || guesses[0].Person != 0 || guesses[1].PerformerID != "bea" || guesses[1].Person != 1 {
		t.Errorf("guesses = %+v", guesses)
	}
}

// TestLearnFallsBackToTheDominantPerson learns a lone credited performer
// nobody knows the face of only when one person clearly dominates.
func TestLearnFallsBackToTheDominantPerson(t *testing.T) {
	dominant := []Person{
		{Sightings: make([]Sighting, 4), Centroid: identity(0)},
		{Sightings: make([]Sighting, 1), Centroid: identity(1)},
	}
	if learned := Learn(dominant, nil, []string{"ada"}); learned[0] != "ada" || len(learned) != 1 {
		t.Errorf("learned = %v", learned)
	}
	even := []Person{
		{Sightings: make([]Sighting, 4), Centroid: identity(0)},
		{Sightings: make([]Sighting, 3), Centroid: identity(1)},
	}
	if learned := Learn(even, nil, []string{"ada"}); len(learned) != 0 {
		t.Errorf("learned from an even video: %v", learned)
	}
	if learned := Learn(dominant, nil, []string{"ada", "bea"}); len(learned) != 0 {
		t.Errorf("learned with two unknown credits: %v", learned)
	}
}

// ids joins a person's sighting ids.
func ids(person Person) string {
	joined := ""
	for index, sighting := range person.Sightings {
		if index > 0 {
			joined += " "
		}
		joined += sighting.EmbeddingID
	}
	return joined
}
