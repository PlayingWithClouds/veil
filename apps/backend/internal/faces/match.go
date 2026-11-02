package faces

import (
	"slices"
	"sort"

	"github.com/playingwithclouds/veil/internal/embeddings"
)

// Thresholds on the cosine similarity of SFace vectors. OpenCV puts the
// same-identity line of single SFace pairs at 0.363; a person's mean face
// against a known face scores higher for the same identity, so the lines
// sit a little above it.
const (
	// linkage is the average similarity at which two groups of a video's
	// faces are one person.
	linkage = 0.40
	// likelySimilarity marks a candidate as probably the person, and is
	// what learning a performer's face from a video needs.
	likelySimilarity = 0.45
	// candidateFloor is the least similarity a candidate is listed with.
	candidateFloor = 0.30
	// candidateLimit bounds the candidates listed per person.
	candidateLimit = 5
	// dominantSightings and dominantRatio decide when a video crediting one
	// performer nobody knows the face of shows them: their person must be
	// seen this often and this many times more than anybody else.
	dominantSightings = 3
	dominantRatio     = 2
)

// Reference is a known face of a performer: a picture's vector, or the
// mean face learned from a video crediting them.
type Reference struct {
	PerformerID string
	Vector      []float32
	Picture     *string
	VideoID     *string
}

// Candidate is a performer a person may be, by their closest known face.
type Candidate struct {
	PerformerID string
	Similarity  float64
	Likely      bool
	Credited    bool
	// What the closest known face came from.
	Picture *string
	VideoID *string
}

// Guess names the performer a person probably is.
type Guess struct {
	PerformerID string
	Similarity  float64
	Credited    bool
	// Person is the index of the person in the video's people.
	Person int
}

// Candidates ranks the performers a person may be: each by the similarity
// of the person's mean face to the performer's closest known face.
func Candidates(person Person, references []Reference, credited []string) []Candidate {
	best := map[string]Candidate{}
	for _, reference := range references {
		similarity, comparable := embeddings.Cosine(person.Centroid, reference.Vector)
		if !comparable || similarity < candidateFloor {
			continue
		}
		if current, seen := best[reference.PerformerID]; seen && current.Similarity >= similarity {
			continue
		}
		best[reference.PerformerID] = Candidate{
			PerformerID: reference.PerformerID, Similarity: similarity, Likely: similarity >= likelySimilarity,
			Credited: slices.Contains(credited, reference.PerformerID), Picture: reference.Picture, VideoID: reference.VideoID,
		}
	}
	candidates := make([]Candidate, 0, len(best))
	for _, candidate := range best {
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(first, second int) bool {
		return candidates[first].Similarity > candidates[second].Similarity
	})
	if len(candidates) > candidateLimit {
		candidates = candidates[:candidateLimit]
	}
	return candidates
}

// Assign picks at most one performer per person and one person per
// performer, best similarity first, from the likely candidates.
func Assign(candidates [][]Candidate) []Guess {
	var pairs []Guess
	for person, listed := range candidates {
		for _, candidate := range listed {
			if candidate.Likely {
				pairs = append(pairs, Guess{PerformerID: candidate.PerformerID, Similarity: candidate.Similarity, Credited: candidate.Credited, Person: person})
			}
		}
	}
	sort.SliceStable(pairs, func(first, second int) bool {
		return pairs[first].Similarity > pairs[second].Similarity
	})
	usedPeople := map[int]bool{}
	usedPerformers := map[string]bool{}
	var guesses []Guess
	for _, pair := range pairs {
		if usedPeople[pair.Person] || usedPerformers[pair.PerformerID] {
			continue
		}
		usedPeople[pair.Person] = true
		usedPerformers[pair.PerformerID] = true
		guesses = append(guesses, pair)
	}
	return guesses
}

// Learn decides which person of a video each credited performer is, for
// learning their face: a person likely matching the performer's known
// faces, or, when exactly one credited performer's face is unknown, the
// person clearly seen most among those nobody else was matched to. The
// result maps person index to performer id.
func Learn(people []Person, references []Reference, credited []string) map[int]string {
	var creditedReferences []Reference
	for _, reference := range references {
		if slices.Contains(credited, reference.PerformerID) {
			creditedReferences = append(creditedReferences, reference)
		}
	}
	candidates := make([][]Candidate, len(people))
	for index, person := range people {
		candidates[index] = Candidates(person, creditedReferences, credited)
	}
	learned := map[int]string{}
	for _, guess := range Assign(candidates) {
		learned[guess.Person] = guess.PerformerID
	}
	unknown := withoutKnownFace(credited, creditedReferences)
	if len(unknown) != 1 {
		return learned
	}
	if person, found := dominantLeftover(people, learned); found {
		learned[person] = unknown[0]
	}
	return learned
}

// withoutKnownFace lists the credited performers no reference shows.
func withoutKnownFace(credited []string, references []Reference) []string {
	var unknown []string
	for _, performerID := range credited {
		known := slices.ContainsFunc(references, func(reference Reference) bool {
			return reference.PerformerID == performerID
		})
		if !known {
			unknown = append(unknown, performerID)
		}
	}
	return unknown
}

// dominantLeftover is the index of the person seen clearly most among those
// not learned already, if any; people come most seen first, as Cluster
// returns them.
func dominantLeftover(people []Person, learned map[int]string) (int, bool) {
	var leftover []int
	for index := range people {
		if _, taken := learned[index]; !taken {
			leftover = append(leftover, index)
		}
	}
	if len(leftover) == 0 || len(people[leftover[0]].Sightings) < dominantSightings {
		return 0, false
	}
	if len(leftover) > 1 && len(people[leftover[0]].Sightings) < dominantRatio*len(people[leftover[1]].Sightings) {
		return 0, false
	}
	return leftover[0], true
}
