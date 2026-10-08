// Package outbound holds the TLS and transport settings every HTTP request the
// backend makes to other sites shares (plugin fetches, streams, images,
// posters, icons).
package outbound

import (
	"crypto/tls"
	"net/http"
)

// TLSConfig returns the client TLS settings for outbound requests. Pornhub
// answers 403 to handshakes offering the post-quantum X25519MLKEM768 key
// share Go sends by default, so only classic curves are offered.
func TLSConfig() *tls.Config {
	return &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256, tls.CurveP384},
	}
}

// NewTransport returns a copy of http.DefaultTransport using TLSConfig.
func NewTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = TLSConfig()
	return transport
}

// Client is the shared client for one-shot requests (images, posters, icons).
// Callers bound each request with a context.
var Client = &http.Client{Transport: NewTransport()}
