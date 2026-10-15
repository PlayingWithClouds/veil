package plugins

import (
	"bytes"
	"encoding/json"
	"log"
)

// itemStreamBuffer is how many emitted lines may wait for onItems before the
// plugin's writes block.
const itemStreamBuffer = 64

// itemStream forwards the item batches a listing plugin emits to onItems on
// one goroutine of its own, so slow consumers (ingest) don't run on the
// plugin runtime's goroutine, and drops items already forwarded.
type itemStream struct {
	onItems func([]DiscoveredItem)
	lines   chan []byte
	done    chan struct{}
	// seen is only touched by the forwarding goroutine, and by deliver after
	// close has waited for it.
	seen map[string]bool
}

// partialItems is a line a plugin writes before its result (SDK emitItems).
type partialItems struct {
	Items []DiscoveredItem `json:"items"`
}

// newItemStream starts forwarding to onItems.
func newItemStream(onItems func([]DiscoveredItem)) *itemStream {
	stream := &itemStream{onItems: onItems, lines: make(chan []byte, itemStreamBuffer), done: make(chan struct{}), seen: map[string]bool{}}
	go stream.forward()
	return stream
}

// line takes one stdout line from the plugin. The result line (and anything
// that isn't an item batch) is left to the caller.
func (stream *itemStream) line(line []byte) {
	if !bytes.HasPrefix(bytes.TrimSpace(line), []byte(`{"items"`)) {
		return
	}
	stream.lines <- line
}

// forward decodes queued batches and hands them on until close.
func (stream *itemStream) forward() {
	defer close(stream.done)
	for line := range stream.lines {
		var batch partialItems
		if err := json.Unmarshal(line, &batch); err != nil {
			log.Printf("plugin item batch: %v", err)
			continue
		}
		stream.deliver(batch.Items)
	}
}

// close waits until every queued batch was handed on.
func (stream *itemStream) close() {
	close(stream.lines)
	<-stream.done
}

// deliver hands on the items not forwarded before.
func (stream *itemStream) deliver(items []DiscoveredItem) {
	fresh := make([]DiscoveredItem, 0, len(items))
	for _, item := range items {
		key := itemIdentity(item)
		if stream.seen[key] {
			continue
		}
		stream.seen[key] = true
		fresh = append(fresh, item)
	}
	if len(fresh) > 0 {
		stream.onItems(fresh)
	}
}

// itemIdentity keys an item like the SDK's ItemSink: source URL, else
// external id.
func itemIdentity(item DiscoveredItem) string {
	if item.SourceURL != "" {
		return item.SourceURL
	}
	return item.ExternalID
}
