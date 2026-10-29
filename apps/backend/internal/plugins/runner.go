package plugins

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
)

// PluginSettingsProvider supplies per-plugin env vars at invocation time.
type PluginSettingsProvider interface {
	PluginEnv(pluginName string) []string
}

// executor runs one plugin invocation in some JS runtime: the JSON request is
// the plugin's stdin, and whatever the plugin writes to stdout comes back.
type executor interface {
	// run returns everything the plugin wrote to stdout. onLine, when set,
	// is called with each complete stdout line as it is written (the
	// embedded runtime) or once the plugin ended (Bun).
	run(ctx context.Context, p *Plugin, stdin []byte, env []string, onLine func([]byte)) ([]byte, error)
	// install fetches the plugin's dependencies, where the runtime needs that.
	install(ctx context.Context, dir string) error
}

// Runner invokes plugin capabilities. Every call runs in a fresh runtime —
// fully isolated, no shared state between calls.
type Runner struct {
	executor         executor
	envVars          []string // global env vars passed to every plugin invocation
	settingsProvider PluginSettingsProvider
}

// NewRunner runs plugins as Bun subprocesses, straight from their TypeScript source.
func NewRunner() (*Runner, error) {
	bun, err := newBunExecutor()
	if err != nil {
		return nil, err
	}
	return &Runner{executor: bun}, nil
}

// NewEmbeddedRunner runs pre-bundled plugins in an in-process JS engine, for
// hosts without Bun (the phone app). See embedded.go.
func NewEmbeddedRunner() *Runner {
	return &Runner{executor: newEmbeddedExecutor()}
}

// SetEnv replaces the global env vars forwarded to all plugin subprocesses.
func (r *Runner) SetEnv(vars []string) {
	r.envVars = vars
}

// SetSettingsProvider wires in the per-plugin settings source.
func (r *Runner) SetSettingsProvider(p PluginSettingsProvider) {
	r.settingsProvider = p
}

type invocation struct {
	Capability string      `json:"capability"`
	Args       interface{} `json:"args,omitempty"`
}

type pluginResponse struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

func (r *Runner) invoke(ctx context.Context, p *Plugin, cap Capability, args interface{}) (json.RawMessage, error) {
	return r.invokeStreaming(ctx, p, cap, args, nil)
}

// invokeStreaming is invoke with a callback for the lines a plugin writes
// before its response (partial list items). The response is the last line.
func (r *Runner) invokeStreaming(ctx context.Context, p *Plugin, cap Capability, args any, onLine func([]byte)) (json.RawMessage, error) {
	stdout, err := r.execute(ctx, p, invocation{Capability: string(cap), Args: args}, onLine)
	if err != nil {
		return nil, err
	}

	var resp pluginResponse
	if err := json.Unmarshal(lastLine(stdout), &resp); err != nil {
		return nil, fmt.Errorf("plugin %q bad response: %w\nraw: %s", p.Meta.Name, err, stdout)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf("plugin %q: %s", p.Meta.Name, resp.Error)
	}
	return resp.Result, nil
}

// loadMeta invokes the plugin with the "meta" capability to read its metadata.
func (r *Runner) loadMeta(ctx context.Context, dir, entryPoint string) (PluginMeta, error) {
	stub := &Plugin{Dir: dir, EntryPoint: entryPoint, Meta: PluginMeta{Name: dir}}
	raw, err := r.invoke(ctx, stub, "meta", nil)
	if err != nil {
		return PluginMeta{}, err
	}
	var meta PluginMeta
	return meta, json.Unmarshal(raw, &meta)
}

// execute runs one invocation of the plugin with the global and per-plugin
// env vars, returning its raw stdout.
func (r *Runner) execute(ctx context.Context, p *Plugin, call invocation, onLine func([]byte)) ([]byte, error) {
	payload, err := json.Marshal(call)
	if err != nil {
		return nil, err
	}
	env := os.Environ()
	env = append(env, r.envVars...)
	if r.settingsProvider != nil {
		env = append(env, r.settingsProvider.PluginEnv(p.Meta.Name)...)
	}
	return r.executor.run(ctx, p, payload, env, onLine)
}

// lastLine returns the last non-blank line of a plugin's output.
func lastLine(stdout []byte) []byte {
	trimmed := bytes.TrimSpace(stdout)
	if index := bytes.LastIndexByte(trimmed, '\n'); index >= 0 {
		return bytes.TrimSpace(trimmed[index+1:])
	}
	return trimmed
}

// install fetches the plugin's dependencies inside its directory.
func (r *Runner) install(ctx context.Context, dir string) error {
	return r.executor.install(ctx, dir)
}

// List invokes "<entity>:list" — reference stubs for one entity type: a keyword
// search when args.Query is set, the paginated catalog otherwise.
func (r *Runner) List(ctx context.Context, p *Plugin, entity MediaType, args ListArgs) (*ItemsResult, error) {
	return r.items(ctx, p, ListCapability(entity), args)
}

// ListStream is List for a caller that shows results as they come: onItems
// gets each batch of items the plugin emits while it is still crawling
// (SDK emitItems), then the final result's items it hadn't sent, each item
// once. Calls to onItems never overlap. Returns the full result.
func (r *Runner) ListStream(ctx context.Context, p *Plugin, entity MediaType, args ListArgs, onItems func([]DiscoveredItem)) (*ItemsResult, error) {
	capability := ListCapability(entity)
	if !p.Meta.Has(capability) {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, capability)
	}
	stream := newItemStream(onItems)
	raw, err := r.invokeStreaming(ctx, p, capability, args, stream.line)
	stream.close()
	if err != nil {
		return nil, err
	}
	var result ItemsResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	stream.deliver(result.Items)
	return &result, nil
}

func (r *Runner) items(ctx context.Context, p *Plugin, cap Capability, args any) (*ItemsResult, error) {
	if !p.Meta.Has(cap) {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, cap)
	}
	raw, err := r.invokeStreaming(ctx, p, cap, args, nil)
	if err != nil {
		return nil, err
	}
	var result ItemsResult
	return &result, json.Unmarshal(raw, &result)
}

// Suggest invokes "search:suggest" — the site's own search-box completions for a
// partly typed query.
func (r *Runner) Suggest(ctx context.Context, p *Plugin, args SuggestArgs) (*SuggestResult, error) {
	if !p.Meta.Has(CapabilitySearchSuggest) {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, CapabilitySearchSuggest)
	}
	raw, err := r.invoke(ctx, p, CapabilitySearchSuggest, args)
	if err != nil {
		return nil, err
	}
	var result SuggestResult
	return &result, json.Unmarshal(raw, &result)
}

// IdentifyFace invokes "face:identify" — the performers a recognition service
// ranks as matches for the face in a picture, best first.
func (r *Runner) IdentifyFace(ctx context.Context, p *Plugin, args FaceIdentifyArgs) (*FaceIdentifyResult, error) {
	if !p.Meta.Has(CapabilityFaceIdentify) {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, CapabilityFaceIdentify)
	}
	raw, err := r.invoke(ctx, p, CapabilityFaceIdentify, args)
	if err != nil {
		return nil, err
	}
	var result FaceIdentifyResult
	return &result, json.Unmarshal(raw, &result)
}

// FindPost invokes "post:find" — one forum thread by URL, its comments in site order.
func (r *Runner) FindPost(ctx context.Context, p *Plugin, url string) (*ForumPost, error) {
	if !p.Meta.Has(CapabilityPostFind) {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, CapabilityPostFind)
	}
	raw, err := r.invoke(ctx, p, CapabilityPostFind, map[string]string{"url": url})
	if err != nil {
		return nil, err
	}
	var result ForumPost
	return &result, json.Unmarshal(raw, &result)
}

// SearchPosts invokes "post:search" — forum threads matching a query, best first.
func (r *Runner) SearchPosts(ctx context.Context, p *Plugin, args ForumSearchArgs) (*ForumSearchResult, error) {
	if !p.Meta.Has(CapabilityPostSearch) {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, CapabilityPostSearch)
	}
	raw, err := r.invoke(ctx, p, CapabilityPostSearch, args)
	if err != nil {
		return nil, err
	}
	var result ForumSearchResult
	return &result, json.Unmarshal(raw, &result)
}

// Find invokes "<entity>:find" — one full record by URL — and reads the NDJSON
// stream. Plugins emit one ScrapeResult per line: root entity first, then
// related standalone entities (performers, studio, tags, collection).
func (r *Runner) Find(ctx context.Context, p *Plugin, entity MediaType, url string) ([]*ScrapeResult, error) {
	cap := FindCapability(entity)
	if !p.Meta.Has(cap) {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, cap)
	}

	stdout, err := r.execute(ctx, p, invocation{Capability: string(cap), Args: FindArgs{URL: url}}, nil)
	if err != nil {
		return nil, err
	}

	var results []*ScrapeResult
	scanner := bufio.NewScanner(bytes.NewReader(stdout))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1 MB per line
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		if failure := scrapeLineError(line); failure != "" {
			return nil, fmt.Errorf("plugin %q: %s", p.Meta.Name, failure)
		}
		var item ScrapeResult
		if err := json.Unmarshal(line, &item); err != nil {
			return nil, fmt.Errorf("plugin %q bad scrape line: %w\nraw: %s", p.Meta.Name, err, line)
		}
		results = append(results, &item)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("plugin %q: read stdout: %w", p.Meta.Name, err)
	}
	return results, nil
}

// scrapeLineError is the message of a {"error": ...} line a find writes when it
// fails (the SDK's usual error reply); empty for a ScrapeResult line.
func scrapeLineError(line []byte) string {
	var failure struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(line, &failure) != nil {
		return ""
	}
	return failure.Error
}

func (r *Runner) Resolve(ctx context.Context, p *Plugin, url string) (*ResolveResult, error) {
	if !p.Meta.Has(CapabilityStreamResolve) {
		return nil, fmt.Errorf("plugin %q does not support stream:resolve", p.Meta.Name)
	}
	raw, err := r.invoke(ctx, p, CapabilityStreamResolve, map[string]string{"url": url})
	if err != nil {
		return nil, err
	}
	var result ResolveResult
	return &result, json.Unmarshal(raw, &result)
}

// CamList invokes "cam:list": the rooms a cam site has online now.
func (r *Runner) CamList(ctx context.Context, p *Plugin, args CamListArgs) (*CamListResult, error) {
	if !p.Meta.Has(CapabilityCamList) {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, CapabilityCamList)
	}
	raw, err := r.invoke(ctx, p, CapabilityCamList, args)
	if err != nil {
		return nil, err
	}
	var result CamListResult
	return &result, json.Unmarshal(raw, &result)
}

// CamStream invokes "cam:stream": a room's live HLS master playlist.
func (r *Runner) CamStream(ctx context.Context, p *Plugin, username string) (*ResolveResult, error) {
	if !p.Meta.Has(CapabilityCamStream) {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, CapabilityCamStream)
	}
	raw, err := r.invoke(ctx, p, CapabilityCamStream, CamStreamArgs{Username: username})
	if err != nil {
		return nil, err
	}
	var result ResolveResult
	return &result, json.Unmarshal(raw, &result)
}

func (r *Runner) Enrich(ctx context.Context, p *Plugin, args EnrichArgs) (*EnrichResult, error) {
	if !p.Meta.Has(CapabilityEnrich) {
		return nil, fmt.Errorf("plugin %q does not support enrich", p.Meta.Name)
	}
	raw, err := r.invoke(ctx, p, CapabilityEnrich, args)
	if err != nil {
		return nil, err
	}
	var result EnrichResult
	return &result, json.Unmarshal(raw, &result)
}
