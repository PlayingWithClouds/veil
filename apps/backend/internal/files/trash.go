package files

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// trashFolder is the folder in each root that removed files move to.
const trashFolder = ".veil-trash"

// TrashFile removes a library file: it moves into its root's trash folder
// (kept until PurgeTrash) and its row is trashed.
func (s *Service) TrashFile(ctx context.Context, fileID string) error {
	file, err := s.File(ctx, fileID)
	if err != nil {
		return err
	}
	if file.LibraryRoot == nil {
		return fmt.Errorf("file %s is not in a library root", fileID)
	}
	opened, err := s.openRoot(ctx, *file.LibraryRoot)
	if err != nil {
		return err
	}
	if err := moveToTrash(opened, *file); err != nil {
		return err
	}
	return s.database.Trash(ctx, tableMediaFile, file.Id)
}

// moveToTrash moves a file's content into the root's trash folder, under its
// root-relative path. A missing file has nothing to move.
func moveToTrash(opened *openedRoot, file schema.MediaFile) error {
	source, err := opened.absolute(file.Path)
	if err != nil {
		return err
	}
	if !pathExists(source) {
		return nil
	}
	target := freeTrashPath(filepath.Join(opened.directory, trashFolder, filepath.FromSlash(opened.rootRelative(file.Path))))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	if err := os.Rename(source, target); err != nil {
		return err
	}
	// The modification time dates the trashing, which PurgeTrash goes by.
	now := time.Now()
	return os.Chtimes(target, now, now)
}

// freeTrashPath is target, or target with a timestamp when it is taken.
func freeTrashPath(target string) string {
	if !pathExists(target) {
		return target
	}
	extension := filepath.Ext(target)
	return target[:len(target)-len(extension)] + time.Now().Format(".20060102-150405.000000000") + extension
}

// PurgeTrash deletes trashed files older than cutoff from the trash folders
// of this device's roots, and file rows trashed before cutoff with their
// fingerprints. It returns how many files it deleted.
func (s *Service) PurgeTrash(ctx context.Context, cutoff time.Time) (int, error) {
	roots, err := s.Roots(ctx)
	if err != nil {
		return 0, err
	}
	purged := 0
	for _, root := range roots {
		opened, err := s.openRoot(ctx, root.Id)
		if err != nil {
			continue
		}
		count, err := purgeTrashFolder(filepath.Join(opened.directory, trashFolder), cutoff)
		purged += count
		if err != nil {
			return purged, err
		}
	}
	return purged, s.purgeTrashedRows(ctx, cutoff)
}

// purgeTrashFolder deletes the files in a trash folder last modified before
// cutoff and returns how many.
func purgeTrashFolder(folder string, cutoff time.Time) (int, error) {
	purged := 0
	err := filepath.WalkDir(folder, func(target string, entry fs.DirEntry, err error) error {
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil || entry.IsDir() {
			return err
		}
		info, err := entry.Info()
		if err != nil || !info.ModTime().Before(cutoff) {
			return nil
		}
		if err := os.Remove(target); err != nil {
			return err
		}
		purged++
		return nil
	})
	return purged, err
}

// purgeTrashedRows deletes file rows trashed before cutoff, and their
// fingerprints.
func (s *Service) purgeTrashedRows(ctx context.Context, cutoff time.Time) error {
	return s.database.Tx(ctx, func(tx *db.DB) error {
		vars := db.Vars{"cutoff": db.FormatTime(cutoff)}
		if _, err := tx.Exec(ctx,
			`DELETE FROM "Fingerprint" WHERE "fileId" IN
			 (SELECT "id" FROM "MediaFile" WHERE "deletedAt" IS NOT NULL AND "deletedAt" < $cutoff)`, vars); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `DELETE FROM "MediaFile" WHERE "deletedAt" IS NOT NULL AND "deletedAt" < $cutoff`, vars)
		return err
	})
}
