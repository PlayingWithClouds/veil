package graphql

import (
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/settings"
)

func strPtrIfSet(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// anyToStringPtr coerces a payload value to a *string.
func anyToStringPtr(v any) *string {
	if s, ok := v.(string); ok && s != "" {
		return &s
	}
	return nil
}

// anyToFloatPtr coerces a payload value (float64/int64/int) to a *float64.
func anyToFloatPtr(v any) *float64 {
	switch n := v.(type) {
	case float64:
		return &n
	case int64:
		f := float64(n)
		return &f
	case int:
		f := float64(n)
		return &f
	default:
		return nil
	}
}

// pluginToModel maps a loaded plugin plus its enabled/available state and
// stored setting values to the GraphQL model.
func pluginToModel(p *plugins.Plugin, enabled, available bool, values map[string]string) *model.Plugin {
	capabilities := make([]string, 0, len(p.Meta.Capabilities))
	for _, c := range p.Meta.Capabilities {
		capabilities = append(capabilities, string(c))
	}

	fields := make([]*model.PluginSettingField, 0, len(p.Meta.Settings))
	for _, f := range p.Meta.Settings {
		fields = append(fields, &model.PluginSettingField{
			Key:         f.Key,
			Label:       f.Label,
			Description: strPtrIfSet(f.Description),
			Type:        f.Type,
			Required:    f.Required,
			Default:     strPtrIfSet(f.Default),
		})
	}

	settingValues := make([]*model.PluginSettingValue, 0, len(values))
	for key, value := range values {
		settingValues = append(settingValues, &model.PluginSettingValue{Key: key, Value: value})
	}

	return &model.Plugin{
		ID:             p.Meta.Name,
		Name:           p.Meta.Name,
		PackageName:    strPtrIfSet(p.Package),
		LocalBuild:     p.LocalBuild,
		DisplayName:    strPtrIfSet(p.Meta.DisplayName),
		IconURL:        strPtrIfSet(p.Meta.Icon),
		Description:    strPtrIfSet(p.Meta.Description),
		Version:        p.Meta.Version,
		Capabilities:   capabilities,
		Domains:        p.Meta.Domains,
		Enabled:        enabled,
		RequiresSolver: p.Meta.RequiresSolver,
		Available:      available,
		Settings:       fields,
		SettingValues:  settingValues,
	}
}

func settingsToModel(s *settings.Settings) *model.Settings {
	kindLimits := make([]*model.KindRateLimit, 0, len(s.KindLimits))
	for kind, limit := range s.KindLimits {
		kindLimits = append(kindLimits, &model.KindRateLimit{
			Kind:            kind,
			MaxConcurrent:   limit.MaxConcurrent,
			RetryInitialMs:  limit.RetryInitialMs,
			RetryMultiplier: limit.RetryMultiplier,
			RetryMaxMs:      limit.RetryMaxMs,
		})
	}

	return &model.Settings{
		MaxConcurrentJobs:            s.MaxConcurrentJobs,
		MaxJobRetries:                s.MaxJobRetries,
		DownloadSpeedLimitKBps:       s.DownloadSpeedLimitKBps,
		AllowDownloadsWhileStreaming: s.AllowDownloadsWhileStreaming,
		AutoEnrichAfterScrape:        s.AutoEnrichAfterScrape,
		RequireVpn:                   s.RequireVpn,
		KindLimits:                   kindLimits,
	}
}

func jobToModel(j jobs.Job) *model.Job {
	m := &model.Job{
		ID:         j.StringID(),
		Kind:       string(j.Kind),
		Status:     string(j.Status),
		PluginName: j.PluginName,
		Attempts:   j.Attempts,
		Error:      strPtrIfSet(j.Error),
		CreatedAt:  j.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:  j.UpdatedAt.UTC().Format(time.RFC3339),
	}
	if j.RunAt != nil {
		runAt := j.RunAt.UTC().Format(time.RFC3339)
		m.RunAt = &runAt
	}
	if j.Payload != nil {
		m.Target = anyToStringPtr(j.Payload["media"])
		m.DownloadTitle = anyToStringPtr(j.Payload["title"])
		m.DownloadURL = anyToStringPtr(j.Payload["url"])
		m.DownloadProgress = anyToFloatPtr(j.Payload["progress"])
		m.DownloadBytesReceived = anyToFloatPtr(j.Payload["bytesReceived"])
		m.DownloadBytesTotal = anyToFloatPtr(j.Payload["bytesTotal"])
	}
	return m
}
