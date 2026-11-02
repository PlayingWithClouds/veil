package automations

import (
	"context"
	"slices"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/schema"
)

// filterableKinds are the kinds a trigger filter or step condition can
// test; records of other kinds pass them untested.
var filterableKinds = []schema.RecordKind{schema.RecordKindVideo, schema.RecordKindGallery}

// firing is one occurrence of a trigger.
type firing struct {
	kind schema.TriggerKind
	// qualifier is the event kind of an event trigger, the
	// "<plugin>:<trigger>" of a plugin trigger.
	qualifier string
	data      any
	subjects  []schema.RecordRef
}

// Fire queues a run of every enabled automation with a trigger of kind
// whose filter the subjects match, and returns the queued job ids. A
// trigger filter narrows the subjects to the matching videos and galleries;
// when none match, the automation doesn't run. Plugin triggers fire through
// FirePlugin, events through HandleEvent.
func (s *Service) Fire(ctx context.Context, kind schema.TriggerKind, triggerData any, subjects []schema.RecordRef) ([]string, error) {
	return s.fire(ctx, firing{kind: kind, data: triggerData, subjects: subjects})
}

// FirePlugin fires the automations triggered by a plugin's trigger
// ("<plugin>:<trigger>").
func (s *Service) FirePlugin(ctx context.Context, pluginTrigger string, triggerData any, subjects []schema.RecordRef) ([]string, error) {
	return s.fire(ctx, firing{kind: schema.TriggerKindPlugin, qualifier: pluginTrigger, data: triggerData, subjects: subjects})
}

// HandleEvent fires the event automations that listen for the event's kind
// (or for every kind, when they name none), with its subject as input.
func (s *Service) HandleEvent(ctx context.Context, event schema.Event) error {
	subjects := []schema.RecordRef{}
	if event.SubjectKind != nil && event.SubjectId != nil {
		subjects = append(subjects, schema.RecordRef{Kind: *event.SubjectKind, Id: *event.SubjectId})
	}
	data := map[string]any{"eventId": event.Id, "eventKind": event.Kind, "data": event.Data}
	_, err := s.fire(ctx, firing{kind: schema.TriggerKindEvent, qualifier: string(event.Kind), data: data, subjects: subjects})
	return err
}

// fire queues the runs of one trigger occurrence.
func (s *Service) fire(ctx context.Context, occurrence firing) ([]string, error) {
	automations, err := s.enabledWithTrigger(ctx, occurrence.kind)
	if err != nil {
		return nil, err
	}
	jobIDs := []string{}
	for _, automation := range automations {
		subjects, matches, err := s.triggerSubjects(ctx, automation.Trigger, occurrence)
		if err != nil {
			return jobIDs, err
		}
		if !matches {
			continue
		}
		jobID, err := s.Enqueue(ctx, Request{AutomationID: automation.Id, Subjects: subjects, TriggerData: occurrence.data})
		if err != nil {
			return jobIDs, err
		}
		jobIDs = append(jobIDs, jobID)
	}
	return jobIDs, nil
}

// triggerSubjects decides whether an occurrence fires a trigger and with
// which subjects.
func (s *Service) triggerSubjects(ctx context.Context, trigger schema.Trigger, occurrence firing) ([]schema.RecordRef, bool, error) {
	if !qualifies(trigger, occurrence) {
		return nil, false, nil
	}
	if trigger.Filter == nil || len(occurrence.subjects) == 0 {
		return occurrence.subjects, true, nil
	}
	subjects, err := s.matching(ctx, *trigger.Filter, occurrence.subjects)
	return subjects, len(subjects) > 0, err
}

// qualifies checks the kind-specific part of a trigger: the event kinds it
// listens for, the plugin trigger it names.
func qualifies(trigger schema.Trigger, occurrence firing) bool {
	switch trigger.Kind {
	case schema.TriggerKindEvent:
		return len(trigger.EventKinds) == 0 || slices.Contains(trigger.EventKinds, occurrence.qualifier)
	case schema.TriggerKindPlugin:
		return trigger.PluginTrigger != nil && *trigger.PluginTrigger == occurrence.qualifier
	}
	return true
}

// enabledWithTrigger lists the enabled automations with a trigger of kind.
func (s *Service) enabledWithTrigger(ctx context.Context, kind schema.TriggerKind) ([]schema.Automation, error) {
	return db.QueryAs[schema.Automation](ctx, s.database,
		`SELECT * FROM "Automation" WHERE "deletedAt" IS NULL AND "enabled" = 1
		 AND json_extract("trigger", '$.kind') = $kind ORDER BY "createdAt", rowid`,
		db.Vars{"kind": kind})
}

// matching keeps the refs that match filter: videos and galleries are
// tested against it, other kinds pass.
func (s *Service) matching(ctx context.Context, filter schema.Filter, refs []schema.RecordRef) ([]schema.RecordRef, error) {
	matched := map[schema.RecordRef]bool{}
	for _, kind := range filterableKinds {
		ids, err := s.library.MatchingIDs(ctx, kind, filter, idsOfKind(refs, kind))
		if err != nil {
			return nil, err
		}
		markMatched(matched, kind, ids)
	}
	kept := []schema.RecordRef{}
	for _, ref := range refs {
		if matched[ref] || !slices.Contains(filterableKinds, ref.Kind) {
			kept = append(kept, ref)
		}
	}
	return kept, nil
}

// markMatched records the ids of kind as matched.
func markMatched(matched map[schema.RecordRef]bool, kind schema.RecordKind, ids []string) {
	for _, id := range ids {
		matched[schema.RecordRef{Kind: kind, Id: id}] = true
	}
}

// idsOfKind picks the ids of refs of kind.
func idsOfKind(refs []schema.RecordRef, kind schema.RecordKind) []string {
	ids := []string{}
	for _, ref := range refs {
		if ref.Kind == kind {
			ids = append(ids, ref.Id)
		}
	}
	return ids
}

// Request asks for one run of an automation.
type Request struct {
	AutomationID string             `json:"automationId"`
	Subjects     []schema.RecordRef `json:"subjects"`
	TriggerData  any                `json:"triggerData,omitempty"`
	DryRun       bool               `json:"dryRun"`
}

// Enqueue queues an automationRun job for request and returns its id.
func (s *Service) Enqueue(ctx context.Context, request Request) (string, error) {
	subjects := request.Subjects
	if subjects == nil {
		subjects = []schema.RecordRef{}
	}
	return s.queue.Enqueue(ctx, jobs.Job{
		Kind: schema.JobKindAutomationRun,
		Payload: map[string]any{
			"automationId": request.AutomationID,
			"subjects":     subjects,
			"triggerData":  request.TriggerData,
			"dryRun":       request.DryRun,
		},
	})
}
