package recommendeval

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/media"
	"github.com/playingwithclouds/veil/internal/recommend"
)

// Variant names the rankings that can be compared.
const (
	VariantDefault    = "default"
	VariantNoLearning = "no-learning"
	VariantNewest     = "baseline-newest"
	VariantRandom     = "baseline-random"
)

// rankedScene is one position of a ranking.
type rankedScene struct {
	sceneID string
	// source is the candidate source label; empty for baselines.
	source string
}

// ranker produces a ranking of the scenes as of the evaluated moment.
type ranker interface {
	rank(ctx context.Context, depth int) ([]rankedScene, error)
}

// evaluationClock is the time the engines rank at, moved by the run from
// cutoff to cutoff.
type evaluationClock struct {
	current time.Time
}

// now is the engine's clock.
func (clock *evaluationClock) now() time.Time {
	return clock.current
}

// engineRanker ranks with the real engine: the Home feed's first page, fresh.
type engineRanker struct {
	engine *recommend.Engine
}

// rank returns the top depth scenes of a re-ranked feed.
func (ranker engineRanker) rank(ctx context.Context, depth int) ([]rankedScene, error) {
	items, err := ranker.engine.Feed(ctx, nil, nil, depth, 0, true)
	if err != nil {
		return nil, fmt.Errorf("feed: %w", err)
	}
	out := make([]rankedScene, 0, len(items))
	for _, item := range items {
		out = append(out, rankedScene{sceneID: item.SceneID, source: item.Source})
	}
	return out, nil
}

// newestRanker is the baseline that lists the newest unwatched scenes.
type newestRanker struct {
	database *db.DB
}

// rank returns the newest scenes the user never watched.
func (ranker newestRanker) rank(ctx context.Context, depth int) ([]rankedScene, error) {
	ids, err := ranker.database.Strings(ctx,
		`SELECT id FROM scene WHERE id NOT IN (SELECT media FROM watch_history)
		 ORDER BY created_at DESC LIMIT $depth`, db.Vars{"depth": depth})
	if err != nil {
		return nil, fmt.Errorf("newest scenes: %w", err)
	}
	return unlabeled(ids), nil
}

// randomRanker is the baseline that lists random unwatched scenes.
type randomRanker struct {
	database *db.DB
	random   *rand.Rand
}

// rank returns depth random scenes the user never watched.
func (ranker randomRanker) rank(ctx context.Context, depth int) ([]rankedScene, error) {
	ids, err := ranker.database.Strings(ctx,
		`SELECT id FROM scene WHERE id NOT IN (SELECT media FROM watch_history) ORDER BY id`, nil)
	if err != nil {
		return nil, fmt.Errorf("unwatched scenes: %w", err)
	}
	ranker.random.Shuffle(len(ids), func(left, right int) { ids[left], ids[right] = ids[right], ids[left] })
	if len(ids) > depth {
		ids = ids[:depth]
	}
	return unlabeled(ids), nil
}

// unlabeled wraps scene ids into a ranking without source labels.
func unlabeled(ids []string) []rankedScene {
	out := make([]rankedScene, 0, len(ids))
	for _, id := range ids {
		out = append(out, rankedScene{sceneID: id})
	}
	return out
}

// newRanker builds the ranker of a variant over the database; engines read
// their time from clock.
func newRanker(name string, database *db.DB, clock *evaluationClock, seed int64) (ranker, error) {
	switch name {
	case VariantDefault, VariantNoLearning:
		options := recommend.Options{Clock: clock.now, Seed: seed, DisableLearning: name == VariantNoLearning}
		return engineRanker{engine: recommend.NewWithOptions(database, media.NewRepository(database), options)}, nil
	case VariantNewest:
		return newestRanker{database: database}, nil
	case VariantRandom:
		return randomRanker{database: database, random: rand.New(rand.NewSource(seed))}, nil
	}
	return nil, fmt.Errorf("unknown variant %q (want %s, %s, %s or %s)", name, VariantDefault, VariantNoLearning, VariantNewest, VariantRandom)
}

// sceneTraits are the site and tags of the scenes of a ranking, for coverage.
type sceneTraits struct {
	site string
	tags []string
}

// loadTraits reads the first observing plugin and the tags of the scenes.
func loadTraits(ctx context.Context, database *db.DB, sceneIDs []string) (map[string]sceneTraits, error) {
	out := map[string]sceneTraits{}
	if len(sceneIDs) == 0 {
		return out, nil
	}
	rows, err := database.Query(ctx,
		`SELECT scene.id, scene.tags,
		   (SELECT plugin FROM observation WHERE observation.target = scene.id ORDER BY observed_at LIMIT 1) AS site
		 FROM scene WHERE scene.id IN (SELECT value FROM json_each($ids))`, db.Vars{"ids": sceneIDs})
	if err != nil {
		return nil, fmt.Errorf("scene traits: %w", err)
	}
	for _, row := range rows {
		traits := sceneTraits{site: asString(row["site"])}
		if tags, isList := row["tags"].([]any); isList {
			for _, tag := range tags {
				traits.tags = append(traits.tags, asString(tag))
			}
		}
		out[asString(row["id"])] = traits
	}
	return out, nil
}

// sortedKeys lists a map's keys in order.
func sortedKeys[Value any](values map[string]Value) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
