package search

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/discovery"
	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/media"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/recommend"
	"github.com/playingwithclouds/veil/internal/searchrank"
)

// Handler serves live search results as Server-Sent Events.
//
// GET /api/search?q=<query>[&limit=<n>]
//
// Events:
//
//	event: matched
//	data: {"entities":[{"kind":"performer","id":"performer:x","name":"...","alias":false,"fuzzy":false}],"rewritten":"..."}
//
//	event: result
//	data: {"source":"db"|"<pluginName>","id":"...","score":1.7,"replaces":"<source url>","item":{...DiscoveredItem}}
//
//	event: done
//	data: {}
//
// matched comes first, only when the query names performers, studios or tags
// of the library (by name, alias or a near-miss spelling); rewritten is the
// query with the ones named by alias or misspelled spelled out, which is
// searched as well. Library results follow, ranked by bm25 over title,
// performers, studio and tags (with a typo-tolerant fallback when few match);
// plugin results stream in as each plugin finishes, beside the pages of the
// recognised performers and studios on the sites that can list them. Every
// result carries a relevance score (see package searchrank) that clients order
// by: streaming order is only arrival order. Copies of one video on several
// sites are collapsed: later copies that score no better are not sent, and one
// that scores better is sent with `replaces` naming the source URL of the copy
// it supersedes (possibly its own).
type Handler struct {
	ingestSvc *ingest.Service
	discovery *discovery.Service
	repo      *media.Repository
	taste     TasteSource
	entities  *entityDirectory
	history   *searchrank.History
	pages     EntityPages
}

// TasteSource provides the user's taste profile (recommend.Engine); nil
// disables the personal term of the score.
type TasteSource interface {
	Profile(ctx context.Context) (*recommend.Profile, error)
}

// EntityPages names the plugins able to list an entity's own page on their
// site (updater.Scheduler); nil disables listing those pages.
type EntityPages interface {
	PagePlugins(origin *media.EntityOrigin) []*plugins.Plugin
}

// New creates the search handler. taste may be nil.
func New(ingestSvc *ingest.Service, discoverySvc *discovery.Service, repo *media.Repository, taste TasteSource) *Handler {
	return &Handler{
		ingestSvc: ingestSvc, discovery: discoverySvc, repo: repo, taste: taste,
		entities: newEntityDirectory(ingestSvc),
	}
}

// SetHistory lets the ranking learn from the clicks the search page logged.
func (h *Handler) SetHistory(history *searchrank.History) { h.history = history }

// SetEntityPages lets a search list the site pages of recognised entities.
func (h *Handler) SetEntityPages(pages EntityPages) { h.pages = pages }

type resultEvent struct {
	Source string  `json:"source"`
	ID     string  `json:"id,omitempty"` // DB record ID for source:"db" results
	Score  float64 `json:"score"`
	// Replaces is the source URL of an earlier result this one supersedes.
	Replaces string                 `json:"replaces,omitempty"`
	Item     plugins.DiscoveredItem `json:"item"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "q required", http.StatusBadRequest)
		return
	}
	limit := 20
	if ls := r.URL.Query().Get("limit"); ls != "" {
		if n, err := strconv.Atoi(ls); err == nil && n > 0 {
			limit = n
		}
	}
	// An explicit offset makes the plugin phase paginate its listing — used
	// by the galleries feed for infinite scroll. Absent offset keeps the default
	// keyword-search behaviour untouched.
	offset := 0
	paginate := false
	if os := r.URL.Query().Get("offset"); os != "" {
		if n, err := strconv.Atoi(os); err == nil && n >= 0 {
			offset = n
			paginate = true
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ctx := r.Context()
	stream := &eventStream{writer: w, flusher: flusher}
	session := h.newSession(ctx, q, offset, stream)
	firstPage := !paginate || offset == 0
	if firstPage && len(session.matches) > 0 {
		stream.write("matched", session.matchedEvent())
	}

	// 1. Library results (immediate). Later pages of a paginated search list
	// the sites only: the library matches were sent with the first.
	if firstPage {
		session.sendLibrary(limit)
	}

	// 2. Plugin search (concurrent, results streamed per plugin, each ingested
	// as a stub). An optional `sources` param (comma-separated plugin names)
	// restricts which providers run; empty means every search-capable plugin.
	searchPlugins := h.discovery.Plugins(discovery.SplitSources(r.URL.Query().Get("sources")))
	pluginNames := make([]string, 0, len(searchPlugins))
	for _, plugin := range searchPlugins {
		pluginNames = append(pluginNames, plugin.Meta.Name)
	}
	// Record the search as a recommendation signal once the stream completes.
	defer h.logSearch(q, pluginNames, session)

	options := discovery.Options{Query: q, Limit: limit, Offset: offset, Paginate: paginate}
	var running sync.WaitGroup
	if firstPage {
		session.startExtras(&running, options, searchPlugins)
	}
	// Plugin failures are already logged; the stream just carries what worked.
	_ = h.discovery.Run(ctx, options, searchPlugins, session.pluginHit(""))
	running.Wait()
	if ctx.Err() != nil {
		return
	}

	stream.write("done", struct{}{})
}

// logSearch records the completed search and the scenes it sent as a
// recommendation signal, mirroring the GraphQL pluginSearch resolver. No-op
// without a repo. Runs in the background with its own context so it survives
// the closed SSE stream.
func (h *Handler) logSearch(query string, pluginNames []string, session *searchSession) {
	if h.repo == nil {
		return
	}
	session.mutex.Lock()
	count := session.resultCount
	session.mutex.Unlock()
	sceneIDs := session.resultSceneIDs()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), logSearchTimeout)
		defer cancel()
		if err := h.repo.LogSearch(ctx, query, pluginNames, count, sceneIDs); err != nil {
			log.Printf("search: log search %q: %v", query, err)
		}
	}()
}

const logSearchTimeout = 15 * time.Second

// blockedSceneIDSet returns the blocked scene ids as a lookup set.
func (h *Handler) blockedSceneIDSet(ctx context.Context) map[string]bool {
	if h.repo == nil {
		return nil
	}
	rids := h.repo.BlockedSceneRIDs(ctx)
	set := make(map[string]bool, len(rids))
	for i := range rids {
		set[rids[i].String()] = true
	}
	return set
}

// blockedTargetIDSet returns the blocked entity ids as a lookup set.
func (h *Handler) blockedTargetIDSet(ctx context.Context) map[string]bool {
	if h.repo == nil {
		return nil
	}
	return h.repo.BlockedTargetIDSet(ctx)
}

func recordID(m map[string]any) string {
	switch v := m["id"].(type) {
	case db.RecordID:
		return v.String()
	case string:
		return v
	}
	return ""
}
