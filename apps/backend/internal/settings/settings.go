package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
)

const (
	scopeGlobal = "global"
	scopePlugin = "plugin"

	// pluginSettingsKey is the Settings field stored as plugin-scoped rows
	// instead of a single global row.
	pluginSettingsKey = "pluginSettings"
	updatedAtKey      = "updatedAt"
)

// KindRateLimit controls concurrency and exponential retry backoff for a specific job kind.
// MaxConcurrent=0 means fall back to the global MaxConcurrentJobs setting.
type KindRateLimit struct {
	MaxConcurrent   int     `json:"maxConcurrent"`
	RetryInitialMs  int     `json:"retryInitialMs"`  // delay before first retry
	RetryMultiplier float64 `json:"retryMultiplier"` // each retry multiplies the previous delay
	RetryMaxMs      int     `json:"retryMaxMs"`      // upper cap on retry delay
}

// Settings holds all user-configurable values. Each field is one global
// setting row keyed by its JSON name; PluginSettings are plugin-scoped rows.
type Settings struct {
	ID *db.RecordID `json:"id,omitempty"`

	// Job execution
	MaxConcurrentJobs int `json:"maxConcurrentJobs"`
	MaxJobRetries     int `json:"maxJobRetries"`

	// Downloads
	DownloadSpeedLimitKBps int `json:"downloadSpeedLimitKBps"` // 0 = unlimited
	// When false, download jobs are held while a stream is actively playing so
	// playback bandwidth isn't starved.
	AllowDownloadsWhileStreaming bool `json:"allowDownloadsWhileStreaming"`

	// Automation
	AutoEnrichAfterScrape bool `json:"autoEnrichAfterScrape"`

	// VPN
	RequireVpn bool `json:"requireVpn"` // block job execution when no VPN interface is up

	// Plugins
	DisabledPlugins []string `json:"disabledPlugins"`

	// Per-plugin settings: pluginName → key → value. Values are injected as env vars.
	PluginSettings map[string]map[string]string `json:"pluginSettings"`

	// Per-kind rate limiting and backoff (key = job kind string).
	KindLimits map[string]KindRateLimit `json:"kindLimits"`

	UpdatedAt time.Time `json:"updatedAt"`
}

var defaultKindLimits = map[string]KindRateLimit{
	"scrape":   {MaxConcurrent: 4, RetryInitialMs: 1000, RetryMultiplier: 2.0, RetryMaxMs: 30000},
	"enrich":   {MaxConcurrent: 2, RetryInitialMs: 2000, RetryMultiplier: 2.0, RetryMaxMs: 60000},
	"refresh":  {MaxConcurrent: 2, RetryInitialMs: 5000, RetryMultiplier: 2.0, RetryMaxMs: 60000},
	"download": {MaxConcurrent: 2, RetryInitialMs: 3000, RetryMultiplier: 1.5, RetryMaxMs: 60000},
}

// defaultValues seeds global setting rows that don't exist yet.
var defaultValues = map[string]any{
	"maxConcurrentJobs":            4,
	"maxJobRetries":                3,
	"downloadSpeedLimitKBps":       0,
	"allowDownloadsWhileStreaming": true,
	"autoEnrichAfterScrape":        false,
	"requireVpn":                   false,
	"disabledPlugins":              []string{},
	"kindLimits":                   defaultKindLimits,
}

// Service manages the setting rows and caches the decoded Settings.
type Service struct {
	database *db.DB
	mu       sync.RWMutex
	cached   Settings
}

func NewService(database *db.DB) *Service {
	return &Service{
		database: database,
		cached: Settings{
			MaxConcurrentJobs: 4,
			MaxJobRetries:     3,
			KindLimits:        defaultKindLimits,
		},
	}
}

// Init seeds missing global settings with their defaults, so settings added
// after the first start pick up their intended default instead of a zero
// value, then loads the cache.
func (s *Service) Init(ctx context.Context) error {
	err := s.database.Tx(ctx, func(tx *db.DB) error {
		for key, value := range defaultValues {
			if err := insertDefault(ctx, tx, key, value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	_, err = s.Get(ctx)
	return err
}

// PluginEnv returns the settings for pluginName as "KEY=value" env var strings.
// Implements plugins.PluginSettingsProvider.
func (s *Service) PluginEnv(pluginName string) []string {
	cfg := s.Cached()
	vals := cfg.PluginSettings[pluginName]
	env := make([]string, 0, len(vals))
	for k, v := range vals {
		env = append(env, k+"="+v)
	}
	return env
}

// load reads every setting row and decodes them into Settings.
func (s *Service) load(ctx context.Context) (*Settings, error) {
	rows, err := s.database.Query(ctx, `SELECT scope, plugin, key, value, updated_at FROM setting`, nil)
	if err != nil {
		return nil, fmt.Errorf("load settings: %w", err)
	}
	fields := map[string]any{}
	pluginSettings := map[string]map[string]string{}
	latestUpdate := ""
	for _, row := range rows {
		updatedAt, _ := row["updated_at"].(string)
		if updatedAt > latestUpdate {
			latestUpdate = updatedAt
		}
		key, _ := row["key"].(string)
		if row["scope"] == scopePlugin {
			addPluginSetting(pluginSettings, row, key)
			continue
		}
		fields[key] = row["value"]
	}
	fields[pluginSettingsKey] = pluginSettings
	if latestUpdate != "" {
		fields[updatedAtKey] = latestUpdate
	}

	var cfg Settings
	if err := db.DecodeRow(fields, &cfg); err != nil {
		return nil, fmt.Errorf("decode settings: %w", err)
	}
	return &cfg, nil
}

// addPluginSetting files a plugin-scoped row under its plugin name.
func addPluginSetting(pluginSettings map[string]map[string]string, row db.Row, key string) {
	pluginID, _ := row["plugin"].(string)
	pluginName := strings.TrimPrefix(pluginID, "plugin:")
	if pluginSettings[pluginName] == nil {
		pluginSettings[pluginName] = map[string]string{}
	}
	pluginSettings[pluginName][key] = settingString(row["value"])
}

// settingString renders a plugin setting value as the string injected into
// the plugin's environment.
func settingString(value any) string {
	if text, isText := value.(string); isText {
		return text
	}
	if value == nil {
		return ""
	}
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

// Get returns fresh settings from DB.
func (s *Service) Get(ctx context.Context) (*Settings, error) {
	cfg, err := s.load(ctx)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.cached = *cfg
	s.mu.Unlock()
	return cfg, nil
}

// Cached returns the last loaded settings without a DB round-trip.
func (s *Service) Cached() Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cached
}

// Update merges the provided fields and returns the updated settings.
// "pluginSettings" (pluginName → key → value) upserts plugin-scoped rows;
// every other key upserts one global row.
func (s *Service) Update(ctx context.Context, updates map[string]any) (*Settings, error) {
	err := s.database.Tx(ctx, func(tx *db.DB) error {
		for key, value := range updates {
			if err := applyUpdate(ctx, tx, key, value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx)
}

// applyUpdate writes one Update entry.
func applyUpdate(ctx context.Context, tx *db.DB, key string, value any) error {
	switch key {
	case updatedAtKey:
		return nil
	case pluginSettingsKey:
		return upsertPluginSettings(ctx, tx, value)
	}
	return upsertSetting(ctx, tx, scopeGlobal, "", key, value)
}

// upsertPluginSettings writes each plugin's key/value pairs as plugin-scoped
// rows. value is any JSON-shaped pluginName → key → value map.
func upsertPluginSettings(ctx context.Context, tx *db.DB, value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode pluginSettings: %w", err)
	}
	var pluginSettings map[string]map[string]string
	if err := json.Unmarshal(encoded, &pluginSettings); err != nil {
		return fmt.Errorf("decode pluginSettings: %w", err)
	}
	for pluginName, values := range pluginSettings {
		for key, settingValue := range values {
			if err := upsertSetting(ctx, tx, scopePlugin, pluginName, key, settingValue); err != nil {
				return err
			}
		}
	}
	return nil
}

// upsertSetting inserts or replaces the value of (scope, plugin, key).
func upsertSetting(ctx context.Context, tx *db.DB, scope, pluginName, key string, value any) error {
	vars, err := settingVars(scope, pluginName, key, value)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO setting (id, scope, plugin, key, value) VALUES ($id, $scope, $plugin, $key, $value)
		ON CONFLICT (scope, ifnull(plugin, ''), key) DO UPDATE SET value = excluded.value`, vars)
	if err != nil {
		return fmt.Errorf("upsert setting %s: %w", key, err)
	}
	return nil
}

// insertDefault writes a global setting only when it doesn't exist yet.
func insertDefault(ctx context.Context, tx *db.DB, key string, value any) error {
	vars, err := settingVars(scopeGlobal, "", key, value)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO setting (id, scope, plugin, key, value) VALUES ($id, $scope, $plugin, $key, $value)
		ON CONFLICT (scope, ifnull(plugin, ''), key) DO NOTHING`, vars)
	if err != nil {
		return fmt.Errorf("seed setting %s: %w", key, err)
	}
	return nil
}

// settingVars binds a setting row. The value is stored as JSON text so any Go
// value round-trips; the plugin column holds "plugin:<name>" or NULL.
func settingVars(scope, pluginName, key string, value any) (db.Vars, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode setting %s: %w", key, err)
	}
	vars := db.Vars{
		"id":     db.NewRecordID("setting"),
		"scope":  scope,
		"plugin": nil,
		"key":    key,
		"value":  string(encoded),
	}
	if pluginName != "" {
		vars["plugin"] = db.RecordID{Table: "plugin", ID: pluginName}
	}
	return vars, nil
}
