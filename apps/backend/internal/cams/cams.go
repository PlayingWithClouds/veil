// Package cams serves live webcam rooms: it lists the rooms cam plugins have
// online, remembers every room it has seen (so they can be followed and shown
// offline), and resolves a room to a live stream through the HLS proxy.
package cams

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

const (
	// listCacheTTL keeps a fetched listing page, so paging through the grid
	// shows one consistent snapshot instead of re-ranking under the user.
	listCacheTTL = 60 * time.Second
	// onlineCacheTTL is how long a liveness check of one room stays valid.
	onlineCacheTTL = 45 * time.Second
	// chunkSize is the page size asked of a plugin; the plugins cap it themselves.
	chunkSize = 100
	// maxWanted bounds how deep a listing may reach into one site.
	maxWanted = 1000
	// onlineCheckWorkers bounds the concurrent liveness checks of the Following row.
	onlineCheckWorkers = 6
	defaultPageLimit   = 48
)

// ErrNotFound is returned for a cam model id that does not exist.
var ErrNotFound = errors.New("cam model not found")

// Sites is what the service needs from the plugin layer.
type Sites interface {
	// Sources returns the active plugins that list cam rooms.
	Sources() []*plugins.Plugin
	// Plugin returns one active cam plugin by name.
	Plugin(name string) (*plugins.Plugin, bool)
	List(ctx context.Context, plugin *plugins.Plugin, args plugins.CamListArgs) (*plugins.CamListResult, error)
	Stream(ctx context.Context, plugin *plugins.Plugin, username string) (*plugins.ResolveResult, error)
}

// LiveProxy turns a resolved live stream into a playable local URL.
type LiveProxy interface {
	LiveManifestURL(ctx context.Context, result *plugins.ResolveResult) (string, error)
}

// Filter narrows a room listing.
type Filter struct {
	// Sources lists plugin names; empty means every cam plugin.
	Sources []string
	Tag     string
	Query   string
	Gender  string
	Limit   int
	Offset  int
}

// Service lists, stores and streams cam rooms.
type Service struct {
	database *db.DB
	sites    Sites
	proxy    LiveProxy
	now      func() time.Time

	mutex  sync.Mutex
	chunks map[string]cachedChunk
	online map[string]cachedOnline
}

type cachedChunk struct {
	rooms   []*model.CamModel
	total   int
	expires time.Time
}

type cachedOnline struct {
	online  bool
	expires time.Time
}

// New returns a Service over the database, cam plugins and live proxy.
func New(database *db.DB, sites Sites, proxy LiveProxy) *Service {
	return &Service{
		database: database,
		sites:    sites,
		proxy:    proxy,
		now:      time.Now,
		chunks:   map[string]cachedChunk{},
		online:   map[string]cachedOnline{},
	}
}

// Sources returns the cam plugins, for the site filter.
func (s *Service) Sources() []*plugins.Plugin {
	return s.sites.Sources()
}

// Rooms returns one page of the rooms online now across the selected cam
// plugins. Each plugin's own popularity order is kept and the
// plugins are merged by viewers, so a later page continues the earlier ones.
func (s *Service) Rooms(ctx context.Context, filter Filter) (*model.CamRoomPage, error) {
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultPageLimit
	}
	wanted := filter.Offset + limit
	if wanted > maxWanted {
		wanted = maxWanted
	}
	selected := s.selectSources(filter.Sources)
	if len(selected) == 0 {
		return &model.CamRoomPage{Rooms: []*model.CamModel{}}, nil
	}

	lists, total, failures := s.fetchAll(ctx, selected, filter, wanted)
	if failures == len(selected) {
		return nil, fmt.Errorf("no cam site answered")
	}
	merged := mergeByViewers(lists)
	s.markFollowed(ctx, merged)

	page := []*model.CamModel{}
	if filter.Offset < len(merged) {
		page = merged[filter.Offset:]
	}
	if len(page) > limit {
		page = page[:limit]
	}
	return &model.CamRoomPage{Rooms: page, TotalCount: total}, nil
}

// selectSources returns the cam plugins named by names, or all when empty.
func (s *Service) selectSources(names []string) []*plugins.Plugin {
	all := s.sites.Sources()
	if len(names) == 0 {
		return all
	}
	allowed := map[string]bool{}
	for _, name := range names {
		allowed[name] = true
	}
	var selected []*plugins.Plugin
	for _, plugin := range all {
		if allowed[plugin.Meta.Name] {
			selected = append(selected, plugin)
		}
	}
	return selected
}

// fetchAll lists every selected plugin concurrently. A failing plugin
// contributes nothing and is counted in failures.
func (s *Service) fetchAll(ctx context.Context, selected []*plugins.Plugin, filter Filter, wanted int) (lists [][]*model.CamModel, total int, failures int) {
	type outcome struct {
		rooms []*model.CamModel
		total int
		err   error
	}
	outcomes := make([]outcome, len(selected))
	var group sync.WaitGroup
	for index, plugin := range selected {
		group.Add(1)
		go func(index int, plugin *plugins.Plugin) {
			defer group.Done()
			rooms, pluginTotal, err := s.pluginRooms(ctx, plugin, filter, wanted)
			outcomes[index] = outcome{rooms: rooms, total: pluginTotal, err: err}
		}(index, plugin)
	}
	group.Wait()

	for index, result := range outcomes {
		if result.err != nil {
			log.Printf("cams: list %s: %v", selected[index].Meta.Name, result.err)
			failures++
			continue
		}
		lists = append(lists, result.rooms)
		total += result.total
	}
	return lists, total, failures
}

// pluginRooms returns the first wanted online rooms of one plugin, paging
// through the plugin in chunks.
func (s *Service) pluginRooms(ctx context.Context, plugin *plugins.Plugin, filter Filter, wanted int) ([]*model.CamModel, int, error) {
	var rooms []*model.CamModel
	total := 0
	for chunkOffset := 0; len(rooms) < wanted; chunkOffset += chunkSize {
		chunk, chunkTotal, err := s.chunk(ctx, plugin, filter, chunkOffset)
		if err != nil {
			if len(rooms) > 0 {
				break
			}
			return nil, 0, err
		}
		total = chunkTotal
		rooms = append(rooms, chunk...)
		if len(chunk) == 0 || chunkOffset+chunkSize >= chunkTotal {
			break
		}
	}
	return rooms, total, nil
}

// chunk returns one cached page of a plugin's listing, fetching and storing the
// rooms on a miss.
func (s *Service) chunk(ctx context.Context, plugin *plugins.Plugin, filter Filter, chunkOffset int) ([]*model.CamModel, int, error) {
	key := fmt.Sprintf("%s|%s|%s|%s|%d", plugin.Meta.Name, filter.Query, filter.Tag, filter.Gender, chunkOffset)
	if cached, ok := s.cachedChunk(key); ok {
		return cached.rooms, cached.total, nil
	}
	result, err := s.sites.List(ctx, plugin, plugins.CamListArgs{
		Query: filter.Query, Tag: filter.Tag, Gender: filter.Gender, Limit: chunkSize, Offset: chunkOffset,
	})
	if err != nil {
		return nil, 0, err
	}
	rooms := s.storeRooms(ctx, plugin.Meta.Name, result.Rooms)
	total := result.Total
	if total == 0 {
		total = chunkOffset + len(result.Rooms)
	}
	s.storeChunk(key, cachedChunk{rooms: rooms, total: total, expires: s.now().Add(listCacheTTL)})
	return rooms, total, nil
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

// mergeByViewers merges lists that are each ordered best first into one list,
// always taking the head with the most viewers. A room listed twice (sites
// reorder rooms between the chunks of one listing, and chunks are cached at
// different times) keeps only its first, best place.
func mergeByViewers(lists [][]*model.CamModel) []*model.CamModel {
	positions := make([]int, len(lists))
	seen := map[string]bool{}
	var merged []*model.CamModel
	for {
		best := -1
		for index, list := range lists {
			if positions[index] >= len(list) {
				continue
			}
			if best == -1 || list[positions[index]].Viewers > lists[best][positions[best]].Viewers {
				best = index
			}
		}
		if best == -1 {
			return merged
		}
		head := lists[best][positions[best]]
		positions[best]++
		if seen[head.ID] {
			continue
		}
		seen[head.ID] = true
		// Copy: cached pages are shared between requests, and the caller sets flags.
		room := *head
		merged = append(merged, &room)
	}
}

// markFollowed sets the followed flag on rooms from the database, since a
// cached page may predate a follow.
func (s *Service) markFollowed(ctx context.Context, rooms []*model.CamModel) {
	followedIDs, err := s.database.Strings(ctx, `SELECT id FROM cam_model WHERE followed = 1`, nil)
	if err != nil {
		return
	}
	followed := map[string]bool{}
	for _, id := range followedIDs {
		followed[id] = true
	}
	for _, room := range rooms {
		room.Followed = followed[room.ID]
	}
}

// storeRooms upserts the rooms a listing showed as online now and returns them
// as models. Rooms in a private show are not watchable and are skipped.
func (s *Service) storeRooms(ctx context.Context, pluginName string, rooms []plugins.CamRoom) []*model.CamModel {
	models := make([]*model.CamModel, 0, len(rooms))
	for _, room := range rooms {
		if !room.Online || room.Username == "" {
			continue
		}
		stored, err := s.upsertRoom(ctx, pluginName, room)
		if err != nil {
			log.Printf("cams: store %s/%s: %v", pluginName, room.Username, err)
			continue
		}
		models = append(models, stored)
	}
	return models
}

// camRecord mirrors the cam_model table.
type camRecord struct {
	ID               string   `json:"id"`
	Plugin           string   `json:"plugin"`
	Username         string   `json:"username"`
	DisplayName      string   `json:"display_name"`
	SourceURL        string   `json:"source_url"`
	ThumbnailURL     *string  `json:"thumbnail_url"`
	Tags             []string `json:"tags"`
	Gender           *string  `json:"gender"`
	Subject          *string  `json:"subject"`
	Viewers          int      `json:"viewers"`
	Online           bool     `json:"online"`
	LastSeenOnlineAt *string  `json:"last_seen_online_at"`
	Followed         bool     `json:"followed"`
	AutoRecord       bool     `json:"auto_record"`
	RecordQuality    string   `json:"record_quality"`
}

// toModel maps a stored row to the GraphQL model.
func (record camRecord) toModel() *model.CamModel {
	tags := record.Tags
	if tags == nil {
		tags = []string{}
	}
	return &model.CamModel{
		ID:               record.ID,
		PluginName:       record.Plugin,
		Username:         record.Username,
		DisplayName:      record.DisplayName,
		SourceURL:        record.SourceURL,
		ThumbnailURL:     record.ThumbnailURL,
		Tags:             tags,
		Gender:           record.Gender,
		Subject:          record.Subject,
		Viewers:          record.Viewers,
		Online:           record.Online,
		LastSeenOnlineAt: record.LastSeenOnlineAt,
		Followed:         record.Followed,
		AutoRecord:       record.AutoRecord,
		RecordQuality:    record.RecordQuality,
	}
}

// upsertRoom inserts a room or refreshes the one with the same source URL,
// marking it online as of now.
func (s *Service) upsertRoom(ctx context.Context, pluginName string, room plugins.CamRoom) (*model.CamModel, error) {
	displayName := room.DisplayName
	if displayName == "" {
		displayName = room.Username
	}
	tags := room.Tags
	if tags == nil {
		tags = []string{}
	}
	record, err := db.QueryOneAs[camRecord](ctx, s.database, `
		INSERT INTO cam_model (id, plugin, username, display_name, source_url, thumbnail_url, tags, gender, subject, viewers, online, last_seen_online_at)
		VALUES ($id, $plugin, $username, $displayName, $sourceUrl, $thumbnail, $tags, $gender, $subject, $viewers, 1, $now)
		ON CONFLICT (source_url) DO UPDATE SET
		  plugin = excluded.plugin, username = excluded.username, display_name = excluded.display_name,
		  thumbnail_url = excluded.thumbnail_url, tags = excluded.tags, gender = excluded.gender,
		  subject = excluded.subject, viewers = excluded.viewers, online = 1,
		  last_seen_online_at = excluded.last_seen_online_at, updated_at = excluded.last_seen_online_at
		RETURNING *`,
		db.Vars{
			"id": db.NewRecordID("cam_model"), "plugin": pluginName, "username": room.Username,
			"displayName": displayName, "sourceUrl": room.SourceURL, "thumbnail": nullIfEmpty(room.ThumbnailURL),
			"tags": tags, "gender": nullIfEmpty(room.Gender), "subject": nullIfEmpty(room.Subject),
			"viewers": room.Viewers, "now": s.now().UTC().Format("2006-01-02T15:04:05.000Z"),
		})
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, fmt.Errorf("no row returned")
	}
	return record.toModel(), nil
}

// nullIfEmpty binds an empty string as NULL.
func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

// load returns the stored record for a cam model id.
func (s *Service) load(ctx context.Context, id string) (*camRecord, error) {
	record, err := db.QueryOneAs[camRecord](ctx, s.database, `SELECT * FROM cam_model WHERE id = $id`, db.Vars{"id": id})
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrNotFound
	}
	return record, nil
}

// Model returns one cam model with its online status checked against the site.
func (s *Service) Model(ctx context.Context, id string) (*model.CamModel, error) {
	record, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	record.Online = s.checkOnline(ctx, record)
	return record.toModel(), nil
}

// Followed returns the followed models, online ones first, each checked
// against its site.
func (s *Service) Followed(ctx context.Context) ([]*model.CamModel, error) {
	records, err := db.QueryAs[camRecord](ctx, s.database,
		`SELECT * FROM cam_model WHERE followed = 1 ORDER BY display_name COLLATE NOCASE`, nil)
	if err != nil {
		return nil, err
	}
	var group sync.WaitGroup
	workers := make(chan struct{}, onlineCheckWorkers)
	for index := range records {
		group.Add(1)
		workers <- struct{}{}
		go func(record *camRecord) {
			defer group.Done()
			defer func() { <-workers }()
			record.Online = s.checkOnline(ctx, record)
		}(&records[index])
	}
	group.Wait()

	sort.SliceStable(records, func(first, second int) bool {
		return records[first].Online && !records[second].Online
	})
	models := make([]*model.CamModel, 0, len(records))
	for _, record := range records {
		models = append(models, record.toModel())
	}
	return models, nil
}

// checkOnline asks the room's site whether the room can be watched right now
// (it can if a stream resolves) and records the answer. Answers are reused for
// onlineCacheTTL. A missing plugin counts as offline.
func (s *Service) checkOnline(ctx context.Context, record *camRecord) bool {
	if cached, ok := s.cachedOnline(record.ID); ok {
		return cached
	}
	plugin, ok := s.sites.Plugin(record.Plugin)
	if !ok {
		return false
	}
	_, err := s.sites.Stream(ctx, plugin, record.Username)
	online := err == nil
	s.recordOnline(ctx, record.ID, online)
	return online
}

// cachedOnline returns an unexpired liveness answer.
func (s *Service) cachedOnline(id string) (bool, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	cached, ok := s.online[id]
	if !ok || !s.now().Before(cached.expires) {
		return false, false
	}
	return cached.online, true
}

// recordOnline caches a liveness answer and stores it on the row.
func (s *Service) recordOnline(ctx context.Context, id string, online bool) {
	s.mutex.Lock()
	s.online[id] = cachedOnline{online: online, expires: s.now().Add(onlineCacheTTL)}
	s.mutex.Unlock()

	query := `UPDATE cam_model SET online = 0 WHERE id = $id`
	if online {
		query = `UPDATE cam_model SET online = 1, last_seen_online_at = $now WHERE id = $id`
	}
	_, err := s.database.Exec(ctx, query, db.Vars{"id": id, "now": s.now().UTC().Format("2006-01-02T15:04:05.000Z")})
	if err != nil {
		log.Printf("cams: record online %s: %v", id, err)
	}
}

// SetFollowed follows or unfollows a model and returns it. Unfollowing also
// turns auto-record off, since only followed models are watched.
func (s *Service) SetFollowed(ctx context.Context, id string, followed bool) (*model.CamModel, error) {
	record, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if _, err := s.database.Exec(ctx, `UPDATE cam_model SET followed = $followed, auto_record = auto_record AND $followed WHERE id = $id`,
		db.Vars{"id": id, "followed": followed}); err != nil {
		return nil, err
	}
	record.Followed = followed
	record.AutoRecord = record.AutoRecord && followed
	return record.toModel(), nil
}

// SetAutoRecord turns automatic recording on or off for a model and, when
// quality is given, sets the quality it records at (see ParseQuality). Turning
// it on also follows the model.
func (s *Service) SetAutoRecord(ctx context.Context, id string, enabled bool, quality *string) (*model.CamModel, error) {
	record, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	if quality != nil {
		normalized, err := NormalizeQuality(*quality)
		if err != nil {
			return nil, err
		}
		record.RecordQuality = normalized
	}
	if _, err := s.database.Exec(ctx, `UPDATE cam_model SET auto_record = $enabled, followed = followed OR $enabled, record_quality = $quality WHERE id = $id`,
		db.Vars{"id": id, "enabled": enabled, "quality": record.RecordQuality}); err != nil {
		return nil, err
	}
	record.AutoRecord = enabled
	record.Followed = record.Followed || enabled
	return record.toModel(), nil
}

// Stream resolves a room to a live stream and returns the proxy URL the player
// loads. A room that cannot be resolved is recorded as offline.
func (s *Service) Stream(ctx context.Context, id string) (*model.CamStream, error) {
	record, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	plugin, ok := s.sites.Plugin(record.Plugin)
	if !ok {
		return nil, fmt.Errorf("cam plugin %q is not available", record.Plugin)
	}
	result, err := s.sites.Stream(ctx, plugin, record.Username)
	if err != nil {
		s.recordOnline(ctx, id, false)
		return nil, fmt.Errorf("%s is offline: %w", record.DisplayName, err)
	}
	s.recordOnline(ctx, id, true)
	playableURL, err := s.proxy.LiveManifestURL(ctx, result)
	if err != nil {
		return nil, err
	}
	mimeType := result.MimeType
	if strings.TrimSpace(mimeType) == "" {
		mimeType = "application/x-mpegURL"
	}
	return &model.CamStream{URL: playableURL, MimeType: mimeType}, nil
}
