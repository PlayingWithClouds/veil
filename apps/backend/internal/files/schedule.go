package files

import (
	"context"
	"log"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// scheduleInterval is how often Run looks for roots due for a scan.
const scheduleInterval = time.Minute

// Run queues scans of this device's roots whose scan interval elapsed, at
// start and then every minute, until ctx ends.
func (s *Service) Run(ctx context.Context) {
	for {
		if _, err := s.EnqueueDueScans(ctx, time.Now()); err != nil {
			log.Printf("library scans: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(scheduleInterval):
		}
	}
}

// EnqueueDueScans queues a scan of every root on this device whose interval
// elapsed by now, and returns how many it queued.
func (s *Service) EnqueueDueScans(ctx context.Context, now time.Time) (int, error) {
	roots, err := db.QueryAs[schema.LibraryRoot](ctx, s.database,
		`SELECT "LibraryRoot".* FROM "LibraryRoot"
		 JOIN "StorageLocation" ON "StorageLocation"."id" = "LibraryRoot"."location"
		 WHERE "LibraryRoot"."deletedAt" IS NULL AND "LibraryRoot"."scanIntervalMinutes" > 0
		   AND "StorageLocation"."device" = $device AND "StorageLocation"."deletedAt" IS NULL`,
		db.Vars{"device": s.config.Devices.Current().Id})
	if err != nil {
		return 0, err
	}
	queued := 0
	for _, root := range roots {
		if !isScanDue(root, now) {
			continue
		}
		if _, err := s.ScanRoot(ctx, root.Id); err != nil {
			return queued, err
		}
		queued++
	}
	return queued, nil
}

// isScanDue reports whether a root's last scan is older than its interval.
func isScanDue(root schema.LibraryRoot, now time.Time) bool {
	if root.ScanIntervalMinutes == nil {
		return false
	}
	if root.LastScanAt == nil {
		return true
	}
	lastScan, err := time.Parse(db.TimeFormat, *root.LastScanAt)
	if err != nil {
		return true
	}
	interval := time.Duration(*root.ScanIntervalMinutes) * time.Minute
	return !now.Before(lastScan.Add(interval))
}
