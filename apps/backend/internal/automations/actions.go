package automations

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/notify"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Built-in action names. Main registers the ones that need other packages
// under the remaining names the schema lists: download, record, archive,
// lookup, moveFile, runPlugin (and AddToCollectionAction as addToCollection).
const (
	ActionTag             = "tag"
	ActionDelete          = "delete"
	ActionNotify          = "notify"
	ActionAddToCollection = "addToCollection"
	ActionHunt            = "hunt"
)

// RunContext is what an action knows about the run it is part of.
type RunContext struct {
	RunID        string
	AutomationID string
	Name         string
	// DryRun asks the action to report what it would do and write nothing.
	DryRun bool
	// Author is who the action writes as; its revisions carry the run id.
	Author      provenance.Author
	TriggerData any
}

// Result is what one step did.
type Result struct {
	// Output is the next step's input.
	Output []schema.RecordRef
	// Touched are the records the step changed (or would change, on a dry
	// run).
	Touched []schema.RecordRef
	// Report describes what was done, for the run's step results.
	Report string
}

// Action runs one step on its input with the step's args (raw JSON, nil
// when none).
type Action func(ctx context.Context, run RunContext, input []schema.RecordRef, args json.RawMessage) (Result, error)

// RegisterAction makes an action available to steps under name, replacing
// any action registered under it. Plugins use "<plugin>:<action>".
func (s *Service) RegisterAction(name string, action Action) {
	s.actionsMutex.Lock()
	defer s.actionsMutex.Unlock()
	s.actions[name] = action
}

// Actions lists the registered action names, sorted.
func (s *Service) Actions() []string {
	s.actionsMutex.RLock()
	defer s.actionsMutex.RUnlock()
	names := make([]string, 0, len(s.actions))
	for name := range s.actions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// action looks up a registered action.
func (s *Service) action(name string) (Action, bool) {
	s.actionsMutex.RLock()
	defer s.actionsMutex.RUnlock()
	action, known := s.actions[name]
	return action, known
}

// decodeArgs reads a step's args into target; no args leave it as it is.
func decodeArgs(args json.RawMessage, target any) error {
	if len(args) == 0 || string(args) == "null" {
		return nil
	}
	if err := json.Unmarshal(args, target); err != nil {
		return fmt.Errorf("invalid args: %w", err)
	}
	return nil
}

// tagArgs are the args of the tag action.
type tagArgs struct {
	Add    []string `json:"add"`
	Remove []string `json:"remove"`
}

// taggedKinds are the kinds with a tags list.
var taggedKinds = []schema.RecordKind{schema.RecordKindVideo, schema.RecordKindGallery, schema.RecordKindPerformer, schema.RecordKindStudio}

// tagAction adds and removes tag ids on each input record that has tags
// ({"add": [tagId], "remove": [tagId]}). Output: the input.
func (s *Service) tagAction(ctx context.Context, run RunContext, input []schema.RecordRef, args json.RawMessage) (Result, error) {
	var tags tagArgs
	if err := decodeArgs(args, &tags); err != nil {
		return Result{}, err
	}
	result := Result{Output: input, Touched: []schema.RecordRef{}}
	for _, ref := range input {
		changed, err := s.retag(ctx, run, ref, tags)
		if err != nil {
			return result, err
		}
		if changed {
			result.Touched = append(result.Touched, ref)
		}
	}
	result.Report = fmt.Sprintf("retagged %d of %d", len(result.Touched), len(input))
	return result, nil
}

// retag applies tag changes to one record and reports whether its tags
// changed.
func (s *Service) retag(ctx context.Context, run RunContext, ref schema.RecordRef, tags tagArgs) (bool, error) {
	if !slices.Contains(taggedKinds, ref.Kind) {
		return false, nil
	}
	current, err := s.tagsOf(ctx, ref)
	if err != nil || current == nil {
		return false, err
	}
	updated := changedTags(current, tags)
	if slices.Equal(current, updated) {
		return false, nil
	}
	if run.DryRun {
		return true, nil
	}
	changed, err := s.library.Recorder().Update(ctx, ref.Kind, ref.Id, db.Vars{"tags": updated}, run.Author)
	return len(changed) > 0, err
}

// tagsOf loads a record's tag ids; nil when the record is gone.
func (s *Service) tagsOf(ctx context.Context, ref schema.RecordRef) ([]string, error) {
	type tagged struct {
		Tags []string `json:"tags"`
	}
	record, err := db.QueryOneAs[tagged](ctx, s.database,
		`SELECT "tags" FROM `+db.Quote(provenance.Table(ref.Kind))+` WHERE "id" = $id AND "deletedAt" IS NULL`,
		db.Vars{"id": ref.Id})
	if err != nil || record == nil {
		return nil, err
	}
	if record.Tags == nil {
		return []string{}, nil
	}
	return record.Tags, nil
}

// changedTags is current with tags.Remove taken out and tags.Add appended.
func changedTags(current []string, tags tagArgs) []string {
	updated := []string{}
	for _, id := range current {
		if !slices.Contains(tags.Remove, id) {
			updated = append(updated, id)
		}
	}
	for _, id := range tags.Add {
		if !slices.Contains(updated, id) && !slices.Contains(tags.Remove, id) {
			updated = append(updated, id)
		}
	}
	return updated
}

// deleteAction moves each trashable input record to the trash (no args).
// Output: nothing, since the records are gone.
func (s *Service) deleteAction(ctx context.Context, run RunContext, input []schema.RecordRef, args json.RawMessage) (Result, error) {
	result := Result{Output: []schema.RecordRef{}, Touched: []schema.RecordRef{}}
	for _, ref := range input {
		if !slices.Contains(library.TrashableKinds, ref.Kind) {
			continue
		}
		result.Touched = append(result.Touched, ref)
		if run.DryRun {
			continue
		}
		if err := s.trash(ctx, ref, run.Author); err != nil {
			return result, err
		}
	}
	result.Report = fmt.Sprintf("trashed %d", len(result.Touched))
	return result, nil
}

// trash moves a record to the trash as author, so the run can restore it.
func (s *Service) trash(ctx context.Context, ref schema.RecordRef, author provenance.Author) error {
	_, err := s.library.Recorder().Update(ctx, ref.Kind, ref.Id, db.Vars{"deletedAt": db.Now()}, author)
	return err
}

// notifyArgs are the args of the notify action.
type notifyArgs struct {
	Kind  string `json:"kind"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// notifyAction sends one notification for the step ({"title", "body",
// "kind"}; title defaults to the automation's name, kind to "automation").
// The subject is the input record when there is exactly one. Output: the
// input.
func (s *Service) notifyAction(ctx context.Context, run RunContext, input []schema.RecordRef, args json.RawMessage) (Result, error) {
	notification := notifyArgs{Kind: "automation", Title: run.Name}
	if err := decodeArgs(args, &notification); err != nil {
		return Result{}, err
	}
	result := Result{Output: input, Touched: []schema.RecordRef{}, Report: "notified: " + notification.Title}
	if run.DryRun || s.notifier == nil {
		return result, nil
	}
	var subject *schema.RecordRef
	if len(input) == 1 {
		subject = &input[0]
	}
	_, err := s.notifier.Notify(ctx, notify.NewNotification{Kind: notification.Kind, Title: notification.Title, Body: notification.Body, Subject: subject})
	return result, err
}

// CollectionAdder adds records to a collection and returns the ones that
// weren't in it yet. Main supplies it from the collections package.
type CollectionAdder interface {
	AddToCollection(ctx context.Context, collectionID string, items []schema.RecordRef) ([]schema.RecordRef, error)
}

// collectionArgs are the args of the addToCollection action.
type collectionArgs struct {
	CollectionID string `json:"collectionId"`
}

// AddToCollectionAction is the addToCollection action over adder
// ({"collectionId": id}). Collection items aren't revisions: undoing a run
// leaves them. Output: the input.
func AddToCollectionAction(adder CollectionAdder) Action {
	return func(ctx context.Context, run RunContext, input []schema.RecordRef, args json.RawMessage) (Result, error) {
		var collection collectionArgs
		if err := decodeArgs(args, &collection); err != nil {
			return Result{}, err
		}
		if collection.CollectionID == "" {
			return Result{}, fmt.Errorf("addToCollection needs a collectionId")
		}
		if run.DryRun {
			return Result{Output: input, Touched: input, Report: fmt.Sprintf("would add %d", len(input))}, nil
		}
		added, err := adder.AddToCollection(ctx, collection.CollectionID, input)
		return Result{Output: input, Touched: added, Report: fmt.Sprintf("added %d", len(added))}, err
	}
}
