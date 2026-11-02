package faces

import (
	"math"
	"sort"

	"github.com/playingwithclouds/veil/internal/embeddings"
)

// Sighting is one face found in a video: its vector and where it was.
type Sighting struct {
	EmbeddingID string
	Vector      []float32
	AtSeconds   *float64
	Region      any
}

// Person is a group of sightings taken for one person, with their mean
// face.
type Person struct {
	Sightings []Sighting
	Centroid  []float32
}

// Cluster groups sightings into people by average-linkage agglomeration:
// the two groups whose faces are most alike on average merge while that
// average reaches linkage. Two faces of one frame are never one person.
// People come back most seen first, sightings in time order.
func Cluster(sightings []Sighting, linkage float64) []Person {
	similarity := pairwise(sightings)
	groups := make([][]int, len(sightings))
	for index := range sightings {
		groups[index] = []int{index}
	}
	for {
		first, second, found := closestPair(groups, sightings, similarity, linkage)
		if !found {
			break
		}
		groups[first] = append(groups[first], groups[second]...)
		groups = append(groups[:second], groups[second+1:]...)
	}
	return people(groups, sightings)
}

// pairwise is the cosine similarity of every two sightings.
func pairwise(sightings []Sighting) [][]float64 {
	similarity := make([][]float64, len(sightings))
	for first := range sightings {
		similarity[first] = make([]float64, len(sightings))
		for second := range sightings {
			value, comparable := embeddings.Cosine(sightings[first].Vector, sightings[second].Vector)
			if !comparable {
				value = -1
			}
			similarity[first][second] = value
		}
	}
	return similarity
}

// closestPair finds the two groups with the highest average similarity at
// or above linkage that may merge; found is false when none qualify.
func closestPair(groups [][]int, sightings []Sighting, similarity [][]float64, linkage float64) (int, int, bool) {
	best := math.Inf(-1)
	bestFirst, bestSecond := -1, -1
	for first := range groups {
		for second := first + 1; second < len(groups); second++ {
			if shareFrame(groups[first], groups[second], sightings) {
				continue
			}
			average := averageLink(groups[first], groups[second], similarity)
			if average >= linkage && average > best {
				best, bestFirst, bestSecond = average, first, second
			}
		}
	}
	return bestFirst, bestSecond, bestFirst >= 0
}

// shareFrame reports whether two groups hold faces of the same frame.
func shareFrame(first, second []int, sightings []Sighting) bool {
	for _, one := range first {
		for _, other := range second {
			if sameFrame(sightings[one], sightings[other]) {
				return true
			}
		}
	}
	return false
}

// sameFrame reports whether two sightings come from one frame; faces
// without a time (pictures) never do.
func sameFrame(first, second Sighting) bool {
	if first.AtSeconds == nil || second.AtSeconds == nil {
		return false
	}
	return math.Abs(*first.AtSeconds-*second.AtSeconds) < 1e-3
}

// averageLink is the mean similarity between the members of two groups.
func averageLink(first, second []int, similarity [][]float64) float64 {
	sum := 0.0
	for _, one := range first {
		for _, other := range second {
			sum += similarity[one][other]
		}
	}
	return sum / float64(len(first)*len(second))
}

// people turns index groups into people, most seen first.
func people(groups [][]int, sightings []Sighting) []Person {
	result := make([]Person, 0, len(groups))
	for _, group := range groups {
		person := Person{}
		for _, index := range group {
			person.Sightings = append(person.Sightings, sightings[index])
		}
		sort.SliceStable(person.Sightings, func(first, second int) bool {
			return seconds(person.Sightings[first]) < seconds(person.Sightings[second])
		})
		person.Centroid = Centroid(person.Sightings)
		result = append(result, person)
	}
	sort.SliceStable(result, func(first, second int) bool {
		return len(result[first].Sightings) > len(result[second].Sightings)
	})
	return result
}

// seconds is a sighting's frame time, 0 without one.
func seconds(sighting Sighting) float64 {
	if sighting.AtSeconds == nil {
		return 0
	}
	return *sighting.AtSeconds
}

// Centroid is the unit-length mean of the sightings' vectors: averaging
// several frames of a face evens out pose and light.
func Centroid(sightings []Sighting) []float32 {
	if len(sightings) == 0 {
		return nil
	}
	sum := make([]float64, len(sightings[0].Vector))
	for _, sighting := range sightings {
		addUnit(sum, sighting.Vector)
	}
	return unit(sum)
}

// addUnit adds a vector scaled to unit length to sum; vectors of another
// length are skipped.
func addUnit(sum []float64, vector []float32) {
	if len(vector) != len(sum) {
		return
	}
	norm := 0.0
	for _, value := range vector {
		norm += float64(value) * float64(value)
	}
	if norm == 0 {
		return
	}
	norm = math.Sqrt(norm)
	for index, value := range vector {
		sum[index] += float64(value) / norm
	}
}

// unit scales a vector to unit length as float32.
func unit(vector []float64) []float32 {
	norm := 0.0
	for _, value := range vector {
		norm += value * value
	}
	result := make([]float32, len(vector))
	if norm == 0 {
		return result
	}
	norm = math.Sqrt(norm)
	for index, value := range vector {
		result[index] = float32(value / norm)
	}
	return result
}
