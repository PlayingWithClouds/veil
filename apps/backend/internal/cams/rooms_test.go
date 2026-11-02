package cams

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// listUsernames lists one page and returns its usernames.
func listUsernames(t *testing.T, f *fixture, query Query) string {
	t.Helper()
	rooms, _, err := f.service.List(f.ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, camModel := range rooms {
		names = append(names, camModel.Username)
	}
	return fmt.Sprint(names)
}

// TestListMergesSitesByViewersAndPages checks the merge order, that a later
// page continues the first and that the site filter applies.
func TestListMergesSitesByViewersAndPages(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	f.sites.rooms = map[string][]plugins.CamRoom{
		"alpha": {room("a1", 900), room("a2", 100)},
		"beta":  {room("b1", 500), room("b2", 400), room("b3", 5)},
	}

	if got := listUsernames(t, f, Query{Limit: 3}); got != "[a1 b1 b2]" {
		t.Errorf("first page = %s", got)
	}
	if got := listUsernames(t, f, Query{Limit: 3, Offset: 3}); got != "[a2 b3]" {
		t.Errorf("second page = %s", got)
	}
	if got := listUsernames(t, f, Query{Plugins: []string{"beta"}, Limit: 5}); got != "[b1 b2 b3]" {
		t.Errorf("beta only = %s", got)
	}
	_, total, _ := f.service.List(f.ctx, Query{})
	if total != 5 {
		t.Errorf("total = %d, want 5", total)
	}
}

// TestListIsCachedAndSkipsPrivateRooms checks that paging does not hit the
// site again within the cache window and that rooms in a private show are
// left out.
func TestListIsCachedAndSkipsPrivateRooms(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	hidden := room("hidden", 1000)
	hidden.Online = false
	f.sites.rooms = map[string][]plugins.CamRoom{"alpha": {hidden, room("a1", 10)}}

	if got := listUsernames(t, f, Query{}); got != "[a1]" {
		t.Errorf("rooms = %s", got)
	}
	listUsernames(t, f, Query{Offset: 1})
	if f.sites.listCalls != 1 {
		t.Errorf("site listed %d times, want 1", f.sites.listCalls)
	}
	later := time.Now().Add(listCacheTTL + time.Second)
	f.service.now = func() time.Time { return later }
	listUsernames(t, f, Query{})
	if f.sites.listCalls != 2 {
		t.Errorf("after expiry site listed %d times, want 2", f.sites.listCalls)
	}
}

// TestListStoresRoomsWithLiveState checks that a listed room becomes a live
// CamModel, and that listing it again refreshes the audience but keeps the
// start of the show.
func TestListStoresRoomsWithLiveState(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	rooms, _, err := f.service.List(f.ctx, Query{})
	if err != nil || len(rooms) != 1 || rooms[0].Id != f.roomID {
		t.Fatalf("rooms = %+v %v, want the stored nova", rooms, err)
	}
	startedAt := liveSince(rooms[0])
	if startedAt == "" {
		t.Fatal("listed room has no live state")
	}

	f.sites.rooms["alpha"] = []plugins.CamRoom{room("nova", 77)}
	later := time.Now().Add(listCacheTTL + time.Minute)
	f.service.now = func() time.Time { return later }
	rooms, _, _ = f.service.List(f.ctx, Query{})
	if rooms[0].Live.Viewers != 77 || liveSince(rooms[0]) != startedAt {
		t.Errorf("relisted live = %+v, want 77 viewers since %s", rooms[0].Live, startedAt)
	}
	if rooms[0].DisplayName != "nova" || len(rooms[0].Tags) != 1 || rooms[0].Thumbnail == nil {
		t.Errorf("relisted room = %+v", rooms[0])
	}
}

// TestMergeByViewersKeepsARepeatedRoomOnce checks that a room a site moved
// between chunks keeps its first, best place.
func TestMergeByViewersKeepsARepeatedRoomOnce(t *testing.T) {
	withViewers := func(id string, viewers int32) schema.CamModel {
		return schema.CamModel{Entity: schema.Entity{Id: id}, Live: &schema.CamLive{Viewers: viewers}}
	}
	first := []schema.CamModel{withViewers("a", 900), withViewers("b", 500)}
	second := []schema.CamModel{withViewers("b", 400), withViewers("c", 300)}

	ids := []string{}
	for _, merged := range mergeByViewers([][]schema.CamModel{first, second}) {
		ids = append(ids, merged.Id)
	}
	if strings.Join(ids, ",") != "a,b,c" {
		t.Errorf("merged = %v, want each room once in viewer order", ids)
	}
}

// TestWatchResolvesTheRoomAndRecordsItsState checks the proxied URL and that
// an unreachable room is stored offline.
func TestWatchResolvesTheRoomAndRecordsItsState(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	f.sites.live = nil
	playableURL, mimeType, err := f.service.Watch(f.ctx, f.roomID)
	if err != nil || playableURL != "http://backend.test/api/stream/manifest?live=https://edge.example/nova.m3u8" || mimeType != "application/x-mpegURL" {
		t.Fatalf("watch = %q %q %v", playableURL, mimeType, err)
	}

	f.sites.setOffline("nova", true)
	if _, _, err := f.service.Watch(f.ctx, f.roomID); err == nil {
		t.Fatal("watching an offline room succeeded")
	}
	camModel, _ := f.service.load(f.ctx, f.roomID)
	if camModel.Live != nil {
		t.Errorf("offline room still live: %+v", camModel.Live)
	}
	live := true
	if models, _ := f.service.Models(f.ctx, ModelFilter{Live: &live}); len(models) != 0 {
		t.Errorf("live models = %+v", models)
	}
}

// TestSetAutoRecordAndLinkPerformer checks the stored settings of a model.
func TestSetAutoRecordAndLinkPerformer(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	camModel, err := f.service.SetAutoRecord(f.ctx, f.roomID, true, "480p")
	if err != nil || !camModel.AutoRecord || camModel.RecordQuality == nil || *camModel.RecordQuality != "480" {
		t.Fatalf("auto-record = %+v %v", camModel, err)
	}
	if camModel, _ = f.service.SetAutoRecord(f.ctx, f.roomID, true, ""); *camModel.RecordQuality != "480" {
		t.Errorf("leaving the quality out lost it: %+v", camModel.RecordQuality)
	}
	if _, err := f.service.SetAutoRecord(f.ctx, f.roomID, true, "tall"); err == nil {
		t.Error("accepted an invalid quality")
	}
	on := true
	if models, _ := f.service.Models(f.ctx, ModelFilter{AutoRecord: &on}); len(models) != 1 {
		t.Errorf("auto-recorded models = %+v", models)
	}

	performerID := insertPerformer(t, f, "Nova Star")
	if camModel, err = f.service.LinkPerformer(f.ctx, f.roomID, performerID); err != nil || camModel.Performer == nil || *camModel.Performer != performerID {
		t.Fatalf("link = %+v %v", camModel, err)
	}
	if models, _ := f.service.Models(f.ctx, ModelFilter{PerformerID: performerID}); len(models) != 1 {
		t.Errorf("models of the performer = %+v", models)
	}
	if _, err := f.service.LinkPerformer(f.ctx, f.roomID, db.NewID()); err == nil {
		t.Error("linked a performer that doesn't exist")
	}
	if camModel, _ = f.service.LinkPerformer(f.ctx, f.roomID, ""); camModel.Performer != nil {
		t.Errorf("unlink left %v", *camModel.Performer)
	}
}

// insertPerformer stores a performer and returns its id.
func insertPerformer(t *testing.T, f *fixture, name string) string {
	t.Helper()
	id, err := f.database.Insert(f.ctx, "Performer", db.Vars{
		"name": name, "nameKey": strings.ToLower(name), "aliases": []string{}, "attributes": map[string]any{},
		"links": []string{}, "tags": []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return id
}
