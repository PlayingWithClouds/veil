package devices

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestCurrentDeviceAndLocations opens the service twice on one database and
// expects one current device, keeping its rename, and one cache location
// whose quota follows the given one.
func TestCurrentDeviceAndLocations(t *testing.T) {
	ctx := context.Background()
	database, err := db.Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	first, err := Open(ctx, database, "nas", schema.PlatformServer)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Rename(ctx, first.Current().Id, "Home server"); err != nil {
		t.Fatal(err)
	}
	second, err := Open(ctx, database, "nas", schema.PlatformServer)
	if err != nil {
		t.Fatal(err)
	}
	if second.Current().Id != first.Current().Id || second.Current().Name != "Home server" {
		t.Fatalf("current = %+v, want the renamed first device", second.Current())
	}

	defaultQuota := int64(300)
	cache, err := second.EnsureLocation(ctx, schema.StorageKindCache, "/data/blobs", nil, &defaultQuota)
	if err != nil || cache.QuotaBytes == nil || *cache.QuotaBytes != 300 {
		t.Fatalf("cache = %+v, err %v", cache, err)
	}
	quota := int64(500)
	again, err := second.EnsureLocation(ctx, schema.StorageKindCache, "/data/blobs", &quota, &defaultQuota)
	if err != nil || again.Id != cache.Id || *again.QuotaBytes != 500 {
		t.Fatalf("again = %+v, err %v", again, err)
	}
	locations, _ := second.Locations(ctx, "")
	if len(locations) != 1 {
		t.Errorf("locations = %+v", locations)
	}
	negative := int64(-1)
	if _, err := second.SetQuota(ctx, cache.Id, &negative); err == nil {
		t.Error("negative quota accepted")
	}
}
