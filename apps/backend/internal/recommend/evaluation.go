package recommend

import (
	"math/rand"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/media"
)

// Options configure an engine that does not run on the wall clock with the
// production random source: the offline evaluation (cmd/recommend-eval).
type Options struct {
	// Clock replaces time.Now; the engine ranks as of the time it returns.
	Clock func() time.Time
	// Seed fixes the draws of the learned priors.
	Seed int64
	// DisableLearning ranks with the hand-set source priors and page quotas
	// whatever the logged impressions say.
	DisableLearning bool
}

// NewWithOptions returns an engine configured by options, for tools that
// replay the engine at past moments.
func NewWithOptions(database *db.DB, repo *media.Repository, options Options) *Engine {
	engine := New(database, repo)
	if options.Clock != nil {
		engine.now = options.Clock
	}
	engine.random = rand.New(rand.NewSource(options.Seed))
	engine.learningDisabled = options.DisableLearning
	return engine
}

// SessionGap is how long nothing else may play after a watch for it to have
// ended its session.
const SessionGap = sessionGap

// SessionEndMinimumSeconds is how much of a scene a session-ending watch must
// have reached.
const SessionEndMinimumSeconds = sessionEndMinimumSeconds
