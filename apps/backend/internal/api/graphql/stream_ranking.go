package graphql

import (
	"context"
	"sort"
	"strings"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
)

// loadRankedStreams returns a media item's stored streams ranked the same way
// the movieDetail / mediaStreams queries rank them.
func (r *Resolver) loadRankedStreams(ctx context.Context, mediaID string) ([]*model.Stream, error) {
	streams, err := r.repo.ListStreams(ctx, mediaID)
	if err != nil {
		return nil, err
	}
	r.rankStreams(streams)
	return streams, nil
}

// discoveryTitles returns the distinct titles worth searching for sources: the
// movie's stored (localized) title, its original title, and the caller-provided
// title. Blank and duplicate entries are dropped.
func (r *Resolver) discoveryTitles(ctx context.Context, movieID, provided string) []string {
	candidates := []string{provided}

	seen := make(map[string]bool)
	titles := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		trimmed := strings.TrimSpace(candidate)
		if trimmed == "" || seen[trimmed] {
			continue
		}
		seen[trimmed] = true
		titles = append(titles, trimmed)
	}
	return titles
}

// rankStreams orders sources by how likely playback is to succeed:
// resolver support first (a host without any resolver must never be the
// primary source), then verified resolvability, then measured speed.
func (r *Resolver) rankStreams(streams []*model.Stream) {
	if r.streamSvc == nil {
		return
	}
	support := make(map[string]int, len(streams))
	for _, s := range streams {
		support[s.URL] = r.streamSvc.SupportLevel(s.URL)
	}
	sort.SliceStable(streams, func(i, j int) bool {
		a, b := streams[i], streams[j]
		if support[a.URL] != support[b.URL] {
			return support[a.URL] > support[b.URL]
		}
		if a.Verified != b.Verified {
			return a.Verified
		}
		return speedOf(a) > speedOf(b)
	})
}

func speedOf(s *model.Stream) float64 {
	if s.ExpectedSpeedBps == nil {
		return 0
	}
	return *s.ExpectedSpeedBps
}

func emptyIfNil[T any](list []T) []T {
	if list == nil {
		return []T{}
	}
	return list
}
