// Package files keeps the files behind records: the user's own library
// folders (LibraryRoot, scanned into MediaFile rows and organized from a
// template), files served over HTTP, content fingerprints, downloads of
// streams into the blob store and archived snapshots of source pages.
//
// Paths of MediaFile rows are slash-separated and relative to their storage
// location, so a row means the same file on every device that mounts it.
// Destructive operations never delete a user's file: it moves to the root's
// .veil-trash folder and is purged after the trash retention.
package files

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/devices"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/outbound"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/settings"
	"github.com/playingwithclouds/veil/internal/storage"
	"github.com/playingwithclouds/veil/internal/stream"
)

// Table names, exactly as the schema names the models.
const (
	tableMediaFile   = "MediaFile"
	tableLibraryRoot = "LibraryRoot"
	tableFingerprint = "Fingerprint"
	tableSource      = "Source"
	tableStream      = "Stream"
	tableJob         = "Job"
)

// StreamResolver turns a stream's provider URL into a playable one;
// *resolve.Service is one.
type StreamResolver interface {
	Resolve(ctx context.Context, providerURL string) (*plugins.ResolveResult, error)
}

// Config is what the service works with.
type Config struct {
	Database *db.DB
	Devices  *devices.Service
	Queue    *jobs.Queue
	// Settings gives the download speed limit, streaming deferral and trash
	// retention; nil uses the defaults (unlimited, allowed).
	Settings *settings.Service
	Store    *storage.Client
	// BlobDir is the directory Store keeps blobs in. Locations with this URI
	// are the blob store: their files are read through Store, which decrypts.
	BlobDir string
	// PublicURL is the backend's externally reachable origin.
	PublicURL   string
	Resolver    StreamResolver
	StreamCache *stream.CacheService
	// HTTPClient fetches snapshot pages and covers; nil = outbound.Client.
	HTTPClient *http.Client
	// OnFileAdded is called for every file a library scan adds, for
	// automations (fileAdded trigger).
	OnFileAdded func(schema.MediaFile)
}

// Service manages library roots, files, fingerprints, downloads and snapshots.
type Service struct {
	config   Config
	database *db.DB
	recorder *provenance.Recorder
}

// New returns the service for config.
func New(config Config) *Service {
	if config.HTTPClient == nil {
		config.HTTPClient = outbound.Client
	}
	return &Service{config: config, database: config.Database, recorder: provenance.New(config.Database)}
}

// location loads a storage location, failing when there is none.
func (s *Service) location(ctx context.Context, id string) (*schema.StorageLocation, error) {
	location, err := s.config.Devices.Location(ctx, id)
	if err != nil {
		return nil, err
	}
	if location == nil {
		return nil, fmt.Errorf("no storage location %s", id)
	}
	return location, nil
}

// isBlobStore reports whether location is the blob store directory.
func (s *Service) isBlobStore(location schema.StorageLocation) bool {
	if s.config.BlobDir == "" {
		return false
	}
	return filepath.Clean(location.Uri) == filepath.Clean(s.config.BlobDir)
}

// isOnThisDevice reports whether location belongs to this install.
func (s *Service) isOnThisDevice(location schema.StorageLocation) bool {
	return location.Device == s.config.Devices.Current().Id
}

// insidePath joins base and a slash-separated relative path, refusing
// results that leave base.
func insidePath(base, relative string) (string, error) {
	joined := filepath.Join(base, filepath.FromSlash(relative))
	if !isWithin(base, joined) {
		return "", fmt.Errorf("%q is outside %s", relative, base)
	}
	return joined, nil
}

// isWithin reports whether target is base or below it.
func isWithin(base, target string) bool {
	relative, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

// resolvedWithin is isWithin after following symlinks, so a link inside a
// location can't hand out a file outside it.
func resolvedWithin(base, target string) bool {
	resolvedBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		return false
	}
	resolvedTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		return false
	}
	return isWithin(resolvedBase, resolvedTarget)
}

// joinSlash joins slash-separated path parts, dropping empty and "." ones.
func joinSlash(parts ...string) string {
	kept := []string{}
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		kept = append(kept, part)
	}
	return strings.Join(kept, "/")
}

// mediaTable is the table of a media record kind that has files, or "".
func mediaTable(kind schema.RecordKind) string {
	if kind == schema.RecordKindAudio {
		return "Audio"
	}
	return provenance.Table(kind)
}

// recordExists reports whether the record ref points at is stored and not
// trashed.
func (s *Service) recordExists(ctx context.Context, ref schema.RecordRef) (bool, error) {
	table := mediaTable(ref.Kind)
	if table == "" {
		return false, fmt.Errorf("a %s has no files", ref.Kind)
	}
	count, err := s.database.Int(ctx, `SELECT COUNT(*) FROM `+db.Quote(table)+` WHERE "id" = $id AND "deletedAt" IS NULL`,
		db.Vars{"id": ref.Id})
	return count > 0, err
}

// requireRecord fails when the record ref points at doesn't exist.
func (s *Service) requireRecord(ctx context.Context, ref schema.RecordRef) error {
	exists, err := s.recordExists(ctx, ref)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("no %s %s", ref.Kind, ref.Id)
	}
	return nil
}

// payloadString reads a string from a job's payload; "" when absent.
func payloadString(job *jobs.Job, key string) string {
	value, _ := job.Payload[key].(string)
	return value
}

// fileExists reports whether a regular file is at target.
func fileExists(target string) bool {
	info, err := os.Stat(target)
	return err == nil && info.Mode().IsRegular()
}

// pathExists reports whether anything is at target.
func pathExists(target string) bool {
	_, err := os.Lstat(target)
	return err == nil
}

// optionalText returns a pointer to text, or nil for "".
func optionalText(text string) *string {
	if text == "" {
		return nil
	}
	return &text
}
