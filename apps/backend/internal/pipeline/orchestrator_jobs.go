package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/ingest"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/subscriptions"
)

func (o *Orchestrator) runScrape(ctx context.Context, j *jobs.Job) error {
	url, _ := j.Payload["url"].(string)
	p, ok := o.registry.Get(j.PluginName)
	if !ok {
		return fmt.Errorf("plugin %q not found", j.PluginName)
	}

	entity, ok := p.ContentFindEntity()
	if !ok {
		return fmt.Errorf("plugin %q has no find capability", j.PluginName)
	}
	result, err := o.runner.Find(ctx, p, entity, url)
	if err != nil {
		return err
	}

	// Back-fill poster from the enqueueing listing's payload when the plugin
	// couldn't extract it from the detail page.
	if poster, _ := j.Payload["poster"].(string); poster != "" {
		for _, sr := range result {
			if sr.Scene != nil && sr.Scene.PosterPath == "" {
				sr.Scene.PosterPath = poster
			}
			if sr.Gallery != nil && sr.Gallery.CoverPath == "" {
				sr.Gallery.CoverPath = poster
			}
		}
	}

	return o.ingestScrapeResult(ctx, j.PluginName, url, result)
}

func (o *Orchestrator) runEnrich(ctx context.Context, j *jobs.Job) error {
	mediaID, _ := j.Payload["media_id"].(string)
	missing, _ := j.Payload["missing"].([]any)

	m, err := o.ingestSvc.GetByID(ctx, mediaID)
	if err != nil || m == nil {
		return fmt.Errorf("media %s not found", mediaID)
	}

	allEnrichers := o.registry.WithCapability(plugins.CapabilityEnrich)
	if len(allEnrichers) == 0 {
		return fmt.Errorf("no enrich-capable plugins available")
	}

	dataBytes, _ := json.Marshal(m)
	missingStrs := make([]string, len(missing))
	for i, v := range missing {
		missingStrs[i], _ = v.(string)
	}

	args := plugins.EnrichArgs{
		MediaType: plugins.MediaType(ingest.TableFromID(mediaID)),
		Existing:  json.RawMessage(dataBytes),
		Missing:   missingStrs,
	}

	for _, p := range allEnrichers {
		result, err := o.runner.Enrich(ctx, p, args)
		if err != nil {
			log.Printf("enrich(%s, %s): %v", p.Meta.Name, mediaID, err)
			continue
		}

		// Use the canonical source URL as Referer so CDNs blocking hotlinks allow the fetch.
		_, sourceURL := o.ingestSvc.PrimarySource(ctx, m)
		o.uploadEnrichImages(ctx, mediaID, sourceURL, result)
		// Prefer the the blob store-hosted avatar over the remote (hotlink-protected) one.
		if len(result.Images) > 0 && result.Images[0].FilePath != "" {
			result.ImagePath = &result.Images[0].FilePath
		}
		merged := mergeEnrich(dataBytes, result)
		if err := o.ingestSvc.UpdateRecord(ctx, mediaID, merged); err != nil {
			return fmt.Errorf("save enriched data: %w", err)
		}
		log.Printf("enriched media %s via %s", mediaID, p.Meta.Name)
		break
	}
	return nil
}

func (o *Orchestrator) runSpeedCheck(ctx context.Context, j *jobs.Job) error {
	mediaID, _ := j.Payload["media_id"].(string)
	m, err := o.ingestSvc.GetByID(ctx, mediaID)
	if err != nil || m == nil {
		return fmt.Errorf("media %s not found", mediaID)
	}
	if o.streamResolver == nil {
		return fmt.Errorf("stream resolver not configured")
	}

	downloads, _ := m["downloads"].([]any)
	if len(downloads) == 0 {
		return nil
	}

	type result struct {
		idx      int
		url      string
		speedBps float64
		resolved bool
	}
	ch := make(chan result, len(downloads))
	var wg sync.WaitGroup
	for i, dl := range downloads {
		dm, ok := dl.(map[string]any)
		if !ok {
			continue
		}
		url, _ := dm["url"].(string)
		if url == "" {
			continue
		}
		wg.Add(1)
		go func(i int, url string) {
			defer wg.Done()
			resolved, err := o.streamResolver.Resolve(ctx, url)
			if err != nil {
				log.Printf("speed-check %s: resolve %s: %v", mediaID, url, err)
				ch <- result{i, url, 0, false}
				return
			}
			ch <- result{i, url, resolved.SpeedBps, true}
		}(i, url)
	}
	go func() { wg.Wait(); close(ch) }()

	updated := false
	for r := range ch {
		// Record the outcome on the stream row: a successful resolve marks the
		// source verified (it can actually produce a playable URL); a failed one
		// unmarks it so ranking pushes it down.
		o.updateStreamCheckResult(ctx, mediaID, r.url, r.speedBps, r.resolved)
		if r.speedBps > 0 {
			if dm, ok := downloads[r.idx].(map[string]any); ok {
				dm["speed_bps"] = r.speedBps
				updated = true
			}
		}
	}
	// A speed-check always changes stream verification/speed, so notify any live
	// movie-page subscriber to re-rank its sources.
	o.publishStreamsChanged(mediaID)
	if !updated {
		return nil
	}
	m["downloads"] = downloads
	return o.ingestSvc.UpdateRecord(ctx, mediaID, m)
}

// publishStreamsChanged notifies subscribers that a media item's playback
// sources changed (new streams ingested, or a speed-check re-ranked them).
func (o *Orchestrator) publishStreamsChanged(mediaID string) {
	if o.hub == nil || mediaID == "" {
		return
	}
	o.hub.Publish(subscriptions.StreamsTopic(mediaID), subscriptions.StreamsChangedEvent{MediaID: mediaID})
}

// updateStreamCheckResult writes a speed-check outcome onto the matching
// stream row so source ranking reflects real resolvability.
func (o *Orchestrator) updateStreamCheckResult(ctx context.Context, mediaID, url string, speedBps float64, resolved bool) {
	if _, err := db.ParseRecordID(mediaID); err != nil {
		return
	}
	vars := db.Vars{"media": mediaID, "url": url, "verified": resolved, "speed": nil}
	if speedBps > 0 {
		vars["speed"] = speedBps
	}
	_, err := o.queue.DB().Exec(ctx,
		`UPDATE stream SET verified = $verified, expected_speed_bps = coalesce($speed, expected_speed_bps)
		WHERE media = $media AND url = $url`, vars)
	if err != nil {
		log.Printf("speed-check %s: update stream row: %v", mediaID, err)
	}
}

// errDownloadDeferred signals runDownload wants the job rescheduled (not failed)
// because a stream is playing and downloads-while-streaming is disabled.
var errDownloadDeferred = errors.New("download deferred: stream active")

func (o *Orchestrator) runDownload(ctx context.Context, j *jobs.Job) error {
	providerURL, _ := j.Payload["url"].(string)
	if providerURL == "" {
		return fmt.Errorf("download job missing url")
	}

	// Hold the download while the user is streaming, unless allowed to overlap.
	if o.settings != nil && !o.settings.Cached().AllowDownloadsWhileStreaming {
		if o.streamCache != nil && o.streamCache.StreamActive(ctx) {
			return errDownloadDeferred
		}
	}
	if o.streamResolver == nil {
		return fmt.Errorf("stream resolver not configured")
	}
	if o.streamCache == nil {
		return fmt.Errorf("stream cache not configured")
	}

	result, err := o.streamResolver.Resolve(ctx, providerURL)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}

	blobURL, err := o.streamCache.CacheNow(ctx, result, result.Headers, func(pct float64, rx, total int64) {
		_ = o.updateDownloadProgress(ctx, j, pct, rx, total)
	})
	if err != nil {
		return fmt.Errorf("cache to the blob store: %w", err)
	}

	return o.queue.DB().Merge(ctx, *j.ID, map[string]any{
		"payload": downloadPayload(j, map[string]any{
			"download_url": blobURL,
			"mime_type":    result.MimeType,
			"progress":     100.0,
		}),
	})
}

// downloadPayload returns the job's payload with changes applied. Merge
// replaces the whole payload column, so everything else (the url, title and
// the media link the library needs) has to be carried over.
func downloadPayload(j *jobs.Job, changes map[string]any) map[string]any {
	payload := make(map[string]any, len(j.Payload)+len(changes))
	for key, value := range j.Payload {
		payload[key] = value
	}
	for key, value := range changes {
		payload[key] = value
	}
	return payload
}

func (o *Orchestrator) updateDownloadProgress(ctx context.Context, j *jobs.Job, pct float64, bytesReceived, bytesTotal int64) error {
	payload := downloadPayload(j, map[string]any{"bytesReceived": bytesReceived})
	if pct >= 0 {
		payload["progress"] = pct
	}
	if bytesTotal > 0 {
		payload["bytesTotal"] = bytesTotal
	}
	err := o.queue.DB().Merge(ctx, *j.ID, map[string]any{"payload": payload})

	ev := subscriptions.JobUpdatedEvent{
		ID:        j.StringID(),
		Kind:      string(j.Kind),
		Status:    "running",
		UpdatedAt: subscriptions.NowRFC3339(),
	}
	if title, ok := j.Payload["title"].(string); ok {
		ev.DownloadTitle = title
	}
	if pct >= 0 {
		p := pct
		ev.Progress = &p
	}
	if bytesReceived > 0 {
		f := float64(bytesReceived)
		ev.BytesReceived = &f
	}
	if bytesTotal > 0 {
		f := float64(bytesTotal)
		ev.BytesTotal = &f
	}
	o.hub.Publish(subscriptions.TopicJobUpdated, ev)

	return err
}
