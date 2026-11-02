// Package accounts keeps the logins of plugins whose sites need one. The
// user fills the plugin's login form, or imports the site's cookies from a
// browser; the session is stored encrypted for this device only and handed
// to the plugin on every call (plugins.AccountProvider). Form values are
// stored encrypted too, to log in again when the site ends the session.
// Neither values nor sessions ever leave the backend through the API.
package accounts

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// accountTable is the table accounts live in.
const accountTable = "Account"

// reloginGrace is how long after a login a refused call is taken for one
// still made with the old session, so it's retried without logging in again.
const reloginGrace = 30 * time.Second

// Logins runs the account capabilities of plugins.
type Logins interface {
	Login(ctx context.Context, plugin *plugins.Plugin, values map[string]string) (*plugins.AccountResult, error)
	CheckAccount(ctx context.Context, plugin *plugins.Plugin, session string) (*plugins.AccountResult, error)
}

// Registry finds installed plugins.
type Registry interface {
	Get(name string) (*plugins.Plugin, bool)
}

// Service stores accounts and supplies their sessions.
type Service struct {
	database *db.DB
	registry Registry
	logins   Logins
	deviceID string
	cipher   cipher.AEAD

	// reloginMutex guards loggedInAt: when each plugin last logged in.
	reloginMutex sync.Mutex
	loggedInAt   map[string]time.Time
}

// New returns the service. key is the 32-byte key sessions are sealed with.
func New(database *db.DB, registry Registry, logins Logins, deviceID string, key []byte) (*Service, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("account key: %w", err)
	}
	sealer, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Service{
		database: database, registry: registry, logins: logins, deviceID: deviceID, cipher: sealer,
		loggedInAt: map[string]time.Time{},
	}, nil
}

// Accounts lists this device's accounts, by plugin.
func (s *Service) Accounts(ctx context.Context) ([]schema.Account, error) {
	return db.QueryAs[schema.Account](ctx, s.database,
		`SELECT * FROM "Account" WHERE "device" = $device AND "deletedAt" IS NULL ORDER BY "plugin", "label"`,
		db.Vars{"device": s.deviceID})
}

// Account loads one account.
func (s *Service) Account(ctx context.Context, id string) (*schema.Account, error) {
	return db.Get[schema.Account](ctx, s.database, accountTable, id)
}

// Login signs into a plugin's site with the values of its login view and
// keeps the session, replacing the plugin's earlier account on this device.
// The values are kept sealed too, so the account can log in again by itself
// when the site ends the session.
func (s *Service) Login(ctx context.Context, pluginName string, values map[string]string) (*schema.Account, error) {
	plugin, err := s.plugin(pluginName)
	if err != nil {
		return nil, err
	}
	result, err := s.logins.Login(ctx, plugin, values)
	if err != nil {
		return nil, err
	}
	if result.Status != string(schema.AccountStatusOk) || result.Session == "" {
		return nil, fmt.Errorf("login failed: %s", defaultText(result.Error, result.Status))
	}
	encoded, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}
	s.markLoggedIn(pluginName)
	return s.store(ctx, pluginName, newAccount{
		label: defaultText(result.Label, pluginName), status: schema.AccountStatusOk,
		session: result.Session, credentials: string(encoded), expiresAt: result.ExpiresAt,
	})
}

// ImportCookies keeps cookies exported from a browser as a plugin's session,
// replacing its earlier account on this device. Only the cookies of the
// plugin's own domains are kept; the plugin checks them when it can.
func (s *Service) ImportCookies(ctx context.Context, pluginName, exported string) (*schema.Account, error) {
	plugin, found := s.registry.Get(pluginName)
	if !found || !plugin.TakesCookies() {
		return nil, fmt.Errorf("%s doesn't take imported cookies", pluginName)
	}
	parsed, err := ParseCookies(exported)
	if err != nil {
		return nil, err
	}
	kept := cookiesFor(parsed, plugin.Meta.Domains)
	if len(kept) == 0 {
		return nil, fmt.Errorf("none of these cookies are for %s", strings.Join(ownDomains(plugin.Meta.Domains), ", "))
	}
	session, err := json.Marshal(plugins.CookieJar{Cookies: kept})
	if err != nil {
		return nil, err
	}
	account := newAccount{label: "Imported cookies", status: schema.AccountStatusOk, session: string(session)}
	s.verify(ctx, plugin, &account)
	return s.store(ctx, pluginName, account)
}

// newAccount is what's stored of a fresh login.
type newAccount struct {
	label       string
	status      schema.AccountStatus
	session     string
	credentials string
	expiresAt   string
}

// verify asks the plugin, when it can, whether an imported session works
// and takes its answer and label.
func (s *Service) verify(ctx context.Context, plugin *plugins.Plugin, account *newAccount) {
	if !plugin.Serves(plugins.CapabilityAccountCheck) {
		return
	}
	result, err := s.logins.CheckAccount(ctx, plugin, account.session)
	if err != nil || result.Status == "" {
		account.status = schema.AccountStatusFailed
		return
	}
	account.status = schema.AccountStatus(result.Status)
	account.label = defaultText(result.Label, account.label)
}

// store seals a fresh login and keeps it as the plugin's only account on
// this device.
func (s *Service) store(ctx context.Context, pluginName string, account newAccount) (*schema.Account, error) {
	sealed, err := s.seal(account.session)
	if err != nil {
		return nil, err
	}
	row := db.Vars{
		"plugin": pluginName, "label": account.label, "status": account.status,
		"device": s.deviceID, "session": sealed, "lastCheckedAt": db.Now(),
	}
	if account.credentials != "" {
		credentials, err := s.seal(account.credentials)
		if err != nil {
			return nil, err
		}
		row["credentials"] = credentials
	}
	if account.expiresAt != "" {
		row["expiresAt"] = account.expiresAt
	}
	if _, err := s.database.Exec(ctx, `DELETE FROM "Account" WHERE "plugin" = $plugin AND "device" = $device`,
		db.Vars{"plugin": pluginName, "device": s.deviceID}); err != nil {
		return nil, err
	}
	id, err := s.database.Insert(ctx, accountTable, row)
	if err != nil {
		return nil, err
	}
	return s.Account(ctx, id)
}

// LoginAgain logs a plugin's account in again with its stored login values
// after the site refused its session. Calls that hit the refusal together
// share one login. Without stored values, or when the login fails, the
// account is marked so the user logs in again, and it returns false.
func (s *Service) LoginAgain(ctx context.Context, pluginName string) bool {
	s.reloginMutex.Lock()
	defer s.reloginMutex.Unlock()
	if time.Since(s.loggedInAt[pluginName]) < reloginGrace {
		return true
	}
	account, err := s.deviceAccount(ctx, pluginName)
	if err != nil || account == nil {
		return false
	}
	values, err := s.credentials(ctx, account.Id)
	if err != nil || values == nil {
		s.setStatus(ctx, account.Id, schema.AccountStatusExpired)
		return false
	}
	if err := s.relogin(ctx, account, values); err != nil {
		log.Printf("account %s: logging in again: %v", pluginName, err)
		s.setStatus(ctx, account.Id, schema.AccountStatusNeedsLogin)
		return false
	}
	s.loggedInAt[pluginName] = time.Now()
	return true
}

// relogin runs the plugin's login with stored values and keeps the new
// session on the account.
func (s *Service) relogin(ctx context.Context, account *schema.Account, values map[string]string) error {
	plugin, err := s.plugin(account.Plugin)
	if err != nil {
		return err
	}
	result, err := s.logins.Login(ctx, plugin, values)
	if err != nil {
		return err
	}
	if result.Status != string(schema.AccountStatusOk) || result.Session == "" {
		return fmt.Errorf("%s", defaultText(result.Error, result.Status))
	}
	sealed, err := s.seal(result.Session)
	if err != nil {
		return err
	}
	fields := db.Vars{"session": sealed, "status": schema.AccountStatusOk, "lastCheckedAt": db.Now()}
	if result.ExpiresAt != "" {
		fields["expiresAt"] = result.ExpiresAt
	}
	return s.database.Update(ctx, accountTable, account.Id, fields)
}

// SessionRenewed keeps the session a plugin renewed during a call (sites
// that rotate their cookies), sealed like the first.
func (s *Service) SessionRenewed(pluginName, session string) {
	ctx := context.Background()
	account, err := s.deviceAccount(ctx, pluginName)
	if err != nil || account == nil {
		return
	}
	sealed, err := s.seal(session)
	if err != nil {
		return
	}
	if err := s.database.Update(ctx, accountTable, account.Id, db.Vars{"session": sealed}); err != nil {
		log.Printf("account %s: keeping the renewed session: %v", pluginName, err)
	}
}

// StoresLogin reports whether an account keeps login values to log in again
// by itself (a form login, not imported cookies).
func (s *Service) StoresLogin(ctx context.Context, id string) bool {
	values, err := s.credentials(ctx, id)
	return err == nil && values != nil
}

// deviceAccount is the plugin's account on this device, whatever its
// status; nil when it has none.
func (s *Service) deviceAccount(ctx context.Context, pluginName string) (*schema.Account, error) {
	return db.QueryOneAs[schema.Account](ctx, s.database,
		`SELECT * FROM "Account" WHERE "plugin" = $plugin AND "device" = $device AND "deletedAt" IS NULL LIMIT 1`,
		db.Vars{"plugin": pluginName, "device": s.deviceID})
}

// credentials opens an account's stored login values; nil when it keeps none.
func (s *Service) credentials(ctx context.Context, id string) (map[string]string, error) {
	rows, err := s.database.Query(ctx, `SELECT "credentials" FROM "Account" WHERE "id" = $id`, db.Vars{"id": id})
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	sealed, isBytes := rows[0]["credentials"].([]byte)
	if !isBytes || len(sealed) == 0 {
		return nil, nil
	}
	opened, err := s.open(sealed)
	if err != nil {
		return nil, err
	}
	var values map[string]string
	return values, json.Unmarshal([]byte(opened), &values)
}

// setStatus records an account's status; a failure to is only logged.
func (s *Service) setStatus(ctx context.Context, id string, status schema.AccountStatus) {
	if err := s.database.Update(ctx, accountTable, id, db.Vars{"status": status, "lastCheckedAt": db.Now()}); err != nil {
		log.Printf("account %s: status %s: %v", id, status, err)
	}
}

// markLoggedIn remembers that the plugin just logged in, so calls refused
// with the old session don't log in once each.
func (s *Service) markLoggedIn(pluginName string) {
	s.reloginMutex.Lock()
	defer s.reloginMutex.Unlock()
	s.loggedInAt[pluginName] = time.Now()
}

// Check asks the plugin whether an account's session still works and
// records the answer.
func (s *Service) Check(ctx context.Context, id string) (*schema.Account, error) {
	account, err := s.Account(ctx, id)
	if err != nil || account == nil {
		return nil, fmt.Errorf("no account %s", id)
	}
	plugin, err := s.plugin(account.Plugin)
	if err != nil {
		return nil, err
	}
	session, err := s.session(ctx, id)
	if err != nil {
		return nil, err
	}
	status := schema.AccountStatusFailed
	result, err := s.logins.CheckAccount(ctx, plugin, session)
	if err == nil && result.Status != "" {
		status = schema.AccountStatus(result.Status)
	}
	if err := s.database.Update(ctx, accountTable, id, db.Vars{"status": status, "lastCheckedAt": db.Now()}); err != nil {
		return nil, err
	}
	return s.Account(ctx, id)
}

// Logout forgets an account and its session for good.
func (s *Service) Logout(ctx context.Context, id string) error {
	return s.database.Delete(ctx, accountTable, id)
}

// AccountSession is the session of the plugin's working account on this
// device, for plugin calls; nil when it has none.
func (s *Service) AccountSession(pluginName string) *plugins.AccountArgs {
	ctx := context.Background()
	rows, err := s.database.Query(ctx,
		`SELECT "id", "label" FROM "Account" WHERE "plugin" = $plugin AND "device" = $device
		 AND "status" = $ok AND "deletedAt" IS NULL LIMIT 1`,
		db.Vars{"plugin": pluginName, "device": s.deviceID, "ok": schema.AccountStatusOk})
	if err != nil || len(rows) == 0 {
		return nil
	}
	session, err := s.session(ctx, fmt.Sprint(rows[0]["id"]))
	if err != nil {
		return nil
	}
	return &plugins.AccountArgs{Label: fmt.Sprint(rows[0]["label"]), Session: session}
}

// session opens an account's stored session.
func (s *Service) session(ctx context.Context, id string) (string, error) {
	rows, err := s.database.Query(ctx, `SELECT "session" FROM "Account" WHERE "id" = $id`, db.Vars{"id": id})
	if err != nil || len(rows) == 0 {
		return "", fmt.Errorf("no account %s", id)
	}
	sealed, isBytes := rows[0]["session"].([]byte)
	if !isBytes {
		return "", fmt.Errorf("account %s has no session", id)
	}
	return s.open(sealed)
}

// plugin finds an installed plugin that can log in.
func (s *Service) plugin(name string) (*plugins.Plugin, error) {
	plugin, found := s.registry.Get(name)
	if !found {
		return nil, fmt.Errorf("plugin %q isn't installed", name)
	}
	if !plugin.Serves(plugins.CapabilityAccountLogin) {
		return nil, fmt.Errorf("%s has no login", name)
	}
	return plugin, nil
}

// seal encrypts a session: a fresh nonce, then the sealed text.
func (s *Service) seal(session string) ([]byte, error) {
	nonce := make([]byte, s.cipher.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return s.cipher.Seal(nonce, nonce, []byte(session), nil), nil
}

// open decrypts a sealed session.
func (s *Service) open(sealed []byte) (string, error) {
	if len(sealed) < s.cipher.NonceSize() {
		return "", fmt.Errorf("sealed session too short")
	}
	nonce, text := sealed[:s.cipher.NonceSize()], sealed[s.cipher.NonceSize():]
	opened, err := s.cipher.Open(nil, nonce, text, nil)
	if err != nil {
		return "", fmt.Errorf("session can't be opened (key changed?): %w", err)
	}
	return string(opened), nil
}

// defaultText returns value, or fallback when it's blank.
func defaultText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
