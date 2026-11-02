package files

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/schema"
)

// FileRoutePrefix is where Handler is mounted: /api/file/{mediaFileId}.
const FileRoutePrefix = "/api/file/"

// Files lists a record's files, present ones first.
func (s *Service) Files(ctx context.Context, kind schema.RecordKind, id string) ([]schema.MediaFile, error) {
	return db.QueryAs[schema.MediaFile](ctx, s.database,
		`SELECT * FROM "MediaFile" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL
		 ORDER BY "missing", "createdAt"`, db.Vars{"kind": kind, "id": id})
}

// File loads one file, failing when there is none.
func (s *Service) File(ctx context.Context, fileID string) (*schema.MediaFile, error) {
	file, err := db.Get[schema.MediaFile](ctx, s.database, tableMediaFile, fileID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, fmt.Errorf("no file %s", fileID)
	}
	return file, nil
}

// FileURL is where a client plays or downloads a file: blob-store files at
// their blob URL, others through Handler.
func (s *Service) FileURL(ctx context.Context, file schema.MediaFile) string {
	location, err := s.config.Devices.Location(ctx, file.Location)
	if err == nil && location != nil && s.isBlobStore(*location) {
		return s.config.Store.DirectURL(file.Path)
	}
	return s.fileRouteURL(file.Id)
}

// fileRouteURL is the URL Handler serves a file at.
func (s *Service) fileRouteURL(fileID string) string {
	return strings.TrimRight(s.config.PublicURL, "/") + FileRoutePrefix + fileID
}

// Handler serves files at FileRoutePrefix + id, with range requests for
// seeking. Files of other devices, missing ones and paths that leave their
// location are not found.
func (s *Service) Handler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		file, err := db.Get[schema.MediaFile](request.Context(), s.database, tableMediaFile, strings.TrimPrefix(request.URL.Path, FileRoutePrefix))
		if err != nil || file == nil {
			http.NotFound(writer, request)
			return
		}
		s.serveFile(writer, request, *file)
	})
}

// serveFile writes one file's content, honouring range requests.
func (s *Service) serveFile(writer http.ResponseWriter, request *http.Request, file schema.MediaFile) {
	opened, err := s.open(request.Context(), file)
	if err != nil {
		http.NotFound(writer, request)
		return
	}
	defer opened.content.Close()
	writer.Header().Set("Content-Type", file.MimeType)
	http.ServeContent(writer, request, path.Base(file.Path), opened.modifiedAt, opened.content)
}

// openedFile is a file's content ready to read and seek.
type openedFile struct {
	content    io.ReadSeekCloser
	modifiedAt time.Time
}

// open opens a file of this device: blob-store files through the store
// (which decrypts), others from disk, refusing paths outside their location.
func (s *Service) open(ctx context.Context, file schema.MediaFile) (*openedFile, error) {
	location, err := s.location(ctx, file.Location)
	if err != nil {
		return nil, err
	}
	if !s.isOnThisDevice(*location) {
		return nil, fmt.Errorf("file %s is on another device", file.Id)
	}
	if s.isBlobStore(*location) {
		return s.openBlob(ctx, file.Path)
	}
	return openInside(location.Uri, file.Path)
}

// openBlob opens a blob of the store by key.
func (s *Service) openBlob(ctx context.Context, key string) (*openedFile, error) {
	listed, err := s.config.Store.List(ctx, key)
	if err != nil || len(listed) != 1 {
		return nil, fmt.Errorf("no blob %s", key)
	}
	content, err := s.config.Store.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	seekable, ok := content.(io.ReadSeekCloser)
	if !ok {
		content.Close()
		return nil, fmt.Errorf("blob %s can't seek", key)
	}
	return &openedFile{content: seekable, modifiedAt: listed[0].LastUsed}, nil
}

// openInside opens the regular file at relative under base.
func openInside(base, relative string) (*openedFile, error) {
	target, err := insidePath(base, relative)
	if err != nil {
		return nil, err
	}
	if !resolvedWithin(base, target) {
		return nil, fmt.Errorf("%q leads outside %s", relative, base)
	}
	content, err := os.Open(target)
	if err != nil {
		return nil, err
	}
	info, err := content.Stat()
	if err != nil || !info.Mode().IsRegular() {
		content.Close()
		return nil, fmt.Errorf("%s is not a file", target)
	}
	return &openedFile{content: content, modifiedAt: info.ModTime()}, nil
}
