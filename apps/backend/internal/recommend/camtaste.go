package recommend

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
)

// CamTaste is the whole taste as cam rooms are ranked by it: every tag,
// performer, site, gender and room the user's videos, galleries and rooms
// gave a weight, each in [-1, 1].
type CamTaste struct {
	// Tags maps names.CompactKey of a tag name to its weight: library tags
	// by their name, and room tags no library tag is spelled like.
	Tags map[string]float64
	// Performers maps a library performer id to its weight.
	Performers map[string]float64
	// PerformerNames names the performers of Performers.
	PerformerNames map[string]string
	// Sites maps a plugin name to its weight.
	Sites map[string]float64
	// Genders maps a room gender (female, male, couple, trans) to its weight.
	Genders map[string]float64
	// Rooms maps a cam room id to its weight: the rooms the user returns to.
	Rooms map[string]float64
}

// camTasteCache keeps the last CamTaste for camTasteTTL, so paging through
// rooms ranked by it doesn't rebuild the profile for every page.
type camTasteCache struct {
	mutex   sync.Mutex
	taste   *CamTaste
	builtAt time.Time
}

// CamTaste returns the current taste for ranking cam rooms.
func (e *Engine) CamTaste(ctx context.Context) (CamTaste, error) {
	e.camTaste.mutex.Lock()
	defer e.camTaste.mutex.Unlock()
	now := e.now()
	if e.camTaste.taste != nil && now.Sub(e.camTaste.builtAt) < camTasteTTL {
		return *e.camTaste.taste, nil
	}
	taste, _, _, err := e.profileAt(ctx, now)
	if err != nil {
		return CamTaste{}, err
	}
	built, err := e.camTasteOf(ctx, taste)
	if err != nil {
		return CamTaste{}, err
	}
	e.camTaste.taste = &built
	e.camTaste.builtAt = now
	return built, nil
}

// camTasteOf turns a profile into a CamTaste, reading the names of its tags
// and performers.
func (e *Engine) camTasteOf(ctx context.Context, taste *profile) (CamTaste, error) {
	tagNames, err := e.namesOf(ctx, "Tag", setKeysOf(taste.tags))
	if err != nil {
		return CamTaste{}, err
	}
	performerNames, err := e.namesOf(ctx, "Performer", setKeysOf(taste.performers))
	if err != nil {
		return CamTaste{}, err
	}
	return CamTaste{
		Tags:           camTagWeights(taste, tagNames),
		Performers:     taste.performers,
		PerformerNames: performerNames,
		Sites:          taste.sites,
		Genders:        taste.camGenders,
		Rooms:          taste.camRooms,
	}, nil
}

// camTagWeights keys the library tags' weights by their compact name and adds
// the room tags without a library tag. Two library tags spelled alike add up.
func camTagWeights(taste *profile, tagNames map[string]string) map[string]float64 {
	weights := make(map[string]float64, len(taste.tags)+len(taste.camTags))
	for id, weight := range taste.tags {
		key := names.CompactKey(tagNames[id])
		if key == "" {
			continue
		}
		weights[key] += weight
	}
	for key, weight := range taste.camTags {
		weights[key] += weight
	}
	return weights
}

// namesOf reads the names of the records of one table by id; trashed ones
// are left out.
func (e *Engine) namesOf(ctx context.Context, table string, ids []string) (map[string]string, error) {
	out := map[string]string{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := e.database.Query(ctx,
		`SELECT "id", "name" FROM `+db.Quote(table)+`
		 WHERE "id" IN (SELECT value FROM json_each($ids)) AND "deletedAt" IS NULL`,
		db.Vars{"ids": ids})
	if err != nil {
		return nil, fmt.Errorf("%s names: %w", table, err)
	}
	for _, row := range rows {
		out[rowString(row, "id")] = rowString(row, "name")
	}
	return out, nil
}

// setKeysOf lists the keys of an affinity map.
func setKeysOf(affinities map[string]float64) []string {
	keys := make([]string, 0, len(affinities))
	for key := range affinities {
		keys = append(keys, key)
	}
	return keys
}
