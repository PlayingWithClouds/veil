package subscriptions

import (
	"strings"
	"sync"
	"time"
)

// Topic constants. Use "category:action" naming; suffix ":*" for wildcard subscriptions.
const (
	TopicSceneAdded     = "scene:added"
	TopicSceneUpdated   = "scene:updated"
	TopicSceneRemoved   = "scene:removed"
	TopicPerformerAdded = "performer:added"
	TopicStudioAdded    = "studio:added"
	TopicGalleryAdded   = "gallery:added"
	TopicGalleryRemoved = "gallery:removed"
	TopicJobUpdated     = "job:updated"
	// Per-media stream changes. Publish to TopicStreamsChanged + ":" + mediaID so
	// a movie/episode page can subscribe to only its own sources.
	TopicStreamsChanged = "streams:changed"
	// Per-scene related-list changes, published to TopicRelatedChanged + ":" +
	// sceneID after a visit links new related scenes.
	TopicRelatedChanged = "related:changed"
)

// StreamsChangedEvent is published when a media item's playback sources change
// (new streams ingested, or a speed-check updated verification/speed).
type StreamsChangedEvent struct {
	MediaID string `json:"mediaId"`
}

// StreamsTopic returns the per-media topic for stream-change events.
func StreamsTopic(mediaID string) string {
	return TopicStreamsChanged + ":" + mediaID
}

// RelatedTopic is the per-scene topic for related-list changes.
func RelatedTopic(sceneID string) string {
	return TopicRelatedChanged + ":" + sceneID
}

// MediaAddedEvent is published when a new media item is ingested.
type MediaAddedEvent struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
}

// MediaRemovedEvent is published when a media item is deleted.
type MediaRemovedEvent struct {
	ID string `json:"id"`
}

// EpisodeAddedEvent is published when a new episode is detected on a series refresh.
type EpisodeAddedEvent struct {
	MediaID string  `json:"mediaId"`
	Title   string  `json:"title"`
	Season  int     `json:"season"`
	Episode int     `json:"episode"`
	AirDate *string `json:"airDate,omitempty"`
}

// JobUpdatedEvent is published on job state transitions and download progress.
type JobUpdatedEvent struct {
	ID            string   `json:"id"`
	Kind          string   `json:"kind"`
	Status        string   `json:"status"`
	Error         string   `json:"error,omitempty"`
	UpdatedAt     string   `json:"updatedAt"`
	DownloadTitle string   `json:"downloadTitle,omitempty"`
	DownloadURL   string   `json:"downloadUrl,omitempty"`
	Progress      *float64 `json:"progress,omitempty"`
	BytesReceived *float64 `json:"bytesReceived,omitempty"`
	BytesTotal    *float64 `json:"bytesTotal,omitempty"`
}

func nowRFC3339() string { return time.Now().UTC().Format(time.RFC3339) }

// NowRFC3339 is exported for use in orchestrator without importing "time" again.
func NowRFC3339() string { return nowRFC3339() }

type Event struct {
	Topic string
	Data  any
}

type subscriber struct {
	ch chan Event
}

// Hub is an in-memory pub/sub broker. Subscribe patterns support exact topics,
// prefix wildcards ("movie:*"), and the global wildcard ("*").
type Hub struct {
	mu   sync.RWMutex
	subs map[string][]*subscriber
}

func NewHub() *Hub {
	return &Hub{subs: make(map[string][]*subscriber)}
}

// Subscribe registers for events matching pattern. Patterns:
//   - "movie:added"  — exact match
//   - "movie:*"      — any topic with prefix "movie:"
//   - "*"            — all topics
func (h *Hub) Subscribe(pattern string) (<-chan Event, func()) {
	s := &subscriber{ch: make(chan Event, 32)}
	h.mu.Lock()
	h.subs[pattern] = append(h.subs[pattern], s)
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		list := h.subs[pattern]
		for i, sub := range list {
			if sub == s {
				h.subs[pattern] = append(list[:i], list[i+1:]...)
				break
			}
		}
		h.mu.Unlock()
		close(s.ch)
	}
	return s.ch, cancel
}

// SubscribeMulti fans-in multiple pattern subscriptions into a single channel.
func (h *Hub) SubscribeMulti(patterns ...string) (<-chan Event, func()) {
	merged := make(chan Event, 64)
	cancels := make([]func(), 0, len(patterns))
	for _, p := range patterns {
		ch, cancel := h.Subscribe(p)
		cancels = append(cancels, cancel)
		go func(c <-chan Event) {
			for ev := range c {
				select {
				case merged <- ev:
				default:
				}
			}
		}(ch)
	}
	return merged, func() {
		for _, c := range cancels {
			c()
		}
	}
}

// Publish sends an event to all subscribers whose pattern matches the topic.
func (h *Hub) Publish(topic string, data any) {
	ev := Event{Topic: topic, Data: data}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for pattern, subs := range h.subs {
		if patternMatches(pattern, topic) {
			for _, s := range subs {
				select {
				case s.ch <- ev:
				default: // drop slow consumer
				}
			}
		}
	}
}

// patternMatches reports whether pattern matches topic.
func patternMatches(pattern, topic string) bool {
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, ":*") {
		return strings.HasPrefix(topic, strings.TrimSuffix(pattern, "*"))
	}
	return pattern == topic
}
