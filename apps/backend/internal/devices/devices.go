// Package devices keeps the Device rows (one per Veil install) and the
// storage locations each device keeps files in. This install's own row is
// created on first start and marked current; events, jobs and locations
// point at it. Sync between devices builds on these rows later.
package devices

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// deviceTable is the table devices live in.
const deviceTable = "Device"

// platforms are the valid platform names.
var platforms = []schema.Platform{
	schema.PlatformServer,
	schema.PlatformDesktop,
	schema.PlatformAndroid,
	schema.PlatformTv,
	schema.PlatformVr,
}

// Service reads and writes devices and storage locations.
type Service struct {
	database *db.DB
	current  schema.Device
}

// Open returns the service with this install's device row, creating it with
// name and platform on first start. Later starts keep the stored name (the
// user may have renamed it) and stamp lastSeenAt.
func Open(ctx context.Context, database *db.DB, name string, platform schema.Platform) (*Service, error) {
	if !slices.Contains(platforms, platform) {
		return nil, fmt.Errorf("unknown platform %q", platform)
	}
	service := &Service{database: database}
	current, err := service.ensureCurrent(ctx, name, platform)
	if err != nil {
		return nil, err
	}
	service.current = *current
	return service, nil
}

// ensureCurrent loads or creates the current device and marks it seen.
func (s *Service) ensureCurrent(ctx context.Context, name string, platform schema.Platform) (*schema.Device, error) {
	existing, err := db.QueryOneAs[schema.Device](ctx, s.database,
		`SELECT * FROM "Device" WHERE "current" = 1 AND "deletedAt" IS NULL ORDER BY "createdAt" LIMIT 1`, nil)
	if err != nil {
		return nil, err
	}
	now := db.Now()
	if existing != nil {
		if err := s.database.Update(ctx, deviceTable, existing.Id, db.Vars{"lastSeenAt": now, "platform": platform}); err != nil {
			return nil, err
		}
		return db.Get[schema.Device](ctx, s.database, deviceTable, existing.Id)
	}
	id, err := s.database.Insert(ctx, deviceTable, db.Vars{"name": name, "platform": platform, "current": true, "lastSeenAt": now})
	if err != nil {
		return nil, err
	}
	return db.Get[schema.Device](ctx, s.database, deviceTable, id)
}

// Current returns this install's device.
func (s *Service) Current() schema.Device {
	return s.current
}

// List returns every known device, this one first.
func (s *Service) List(ctx context.Context) ([]schema.Device, error) {
	return db.QueryAs[schema.Device](ctx, s.database,
		`SELECT * FROM "Device" WHERE "deletedAt" IS NULL ORDER BY "current" DESC, "name" COLLATE NOCASE`, nil)
}

// Device loads one device; nil when it doesn't exist.
func (s *Service) Device(ctx context.Context, id string) (*schema.Device, error) {
	return db.Get[schema.Device](ctx, s.database, deviceTable, id)
}

// Rename gives a device a new name.
func (s *Service) Rename(ctx context.Context, id, name string) (*schema.Device, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("a device needs a name")
	}
	if err := s.database.Update(ctx, deviceTable, id, db.Vars{"name": name}); err != nil {
		return nil, err
	}
	if id == s.current.Id {
		s.current.Name = name
	}
	return db.Get[schema.Device](ctx, s.database, deviceTable, id)
}

// ParsePlatform reads a platform name; "" means a server.
func ParsePlatform(name string) (schema.Platform, error) {
	if name == "" {
		return schema.PlatformServer, nil
	}
	platform := schema.Platform(strings.ToLower(name))
	if !slices.Contains(platforms, platform) {
		return "", fmt.Errorf("unknown platform %q", name)
	}
	return platform, nil
}
