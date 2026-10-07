package resolve

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/stream"
)

// Service resolves a provider URL (e.g. https://voe.sx/abc) to a direct stream URL.
// Resolution order:
//  1. MinIO cache (via CacheService) — fully local, NAS-speed
//  2. Domain-specific resolver plugin (resolve capability, exact domain match)
//  3. Wildcard resolver plugin (resolve capability, domain = "*") — generic fallback
type Service struct {
	registry *plugins.Registry
	runner   *plugins.Runner
	cache    *stream.CacheService // nil if not configured
}

func New(registry *plugins.Registry, runner *plugins.Runner) *Service {
	return &Service{registry: registry, runner: runner}
}

// SetCache wires in the stream cache service. Call after New() to enable caching.
func (s *Service) SetCache(c *stream.CacheService) {
	s.cache = c
}

// Resolve returns a direct playable URL for the given provider URL.
func (s *Service) Resolve(ctx context.Context, providerURL string) (*plugins.ResolveResult, error) {
	if s.cache != nil {
		if cached, ok := s.cache.Lookup(ctx, providerURL); ok {
			log.Printf("resolve: cache hit %s → %s", providerURL, cached.URL)
			return cached, nil
		}
	}

	result, err := s.doResolve(ctx, providerURL)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		s.cache.StartCaching(result, result.Headers)
	}
	return result, nil
}

func (s *Service) doResolve(ctx context.Context, providerURL string) (*plugins.ResolveResult, error) {
	// Direct media URLs need no plugin — return them as-is.
	if isDirectMedia(providerURL) {
		mimeType := "video/mp4"
		if strings.Contains(providerURL, ".m3u8") {
			mimeType = "application/x-mpegURL"
		}
		return &plugins.ResolveResult{URL: providerURL, MimeType: mimeType}, nil
	}

	result, err := s.resolvePlugin(ctx, providerURL)
	if err == nil {
		return result, nil
	}
	if err != errNoPlugin {
		return nil, err
	}
	return nil, fmt.Errorf("no resolver could handle %s", providerURL)
}

func isDirectMedia(rawURL string) bool {
	p := strings.ToLower(strings.Split(rawURL, "?")[0])
	return strings.HasSuffix(p, ".mp4") ||
		strings.HasSuffix(p, ".mkv") ||
		strings.HasSuffix(p, ".webm") ||
		strings.HasSuffix(p, ".m3u8") ||
		strings.HasSuffix(p, ".ts")
}

var errNoPlugin = fmt.Errorf("no resolver plugin for this domain")

// resolvePlugin tries domain-specific plugins first, then wildcard ("*") plugins as fallback.
func (s *Service) resolvePlugin(ctx context.Context, providerURL string) (*plugins.ResolveResult, error) {
	domain := hostOf(providerURL)
	var wildcards []*plugins.Plugin

	for _, p := range s.registry.WithCapability(plugins.CapabilityStreamResolve) {
		for _, d := range p.Meta.Domains {
			if d == "*" {
				wildcards = append(wildcards, p)
			} else if d == domain || strings.HasSuffix(domain, "."+d) {
				result, err := s.runner.Resolve(ctx, p, providerURL)
				if err == nil {
					return result, nil
				}
				log.Printf("resolve: domain plugin %s failed for %s: %v", p.Meta.Name, providerURL, err)
				// Fall through to wildcard plugins rather than hard-failing.
			}
		}
	}

	for _, p := range wildcards {
		result, err := s.runner.Resolve(ctx, p, providerURL)
		if err == nil {
			return result, nil
		}
		log.Printf("resolve: fallback plugin %s failed for %s: %v", p.Meta.Name, providerURL, err)
	}

	return nil, errNoPlugin
}

// SupportLevel says how well a provider URL can be resolved to a playable
// stream: 2 = direct media or domain-specific resolver plugin, 1 = only a
// wildcard resolver could try it, 0 = no resolver at all. Used to rank
// sources — an unsupported host must never become the primary source.
func (s *Service) SupportLevel(providerURL string) int {
	if isDirectMedia(providerURL) {
		return 2
	}
	domain := hostOf(providerURL)
	hasWildcard := false
	for _, p := range s.registry.WithCapability(plugins.CapabilityStreamResolve) {
		for _, d := range p.Meta.Domains {
			if d == "*" {
				hasWildcard = true
			} else if d == domain || strings.HasSuffix(domain, "."+d) {
				return 2
			}
		}
	}
	if hasWildcard {
		return 1
	}
	return 0
}

func hostOf(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(u.Hostname(), "www.")
}
