package accounts

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

// Get returns the creator plugin, which also takes cookies for fans.test.
func (fakeRegistry) Get(name string) (*plugins.Plugin, bool) {
	return &plugins.Plugin{Meta: plugins.PluginMeta{
		Name: name, Capabilities: []plugins.Capability{plugins.CapabilityAccountLogin, plugins.CapabilityAccountCheck},
		Domains: []string{"fans.test"}, Login: &plugins.LoginSpec{Cookies: true},
	}}, true
}

// openTestService returns the service on a fresh database with one device,
// and the database.
func openTestService(t *testing.T) (*Service, *db.DB) {
	t.Helper()
	ctx := context.Background()
	dataDir := t.TempDir()
	database, err := db.Open(ctx, filepath.Join(dataDir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
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
	return service, database
}

// TestLoginAgainWithStoredValues logs in, has the site end the session,
// and expects the stored values to log in again, once for calls that come
// together; a cookie import keeps no values and expires instead.
func TestLoginAgainWithStoredValues(t *testing.T) {
	ctx := context.Background()
	service, database := openTestService(t)
	account, err := service.Login(ctx, "fans", map[string]string{"user": "me", "password": "secret"})
	if err != nil {
		t.Fatal(err)
	}
	stored, _ := database.Strings(ctx, `SELECT CAST("credentials" AS TEXT) FROM "Account"`, nil)
	if len(stored) != 1 || strings.Contains(stored[0], "secret") || !service.StoresLogin(ctx, account.Id) {
		t.Fatalf("credentials stored as %q, want them sealed", stored)
	}
	service.SessionRenewed("fans", "cookie=old")
	service.loggedInAt = map[string]time.Time{}
	if !service.LoginAgain(ctx, "fans") {
		t.Fatal("didn't log in again")
	}
	if !service.LoginAgain(ctx, "fans") {
		t.Fatal("a call right after the login didn't count it")
	}
	if session := service.AccountSession("fans"); session == nil || session.Session != "cookie=abc" {
		t.Errorf("session after logging in again = %+v", session)
	}

	imported, err := service.ImportCookies(ctx, "fans", "Cookie: sid=1")
	if err != nil || service.StoresLogin(ctx, imported.Id) {
		t.Fatalf("imported = %+v, err %v", imported, err)
	}
	service.loggedInAt = map[string]time.Time{}
	if service.LoginAgain(ctx, "fans") {
		t.Error("logged in again without stored values")
	}
	if expired, _ := service.Account(ctx, imported.Id); expired.Status != schema.AccountStatusExpired {
		t.Errorf("status = %s, want expired", expired.Status)
	}
}

// TestImportCookiesKeepsTheSitesOwn reads each export format and keeps only
// the cookies of the plugin's domains.
func TestImportCookiesKeepsTheSitesOwn(t *testing.T) {
	exports := map[string]string{
		"json":      `[{"name":"sid","value":"1","domain":".fans.test","path":"/","expirationDate":1893456000.5},{"name":"other","value":"2","domain":"bank.test"}]`,
		"netscape":  "# Netscape HTTP Cookie File\n#HttpOnly_.fans.test\tTRUE\t/\tTRUE\t1893456000\tsid\t1\nbank.test\tFALSE\t/\tFALSE\t0\tother\t2\n",
		"header":    "Cookie: sid=1",
		"wrapped":   `{"cookies":[{"name":"sid","value":"1","domain":"www.fans.test"}]}`,
		"top level": `[{"name":"sid","value":"1","domain":".test"},{"name":"sid","value":"1","domain":"fans.test"}]`,
	}
	for format, exported := range exports {
		parsed, err := ParseCookies(exported)
		if err != nil {
			t.Fatalf("%s: %v", format, err)
		}
		kept := cookiesFor(parsed, []string{"fans.test", "*"})
		if len(kept) != 1 || kept[0].Name != "sid" || kept[0].Value != "1" {
			t.Errorf("%s: kept %+v", format, kept)
		}
	}
	parsed, _ := ParseCookies(exports["netscape"])
	if !parsed[0].Secure || parsed[0].Expires != 1893456000 || parsed[0].Domain != ".fans.test" {
		t.Errorf("netscape cookie = %+v", parsed[0])
	}
	service, _ := openTestService(t)
	if _, err := service.ImportCookies(context.Background(), "fans", `[{"name":"x","value":"1","domain":"bank.test"}]`); err == nil {
		t.Error("imported cookies of another site")
	}
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
