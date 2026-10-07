package settings

import (
	"context"
	"path/filepath"
	"slices"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
)

func openTestService(t *testing.T) (*Service, *db.DB, context.Context) {
	t.Helper()
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	service := NewService(database)
	if err := service.Init(ctx); err != nil {
		t.Fatalf("init: %v", err)
	}
	return service, database, ctx
}

func TestInitSeedsDefaults(t *testing.T) {
	service, _, _ := openTestService(t)
	cfg := service.Cached()
	if cfg.MaxConcurrentJobs != 4 || cfg.MaxJobRetries != 3 {
		t.Fatalf("job defaults: %+v", cfg)
	}
	if !cfg.AllowDownloadsWhileStreaming || cfg.RequireVpn {
		t.Fatalf("bool defaults: %+v", cfg)
	}
	if cfg.KindLimits["download"].RetryMultiplier != 1.5 {
		t.Fatalf("kind limits: %+v", cfg.KindLimits)
	}
	if cfg.UpdatedAt.IsZero() {
		t.Fatalf("updatedAt not set")
	}
}

func TestUpdateRoundTripAndInitKeepsValues(t *testing.T) {
	service, database, ctx := openTestService(t)
	limits := map[string]KindRateLimit{"scrape": {MaxConcurrent: 9, RetryInitialMs: 10, RetryMultiplier: 3, RetryMaxMs: 99}}
	updated, err := service.Update(ctx, map[string]any{
		"maxConcurrentJobs":            7,
		"allowDownloadsWhileStreaming": false,
		"disabledPlugins":              []string{"a", "b"},
		"kindLimits":                   limits,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.MaxConcurrentJobs != 7 || updated.AllowDownloadsWhileStreaming {
		t.Fatalf("updated scalars: %+v", updated)
	}
	if !slices.Equal(updated.DisabledPlugins, []string{"a", "b"}) {
		t.Fatalf("disabled plugins: %v", updated.DisabledPlugins)
	}
	if updated.KindLimits["scrape"] != limits["scrape"] {
		t.Fatalf("kind limits: %+v", updated.KindLimits)
	}

	// A restart must not reset stored values back to defaults.
	restarted := NewService(database)
	if err := restarted.Init(ctx); err != nil {
		t.Fatalf("re-init: %v", err)
	}
	if cfg := restarted.Cached(); cfg.MaxConcurrentJobs != 7 || cfg.AllowDownloadsWhileStreaming {
		t.Fatalf("values after re-init: %+v", cfg)
	}
	globalRows, err := database.Int(ctx, `SELECT count(*) FROM setting WHERE scope = 'global' AND key = 'maxConcurrentJobs'`, nil)
	if err != nil || globalRows != 1 {
		t.Fatalf("maxConcurrentJobs rows = %d (err %v), want 1", globalRows, err)
	}
}

func TestPluginSettingsAreScopedPerPlugin(t *testing.T) {
	service, database, ctx := openTestService(t)
	_, err := service.Update(ctx, map[string]any{"pluginSettings": map[string]map[string]string{
		"alpha": {"API_KEY": "one", "REGION": "eu"},
		"beta":  {"API_KEY": "two"},
	}})
	if err != nil {
		t.Fatalf("update plugin settings: %v", err)
	}
	// Overwrite one plugin's key; the other plugin's same key must not change.
	updated, err := service.Update(ctx, map[string]any{"pluginSettings": map[string]map[string]string{
		"alpha": {"API_KEY": "three", "REGION": "eu"},
	}})
	if err != nil {
		t.Fatalf("second update: %v", err)
	}
	if updated.PluginSettings["alpha"]["API_KEY"] != "three" || updated.PluginSettings["alpha"]["REGION"] != "eu" {
		t.Fatalf("alpha settings: %+v", updated.PluginSettings["alpha"])
	}
	if updated.PluginSettings["beta"]["API_KEY"] != "two" {
		t.Fatalf("beta settings: %+v", updated.PluginSettings["beta"])
	}

	env := service.PluginEnv("beta")
	if !slices.Equal(env, []string{"API_KEY=two"}) {
		t.Fatalf("plugin env: %v", env)
	}

	pluginRows, err := database.Int(ctx, `SELECT count(*) FROM setting WHERE scope = 'plugin' AND plugin = 'plugin:alpha'`, nil)
	if err != nil || pluginRows != 2 {
		t.Fatalf("alpha rows = %d (err %v), want 2", pluginRows, err)
	}
	// A plugin key must not collide with a global key of the same name.
	if _, err := service.Update(ctx, map[string]any{"API_KEY": "global"}); err != nil {
		t.Fatalf("global key update: %v", err)
	}
	cfg, err := service.Get(ctx)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if cfg.PluginSettings["alpha"]["API_KEY"] != "three" {
		t.Fatalf("global key overwrote plugin key: %+v", cfg.PluginSettings)
	}
}
