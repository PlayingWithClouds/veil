package automations

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/notify"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// RunTask is the task of automationRun jobs. A failing step fails the run,
// not the job: retrying would repeat the steps that already worked.
func (s *Service) RunTask() jobs.Task {
	return func(ctx context.Context, job *jobs.Job, progress func(float64)) error {
		var request Request
		if err := remarshal(job.Payload, &request); err != nil {
			return fmt.Errorf("automationRun payload: %w", err)
		}
		_, err := s.Execute(ctx, request)
		return err
	}
}

// remarshal converts a decoded JSON value into target through JSON.
func remarshal(value any, target any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, target)
}

// stepOutput is what a step result's output holds.
type stepOutput struct {
	Records []schema.RecordRef `json:"records"`
	Report  string             `json:"report,omitempty"`
	Skipped bool               `json:"skipped,omitempty"`
}

// Execute runs an automation now, recording the run: the steps in order,
// each on the previous step's output (the first on the request's
// subjects). The returned error is about storage; a failed step shows in
// the run's status and step results.
func (s *Service) Execute(ctx context.Context, request Request) (*schema.AutomationRun, error) {
	automation, err := s.Automation(ctx, request.AutomationID)
	if err != nil || automation == nil {
		return nil, missing("automation", request.AutomationID, err)
	}
	runID, err := s.database.Insert(ctx, runTable, schema.AutomationRun{
		Automation:  automation.Id,
		StartedAt:   db.FormatTime(s.now()),
		Status:      schema.RunStatusRunning,
		DryRun:      request.DryRun,
		TriggerData: request.TriggerData,
		Steps:       []schema.StepResult{},
	})
	if err != nil {
		return nil, err
	}
	run := RunContext{
		RunID:        runID,
		AutomationID: automation.Id,
		Name:         automation.Name,
		DryRun:       request.DryRun,
		Author:       provenance.Automation(automation.Name, runID),
		TriggerData:  request.TriggerData,
	}
	results, failed := s.runSteps(ctx, automation.Steps, run, request.Subjects)
	if err := s.finishRun(ctx, runID, results, failed); err != nil {
		return nil, err
	}
	if failed && !request.DryRun {
		s.notifyFailure(ctx, *automation)
	}
	return s.AutomationRun(ctx, runID)
}

// runSteps runs steps in order and reports whether the run failed: a step
// erred without continueOnError, which stops the run.
func (s *Service) runSteps(ctx context.Context, steps []schema.Step, run RunContext, subjects []schema.RecordRef) ([]schema.StepResult, bool) {
	input := subjects
	if input == nil {
		input = []schema.RecordRef{}
	}
	results := []schema.StepResult{}
	for _, step := range steps {
		result, output, err := s.runStep(ctx, step, run, input)
		results = append(results, result)
		if err != nil && !step.ContinueOnError {
			return results, true
		}
		input = output
	}
	return results, false
}

// runStep runs one step and returns its result and the next step's input.
// A step whose condition no input matches is skipped and passes its input
// on; a step that erred passes its input on too.
func (s *Service) runStep(ctx context.Context, step schema.Step, run RunContext, input []schema.RecordRef) (schema.StepResult, []schema.RecordRef, error) {
	if step.Condition != nil {
		matching, err := s.matching(ctx, *step.Condition, input)
		if err != nil {
			return failedStep(step, err), input, err
		}
		if len(matching) == 0 {
			return schema.StepResult{Action: step.Action, Ok: true, Output: stepOutput{Records: input, Skipped: true}, Touched: []schema.RecordRef{}}, input, nil
		}
		input = matching
	}
	action, known := s.action(step.Action)
	if !known {
		err := fmt.Errorf("unknown action %q", step.Action)
		return failedStep(step, err), input, err
	}
	args, err := json.Marshal(step.Args)
	if err != nil {
		return failedStep(step, err), input, err
	}
	result, err := action(ctx, run, input, args)
	if err != nil {
		stepResult := failedStep(step, err)
		stepResult.Touched = nonNil(result.Touched)
		return stepResult, input, err
	}
	output := nonNil(result.Output)
	return schema.StepResult{Action: step.Action, Ok: true, Output: stepOutput{Records: output, Report: result.Report}, Touched: nonNil(result.Touched)}, output, nil
}

// failedStep is the result of a step that erred.
func failedStep(step schema.Step, err error) schema.StepResult {
	message := err.Error()
	return schema.StepResult{Action: step.Action, Ok: false, Error: &message, Touched: []schema.RecordRef{}}
}

// nonNil is refs, or an empty list for nil.
func nonNil(refs []schema.RecordRef) []schema.RecordRef {
	if refs == nil {
		return []schema.RecordRef{}
	}
	return refs
}

// finishRun stores a run's step results and final status.
func (s *Service) finishRun(ctx context.Context, runID string, results []schema.StepResult, failed bool) error {
	status := schema.RunStatusSucceeded
	if failed {
		status = schema.RunStatusFailed
	}
	return s.database.Update(ctx, runTable, runID, db.Vars{
		"finishedAt": db.FormatTime(s.now()),
		"status":     status,
		"steps":      results,
	})
}

// notifyFailure tells the user an automation failed.
func (s *Service) notifyFailure(ctx context.Context, automation schema.Automation) {
	if s.notifier == nil {
		return
	}
	notification := notify.NewNotification{Kind: "automationFailed", Title: "Automation failed: " + automation.Name}
	if _, err := s.notifier.Notify(ctx, notification); err != nil {
		log.Printf("automations: notify failure of %s: %v", automation.Name, err)
	}
}

// Undo reverts every revision a run wrote, newest first, and marks it
// undone. Reverts are user edits, so the fields stay locked to the restored
// values. What isn't a revision (notifications, collection items, sources
// the wanted hunt moved) stays.
func (s *Service) Undo(ctx context.Context, runID string) error {
	run, err := s.AutomationRun(ctx, runID)
	if err != nil || run == nil {
		return missing("automation run", runID, err)
	}
	if err := undoable(*run); err != nil {
		return err
	}
	revisionIDs, err := s.database.Strings(ctx,
		`SELECT "id" FROM "Revision" WHERE "automationRunId" = $run AND "deletedAt" IS NULL
		 ORDER BY "createdAt" DESC, rowid DESC`,
		db.Vars{"run": runID})
	if err != nil {
		return err
	}
	for _, revisionID := range revisionIDs {
		if err := s.library.Recorder().Revert(ctx, revisionID); err != nil {
			return fmt.Errorf("revert %s: %w", revisionID, err)
		}
	}
	return s.database.Update(ctx, runTable, runID, db.Vars{"status": schema.RunStatusUndone})
}

// undoable explains why a run can't be undone, or returns nil.
func undoable(run schema.AutomationRun) error {
	switch {
	case run.DryRun:
		return fmt.Errorf("a dry run changed nothing")
	case run.Status == schema.RunStatusRunning:
		return fmt.Errorf("the run is still going")
	case run.Status == schema.RunStatusUndone:
		return fmt.Errorf("the run was already undone")
	}
	return nil
}

// AutomationRun loads one run; nil when there is none.
func (s *Service) AutomationRun(ctx context.Context, id string) (*schema.AutomationRun, error) {
	return db.Get[schema.AutomationRun](ctx, s.database, runTable, id)
}

// Runs lists an automation's runs, newest first.
func (s *Service) Runs(ctx context.Context, automationID string, limit, offset int) ([]schema.AutomationRun, error) {
	if limit <= 0 {
		limit = 50
	}
	return db.QueryAs[schema.AutomationRun](ctx, s.database,
		`SELECT * FROM "AutomationRun" WHERE "automation" = $automation AND "deletedAt" IS NULL
		 ORDER BY "startedAt" DESC, rowid DESC LIMIT $limit OFFSET $offset`,
		db.Vars{"automation": automationID, "limit": limit, "offset": offset})
}
