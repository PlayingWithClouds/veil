package search

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/discovery"
	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/media"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// Handler serves live search results as Server-Sent Events.
//
// GET /api/search?q=<query>[&limit=<n>]
//
// Events:
//
//	event: result
//	data: {"source":"db"|"<pluginName>","item":{...DiscoveredItem}}
//
//	event: done
//	data: {}
//
// DB results arrive first; plugin results stream in as each plugin finishes.
type Handler struct {
	ingestSvc *ingest.Service
	discovery *discovery.Service
	repo      *media.Repository
}

// New creates the search handler.
func New(ingestSvc *ingest.Service, discoverySvc *discovery.Service, repo *media.Repository) *Handler {
	return &Handler{ingestSvc: ingestSvc, discovery: discoverySvc, repo: repo}
}

type resultEvent struct {
	Source string                 `json:"source"`
	ID     string                 `json:"id,omitempty"` // DB record ID for source:"db" results
	Item   plugins.DiscoveredItem `json:"item"`
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
	// An explicit offset makes the plugin phase paginate its listing — used by
	// the galleries feed for infinite scroll. Absent offset keeps the default
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
	resultCount := 0
	write := func(eventType string, v any) {
		data, err := json.Marshal(v)
		if err != nil {
			return
		}
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, data)
		flusher.Flush()
	}

	// Blocklist sets: scenes carrying a blocked entity, and the blocked entities
	// themselves (to drop blocked performers/studios from results).
	blockedScenes := h.blockedSceneIDSet(ctx)
	blockedTargets := h.blockedTargetIDSet(ctx)

	// 1. DB results (immediate).
	for _, mt := range []plugins.MediaType{plugins.MediaTypeScene, plugins.MediaTypeGallery, plugins.MediaTypePerformer, plugins.MediaTypeStudio} {
		recs, err := h.ingestSvc.ListByTable(ctx, string(mt), q, limit, 0)
		if err != nil {
			log.Printf("search db(%s, %q): %v", mt, q, err)
			continue
		}
		for _, rec := range recs {
			id := recordID(rec)
			if blockedScenes[id] || blockedTargets[id] {
				continue
			}
			sourceURL, _ := rec["source_url"].(string)
			// Performers and studios label with `name`; scenes/galleries with `title`.
			title, _ := rec["title"].(string)
			if title == "" {
				title, _ = rec["name"].(string)
			}
			externalID, _ := rec["external_id"].(string)
			posterPath, _ := rec["poster_path"].(string)
			write("result", resultEvent{
				Source: "db",
				ID:     id,
				Item: plugins.DiscoveredItem{
					Title:      title,
					MediaType:  mt,
					SourceURL:  sourceURL,
					ExternalID: externalID,
					PosterPath: posterPath,
				},
			})
			resultCount++
		}
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
	defer h.logSearch(q, pluginNames, &resultCount)

	options := discovery.Options{Query: q, Limit: limit, Offset: offset, Paginate: paginate}
	// Plugin failures are already logged; the stream just carries what worked.
	_ = h.discovery.Run(ctx, options, searchPlugins, func(hit discovery.Hit) {
		write("result", resultEvent{Source: hit.Plugin, ID: hit.ID, Item: hit.Item})
		resultCount++
	})
	if ctx.Err() != nil {
		return
	}

	write("done", struct{}{})
}

// logSearch records the completed search as a recommendation signal, mirroring
// the GraphQL pluginSearch resolver. No-op without a repo. Runs in the
// background with its own context so it survives the closed SSE stream.
func (h *Handler) logSearch(query string, pluginNames []string, resultCount *int) {
	if h.repo == nil {
		return
	}
	count := *resultCount
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), logSearchTimeout)
		defer cancel()
		if err := h.repo.LogSearch(ctx, query, pluginNames, count); err != nil {
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
