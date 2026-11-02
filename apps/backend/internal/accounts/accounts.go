// Package accounts keeps the logins of plugins whose sites need one. The
// user fills the plugin's login view; the plugin turns it into a session,
// which is stored encrypted for this device only and handed to the plugin on
// every call (plugins.AccountProvider). Login values are never stored, and
// neither they nor sessions ever leave the backend through the API.
package accounts

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// accountTable is the table accounts live in.
const accountTable = "Account"

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
	return &Service{database: database, registry: registry, logins: logins, deviceID: deviceID, cipher: sealer}, nil
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
	sealed, err := s.seal(result.Session)
	if err != nil {
		return nil, err
	}
	if _, err := s.database.Exec(ctx, `DELETE FROM "Account" WHERE "plugin" = $plugin AND "device" = $device`,
		db.Vars{"plugin": pluginName, "device": s.deviceID}); err != nil {
		return nil, err
	}
	row := db.Vars{
		"plugin": pluginName, "label": defaultText(result.Label, pluginName), "status": schema.AccountStatusOk,
		"device": s.deviceID, "session": sealed, "lastCheckedAt": db.Now(),
	}
	if result.ExpiresAt != "" {
		row["expiresAt"] = result.ExpiresAt
	}
	id, err := s.database.Insert(ctx, accountTable, row)
	if err != nil {
		return nil, err
	}
	return s.Account(ctx, id)
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
