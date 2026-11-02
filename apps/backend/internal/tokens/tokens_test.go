package tokens

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

// openTestService opens the service on a fresh database.
func openTestService(t *testing.T) *Service {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return New(database)
}

// TestTokenLifecycle creates a token, authenticates with it, and expects a
// revoked and an expired token to fail.
func TestTokenLifecycle(t *testing.T) {
	ctx := context.Background()
	service := openTestService(t)
	secret, token, err := service.Create(ctx, "phone", []string{ScopeRead, ScopeStream}, nil)
	if err != nil {
		t.Fatal(err)
	}
	scopes, ok := service.Authenticate(ctx, secret)
	if !ok || !Allows(scopes, ScopeStream) || Allows(scopes, ScopeWrite) {
		t.Fatalf("scopes = %v, ok %v", scopes, ok)
	}
	if err := service.Revoke(ctx, token.Id); err != nil {
		t.Fatal(err)
	}
	if _, ok := service.Authenticate(ctx, secret); ok {
		t.Error("a revoked token still works")
	}
	past := time.Now().Add(-time.Hour)
	expired, _, err := service.Create(ctx, "old", []string{ScopeRead}, &past)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := service.Authenticate(ctx, expired); ok {
		t.Error("an expired token still works")
	}
	if _, _, err := service.Create(ctx, "bad", []string{"everything"}, nil); err == nil {
		t.Error("an unknown scope was accepted")
	}
}

// TestMiddleware lets loopback through, refuses a remote caller without a
// token, and checks the stream scope on media routes.
func TestMiddleware(t *testing.T) {
	ctx := context.Background()
	service := openTestService(t)
	readOnly, _, _ := service.Create(ctx, "reader", []string{ScopeRead}, nil)
	handler := service.Middleware(func() bool { return true }, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	request := func(remote, path, token string) int {
		incoming := httptest.NewRequest(http.MethodGet, path, nil)
		incoming.RemoteAddr = remote
		if token != "" {
			incoming.Header.Set("Authorization", "Bearer "+token)
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, incoming)
		return recorder.Code
	}
	if code := request("127.0.0.1:5000", "/graphql", ""); code != http.StatusNoContent {
		t.Errorf("loopback = %d", code)
	}
	if code := request("192.168.1.20:5000", "/graphql", ""); code != http.StatusUnauthorized {
		t.Errorf("remote without token = %d", code)
	}
	if code := request("192.168.1.20:5000", "/graphql", readOnly); code != http.StatusNoContent {
		t.Errorf("remote with token = %d", code)
	}
	if code := request("192.168.1.20:5000", "/api/stream/abc", readOnly); code != http.StatusForbidden {
		t.Errorf("stream without the scope = %d", code)
	}
}
