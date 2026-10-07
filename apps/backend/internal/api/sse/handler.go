package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/playingwithclouds/veil/internal/subscriptions"
)

// Handler streams hub events as Server-Sent Events.
//
// GET /api/events?topics=movie:added,show:*,job:updated
//
// If topics is omitted, subscribes to "*" (all events).
// Each event is sent as:
//
//	event: <topic>
//	data: <json>
type Handler struct {
	hub *subscriptions.Hub
}

func New(hub *subscriptions.Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

	patterns := parseTopics(r.URL.Query().Get("topics"))

	merged := make(chan subscriptions.Event, 64)
	cancels := make([]func(), 0, len(patterns))
	for _, p := range patterns {
		ch, cancel := h.hub.Subscribe(p)
		cancels = append(cancels, cancel)
		go func(c <-chan subscriptions.Event) {
			for ev := range c {
				select {
				case merged <- ev:
				default:
				}
			}
		}(ch)
	}
	defer func() {
		for _, c := range cancels {
			c()
		}
	}()

	// Establish connection.
	fmt.Fprintf(w, ": connected topics=%s\n\n", strings.Join(patterns, ","))
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-merged:
			if !ok {
				return
			}
			data, err := json.Marshal(ev.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Topic, data)
			flusher.Flush()
		}
	}
}

func parseTopics(raw string) []string {
	if raw == "" {
		return []string{"*"}
	}
	var out []string
	for _, t := range strings.Split(raw, ",") {
		if t = strings.TrimSpace(t); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}
