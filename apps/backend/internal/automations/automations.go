// Package automations runs the user's multi-step automations: a trigger
// (an event, a schedule, new media from a follow, a cam going live, a new
// file, a wanted record found, a plugin) fires the enabled automations that
// match it, and each run executes their steps in order as an automationRun
// job, each step seeing the previous step's output.
//
// Steps are actions from a registry (RegisterAction). Writes go through
// provenance with the run as author, so Undo reverts exactly what a run
// changed. Writes that aren't revisions (collection items, moved sources)
// stay.
package automations

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/notify"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Table names, exactly as the schema names the models.
const (
	automationTable = "Automation"
	runTable        = "AutomationRun"
)

// Options are what the service works with. Notifier and Searcher may be
// nil: without a searcher the wanted hunt finds nothing.
type Options struct {
	Library  *library.Library
	Queue    *jobs.Queue
	Notifier *notify.Service
	Searcher Searcher
}

// Service stores automations and runs them.
type Service struct {
	database *db.DB
	library  *library.Library
	queue    *jobs.Queue
	notifier *notify.Service
	searcher Searcher
	now      func() time.Time

	actionsMutex sync.RWMutex
	actions      map[string]Action

	scheduleMutex sync.Mutex
	// lastScheduled is when each schedule automation last fired.
	lastScheduled map[string]time.Time
}

// New returns an automation service with the built-in actions registered.
func New(options Options) *Service {
	service := &Service{
		database:      options.Library.DB(),
		library:       options.Library,
		queue:         options.Queue,
		notifier:      options.Notifier,
		searcher:      options.Searcher,
		now:           time.Now,
		actions:       map[string]Action{},
		lastScheduled: map[string]time.Time{},
	}
	service.RegisterAction(ActionTag, service.tagAction)
	service.RegisterAction(ActionDelete, service.deleteAction)
	service.RegisterAction(ActionNotify, service.notifyAction)
	service.RegisterAction(ActionHunt, service.huntAction)
	return service
}

// NewAutomation is an automation to add.
type NewAutomation struct {
	Name    string
	Enabled bool
	Trigger schema.Trigger
	Steps   []schema.Step
}

// AutomationEdit changes an automation; nil fields stay as they are.
type AutomationEdit struct {
	Name    *string
	Enabled *bool
	Trigger *schema.Trigger
	Steps   *[]schema.Step
}

// Create adds an automation after checking its trigger and steps.
func (s *Service) Create(ctx context.Context, automation NewAutomation) (*schema.Automation, error) {
	row := schema.Automation{Name: strings.TrimSpace(automation.Name), Enabled: automation.Enabled, Trigger: automation.Trigger, Steps: automation.Steps}
	if err := s.validate(row); err != nil {
		return nil, err
	}
	id, err := s.database.Insert(ctx, automationTable, row)
	if err != nil {
		return nil, err
	}
	return s.Automation(ctx, id)
}

// Update applies an edit. Built-in automations can be edited like any other.
func (s *Service) Update(ctx context.Context, id string, edit AutomationEdit) (*schema.Automation, error) {
	automation, err := s.Automation(ctx, id)
	if err != nil || automation == nil {
		return nil, missing("automation", id, err)
	}
	applyEdit(automation, edit)
	if err := s.validate(*automation); err != nil {
		return nil, err
	}
	err = s.database.Update(ctx, automationTable, id, db.Vars{
		"name":    automation.Name,
		"enabled": automation.Enabled,
		"trigger": automation.Trigger,
		"steps":   automation.Steps,
	})
	if err != nil {
		return nil, err
	}
	return s.Automation(ctx, id)
}

// applyEdit copies the set fields of edit onto automation.
func applyEdit(automation *schema.Automation, edit AutomationEdit) {
	if edit.Name != nil {
		automation.Name = strings.TrimSpace(*edit.Name)
	}
	if edit.Enabled != nil {
		automation.Enabled = *edit.Enabled
	}
	if edit.Trigger != nil {
		automation.Trigger = *edit.Trigger
	}
	if edit.Steps != nil {
		automation.Steps = *edit.Steps
	}
}

// Delete moves an automation to the trash. Built-in ones can only be
// disabled.
func (s *Service) Delete(ctx context.Context, id string) error {
	automation, err := s.Automation(ctx, id)
	if err != nil || automation == nil {
		return missing("automation", id, err)
	}
	if automation.BuiltIn {
		return fmt.Errorf("%q is built in: disable it instead", automation.Name)
	}
	return s.database.Trash(ctx, automationTable, id)
}

// Automation loads one automation; nil when there is none.
func (s *Service) Automation(ctx context.Context, id string) (*schema.Automation, error) {
	return db.Get[schema.Automation](ctx, s.database, automationTable, id)
}

// Automations lists every automation, by name.
func (s *Service) Automations(ctx context.Context) ([]schema.Automation, error) {
	return db.QueryAs[schema.Automation](ctx, s.database,
		`SELECT * FROM "Automation" WHERE "deletedAt" IS NULL ORDER BY "name" COLLATE NOCASE, rowid`, nil)
}

// validate checks an automation's name, trigger and steps.
func (s *Service) validate(automation schema.Automation) error {
	if automation.Name == "" {
		return fmt.Errorf("an automation needs a name")
	}
	if err := validateTrigger(automation.Trigger); err != nil {
		return err
	}
	if len(automation.Steps) == 0 {
		return fmt.Errorf("an automation needs at least one step")
	}
	for _, step := range automation.Steps {
		if _, known := s.action(step.Action); !known {
			return fmt.Errorf("unknown action %q", step.Action)
		}
	}
	return nil
}

// validateTrigger checks that a trigger has what its kind needs.
func validateTrigger(trigger schema.Trigger) error {
	switch trigger.Kind {
	case schema.TriggerKindSchedule:
		if trigger.Cron == nil {
			return fmt.Errorf("a schedule trigger needs a cron expression")
		}
		_, err := parseCron(*trigger.Cron)
		return err
	case schema.TriggerKindPlugin:
		if trigger.PluginTrigger == nil || !strings.Contains(*trigger.PluginTrigger, ":") {
			return fmt.Errorf(`a plugin trigger needs pluginTrigger "<plugin>:<trigger>"`)
		}
		return nil
	case schema.TriggerKindEvent, schema.TriggerKindNewMedia, schema.TriggerKindCamLive,
		schema.TriggerKindFileAdded, schema.TriggerKindWantedFound:
		return nil
	}
	return fmt.Errorf("unknown trigger kind %q", trigger.Kind)
}

// missing is the error for a record that couldn't be loaded.
func missing(what, id string, err error) error {
	if err != nil {
		return err
	}
	return fmt.Errorf("no %s %s", what, id)
}
