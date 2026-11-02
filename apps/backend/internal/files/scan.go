package files

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/schema"
)

// mediaType is what a file extension holds.
type mediaType struct {
	kind     schema.RecordKind
	mimeType string
}

// mediaTypes are the file extensions a scan picks up.
var mediaTypes = map[string]mediaType{
	".mp4":  {schema.RecordKindVideo, "video/mp4"},
	".m4v":  {schema.RecordKindVideo, "video/x-m4v"},
	".mkv":  {schema.RecordKindVideo, "video/x-matroska"},
	".webm": {schema.RecordKindVideo, "video/webm"},
	".mov":  {schema.RecordKindVideo, "video/quicktime"},
	".avi":  {schema.RecordKindVideo, "video/x-msvideo"},
	".wmv":  {schema.RecordKindVideo, "video/x-ms-wmv"},
	".flv":  {schema.RecordKindVideo, "video/x-flv"},
	".mpg":  {schema.RecordKindVideo, "video/mpeg"},
	".mpeg": {schema.RecordKindVideo, "video/mpeg"},
	".ts":   {schema.RecordKindVideo, "video/mp2t"},
	".mp3":  {schema.RecordKindAudio, "audio/mpeg"},
	".m4a":  {schema.RecordKindAudio, "audio/mp4"},
	".aac":  {schema.RecordKindAudio, "audio/aac"},
	".flac": {schema.RecordKindAudio, "audio/flac"},
	".wav":  {schema.RecordKindAudio, "audio/wav"},
	".ogg":  {schema.RecordKindAudio, "audio/ogg"},
	".opus": {schema.RecordKindAudio, "audio/opus"},
	".jpg":  {schema.RecordKindImage, "image/jpeg"},
	".jpeg": {schema.RecordKindImage, "image/jpeg"},
	".png":  {schema.RecordKindImage, "image/png"},
	".gif":  {schema.RecordKindImage, "image/gif"},
	".webp": {schema.RecordKindImage, "image/webp"},
	".avif": {schema.RecordKindImage, "image/avif"},
	".heic": {schema.RecordKindImage, "image/heic"},
}

// mediaTypeOf is the media type of a file name; ok is false for other files.
func mediaTypeOf(name string) (mediaType, bool) {
	found, ok := mediaTypes[strings.ToLower(filepath.Ext(name))]
	return found, ok
}

// foundFile is a media file a scan found on disk.
type foundFile struct {
	// path is relative to the location, slash-separated.
	path      string
	absolute  string
	sizeBytes int64
	mediaType mediaType
}

// ScanRoot queues a scan of a library root and returns the job's id.
func (s *Service) ScanRoot(ctx context.Context, rootID string) (string, error) {
	return s.config.Queue.Enqueue(ctx, jobs.Job{
		Kind:      schema.JobKindLibraryScan,
		Payload:   map[string]any{"root": rootID},
		DedupeKey: "libraryScan:" + rootID,
	})
}

// ScanTask runs libraryScan jobs: the payload's root is walked, its media
// files recorded, new ones fingerprinted and matched to records.
func (s *Service) ScanTask() jobs.Task {
	return func(ctx context.Context, job *jobs.Job, progress func(float64)) error {
		rootID := payloadString(job, "root")
		if rootID == "" {
			return fmt.Errorf("a library scan needs a root")
		}
		return s.scan(ctx, rootID, progress)
	}
}

// scan walks a root, upserts its files and marks vanished ones missing.
func (s *Service) scan(ctx context.Context, rootID string, progress func(float64)) error {
	opened, err := s.openRoot(ctx, rootID)
	if err != nil {
		return err
	}
	found, err := listMediaFiles(opened)
	if err != nil {
		return err
	}
	known, err := s.rootFiles(ctx, rootID)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for index, file := range found {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		seen[file.path] = true
		if err := s.scanFile(ctx, opened, file, known[file.path]); err != nil {
			log.Printf("library scan: %s: %v", file.absolute, err)
		}
		progress(float64(index+1) / float64(len(found)+1))
	}
	if err := s.markMissing(ctx, known, seen); err != nil {
		return err
	}
	return s.database.Update(ctx, tableLibraryRoot, rootID, db.Vars{"lastScanAt": db.Now()})
}

// listMediaFiles walks a root for media files. Hidden folders, the trash
// folder among them, are skipped.
func listMediaFiles(opened *openedRoot) ([]foundFile, error) {
	var found []foundFile
	err := filepath.WalkDir(opened.directory, func(target string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && target != opened.directory && strings.HasPrefix(entry.Name(), ".") {
			return filepath.SkipDir
		}
		file, ok := foundFileAt(opened, target, entry)
		if ok {
			found = append(found, file)
		}
		return nil
	})
	return found, err
}

// foundFileAt describes the walked entry at target when it is a media file.
func foundFileAt(opened *openedRoot, target string, entry fs.DirEntry) (foundFile, bool) {
	if !entry.Type().IsRegular() || strings.HasPrefix(entry.Name(), ".") {
		return foundFile{}, false
	}
	found, ok := mediaTypeOf(entry.Name())
	if !ok {
		return foundFile{}, false
	}
	info, err := entry.Info()
	if err != nil {
		return foundFile{}, false
	}
	relative, err := filepath.Rel(opened.location.Uri, target)
	if err != nil {
		return foundFile{}, false
	}
	return foundFile{path: filepath.ToSlash(relative), absolute: target, sizeBytes: info.Size(), mediaType: found}, true
}

// rootFiles maps the location path of every file row of a root to the row.
func (s *Service) rootFiles(ctx context.Context, rootID string) (map[string]schema.MediaFile, error) {
	rows, err := db.QueryAs[schema.MediaFile](ctx, s.database,
		`SELECT * FROM "MediaFile" WHERE "libraryRoot" = $root AND "deletedAt" IS NULL`, db.Vars{"root": rootID})
	if err != nil {
		return nil, err
	}
	byPath := make(map[string]schema.MediaFile, len(rows))
	for _, row := range rows {
		byPath[row.Path] = row
	}
	return byPath, nil
}

// scanFile records one found file: new files are added, changed or
// reappeared ones refreshed, unchanged ones left alone.
func (s *Service) scanFile(ctx context.Context, opened *openedRoot, file foundFile, existing schema.MediaFile) error {
	if existing.Id == "" {
		return s.addFile(ctx, opened, file)
	}
	if existing.SizeBytes == file.sizeBytes && !existing.Missing {
		return nil
	}
	return s.refreshFile(ctx, existing, file)
}

// refreshFile updates a known file whose size changed or that came back,
// fingerprinting it again when its content changed.
func (s *Service) refreshFile(ctx context.Context, existing schema.MediaFile, file foundFile) error {
	if err := s.database.Update(ctx, tableMediaFile, existing.Id, db.Vars{"sizeBytes": file.sizeBytes, "missing": false}); err != nil {
		return err
	}
	if existing.SizeBytes == file.sizeBytes {
		return nil
	}
	hashes, err := hashPath(file.absolute)
	if err != nil {
		return err
	}
	return s.replaceFingerprints(ctx, existing, hashes)
}

// addFile fingerprints a new file, attaches it to the record it matches or
// a new one, and reports it to OnFileAdded.
func (s *Service) addFile(ctx context.Context, opened *openedRoot, file foundFile) error {
	hashes, err := hashPath(file.absolute)
	if err != nil {
		return err
	}
	fileID := db.NewID()
	subject, err := s.matchingSubject(ctx, file.mediaType.kind, hashes)
	if err != nil {
		return err
	}
	if subject == nil {
		subject, err = s.createRecord(ctx, opened, file, fileID)
	}
	if err != nil {
		return err
	}
	row := schema.MediaFile{
		SubjectKind: subject.Kind,
		SubjectId:   subject.Id,
		Location:    opened.location.Id,
		Path:        file.path,
		SizeBytes:   file.sizeBytes,
		MimeType:    file.mediaType.mimeType,
		LibraryRoot: &opened.root.Id,
	}
	row.Id = fileID
	if err := s.insertFile(ctx, row); err != nil {
		return err
	}
	if err := s.replaceFingerprints(ctx, row, hashes); err != nil {
		return err
	}
	if s.config.OnFileAdded != nil {
		s.config.OnFileAdded(row)
	}
	return nil
}

// insertFile inserts a file row. A trashed row at the same location and
// path still holds the unique index, so it goes first.
func (s *Service) insertFile(ctx context.Context, row schema.MediaFile) error {
	return s.database.Tx(ctx, func(tx *db.DB) error {
		if err := clearTrashedRow(ctx, tx, row.Location, row.Path); err != nil {
			return err
		}
		_, err := tx.Insert(ctx, tableMediaFile, row)
		return err
	})
}

// clearTrashedRow deletes the trashed file row at location and path, which
// would otherwise block the unique index for a live one, with its
// fingerprints.
func clearTrashedRow(ctx context.Context, tx *db.DB, locationID, path string) error {
	vars := db.Vars{"location": locationID, "path": path}
	if _, err := tx.Exec(ctx,
		`DELETE FROM "Fingerprint" WHERE "fileId" IN (SELECT "id" FROM "MediaFile"
		 WHERE "location" = $location AND "path" = $path AND "deletedAt" IS NOT NULL)`, vars); err != nil {
		return err
	}
	_, err := tx.Exec(ctx,
		`DELETE FROM "MediaFile" WHERE "location" = $location AND "path" = $path AND "deletedAt" IS NOT NULL`, vars)
	return err
}

// markMissing flags the known files a scan didn't see.
func (s *Service) markMissing(ctx context.Context, known map[string]schema.MediaFile, seen map[string]bool) error {
	for path, file := range known {
		if seen[path] || file.Missing {
			continue
		}
		if err := s.database.Update(ctx, tableMediaFile, file.Id, db.Vars{"missing": true}); err != nil {
			return err
		}
	}
	return nil
}

// hashPath fingerprints the file at target.
func hashPath(target string) ([]hashValue, error) {
	file, err := os.Open(target)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	return computeHashes(file, info.Size())
}
