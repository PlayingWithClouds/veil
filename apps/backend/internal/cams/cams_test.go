package cams

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// fakeSites serves canned rooms per plugin and counts listing calls.
type fakeSites struct {
	rooms     map[string][]plugins.CamRoom
	offline   map[string]bool
	listCalls int
}

func (sites *fakeSites) Sources() []*plugins.Plugin {
	var sources []*plugins.Plugin
	for _, name := range []string{"alpha", "beta"} {
		if _, ok := sites.rooms[name]; ok {
			sources = append(sources, &plugins.Plugin{Meta: plugins.PluginMeta{Name: name}})
		}
	}
	return sources
}

func (sites *fakeSites) Plugin(name string) (*plugins.Plugin, bool) {
	for _, plugin := range sites.Sources() {
		if plugin.Meta.Name == name {
			return plugin, true
		}
	}
	return nil, false
}

func (sites *fakeSites) List(ctx context.Context, plugin *plugins.Plugin, args plugins.CamListArgs) (*plugins.CamListResult, error) {
	sites.listCalls++
	rooms := sites.rooms[plugin.Meta.Name]
	return &plugins.CamListResult{Rooms: rooms, Total: len(rooms)}, nil
}

func (sites *fakeSites) Stream(ctx context.Context, plugin *plugins.Plugin, username string) (*plugins.ResolveResult, error) {
	if sites.offline[username] {
		return nil, errors.New("offline")
	}
	return &plugins.ResolveResult{URL: "https://edge.example/" + username + ".m3u8", MimeType: "application/x-mpegURL"}, nil
}

type fakeProxy struct{}

func (fakeProxy) LiveManifestURL(ctx context.Context, result *plugins.ResolveResult) (string, error) {
	return "http://backend.test/api/stream/manifest?live=" + result.URL, nil
}

func room(username string, viewers int) plugins.CamRoom {
	return plugins.CamRoom{
		Username: username, SourceURL: "https://site.example/" + username + "/",
		Viewers: viewers, Tags: []string{"tag"}, Online: true,
	}
}

func newTestService(t *testing.T, sites *fakeSites) *Service {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	return New(database, sites, fakeProxy{})
}

func usernames(page []string) string { return fmt.Sprint(page) }

func pageUsernames(t *testing.T, service *Service, filter Filter) []string {
	t.Helper()
	page, err := service.Rooms(context.Background(), filter)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, camModel := range page.Rooms {
		names = append(names, camModel.Username)
	}
	return names
}

// TestRoomsMergeSitesByViewersAndPage checks the merge order and that a later
// page continues the first.
func TestRoomsMergeSitesByViewersAndPage(t *testing.T) {
	sites := &fakeSites{rooms: map[string][]plugins.CamRoom{
		"alpha": {room("a1", 900), room("a2", 100)},
		"beta":  {room("b1", 500), room("b2", 400), room("b3", 5)},
	}}
	service := newTestService(t, sites)

	if got := usernames(pageUsernames(t, service, Filter{Limit: 3})); got != "[a1 b1 b2]" {
		t.Errorf("first page = %s", got)
	}
	if got := usernames(pageUsernames(t, service, Filter{Limit: 3, Offset: 3})); got != "[a2 b3]" {
		t.Errorf("second page = %s", got)
	}
	if got := usernames(pageUsernames(t, service, Filter{Sources: []string{"beta"}, Limit: 5})); got != "[b1 b2 b3]" {
		t.Errorf("beta only = %s", got)
	}
}

// TestRoomsAreCachedAndSkipPrivate checks that paging does not hit the site
// again within the cache window and that rooms in a private show are left out.
func TestRoomsAreCachedAndSkipPrivate(t *testing.T) {
	hidden := room("hidden", 1000)
	hidden.Online = false
	sites := &fakeSites{rooms: map[string][]plugins.CamRoom{"alpha": {hidden, room("a1", 10)}}}
	service := newTestService(t, sites)

	if got := usernames(pageUsernames(t, service, Filter{})); got != "[a1]" {
		t.Errorf("rooms = %s", got)
	}
	pageUsernames(t, service, Filter{Offset: 1})
	if sites.listCalls != 1 {
		t.Errorf("site listed %d times, want 1", sites.listCalls)
	}

	now := time.Now().Add(listCacheTTL + time.Second)
	service.now = func() time.Time { return now }
	pageUsernames(t, service, Filter{})
	if sites.listCalls != 2 {
		t.Errorf("after expiry site listed %d times, want 2", sites.listCalls)
	}
}

// TestFollowShowsOnCachedPageAndOffline checks that following survives the
// cache, that the Following row reports offline rooms last, and that streaming
// an offline room fails and marks it offline.
func TestFollowShowsOnCachedPageAndOffline(t *testing.T) {
	sites := &fakeSites{
		rooms:   map[string][]plugins.CamRoom{"alpha": {room("a1", 10), room("a2", 5)}},
		offline: map[string]bool{},
	}
	service := newTestService(t, sites)
	ctx := context.Background()
	page, err := service.Rooms(ctx, Filter{})
	if err != nil {
		t.Fatal(err)
	}
	first, second := page.Rooms[0], page.Rooms[1]
	if _, err := service.SetFollowed(ctx, first.ID, true); err != nil {
		t.Fatal(err)
	}
	if _, err := service.SetFollowed(ctx, second.ID, true); err != nil {
		t.Fatal(err)
	}

	page, _ = service.Rooms(ctx, Filter{})
	if !page.Rooms[0].Followed {
		t.Error("cached page lost the follow")
	}

	sites.offline["a1"] = true
	service.online = map[string]cachedOnline{}
	followed, err := service.Followed(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(followed) != 2 || followed[0].Username != "a2" || !followed[0].Online || followed[1].Online {
		t.Errorf("followed = %+v %+v, want online a2 first", followed[0], followed[1])
	}

	if _, err := service.Stream(ctx, first.ID); err == nil {
		t.Error("streaming an offline room succeeded")
	}
	stream, err := service.Stream(ctx, second.ID)
	if err != nil || stream.URL != "http://backend.test/api/stream/manifest?live=https://edge.example/a2.m3u8" {
		t.Errorf("stream = %+v, %v", stream, err)
	}
}

func TestMergeByViewersKeepsARepeatedRoomOnce(t *testing.T) {
	first := []*model.CamModel{{ID: "cam_model:a", Viewers: 900}, {ID: "cam_model:b", Viewers: 500}}
	// The site moved b ahead between two chunks, so it shows up again later.
	second := []*model.CamModel{{ID: "cam_model:b", Viewers: 400}, {ID: "cam_model:c", Viewers: 300}}

	merged := mergeByViewers([][]*model.CamModel{first, second})
	ids := make([]string, 0, len(merged))
	for _, room := range merged {
		ids = append(ids, room.ID)
	}
	if strings.Join(ids, ",") != "cam_model:a,cam_model:b,cam_model:c" {
		t.Errorf("merged = %v, want each room once in viewer order", ids)
	}
}
