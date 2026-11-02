package plugins

import "slices"

// Contribution kinds.
const (
	ContributionSource = "source"
	ContributionLookup = "lookup"
	ContributionStream = "stream"
	ContributionCams   = "cams"
	ContributionLogin  = "login"
)

// sourceMediaKinds maps the entities a plugin can find or list to the media
// kinds of the schema.
var sourceMediaKinds = []struct {
	entity    MediaType
	mediaKind string
}{
	{MediaTypeScene, "video"},
	{MediaTypeGallery, "gallery"},
	{MediaTypeAudio, "audio"},
	{MediaTypeStory, "story"},
	{MediaTypePost, "post"},
}

// Contributions lists what the plugin adds to Veil: what its capabilities
// imply (a source of the media it finds or lists, a lookup, stream
// resolving for its domains, live cams, a login) and what it declares in
// meta.contributions (views, feeds, triggers, actions, fields).
func (p *Plugin) Contributions() []Contribution {
	contributions := []Contribution{}
	if source, ok := p.sourceContribution(); ok {
		contributions = append(contributions, source)
	}
	if p.Serves(CapabilityLookup) || p.Serves(CapabilityFaceIdentify) || p.ServesForum() {
		contributions = append(contributions, Contribution{Kind: ContributionLookup, Name: p.Meta.Name, Title: p.Meta.DisplayName})
	}
	if p.Serves(CapabilityStreamResolve) {
		contributions = append(contributions, Contribution{Kind: ContributionStream, Name: p.Meta.Name, Title: p.Meta.DisplayName, Domains: p.Meta.Domains})
	}
	if p.Serves(CapabilityCamList) {
		contributions = append(contributions, Contribution{Kind: ContributionCams, Name: p.Meta.Name, Title: p.Meta.DisplayName})
	}
	if p.Serves(CapabilityAccountLogin) {
		contributions = append(contributions, Contribution{Kind: ContributionLogin, Name: p.Meta.Name, Title: p.Meta.DisplayName})
	}
	return append(contributions, p.Meta.Contributions...)
}

// sourceContribution is the plugin's source contribution, when it finds or
// lists any media.
func (p *Plugin) sourceContribution() (Contribution, bool) {
	kinds := []string{}
	for _, entry := range sourceMediaKinds {
		if p.Serves(FindCapability(entry.entity)) || p.Serves(ListCapability(entry.entity)) {
			kinds = append(kinds, entry.mediaKind)
		}
	}
	if len(kinds) == 0 {
		return Contribution{}, false
	}
	return Contribution{Kind: ContributionSource, Name: p.Meta.Name, Title: p.Meta.DisplayName, MediaKinds: kinds, Domains: p.Meta.Domains}, true
}

// HasContribution reports whether the plugin contributes kind under name
// (any name when name is empty).
func (p *Plugin) HasContribution(kind, name string) bool {
	return slices.ContainsFunc(p.Contributions(), func(contribution Contribution) bool {
		return contribution.Kind == kind && (name == "" || contribution.Name == name)
	})
}
