package pipeline

import (
	"context"
	"fmt"
	"log"

	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

const (
	// performerVideosPage is how many videos one page of a performer's
	// profile asks for.
	performerVideosPage = 40
	// performerVideosMaxPages bounds a favourite's crawl, for a site that
	// never runs out.
	performerVideosMaxPages = 100
	// payloadListURL and payloadAllPages are a performer-videos job's payload.
	payloadListURL  = "listUrl"
	payloadAllPages = "allPages"
)

// QueuePerformerVideos queues listing the videos on each of a performer's
// profiles on sites that can list a page (onlyPlugin's, when set): the first
// page, every page for a favourite. What's listed is stored and credited to
// the performer.
func (p *Pipeline) QueuePerformerVideos(ctx context.Context, performerID, onlyPlugin string) error {
	if p.queue == nil {
		return nil
	}
	sources, err := p.library.Sources(ctx, schema.RecordKindPerformer, performerID)
	if err != nil {
		return err
	}
	allPages, err := p.library.IsFavorite(ctx, schema.RecordKindPerformer, performerID)
	if err != nil {
		return err
	}
	for _, source := range sources {
		if onlyPlugin != "" && source.Plugin != onlyPlugin {
			continue
		}
		if err := p.queuePerformerSource(ctx, performerID, source, allPages); err != nil {
			return err
		}
	}
	return nil
}

// queuePerformerSource queues listing one profile's videos, when its plugin can list pages.
func (p *Pipeline) queuePerformerSource(ctx context.Context, performerID string, source schema.Source, allPages bool) error {
	if _, ok := p.pageLister(source.Plugin); !ok {
		return nil
	}
	dedupeKey := "performer-videos:" + source.Id
	if allPages {
		dedupeKey += ":all"
	}
	_, err := p.queue.Enqueue(ctx, jobs.Job{
		Kind:       jobs.KindScrape,
		Subject:    &schema.RecordRef{Kind: schema.RecordKindPerformer, Id: performerID},
		PluginName: source.Plugin,
		Payload:    map[string]any{payloadListURL: source.Url, payloadAllPages: allPages, "sourceId": source.Id},
		DedupeKey:  dedupeKey,
	})
	return err
}

// IsPerformerVideosJob reports whether a scrape job lists a performer's videos.
func IsPerformerVideosJob(job *jobs.Job) bool {
	_, ok := job.Payload[payloadListURL].(string)
	return ok
}

// CollectPerformerVideos runs a job QueuePerformerVideos queued: it lists
// the profile's videos page by page (only the first unless the job asks for
// all), stores them and credits the performer, until a page brings nothing new.
func (p *Pipeline) CollectPerformerVideos(ctx context.Context, job *jobs.Job, progress func(float64)) error {
	if job.Subject == nil {
		return fmt.Errorf("a performer-videos job needs a performer")
	}
	plugin, ok := p.pageLister(job.PluginName)
	if !ok {
		return fmt.Errorf("%s can't list a performer's videos", job.PluginName)
	}
	listURL, _ := job.Payload[payloadListURL].(string)
	sourceID, _ := job.Payload["sourceId"].(string)
	allPages, _ := job.Payload[payloadAllPages].(bool)
	pages := 1
	if allPages {
		pages = performerVideosMaxPages
	}
	seen := map[string]bool{}
	for page := range pages {
		added, err := p.collectPage(ctx, plugin, job.Subject.Id, sourceID, listURL, page, seen)
		if err != nil {
			return err
		}
		progress(float64(page+1) / float64(pages))
		if added == 0 {
			break
		}
	}
	return nil
}

// collectPage stores one page of a performer's videos, credited to them,
// and returns how many weren't on an earlier page.
func (p *Pipeline) collectPage(ctx context.Context, plugin *plugins.Plugin, performerID, sourceID, listURL string, page int, seen map[string]bool) (int, error) {
	args := plugins.ListArgs{URL: listURL, Limit: performerVideosPage, Offset: page * performerVideosPage}
	result, err := p.runner.List(ctx, plugin, plugins.MediaTypeScene, args)
	if err != nil {
		return 0, err
	}
	added := 0
	for _, ref := range p.storeListed(ctx, plugin.Meta.Name, result.Items) {
		if ref.Kind != schema.RecordKindVideo || seen[ref.Id] {
			continue
		}
		seen[ref.Id] = true
		added++
		if err := p.library.CreditPerformer(ctx, ref.Id, performerID, plugin.Meta.Name, sourceID); err != nil {
			log.Printf("credit %s on %s: %v", performerID, ref.Id, err)
		}
	}
	return added, nil
}

// pageLister returns the active plugin when it can list the videos on a page.
func (p *Pipeline) pageLister(pluginName string) (*plugins.Plugin, bool) {
	if !p.registry.IsActive(pluginName) {
		return nil, false
	}
	plugin, found := p.registry.Get(pluginName)
	if !found || !plugin.Serves(plugins.CapabilitySceneListPage) {
		return nil, false
	}
	return plugin, true
}
