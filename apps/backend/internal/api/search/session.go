package search

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/playingwithclouds/veil/internal/discovery"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/recommend"
	"github.com/playingwithclouds/veil/internal/searchrank"
)

// eventStream writes Server-Sent Events; safe for concurrent use.
type eventStream struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	mutex   sync.Mutex
}

// write sends one event with a JSON body, ignoring values that cannot be encoded.
func (stream *eventStream) write(eventType string, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	fmt.Fprintf(stream.writer, "event: %s\ndata: %s\n\n", eventType, data)
	stream.flusher.Flush()
}

// searchSession is one running search: what the query names, how its
// results are scored, and the lists that feed them.
type searchSession struct {
	handler *Handler
	ctx     context.Context
	query   string
	stream  *eventStream
	profile *recommend.Profile

	// matches are the performers, studios and tags the query names (blocked
	// ones left out); rewritten is the query with the ones named by alias or
	// misspelled spelled out, empty when nothing needed it.
	matches   []searchrank.EntityMatch
	rewritten string

	blockedScenes, blockedTargets map[string]bool

	// mutex guards ranker and resultCount: lists feed them concurrently.
	mutex       sync.Mutex
	ranker      *searchrank.Ranker
	resultCount int
}

// newSession recognises the entities in the query and configures the ranker:
// text variants, the taste weight for the query's specificity, and what the
// user's earlier clicks say.
func (h *Handler) newSession(ctx context.Context, query string, offset int, stream *eventStream) *searchSession {
	session := &searchSession{
		handler:        h,
		ctx:            ctx,
		query:          query,
		stream:         stream,
		profile:        h.tasteProfile(ctx),
		blockedScenes:  h.blockedSceneIDSet(ctx),
		blockedTargets: h.blockedTargetIDSet(ctx),
		ranker:         searchrank.New(query, offset),
	}
	session.matches = session.recognise()
	session.rewritten = searchrank.Rewrite(query, session.matches)
	session.ranker.SetVariants(session.rewritten)
	session.ranker.SetPersonalWeight(searchrank.PersonalWeight(searchrank.Specificity(session.queryProfile())))
	session.ranker.SetLearned(h.learn(ctx, query))
	return session
}

// recognise finds the library entities the query names, dropping blocked ones.
func (s *searchSession) recognise() []searchrank.EntityMatch {
	index := s.handler.entities.Index(s.ctx)
	if index == nil {
		return nil
	}
	var kept []searchrank.EntityMatch
	for _, match := range index.Find(s.query) {
		if !s.blockedTargets[match.ID] {
			kept = append(kept, match)
		}
	}
	return kept
}

// queryProfile describes the query's shape for searchrank.Specificity.
func (s *searchSession) queryProfile() searchrank.QueryProfile {
	wordCount := len(searchrank.QueryWords(s.query))
	profile := searchrank.QueryProfile{
		WordCount:      wordCount,
		EntityCoverage: searchrank.Coverage(s.matches, wordCount, searchrank.KindPerformer, searchrank.KindStudio),
		TagCoverage:    searchrank.Coverage(s.matches, wordCount, searchrank.KindTag),
		LibraryMatches: -1,
	}
	count, err := s.handler.ingestSvc.CountSceneMatches(s.ctx, s.query, searchrank.LibraryManyMatches)
	if err == nil {
		profile.LibraryMatches = count
	}
	return profile
}

// learn reads what the search history says about the query; nothing without a
// history or on failure.
func (h *Handler) learn(ctx context.Context, query string) searchrank.Learned {
	if h.history == nil {
		return searchrank.Learned{}
	}
	learned, err := h.history.Learn(ctx, query)
	if err != nil {
		log.Printf("search: history: %v", err)
	}
	return learned
}

// matchedEvent is the body of the `matched` event.
type matchedEvent struct {
	Entities  []matchedEntity `json:"entities"`
	Rewritten string          `json:"rewritten,omitempty"`
}

// matchedEntity is one recognised entity as the client shows it.
type matchedEntity struct {
	Kind searchrank.EntityKind `json:"kind"`
	ID   string                `json:"id"`
	Name string                `json:"name"`
	// Alias is true when the query used an alias, Fuzzy when it only came
	// close to the name.
	Alias bool `json:"alias,omitempty"`
	Fuzzy bool `json:"fuzzy,omitempty"`
}

// matchedEvent describes the recognised entities for the client.
func (s *searchSession) matchedEvent() matchedEvent {
	event := matchedEvent{Entities: make([]matchedEntity, 0, len(s.matches)), Rewritten: s.rewritten}
	for _, match := range s.matches {
		event.Entities = append(event.Entities, matchedEntity{
			Kind: match.Kind, ID: match.ID, Name: match.Name, Alias: match.ViaAlias, Fuzzy: match.Fuzzy,
		})
	}
	return event
}

// annotate fills the terms of a hit that need the library: taste and entity
// credit, from the record the hit was stored as. pageEntityID, when set, is
// the entity whose site page listed the hit, which credits it by definition.
func (s *searchSession) annotate(hit *searchrank.Hit, pageEntityID string) {
	record := s.handler.storedRecord(s.ctx, *hit)
	hit.Personal = s.handler.personalAffinity(s.profile, *hit, record)
	if len(s.matches) == 0 {
		return
	}
	credited := creditedIDs(record)
	if pageEntityID != "" {
		credited[pageEntityID] = true
	}
	hit.EntityCredit = searchrank.EntityCredit(s.matches, credited, hit.ID, hit.Item.Title)
}

// send scores a hit, folds it into the copies seen and streams it unless a
// better copy was sent already.
func (s *searchSession) send(source, id string, hit searchrank.Hit) {
	s.mutex.Lock()
	verdict := s.ranker.Add(hit)
	if verdict.Emit {
		s.resultCount++
	}
	s.mutex.Unlock()
	if !verdict.Emit {
		return
	}
	s.stream.write("result", resultEvent{Source: source, ID: id, Score: verdict.Score, Replaces: verdict.Replaces, Item: hit.Item})
}

// pluginHit returns the callback taking a plugin's ingested results into the
// list named list (empty for the plain keyword search). pageEntityID is the
// entity whose site page the list is, if any.
func (s *searchSession) pluginHit(list string) func(discovery.Hit) {
	return s.pageHit(list, "")
}

// pageHit is pluginHit for a list that is an entity's site page.
func (s *searchSession) pageHit(list, pageEntityID string) func(discovery.Hit) {
	return func(found discovery.Hit) {
		hit := searchrank.Hit{Source: found.Plugin, ID: found.ID, Item: found.Item, List: list}
		s.annotate(&hit, pageEntityID)
		s.send(found.Plugin, found.ID, hit)
	}
}

// sendLibrary streams the library's matches: the query as typed and its
// rewrite, then, when the exact index found few scenes, the typo-tolerant
// candidates.
func (s *searchSession) sendLibrary(limit int) {
	queries := []string{s.query}
	if s.rewritten != "" {
		queries = append(queries, s.rewritten)
	}
	emitter := libraryEmitter{
		blockedScenes:  s.blockedScenes,
		blockedTargets: s.blockedTargets,
		annotate:       func(hit *searchrank.Hit) { s.annotate(hit, "") },
		send:           s.send,
	}
	sentScenes := s.handler.sendLibraryHits(s.ctx, queries, limit, emitter)
	if len(sentScenes) < searchrank.FuzzyFallbackBelow {
		s.handler.sendFuzzyScenes(s.ctx, s.query, limit, sentScenes, emitter)
	}
}

// startExtras starts, beside the plain keyword search, the rewritten query on
// the same plugins and the site pages of the recognised performers and
// studios; running tracks them.
func (s *searchSession) startExtras(running *sync.WaitGroup, options discovery.Options, searchPlugins []*plugins.Plugin) {
	if s.rewritten != "" {
		rewrittenOptions := options
		rewrittenOptions.Query = s.rewritten
		running.Go(func() {
			_ = s.handler.discovery.Run(s.ctx, rewrittenOptions, searchPlugins, s.pluginHit("rewrite"))
		})
	}
	for _, match := range s.pageEntities() {
		running.Go(func() { s.listPage(match, options.Limit) })
	}
}

// pageEntities picks the recognised entities whose site page is listed: up
// to searchrank.MaxEntityPages exactly spelled performers and studios, plus a
// tag only when it is the whole query (otherwise the keyword search already
// covers the topic).
func (s *searchSession) pageEntities() []searchrank.EntityMatch {
	if s.handler.pages == nil || s.handler.repo == nil {
		return nil
	}
	wordCount := len(searchrank.QueryWords(s.query))
	var picked []searchrank.EntityMatch
	for _, kind := range []searchrank.EntityKind{searchrank.KindPerformer, searchrank.KindStudio, searchrank.KindTag} {
		for _, match := range s.matches {
			wholeQuery := match.End-match.Start == wordCount
			if match.Kind != kind || match.Fuzzy || (kind == searchrank.KindTag && !wholeQuery) {
				continue
			}
			if len(picked) < searchrank.MaxEntityPages {
				picked = append(picked, match)
			}
		}
	}
	return picked
}

// listPage lists the entity's own page on the first plugin able to, handing
// its scenes on as a list of their own, credited to the entity.
func (s *searchSession) listPage(match searchrank.EntityMatch, limit int) {
	origin, err := s.handler.repo.EntityOrigin(s.ctx, match.ID)
	if err != nil {
		log.Printf("search: entity origin %s: %v", match.ID, err)
		return
	}
	onHit := s.pageHit("page:"+match.ID, match.ID)
	for _, plugin := range s.handler.pages.PagePlugins(origin) {
		listed := 0
		err := s.handler.discovery.ListPage(s.ctx, plugin, origin.SourceURL, limit, func(found discovery.Hit) {
			listed++
			onHit(found)
		})
		if err == nil || listed > 0 {
			return
		}
		log.Printf("search: page of %s via %s: %v", match.Name, plugin.Meta.Name, err)
	}
}
