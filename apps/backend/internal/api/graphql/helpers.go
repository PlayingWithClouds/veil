package graphql

import (
	"context"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// pointers turns a slice of values into the slice of pointers gqlgen wants.
func pointers[T any](values []T) []*T {
	result := make([]*T, len(values))
	for index := range values {
		result[index] = &values[index]
	}
	return result
}

// text returns the string a pointer holds, or "".
func text(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// number returns the int a pointer holds, or 0.
func number(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

// optionalText returns nil for "", else a pointer to the string.
func optionalText(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// mediaFilter maps the API filter and page to the library's.
func mediaFilter(filter *model.MediaFilter, limit, offset *int) library.MediaFilter {
	result := library.MediaFilter{Limit: number(limit), Offset: number(offset)}
	if filter == nil {
		return result
	}
	result.Text = text(filter.Text)
	result.PerformerID = text(filter.PerformerID)
	result.StudioID = text(filter.StudioID)
	result.TagID = text(filter.TagID)
	return result
}

// nameFilter maps a name search and page to the library's.
func nameFilter(search *string, limit, offset *int) library.NameFilter {
	return library.NameFilter{Text: text(search), Limit: number(limit), Offset: number(offset)}
}

// listEntity maps an API list kind to the plugin entity it lists.
func listEntity(kind model.ListKind) plugins.MediaType {
	switch kind {
	case model.ListKindGallery:
		return plugins.MediaTypeGallery
	case model.ListKindPerformer:
		return plugins.MediaTypePerformer
	case model.ListKindStudio:
		return plugins.MediaTypeStudio
	case model.ListKindTag:
		return plugins.MediaTypeTag
	}
	return plugins.MediaTypeScene
}

// pageEntity picks what a plugin's pages describe: videos first, then
// galleries, performers and studios.
func pageEntity(plugin *plugins.Plugin) (plugins.MediaType, bool) {
	for _, entity := range []plugins.MediaType{plugins.MediaTypeScene, plugins.MediaTypeGallery, plugins.MediaTypePerformer, plugins.MediaTypeStudio} {
		if plugin.Meta.Has(plugins.FindCapability(entity)) {
			return entity, true
		}
	}
	return "", false
}

// listed loads stored records by reference, grouped by kind, in the order the
// references came.
func (r *Resolver) listed(ctx context.Context, refs []schema.RecordRef) (*model.Listed, error) {
	result := &model.Listed{
		Videos: []*schema.Video{}, Galleries: []*schema.Gallery{}, Performers: []*schema.Performer{},
		Studios: []*schema.Studio{}, Tags: []*schema.Tag{},
	}
	for _, ref := range refs {
		if err := r.addListed(ctx, result, ref); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// addListed loads one referenced record into its group; missing ones are skipped.
func (r *Resolver) addListed(ctx context.Context, result *model.Listed, ref schema.RecordRef) error {
	switch ref.Kind {
	case schema.RecordKindVideo:
		return appendLoaded(&result.Videos, r.library.Video, ctx, ref.Id)
	case schema.RecordKindGallery:
		return appendLoaded(&result.Galleries, r.library.Gallery, ctx, ref.Id)
	case schema.RecordKindPerformer:
		return appendLoaded(&result.Performers, r.library.Performer, ctx, ref.Id)
	case schema.RecordKindStudio:
		return appendLoaded(&result.Studios, r.library.Studio, ctx, ref.Id)
	case schema.RecordKindTag:
		return appendLoaded(&result.Tags, r.library.Tag, ctx, ref.Id)
	}
	return nil
}

// appendLoaded loads a record with load and appends it when it exists.
func appendLoaded[T any](list *[]*T, load func(context.Context, string) (*T, error), ctx context.Context, id string) error {
	record, err := load(ctx, id)
	if err != nil || record == nil {
		return err
	}
	*list = append(*list, record)
	return nil
}

// loadOptional loads a referenced record, or nil when the reference is empty.
func loadOptional[T any](ctx context.Context, load func(context.Context, string) (*T, error), id *string) (*T, error) {
	if id == nil || *id == "" {
		return nil, nil
	}
	return load(ctx, *id)
}
