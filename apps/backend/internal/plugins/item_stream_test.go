package plugins

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// streamingBundle emits two batches, the second repeating an item, then the
// full result, like a plugin using the SDK's ItemSink.
const streamingBundle = `
const fs = require("fs");
const item = (id) => ({ title: id, media_type: "scene", source_url: "https://x.test/" + id, external_id: id });
fs.writeSync(1, JSON.stringify({ items: [item("a"), item("b")] }) + "\n");
fs.writeSync(1, JSON.stringify({ items: [item("b"), item("c")] }) + "\n");
fs.writeSync(1, JSON.stringify({ result: { items: [item("a"), item("b"), item("c"), item("d")] } }) + "\n");
process.exit(0);
`

func TestListStreamForwardsBatchesOnce(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "plugin.js"), []byte(streamingBundle), 0o644); err != nil {
		t.Fatal(err)
	}
	plugin := &Plugin{Dir: dir, EntryPoint: "plugin.js", Meta: PluginMeta{Name: "streamer", Capabilities: []Capability{CapabilitySceneList}}}
	runner := &Runner{executor: newEmbeddedExecutor()}

	var batches []string
	result, err := runner.ListStream(context.Background(), plugin, MediaTypeScene, ListArgs{Query: "q"}, func(items []DiscoveredItem) {
		ids := make([]string, 0, len(items))
		for _, item := range items {
			ids = append(ids, item.ExternalID)
		}
		batches = append(batches, strings.Join(ids, ","))
	})
	if err != nil {
		t.Fatalf("list stream: %v", err)
	}
	if got := strings.Join(batches, " | "); got != "a,b | c | d" {
		t.Errorf("batches = %q, want %q", got, "a,b | c | d")
	}
	if len(result.Items) != 4 {
		t.Errorf("result has %d items, want 4", len(result.Items))
	}
	plain, err := runner.List(context.Background(), plugin, MediaTypeScene, ListArgs{Query: "q"})
	if err != nil || len(plain.Items) != 4 {
		t.Errorf("plain List after partial lines: %v, %+v", err, plain)
	}
}
