package search

import (
	"context"
	"log"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/recommend"
	"github.com/playingwithclouds/veil/internal/searchrank"
)

// tasteProfile loads the user's taste profile once per search. Nil without a
// source, on failure or while the profile has no signal (cold start).
func (h *Handler) tasteProfile(ctx context.Context) *recommend.Profile {
	if h.taste == nil {
		return nil
	}
	profile, err := h.taste.Profile(ctx)
	if err != nil {
		log.Printf("search: taste profile: %v", err)
		return nil
	}
	if profile.Empty() {
		return nil
	}
	return profile
}

// personalAffinity is how much the taste profile likes a scene hit, in
// [-1, 1]: its stored performers, tags and studio (a stub without them counts
// for nothing) and the site that listed it. Zero for other media types and
// without a profile.
func (h *Handler) personalAffinity(ctx context.Context, profile *recommend.Profile, hit searchrank.Hit) float64 {
	if profile == nil || hit.Item.MediaType != plugins.MediaTypeScene {
		return 0
	}
	affinities := searchrank.Affinities{}
	if hit.Source != searchrank.LibrarySource {
		affinities.Site = profile.Sites[hit.Source]
	}
	if hit.ID != "" {
		h.addStoredAffinities(ctx, profile, hit.ID, &affinities)
	}
	return searchrank.Personal(affinities)
}

// addStoredAffinities fills the performer, tag and studio affinities of a
// stored scene.
func (h *Handler) addStoredAffinities(ctx context.Context, profile *recommend.Profile, sceneID string, affinities *searchrank.Affinities) {
	scene, err := h.ingestSvc.GetByID(ctx, sceneID)
	if err != nil || scene == nil {
		return
	}
	for _, id := range idList(scene["performers"]) {
		affinities.Performers = append(affinities.Performers, profile.Performers[id])
	}
	for _, id := range idList(scene["tags"]) {
		affinities.Tags = append(affinities.Tags, profile.Tags[id])
	}
	if studio, ok := scene["studio"].(string); ok {
		affinities.Studio = profile.Studios[studio]
	}
}

// idList reads a JSON link column ("table:id" strings).
func idList(value any) []string {
	items, _ := value.([]any)
	ids := make([]string, 0, len(items))
	for _, item := range items {
		if id, ok := item.(string); ok {
			ids = append(ids, id)
		}
	}
	return ids
}
