package cams

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

const (
	// onlineCacheTTL is how long a liveness check of one room stays valid.
	onlineCacheTTL = 45 * time.Second
	// defaultModelLimit is the page size of Models when the filter names none.
	defaultModelLimit = 100
)

// cachedOnline is one liveness answer.
type cachedOnline struct {
	online  bool
	expires time.Time
}

// ModelFilter narrows the stored cam models.
type ModelFilter struct {
	// Live keeps rooms that are (true) or are not (false) live; nil keeps both.
	Live *bool
	// AutoRecord keeps rooms with auto-record on (true) or off (false).
	AutoRecord *bool
	// PerformerID keeps rooms linked to the performer.
	PerformerID string
	Limit       int
	Offset      int
}

// Models returns the stored cam models matching filter, live ones first,
// then by name. Their live state is what listings and the watcher saw last.
func (s *Service) Models(ctx context.Context, filter ModelFilter) ([]schema.CamModel, error) {
	conditions := []string{`"deletedAt" IS NULL`}
	vars := db.Vars{}
	if filter.Live != nil {
		conditions = append(conditions, liveCondition(*filter.Live))
	}
	if filter.AutoRecord != nil {
		conditions = append(conditions, `"autoRecord" = $autoRecord`)
		vars["autoRecord"] = *filter.AutoRecord
	}
	if filter.PerformerID != "" {
		conditions = append(conditions, `"performer" = $performer`)
		vars["performer"] = filter.PerformerID
	}
	limit := filter.Limit
	if limit <= 0 {
		limit = defaultModelLimit
	}
	vars["limit"] = limit
	vars["offset"] = max(filter.Offset, 0)
	return db.QueryAs[schema.CamModel](ctx, s.services.Database,
		`SELECT * FROM "CamModel" WHERE `+strings.Join(conditions, " AND ")+
			` ORDER BY "live" IS NULL, "displayName" COLLATE NOCASE LIMIT $limit OFFSET $offset`, vars)
}

// liveCondition is the condition keeping live (or offline) rooms.
func liveCondition(live bool) string {
	if live {
		return `"live" IS NOT NULL`
	}
	return `"live" IS NULL`
}

// Model returns one cam model with its live state checked against the site.
func (s *Service) Model(ctx context.Context, id string) (*schema.CamModel, error) {
	camModel, err := s.load(ctx, id)
	if err != nil {
		return nil, err
	}
	s.checkOnline(ctx, camModel)
	return s.load(ctx, id)
}

// load returns the stored cam model, ErrNotFound when there is none.
func (s *Service) load(ctx context.Context, id string) (*schema.CamModel, error) {
	camModel, err := db.Get[schema.CamModel](ctx, s.services.Database, camModelTable, id)
	if err != nil {
		return nil, err
	}
	if camModel == nil {
		return nil, ErrNotFound
	}
	return camModel, nil
}

// SetAutoRecord turns automatic recording on or off for a model and, when
// quality is not empty, sets the quality it records at (see ParseQuality).
func (s *Service) SetAutoRecord(ctx context.Context, id string, on bool, quality string) (*schema.CamModel, error) {
	if _, err := s.load(ctx, id); err != nil {
		return nil, err
	}
	fields := db.Vars{"autoRecord": on}
	if strings.TrimSpace(quality) != "" {
		normalized, err := NormalizeQuality(quality)
		if err != nil {
			return nil, err
		}
		fields["recordQuality"] = normalized
	}
	if err := s.services.Database.Update(ctx, camModelTable, id, fields); err != nil {
		return nil, err
	}
	return s.load(ctx, id)
}

// LinkPerformer marks the model as the performer in the library, so her
// recordings credit the performer; an empty performerID unlinks.
func (s *Service) LinkPerformer(ctx context.Context, id, performerID string) (*schema.CamModel, error) {
	if _, err := s.load(ctx, id); err != nil {
		return nil, err
	}
	var performer any
	if performerID != "" {
		if err := s.requirePerformer(ctx, performerID); err != nil {
			return nil, err
		}
		performer = performerID
	}
	if err := s.services.Database.Update(ctx, camModelTable, id, db.Vars{"performer": performer}); err != nil {
		return nil, err
	}
	return s.load(ctx, id)
}

// requirePerformer fails unless the performer exists.
func (s *Service) requirePerformer(ctx context.Context, performerID string) error {
	performer, err := db.Get[schema.Performer](ctx, s.services.Database, "Performer", performerID)
	if err != nil {
		return err
	}
	if performer == nil {
		return fmt.Errorf("performer %s not found", performerID)
	}
	return nil
}

// Watch resolves a room to a live stream and returns the proxy URL the player
// loads and its mime type. A room that cannot be resolved is stored as offline.
func (s *Service) Watch(ctx context.Context, id string) (playableURL, mimeType string, err error) {
	camModel, err := s.load(ctx, id)
	if err != nil {
		return "", "", err
	}
	plugin, ok := s.services.Sites.Plugin(camModel.Plugin)
	if !ok {
		return "", "", fmt.Errorf("cam plugin %q is not available", camModel.Plugin)
	}
	result, err := s.services.Sites.Stream(ctx, plugin, camModel.Username)
	if err != nil {
		s.recordOnline(ctx, id, false)
		return "", "", fmt.Errorf("%w: %s: %v", ErrOffline, camModel.DisplayName, err)
	}
	s.recordOnline(ctx, id, true)
	playableURL, err = s.services.Proxy.LiveManifestURL(ctx, result, id)
	if err != nil {
		return "", "", err
	}
	mimeType = strings.TrimSpace(result.MimeType)
	if mimeType == "" {
		mimeType = "application/x-mpegURL"
	}
	return playableURL, mimeType, nil
}

// checkOnline asks the room's site whether the room can be watched right now
// (it can if a stream resolves) and stores the answer. Answers are reused for
// onlineCacheTTL. A missing plugin counts as offline.
func (s *Service) checkOnline(ctx context.Context, camModel *schema.CamModel) bool {
	if cached, ok := s.cachedOnline(camModel.Id); ok {
		return cached
	}
	plugin, ok := s.services.Sites.Plugin(camModel.Plugin)
	if !ok {
		return false
	}
	_, err := s.services.Sites.Stream(ctx, plugin, camModel.Username)
	online := err == nil
	s.recordOnline(ctx, camModel.Id, online)
	return online
}

// cachedOnline returns an unexpired liveness answer.
func (s *Service) cachedOnline(id string) (online bool, found bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	cached, ok := s.online[id]
	if !ok || !s.now().Before(cached.expires) {
		return false, false
	}
	return cached.online, true
}

// markOnlineQuery keeps a live room's state and marks an offline one live as
// of now.
const markOnlineQuery = `UPDATE "CamModel" SET
	"live" = COALESCE("live", json_object('viewers', 0, 'startedAt', $now)),
	"lastSeenOnlineAt" = $now, "updatedAt" = $now
	WHERE "id" = $id`

// recordOnline caches a liveness answer and stores it on the row; going
// offline stores the show that ended.
func (s *Service) recordOnline(ctx context.Context, id string, online bool) {
	s.mutex.Lock()
	s.online[id] = cachedOnline{online: online, expires: s.now().Add(onlineCacheTTL)}
	s.mutex.Unlock()

	var err error
	if online {
		_, err = s.services.Database.Exec(ctx, markOnlineQuery, db.Vars{"id": id, "now": db.FormatTime(s.now())})
	} else {
		err = s.takeOffline(ctx, `"id" = $id`, db.Vars{"id": id})
	}
	if err != nil {
		log.Printf("cams: record online %s: %v", id, err)
	}
}
