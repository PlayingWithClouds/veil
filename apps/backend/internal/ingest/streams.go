package ingest

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// streamRow is the insert shape of a `stream` row (see internal/media/stream.go).
// created_at/updated_at are left to their column defaults.
type streamRow struct {
	Media            db.RecordID `json:"media"`
	URL              string      `json:"url"`
	Kind             string      `json:"kind"`
	Label            *string     `json:"label,omitempty"`
	Provider         *string     `json:"provider,omitempty"`
	Resolution       *string     `json:"resolution,omitempty"`
	Language         *string     `json:"language,omitempty"`
	Format           *string     `json:"format,omitempty"`
	ExpectedSpeedBps *float64    `json:"expected_speed_bps,omitempty"`
	FileSizeBytes    *float64    `json:"file_size_bytes,omitempty"`
	Verified         bool        `json:"verified"`
	PluginName       *string     `json:"plugin_name,omitempty"`
}

// syncStreamsLogged runs SyncStreams and logs failures instead of aborting the
// ingest loop — a stream write must not lose the observation itself.
func (s *Service) syncStreamsLogged(ctx context.Context, canonicalID, pluginName string, downloads []plugins.Download) {
	if err := s.SyncStreams(ctx, canonicalID, pluginName, downloads); err != nil {
		log.Printf("ingest: sync streams for %s: %v", canonicalID, err)
	}
}

// SyncStreams upserts stream records for a canonical media record from the
// downloads a plugin scraped. Existing rows are matched by URL so re-scrapes
// stay idempotent; rows created by other plugins or by hand are left alone.
func (s *Service) SyncStreams(ctx context.Context, canonicalID, pluginName string, downloads []plugins.Download) error {
	if len(downloads) == 0 {
		return nil
	}
	mediaRID, err := db.ParseRecordID(canonicalID)
	if err != nil {
		return fmt.Errorf("invalid media id %q: %w", canonicalID, err)
	}

	existingURLs, err := s.database.Strings(ctx,
		`SELECT url FROM stream WHERE media = $media`,
		db.Vars{"media": *mediaRID})
	if err != nil {
		return fmt.Errorf("list existing streams: %w", err)
	}

	existingByURL := make(map[string]bool, len(existingURLs))
	for _, existingURL := range existingURLs {
		existingByURL[existingURL] = true
	}

	for _, download := range downloads {
		if download.URL == "" || existingByURL[download.URL] {
			continue
		}
		existingByURL[download.URL] = true
		row := downloadToStreamRow(*mediaRID, pluginName, download)
		if _, err := s.database.Insert(ctx, "stream", row); err != nil {
			return fmt.Errorf("create stream for %s: %w", download.URL, err)
		}
	}
	return nil
}

func downloadToStreamRow(media db.RecordID, pluginName string, download plugins.Download) streamRow {
	row := streamRow{
		Media: media,
		URL:   download.URL,
		Kind:  "stream",
	}
	if download.Label != "" {
		row.Label = &download.Label
	}
	if provider := providerFromURL(download.URL); provider != "" {
		row.Provider = &provider
	}
	if download.Quality != "" {
		row.Resolution = &download.Quality
	}
	if download.Language != "" {
		row.Language = &download.Language
	}
	if download.Format != "" {
		row.Format = &download.Format
	}
	if download.SpeedBps > 0 {
		speed := download.SpeedBps
		row.ExpectedSpeedBps = &speed
	}
	if download.SizeBytes > 0 {
		size := float64(download.SizeBytes)
		row.FileSizeBytes = &size
	}
	if pluginName != "" {
		row.PluginName = &pluginName
	}
	return row
}

func providerFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(parsed.Hostname(), "www.")
}
