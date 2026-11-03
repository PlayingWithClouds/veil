package cams

import (
	"cmp"
	"context"
	"math"
	"slices"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/schema"
)

const (
	// recommendDepth is how many of the most watched live rooms are ranked.
	recommendDepth = 300
	// forYouDepth is how many of the most watched rooms SortForYou ranks.
	forYouDepth = 600
	// performerFactor weighs a linked performer the user likes over one tag.
	performerFactor = 2.0
	// roomFactor weighs a room the user keeps returning to.
	roomFactor = 3.0
	// similarFactor weighs a room like one the user returns to.
	similarFactor = 1.5
	// similarMinimumShared is how many tags a room shares with a watched one
	// to count as like it.
	similarMinimumShared = 2
	// anchorMinimumWeight and anchorLimit pick the watched rooms others are
	// compared to: the strongest, above the minimum.
	anchorMinimumWeight = 0.3
	anchorLimit         = 8
	// siteFactor and genderFactor lean a matched room toward the sites and
	// genders the user watches; on their own they don't make a match.
	siteFactor   = 0.3
	genderFactor = 0.5
	// audienceFactor lets a bigger audience break ties between equal matches.
	audienceFactor = 0.05
	// exploreEvery puts an unmatched popular room at every exploreEvery-th
	// place of the recommendations, so they don't narrow to the same few.
	exploreEvery = 6
)

// Reasons a room is recommended.
const (
	reasonExplore = "Something different"
	reasonRoom    = "You often watch this room"
)

// Affinity is what the user likes and dislikes, as weights in [-1, 1].
type Affinity struct {
	// Tags maps a tag name, spelled any way, to its weight.
	Tags map[string]float64
	// Performers maps a library performer id to its weight.
	Performers map[string]float64
	// PerformerNames names the performers, for the reasons.
	PerformerNames map[string]string
	// Sites maps a plugin name to its weight.
	Sites map[string]float64
	// Genders maps a room gender to its weight.
	Genders map[string]float64
	// Rooms maps a room id to its weight.
	Rooms map[string]float64
}

// Recommendation is a room and why it is recommended.
type Recommendation struct {
	Room schema.CamModel
	// Reason is shown with the room, e.g. "Because you like #asian"; empty
	// for a room that matched nothing.
	Reason string
}

// anchor is a room the user returns to, which others are compared with.
type anchor struct {
	id     string
	name   string
	weight float64
	tags   map[string]bool
}

// scoredRoom is a room with how well it fits the user and why.
type scoredRoom struct {
	room schema.CamModel
	// matched is whether anything of the room is liked; score then ranks it.
	matched bool
	score   float64
	reason  string
}

// Recommended returns up to limit rooms live now that match what the user
// likes, best first, every exploreEvery-th an unmatched popular one.
// Followed, auto-recorded and blocked rooms are left out: the user has those
// already or doesn't want them.
func (s *Service) Recommended(ctx context.Context, affinity Affinity, limit int) ([]Recommendation, error) {
	rooms, _, err := s.List(ctx, Query{Limit: recommendDepth})
	if err != nil {
		return nil, err
	}
	keptIDs, err := s.services.Database.Strings(ctx, `SELECT "id" FROM (`+watchedModelsQuery+`)`,
		db.Vars{"camModel": schema.RecordKindCamModel})
	if err != nil {
		return nil, err
	}
	anchors, err := s.anchors(ctx, affinity.Rooms)
	if err != nil {
		return nil, err
	}
	rooms = slices.DeleteFunc(rooms, func(room schema.CamModel) bool {
		return slices.Contains(keptIDs, room.Id)
	})
	matched, unmatched := splitMatched(scoreRooms(rooms, affinity, anchors))
	if len(matched) == 0 {
		return []Recommendation{}, nil
	}
	return withExploration(matched, unmatched, limit), nil
}

// ForYou returns one page of the rooms the query lists, ranked by the user's
// taste among the forYouDepth most watched: matched rooms best first, then
// the rest by viewers.
func (s *Service) ForYou(ctx context.Context, query Query, affinity Affinity) ([]Recommendation, int, error) {
	deep := query
	deep.Sort = SortViewers
	deep.Offset = 0
	deep.Limit = forYouDepth
	rooms, total, err := s.List(ctx, deep)
	if err != nil {
		return nil, 0, err
	}
	anchors, err := s.anchors(ctx, affinity.Rooms)
	if err != nil {
		return nil, 0, err
	}
	matched, unmatched := splitMatched(scoreRooms(rooms, affinity, anchors))
	ranked := recommendations(append(matched, unmatched...))
	limit := query.Limit
	if limit <= 0 {
		limit = defaultPageLimit
	}
	return pageOf(ranked, query.Offset, limit), total, nil
}

// anchors reads the rooms the user returns to most, strongest first.
func (s *Service) anchors(ctx context.Context, roomWeights map[string]float64) ([]anchor, error) {
	ids := []string{}
	for id, weight := range roomWeights {
		if weight >= anchorMinimumWeight {
			ids = append(ids, id)
		}
	}
	slices.SortFunc(ids, func(first, second string) int {
		return cmp.Or(cmp.Compare(roomWeights[second], roomWeights[first]), cmp.Compare(first, second))
	})
	ids = ids[:min(len(ids), anchorLimit)]
	if len(ids) == 0 {
		return nil, nil
	}
	rooms, err := db.QueryAs[schema.CamModel](ctx, s.services.Database,
		`SELECT * FROM "CamModel" WHERE "id" IN (SELECT value FROM json_each($ids)) AND "deletedAt" IS NULL`,
		db.Vars{"ids": ids})
	if err != nil {
		return nil, err
	}
	anchors := make([]anchor, 0, len(rooms))
	for _, room := range rooms {
		anchors = append(anchors, anchor{id: room.Id, name: room.DisplayName, weight: roomWeights[room.Id], tags: tagKeySet(room.Tags)})
	}
	slices.SortFunc(anchors, func(first, second anchor) int {
		return cmp.Compare(second.weight, first.weight)
	})
	return anchors, nil
}

// scoreRooms scores every room, keeping their order.
func scoreRooms(rooms []schema.CamModel, affinity Affinity, anchors []anchor) []scoredRoom {
	affinity.Tags = tagKeyWeights(affinity.Tags)
	scored := make([]scoredRoom, 0, len(rooms))
	for _, room := range rooms {
		scored = append(scored, scoreRoom(room, affinity, anchors))
	}
	return scored
}

// splitMatched returns the rooms that matched the taste, best first, and the
// ones that didn't, in their order.
func splitMatched(scored []scoredRoom) ([]scoredRoom, []scoredRoom) {
	matched, unmatched := []scoredRoom{}, []scoredRoom{}
	for _, entry := range scored {
		if entry.matched {
			matched = append(matched, entry)
			continue
		}
		unmatched = append(unmatched, entry)
	}
	slices.SortStableFunc(matched, func(first, second scoredRoom) int {
		return cmp.Compare(second.score, first.score)
	})
	return matched, unmatched
}

// contribution is one part of a room's match and the reason it gives.
type contribution struct {
	value  float64
	reason string
}

// scoreRoom adds up what the room shares with the user's taste: its tags,
// linked performer, the room itself and how like it is to the rooms the user
// returns to. A room matching none of them scores 0 whatever its site,
// gender or audience; otherwise those lean the score. The reason is the
// biggest part.
func scoreRoom(room schema.CamModel, affinity Affinity, anchors []anchor) scoredRoom {
	parts := []contribution{tagContribution(room, affinity.Tags)}
	if room.Performer != nil {
		performer := *room.Performer
		parts = append(parts, contribution{performerFactor * affinity.Performers[performer], "More with " + affinity.PerformerNames[performer]})
	}
	parts = append(parts, contribution{roomFactor * affinity.Rooms[room.Id], reasonRoom})
	parts = append(parts, similarContribution(room, anchors))
	match := 0.0
	best := contribution{}
	for _, part := range parts {
		match += part.value
		if part.value > best.value {
			best = part
		}
	}
	if match <= 0 || best.value <= 0 {
		return scoredRoom{room: room}
	}
	score := match + siteFactor*affinity.Sites[room.Plugin] + audienceFactor*math.Log10(float64(viewersOf(room))+1)
	if room.Gender != nil {
		score += genderFactor * affinity.Genders[*room.Gender]
	}
	return scoredRoom{room: room, matched: true, score: score, reason: best.reason}
}

// tagContribution adds up the weights of the room's tags; the reason names
// the one that weighs most.
func tagContribution(room schema.CamModel, tagWeights map[string]float64) contribution {
	total := 0.0
	strongest := 0.0
	reason := ""
	for _, tag := range room.Tags {
		weight := tagWeights[tagKey(tag)]
		total += weight
		if weight > strongest {
			strongest = weight
			reason = "Because you like #" + tag
		}
	}
	return contribution{total, reason}
}

// similarContribution is how like the room is to the most alike room the
// user returns to: their weight times the cosine of their tag sets, counted
// only when they share similarMinimumShared tags.
func similarContribution(room schema.CamModel, anchors []anchor) contribution {
	tags := tagKeySet(room.Tags)
	best := contribution{}
	for _, candidate := range anchors {
		if candidate.id == room.Id {
			continue
		}
		shared := 0
		for tag := range tags {
			if candidate.tags[tag] {
				shared++
			}
		}
		if shared < similarMinimumShared {
			continue
		}
		similarity := float64(shared) / math.Sqrt(float64(len(tags)*len(candidate.tags)))
		value := similarFactor * candidate.weight * similarity
		if value > best.value {
			best = contribution{value, "Because you watched " + candidate.name}
		}
	}
	return best
}

// withExploration takes up to limit rooms: the matched ones in order, with an
// unmatched one at every exploreEvery-th place while there are any. It ends
// with the matched rooms.
func withExploration(matched, unmatched []scoredRoom, limit int) []Recommendation {
	picked := []Recommendation{}
	for len(picked) < limit && len(matched) > 0 {
		exploring := (len(picked)+1)%exploreEvery == 0
		if exploring && len(unmatched) > 0 {
			picked = append(picked, Recommendation{Room: unmatched[0].room, Reason: reasonExplore})
			unmatched = unmatched[1:]
			continue
		}
		picked = append(picked, Recommendation{Room: matched[0].room, Reason: matched[0].reason})
		matched = matched[1:]
	}
	return picked
}

// recommendations turns scored rooms into recommendations, in order.
func recommendations(scored []scoredRoom) []Recommendation {
	out := make([]Recommendation, 0, len(scored))
	for _, entry := range scored {
		out = append(out, Recommendation{Room: entry.room, Reason: entry.reason})
	}
	return out
}

// tagKeyWeights keys tag weights by tagKey; spellings of one tag add up.
func tagKeyWeights(weights map[string]float64) map[string]float64 {
	keyed := make(map[string]float64, len(weights))
	for name, weight := range weights {
		keyed[tagKey(name)] += weight
	}
	return keyed
}

// tagKeySet is the set of a room's tag keys.
func tagKeySet(tags []string) map[string]bool {
	set := make(map[string]bool, len(tags))
	for _, tag := range tags {
		key := tagKey(tag)
		if key != "" {
			set[key] = true
		}
	}
	return set
}

// tagKey folds a tag name for comparison, without separators: "Big Boobs"
// and "bigboobs" are one tag.
func tagKey(name string) string {
	return names.CompactKey(name)
}
