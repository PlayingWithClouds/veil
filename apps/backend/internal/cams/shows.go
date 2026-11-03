package cams

import (
	"context"
	"log"
	"maps"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// showTable is the table finished shows are stored in.
const showTable = "CamShow"

// showHistory is how far back Shows reads, enough for a room's usual week.
const showHistory = 8 * 7 * 24 * time.Hour

// takeOffline stores the show of every live room matching where (a condition
// on "CamModel" over vars), then clears their live state.
func (s *Service) takeOffline(ctx context.Context, where string, vars db.Vars) error {
	ending, err := db.QueryAs[schema.CamModel](ctx, s.services.Database,
		`SELECT * FROM "CamModel" WHERE "live" IS NOT NULL AND (`+where+`)`, vars)
	if err != nil {
		return err
	}
	for _, camModel := range ending {
		s.storeShow(ctx, camModel)
	}
	update := db.Vars{"now": db.FormatTime(s.now())}
	maps.Copy(update, vars)
	_, err = s.services.Database.Exec(ctx,
		`UPDATE "CamModel" SET "live" = NULL, "updatedAt" = $now WHERE "live" IS NOT NULL AND (`+where+`)`, update)
	return err
}

// storeShow stores the show a room is ending: from its start to when it was
// last seen online. A room without a known start has no show to store.
func (s *Service) storeShow(ctx context.Context, camModel schema.CamModel) {
	startedAt := liveSince(camModel)
	if startedAt == "" {
		return
	}
	endedAt := db.FormatTime(s.now())
	if camModel.LastSeenOnlineAt != nil {
		endedAt = *camModel.LastSeenOnlineAt
	}
	if endedAt < startedAt {
		return
	}
	show := schema.CamShow{CamModel: camModel.Id, StartedAt: startedAt, EndedAt: endedAt}
	if _, err := s.services.Database.Insert(ctx, showTable, show); err != nil {
		log.Printf("cams: store show of %s: %v", camModel.DisplayName, err)
	}
}

// Shows returns a room's stored shows of the last weeks, newest first.
func (s *Service) Shows(ctx context.Context, camModelID string) ([]schema.CamShow, error) {
	return db.QueryAs[schema.CamShow](ctx, s.services.Database,
		`SELECT * FROM "CamShow" WHERE "camModel" = $id AND "deletedAt" IS NULL AND "startedAt" >= $since
		 ORDER BY "startedAt" DESC`,
		db.Vars{"id": camModelID, "since": db.FormatTime(s.now().Add(-showHistory))})
}
