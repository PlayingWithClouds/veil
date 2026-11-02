package notify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// openTestService opens a notification service on a fresh database.
func openTestService(t *testing.T) *Service {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return New(database)
}

// capturedRequest is what a test server received.
type capturedRequest struct {
	Path    string
	Headers http.Header
	Body    string
}

// captureServer starts a server that records every request it gets.
func captureServer(t *testing.T) (*httptest.Server, func() []capturedRequest) {
	t.Helper()
	var mutex sync.Mutex
	var received []capturedRequest
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		mutex.Lock()
		received = append(received, capturedRequest{Path: request.URL.Path, Headers: request.Header.Clone(), Body: string(body)})
		mutex.Unlock()
		writer.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	return server, func() []capturedRequest {
		mutex.Lock()
		defer mutex.Unlock()
		return append([]capturedRequest{}, received...)
	}
}

// TestInbox stores notifications, counts and lists the unread ones, and
// marks them read.
func TestInbox(t *testing.T) {
	ctx := context.Background()
	service := openTestService(t)
	first, err := service.Notify(ctx, NewNotification{Kind: "newMedia", Title: "3 new"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Notify(ctx, NewNotification{Kind: "camLive", Title: "Live now", Subject: &schema.RecordRef{Kind: schema.RecordKindCamModel, Id: "m1"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Notify(ctx, NewNotification{Kind: "newMedia"}); err == nil {
		t.Error("a notification without a title was stored")
	}
	assertUnread(t, service, 2)
	if err := service.MarkRead(ctx, []string{first.Id}); err != nil {
		t.Fatal(err)
	}
	unread, err := service.Notifications(ctx, true, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(unread) != 1 || unread[0].Kind != "camLive" || unread[0].Subject == nil || unread[0].Subject.Id != "m1" {
		t.Errorf("unread = %+v, want the camLive one", unread)
	}
	if err := service.MarkAllRead(ctx); err != nil {
		t.Fatal(err)
	}
	assertUnread(t, service, 0)
	all, _ := service.Notifications(ctx, false, 10, 0)
	if len(all) != 2 || !all[0].Read {
		t.Errorf("all = %+v, want two read notifications", all)
	}
}

// assertUnread expects count unread notifications.
func assertUnread(t *testing.T, service *Service, count int) {
	t.Helper()
	unread, err := service.UnreadCount(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if unread != count {
		t.Errorf("unread = %d, want %d", unread, count)
	}
}

// TestChannelValidation refuses configs a channel kind can't work with and
// edits a valid one.
func TestChannelValidation(t *testing.T) {
	ctx := context.Background()
	service := openTestService(t)
	if _, err := service.CreateChannel(ctx, NewChannel{Kind: schema.ChannelKindNtfy, Name: "Phone", Config: map[string]any{}}); err == nil {
		t.Error("ntfy channel without a topic was accepted")
	}
	if _, err := service.CreateChannel(ctx, NewChannel{Kind: schema.ChannelKindWebhook, Name: "Hook", Config: map[string]any{"url": "ftp://x"}}); err == nil {
		t.Error("webhook with a non-http URL was accepted")
	}
	channel, err := service.CreateChannel(ctx, NewChannel{Kind: schema.ChannelKindPush, Name: "Pixel", Config: map[string]any{"device": "d1"}, Enabled: true})
	if err != nil {
		t.Fatal(err)
	}
	disabled := false
	edited, err := service.UpdateChannel(ctx, channel.Id, ChannelEdit{Enabled: &disabled})
	if err != nil {
		t.Fatal(err)
	}
	if edited.Enabled {
		t.Error("channel still enabled")
	}
	if err := service.DeleteChannel(ctx, channel.Id); err != nil {
		t.Fatal(err)
	}
	channels, _ := service.Channels(ctx)
	if len(channels) != 0 {
		t.Errorf("channels after delete = %+v", channels)
	}
}

// TestDeliveryIsNeutral sends to ntfy and a webhook: without detail they get
// neutral wording, with detail the real title.
func TestDeliveryIsNeutral(t *testing.T) {
	ctx := context.Background()
	service := openTestService(t)
	server, received := captureServer(t)
	createChannel(t, service, NewChannel{Kind: schema.ChannelKindNtfy, Name: "Phone", Enabled: true,
		Config: map[string]any{"server": server.URL, "topic": "secret", "token": "tk"}})
	createChannel(t, service, NewChannel{Kind: schema.ChannelKindWebhook, Name: "Hook", Enabled: true,
		Config: map[string]any{"url": server.URL + "/hook", "headers": map[string]any{"X-Key": "k"}, "detail": true}})
	createChannel(t, service, NewChannel{Kind: schema.ChannelKindWebhook, Name: "Off", Enabled: false,
		Config: map[string]any{"url": server.URL + "/off"}})

	subject := schema.RecordRef{Kind: schema.RecordKindVideo, Id: "v1"}
	if _, err := service.Notify(ctx, NewNotification{Kind: "newMedia", Title: "Explicit title", Body: "Explicit body", Subject: &subject}); err != nil {
		t.Fatal(err)
	}
	service.Wait()

	requests := received()
	if len(requests) != 2 {
		t.Fatalf("got %d requests, want 2: %+v", len(requests), requests)
	}
	for _, request := range requests {
		switch request.Path {
		case "/secret":
			assertNtfy(t, request)
		case "/hook":
			assertWebhook(t, request)
		default:
			t.Errorf("unexpected request to %s", request.Path)
		}
	}
}

// createChannel adds a channel or fails the test.
func createChannel(t *testing.T, service *Service, channel NewChannel) {
	t.Helper()
	if _, err := service.CreateChannel(context.Background(), channel); err != nil {
		t.Fatal(err)
	}
}

// assertNtfy expects a neutral ntfy message with the bearer token.
func assertNtfy(t *testing.T, request capturedRequest) {
	t.Helper()
	if request.Headers.Get("Title") != neutralTitle || request.Body != neutralBodies["newMedia"] {
		t.Errorf("ntfy got title %q body %q, want neutral wording", request.Headers.Get("Title"), request.Body)
	}
	if request.Headers.Get("Authorization") != "Bearer tk" {
		t.Errorf("ntfy authorization = %q", request.Headers.Get("Authorization"))
	}
}

// assertWebhook expects the detailed JSON payload with the custom header.
func assertWebhook(t *testing.T, request capturedRequest) {
	t.Helper()
	var payload webhookPayload
	if err := json.Unmarshal([]byte(request.Body), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Title != "Explicit title" || payload.Subject == nil || payload.Subject.Id != "v1" {
		t.Errorf("webhook payload = %+v, want the details", payload)
	}
	if request.Headers.Get("X-Key") != "k" {
		t.Errorf("webhook header X-Key = %q", request.Headers.Get("X-Key"))
	}
}

// TestSubscribe receives notifications stored after subscribing, until
// cancelled.
func TestSubscribe(t *testing.T) {
	ctx := context.Background()
	service := openTestService(t)
	received, cancel := service.Subscribe()
	if _, err := service.Notify(ctx, NewNotification{Kind: "downloadDone", Title: "Done"}); err != nil {
		t.Fatal(err)
	}
	select {
	case notification := <-received:
		if notification.Kind != "downloadDone" {
			t.Errorf("received %+v", notification)
		}
	case <-time.After(time.Second):
		t.Fatal("no notification received")
	}
	cancel()
	cancel()
	if _, open := <-received; open {
		t.Error("channel still open after cancel")
	}
}
