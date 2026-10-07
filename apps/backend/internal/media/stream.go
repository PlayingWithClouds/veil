package media

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// streamRecord mirrors the `stream` table: a stored playback/download source for
// a movie or episode.
type streamRecord struct {
	ID               *db.RecordID `json:"id,omitempty"`
	Media            db.RecordID  `json:"media"`
	URL              string       `json:"url"`
	Kind             string       `json:"kind"`
	Label            *string      `json:"label,omitempty"`
	Provider         *string      `json:"provider,omitempty"`
	Resolution       *string      `json:"resolution,omitempty"`
	Width            *int         `json:"width,omitempty"`
	Height           *int         `json:"height,omitempty"`
	Language         *string      `json:"language,omitempty"`
	Format           *string      `json:"format,omitempty"`
	MimeType         *string      `json:"mime_type,omitempty"`
	ExpectedSpeedBps *float64     `json:"expected_speed_bps,omitempty"`
	FileSizeBytes    *float64     `json:"file_size_bytes,omitempty"`
	Verified         bool         `json:"verified"`
	PluginName       *string      `json:"plugin_name,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

func (s streamRecord) ToModel() *model.Stream {
	return &model.Stream{
		ID:               recordIDString(s.ID),
		MediaID:          s.Media.String(),
		URL:              s.URL,
		Kind:             s.Kind,
		Label:            s.Label,
		Provider:         s.Provider,
		Resolution:       s.Resolution,
		Width:            s.Width,
		Height:           s.Height,
		Language:         s.Language,
		Format:           s.Format,
		MimeType:         s.MimeType,
		ExpectedSpeedBps: s.ExpectedSpeedBps,
		FileSizeBytes:    s.FileSizeBytes,
		Verified:         s.Verified,
		PluginName:       s.PluginName,
		CreatedAt:        s.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:        s.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// SceneOrigin returns the plugin name and source URL of an observation for a
// scene — which plugin discovered it and from what URL. Used to lazily scrape
// stubs (discovered-but-never-scraped scenes) that have no streams. Any
// observation works: the plugin that can scrape a scene is the same across them.
func (r *Repository) SceneOrigin(ctx context.Context, sceneID string) (pluginName, sourceURL string, err error) {
	rid, err := db.ParseRecordID(sceneID)
	if err != nil {
		return "", "", fmt.Errorf("invalid scene id: %w", err)
	}
	row, err := r.database.QueryRow(ctx,
		`SELECT plugin, source_url FROM observation WHERE target = $scene LIMIT 1`,
		db.Vars{"scene": *rid})
	if err != nil {
		return "", "", fmt.Errorf("scene origin: %w", err)
	}
	if row == nil {
		return "", "", nil
	}
	return mString(row, "plugin"), mString(row, "source_url"), nil
}

// ListStreams returns stored streams for a movie or episode, best sources first.
func (r *Repository) ListStreams(ctx context.Context, mediaID string) ([]*model.Stream, error) {
	rid, err := db.ParseRecordID(mediaID)
	if err != nil {
		return nil, fmt.Errorf("invalid media id: %w", err)
	}
	rows, err := db.QueryAs[streamRecord](ctx, r.database,
		`SELECT * FROM stream WHERE media = $media ORDER BY verified DESC, expected_speed_bps DESC`,
		db.Vars{"media": *rid})
	if err != nil {
		return nil, fmt.Errorf("list streams: %w", err)
	}

	streams := make([]*model.Stream, 0, len(rows))
	for _, rec := range rows {
		streams = append(streams, rec.ToModel())
	}
	return streams, nil
}

// CreateStream stores a new stream source for a media record.
func (r *Repository) CreateStream(ctx context.Context, input model.CreateStreamInput) (*model.Stream, error) {
	mediaRID, err := db.ParseRecordID(input.MediaID)
	if err != nil {
		return nil, fmt.Errorf("invalid media id: %w", err)
	}

	fields := map[string]any{
		"media":              *mediaRID,
		"url":                input.URL,
		"kind":               "stream",
		"label":              input.Label,
		"provider":           input.Provider,
		"resolution":         input.Resolution,
		"width":              input.Width,
		"height":             input.Height,
		"language":           input.Language,
		"format":             input.Format,
		"mime_type":          input.MimeType,
		"expected_speed_bps": input.ExpectedSpeedBps,
		"file_size_bytes":    input.FileSizeBytes,
		"verified":           false,
		"plugin_name":        input.PluginName,
	}
	if input.Kind != nil && *input.Kind != "" {
		fields["kind"] = *input.Kind
	}
	if input.Verified != nil {
		fields["verified"] = *input.Verified
	}

	id, err := r.database.Insert(ctx, "stream", fields)
	if err != nil {
		return nil, fmt.Errorf("create stream: %w", err)
	}
	created, err := db.QueryOneAs[streamRecord](ctx, r.database,
		`SELECT * FROM stream WHERE id = $id`, db.Vars{"id": id})
	if err != nil {
		return nil, fmt.Errorf("create stream: %w", err)
	}
	if created == nil {
		return nil, fmt.Errorf("create stream: no record returned")
	}
	return created.ToModel(), nil
}

// DeleteStream removes a stream by record id.
func (r *Repository) DeleteStream(ctx context.Context, id string) (bool, error) {
	rid, err := db.ParseRecordID(id)
	if err != nil {
		return false, fmt.Errorf("invalid stream id: %w", err)
	}
	if err := r.database.Delete(ctx, *rid); err != nil {
		return false, err
	}
	return true, nil
}
