package pipeline

import (
	"context"
	"log"
	"time"

	"github.com/playingwithclouds/veil/internal/cache"
	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/plugins"
)

const (
	// profilePluginName is the plugin whose performer pages describe the people
	// in the library: their stage names let records of one person be merged.
	profilePluginName = "freeones"

	// profileCandidates is how many search hits are opened to find the performer.
	profileCandidates = 3

	// profilePause spaces the requests of a profiling run out.
	profilePause = 1500 * time.Millisecond

	// profileMissTTL is how long a performer the site could not place is left alone.
	profileMissTTL = 30 * 24 * time.Hour
)

// SetProfileCache wires in the store that remembers which performers the profile
// site could not place.
func (o *Orchestrator) SetProfileCache(client *cache.Client) {
	o.profileCache = client
}

// ProfilePerformers looks up to limit performers without a profile on the
// profile site and ingests the profile of the one that is the performer, which
// fills their bio and stage names and merges records of the same person. Returns
// how many were profiled. A site that is not installed is not an error.
func (o *Orchestrator) ProfilePerformers(ctx context.Context, limit int) (int, error) {
	plugin, found := o.registry.Get(profilePluginName)
	if !found || !o.registry.IsActive(profilePluginName) {
		return 0, nil
	}
	candidates, err := o.ingestSvc.PerformersWithoutProfile(ctx, profilePluginName)
	if err != nil {
		return 0, err
	}
	profiled, attempted := 0, 0
	for _, candidate := range candidates {
		if attempted >= limit || ctx.Err() != nil {
			break
		}
		if o.recentlyUnplaced(ctx, candidate.ID) {
			continue
		}
		attempted++
		placed, err := o.profilePerformer(ctx, plugin, candidate)
		if err != nil {
			log.Printf("orchestrator: profile %q: %v", candidate.Name, err)
		} else if placed {
			profiled++
		} else {
			o.rememberUnplaced(ctx, candidate.ID)
		}
		o.pauseBetweenProfiles(ctx)
	}
	return profiled, nil
}

// profilePerformer searches the profile site for the performer and ingests the
// first hit whose name or stage names include the performer's name. False when
// none does.
func (o *Orchestrator) profilePerformer(ctx context.Context, plugin *plugins.Plugin, performer ingest.PerformerName) (bool, error) {
	hits, err := o.runner.List(ctx, plugin, plugins.MediaTypePerformer, plugins.ListArgs{Query: performer.Name, Limit: profileCandidates})
	if err != nil {
		return false, err
	}
	for index, hit := range hits.Items {
		if index >= profileCandidates {
			break
		}
		results, err := o.runner.Find(ctx, plugin, plugins.MediaTypePerformer, hit.SourceURL)
		if err != nil || len(results) == 0 || results[0].Performer == nil {
			continue
		}
		if !knownAs(results[0].Performer, performer.Name) {
			continue
		}
		return true, o.ingestScrapeResult(ctx, plugin.Meta.Name, hit.SourceURL, results)
	}
	return false, nil
}

// knownAs reports whether the profile's name or one of its stage names is name.
func knownAs(profile *plugins.Performer, name string) bool {
	key := names.Key(name)
	if names.Key(profile.Name) == key {
		return true
	}
	for _, alias := range profile.Aliases {
		if names.Key(alias) == key {
			return true
		}
	}
	return false
}

// unplacedKey is the cache key under which a performer the site could not
// place is remembered.
func unplacedKey(performerID string) string {
	return "performer-profile-miss:" + performerID
}

// recentlyUnplaced reports whether the profile site was already asked about the
// performer and had no page for them.
func (o *Orchestrator) recentlyUnplaced(ctx context.Context, performerID string) bool {
	if o.profileCache == nil {
		return false
	}
	value, err := o.profileCache.Get(ctx, unplacedKey(performerID))
	return err == nil && value != ""
}

// rememberUnplaced stops the performer from being looked up again for a while.
func (o *Orchestrator) rememberUnplaced(ctx context.Context, performerID string) {
	if o.profileCache == nil {
		return
	}
	if err := o.profileCache.Set(ctx, unplacedKey(performerID), "1", profileMissTTL); err != nil {
		log.Printf("orchestrator: remember unplaced performer: %v", err)
	}
}

// pauseBetweenProfiles waits profilePause or until ctx ends.
func (o *Orchestrator) pauseBetweenProfiles(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-time.After(profilePause):
	}
}
