package plugins

import (
	"context"
	"encoding/base64"
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

// TestFindReportsPluginError checks a find that answers {"error": ...} fails
// with that message instead of yielding an empty record.
func TestFindReportsPluginError(t *testing.T) {
	plugin := writeBundle(t, `
const { writeSync } = require("fs");
writeSync(1, JSON.stringify({ error: "post was removed" }) + "\n");
process.exit(0);
`)
	plugin.Meta.Capabilities = []Capability{FindCapability(MediaTypeScene)}
	_, err := NewEmbeddedRunner().Find(context.Background(), plugin, MediaTypeScene, "https://example.test/post/1")
	if err == nil || !strings.Contains(err.Error(), "post was removed") {
		t.Fatalf("err = %v", err)
	}
}

// TestEmbeddedRunnerMultipartUpload posts a FormData with a text field and a
// binary Blob, and a raw Uint8Array body, checking the server receives the bytes intact.
func TestEmbeddedRunnerMultipartUpload(t *testing.T) {
	binary := []byte{0, 1, 2, 0xff, 0xfe, '\r', '\n', 0x80}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		answer := map[string]any{"contentType": r.Header.Get("Content-Type")}
		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			file, header, err := r.FormFile("upload")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			content := make([]byte, 64)
			count, _ := file.Read(content)
			answer["file"] = content[:count]
			answer["filename"] = header.Filename
			answer["fileType"] = header.Header.Get("Content-Type")
			answer["field"] = r.FormValue("note")
		} else {
			raw := make([]byte, 64)
			count, _ := r.Body.Read(raw)
			answer["raw"] = raw[:count]
		}
		_ = json.NewEncoder(w).Encode(answer)
	}))
	defer server.Close()

	bundle := `
const { writeSync } = require("fs");
async function main() {
  const input = JSON.parse(await Bun.stdin.text());
  const bytes = new Uint8Array(input.args.bytes);
  const form = new FormData();
  form.append("note", "hello");
  form.append("upload", new Blob([bytes], { type: "image/jpeg" }), "face.jpg");
  const multipart = await (await fetch(input.args.url, { method: "POST", body: form })).json();
  const raw = await (await fetch(input.args.url, { method: "POST", body: bytes })).json();
  writeSync(1, JSON.stringify({ result: { multipart, raw } }) + "\n");
  process.exit(0);
}
main().catch((error) => { writeSync(1, JSON.stringify({ error: String(error) }) + "\n"); process.exit(0); });
`
	raw, err := NewEmbeddedRunner().invoke(context.Background(), writeBundle(t, bundle), "test", map[string]any{"url": server.URL, "bytes": []int{0, 1, 2, 255, 254, 13, 10, 128}})
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Multipart struct {
			File     []byte `json:"file"`
			Filename string `json:"filename"`
			FileType string `json:"fileType"`
			Field    string `json:"field"`
		} `json:"multipart"`
		Raw struct {
			Raw []byte `json:"raw"`
		} `json:"raw"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if string(result.Multipart.File) != string(binary) || result.Multipart.Filename != "face.jpg" ||
		result.Multipart.FileType != "image/jpeg" || result.Multipart.Field != "hello" {
		t.Errorf("multipart upload altered: %+v", result.Multipart)
	}
	if string(result.Raw.Raw) != string(binary) {
		t.Errorf("Uint8Array body altered: %v", result.Raw.Raw)
	}
}

// TestEmbeddedRunnerSetTimeout waits on setTimeout callbacks before answering.
func TestEmbeddedRunnerSetTimeout(t *testing.T) {
	bundle := `
const { writeSync } = require("fs");
const sleep = (milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds));
(async () => {
  const started = Date.now();
  await sleep(30);
  await sleep(30);
  writeSync(1, JSON.stringify({ result: { waited: Date.now() - started } }) + "\n");
  process.exit(0);
})();
`
	raw, err := NewEmbeddedRunner().invoke(context.Background(), writeBundle(t, bundle), "test", nil)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Waited int `json:"waited"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatal(err)
	}
	if result.Waited < 55 {
		t.Errorf("setTimeout returned early: waited %d ms", result.Waited)
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
	// The test is the owner here: FaceDigger needs its agreement accepted in settings.
	runner.SetEnv([]string{"FACEDIGGER_ACCEPT_AGREEMENT=true"})
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

// checkFaceIdentifyLive uploads the pornstarbyface plugin's fixture photo (a
// performer's own picture) and expects that performer back as the best match.
func checkFaceIdentifyLive(t *testing.T, ctx context.Context, runner *Runner, plugin *Plugin) {
	picture, err := os.ReadFile("../../plugins/pornstarbyface/tests/fixtures/face.jpg")
	if err != nil {
		t.Skip("no face fixture")
	}
	started := time.Now()
	result, err := runner.IdentifyFace(ctx, plugin, FaceIdentifyArgs{ImageBase64: base64.StdEncoding.EncodeToString(picture), MimeType: "image/jpeg"})
	if err != nil {
		t.Fatalf("face:identify: %v", err)
	}
	t.Logf("face:identify → %d candidates in %s", len(result.Candidates), time.Since(started))
	if len(result.Candidates) == 0 {
		t.Fatal("no candidates")
	}
	if plugin.Meta.Name == "pornstarbyface" && result.Candidates[0].Name != "Lexi Lore" {
		t.Fatalf("expected Lexi Lore first, got %+v", result.Candidates)
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

	if meta.Has(CapabilityCamList) {
		checkCamBundleLive(t, runner, plugin)
		return
	}
	if meta.Has(CapabilityFaceIdentify) {
		checkFaceIdentifyLive(t, ctx, runner, plugin)
		return
	}

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

// checkCamBundleLive lists a cam plugin's online rooms and resolves the first
// watchable one to its live stream URL.
func checkCamBundleLive(t *testing.T, runner *Runner, plugin *Plugin) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	listing, err := runner.CamList(ctx, plugin, CamListArgs{Limit: 5})
	if err != nil {
		t.Fatalf("cam:list: %v", err)
	}
	t.Logf("cam:list → %d rooms of %d", len(listing.Rooms), listing.Total)
	if len(listing.Rooms) == 0 {
		t.Fatal("no rooms")
	}
	for _, room := range listing.Rooms {
		if !room.Online {
			continue
		}
		stream, err := runner.CamStream(ctx, plugin, room.Username)
		if err != nil {
			t.Fatalf("cam:stream %s: %v", room.Username, err)
		}
		t.Logf("cam:stream %s → %s", room.Username, stream.URL)
		return
	}
	t.Fatal("no online room to resolve")
}
