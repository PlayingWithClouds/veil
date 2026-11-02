package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"

	"github.com/playingwithclouds/veil/internal/schema"
)

// neutralTitle is the title every delivery without details carries.
const neutralTitle = "Veil"

// neutralBodies word each kind of notification without naming anything.
var neutralBodies = map[string]string{
	"newMedia":         "New media is available.",
	"camLive":          "Someone you follow is live.",
	"downloadDone":     "A download finished.",
	"automationFailed": "An automation failed.",
	"wantedFound":      "Something you wanted was found.",
	"limitReached":     "A usage limit was reached.",
}

// neutralDefault words the kinds neutralBodies doesn't know.
const neutralDefault = "You have a new notification."

// deliver sends a notification to one channel and logs a failure.
func (s *Service) deliver(ctx context.Context, channel schema.NotificationChannel, notification schema.Notification) {
	var err error
	switch channel.Kind {
	case schema.ChannelKindNtfy:
		err = s.sendNtfy(ctx, channel, notification)
	case schema.ChannelKindWebhook:
		err = s.sendWebhook(ctx, channel, notification)
	default:
		return
	}
	if err != nil {
		log.Printf("notify: deliver to %s channel %q: %v", channel.Kind, channel.Name, err)
	}
}

// sendNtfy publishes a notification to an ntfy topic: the body as the
// message, the title as the Title header.
func (s *Service) sendNtfy(ctx context.Context, channel schema.NotificationChannel, notification schema.Notification) error {
	var config NtfyConfig
	if err := decodeConfig(channel.Config, &config); err != nil {
		return err
	}
	server := config.Server
	if server == "" {
		server = defaultNtfyServer
	}
	title, body := wording(notification, config.Detail)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(server, "/")+"/"+config.Topic, strings.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Title", headerText(title))
	if config.Token != "" {
		request.Header.Set("Authorization", "Bearer "+config.Token)
	}
	return s.send(request)
}

// webhookPayload is the JSON a webhook receives.
type webhookPayload struct {
	ID      string            `json:"id"`
	At      string            `json:"at"`
	Kind    string            `json:"kind"`
	Title   string            `json:"title"`
	Body    string            `json:"body"`
	Subject *schema.RecordRef `json:"subject,omitempty"`
}

// sendWebhook posts a notification as JSON to a webhook URL, with the
// configured headers.
func (s *Service) sendWebhook(ctx context.Context, channel schema.NotificationChannel, notification schema.Notification) error {
	var config WebhookConfig
	if err := decodeConfig(channel.Config, &config); err != nil {
		return err
	}
	title, body := wording(notification, config.Detail)
	payload := webhookPayload{ID: notification.Id, At: notification.At, Kind: notification.Kind, Title: title, Body: body}
	if config.Detail {
		payload.Subject = notification.Subject
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, config.URL, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	for name, value := range config.Headers {
		request.Header.Set(name, value)
	}
	return s.send(request)
}

// send runs a request and turns a non-2xx answer into an error.
func (s *Service) send(request *http.Request) error {
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	io.Copy(io.Discard, io.LimitReader(response.Body, 64*1024))
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("%s answered %s", request.URL.Host, response.Status)
	}
	return nil
}

// wording is the title and body a channel sees: the real ones when it opted
// in to details, else neutral text by kind.
func wording(notification schema.Notification, detail bool) (string, string) {
	if detail {
		return notification.Title, bodyOf(notification)
	}
	body, known := neutralBodies[notification.Kind]
	if !known {
		body = neutralDefault
	}
	return neutralTitle, body
}

// bodyOf is a notification's body, or "" when it has none.
func bodyOf(notification schema.Notification) string {
	if notification.Body == nil {
		return ""
	}
	return *notification.Body
}

// headerText encodes text for an HTTP header: non-ASCII titles as an
// RFC 2047 word, which ntfy decodes.
func headerText(text string) string {
	for _, character := range text {
		if character > 127 {
			return mime.QEncoding.Encode("utf-8", text)
		}
	}
	return text
}
