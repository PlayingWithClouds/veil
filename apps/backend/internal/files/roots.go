package files

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// RootEdit changes a library root. Nil fields stay as they are; a zero
// interval stops scheduled scans, an empty template stops organizing.
type RootEdit struct {
	ScanIntervalMinutes *int
	OrganizeTemplate    *string
}

// AddRoot adds a folder of the user's own files. path is absolute (and
// must lie inside the location) or relative to the location's URI. A nil
// interval means the root is only scanned on request.
func (s *Service) AddRoot(ctx context.Context, locationID, path string, scanIntervalMinutes *int, organizeTemplate *string) (*schema.LibraryRoot, error) {
	location, err := s.location(ctx, locationID)
	if err != nil {
		return nil, err
	}
	relative, err := rootPath(*location, path)
	if err != nil {
		return nil, err
	}
	fields := db.Vars{"location": location.Id, "path": relative}
	edit := RootEdit{ScanIntervalMinutes: scanIntervalMinutes, OrganizeTemplate: organizeTemplate}
	if err := setRootEdit(fields, edit); err != nil {
		return nil, err
	}
	id, err := s.database.Insert(ctx, tableLibraryRoot, fields)
	if err != nil {
		return nil, err
	}
	return s.Root(ctx, id)
}

// rootPath turns a root's path into one relative to its location, checking
// that the folder exists when the location is on this machine.
func rootPath(location schema.StorageLocation, path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("a library root needs a path")
	}
	absolute := path
	if !filepath.IsAbs(path) {
		absolute = filepath.Join(location.Uri, path)
	}
	if !isWithin(location.Uri, absolute) {
		return "", fmt.Errorf("%s is outside the storage location %s", path, location.Uri)
	}
	info, err := os.Stat(absolute)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("%s is not a folder", absolute)
	}
	relative, err := filepath.Rel(location.Uri, absolute)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(relative), nil
}

// setRootEdit validates edit and writes it into fields.
func setRootEdit(fields db.Vars, edit RootEdit) error {
	if edit.ScanIntervalMinutes != nil {
		if *edit.ScanIntervalMinutes < 0 {
			return fmt.Errorf("a scan interval is a positive number of minutes")
		}
		fields["scanIntervalMinutes"] = nil
		if *edit.ScanIntervalMinutes > 0 {
			fields["scanIntervalMinutes"] = *edit.ScanIntervalMinutes
		}
	}
	if edit.OrganizeTemplate != nil {
		template := strings.TrimSpace(*edit.OrganizeTemplate)
		if err := validateTemplate(template); err != nil {
			return err
		}
		fields["organizeTemplate"] = optionalText(template)
	}
	return nil
}

// UpdateRoot changes a library root's scan interval or organize template.
func (s *Service) UpdateRoot(ctx context.Context, rootID string, edit RootEdit) (*schema.LibraryRoot, error) {
	fields := db.Vars{}
	if err := setRootEdit(fields, edit); err != nil {
		return nil, err
	}
	if err := s.database.Update(ctx, tableLibraryRoot, rootID, fields); err != nil {
		return nil, err
	}
	return s.Root(ctx, rootID)
}

// RemoveRoot trashes a library root: it is no longer scanned. Its files and
// their rows stay where they are.
func (s *Service) RemoveRoot(ctx context.Context, rootID string) error {
	return s.database.Trash(ctx, tableLibraryRoot, rootID)
}

// Roots lists every library root, oldest first.
func (s *Service) Roots(ctx context.Context) ([]schema.LibraryRoot, error) {
	return db.QueryAs[schema.LibraryRoot](ctx, s.database,
		`SELECT * FROM "LibraryRoot" WHERE "deletedAt" IS NULL ORDER BY "createdAt"`, nil)
}

// Root loads one library root; nil when there is none.
func (s *Service) Root(ctx context.Context, rootID string) (*schema.LibraryRoot, error) {
	return db.Get[schema.LibraryRoot](ctx, s.database, tableLibraryRoot, rootID)
}

// openedRoot is a root with its location and folder on this device.
type openedRoot struct {
	root     schema.LibraryRoot
	location schema.StorageLocation
	// directory is the root's absolute folder.
	directory string
}

// openRoot loads a root that lives on this device.
func (s *Service) openRoot(ctx context.Context, rootID string) (*openedRoot, error) {
	root, err := s.Root(ctx, rootID)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, fmt.Errorf("no library root %s", rootID)
	}
	location, err := s.location(ctx, root.Location)
	if err != nil {
		return nil, err
	}
	if !s.isOnThisDevice(*location) {
		return nil, fmt.Errorf("library root %s is on another device", rootID)
	}
	directory, err := insidePath(location.Uri, root.Path)
	if err != nil {
		return nil, err
	}
	return &openedRoot{root: *root, location: *location, directory: directory}, nil
}

// locationPath is a root-relative path as a path relative to the location.
func (r *openedRoot) locationPath(rootRelative string) string {
	return joinSlash(r.root.Path, rootRelative)
}

// rootRelative is a location-relative path as a path relative to the root.
func (r *openedRoot) rootRelative(locationPath string) string {
	prefix := joinSlash(r.root.Path)
	if prefix == "" {
		return locationPath
	}
	return strings.TrimPrefix(locationPath, prefix+"/")
}

// absolute is the file at a location-relative path.
func (r *openedRoot) absolute(locationPath string) (string, error) {
	return insidePath(r.location.Uri, locationPath)
}
