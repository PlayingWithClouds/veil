package accounts

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// fakeSite logs in with one password and accepts its session.
type fakeSite struct{}

// Login accepts the password "secret".
func (fakeSite) Login(ctx context.Context, plugin *plugins.Plugin, values map[string]string) (*plugins.AccountResult, error) {
	if values["password"] != "secret" {
		return &plugins.AccountResult{Status: "failed", Error: "wrong password"}, nil
	}
	return &plugins.AccountResult{Status: "ok", Label: values["user"], Session: "cookie=abc"}, nil
}

// CheckAccount says the session still works.
func (fakeSite) CheckAccount(ctx context.Context, plugin *plugins.Plugin, session string) (*plugins.AccountResult, error) {
	if session != "cookie=abc" {
		return &plugins.AccountResult{Status: "expired"}, nil
	}
	return &plugins.AccountResult{Status: "ok"}, nil
}

// fakeRegistry knows one plugin with a login.
type fakeRegistry struct{}

// Get returns the creator plugin.
func (fakeRegistry) Get(name string) (*plugins.Plugin, bool) {
	return &plugins.Plugin{Meta: plugins.PluginMeta{Name: name, Capabilities: []plugins.Capability{plugins.CapabilityAccountLogin, plugins.CapabilityAccountCheck}}}, true
}

// TestLoginKeepsTheSessionSealed logs in, reads the session back for the
// plugin, checks it, and rejects a wrong password.
func TestLoginKeepsTheSessionSealed(t *testing.T) {
	ctx := context.Background()
	dataDir := t.TempDir()
	database, err := db.Open(ctx, filepath.Join(dataDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	deviceID, err := database.Insert(ctx, "Device", db.Vars{"name": "nas", "platform": "server", "current": true})
	if err != nil {
		t.Fatal(err)
	}
	key, err := LoadKey(dataDir, nil)
	if err != nil {
		t.Fatal(err)
	}
	service, err := New(database, fakeRegistry{}, fakeSite{}, deviceID, key)
	if err != nil {
		t.Fatal(err)
	}
	account, err := service.Login(ctx, "fans", map[string]string{"user": "me", "password": "secret"})
	if err != nil || account.Label != "me" || account.Status != schema.AccountStatusOk {
		t.Fatalf("account = %+v, err %v", account, err)
	}
	stored, _ := database.Strings(ctx, `SELECT CAST("session" AS TEXT) FROM "Account"`, nil)
	if len(stored) != 1 || stored[0] == "cookie=abc" {
		t.Errorf("session stored as %q, want it sealed", stored)
	}
	if session := service.AccountSession("fans"); session == nil || session.Session != "cookie=abc" {
		t.Errorf("session = %+v", session)
	}
	if checked, err := service.Check(ctx, account.Id); err != nil || checked.Status != schema.AccountStatusOk {
		t.Errorf("checked = %+v, err %v", checked, err)
	}
	if _, err := service.Login(ctx, "fans", map[string]string{"password": "nope"}); err == nil {
		t.Error("a wrong password logged in")
	}
	again, _ := LoadKey(dataDir, nil)
	if string(again) != string(key) {
		t.Error("the key file wasn't reused")
	}
}
