package plugins

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// PluginSettingsProvider supplies per-plugin env vars at invocation time.
type PluginSettingsProvider interface {
	PluginEnv(pluginName string) []string
}

// AccountProvider supplies the session of the plugin's account on this
// device (nil when it has none), keeps the sessions plugins renew, and logs
// in again when a site refuses one.
type AccountProvider interface {
	AccountSession(pluginName string) *AccountArgs
	// SessionRenewed stores the session a plugin renewed during a call.
	SessionRenewed(pluginName, session string)
	// LoginAgain logs the plugin's account in again after its site wanted a
	// login; false when it can't (no stored login, or the login failed).
	LoginAgain(ctx context.Context, pluginName string) bool
}

// accountEnvVar is the env var a plugin finds its session in.
const accountEnvVar = "VEIL_ACCOUNT"

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
	accountProvider  AccountProvider
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

// SetAccountProvider wires in the source of plugin account sessions.
func (r *Runner) SetAccountProvider(provider AccountProvider) {
	r.accountProvider = provider
}

// accountEnv is the env var carrying the plugin's session, if it has one.
func (r *Runner) accountEnv(pluginName string) []string {
	if r.accountProvider == nil {
		return nil
	}
	session := r.accountProvider.AccountSession(pluginName)
	if session == nil {
		return nil
	}
	encoded, err := json.Marshal(session)
	if err != nil {
		return nil
	}
	return []string{accountEnvVar + "=" + string(encoded)}
}

type invocation struct {
	Capability string      `json:"capability"`
	Args       interface{} `json:"args,omitempty"`
}

type pluginResponse struct {
	Result json.RawMessage `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
	// NeedsLogin marks an error the site gave because it wants a login.
	NeedsLogin bool `json:"needs_login,omitempty"`
}

// ErrNeedsLogin is what a call the site refused for want of a login the
// backend couldn't renew is (errors.Is); errors.As finds the LoginNeeded
// naming the plugin.
var ErrNeedsLogin = errors.New("needs a login")

// LoginNeeded is the error of a call the site refused for want of a login:
// the user has to log into the plugin's site (or import its cookies).
type LoginNeeded struct {
	Plugin  string
	Message string
}

// Error says which plugin wants a login, and why.
func (err *LoginNeeded) Error() string {
	return fmt.Sprintf("plugin %q needs a login: %s", err.Plugin, err.Message)
}

// Is makes a LoginNeeded match ErrNeedsLogin.
func (err *LoginNeeded) Is(target error) bool {
	return target == ErrNeedsLogin
}

// sessionLinePrefix starts the line a plugin writes when it renewed its
// account's session during a call (SDK sessionFetch).
var sessionLinePrefix = []byte(`{"session"`)

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
		return nil, resp.err(p.Meta.Name)
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
// env vars, returning its raw stdout without session lines. When the site
// wanted a login, the account logs in again and the call runs once more.
func (r *Runner) execute(ctx context.Context, p *Plugin, call invocation, onLine func([]byte)) ([]byte, error) {
	stdout, err := r.executeOnce(ctx, p, call, onLine)
	if err != nil || !wantsLogin(stdout) || r.accountProvider == nil || call.Capability == string(CapabilityAccountLogin) {
		return stdout, err
	}
	if !r.accountProvider.LoginAgain(ctx, p.Meta.Name) {
		return stdout, nil
	}
	return r.executeOnce(ctx, p, call, onLine)
}

// executeOnce runs one invocation and keeps the session the plugin renewed.
func (r *Runner) executeOnce(ctx context.Context, p *Plugin, call invocation, onLine func([]byte)) ([]byte, error) {
	payload, err := json.Marshal(call)
	if err != nil {
		return nil, err
	}
	env := os.Environ()
	env = append(env, r.envVars...)
	if r.settingsProvider != nil {
		env = append(env, r.settingsProvider.PluginEnv(p.Meta.Name)...)
	}
	env = append(env, r.accountEnv(p.Meta.Name)...)
	stdout, err := r.executor.run(ctx, p, payload, env, withoutSessionLines(onLine))
	stdout, session := takeSession(stdout)
	if session != "" && r.accountProvider != nil {
		r.accountProvider.SessionRenewed(p.Meta.Name, session)
	}
	return stdout, err
}

// withoutSessionLines passes every line but session lines on to onLine.
func withoutSessionLines(onLine func([]byte)) func([]byte) {
	if onLine == nil {
		return nil
	}
	return func(line []byte) {
		if bytes.HasPrefix(bytes.TrimSpace(line), sessionLinePrefix) {
			return
		}
		onLine(line)
	}
}

// takeSession removes session lines from a plugin's output and returns the
// last session they carried, "" when there's none.
func takeSession(stdout []byte) ([]byte, string) {
	if !bytes.Contains(stdout, sessionLinePrefix) {
		return stdout, ""
	}
	var kept [][]byte
	session := ""
	for _, line := range bytes.Split(stdout, []byte("\n")) {
		renewed, isSession := sessionOf(line)
		if !isSession {
			kept = append(kept, line)
			continue
		}
		session = renewed
	}
	return bytes.Join(kept, []byte("\n")), session
}

// sessionOf reads a session line; false for any other line.
func sessionOf(line []byte) (string, bool) {
	if !bytes.HasPrefix(bytes.TrimSpace(line), sessionLinePrefix) {
		return "", false
	}
	var renewal struct {
		Session string `json:"session"`
	}
	if json.Unmarshal(line, &renewal) != nil {
		return "", false
	}
	return renewal.Session, true
}

// wantsLogin reports whether a call ended with the site wanting a login.
func wantsLogin(stdout []byte) bool {
	var response pluginResponse
	if json.Unmarshal(lastLine(stdout), &response) != nil {
		return false
	}
	return response.NeedsLogin
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
	if !p.Serves(capability) {
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
	if !p.Serves(cap) {
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

// IdentifyPerformer invokes "performer:identify" — the site's profile of a
// performer the library knows, nil when the plugin isn't sure it's them.
func (r *Runner) IdentifyPerformer(ctx context.Context, p *Plugin, args PerformerIdentifyArgs) (*Performer, error) {
	if !p.Meta.Has(CapabilityPerformerIdentify) {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, CapabilityPerformerIdentify)
	}
	raw, err := r.invoke(ctx, p, CapabilityPerformerIdentify, args)
	if err != nil {
		return nil, err
	}
	var result PerformerIdentifyResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result.Performer, nil
}

// FindForumPost invokes "forum:find" — one forum thread by URL, its comments
// in site order.
func (r *Runner) FindForumPost(ctx context.Context, p *Plugin, url string) (*ForumPost, error) {
	if !p.ServesForum() {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, CapabilityForumFind)
	}
	capability := p.forumCapability(CapabilityForumFind, CapabilityLegacyForumFind)
	raw, err := r.invoke(ctx, p, capability, map[string]string{"url": url})
	if err != nil {
		return nil, err
	}
	var result ForumPost
	return &result, json.Unmarshal(raw, &result)
}

// SearchForum invokes "forum:search" — forum threads matching a query, best
// first.
func (r *Runner) SearchForum(ctx context.Context, p *Plugin, args ForumSearchArgs) (*ForumSearchResult, error) {
	if !p.ServesForum() {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, CapabilityForumSearch)
	}
	capability := p.forumCapability(CapabilityForumSearch, CapabilityLegacyForumSearch)
	raw, err := r.invoke(ctx, p, capability, args)
	if err != nil {
		return nil, err
	}
	var result ForumSearchResult
	return &result, json.Unmarshal(raw, &result)
}

// call invokes a capability the plugin must serve and decodes its result.
func call[T any](ctx context.Context, r *Runner, p *Plugin, capability Capability, args any) (*T, error) {
	if !p.Serves(capability) {
		return nil, fmt.Errorf("plugin %q does not support %s", p.Meta.Name, capability)
	}
	raw, err := r.invoke(ctx, p, capability, args)
	if err != nil {
		return nil, err
	}
	var result T
	return &result, json.Unmarshal(raw, &result)
}

// Lookup invokes "lookup" — records matching a text or picture, best first.
func (r *Runner) Lookup(ctx context.Context, p *Plugin, args LookupArgs) (*LookupResult, error) {
	return call[LookupResult](ctx, r, p, CapabilityLookup, args)
}

// StoryChapter invokes "story:chapter" — one chapter's text by URL.
func (r *Runner) StoryChapter(ctx context.Context, p *Plugin, url string) (*StoryChapter, error) {
	return call[StoryChapter](ctx, r, p, CapabilityStoryChapter, ChapterArgs{URL: url})
}

// Login invokes "account:login" with the login view's values.
func (r *Runner) Login(ctx context.Context, p *Plugin, values map[string]string) (*AccountResult, error) {
	return call[AccountResult](ctx, r, p, CapabilityAccountLogin, LoginArgs{Values: values})
}

// CheckAccount invokes "account:check" — whether a session still works.
func (r *Runner) CheckAccount(ctx context.Context, p *Plugin, session string) (*AccountResult, error) {
	return call[AccountResult](ctx, r, p, CapabilityAccountCheck, CheckArgs{Session: session})
}

// RenderView invokes "view:render" — one of the plugin's declarative views.
func (r *Runner) RenderView(ctx context.Context, p *Plugin, args ViewArgs) (*View, error) {
	return call[View](ctx, r, p, CapabilityViewRender, args)
}

// SubmitView invokes "view:submit" — runs a view's action.
func (r *Runner) SubmitView(ctx context.Context, p *Plugin, args SubmitArgs) (*SubmitResult, error) {
	return call[SubmitResult](ctx, r, p, CapabilityViewSubmit, args)
}

// Find invokes "<entity>:find" — one full record by URL — and reads the NDJSON
// stream. Plugins emit one ScrapeResult per line: root entity first, then
// related standalone entities (performers, studio, tags, collection).
func (r *Runner) Find(ctx context.Context, p *Plugin, entity MediaType, url string) ([]*ScrapeResult, error) {
	cap := FindCapability(entity)
	if !p.Serves(cap) {
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
		if failure := scrapeLineError(line); failure.Error != "" {
			return nil, failure.err(p.Meta.Name)
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

// scrapeLineError reads the {"error": ...} line a find writes when it fails
// (the SDK's usual error reply); its Error is empty for a ScrapeResult line.
func scrapeLineError(line []byte) pluginResponse {
	var failure pluginResponse
	if json.Unmarshal(line, &failure) != nil {
		return pluginResponse{}
	}
	return failure
}

// err is the error a failed response stands for.
func (response pluginResponse) err(pluginName string) error {
	if response.NeedsLogin {
		return &LoginNeeded{Plugin: pluginName, Message: response.Error}
	}
	return fmt.Errorf("plugin %q: %s", pluginName, response.Error)
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
