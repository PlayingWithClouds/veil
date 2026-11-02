package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/tokens"
)

// requireScope fails unless the request's token grants scope (requests
// without a token check are trusted).
func requireScope(ctx context.Context, scope string) error {
	if tokens.FromContext(ctx).Allows(scope) {
		return nil
	}
	return fmt.Errorf("this token lacks the %s scope", scope)
}

// optionalTime parses an optional RFC 3339 timestamp argument.
func optionalTime(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, *value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

// objectOf reads a JSON argument that must be an object.
func objectOf(value any) (map[string]any, error) {
	object, isObject := value.(map[string]any)
	if !isObject {
		return nil, fmt.Errorf("expected an object")
	}
	return object, nil
}

// asObject renders a value as a JSON object for a JSON field.
func asObject(value any) (any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	object := map[string]any{}
	return object, json.Unmarshal(encoded, &object)
}

// stringValues reads a JSON object of form values as text by key.
func stringValues(value any) (map[string]string, error) {
	if value == nil {
		return map[string]string{}, nil
	}
	object, err := objectOf(value)
	if err != nil {
		return nil, err
	}
	values := make(map[string]string, len(object))
	for key, entry := range object {
		values[key] = fmt.Sprint(entry)
	}
	return values, nil
}

// rawJSON encodes an optional JSON argument for a plugin; nil stays empty.
func rawJSON(value any) json.RawMessage {
	if value == nil {
		return nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return encoded
}

// contributionsOf lists a plugin's contributions, of one kind when given.
func contributionsOf(plugin *plugins.Plugin, kind *string) []*model.PluginContribution {
	contributions := []*model.PluginContribution{}
	for _, contribution := range plugin.Contributions() {
		if kind != nil && contribution.Kind != *kind {
			continue
		}
		contributions = append(contributions, &model.PluginContribution{
			Plugin:     plugin.Meta.Name,
			Kind:       contribution.Kind,
			Name:       contribution.Name,
			Title:      optionalText(contribution.Title),
			MediaKinds: nonNil(contribution.MediaKinds),
			Domains:    nonNil(contribution.Domains),
		})
	}
	return contributions
}

// accountOf returns this device's account of a plugin, nil when it has
// none.
func (r *Resolver) accountOf(ctx context.Context, pluginName string) (*schema.Account, error) {
	found, err := r.accounts.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	for index := range found {
		if found[index].Plugin == pluginName {
			return &found[index], nil
		}
	}
	return nil, nil
}

// viewPlugin finds an active plugin with declarative views.
func (r *Resolver) viewPlugin(name string) (*plugins.Plugin, error) {
	found, ok := r.registry.Get(name)
	if !ok || !r.registry.IsActive(name) {
		return nil, fmt.Errorf("plugin %q isn't active", name)
	}
	if !found.Serves(plugins.CapabilityViewRender) {
		return nil, fmt.Errorf("%s has no views", name)
	}
	return found, nil
}

// pluginViewOf maps a plugin's view for the API; rows that stand for a
// record have it stored as a stub of the plugin.
func (r *Resolver) pluginViewOf(ctx context.Context, pluginName string, view plugins.View) *model.PluginView {
	mapped := &model.PluginView{
		Kind: view.Kind, Title: view.Title, Body: optionalText(view.Body),
		Fields: []*model.PluginViewField{}, Items: []*model.PluginViewItem{}, Actions: []*model.PluginViewAction{},
	}
	for _, field := range view.Fields {
		mapped.Fields = append(mapped.Fields, &model.PluginViewField{
			Key: field.Key, Label: field.Label, Type: field.Type, Value: optionalText(field.Value),
			Options: nonNil(field.Options), Required: field.Required, Hint: optionalText(field.Hint),
		})
	}
	for _, item := range view.Items {
		mapped.Items = append(mapped.Items, r.viewItemOf(ctx, pluginName, item))
	}
	for _, action := range view.Actions {
		mapped.Actions = append(mapped.Actions, &model.PluginViewAction{Name: action.Name, Label: action.Label, Primary: action.Primary})
	}
	return mapped
}

// viewItemOf maps one row of a view.
func (r *Resolver) viewItemOf(ctx context.Context, pluginName string, item plugins.ViewItem) *model.PluginViewItem {
	mapped := &model.PluginViewItem{Title: item.Title, Subtitle: optionalText(item.Subtitle), ImageURL: optionalText(item.ImageURL), URL: optionalText(item.URL)}
	if item.Item == nil {
		return mapped
	}
	ref, err := r.library.IngestListed(ctx, pluginName, *item.Item)
	if err != nil || ref.Id == "" {
		return mapped
	}
	record, err := r.loadRecord(ctx, ref.Kind, ref.Id)
	if err == nil {
		mapped.Record = record
	}
	return mapped
}
