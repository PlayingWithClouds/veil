package cams

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

const (
	// listCacheTTL keeps a fetched listing chunk, so paging through the grid
	// shows one consistent snapshot instead of re-ranking under the user.
	listCacheTTL = 60 * time.Second
	// chunkSize is the page size asked of a plugin; the plugins cap it themselves.
	chunkSize = 100
	// maxWanted bounds how deep a listing may reach into one site.
	maxWanted = 1000
	// defaultPageLimit is the page size when the query names none.
	defaultPageLimit = 48
	// startedDepth is how many of each site's most watched rooms SortStarted
	// ranks: a site lists rooms by popularity only.
	startedDepth = 500
)

// The orders of a room listing.
const (
	// SortViewers lists the most watched rooms first; the default.
	SortViewers = "viewers"
	// SortFewest lists the least watched rooms first, from the end of each site's list.
	SortFewest = "fewest"
	// SortStarted lists the rooms whose show began last first, among the
	// startedDepth most watched of each site.
	SortStarted = "started"
	// SortForYou ranks rooms by the user's taste (see ForYou); List itself
	// lists them as SortViewers.
	SortForYou = "forYou"
)

// Query narrows a room listing.
type Query struct {
	// Plugins lists plugin names; empty means every cam plugin.
	Plugins []string
	Text    string
	Tag     string
	Gender  string
	// Sort is one of the Sort constants; empty is SortViewers.
	Sort   string
	Limit  int
	Offset int
}

// cachedChunk is one listing page of one plugin.
type cachedChunk struct {
	rooms   []schema.CamModel
	total   int
	expires time.Time
}

// pluginListing is what one plugin answered for a listing.
type pluginListing struct {
	rooms []schema.CamModel
	total int
	err   error
}

// List returns one page of the rooms online now across the selected cam
// plugins, stored as CamModels with fresh live state, and the number of rooms
// the sites have in total. Each site's list is put in the query's order chunk
// by chunk (a site's popularity order only roughly follows viewers), and the
// sites are merged in it, so a later page continues the earlier ones.
func (s *Service) List(ctx context.Context, query Query) ([]schema.CamModel, int, error) {
	limit := query.Limit
	if limit <= 0 {
		limit = defaultPageLimit
	}
	wanted := min(query.Offset+limit, maxWanted)
	selected := s.selectPlugins(query.Plugins)
	if len(selected) == 0 {
		return []schema.CamModel{}, 0, nil
	}
	listings := s.listAll(ctx, selected, query, wanted)
	lists, total, err := collectListings(selected, listings)
	if err != nil {
		return nil, 0, err
	}
	unblocked, err := s.withoutBlocked(ctx, mergeBy(lists, rankOf(query.Sort)))
	if err != nil {
		return nil, 0, err
	}
	return pageOf(unblocked, query.Offset, limit), total, nil
}

// blockedRoomsQuery selects the ids of the rooms the user blocked.
const blockedRoomsQuery = `
	SELECT "subjectId" FROM "UserState"
	WHERE "subjectKind" = $camModel AND "blocked" = 1 AND "deletedAt" IS NULL`

// withoutBlocked leaves out the rooms the user blocked.
func (s *Service) withoutBlocked(ctx context.Context, rooms []schema.CamModel) ([]schema.CamModel, error) {
	blockedIDs, err := s.services.Database.Strings(ctx, blockedRoomsQuery, db.Vars{"camModel": schema.RecordKindCamModel})
	if err != nil || len(blockedIDs) == 0 {
		return rooms, err
	}
	return slices.DeleteFunc(rooms, func(camModel schema.CamModel) bool {
		return slices.Contains(blockedIDs, camModel.Id)
	}), nil
}

// rankOf is the order of a sort: whether the first room goes before the second.
func rankOf(sort string) func(first, second schema.CamModel) bool {
	switch sort {
	case SortFewest:
		return fewerViewers
	case SortStarted:
		return laterStart
	}
	return moreViewers
}

// moreViewers orders the bigger audience first.
func moreViewers(first, second schema.CamModel) bool {
	return viewersOf(first) > viewersOf(second)
}

// fewerViewers orders the smaller audience first.
func fewerViewers(first, second schema.CamModel) bool {
	return viewersOf(first) < viewersOf(second)
}

// laterStart orders the show that began last first.
func laterStart(first, second schema.CamModel) bool {
	return liveSince(first) > liveSince(second)
}

// sortedRooms returns the rooms in rank order, keeping the order of ties.
func sortedRooms(rooms []schema.CamModel, before func(first, second schema.CamModel) bool) []schema.CamModel {
	sorted := slices.Clone(rooms)
	slices.SortStableFunc(sorted, func(first, second schema.CamModel) int {
		if before(first, second) {
			return -1
		}
		if before(second, first) {
			return 1
		}
		return 0
	})
	return sorted
}

// selectPlugins returns the cam plugins named by names, or all when empty.
func (s *Service) selectPlugins(names []string) []*plugins.Plugin {
	all := s.services.Sites.Sources()
	if len(names) == 0 {
		return all
	}
	var selected []*plugins.Plugin
	for _, plugin := range all {
		if containsString(names, plugin.Meta.Name) {
			selected = append(selected, plugin)
		}
	}
	return selected
}

// listAll lists every selected plugin concurrently.
func (s *Service) listAll(ctx context.Context, selected []*plugins.Plugin, query Query, wanted int) []pluginListing {
	listings := make([]pluginListing, len(selected))
	var group sync.WaitGroup
	for index, plugin := range selected {
		group.Add(1)
		go func() {
			defer group.Done()
			rooms, total, err := s.pluginRooms(ctx, plugin, query, wanted)
			listings[index] = pluginListing{rooms: rooms, total: total, err: err}
		}()
	}
	group.Wait()
	return listings
}

// collectListings keeps the listings that worked and adds up their totals.
// A failing plugin contributes nothing; it is an error only when all failed.
func collectListings(selected []*plugins.Plugin, listings []pluginListing) ([][]schema.CamModel, int, error) {
	var lists [][]schema.CamModel
	total := 0
	for index, listing := range listings {
		if listing.err != nil {
			log.Printf("cams: list %s: %v", selected[index].Meta.Name, listing.err)
			continue
		}
		lists = append(lists, listing.rooms)
		total += listing.total
	}
	if lists == nil {
		return nil, 0, errors.New("no cam site answered")
	}
	return lists, total, nil
}

// pageOf cuts one page out of the merged rooms.
func pageOf[Room any](rooms []Room, offset, limit int) []Room {
	if offset >= len(rooms) {
		return []Room{}
	}
	page := rooms[offset:]
	if len(page) > limit {
		page = page[:limit]
	}
	return page
}

// pluginRooms returns the first wanted online rooms of one plugin in the
// query's order, and how many the site has.
func (s *Service) pluginRooms(ctx context.Context, plugin *plugins.Plugin, query Query, wanted int) ([]schema.CamModel, int, error) {
	switch query.Sort {
	case SortFewest:
		return s.roomsFromEnd(ctx, plugin, query, wanted)
	case SortStarted:
		rooms, total, err := s.roomsFromStart(ctx, plugin, query, max(wanted, startedDepth))
		return sortedRooms(rooms, laterStart), total, err
	}
	return s.roomsFromStart(ctx, plugin, query, wanted)
}

// roomsFromStart pages through the site's list from its most popular room
// until it has wanted rooms, each chunk most watched first.
func (s *Service) roomsFromStart(ctx context.Context, plugin *plugins.Plugin, query Query, wanted int) ([]schema.CamModel, int, error) {
	var rooms []schema.CamModel
	total := 0
	for chunkOffset := 0; len(rooms) < wanted; chunkOffset += chunkSize {
		chunk, err := s.chunk(ctx, plugin, query, chunkOffset)
		if err != nil && len(rooms) > 0 {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		total = chunk.total
		rooms = append(rooms, sortedRooms(chunk.rooms, moreViewers)...)
		if len(chunk.rooms) == 0 || chunkOffset+chunkSize >= chunk.total {
			break
		}
	}
	return rooms, total, nil
}

// roomsFromEnd pages backwards from the end of the site's list, where its
// least watched rooms are, until it has wanted rooms, each chunk fewest first.
// The first chunk tells where the list ends.
func (s *Service) roomsFromEnd(ctx context.Context, plugin *plugins.Plugin, query Query, wanted int) ([]schema.CamModel, int, error) {
	first, err := s.chunk(ctx, plugin, query, 0)
	if err != nil {
		return nil, 0, err
	}
	var rooms []schema.CamModel
	for chunkOffset := lastChunkOffset(first.total); chunkOffset >= 0 && len(rooms) < wanted; chunkOffset -= chunkSize {
		chunk, err := s.chunk(ctx, plugin, query, chunkOffset)
		if err != nil && len(rooms) > 0 {
			break
		}
		if err != nil {
			return nil, 0, err
		}
		rooms = append(rooms, sortedRooms(chunk.rooms, fewerViewers)...)
	}
	return rooms, first.total, nil
}

// lastChunkOffset is the offset of the chunk holding the last of total rooms.
func lastChunkOffset(total int) int {
	if total <= 0 {
		return 0
	}
	return (total - 1) / chunkSize * chunkSize
}

// chunk returns one cached page of a plugin's listing, fetching and storing the
// rooms on a miss.
func (s *Service) chunk(ctx context.Context, plugin *plugins.Plugin, query Query, chunkOffset int) (cachedChunk, error) {
	key := fmt.Sprintf("%s|%s|%s|%s|%d", plugin.Meta.Name, query.Text, query.Tag, query.Gender, chunkOffset)
	if cached, ok := s.cachedChunk(key); ok {
		return cached, nil
	}
	result, err := s.services.Sites.List(ctx, plugin, plugins.CamListArgs{
		Query: query.Text, Tag: query.Tag, Gender: query.Gender, Limit: chunkSize, Offset: chunkOffset,
	})
	if err != nil {
		return cachedChunk{}, err
	}
	total := result.Total
	if total == 0 {
		total = chunkOffset + len(result.Rooms)
	}
	chunk := cachedChunk{rooms: s.storeRooms(ctx, plugin.Meta.Name, result.Rooms), total: total, expires: s.now().Add(listCacheTTL)}
	s.storeChunk(key, chunk)
	return chunk, nil
}

// cachedChunk returns an unexpired cached page.
func (s *Service) cachedChunk(key string) (cachedChunk, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	cached, ok := s.chunks[key]
	if !ok || !s.now().Before(cached.expires) {
		return cachedChunk{}, false
	}
	return cached, true
}

// storeChunk caches a page and drops the expired ones.
func (s *Service) storeChunk(key string, chunk cachedChunk) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	for oldKey, old := range s.chunks {
		if !s.now().Before(old.expires) {
			delete(s.chunks, oldKey)
		}
	}
	s.chunks[key] = chunk
}

// mergeBy merges lists that are each in rank order into one list, always
// taking the head that ranks first. A room listed twice (sites reorder rooms
// between the chunks of one listing, and chunks are cached at different
// times) keeps only its first, best place.
func mergeBy(lists [][]schema.CamModel, before func(first, second schema.CamModel) bool) []schema.CamModel {
	positions := make([]int, len(lists))
	seen := map[string]bool{}
	merged := []schema.CamModel{}
	for {
		best := bestHead(lists, positions, before)
		if best == -1 {
			return merged
		}
		head := lists[best][positions[best]]
		positions[best]++
		if seen[head.Id] {
			continue
		}
		seen[head.Id] = true
		merged = append(merged, head)
	}
}

// bestHead returns the index of the list whose next room ranks first, or -1
// when every list is used up.
func bestHead(lists [][]schema.CamModel, positions []int, before func(first, second schema.CamModel) bool) int {
	best := -1
	for index, list := range lists {
		if positions[index] >= len(list) {
			continue
		}
		if best == -1 || before(list[positions[index]], lists[best][positions[best]]) {
			best = index
		}
	}
	return best
}

// viewersOf is the room's audience, 0 when offline.
func viewersOf(camModel schema.CamModel) int32 {
	if camModel.Live == nil {
		return 0
	}
	return camModel.Live.Viewers
}

// storeRooms upserts the rooms a listing showed as online now and returns
// them. Rooms in a private show are not watchable and are skipped.
func (s *Service) storeRooms(ctx context.Context, pluginName string, rooms []plugins.CamRoom) []schema.CamModel {
	stored := make([]schema.CamModel, 0, len(rooms))
	for _, room := range rooms {
		if !room.Online || room.Username == "" {
			continue
		}
		camModel, err := s.upsertRoom(ctx, pluginName, room)
		if err != nil {
			log.Printf("cams: store %s/%s: %v", pluginName, room.Username, err)
			continue
		}
		stored = append(stored, *camModel)
	}
	return stored
}

// upsertRoomQuery inserts a room or refreshes the one with the same plugin
// and username. The show starts when the site says, else a room already live
// keeps the start it has.
const upsertRoomQuery = `
	INSERT INTO "CamModel" ("id", "plugin", "username", "displayName", "thumbnail", "tags", "gender", "live", "lastSeenOnlineAt", "autoRecord")
	VALUES ($id, $plugin, $username, $displayName, $thumbnail, $tags, $gender, $live, $now, 0)
	ON CONFLICT ("plugin", "username") DO UPDATE SET
		"displayName" = excluded."displayName",
		"thumbnail" = COALESCE(excluded."thumbnail", "CamModel"."thumbnail"),
		"tags" = excluded."tags",
		"gender" = COALESCE(excluded."gender", "CamModel"."gender"),
		"live" = json_set(excluded."live", '$.startedAt', CASE WHEN $siteStarted
			THEN json_extract(excluded."live", '$.startedAt')
			ELSE COALESCE(json_extract("CamModel"."live", '$.startedAt'), json_extract(excluded."live", '$.startedAt')) END),
		"lastSeenOnlineAt" = excluded."lastSeenOnlineAt",
		"updatedAt" = excluded."lastSeenOnlineAt"
	RETURNING *`

// upsertRoom stores a room a listing showed online, with its live state.
func (s *Service) upsertRoom(ctx context.Context, pluginName string, room plugins.CamRoom) (*schema.CamModel, error) {
	now := db.FormatTime(s.now())
	startedAt, siteStarted := siteStart(room)
	if !siteStarted {
		startedAt = now
	}
	thumbnail := assetOf(room.ThumbnailURL)
	camModel, err := db.QueryOneAs[schema.CamModel](ctx, s.services.Database, upsertRoomQuery, db.Vars{
		"id":          db.NewID(),
		"plugin":      pluginName,
		"username":    room.Username,
		"displayName": displayNameOf(room),
		"thumbnail":   thumbnail,
		"tags":        room.Tags,
		"gender":      optionalText(room.Gender),
		"live":        schema.CamLive{Viewers: int32(room.Viewers), Subject: optionalText(room.Subject), StartedAt: &startedAt, Thumbnail: thumbnail},
		"siteStarted": siteStarted,
		"now":         now,
	})
	if err != nil {
		return nil, err
	}
	if camModel == nil {
		return nil, errors.New("no row returned")
	}
	return camModel, nil
}

// siteStart is the show start the site gave, in the stored time format, and
// whether it gave a readable one.
func siteStart(room plugins.CamRoom) (string, bool) {
	if room.StartedAt == "" {
		return "", false
	}
	parsed, err := time.Parse(time.RFC3339Nano, room.StartedAt)
	if err != nil {
		return "", false
	}
	return db.FormatTime(parsed), true
}

// displayNameOf is the room's shown name, its username when the site gives none.
func displayNameOf(room plugins.CamRoom) string {
	name := strings.TrimSpace(room.DisplayName)
	if name == "" {
		return room.Username
	}
	return name
}

// assetOf wraps a URL as an asset; nil for a blank one, so it binds as NULL.
func assetOf(uri string) *schema.Asset {
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return nil
	}
	return &schema.Asset{Uri: uri}
}

// optionalText returns nil for a blank string, so it binds as NULL.
func optionalText(text string) *string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	return &text
}

// containsString reports whether list holds value.
func containsString(list []string, value string) bool {
	for _, candidate := range list {
		if candidate == value {
			return true
		}
	}
	return false
}
