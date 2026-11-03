package cams

import (
	"fmt"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// startedRoom is an online listing room whose show began at startedAt.
func startedRoom(username string, viewers int, startedAt string) plugins.CamRoom {
	listed := room(username, viewers)
	listed.StartedAt = startedAt
	return listed
}

// TestListSortsTheSiteOrderByViewers checks that a site's popularity order,
// which only roughly follows viewers, is listed most viewers first.
func TestListSortsTheSiteOrderByViewers(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	f.sites.rooms = map[string][]plugins.CamRoom{"alpha": {room("a", 10), room("b", 50), room("c", 30)}}

	if got := listUsernames(t, f, Query{}); got != "[b c a]" {
		t.Errorf("rooms = %s, want most viewers first", got)
	}
}

// TestListFewestStartsAtTheEndOfTheSite checks that the fewest-viewers sort
// reads the site's list from its last chunk.
func TestListFewestStartsAtTheEndOfTheSite(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	var rooms []plugins.CamRoom
	for index := range 150 {
		rooms = append(rooms, room(fmt.Sprintf("r%d", index), 1000-index))
	}
	f.sites.rooms = map[string][]plugins.CamRoom{"alpha": rooms}

	if got := listUsernames(t, f, Query{Sort: SortFewest, Limit: 3}); got != "[r149 r148 r147]" {
		t.Errorf("fewest = %s", got)
	}
	if got := listUsernames(t, f, Query{Sort: SortFewest, Limit: 2, Offset: 50}); got != "[r99 r98]" {
		t.Errorf("fewest past the last chunk = %s", got)
	}
}

// TestListStartedUsesTheSiteShowStart checks that the site's show start is
// stored and orders the started sort.
func TestListStartedUsesTheSiteShowStart(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	f.sites.rooms = map[string][]plugins.CamRoom{"alpha": {
		startedRoom("early", 900, "2026-11-03T10:00:00+00:00"),
		startedRoom("late", 100, "2026-11-03T12:30:00.5+00:00"),
	}}

	if got := listUsernames(t, f, Query{Sort: SortStarted}); got != "[late early]" {
		t.Errorf("started = %s", got)
	}
	rooms, _, _ := f.service.List(f.ctx, Query{Sort: SortStarted, Limit: 1})
	if liveSince(rooms[0]) != db.FormatTime(time.Date(2026, 11, 3, 12, 30, 0, 500_000_000, time.UTC)) {
		t.Errorf("stored start = %q", liveSince(rooms[0]))
	}
}

// TestListLeavesOutBlockedRooms checks that a room the user blocked isn't listed.
func TestListLeavesOutBlockedRooms(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	f.sites.rooms = map[string][]plugins.CamRoom{"alpha": {room("nova", 50), room("other", 10)}}
	_, err := f.database.Insert(f.ctx, "UserState", db.Vars{
		"subjectKind": schema.RecordKindCamModel, "subjectId": f.roomID, "favorite": false,
		"blocked": true, "viewCount": 0, "fields": map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}

	if got := listUsernames(t, f, Query{}); got != "[other]" {
		t.Errorf("rooms = %s, want the blocked nova left out", got)
	}
}

// TestGoingOfflineStoresTheShow checks that a room nobody saw online lately
// is taken offline and its show is kept.
func TestGoingOfflineStoresTheShow(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	f.sites.rooms = map[string][]plugins.CamRoom{"alpha": {startedRoom("nova", 50, "2026-11-03T10:00:00Z")}}
	listUsernames(t, f, Query{})
	seen, _ := f.service.load(f.ctx, f.roomID)

	later := time.Now().Add(liveStaleAfter + time.Minute)
	f.service.now = func() time.Time { return later }
	f.service.expireStaleLive(f.ctx)

	camModel, _ := f.service.load(f.ctx, f.roomID)
	if camModel.Live != nil {
		t.Errorf("stale room still live: %+v", camModel.Live)
	}
	shows, err := f.service.Shows(f.ctx, f.roomID)
	if err != nil || len(shows) != 1 {
		t.Fatalf("shows = %+v %v, want one", shows, err)
	}
	if shows[0].StartedAt != "2026-11-03T10:00:00.000Z" || shows[0].EndedAt != *seen.LastSeenOnlineAt {
		t.Errorf("show = %s to %s, want the site's start to %s", shows[0].StartedAt, shows[0].EndedAt, *seen.LastSeenOnlineAt)
	}
}

// TestRecommendedRanksRoomsByLikedTags checks that rooms are ranked by the
// tags the user likes, however the site spells them, and that rooms matching
// nothing or disliked tags are left out.
func TestRecommendedRanksRoomsByLikedTags(t *testing.T) {
	f := newFixture(t, [][]int{{0}}, false)
	tagged := func(username string, viewers int, tags ...string) plugins.CamRoom {
		listed := room(username, viewers)
		listed.Tags = tags
		return listed
	}
	f.sites.rooms = map[string][]plugins.CamRoom{"alpha": {
		tagged("busty", 10, "bigboobs"),
		tagged("both", 5, "bigboobs", "redhead"),
		tagged("plain", 900, "chat"),
		tagged("disliked", 50, "feet"),
	}}
	affinity := Affinity{Tags: map[string]float64{"Big Boobs": 0.6, "Redhead": 0.3, "Feet": -0.8}}

	rooms, err := f.service.Recommended(f.ctx, affinity, 10)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, camModel := range rooms {
		names = append(names, camModel.Username)
	}
	if fmt.Sprint(names) != "[both busty]" {
		t.Errorf("recommended = %v", names)
	}
}
