package library

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// AttachPerformerProfile stores a site's profile of the performer with id,
// which the plugin identified as them (performer:identify): the profile
// becomes one of the performer's sources, read now, and fills the fields
// nobody else owns. A profile already stored as another performer's source is
// left there and reports false: that's two records of one person, for the user
// to merge.
func (l *Library) AttachPerformerProfile(ctx context.Context, id, plugin string, profile *plugins.Performer) (bool, error) {
	site, ok := identity(plugin, profile.ExternalID, profile.SourceURL)
	if !ok {
		return false, fmt.Errorf("%s identified a performer without id or URL", plugin)
	}
	source, err := l.findSource(ctx, site)
	if err != nil {
		return false, err
	}
	if source != nil && source.SubjectId != id {
		return false, nil
	}
	sourceID, err := l.profileSourceID(ctx, id, site, source, profile)
	if err != nil {
		return false, err
	}
	if _, err := l.recorder.Update(ctx, schema.RecordKindPerformer, id, performerFields(profile), provenance.Plugin(plugin, sourceID)); err != nil {
		return false, err
	}
	return true, l.markFetched(ctx, sourceID)
}

// profileSourceID returns the existing source's id, or adds the profile as a
// new source of the performer.
func (l *Library) profileSourceID(ctx context.Context, id string, site siteIdentity, source *schema.Source, profile *plugins.Performer) (string, error) {
	if source != nil {
		return source.Id, nil
	}
	return l.addSource(ctx, schema.RecordKindPerformer, id, site, profile.Name, assetOf(profile.ImagePath))
}
