package devices

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// locationTable is the table storage locations live in.
const locationTable = "StorageLocation"

// storageKinds are the valid kinds of storage location.
var storageKinds = []schema.StorageKind{schema.StorageKindLibrary, schema.StorageKindCache, schema.StorageKindArchive}

// Locations lists a device's storage locations ("" = this device), by kind.
func (s *Service) Locations(ctx context.Context, deviceID string) ([]schema.StorageLocation, error) {
	if deviceID == "" {
		deviceID = s.current.Id
	}
	return db.QueryAs[schema.StorageLocation](ctx, s.database,
		`SELECT * FROM "StorageLocation" WHERE "device" = $device AND "deletedAt" IS NULL ORDER BY "kind", "createdAt"`,
		db.Vars{"device": deviceID})
}

// Location loads one storage location; nil when it doesn't exist.
func (s *Service) Location(ctx context.Context, id string) (*schema.StorageLocation, error) {
	return db.Get[schema.StorageLocation](ctx, s.database, locationTable, id)
}

// EnsureLocation returns this device's location of kind at uri, creating it
// when missing. A non-nil quota replaces the stored one; a new location
// without one gets defaultQuota (nil = unlimited).
func (s *Service) EnsureLocation(ctx context.Context, kind schema.StorageKind, uri string, quota, defaultQuota *int64) (*schema.StorageLocation, error) {
	existing, err := db.QueryOneAs[schema.StorageLocation](ctx, s.database,
		`SELECT * FROM "StorageLocation" WHERE "device" = $device AND "kind" = $kind AND "uri" = $uri AND "deletedAt" IS NULL`,
		db.Vars{"device": s.current.Id, "kind": kind, "uri": uri})
	if err != nil {
		return nil, err
	}
	if existing == nil {
		if quota == nil {
			quota = defaultQuota
		}
		return s.AddLocation(ctx, kind, uri, quota)
	}
	if quota == nil {
		return existing, nil
	}
	return s.SetQuota(ctx, existing.Id, quota)
}

// AddLocation adds a storage location on this device.
func (s *Service) AddLocation(ctx context.Context, kind schema.StorageKind, uri string, quota *int64) (*schema.StorageLocation, error) {
	if !slices.Contains(storageKinds, kind) {
		return nil, fmt.Errorf("unknown storage kind %q", kind)
	}
	uri = strings.TrimSpace(uri)
	if uri == "" {
		return nil, fmt.Errorf("a storage location needs a path or URI")
	}
	if err := validQuota(quota); err != nil {
		return nil, err
	}
	fields := db.Vars{"device": s.current.Id, "kind": kind, "uri": uri, "usedBytes": 0}
	if quota != nil {
		fields["quotaBytes"] = *quota
	}
	id, err := s.database.Insert(ctx, locationTable, fields)
	if err != nil {
		return nil, err
	}
	return s.Location(ctx, id)
}

// SetQuota changes a location's quota; nil removes it.
func (s *Service) SetQuota(ctx context.Context, id string, quota *int64) (*schema.StorageLocation, error) {
	if err := validQuota(quota); err != nil {
		return nil, err
	}
	fields := db.Vars{"quotaBytes": nil}
	if quota != nil {
		fields["quotaBytes"] = *quota
	}
	if err := s.database.Update(ctx, locationTable, id, fields); err != nil {
		return nil, err
	}
	return s.Location(ctx, id)
}

// validQuota rejects quotas that aren't positive.
func validQuota(quota *int64) error {
	if quota != nil && *quota <= 0 {
		return fmt.Errorf("a quota is a positive number of bytes, not %d", *quota)
	}
	return nil
}

// SetUsedBytes records how much a location holds.
func (s *Service) SetUsedBytes(ctx context.Context, id string, used int64) error {
	_, err := s.database.Exec(ctx, `UPDATE "StorageLocation" SET "usedBytes" = $used WHERE "id" = $id`,
		db.Vars{"id": id, "used": used})
	return err
}

// RemoveLocation trashes a storage location. Its files stay where they are.
func (s *Service) RemoveLocation(ctx context.Context, id string) error {
	return s.database.Trash(ctx, locationTable, id)
}
