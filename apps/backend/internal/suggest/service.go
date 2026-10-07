// Package suggest is the search suggestion engine behind the search box:
// recent searches, tags, performers and studios (names and aliases), and
// completions mined from scene titles, ranked by match quality, popularity
// and the user's taste profile.
//
// Entity names and title phrases live in an in-memory index built on first
// use and rebuilt in the background every indexRefreshInterval, so a
// keystroke costs a couple of binary searches plus one small history read.
package suggest

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/recommend"
)

// Kind is what a suggestion stands for; the values match the GraphQL enum.
type Kind string

const (
	KindRecent    Kind = "RECENT"
	KindQuery     Kind = "QUERY"
	KindTag       Kind = "TAG"
	KindPerformer Kind = "PERFORMER"
	KindStudio    Kind = "STUDIO"
)

// Suggestion is one entry of the suggestion list.
type Suggestion struct {
	Kind Kind
	// Text is shown and searched for when the suggestion is chosen.
	Text string
	// EntityID is the tag/performer/studio id; empty for searches.
	EntityID string
	// ImageURL is the performer's or studio's picture; empty otherwise.
	ImageURL string
	// Detail is the short grey line under the text ("Tag · 128 videos").
	Detail string
	Score  float64
}

// TasteSource supplies the recommendation taste profile.
type TasteSource interface {
	Profile(ctx context.Context) (*recommend.Profile, error)
}

// BlockSource supplies the blocked tag/performer/studio ids.
type BlockSource interface {
	BlockedTargetIDSet(ctx context.Context) map[string]bool
}

// Service answers suggestion queries and keeps the search history. Safe for
// concurrent use.
type Service struct {
	database *db.DB
	taste    TasteSource
	blocks   BlockSource
	// now is the clock, swappable in tests.
	now func() time.Time

	// indexMutex guards index and rebuilding; buildMutex serializes builds.
	indexMutex sync.Mutex
	index      *index
	rebuilding bool
	buildMutex sync.Mutex

	profileMutex   sync.Mutex
	profile        *recommend.Profile
	profileBuiltAt time.Time
}

// New returns a suggestion service. taste may be nil (no taste affinity).
func New(database *db.DB, taste TasteSource, blocks BlockSource) *Service {
	return &Service{database: database, taste: taste, blocks: blocks, now: time.Now}
}

// Warm builds the index ahead of the first query.
func (s *Service) Warm(ctx context.Context) {
	if _, err := s.currentIndex(ctx); err != nil {
		log.Printf("suggestion index: %v", err)
	}
}

// Suggest returns up to limit suggestions for the typed query. An empty query
// yields recent searches followed by taste picks.
func (s *Service) Suggest(ctx context.Context, query string, limit int) ([]Suggestion, error) {
	limit = min(limit, maxLimit)
	if limit <= 0 {
		return []Suggestion{}, nil
	}
	built, err := s.currentIndex(ctx)
	if err != nil {
		return nil, err
	}
	request := &request{
		service: s,
		index:   built,
		now:     s.now(),
		profile: s.currentProfile(ctx),
		blocked: s.blockedIDs(ctx),
		limit:   limit,
	}
	typed := matchText(query)
	if typed == "" {
		return request.emptyQuery(ctx)
	}
	return request.typedQuery(ctx, typed)
}

// currentIndex returns the index, building it on first use and starting a
// background rebuild once it is stale (the stale index keeps serving).
func (s *Service) currentIndex(ctx context.Context) (*index, error) {
	s.indexMutex.Lock()
	current := s.index
	if current != nil && s.now().Sub(current.builtAt) > indexRefreshInterval && !s.rebuilding {
		s.rebuilding = true
		go s.rebuildInBackground()
	}
	s.indexMutex.Unlock()
	if current != nil {
		return current, nil
	}
	return s.rebuild(ctx, true)
}

// rebuildInBackground refreshes a stale index, logging failures.
func (s *Service) rebuildInBackground() {
	ctx, cancel := context.WithTimeout(context.Background(), indexBuildTimeout)
	defer cancel()
	if _, err := s.rebuild(ctx, false); err != nil {
		log.Printf("suggestion index rebuild: %v", err)
	}
	s.indexMutex.Lock()
	s.rebuilding = false
	s.indexMutex.Unlock()
}

// rebuild builds and stores a fresh index. With onlyIfMissing it returns the
// index another caller built meanwhile instead of building again.
func (s *Service) rebuild(ctx context.Context, onlyIfMissing bool) (*index, error) {
	s.buildMutex.Lock()
	defer s.buildMutex.Unlock()
	if onlyIfMissing {
		s.indexMutex.Lock()
		existing := s.index
		s.indexMutex.Unlock()
		if existing != nil {
			return existing, nil
		}
	}
	built, err := buildIndex(ctx, s.database, s.now())
	if err != nil {
		return nil, err
	}
	s.indexMutex.Lock()
	s.index = built
	s.indexMutex.Unlock()
	return built, nil
}

// currentProfile returns the taste profile, recomputed at most every
// profileCacheTTL. Failures degrade to an empty profile (no taste boost).
func (s *Service) currentProfile(ctx context.Context) *recommend.Profile {
	empty := &recommend.Profile{}
	if s.taste == nil {
		return empty
	}
	s.profileMutex.Lock()
	defer s.profileMutex.Unlock()
	if s.profile != nil && s.now().Sub(s.profileBuiltAt) < profileCacheTTL {
		return s.profile
	}
	profile, err := s.taste.Profile(ctx)
	if err != nil || profile == nil {
		log.Printf("suggestion taste profile: %v", err)
		return empty
	}
	s.profile = profile
	s.profileBuiltAt = s.now()
	return profile
}

// blockedIDs returns the blocked entity ids, read fresh so a block applies at once.
func (s *Service) blockedIDs(ctx context.Context) map[string]bool {
	if s.blocks == nil {
		return map[string]bool{}
	}
	blocked := s.blocks.BlockedTargetIDSet(ctx)
	if blocked == nil {
		return map[string]bool{}
	}
	return blocked
}
