package plugins

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/netdns"
)

// testBundle mimics what the SDK's runPlugin does in a bundle: read the
// request from Bun.stdin, fetch, answer through fs.writeSync, then exit.
const testBundle = `
const { writeSync } = require("fs");
async function main() {
  const input = JSON.parse(await Bun.stdin.text());
  const response = await fetch(input.args.url + "?q=" + encodeURIComponent(new URL(input.args.url).host), {
    headers: { "X-Plugin": process.env.PLUGIN_TOKEN, "Accept-Encoding": "br" },
    signal: AbortSignal.timeout(5000),
  });
  const data = await response.json();
  writeSync(1, JSON.stringify({ result: { status: response.status, type: response.headers.get("content-type"), data } }) + "\n");
  process.exit(0);
  writeSync(1, "written after exit\n");
}
main().catch((error) => {
  writeSync(1, JSON.stringify({ error: String(error) }) + "\n");
  process.exit(0);
});
`

// writeBundle stores a bundle as a plugin directory and returns its Plugin.
func writeBundle(t *testing.T, source string) *Plugin {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "plugin.js"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	return &Plugin{Dir: dir, EntryPoint: "plugin.js", Meta: PluginMeta{Name: "test"}}
}

// TestEmbeddedRunnerRoundTrip runs a bundle through the embedded runtime:
// stdin, env, fetch with headers, JSON body, stdout, and exit.
func TestEmbeddedRunnerRoundTrip(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"query":    r.URL.Query().Get("q"),
			"token":    r.Header.Get("X-Plugin"),
			"encoding": r.Header.Get("Accept-Encoding"),
		})
	}))
	defer server.Close()

	runner := NewEmbeddedRunner()
	runner.SetEnv([]string{"PLUGIN_TOKEN=secret"})
	raw, err := runner.invoke(context.Background(), writeBundle(t, testBundle), "test", map[string]string{"url": server.URL})
	if err != nil {
		t.Fatal(err)
	}

	var result struct {
		Status int               `json:"status"`
		Type   string            `json:"type"`
		Data   map[string]string `json:"data"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != 200 || result.Type != "application/json" {
		t.Errorf("status %d type %q", result.Status, result.Type)
	}
	if result.Data["token"] != "secret" {
		t.Errorf("env header not sent: %v", result.Data)
	}
	if !strings.HasPrefix(result.Data["query"], "127.0.0.1:") {
		t.Errorf("URL global or query encoding broken: %v", result.Data)
	}
	if result.Data["encoding"] == "br" {
		t.Errorf("plugin's Accept-Encoding should be left to the transport")
	}
}

// TestEmbeddedRunnerTimeout stops a plugin that never finishes.
func TestEmbeddedRunnerTimeout(t *testing.T) {
	plugin := writeBundle(t, `while (true) {}`)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := NewEmbeddedRunner().invoke(ctx, plugin, "meta", nil); err == nil {
		t.Fatal("expected the busy loop to be interrupted")
	}
}

// TestEmbeddedBundlesLive runs the real plugin bundles against the live sites.
// Opt-in: VEIL_EMBEDDED_BUNDLES=<dir from apps/backend/scripts/bundle-plugins.ts>.
func TestEmbeddedBundlesLive(t *testing.T) {
	bundles := os.Getenv("VEIL_EMBEDDED_BUNDLES")
	if bundles == "" {
		t.Skip("VEIL_EMBEDDED_BUNDLES not set")
	}
	query := os.Getenv("VEIL_EMBEDDED_QUERY")
	if query == "" {
		query = "blonde"
	}
	netdns.Configure(os.Getenv("DNS_SERVERS"), "")
	runner := NewEmbeddedRunner()
	entries, err := os.ReadDir(bundles)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			checkBundleLive(t, runner, filepath.Join(bundles, entry.Name()), query)
		})
	}
}

// checkBundleLive loads one bundle's meta, searches, opens the first hit and
// resolves its page to a playable stream, logging how long each step took.
func checkBundleLive(t *testing.T, runner *Runner, dir, query string) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	started := time.Now()
	meta, err := runner.loadMeta(ctx, dir, "plugin.js")
	if err != nil {
		t.Fatalf("meta: %v", err)
	}
	t.Logf("meta %s in %s", meta.Name, time.Since(started))
	if meta.RequiresSolver {
		t.Skip("needs FlareSolverr")
	}
	plugin := &Plugin{Dir: dir, EntryPoint: "plugin.js", Meta: meta}

	entity := MediaTypeScene
	if !meta.Has(ListCapability(entity)) {
		entity = MediaTypeGallery
	}
	started = time.Now()
	items, err := runner.List(ctx, plugin, entity, ListArgs{Query: query, Limit: 5})
	if err != nil {
		t.Fatalf("%s list: %v", entity, err)
	}
	t.Logf("%s:list %q → %d items in %s", entity, query, len(items.Items), time.Since(started))
	if len(items.Items) == 0 {
		t.Fatal("no items")
	}

	started = time.Now()
	results, err := runner.Find(ctx, plugin, entity, items.Items[0].SourceURL)
	if err != nil {
		t.Fatalf("%s find: %v", entity, err)
	}
	t.Logf("%s:find → %d results in %s", entity, len(results), time.Since(started))
	compareWithBun(t, ctx, plugin, entity, query, items, results)
	if entity != MediaTypeScene || !meta.Has(CapabilityStreamResolve) {
		return
	}
	started = time.Now()
	resolved, err := runner.Resolve(ctx, plugin, items.Items[0].SourceURL)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	t.Logf("stream:resolve → %s in %s", resolved.MimeType, time.Since(started))
}

// compareWithBun runs the same list and find through Bun on the plugin's
// source and fails on any difference, so the native HTML parser is held to
// node-html-parser's results. Sites reorder listings between requests, so
// list items are compared as sets by source URL.
func compareWithBun(t *testing.T, ctx context.Context, plugin *Plugin, entity MediaType, query string, items *ItemsResult, results []*ScrapeResult) {
	sourceDir := filepath.Join("..", "..", "plugins", filepath.Base(plugin.Dir))
	if _, err := os.Stat(filepath.Join(sourceDir, "package.json")); err != nil {
		t.Logf("no plugin source at %s, skipping the Bun comparison", sourceDir)
		return
	}
	bun, err := NewRunner()
	if err != nil {
		t.Logf("skipping the Bun comparison: %v", err)
		return
	}
	source := &Plugin{Dir: sourceDir, EntryPoint: "index.ts", Meta: plugin.Meta}

	bunItems, err := bun.List(ctx, source, entity, ListArgs{Query: query, Limit: 5})
	if err != nil {
		t.Fatalf("bun list: %v", err)
	}
	embeddedByURL := map[string]string{}
	for _, item := range items.Items {
		embeddedByURL[item.SourceURL] = mustJSON(t, item)
	}
	for _, item := range bunItems.Items {
		embedded, found := embeddedByURL[item.SourceURL]
		if found && embedded != mustJSON(t, item) {
			t.Errorf("list item differs\n bun:      %s\n embedded: %s", mustJSON(t, item), embedded)
		}
	}

	bunResults, err := bun.Find(ctx, source, entity, items.Items[0].SourceURL)
	if err != nil {
		t.Fatalf("bun find: %v", err)
	}
	if got, want := mustJSON(t, results), mustJSON(t, bunResults); got != want {
		t.Errorf("find differs\n bun:      %s\n embedded: %s", want, got)
	}
}

// mustJSON marshals value for comparison.
func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
