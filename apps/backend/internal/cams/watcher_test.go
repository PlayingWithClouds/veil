package cams

import (
	"sync"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// liveHook records the models OnLive was called with.
type liveHook struct {
	mutex  sync.Mutex
	models []string
}

// called records one call.
func (hook *liveHook) called(camModel schema.CamModel) {
	hook.mutex.Lock()
	defer hook.mutex.Unlock()
	hook.models = append(hook.models, camModel.Username)
}

// count is how many calls there were.
func (hook *liveHook) count() int {
	hook.mutex.Lock()
	defer hook.mutex.Unlock()
	return len(hook.models)
}

// follow stores an enabled Follow of the room.
func (f *fixture) follow(t *testing.T) {
	t.Helper()
	_, err := f.database.Insert(f.ctx, "Follow", schema.Follow{
		Name: "Nova", TargetKind: ptrTo(schema.RecordKindCamModel), TargetId: &f.roomID,
		IntervalHours: 1, Enabled: true, Notify: true,
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ptrTo returns a pointer to value.
func ptrTo[T any](value T) *T {
	return &value
}

// setOffline stores the room as offline and forgets the cached answer.
func (f *fixture) setOffline(t *testing.T) {
	t.Helper()
	if _, err := f.database.Exec(f.ctx, `UPDATE "CamModel" SET "live" = NULL WHERE "id" = $id`, db.Vars{"id": f.roomID}); err != nil {
		t.Fatal(err)
	}
	f.service.online = map[string]cachedOnline{}
}

// TestCheckAnnouncesFollowedModelsThatGoLive checks the OnLive hook: not for
// a show already running when first seen, once for a new show, again after
// the model was offline.
func TestCheckAnnouncesFollowedModelsThatGoLive(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	hook := &liveHook{}
	f.service.SetOnLive(hook.called)
	f.follow(t)

	f.service.Check(f.ctx)
	if hook.count() != 0 {
		t.Fatalf("a show running when first seen was announced: %v", hook.models)
	}

	f.sites.setOffline("nova", true)
	f.service.online = map[string]cachedOnline{}
	f.service.Check(f.ctx)
	camModel, _ := f.service.load(f.ctx, f.roomID)
	if camModel.Live != nil {
		t.Fatalf("offline model still live: %+v", camModel.Live)
	}

	f.sites.setOffline("nova", false)
	f.service.online = map[string]cachedOnline{}
	f.service.Check(f.ctx)
	f.service.online = map[string]cachedOnline{}
	f.service.Check(f.ctx)
	if hook.count() != 1 {
		t.Fatalf("OnLive called %d times for one new show, want once", hook.count())
	}
}

// TestCheckIgnoresUnwatchedModels checks that rooms neither followed nor
// auto-recorded are not probed or announced.
func TestCheckIgnoresUnwatchedModels(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	hook := &liveHook{}
	f.service.SetOnLive(hook.called)
	f.setOffline(t)

	f.service.Check(f.ctx)
	if f.sites.streamCalls.Load() != 0 || hook.count() != 0 {
		t.Fatalf("an unwatched room was probed (%d) or announced (%d)", f.sites.streamCalls.Load(), hook.count())
	}
}

// TestCheckStartsAutoRecordingAtTheModelsQuality checks that a live
// auto-record model is recorded, at its quality.
func TestCheckStartsAutoRecordingAtTheModelsQuality(t *testing.T) {
	f := newFixture(t, [][]int{{0, 1, 2}}, true)
	hook := &liveHook{}
	f.service.SetOnLive(hook.called)
	if _, err := f.service.SetAutoRecord(f.ctx, f.roomID, true, "480"); err != nil {
		t.Fatal(err)
	}
	f.setOffline(t)

	f.service.Check(f.ctx)
	waitFor(t, "the auto-recording", func() bool { return f.recordedSeconds() >= 4 })
	if active := f.service.ActiveRecordings(); len(active) != 1 || active[0].Quality != "480p" {
		t.Fatalf("the watcher did not record at 480p: %+v", active)
	}
	if hook.count() != 1 {
		t.Errorf("OnLive called %d times, want once", hook.count())
	}
	f.stop(t)
}

// TestCheckTakesStaleRoomsOffline checks that a room nobody saw online for a
// while loses its live state.
func TestCheckTakesStaleRoomsOffline(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	later := time.Now().Add(liveStaleAfter + time.Minute)
	f.service.now = func() time.Time { return later }

	f.service.Check(f.ctx)

	camModel, _ := f.service.load(f.ctx, f.roomID)
	if camModel.Live != nil {
		t.Errorf("stale room still live: %+v", camModel.Live)
	}
}
