package pipeline

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
)

// EnqueueScrape schedules a scrape job for a specific plugin and URL.
func (o *Orchestrator) EnqueueScrape(ctx context.Context, pluginName, url string) (string, error) {
	return o.queue.Enqueue(ctx, jobs.Job{
		Kind:       jobs.KindScrape,
		PluginName: pluginName,
		Payload:    map[string]any{"url": url},
		DedupeKey:  fmt.Sprintf("scrape:%s:%s", pluginName, url),
	})
}

// EnqueueEnrich schedules an enrich job for a media item (e.g. a performer),
// letting enrich-capable plugins fill in missing details.
func (o *Orchestrator) EnqueueEnrich(ctx context.Context, mediaID string) error {
	m, err := o.ingestSvc.GetByID(ctx, mediaID)
	if err != nil || m == nil {
		return fmt.Errorf("media %s not found", mediaID)
	}
	pluginName, _ := o.ingestSvc.PrimarySource(ctx, m)
	_, err = o.queue.Enqueue(ctx, jobs.Job{
		Kind:       jobs.KindEnrich,
		PluginName: pluginName,
		Payload:    map[string]any{"media_id": mediaID, "missing": []string{}},
		DedupeKey:  fmt.Sprintf("enrich:%s", mediaID),
	})
	return err
}

// EnqueueSpeedCheck schedules a low-priority speed-check job for a media item.
func (o *Orchestrator) EnqueueSpeedCheck(ctx context.Context, mediaID string) error {
	_, err := o.queue.Enqueue(ctx, jobs.Job{
		Kind:      jobs.KindSpeedCheck,
		Payload:   map[string]any{"media_id": mediaID},
		DedupeKey: fmt.Sprintf("speed-check:%s", mediaID),
	})
	return err
}

// EnqueueDownload schedules a download job for the given provider URL.
func (o *Orchestrator) EnqueueDownload(ctx context.Context, providerURL, title, pluginName string) (string, error) {
	payload := map[string]any{
		"url":   providerURL,
		"title": title,
	}
	// Link the job back to its media (via the stream URL) so the UI can show
	// download progress on the corresponding media card.
	if mediaID := o.mediaForStreamURL(ctx, providerURL); mediaID != "" {
		payload["media"] = mediaID
	}
	return o.queue.Enqueue(ctx, jobs.Job{
		Kind:       jobs.KindDownload,
		PluginName: pluginName,
		Payload:    payload,
	})
}

// mediaForStreamURL resolves the media record that owns a stream URL.
func (o *Orchestrator) mediaForStreamURL(ctx context.Context, url string) string {
	mediaIDs, err := o.queue.DB().Strings(ctx,
		`SELECT media FROM stream WHERE url = $url LIMIT 1`,
		db.Vars{"url": url})
	if err != nil || len(mediaIDs) == 0 {
		return ""
	}
	return mediaIDs[0]
}
