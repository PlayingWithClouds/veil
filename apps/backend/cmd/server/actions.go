package main

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/playingwithclouds/veil/internal/automations"
	"github.com/playingwithclouds/veil/internal/cams"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/lookup"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/user"
)

// Automation actions that need services beyond the automations package.
// Their args are {} unless noted.
const (
	actionDownload = "download"
	actionRecord   = "record"
	actionArchive  = "archive"
	actionLookup   = "lookup"
)

// registerActions makes the steps of automations able to download, record,
// archive and look up, and to add to collections.
func (f *features) registerActions(base core) {
	f.automations.RegisterAction(automations.ActionAddToCollection, automations.AddToCollectionAction(collectionAdder{user: base.user}))
	f.automations.RegisterAction(actionDownload, f.downloadAction)
	f.automations.RegisterAction(actionRecord, f.recordAction)
	f.automations.RegisterAction(actionArchive, f.archiveAction(base.library))
	f.automations.RegisterAction(actionLookup, f.lookupAction(base.database))
}

// collectionAdder adds automation output to a collection.
type collectionAdder struct {
	user *user.Service
}

// AddToCollection appends items to the collection and returns them.
func (adder collectionAdder) AddToCollection(ctx context.Context, collectionID string, items []schema.RecordRef) ([]schema.RecordRef, error) {
	if err := adder.user.AddItems(ctx, collectionID, items, nil); err != nil {
		return nil, err
	}
	return items, nil
}

// downloadAction downloads each video or audio of the input for keeps.
func (f *features) downloadAction(ctx context.Context, run automations.RunContext, input []schema.RecordRef, args json.RawMessage) (automations.Result, error) {
	media := ofKinds(input, schema.RecordKindVideo, schema.RecordKindAudio)
	if run.DryRun {
		return automations.Result{Output: input, Report: fmt.Sprintf("would download %d", len(media))}, nil
	}
	for _, ref := range media {
		if _, err := f.files.Download(ctx, ref, nil); err != nil {
			return automations.Result{}, err
		}
	}
	return automations.Result{Output: input, Touched: media, Report: fmt.Sprintf("queued %d downloads", len(media))}, nil
}

// recordAction starts recording each cam room of the input that's live.
func (f *features) recordAction(ctx context.Context, run automations.RunContext, input []schema.RecordRef, args json.RawMessage) (automations.Result, error) {
	rooms := ofKinds(input, schema.RecordKindCamModel)
	if run.DryRun {
		return automations.Result{Output: input, Report: fmt.Sprintf("would record %d rooms", len(rooms))}, nil
	}
	started := []schema.RecordRef{}
	for _, room := range rooms {
		if _, err := f.cams.StartRecording(ctx, room.Id, cams.RecordOptions{}); err != nil {
			continue
		}
		started = append(started, room)
	}
	return automations.Result{Output: input, Touched: started, Report: fmt.Sprintf("recording %d of %d rooms", len(started), len(rooms))}, nil
}

// archiveAction snapshots every source of the input's records.
func (f *features) archiveAction(lib *library.Library) automations.Action {
	return func(ctx context.Context, run automations.RunContext, input []schema.RecordRef, args json.RawMessage) (automations.Result, error) {
		archived := 0
		for _, ref := range input {
			sources, err := lib.Sources(ctx, ref.Kind, ref.Id)
			if err != nil {
				return automations.Result{}, err
			}
			for _, source := range sources {
				archived++
				if run.DryRun {
					continue
				}
				if _, err := f.files.Snapshot(ctx, source.Id); err != nil {
					return automations.Result{}, err
				}
			}
		}
		return automations.Result{Output: input, Report: fmt.Sprintf("%d source snapshots", archived)}, nil
	}
}

// lookupAction looks each media input up by its title and passes on what
// the identification plugins name.
func (f *features) lookupAction(database *db.DB) automations.Action {
	return func(ctx context.Context, run automations.RunContext, input []schema.RecordRef, args json.RawMessage) (automations.Result, error) {
		if run.DryRun {
			return automations.Result{Output: input, Report: fmt.Sprintf("would look up %d records", len(input))}, nil
		}
		found := []schema.RecordRef{}
		for _, ref := range input {
			title := titleOf(ctx, database, ref)
			if title == "" {
				continue
			}
			result, err := f.lookup.Lookup(ctx, lookup.Query{Text: title})
			if err != nil {
				return automations.Result{}, err
			}
			for _, match := range result.Matches {
				found = append(found, match.Item)
			}
		}
		return automations.Result{Output: found, Report: fmt.Sprintf("%d matches", len(found))}, nil
	}
}

// titleOf is a media record's title, "" for anything else.
func titleOf(ctx context.Context, database *db.DB, ref schema.RecordRef) string {
	if !library.IsMedia(ref.Kind) {
		return ""
	}
	titles, err := database.Strings(ctx, `SELECT "title" FROM `+db.Quote(provenance.Table(ref.Kind))+` WHERE "id" = $id`, db.Vars{"id": ref.Id})
	if err != nil || len(titles) == 0 {
		return ""
	}
	return titles[0]
}

// ofKinds keeps the records of the given kinds.
func ofKinds(refs []schema.RecordRef, kinds ...schema.RecordKind) []schema.RecordRef {
	kept := []schema.RecordRef{}
	for _, ref := range refs {
		if slices.Contains(kinds, ref.Kind) {
			kept = append(kept, ref)
		}
	}
	return kept
}
