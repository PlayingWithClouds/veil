package stream

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/storage"
)

const (
	sessPrefix = "stream:sess:v1:"
	sessTTL    = 4 * time.Hour
)

type sessionData struct {
	Headers map[string]string `json:"headers"`
	// Live marks a live broadcast: nothing is written to the stream cache and
	// low-latency playlist tags are stripped (see serveLiveManifest).
	Live bool `json:"live,omitempty"`
}

// Proxy is an HTTP handler that serves HLS streams through the backend.
// On first request for a segment it fetches from the CDN and writes to
// the blob store in parallel (write-through cache); subsequent requests redirect
// straight to the blob store. The player always talks to local infrastructure.
//
// Routes (all under a common prefix, e.g. /api/stream/):
//
//	GET /api/stream/manifest?s=<session>&u=<encodedURL>
//	GET /api/stream/seg?s=<session>&h=<hash>&u=<encodedURL>
type Proxy struct {
	cache   *CacheService
	kv      *cache.Client
	baseURL string // public backend URL, e.g. "http://localhost:8080"
}

func NewProxy(cs *CacheService, kv *cache.Client, baseURL string) *Proxy {
	return &Proxy{cache: cs, kv: kv, baseURL: baseURL}
}

// ManifestURL returns a local proxy URL for the given resolved stream.
// For HLS: stores headers in a kv session and returns the proxy manifest URL.
// For MP4: returns the blob store URL if cached, otherwise returns a proxy URL that
// forwards Range requests to the CDN while background-caching the file.
func (p *Proxy) ManifestURL(ctx context.Context, result *plugins.ResolveResult) (string, error) {
	// A cache hit is already a local blob. Proxying it would fetch it from
	// ourselves and cache a second copy.
	if strings.HasPrefix(result.URL, p.cache.storage.DirectURL(blobPrefix)) {
		return result.URL, nil
	}
	if !isHLS(result) {
		id, err := p.newSession(ctx, result.Headers, false)
		if err != nil {
			return result.URL, nil
		}
		p.cache.StartCaching(result, result.Headers)
		return p.baseURL + "/api/stream/mp4?s=" + id + "&u=" + encURL(result.URL), nil
	}

	id, err := p.newSession(ctx, result.Headers, false)
	if err != nil {
		return result.URL, nil // fallback: return remote URL on session failure
	}
	return p.baseURL + "/api/stream/manifest?s=" + id + "&u=" + encURL(result.URL), nil
}

// LiveManifestURL returns a local proxy URL for a live HLS broadcast (a cam
// room). The session is marked live: playlists change constantly and segments
// are ephemeral, so nothing is cached — no blob, no kv entry, no redirect to
// the blob store.
func (p *Proxy) LiveManifestURL(ctx context.Context, result *plugins.ResolveResult) (string, error) {
	id, err := p.newSession(ctx, result.Headers, true)
	if err != nil {
		return "", err
	}
	return p.baseURL + "/api/stream/manifest?s=" + id + "&u=" + encURL(result.URL), nil
}

// ServeHTTP handles /api/stream/manifest, /api/stream/seg, and /api/stream/mp4.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Any playback request marks the stream as active, so the download queue can
	// hold off while the user is watching (when that setting is enabled).
	if p.cache != nil {
		p.cache.MarkStreamActivity(r.Context())
	}
	switch {
	case strings.HasSuffix(r.URL.Path, "/manifest"):
		p.serveManifest(w, r)
	case strings.HasSuffix(r.URL.Path, "/seg"):
		p.serveSeg(w, r)
	case strings.HasSuffix(r.URL.Path, "/mp4"):
		p.serveMP4(w, r)
	default:
		http.NotFound(w, r)
	}
}

// serveManifest fetches the upstream m3u8, rewrites all URIs to proxy URLs, and
// serves the rewritten manifest. Always fetches fresh (supports live streams).
func (p *Proxy) serveManifest(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	sessID := q.Get("s")
	manifestURL, err := decURL(q.Get("u"))
	if err != nil {
		http.Error(w, "bad url", http.StatusBadRequest)
		return
	}

	session := p.loadSessionData(r.Context(), sessID)
	headers := session.Headers

	resp, err := p.cache.fetch(r.Context(), manifestURL, headers)
	if err != nil {
		http.Error(w, "upstream: "+err.Error(), http.StatusBadGateway)
		return
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		http.Error(w, "read error", http.StatusBadGateway)
		return
	}

	base, _ := url.Parse(manifestURL)
	manifest := string(body)
	if session.Live {
		manifest = stripLowLatencyTags(manifest)
	}
	rewritten := p.rewriteManifest(manifest, base, sessID)

	w.Header().Set("Content-Type", "application/x-mpegURL")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write([]byte(rewritten))
}

// rewriteManifest replaces all segment and variant URIs in an m3u8 with proxy
// URLs. They are root-relative, so the player resolves them against the host it
// fetched the manifest from.
func (p *Proxy) rewriteManifest(manifest string, base *url.URL, sessID string) string {
	lines := strings.Split(manifest, "\n")
	out := make([]string, 0, len(lines))
	prevExtInf := false
	prevStreamInf := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		isURI := line != "" && !strings.HasPrefix(line, "#")

		if prevExtInf && isURI {
			// Segment line → proxy through /seg
			abs := absURL(base, line)
			h := urlHash(abs)
			out = append(out, "/api/stream/seg?s="+sessID+"&h="+h+"&u="+encURL(abs))
			prevExtInf = false
			continue
		}
		if prevStreamInf && isURI {
			// Variant playlist → proxy through /manifest recursively
			abs := absURL(base, line)
			out = append(out, "/api/stream/manifest?s="+sessID+"&u="+encURL(abs))
			prevStreamInf = false
			continue
		}

		prevExtInf = strings.HasPrefix(line, "#EXTINF")
		prevStreamInf = strings.HasPrefix(line, "#EXT-X-STREAM-INF")
		out = append(out, p.rewriteTagURI(line, base, sessID))
	}
	return strings.Join(out, "\n")
}

// uriAttribute matches the URI="…" attribute of an HLS tag.
var uriAttribute = regexp.MustCompile(`URI="([^"]*)"`)

// rewriteTagURI proxies the URI attribute of tags that reference files:
// fMP4 init segments and keys through /seg, alternate renditions and
// I-frame playlists through /manifest. Other lines pass unchanged.
func (p *Proxy) rewriteTagURI(line string, base *url.URL, sessID string) string {
	var route string
	switch {
	case strings.HasPrefix(line, "#EXT-X-MAP"), strings.HasPrefix(line, "#EXT-X-KEY"):
		route = "seg"
	case strings.HasPrefix(line, "#EXT-X-MEDIA"), strings.HasPrefix(line, "#EXT-X-I-FRAME-STREAM-INF"):
		route = "manifest"
	default:
		return line
	}
	return uriAttribute.ReplaceAllStringFunc(line, func(attribute string) string {
		abs := absURL(base, uriAttribute.FindStringSubmatch(attribute)[1])
		if route == "seg" {
			return `URI="/api/stream/seg?s=` + sessID + "&h=" + urlHash(abs) + "&u=" + encURL(abs) + `"`
		}
		return `URI="/api/stream/manifest?s=` + sessID + "&u=" + encURL(abs) + `"`
	})
}

// serveSeg serves an HLS segment from the blob store (redirect) if cached, otherwise
// fetches from the CDN and tee-writes to the blob store while streaming to the client.
func (p *Proxy) serveSeg(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	sessID := q.Get("s")
	h := q.Get("h")
	segURL, err := decURL(q.Get("u"))
	if err != nil {
		http.Error(w, "bad url", http.StatusBadRequest)
		return
	}

	session := p.loadSessionData(r.Context(), sessID)
	if session.Live {
		p.serveLiveSeg(w, r, segURL, session.Headers)
		return
	}

	key := blobPrefix + "seg/" + h + segExt(segURL)

	// Cache hit: redirect to the blob store — the player fetches directly, saving backend bandwidth.
	// The touch also confirms the blob still exists; an evicted one is fetched again.
	if v, _ := p.cache.kv.Get(r.Context(), kvSegmentPrefix+h); v == "1" && p.cache.storage.Touch(r.Context(), key) == nil {
		http.Redirect(w, r, storage.RelativeURL(key), http.StatusFound)
		return
	}

	// Cache miss: fetch from CDN and write-through to the blob store.
	headers := session.Headers
	resp, err := p.cache.fetch(r.Context(), segURL, headers)
	if err != nil {
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "video/MP2T"
	}
	w.Header().Set("Content-Type", ct)
	if resp.ContentLength > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", resp.ContentLength))
	}

	// Pipe body to: (a) HTTP response, (b) the blob store via a goroutine.
	pr, pw := io.Pipe()
	go func() {
		defer pr.Close()
		if err := p.cache.storage.Put(context.Background(), key, pr, resp.ContentLength, ct); err == nil {
			_ = p.cache.kv.Set(context.Background(), kvSegmentPrefix+h, "1", ttlSegment)
		}
	}()

	if _, err := io.Copy(w, io.TeeReader(resp.Body, pw)); err != nil {
		_ = pw.CloseWithError(err)
		return
	}
	pw.Close()
}

// serveLiveSeg relays a live segment (or init/key file) from the origin without
// touching the stream cache.
func (p *Proxy) serveLiveSeg(w http.ResponseWriter, r *http.Request, segURL string, headers map[string]string) {
	resp, err := p.cache.fetch(r.Context(), segURL, headers)
	if err != nil {
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Cache-Control", "no-store")
	_, _ = io.Copy(w, resp.Body)
}

// lowLatencyTagPrefixes are the LL-HLS tags a live playlist loses: they make a
// player request partial segments with blocking-reload query parameters that
// the proxy cannot forward, so the stream is served as plain HLS instead.
var lowLatencyTagPrefixes = []string{
	"#EXT-X-PART:",
	"#EXT-X-PART-INF",
	"#EXT-X-PRELOAD-HINT",
	"#EXT-X-RENDITION-REPORT",
	"#EXT-X-SERVER-CONTROL",
}

// stripLowLatencyTags removes the low-latency lines from a live playlist.
func stripLowLatencyTags(manifest string) string {
	lines := strings.Split(manifest, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if hasAnyPrefix(strings.TrimSpace(line), lowLatencyTagPrefixes) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// hasAnyPrefix reports whether text starts with one of the prefixes.
func hasAnyPrefix(text string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}

// serveMP4 proxies an MP4: redirects to the blob store if cached, otherwise forwards the
// request (including any Range header) to the CDN. Background caching is started
// by ManifestURL so the file lands in the blob store without blocking playback.
func (p *Proxy) serveMP4(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	sessID := q.Get("s")
	mediaURL, err := decURL(q.Get("u"))
	if err != nil {
		http.Error(w, "bad url", http.StatusBadRequest)
		return
	}

	// Cache hit: redirect directly to the blob store.
	if cached, ok := p.cache.Lookup(r.Context(), mediaURL); ok {
		http.Redirect(w, r, cached.URL, http.StatusFound)
		return
	}

	headers, _ := p.loadSession(r.Context(), sessID)

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, mediaURL, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if rng := r.Header.Get("Range"); rng != "" {
		req.Header.Set("Range", rng)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, "upstream error", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Forward relevant response headers.
	for _, h := range []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges"} {
		if v := resp.Header.Get(h); v != "" {
			w.Header().Set(h, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

// -- session --

func (p *Proxy) newSession(ctx context.Context, headers map[string]string, live bool) (string, error) {
	id := strings.ReplaceAll(uuid.New().String(), "-", "")[:12]
	data, _ := json.Marshal(sessionData{Headers: headers, Live: live})
	if err := p.kv.Set(ctx, sessPrefix+id, string(data), sessTTL); err != nil {
		return "", err
	}
	return id, nil
}

func (p *Proxy) loadSession(ctx context.Context, id string) (map[string]string, error) {
	session, err := p.readSession(ctx, id)
	return session.Headers, err
}

// loadSessionData returns the session, or an empty one when it expired.
func (p *Proxy) loadSessionData(ctx context.Context, id string) sessionData {
	session, _ := p.readSession(ctx, id)
	return session
}

// readSession loads a stream session from the kv store.
func (p *Proxy) readSession(ctx context.Context, id string) (sessionData, error) {
	val, err := p.kv.Get(ctx, sessPrefix+id)
	if err != nil {
		return sessionData{}, err
	}
	var session sessionData
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return sessionData{}, err
	}
	return session, nil
}

// -- URL encoding --

func encURL(raw string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decURL(encoded string) (string, error) {
	b, err := base64.RawURLEncoding.DecodeString(encoded)
	return string(b), err
}
