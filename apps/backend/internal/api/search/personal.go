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

// storedRecord loads the canonical record a hit was stored as, nil for a hit
// that was not stored, for other media types than scenes and galleries, and on
// failure.
func (h *Handler) storedRecord(ctx context.Context, hit searchrank.Hit) map[string]any {
	if hit.ID == "" || h.ingestSvc == nil {
		return nil
	}
	if hit.Item.MediaType != plugins.MediaTypeScene && hit.Item.MediaType != plugins.MediaTypeGallery {
		return nil
	}
	record, err := h.ingestSvc.GetByID(ctx, hit.ID)
	if err != nil {
		return nil
	}
	return record
}

// personalAffinity is how much the taste profile likes a scene hit, in
// [-1, 1]: its stored performers, tags and studio (a stub without them counts
// for nothing, record is nil for one that was not stored) and the site that
// listed it. Zero for other media types and without a profile.
func (h *Handler) personalAffinity(profile *recommend.Profile, hit searchrank.Hit, record map[string]any) float64 {
	if profile == nil || hit.Item.MediaType != plugins.MediaTypeScene {
		return 0
	}
	affinities := searchrank.Affinities{}
	if hit.Source != searchrank.LibrarySource {
		affinities.Site = profile.Sites[hit.Source]
	}
	addStoredAffinities(profile, record, &affinities)
	return searchrank.Personal(affinities)
}

// addStoredAffinities fills the performer, tag and studio affinities of a
// stored scene.
func addStoredAffinities(profile *recommend.Profile, scene map[string]any, affinities *searchrank.Affinities) {
	if scene == nil {
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

// creditedIDs is the set of performer, studio and tag ids a stored record is
// credited to.
func creditedIDs(record map[string]any) map[string]bool {
	credited := map[string]bool{}
	if record == nil {
		return credited
	}
	for _, field := range []string{"performers", "tags"} {
		for _, id := range idList(record[field]) {
			credited[id] = true
		}
	}
	if studio, ok := record["studio"].(string); ok && studio != "" {
		credited[studio] = true
	}
	return credited
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
