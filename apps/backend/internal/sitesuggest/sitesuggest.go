// Package sitesuggest asks the plugins' sites for their own search-box
// completions, so the search field can offer things the library doesn't hold
// yet. Sites are asked concurrently with a short time limit, and answers are
// cached briefly because a query is typed one letter at a time.
package sitesuggest

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/plugins"
)

const (
	// minimumQueryLength is the shortest typed text sites are asked about.
	minimumQueryLength = 2
	// siteTimeout is how long one site may take; a slower one is left out.
	siteTimeout = 800 * time.Millisecond
	// perSiteLimit is how many completions are read from each site.
	perSiteLimit = 8
	// cacheTTL is how long one site's answer to a query is reused.
	cacheTTL = 2 * time.Minute
	// cacheLimit bounds the cache; it is emptied when it grows past it.
	cacheLimit = 500
)

// Suggestion is a completion offered by one or more sites.
type Suggestion struct {
	Text string
	// Kind is "query", "performer", "tag" or "studio".
	Kind string
	// Sources are the names of the plugins that offered it.
	Sources []string
}

// suggestFunc asks one plugin's site for completions; plugins.Runner.Suggest in production.
type suggestFunc func(ctx context.Context, plugin *plugins.Plugin, args plugins.SuggestArgs) (*plugins.SuggestResult, error)

// Service collects site completions. Safe for concurrent use.
type Service struct {
	// sites lists the active plugins that offer search:suggest.
	sites   func() []*plugins.Plugin
	suggest suggestFunc
	now     func() time.Time

	mutex sync.Mutex
	cache map[string]cachedAnswer
}

// cachedAnswer is one site's answer to one query.
type cachedAnswer struct {
	suggestions []plugins.SiteSuggestion
	storedAt    time.Time
}

// New returns a service asking the active plugins that offer search:suggest.
func New(registry *plugins.Registry, runner *plugins.Runner) *Service {
	sites := func() []*plugins.Plugin { return registry.WithCapability(plugins.CapabilitySearchSuggest) }
	return &Service{sites: sites, suggest: runner.Suggest, now: time.Now, cache: map[string]cachedAnswer{}}
}

// Suggest returns up to limit completions for the typed query, collected from
// every active site that offers them. Completions several sites agree on come
// first; a site that is slow or fails is left out.
func (s *Service) Suggest(ctx context.Context, query string, limit int) []Suggestion {
	typed := names.Key(query)
	if len(typed) < minimumQueryLength || limit <= 0 {
		return []Suggestion{}
	}
	sites := s.sites()
	sort.Slice(sites, func(left, right int) bool { return sites[left].Meta.Name < sites[right].Meta.Name })
	answers := make([][]plugins.SiteSuggestion, len(sites))
	var group sync.WaitGroup
	for index, site := range sites {
		group.Add(1)
		go func() {
			defer group.Done()
			answers[index] = s.askSite(ctx, site, query, typed)
		}()
	}
	group.Wait()
	return merge(sites, answers, typed, limit)
}

// askSite returns one site's completions, from the cache when it answered this
// query recently.
func (s *Service) askSite(ctx context.Context, site *plugins.Plugin, query, typed string) []plugins.SiteSuggestion {
	key := site.Meta.Name + "\x00" + typed
	if cached, found := s.cached(key); found {
		return cached
	}
	siteContext, cancel := context.WithTimeout(ctx, siteTimeout)
	defer cancel()
	result, err := s.suggest(siteContext, site, plugins.SuggestArgs{Query: query, Limit: perSiteLimit})
	if err != nil {
		log.Printf("suggest %q via plugin %s: %v", query, site.Meta.Name, err)
		return nil
	}
	s.store(key, result.Suggestions)
	return result.Suggestions
}

// cached returns a stored answer that has not expired.
func (s *Service) cached(key string) ([]plugins.SiteSuggestion, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	answer, found := s.cache[key]
	if !found || s.now().Sub(answer.storedAt) > cacheTTL {
		return nil, false
	}
	return answer.suggestions, true
}

// store remembers an answer, emptying the cache first when it is full.
func (s *Service) store(key string, suggestions []plugins.SiteSuggestion) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if len(s.cache) >= cacheLimit {
		s.cache = map[string]cachedAnswer{}
	}
	s.cache[key] = cachedAnswer{suggestions: suggestions, storedAt: s.now()}
}

// merge joins the sites' answers into one list: a completion several sites
// offer counts once and ranks above single-site ones, then by how early the
// sites list it. The typed text itself is not a suggestion.
func merge(sites []*plugins.Plugin, answers [][]plugins.SiteSuggestion, typed string, limit int) []Suggestion {
	type entry struct {
		suggestion Suggestion
		bestRank   int
	}
	entries := map[string]*entry{}
	order := []string{}
	for index, answer := range answers {
		for rank, offered := range answer {
			key := names.Key(offered.Text)
			if key == "" || key == typed {
				continue
			}
			existing, found := entries[key]
			if !found {
				existing = &entry{suggestion: Suggestion{Text: offered.Text, Kind: kindOf(offered), Sources: []string{}}, bestRank: rank}
				entries[key] = existing
				order = append(order, key)
			}
			existing.suggestion.Sources = append(existing.suggestion.Sources, sites[index].Meta.Name)
			if rank < existing.bestRank {
				existing.bestRank = rank
			}
		}
	}
	sort.SliceStable(order, func(left, right int) bool {
		first, second := entries[order[left]], entries[order[right]]
		if len(first.suggestion.Sources) != len(second.suggestion.Sources) {
			return len(first.suggestion.Sources) > len(second.suggestion.Sources)
		}
		return first.bestRank < second.bestRank
	})
	out := make([]Suggestion, 0, min(limit, len(order)))
	for _, key := range order {
		if len(out) >= limit {
			break
		}
		out = append(out, entries[key].suggestion)
	}
	return out
}

// kindOf is the completion's kind, "query" when the site gave none.
func kindOf(offered plugins.SiteSuggestion) string {
	if offered.Kind == "" {
		return "query"
	}
	return offered.Kind
}
