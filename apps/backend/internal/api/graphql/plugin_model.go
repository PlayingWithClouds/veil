package graphql

import (
	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// pluginToModel describes a loaded plugin for the API.
func pluginToModel(plugin *plugins.Plugin, enabled, available bool, values map[string]string) *model.Plugin {
	capabilities := make([]string, 0, len(plugin.Meta.Capabilities))
	for _, capability := range plugin.Meta.Capabilities {
		capabilities = append(capabilities, string(capability))
	}
	domains := plugin.Meta.Domains
	if domains == nil {
		domains = []string{}
	}
	return &model.Plugin{
		Name:           plugin.Meta.Name,
		PackageName:    optionalText(plugin.Package),
		LocalBuild:     plugin.LocalBuild,
		DisplayName:    optionalText(plugin.Meta.DisplayName),
		IconURL:        optionalText(plugin.Meta.Icon),
		Description:    optionalText(plugin.Meta.Description),
		Version:        plugin.Meta.Version,
		Capabilities:   capabilities,
		Domains:        domains,
		Content:        nonNil(plugin.ContentKinds()),
		Enabled:        enabled,
		RequiresSolver: plugin.Meta.RequiresSolver,
		Available:      available,
		Settings:       settingFields(plugin.Meta.Settings),
		SettingValues:  settingValues(values),
	}
}

// settingFields describes the settings a plugin asks for.
func settingFields(fields []plugins.PluginSettingField) []*model.PluginSettingField {
	result := make([]*model.PluginSettingField, 0, len(fields))
	for _, field := range fields {
		result = append(result, &model.PluginSettingField{
			Key:         field.Key,
			Label:       field.Label,
			Description: optionalText(field.Description),
			Type:        field.Type,
			Required:    field.Required,
			Default:     optionalText(field.Default),
		})
	}
	return result
}

// settingValues lists a plugin's stored setting values.
func settingValues(values map[string]string) []*model.PluginSettingValue {
	result := make([]*model.PluginSettingValue, 0, len(values))
	for key, value := range values {
		result = append(result, &model.PluginSettingValue{Key: key, Value: value})
	}
	return result
}

// nonNil returns an empty list instead of nil.
func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
