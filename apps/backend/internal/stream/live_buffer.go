package stream

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/playingwithclouds/veil/internal/livebuffer"
)

// serveBufferedPlaylist answers a live room's media playlist with the buffered
// window instead of the origin's short one, so the player can seek back. It
// reports false when the playlist is not one to buffer (a master playlist, a
// session without a room, an encrypted stream); the caller serves it as is.
func (p *Proxy) serveBufferedPlaylist(w http.ResponseWriter, session sessionData, sessionID, playlistURL, text string, isAudio bool) bool {
	if p.liveBuffer == nil || session.Room == "" || !strings.Contains(text, "#EXTINF") {
		return false
	}
	base, err := url.Parse(playlistURL)
	if err != nil {
		return false
	}
	playlist, err := livebuffer.ParsePlaylist(text, base)
	if err != nil {
		return false
	}
	kind := livebuffer.KindVideo
	if isAudio {
		kind = livebuffer.KindAudio
	}
	source := livebuffer.Source{PlaylistURL: playlistURL, Kind: kind, Fetch: p.cache.fetch, Headers: session.Headers}
	window := p.liveBuffer.Observe(session.Room, source, playlist, true)
	rendered := window.Render(func(name, sourceURL string) string {
		return bufferedSegmentURI(sessionID, name, sourceURL)
	})
	w.Header().Set("Content-Type", "application/x-mpegURL")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write([]byte(rendered))
	return true
}

// bufferedSegmentURI is the proxy URI of a buffered file. sourceURL, when
// known, lets the proxy fall back to the origin if the blob is gone.
func bufferedSegmentURI(sessionID, name, sourceURL string) string {
	uri := "/api/stream/seg?s=" + sessionID + "&k=" + url.QueryEscape(name)
	if sourceURL != "" {
		uri += "&u=" + encURL(sourceURL)
	}
	return uri
}

// serveBufferedSeg serves a segment named by the k parameter from the live
// buffer. It reports false when the request is not for a buffered file or the
// buffer cannot provide it, so the caller relays from the origin.
func (p *Proxy) serveBufferedSeg(w http.ResponseWriter, r *http.Request, session sessionData) bool {
	name := r.URL.Query().Get("k")
	if p.liveBuffer == nil || session.Room == "" || name == "" {
		return false
	}
	reader, contentType, err := p.liveBuffer.Open(r.Context(), session.Room, name)
	if err != nil {
		return false
	}
	defer reader.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-store")
	_, _ = io.Copy(w, reader)
	return true
}
