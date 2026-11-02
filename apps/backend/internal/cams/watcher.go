package cams

import (
	"context"
	"log"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// liveStaleAfter is how long a room's live state stands without anyone
// seeing it online again. Listings only show who is online, so a room that
// left them is taken offline by age.
const liveStaleAfter = 10 * time.Minute

// watchedModelsQuery selects the models the watcher checks: auto-recorded
// ones and those followed (an enabled Follow targeting the model).
const watchedModelsQuery = `
	SELECT * FROM "CamModel" WHERE "deletedAt" IS NULL AND (
		"autoRecord" = 1 OR "id" IN (
			SELECT "targetId" FROM "Follow"
			WHERE "targetKind" = $camModel AND "enabled" = 1 AND "deletedAt" IS NULL))
	ORDER BY "displayName" COLLATE NOCASE`

// Run checks the watched models now and then every few minutes until ctx
// ends: it takes rooms nobody saw online lately offline, refreshes the live
// state of auto-recorded and followed models, starts recordings of
// auto-recorded models that are live, and calls the OnLive hook for watched
// models that went live.
func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(s.timing.watchInterval)
	defer ticker.Stop()
	for {
		s.Check(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// Check runs one pass of Run.
func (s *Service) Check(ctx context.Context) {
	s.expireStaleLive(ctx)
	watched, err := db.QueryAs[schema.CamModel](ctx, s.services.Database, watchedModelsQuery,
		db.Vars{"camModel": schema.RecordKindCamModel})
	if err != nil {
		log.Printf("cams: watch: %v", err)
		return
	}
	for index := range watched {
		if ctx.Err() != nil {
			return
		}
		s.checkWatched(ctx, watched[index])
		sleepContext(ctx, s.timing.watchStagger)
	}
}

// expireStaleLive takes rooms offline that nobody saw online for liveStaleAfter.
func (s *Service) expireStaleLive(ctx context.Context) {
	cutoff := db.FormatTime(s.now().Add(-liveStaleAfter))
	_, err := s.services.Database.Exec(ctx,
		`UPDATE "CamModel" SET "live" = NULL, "updatedAt" = $now
		 WHERE "live" IS NOT NULL AND ("lastSeenOnlineAt" IS NULL OR "lastSeenOnlineAt" < $cutoff)`,
		db.Vars{"cutoff": cutoff, "now": db.FormatTime(s.now())})
	if err != nil {
		log.Printf("cams: expire live state: %v", err)
	}
}

// checkWatched refreshes one watched model's live state, announces it when it
// went live and starts its auto-recording. A model already being recorded is
// live by the recording's account and isn't probed.
func (s *Service) checkWatched(ctx context.Context, before schema.CamModel) {
	if s.runningJob(before.Id) != "" {
		return
	}
	s.checkOnline(ctx, &before)
	after, err := s.load(ctx, before.Id)
	if err != nil {
		return
	}
	s.announce(before, *after)
	if after.Live == nil || !after.AutoRecord {
		return
	}
	if _, err := s.StartRecording(ctx, after.Id, RecordOptions{}); err != nil {
		log.Printf("cams: auto-record %s: %v", after.DisplayName, err)
	}
}

// announce calls the OnLive hook when the model began a show the hook wasn't
// told about. A model first seen already live (before this check) is taken
// as announced, so a restart doesn't announce every show again.
func (s *Service) announce(before, after schema.CamModel) {
	current := liveSince(after)
	s.mutex.Lock()
	previous, known := s.announced[after.Id]
	s.announced[after.Id] = current
	hook := s.onLive
	s.mutex.Unlock()
	if current == "" || hook == nil {
		return
	}
	if !known && before.Live != nil {
		return
	}
	if current != previous {
		hook(after)
	}
}

// liveSince identifies the model's current show by its start, "" when offline.
func liveSince(camModel schema.CamModel) string {
	if camModel.Live == nil || camModel.Live.StartedAt == nil {
		return ""
	}
	return *camModel.Live.StartedAt
}
