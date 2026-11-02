package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// defaultNtfyServer is used when an ntfy channel names no server.
const defaultNtfyServer = "https://ntfy.sh"

// NtfyConfig is the config of an ntfy channel.
type NtfyConfig struct {
	// Server defaults to defaultNtfyServer.
	Server string `json:"server,omitempty"`
	Topic  string `json:"topic"`
	// Token is sent as a bearer token, for protected topics.
	Token string `json:"token,omitempty"`
	// Detail opts in to real titles and bodies instead of neutral wording.
	Detail bool `json:"detail,omitempty"`
}

// WebhookConfig is the config of a webhook channel.
type WebhookConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	// Detail opts in to real titles, bodies and subjects.
	Detail bool `json:"detail,omitempty"`
}

// NewChannel is a channel to add. Config is the kind's config as JSON
// (NtfyConfig, WebhookConfig); push channels store what their clients need.
type NewChannel struct {
	Kind    schema.ChannelKind
	Name    string
	Config  map[string]any
	Enabled bool
}

// ChannelEdit changes a channel; nil fields stay as they are.
type ChannelEdit struct {
	Name    *string
	Config  *map[string]any
	Enabled *bool
}

// Delivers reports whether the server delivers to channels of kind. Push
// channels are stored for the clients, which deliver them themselves; the
// in-app inbox needs no delivery.
func Delivers(kind schema.ChannelKind) bool {
	return kind == schema.ChannelKindNtfy || kind == schema.ChannelKindWebhook
}

// CreateChannel adds a notification channel after checking its config.
func (s *Service) CreateChannel(ctx context.Context, channel NewChannel) (*schema.NotificationChannel, error) {
	name := strings.TrimSpace(channel.Name)
	if name == "" {
		return nil, fmt.Errorf("a channel needs a name")
	}
	if err := validateConfig(channel.Kind, channel.Config); err != nil {
		return nil, err
	}
	row := schema.NotificationChannel{Kind: channel.Kind, Name: name, Enabled: channel.Enabled}
	if channel.Config != nil {
		row.Config = channel.Config
	}
	id, err := s.database.Insert(ctx, channelTable, row)
	if err != nil {
		return nil, err
	}
	return s.Channel(ctx, id)
}

// UpdateChannel applies an edit to a channel.
func (s *Service) UpdateChannel(ctx context.Context, id string, edit ChannelEdit) (*schema.NotificationChannel, error) {
	channel, err := s.Channel(ctx, id)
	if err != nil || channel == nil {
		return nil, missingChannel(id, err)
	}
	fields := db.Vars{}
	if edit.Name != nil {
		name := strings.TrimSpace(*edit.Name)
		if name == "" {
			return nil, fmt.Errorf("a channel needs a name")
		}
		fields["name"] = name
	}
	if edit.Config != nil {
		if err := validateConfig(channel.Kind, *edit.Config); err != nil {
			return nil, err
		}
		fields["config"] = *edit.Config
	}
	if edit.Enabled != nil {
		fields["enabled"] = *edit.Enabled
	}
	if err := s.database.Update(ctx, channelTable, id, fields); err != nil {
		return nil, err
	}
	return s.Channel(ctx, id)
}

// DeleteChannel moves a channel to the trash.
func (s *Service) DeleteChannel(ctx context.Context, id string) error {
	return s.database.Trash(ctx, channelTable, id)
}

// Channel loads one channel; nil when there is none.
func (s *Service) Channel(ctx context.Context, id string) (*schema.NotificationChannel, error) {
	return db.Get[schema.NotificationChannel](ctx, s.database, channelTable, id)
}

// Channels lists every channel, oldest first.
func (s *Service) Channels(ctx context.Context) ([]schema.NotificationChannel, error) {
	return db.QueryAs[schema.NotificationChannel](ctx, s.database,
		`SELECT * FROM "NotificationChannel" WHERE "deletedAt" IS NULL ORDER BY "createdAt", rowid`, nil)
}

// enabledChannels lists the enabled channels the server delivers to.
func (s *Service) enabledChannels(ctx context.Context) ([]schema.NotificationChannel, error) {
	return db.QueryAs[schema.NotificationChannel](ctx, s.database,
		`SELECT * FROM "NotificationChannel" WHERE "deletedAt" IS NULL AND "enabled" = 1
		 AND "kind" IN ($ntfy, $webhook) ORDER BY "createdAt", rowid`,
		db.Vars{"ntfy": schema.ChannelKindNtfy, "webhook": schema.ChannelKindWebhook})
}

// missingChannel is the error for a channel that couldn't be loaded.
func missingChannel(id string, err error) error {
	if err != nil {
		return err
	}
	return fmt.Errorf("no notification channel %s", id)
}

// validateConfig checks that a channel's config has what its kind needs.
func validateConfig(kind schema.ChannelKind, config map[string]any) error {
	switch kind {
	case schema.ChannelKindInApp, schema.ChannelKindPush:
		return nil
	case schema.ChannelKindNtfy:
		return validateNtfy(config)
	case schema.ChannelKindWebhook:
		return validateWebhook(config)
	}
	return fmt.Errorf("unknown channel kind %q", kind)
}

// validateNtfy checks an ntfy config: a topic, and a server that is a URL.
func validateNtfy(config map[string]any) error {
	var ntfy NtfyConfig
	if err := decodeConfig(config, &ntfy); err != nil {
		return err
	}
	if strings.TrimSpace(ntfy.Topic) == "" {
		return fmt.Errorf("an ntfy channel needs a topic")
	}
	if ntfy.Server == "" {
		return nil
	}
	return validateURL(ntfy.Server)
}

// validateWebhook checks a webhook config: an http(s) URL.
func validateWebhook(config map[string]any) error {
	var webhook WebhookConfig
	if err := decodeConfig(config, &webhook); err != nil {
		return err
	}
	return validateURL(webhook.URL)
}

// validateURL accepts absolute http and https URLs.
func validateURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("%q isn't an http(s) URL", raw)
	}
	return nil
}

// decodeConfig reads a channel's stored config into target.
func decodeConfig(config any, target any) error {
	encoded, err := json.Marshal(config)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(encoded, target); err != nil {
		return fmt.Errorf("invalid channel config: %w", err)
	}
	return nil
}
