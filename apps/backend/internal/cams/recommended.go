package cams

import (
	"cmp"
	"context"
	"math"
	"slices"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/schema"
)

const (
	// recommendDepth is how many of the most watched live rooms are ranked.
	recommendDepth = 300
	// performerFactor weighs a linked performer the user likes over one tag.
	performerFactor = 2.0
	// audienceFactor lets a bigger audience break ties between equal matches.
	audienceFactor = 0.05
)

// Affinity is what the user likes and dislikes, as weights in [-1, 1].
type Affinity struct {
	// Tags maps a tag name to its weight.
	Tags map[string]float64
	// Performers maps a library performer id to its weight.
	Performers map[string]float64
}

// scoredRoom is a room with how well it fits the user.
type scoredRoom struct {
	room  schema.CamModel
	score float64
}

// Recommended returns up to limit rooms live now that match what the user
// likes, best first. Followed, auto-recorded and blocked rooms are left out:
// the user has those already or doesn't want them.
func (s *Service) Recommended(ctx context.Context, affinity Affinity, limit int) ([]schema.CamModel, error) {
	rooms, _, err := s.List(ctx, Query{Limit: recommendDepth})
	if err != nil {
		return nil, err
	}
	watchedIDs, err := s.services.Database.Strings(ctx, `SELECT "id" FROM (`+watchedModelsQuery+`)`,
		db.Vars{"camModel": schema.RecordKindCamModel})
	if err != nil {
		return nil, err
	}
	tagWeights := tagKeyWeights(affinity.Tags)
	var scored []scoredRoom
	for _, room := range rooms {
		if slices.Contains(watchedIDs, room.Id) {
			continue
		}
		score := roomScore(room, tagWeights, affinity.Performers)
		if score <= 0 {
			continue
		}
		scored = append(scored, scoredRoom{room: room, score: score})
	}
	slices.SortStableFunc(scored, func(first, second scoredRoom) int {
		return cmp.Compare(second.score, first.score)
	})
	return bestRooms(scored, limit), nil
}

// roomScore adds up the weights of the room's tags and linked performer; a
// room matching nothing scores 0 whatever its audience.
func roomScore(room schema.CamModel, tagWeights, performerWeights map[string]float64) float64 {
	match := 0.0
	for _, tag := range room.Tags {
		match += tagWeights[tagKey(tag)]
	}
	if room.Performer != nil {
		match += performerFactor * performerWeights[*room.Performer]
	}
	if match == 0 {
		return 0
	}
	return match + audienceFactor*math.Log10(float64(viewersOf(room))+1)
}

// tagKeyWeights keys tag weights the way cam sites spell tags: "Big Boobs"
// and "bigboobs" are one tag.
func tagKeyWeights(weights map[string]float64) map[string]float64 {
	keyed := make(map[string]float64, len(weights))
	for name, weight := range weights {
		keyed[tagKey(name)] = weight
	}
	return keyed
}

// tagKey folds a tag name for comparison, without separators.
func tagKey(name string) string {
	return strings.ReplaceAll(names.Key(name), " ", "")
}

// bestRooms returns the rooms of the first limit scored ones.
func bestRooms(scored []scoredRoom, limit int) []schema.CamModel {
	rooms := []schema.CamModel{}
	for _, entry := range scored {
		if len(rooms) >= limit {
			break
		}
		rooms = append(rooms, entry.room)
	}
	return rooms
}
