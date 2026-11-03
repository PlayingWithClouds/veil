package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/blobcache"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/pipeline"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/user"
)

// newEventOf turns an event input into the event to log; its data must be
// an object.
func newEventOf(event model.EventInput) (user.NewEvent, error) {
	data, isObject := event.Data.(map[string]any)
	if event.Data != nil && !isObject {
		return user.NewEvent{}, fmt.Errorf("an event's data is an object")
	}
	recorded := user.NewEvent{Kind: event.Kind, Data: data, SessionID: event.SessionID}
	if event.Subject != nil {
		recorded.SubjectKind = &event.Subject.Kind
		recorded.SubjectID = &event.Subject.Id
	}
	return recorded, nil
}

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

// defaultCreditedLimit is how many credited performers a row shows unless asked otherwise.
const defaultCreditedLimit = 20

// defaultSimilarLimit is how many similar records a page shows unless asked otherwise.
const defaultSimilarLimit = 12

// similarLimit is the asked-for number of similar records, else the default.
func similarLimit(limit *int) int {
	if limit == nil || *limit <= 0 {
		return defaultSimilarLimit
	}
	return *limit
}

// optionalText returns nil for "", else a pointer to the string.
func optionalText(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

// defaultPage and maxPage bound a page of results the API returns.
const (
	defaultPage = 40
	maxPage     = 200
)

// page clamps a requested page to the allowed size.
func page(limit, offset *int) (int, int) {
	size := number(limit)
	if size <= 0 {
		size = defaultPage
	}
	if size > maxPage {
		size = maxPage
	}
	start := number(offset)
	if start < 0 {
		start = 0
	}
	return size, start
}

// mediaQuery maps the API filter and page to the library's.
func mediaQuery(filter *schema.Filter, limit, offset *int) library.MediaQuery {
	query := library.MediaQuery{Limit: number(limit), Offset: number(offset)}
	if filter != nil {
		query.Filter = *filter
	}
	return query
}

// tagCounts maps the library's tag counts to the API's.
func tagCounts(counts []library.TagCount) []*model.TagCount {
	result := make([]*model.TagCount, 0, len(counts))
	for index := range counts {
		result = append(result, &model.TagCount{Tag: &counts[index].Tag, Count: counts[index].Count})
	}
	return result
}

// filterOrEmpty returns the filter, or one that matches everything.
func filterOrEmpty(filter *schema.Filter) schema.Filter {
	if filter == nil {
		return schema.Filter{}
	}
	return *filter
}

// mediaEdit maps the API's media edit to the library's.
func mediaEdit(edit model.MediaEdit) library.MediaEdit {
	return library.MediaEdit{
		Title:        edit.Title,
		Details:      edit.Details,
		Date:         edit.Date,
		StudioID:     edit.StudioID,
		PerformerIDs: optionalList(edit.PerformerIds),
		TagIDs:       optionalList(edit.TagIds),
	}
}

// optionalList returns nil for an omitted list, else a pointer to it.
func optionalList(values []string) *[]string {
	if values == nil {
		return nil
	}
	return &values
}

// userStateOf loads the user's state on a record, or the defaults when the
// user never touched it.
func (r *Resolver) userStateOf(ctx context.Context, kind schema.RecordKind, id string) (*schema.UserState, error) {
	state, err := r.user.State(ctx, kind, id)
	if err != nil || state != nil {
		return state, err
	}
	return &schema.UserState{SubjectKind: kind, SubjectId: id, Fields: map[string]any{}}, nil
}

// oCount counts the o-events logged on a record.
func (r *Resolver) oCount(ctx context.Context, kind schema.RecordKind, id string) (int, error) {
	return r.user.CountEvents(ctx, user.EventQuery{Kinds: []schema.EventKind{schema.EventKindO}, SubjectKind: &kind, SubjectID: &id})
}

// fieldSourcesOf lists who owns each field of a record.
func (r *Resolver) fieldSourcesOf(ctx context.Context, kind schema.RecordKind, id string) ([]*schema.FieldSource, error) {
	owners, err := r.library.Recorder().FieldSources(ctx, kind, id)
	return pointers(owners), err
}

// revisionsOf lists a page of a record's revisions, newest first.
func (r *Resolver) revisionsOf(ctx context.Context, kind schema.RecordKind, id string, limit, offset *int) ([]*schema.Revision, error) {
	size, start := page(limit, offset)
	revisions, err := r.library.Recorder().Revisions(ctx, kind, id, size, start)
	return pointers(revisions), err
}

// loadRecord loads the record a reference names; nil when it's gone or of
// a kind the API doesn't serve yet.
func (r *Resolver) loadRecord(ctx context.Context, kind schema.RecordKind, id string) (model.Record, error) {
	switch kind {
	case schema.RecordKindVideo:
		return nilIfMissing(r.library.Video(ctx, id))
	case schema.RecordKindGallery:
		return nilIfMissing(r.library.Gallery(ctx, id))
	case schema.RecordKindPerformer:
		return nilIfMissing(r.library.Performer(ctx, id))
	case schema.RecordKindStudio:
		return nilIfMissing(r.library.Studio(ctx, id))
	case schema.RecordKindTag:
		return nilIfMissing(r.library.Tag(ctx, id))
	}
	return r.loadOtherRecord(ctx, kind, id)
}

// loadOtherRecord loads the records of the kinds beyond the core ones.
func (r *Resolver) loadOtherRecord(ctx context.Context, kind schema.RecordKind, id string) (model.Record, error) {
	switch kind {
	case schema.RecordKindAudio:
		return nilIfMissing(r.library.Audio(ctx, id))
	case schema.RecordKindStory:
		return nilIfMissing(r.library.Story(ctx, id))
	case schema.RecordKindPost:
		return nilIfMissing(r.library.Post(ctx, id))
	case schema.RecordKindGroup:
		return nilIfMissing(r.library.Group(ctx, id))
	case schema.RecordKindCategory:
		return nilIfMissing(r.library.Category(ctx, id))
	case schema.RecordKindCollection:
		return nilIfMissing(r.user.Collection(ctx, id))
	case schema.RecordKindCamModel:
		return nilIfMissing(db.Get[schema.CamModel](ctx, r.library.DB(), "CamModel", id))
	case schema.RecordKindSegment:
		return nilIfMissing(r.library.Segment(ctx, id))
	case schema.RecordKindClip:
		return nilIfMissing(r.library.Clip(ctx, id))
	}
	return nil, nil
}

// loadRecords loads records by reference in order, skipping missing ones.
func (r *Resolver) loadRecords(ctx context.Context, refs []schema.RecordRef) ([]model.Record, error) {
	records := make([]model.Record, 0, len(refs))
	for _, ref := range refs {
		record, err := r.loadRecord(ctx, ref.Kind, ref.Id)
		if err != nil {
			return nil, err
		}
		if record != nil {
			records = append(records, record)
		}
	}
	return records, nil
}

// refOf is the reference of a record of kind with id.
func refOf(kind schema.RecordKind, id string) schema.RecordRef {
	return schema.RecordRef{Kind: kind, Id: id}
}

// ok turns an error into the Boolean! answer of a mutation.
func ok(err error) (bool, error) {
	if err != nil {
		return false, err
	}
	return true, nil
}

// mediaEditOf maps the API's media edit to the library's, VR and size
// included.
func mediaEditOf(edit model.MediaEdit) library.MediaEdit {
	converted := mediaEdit(edit)
	converted.Status = edit.Status
	converted.VR = edit.Vr
	converted.ClearVR = edit.ClearVr != nil && *edit.ClearVr
	converted.Width = edit.Width
	converted.Height = edit.Height
	converted.DurationSeconds = edit.DurationSeconds
	return converted
}

// refsOf collects the references of a list of records.
func refsOf(subjects []*schema.RecordRef) []schema.RecordRef {
	refs := make([]schema.RecordRef, 0, len(subjects))
	for _, subject := range subjects {
		if subject != nil {
			refs = append(refs, *subject)
		}
	}
	return refs
}

// pointerTo returns a pointer to a copy of value.
func pointerTo[T any](value T) *T {
	return &value
}

// textOrNil returns nil for an empty string.
func textOrNil(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}
	return value
}

// nilIfMissing turns a typed nil record into an untyped nil, so a missing
// record resolves to null rather than an empty member of the union.
func nilIfMissing[T any](record *T, err error) (model.Record, error) {
	if err != nil || record == nil {
		return nil, err
	}
	return record, nil
}

// eventQuery maps the API's event filter and page to the user service's.
func eventQuery(filter *model.EventFilter, limit, offset *int) user.EventQuery {
	query := user.EventQuery{}
	query.Limit, query.Offset = page(limit, offset)
	if filter == nil {
		return query
	}
	query.Kinds = filter.Kinds
	query.Since = text(filter.Since)
	query.Until = text(filter.Until)
	if filter.Subject != nil {
		query.SubjectKind = &filter.Subject.Kind
		query.SubjectID = &filter.Subject.Id
	}
	return query
}

// timeText renders an optional time in the stored timestamp format.
func timeText(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := db.FormatTime(*value)
	return &formatted
}

// applyCacheQuota hands a changed quota of this device's cache location to
// the evictor, so it applies without a restart. No quota means the default.
func (r *Resolver) applyCacheQuota(location *schema.StorageLocation) {
	if r.evictor == nil || location.Kind != schema.StorageKindCache || location.Device != r.devices.Current().Id {
		return
	}
	if location.QuotaBytes == nil {
		r.evictor.SetMaxBytes(blobcache.DefaultMaxBytes)
		return
	}
	r.evictor.SetMaxBytes(*location.QuotaBytes)
}

// quota converts an API byte count to the storage service's.
func quota(bytes *int) *int64 {
	if bytes == nil {
		return nil
	}
	converted := int64(*bytes)
	return &converted
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
	case model.ListKindAudio:
		return plugins.MediaTypeAudio
	case model.ListKindStory:
		return plugins.MediaTypeStory
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
		Studios: []*schema.Studio{}, Tags: []*schema.Tag{}, Audios: []*schema.Audio{}, Stories: []*schema.Story{},
	}
	for _, ref := range refs {
		if err := r.addListed(ctx, result, ref); err != nil {
			return nil, err
		}
	}
	return result, nil
}

// defaultSearchLimit is how many results a search page asks each site for.
const defaultSearchLimit = 20

// searchBatchStream turns a search's batches into the subscription's, with
// their records loaded, until the search ends or the subscriber leaves.
func (r *Resolver) searchBatchStream(ctx context.Context, batches <-chan pipeline.SearchBatch, sites int) <-chan *model.SearchBatch {
	forwarded := make(chan *model.SearchBatch)
	go func() {
		defer close(forwarded)
		for batch := range batches {
			listed, err := r.listed(ctx, batch.Refs)
			if err != nil {
				return
			}
			select {
			case forwarded <- &model.SearchBatch{Plugin: batch.Plugin, Listed: listed, Sites: sites, Stored: batch.Stored, Final: batch.Final, Exhausted: batch.Exhausted, Returned: batch.Returned}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return forwarded
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
	case schema.RecordKindAudio:
		return appendLoaded(&result.Audios, r.library.Audio, ctx, ref.Id)
	case schema.RecordKindStory:
		return appendLoaded(&result.Stories, r.library.Story, ctx, ref.Id)
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
