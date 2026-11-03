package pipeline

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

const (
	// identifyInterval is how long a site that didn't know a performer isn't
	// asked again, unless the user refreshes.
	identifyInterval = 7 * 24 * time.Hour
	// identifyTimeout bounds one site's answer, so a slow site doesn't hold
	// up the performer's page.
	identifyTimeout = 30 * time.Second
)

// IdentifyPerformer asks every plugin that can identify performers to find
// this one on its site, by what the library knows about them (name, aliases,
// birthdate, known profiles). Sites the performer already has a profile on
// are skipped, and so are sites asked within identifyInterval unless again is
// set; when only names plugins, the others are too. Each profile a site is
// sure of becomes a source of the performer and fills their details. Sites
// are asked at once; failures are logged.
func (p *Pipeline) IdentifyPerformer(ctx context.Context, id string, again bool, only []string) error {
	performer, err := p.library.Performer(ctx, id)
	if err != nil || performer == nil {
		return err
	}
	sources, err := p.library.Sources(ctx, schema.RecordKindPerformer, id)
	if err != nil {
		return err
	}
	args := identifyArgs(performer, sources)
	known := sourcePlugins(sources)
	var group sync.WaitGroup
	for _, plugin := range p.registry.WithCapability(plugins.CapabilityPerformerIdentify) {
		if known[plugin.Meta.Name] || (!again && p.askedRecently(plugin.Meta.Name, id)) {
			continue
		}
		if !pluginWanted(only, plugin.Meta.Name) {
			continue
		}
		group.Go(func() { p.identifyOn(ctx, plugin, id, args) })
	}
	group.Wait()
	return nil
}

// identifyOn asks one plugin about the performer and stores what it's sure of.
func (p *Pipeline) identifyOn(ctx context.Context, plugin *plugins.Plugin, id string, args plugins.PerformerIdentifyArgs) {
	ctx, cancel := context.WithTimeout(ctx, identifyTimeout)
	defer cancel()
	profile, err := p.runner.IdentifyPerformer(ctx, plugin, args)
	if err != nil {
		log.Printf("identify %s on %s: %v", args.Name, plugin.Meta.Name, err)
		return
	}
	p.identified.Store(identifyKey(plugin.Meta.Name, id), time.Now())
	if profile == nil {
		return
	}
	attached, err := p.library.AttachPerformerProfile(ctx, id, plugin.Meta.Name, profile)
	if err != nil {
		log.Printf("identify %s on %s: store: %v", args.Name, plugin.Meta.Name, err)
		return
	}
	if !attached {
		log.Printf("identify %s on %s: %s is another performer's profile", args.Name, plugin.Meta.Name, profile.SourceURL)
		return
	}
	if err := p.QueuePerformerVideos(ctx, id, plugin.Meta.Name); err != nil {
		log.Printf("identify %s on %s: queue videos: %v", args.Name, plugin.Meta.Name, err)
	}
}

// askedRecently reports whether the plugin answered about the performer
// within identifyInterval. Kept in memory: after a restart each site is asked
// once more.
func (p *Pipeline) askedRecently(pluginName, id string) bool {
	asked, found := p.identified.Load(identifyKey(pluginName, id))
	if !found {
		return false
	}
	return time.Since(asked.(time.Time)) < identifyInterval
}

// identifyKey names one plugin's answer about one performer.
func identifyKey(pluginName, id string) string {
	return pluginName + "|" + id
}

// identifyArgs is what the library knows about a performer, for the sites to
// recognise them by.
func identifyArgs(performer *schema.Performer, sources []schema.Source) plugins.PerformerIdentifyArgs {
	args := plugins.PerformerIdentifyArgs{Name: performer.Name, Aliases: performer.Aliases, URLs: performer.Links}
	for _, source := range sources {
		args.URLs = append(args.URLs, source.Url)
	}
	if performer.Image != nil {
		args.ImageURL = performer.Image.Uri
	}
	setOptional(&args.Birthdate, performer.Attributes.Birthdate)
	setOptional(&args.Country, performer.Attributes.Country)
	setOptional(&args.Gender, performer.Attributes.Gender)
	return args
}

// setOptional copies an optional value when it's set.
func setOptional(target *string, value *string) {
	if value != nil {
		*target = *value
	}
}

// sourcePlugins is the set of plugins a record already has a source on.
func sourcePlugins(sources []schema.Source) map[string]bool {
	known := map[string]bool{}
	for _, source := range sources {
		known[source.Plugin] = true
	}
	return known
}
