package automations

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/library"
	"github.com/playingwithclouds/veil/internal/notify"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// fixture is an automation service with its collaborators.
type fixture struct {
	service  *Service
	library  *library.Library
	database *db.DB
	notifier *notify.Service
	queue    *jobs.Queue
	searcher *fakeSearcher
}

// fakeSearcher lists fixed scenes into the library as a plugin search would.
type fakeSearcher struct {
	library *library.Library
	items   []plugins.DiscoveredItem
	queries []string
}

// Search stores the fixed items as listing stubs of plugin "site".
func (f *fakeSearcher) Search(ctx context.Context, entity plugins.MediaType, query string, limit int) []schema.RecordRef {
	f.queries = append(f.queries, query)
	refs := []schema.RecordRef{}
	for _, item := range f.items {
		ref, err := f.library.IngestListed(ctx, "site", item)
		if err == nil {
			refs = append(refs, ref)
		}
	}
	return refs
}

// newFixture opens everything on a fresh database.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	database, err := db.Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	lib := library.New(database)
	f := &fixture{library: lib, database: database, notifier: notify.New(database), queue: jobs.NewQueue(database, 3)}
	f.searcher = &fakeSearcher{library: lib}
	f.service = New(Options{Library: lib, Queue: f.queue, Notifier: f.notifier, Searcher: f.searcher})
	return f
}

// addVideo inserts a video with tags and returns its ref.
func (f *fixture) addVideo(t *testing.T, title string, status schema.MediaStatus, tags ...string) schema.RecordRef {
	t.Helper()
	if tags == nil {
		tags = []string{}
	}
	id, err := f.database.Insert(context.Background(), "Video", db.Vars{"title": title, "status": status, "performers": []string{}, "tags": tags, "durationSeconds": 600})
	if err != nil {
		t.Fatal(err)
	}
	return schema.RecordRef{Kind: schema.RecordKindVideo, Id: id}
}

// create adds an automation or fails the test.
func (f *fixture) create(t *testing.T, automation NewAutomation) *schema.Automation {
	t.Helper()
	created, err := f.service.Create(context.Background(), automation)
	if err != nil {
		t.Fatal(err)
	}
	return created
}

// runQueued runs every queued automationRun job and returns the runs.
func (f *fixture) runQueued(t *testing.T) []schema.AutomationRun {
	t.Helper()
	ctx := context.Background()
	queued, err := f.queue.List(ctx, jobs.Query{Kinds: []jobs.Kind{schema.JobKindAutomationRun}, Statuses: []jobs.Status{jobs.StatusQueued}, Limit: 50})
	if err != nil {
		t.Fatal(err)
	}
	for _, job := range queued {
		if err := f.service.RunTask()(ctx, &job, func(float64) {}); err != nil {
			t.Fatal(err)
		}
		f.queue.Complete(ctx, job.ID)
	}
	runs, err := db.QueryAs[schema.AutomationRun](ctx, f.database, `SELECT * FROM "AutomationRun" ORDER BY rowid`, nil)
	if err != nil {
		t.Fatal(err)
	}
	return runs
}

// tagsOf loads a video's tags.
func (f *fixture) tagsOf(t *testing.T, ref schema.RecordRef) []string {
	t.Helper()
	video, err := f.library.Video(context.Background(), ref.Id)
	if err != nil || video == nil {
		t.Fatalf("video %s: %v", ref.Id, err)
	}
	return video.Tags
}

// tagStep is a step adding tag ids.
func tagStep(add ...string) schema.Step {
	return schema.Step{Action: ActionTag, Args: map[string]any{"add": add}}
}

// TestValidation refuses unknown actions, bad cron expressions and
// deleting a built-in automation.
func TestValidation(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	if _, err := f.service.Create(ctx, NewAutomation{Name: "x", Trigger: schema.Trigger{Kind: schema.TriggerKindNewMedia}, Steps: []schema.Step{{Action: "nope"}}}); err == nil {
		t.Error("unknown action accepted")
	}
	badCron := "every day"
	if _, err := f.service.Create(ctx, NewAutomation{Name: "x", Trigger: schema.Trigger{Kind: schema.TriggerKindSchedule, Cron: &badCron}, Steps: []schema.Step{tagStep("t")}}); err == nil {
		t.Error("bad cron accepted")
	}
	for range 2 {
		if err := f.service.EnsureBuiltIns(ctx); err != nil {
			t.Fatal(err)
		}
	}
	automations, _ := f.service.Automations(ctx)
	if len(automations) != 1 || !automations[0].BuiltIn {
		t.Fatalf("automations = %+v, want the wanted hunt once", automations)
	}
	if err := f.service.Delete(ctx, WantedHuntID); err == nil {
		t.Error("built-in automation deleted")
	}
	disabled := false
	if _, err := f.service.Update(ctx, WantedHuntID, AutomationEdit{Enabled: &disabled}); err != nil {
		t.Fatal(err)
	}
}

// TestFireRunUndo fires on the subjects matching the trigger filter, tags
// them as the run, notifies, and undoes the tagging.
func TestFireRunUndo(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	kept := f.addVideo(t, "Kept", schema.MediaStatusAvailable, "t0")
	skipped := f.addVideo(t, "Skipped", schema.MediaStatusAvailable, "blocked")
	automation := f.create(t, NewAutomation{Name: "Tag new", Enabled: true,
		Trigger: schema.Trigger{Kind: schema.TriggerKindNewMedia, Filter: &schema.Filter{ExcludeTagIds: []string{"blocked"}}},
		Steps:   []schema.Step{tagStep("t1"), {Action: ActionNotify, Args: map[string]any{"title": "Tagged"}}}})
	f.create(t, NewAutomation{Name: "Other trigger", Enabled: true, Trigger: schema.Trigger{Kind: schema.TriggerKindCamLive}, Steps: []schema.Step{tagStep("t9")}})

	jobIDs, err := f.service.Fire(ctx, schema.TriggerKindNewMedia, map[string]any{"follow": "f1"}, []schema.RecordRef{kept, skipped})
	if err != nil || len(jobIDs) != 1 {
		t.Fatalf("Fire = %v, %v; want one job", jobIDs, err)
	}
	runs := f.runQueued(t)
	if len(runs) != 1 || runs[0].Status != schema.RunStatusSucceeded || runs[0].Automation != automation.Id {
		t.Fatalf("runs = %+v", runs)
	}
	if !slices.Equal(runs[0].Steps[0].Touched, []schema.RecordRef{kept}) {
		t.Errorf("tag step touched %+v, want only the kept video", runs[0].Steps[0].Touched)
	}
	if tags := f.tagsOf(t, kept); !slices.Equal(tags, []string{"t0", "t1"}) {
		t.Errorf("tags = %v, want t0 t1", tags)
	}
	if tags := f.tagsOf(t, skipped); !slices.Equal(tags, []string{"blocked"}) {
		t.Errorf("filtered-out video was tagged: %v", tags)
	}
	notifications, _ := f.notifier.Notifications(ctx, false, 10, 0)
	if len(notifications) != 1 || notifications[0].Title != "Tagged" || notifications[0].Subject == nil {
		t.Errorf("notifications = %+v", notifications)
	}

	if err := f.service.Undo(ctx, runs[0].Id); err != nil {
		t.Fatal(err)
	}
	if tags := f.tagsOf(t, kept); !slices.Equal(tags, []string{"t0"}) {
		t.Errorf("tags after undo = %v, want t0", tags)
	}
	if err := f.service.Undo(ctx, runs[0].Id); err == nil {
		t.Error("run undone twice")
	}
}

// TestDryRun reports what the steps would touch and writes nothing.
func TestDryRun(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	video := f.addVideo(t, "Video", schema.MediaStatusAvailable)
	automation := f.create(t, NewAutomation{Name: "Clean", Enabled: true, Trigger: schema.Trigger{Kind: schema.TriggerKindFileAdded},
		Steps: []schema.Step{tagStep("t1"), {Action: ActionDelete}, {Action: ActionNotify}}})
	run, err := f.service.Execute(ctx, Request{AutomationID: automation.Id, Subjects: []schema.RecordRef{video}, DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if !run.DryRun || run.Status != schema.RunStatusSucceeded || len(run.Steps) != 3 {
		t.Fatalf("run = %+v", run)
	}
	if !slices.Equal(run.Steps[0].Touched, []schema.RecordRef{video}) || !slices.Equal(run.Steps[1].Touched, []schema.RecordRef{video}) {
		t.Errorf("dry run touched %+v / %+v, want the video", run.Steps[0].Touched, run.Steps[1].Touched)
	}
	if tags := f.tagsOf(t, video); len(tags) != 0 {
		t.Errorf("dry run tagged: %v", tags)
	}
	notifications, _ := f.notifier.Notifications(ctx, false, 10, 0)
	if len(notifications) != 0 {
		t.Errorf("dry run notified: %+v", notifications)
	}
	if err := f.service.Undo(ctx, run.Id); err == nil {
		t.Error("dry run undone")
	}
}

// TestStepFailures continues past a failing step that allows it, stops at
// one that doesn't, and skips a step whose condition doesn't match.
func TestStepFailures(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	f.service.RegisterAction("fail", func(ctx context.Context, run RunContext, input []schema.RecordRef, args json.RawMessage) (Result, error) {
		return Result{}, errors.New("boom")
	})
	video := f.addVideo(t, "Video", schema.MediaStatusAvailable)
	trigger := schema.Trigger{Kind: schema.TriggerKindWantedFound}
	tolerant := f.create(t, NewAutomation{Name: "Tolerant", Enabled: true, Trigger: trigger, Steps: []schema.Step{
		{Action: "fail", ContinueOnError: true},
		{Action: ActionTag, Args: map[string]any{"add": []string{"never"}}, Condition: &schema.Filter{IncludeTagIds: []string{"absent"}}},
		{Action: ActionDelete},
	}})
	run, err := f.service.Execute(ctx, Request{AutomationID: tolerant.Id, Subjects: []schema.RecordRef{video}})
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != schema.RunStatusSucceeded || run.Steps[0].Ok || run.Steps[0].Error == nil {
		t.Errorf("tolerant run = %+v", run)
	}
	if trashed, _ := f.library.Video(ctx, video.Id); trashed != nil {
		t.Error("delete step didn't trash the video")
	}
	if err := f.service.Undo(ctx, run.Id); err != nil {
		t.Fatal(err)
	}
	if restored, _ := f.library.Video(ctx, video.Id); restored == nil || len(restored.Tags) != 0 {
		t.Errorf("undo didn't restore the untagged video: %+v", restored)
	}

	strict := f.create(t, NewAutomation{Name: "Strict", Enabled: true, Trigger: trigger, Steps: []schema.Step{{Action: "fail"}, {Action: ActionDelete}}})
	run, err = f.service.Execute(ctx, Request{AutomationID: strict.Id, Subjects: []schema.RecordRef{video}})
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != schema.RunStatusFailed || len(run.Steps) != 1 {
		t.Errorf("strict run = %+v, want failed after one step", run)
	}
	notifications, _ := f.notifier.Notifications(ctx, false, 10, 0)
	if len(notifications) != 1 || notifications[0].Kind != "automationFailed" {
		t.Errorf("notifications = %+v, want one automationFailed", notifications)
	}
}

// TestHandleEvent fires only automations listening for the event's kind.
func TestHandleEvent(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	video := f.addVideo(t, "Video", schema.MediaStatusAvailable)
	f.create(t, NewAutomation{Name: "On o", Enabled: true, Trigger: schema.Trigger{Kind: schema.TriggerKindEvent, EventKinds: []string{"o"}}, Steps: []schema.Step{tagStep("done")}})
	kind := schema.RecordKindVideo
	event := schema.Event{Entity: schema.Entity{Id: "e1"}, Kind: schema.EventKindView, SubjectKind: &kind, SubjectId: &video.Id}
	if err := f.service.HandleEvent(ctx, event); err != nil {
		t.Fatal(err)
	}
	if runs := f.runQueued(t); len(runs) != 0 {
		t.Fatalf("a view fired the o automation: %+v", runs)
	}
	event.Kind = schema.EventKindO
	if err := f.service.HandleEvent(ctx, event); err != nil {
		t.Fatal(err)
	}
	f.runQueued(t)
	if tags := f.tagsOf(t, video); !slices.Equal(tags, []string{"done"}) {
		t.Errorf("tags = %v, want done", tags)
	}
}

// TestFireDue fires a schedule once its next time after the last edit has
// come, and once per time.
func TestFireDue(t *testing.T) {
	f := newFixture(t)
	every := "*/5 * * * *"
	f.create(t, NewAutomation{Name: "Often", Enabled: true, Trigger: schema.Trigger{Kind: schema.TriggerKindSchedule, Cron: &every}, Steps: []schema.Step{{Action: ActionNotify}}})
	now := time.Now()
	assertFired(t, f, now.Add(-time.Minute), 0)
	assertFired(t, f, now.Add(6*time.Minute), 1)
	assertFired(t, f, now.Add(6*time.Minute), 0)
	assertFired(t, f, now.Add(12*time.Minute), 1)
}

// assertFired runs FireDue at now and expects count automations queued.
func assertFired(t *testing.T, f *fixture, now time.Time, count int) {
	t.Helper()
	fired, err := f.service.FireDue(context.Background(), now)
	if err != nil {
		t.Fatal(err)
	}
	if fired != count {
		t.Errorf("FireDue(%s) fired %d, want %d", now.Format(time.TimeOnly), fired, count)
	}
}

// TestWantedHunt finds a wanted video on a site, takes over the copy's
// source and stream, trashes the copy and fires wantedFound.
func TestWantedHunt(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t)
	wanted := f.addVideo(t, "Sunset Pool Party Scene", schema.MediaStatusWanted)
	f.searcher.items = []plugins.DiscoveredItem{
		{MediaType: plugins.MediaTypeScene, ExternalID: "other", SourceURL: "https://site.test/other", Title: "Something Else Entirely", DurationSeconds: 600},
		{MediaType: plugins.MediaTypeScene, ExternalID: "copy", SourceURL: "https://site.test/copy", Title: "Sunset Pool Party Scene", DurationSeconds: 602,
			Downloads: []plugins.Download{{URL: "https://cdn.test/copy.mp4"}}},
	}
	f.create(t, NewAutomation{Name: "On found", Enabled: true, Trigger: schema.Trigger{Kind: schema.TriggerKindWantedFound}, Steps: []schema.Step{tagStep("found")}})
	if err := f.service.EnsureBuiltIns(ctx); err != nil {
		t.Fatal(err)
	}

	run, err := f.service.Execute(ctx, Request{AutomationID: WantedHuntID})
	if err != nil {
		t.Fatal(err)
	}
	if run.Status != schema.RunStatusSucceeded || len(f.searcher.queries) != 1 || f.searcher.queries[0] != "Sunset Pool Party Scene" {
		t.Fatalf("run = %+v, queries = %v", run, f.searcher.queries)
	}
	video, _ := f.library.Video(ctx, wanted.Id)
	if video.Status != schema.MediaStatusAvailable {
		t.Errorf("wanted video status = %s, want available", video.Status)
	}
	sources, _ := f.library.Sources(ctx, schema.RecordKindVideo, wanted.Id)
	streams, _ := f.library.Streams(ctx, schema.RecordKindVideo, wanted.Id)
	if len(sources) != 1 || sources[0].ExternalId != "copy" || len(streams) != 1 {
		t.Errorf("sources = %+v, streams = %+v; want the copy's", sources, streams)
	}
	assertCopyTrashed(t, f)
	notifications, _ := f.notifier.Notifications(ctx, false, 10, 0)
	if len(notifications) != 1 || notifications[0].Kind != "wantedFound" {
		t.Errorf("notifications = %+v", notifications)
	}
	f.runQueued(t)
	if tags := f.tagsOf(t, wanted); !slices.Equal(tags, []string{"found"}) {
		t.Errorf("wantedFound automation didn't run: tags = %v", tags)
	}
}

// assertCopyTrashed expects the found copy in the trash and the unrelated
// result left alone.
func assertCopyTrashed(t *testing.T, f *fixture) {
	t.Helper()
	trashed, err := f.library.Trashed(context.Background(), schema.RecordKindVideo, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(trashed) != 1 || trashed[0].Label != "Sunset Pool Party Scene" {
		t.Errorf("trash = %+v, want the copy", trashed)
	}
}
