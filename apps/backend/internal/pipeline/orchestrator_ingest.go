package pipeline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"

	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/outbound"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/subscriptions"
)

func (o *Orchestrator) ingestScrapeResult(ctx context.Context, pluginName, sourceURL string, results []*plugins.ScrapeResult) error {
	if o.store != nil {
		for _, result := range results {
			o.rewriteScrapeResultPoster(ctx, result)
		}
	}

	ingestResult, err := o.ingestSvc.IngestScrapeResults(ctx, pluginName, sourceURL, results)
	if err != nil {
		return fmt.Errorf("ingest: %w", err)
	}

	// Sources were (re)synced onto the canonical record — notify live pages.
	o.publishStreamsChanged(ingestResult.CanonicalID)

	// Collection members that matched no record yet get scraped so they can
	// slot into the collection on their own ingest.
	for _, memberURL := range ingestResult.UnresolvedMemberURLs {
		if _, err := o.EnqueueScrape(ctx, pluginName, memberURL); err != nil {
			log.Printf("orchestrator: enqueue member scrape for %q: %v", memberURL, err)
		}
	}

	if ingestResult.IsNew {
		// Determine topic and title from the root (first) result.
		var addedTopic, title string
		if len(results) > 0 {
			root := results[0]
			switch root.Type {
			case plugins.MediaTypeScene:
				addedTopic = subscriptions.TopicSceneAdded
				if root.Scene != nil {
					title = root.Scene.Title
				}
			case plugins.MediaTypePerformer:
				addedTopic = subscriptions.TopicPerformerAdded
				if root.Performer != nil {
					title = root.Performer.Name
				}
			case plugins.MediaTypeStudio:
				addedTopic = subscriptions.TopicStudioAdded
				if root.Studio != nil {
					title = root.Studio.Name
				}
			case plugins.MediaTypeGallery:
				addedTopic = subscriptions.TopicGalleryAdded
				if root.Gallery != nil {
					title = root.Gallery.Title
				}
			}
		}
		if addedTopic != "" {
			rootType := ""
			if len(results) > 0 {
				rootType = string(results[0].Type)
			}
			o.hub.Publish(addedTopic, subscriptions.MediaAddedEvent{
				ID:    ingestResult.CanonicalID,
				Type:  rootType,
				Title: title,
			})
		}

		if err := o.EnqueueSpeedCheck(ctx, ingestResult.CanonicalID); err != nil {
			log.Printf("orchestrator: enqueue speed-check for %s: %v", ingestResult.CanonicalID, err)
		}

		autoEnrich := false
		if o.settings != nil {
			autoEnrich = o.settings.Cached().AutoEnrichAfterScrape
		}
		if autoEnrich {
			_, _ = o.queue.Enqueue(ctx, jobs.Job{
				Kind:       jobs.KindEnrich,
				PluginName: pluginName,
				Payload:    map[string]any{"media_id": ingestResult.CanonicalID, "missing": []string{}},
				DedupeKey:  fmt.Sprintf("enrich:%s", ingestResult.CanonicalID),
			})
		}
	}

	return nil
}

// mergeEnrich applies non-nil EnrichResult fields onto raw JSON data. Keys are
// the scene/performer/studio canonical column names.
func mergeEnrich(dataBytes []byte, r *plugins.EnrichResult) map[string]any {
	var m map[string]any
	_ = json.Unmarshal(dataBytes, &m)
	if m == nil {
		m = make(map[string]any)
	}
	setStr := func(key string, v *string) {
		if v != nil {
			m[key] = *v
		}
	}
	setInt := func(key string, v *int) {
		if v != nil {
			m[key] = *v
		}
	}
	setFloat := func(key string, v *float64) {
		if v != nil {
			m[key] = *v
		}
	}
	setStr("title", r.Title)
	setStr("name", r.Name)
	setStr("details", r.Details)
	setStr("date", r.Date)
	setStr("poster_path", r.PosterPath)
	setStr("preview_video", r.PreviewVideo)
	// r.Studio is a StudioRef resolved to a record link by the ingest
	// relation-linking step, never written as a column here.
	setStr("gender", r.Gender)
	setStr("birthdate", r.Birthdate)
	setStr("death_date", r.DeathDate)
	setStr("country", r.Country)
	setStr("ethnicity", r.Ethnicity)
	setStr("eye_color", r.EyeColor)
	setStr("hair_color", r.HairColor)
	setInt("height_cm", r.HeightCm)
	setInt("weight_kg", r.WeightKg)
	setStr("measurements", r.Measurements)
	setStr("fake_tits", r.FakeTits)
	setStr("tattoos", r.Tattoos)
	setStr("piercings", r.Piercings)
	setStr("career_length", r.CareerLength)
	setStr("url", r.URL)
	setStr("twitter", r.Twitter)
	setStr("instagram", r.Instagram)
	setStr("image_path", r.ImagePath)
	setInt("duration_seconds", r.Duration)
	setFloat("rating", r.Rating)
	setSlice := func(key string, ok bool, v any) {
		if ok {
			m[key] = v
		}
	}
	// aliases/tags/performers are name strings; the canonical columns are record
	// links resolved by the ingest relation-linking step, not written here.
	// `images` is intentionally NOT written as a column — no scene/performer table
	// has one; images are materialized separately (uploadEnrichImages / gallery
	// image rows) and surfaced via image_path.
	// Performer aliases are a plain string-array column (unlike scene cast/tags,
	// which are name strings resolved to record links by the ingest step).
	setSlice("aliases", len(r.Aliases) > 0, r.Aliases)
	setSlice("preview_images", len(r.PreviewImages) > 0, r.PreviewImages)
	setSlice("downloads", len(r.Downloads) > 0, r.Downloads)
	return m
}

// rewriteScrapeResultPoster uploads the poster from the scrape result to MinIO
// and replaces the URL in-place so the observation stores a local URL.
func (o *Orchestrator) rewriteScrapeResultPoster(ctx context.Context, result *plugins.ScrapeResult) {
	rewrite := func(externalID string, poster *string, referer string) {
		if *poster == "" || strings.HasPrefix(*poster, "data:") {
			return
		}
		normalized := normalizeURL(*poster)
		if (!strings.HasPrefix(normalized, "http://") && !strings.HasPrefix(normalized, "https://")) ||
			strings.HasPrefix(normalized, o.store.DirectURL("")) {
			return
		}
		url, err := o.uploadPoster(ctx, externalID, normalized, referer)
		if err != nil {
			log.Printf("poster upload %s: %v", externalID, err)
			return
		}
		*poster = url
	}
	switch result.Type {
	case plugins.MediaTypeScene:
		if result.Scene != nil {
			rewrite(result.Scene.ExternalID, &result.Scene.PosterPath, result.Scene.SourceURL)
		}
	case plugins.MediaTypePerformer:
		if result.Performer != nil {
			rewrite(result.Performer.ExternalID, &result.Performer.ImagePath, result.Performer.SourceURL)
		}
	case plugins.MediaTypeStudio:
		if result.Studio != nil {
			rewrite(result.Studio.ExternalID, &result.Studio.ImagePath, result.Studio.SourceURL)
		}
	case plugins.MediaTypeGallery:
		if result.Gallery != nil {
			rewrite(result.Gallery.ExternalID, &result.Gallery.CoverPath, result.Gallery.SourceURL)
		}
	case plugins.MediaTypeCollection:
		if result.Collection != nil {
			rewrite(result.Collection.ExternalID, &result.Collection.CoverPath, result.Collection.SourceURL)
		}
	}
}

// normalizeURL turns protocol-relative URLs (//host/path) into https:// URLs.
func normalizeURL(rawURL string) string {
	if strings.HasPrefix(rawURL, "//") {
		return "https:" + rawURL
	}
	return rawURL
}

// uploadEnrichImages uploads each image URL in the EnrichResult to MinIO and
// rewrites the URL in-place. referer is sent as the HTTP Referer header so CDNs
// that block hotlinking allow the request. No-ops when store is nil.
func (o *Orchestrator) uploadEnrichImages(ctx context.Context, mediaID, referer string, result *plugins.EnrichResult) {
	if o.store == nil || result == nil {
		return
	}
	storeBase := o.store.DirectURL("")
	for i := range result.Images {
		rawURL := result.Images[i].FilePath
		if rawURL == "" {
			continue
		}
		// Skip already-local URLs.
		if strings.HasPrefix(rawURL, storeBase) {
			continue
		}
		normalized := normalizeURL(rawURL)
		if !strings.HasPrefix(normalized, "http://") && !strings.HasPrefix(normalized, "https://") {
			continue
		}
		ext := path.Ext(normalized)
		if ext == "" || len(ext) > 5 {
			ext = ".jpg"
		}
		// Use a stable key derived from mediaID + index so re-runs are idempotent.
		key := fmt.Sprintf("images/%s_%02d%s", strings.ReplaceAll(mediaID, ":", "_"), i, ext)
		localURL, err := o.uploadAsset(ctx, normalized, key, referer)
		if err != nil {
			log.Printf("enrich image upload [%d] for %s: %v", i, mediaID, err)
			continue
		}
		result.Images[i].FilePath = localURL
	}
}

// uploadAsset fetches assetURL and stores the body at the given MinIO key.
// referer is sent as the HTTP Referer header; pass empty string to omit it.
func (o *Orchestrator) uploadAsset(ctx context.Context, assetURL, key, referer string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return "", err
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := outbound.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("asset fetch %s: HTTP %d", assetURL, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}
	if err := o.store.Put(ctx, key, bytes.NewReader(body), int64(len(body)), ct); err != nil {
		return "", err
	}
	return o.store.DirectURL(key), nil
}

func (o *Orchestrator) uploadPoster(ctx context.Context, externalID, posterURL, referer string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, posterURL, nil)
	if err != nil {
		return "", err
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := outbound.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("poster fetch %s: HTTP %d", posterURL, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/jpeg"
	}
	ext := path.Ext(posterURL)
	if ext == "" || len(ext) > 5 {
		ext = ".jpg"
	}
	key := fmt.Sprintf("posters/%s%s", externalID, ext)
	if err := o.store.Put(ctx, key, bytes.NewReader(body), int64(len(body)), ct); err != nil {
		return "", err
	}
	return o.store.DirectURL(key), nil
}
