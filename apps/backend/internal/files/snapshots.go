package files

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/storage"
)

const (
	// archivePrefix is the blob key prefix of snapshots. It is outside the
	// cache prefixes, so eviction never touches it.
	archivePrefix = "archive/"
	// maxPageBytes and maxCoverBytes bound what a snapshot stores.
	maxPageBytes  = 20 * 1024 * 1024
	maxCoverBytes = 20 * 1024 * 1024
	// snapshotUserAgent is sent with snapshot requests; sites answer bare Go
	// clients with bot pages.
	snapshotUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
)

// Snapshot queues archiving a source's page and its record's cover, and
// returns the job's id.
func (s *Service) Snapshot(ctx context.Context, sourceID string) (string, error) {
	source, err := s.source(ctx, sourceID)
	if err != nil {
		return "", err
	}
	return s.config.Queue.Enqueue(ctx, jobs.Job{
		Kind:      schema.JobKindSnapshot,
		Subject:   &schema.RecordRef{Kind: source.SubjectKind, Id: source.SubjectId},
		Payload:   map[string]any{"source": sourceID},
		DedupeKey: "snapshot:" + sourceID,
	})
}

// SnapshotTask runs snapshot jobs: the source's page HTML and its record's
// cover are stored under archive/<sourceId>/ and the page becomes the
// source's snapshot.
func (s *Service) SnapshotTask() jobs.Task {
	return func(ctx context.Context, job *jobs.Job, progress func(float64)) error {
		sourceID := payloadString(job, "source")
		if sourceID == "" {
			return fmt.Errorf("a snapshot needs a source")
		}
		return s.snapshot(ctx, sourceID)
	}
}

// snapshot archives one source.
func (s *Service) snapshot(ctx context.Context, sourceID string) error {
	source, err := s.source(ctx, sourceID)
	if err != nil {
		return err
	}
	if _, err := s.config.Devices.EnsureLocation(ctx, schema.StorageKindArchive, s.config.BlobDir, nil, nil); err != nil {
		return err
	}
	pageKey := archivePrefix + source.Id + "/page.html"
	if err := s.archiveRemote(ctx, source.Url, pageKey, maxPageBytes); err != nil {
		return fmt.Errorf("archive page: %w", err)
	}
	s.archiveCover(ctx, *source)
	mimeType := "text/html"
	snapshot := schema.Asset{Uri: s.config.Store.DirectURL(pageKey), MimeType: &mimeType}
	return s.database.Update(ctx, tableSource, source.Id, db.Vars{"snapshot": snapshot})
}

// source loads a source, failing when there is none.
func (s *Service) source(ctx context.Context, sourceID string) (*schema.Source, error) {
	source, err := db.Get[schema.Source](ctx, s.database, tableSource, sourceID)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, fmt.Errorf("no source %s", sourceID)
	}
	return source, nil
}

// archiveCover stores the cover of the source's record (else the source's
// own image) beside the page. A cover that can't be fetched doesn't fail the
// snapshot: the page is what it is for.
func (s *Service) archiveCover(ctx context.Context, source schema.Source) {
	cover := s.recordCover(ctx, source)
	if cover == nil {
		return
	}
	key := archivePrefix + source.Id + "/cover" + coverExtension(*cover)
	if err := s.archiveAsset(ctx, cover.Uri, key); err != nil {
		log.Printf("snapshot %s: cover: %v", source.Id, err)
	}
}

// recordCover is the cover of the source's record, else the source's image.
func (s *Service) recordCover(ctx context.Context, source schema.Source) *schema.Asset {
	table := mediaTable(source.SubjectKind)
	if !isCoveredMedia(source.SubjectKind) || table == "" {
		return source.Image
	}
	media, err := db.QueryOneAs[schema.Media](ctx, s.database, `SELECT * FROM `+db.Quote(table)+` WHERE "id" = $id`,
		db.Vars{"id": source.SubjectId})
	if err != nil || media == nil || media.Cover == nil {
		return source.Image
	}
	return media.Cover
}

// isCoveredMedia reports whether records of kind have a cover column.
func isCoveredMedia(kind schema.RecordKind) bool {
	return kind == schema.RecordKindVideo || kind == schema.RecordKindGallery || kind == schema.RecordKindAudio
}

// coverExtension is the file extension an archived cover gets, from its URL
// or its type; ".jpg" when neither tells. The URL goes first: the type maps
// to several extensions, sorted, so image/jpeg would give ".jfif".
func coverExtension(cover schema.Asset) string {
	extension := strings.ToLower(path.Ext(strings.SplitN(cover.Uri, "?", 2)[0]))
	if len(extension) >= 2 && len(extension) <= 5 {
		return extension
	}
	if cover.MimeType == nil {
		return ".jpg"
	}
	extensions, _ := mime.ExtensionsByType(*cover.MimeType)
	if len(extensions) == 0 {
		return ".jpg"
	}
	return extensions[0]
}

// archiveAsset copies an asset into the archive: blob-store assets from the
// store, remote ones over HTTP.
func (s *Service) archiveAsset(ctx context.Context, uri, key string) error {
	blobKey, isBlob := storage.KeyFromURL(uri)
	if isBlob {
		return s.copyBlob(ctx, blobKey, key)
	}
	if !strings.HasPrefix(uri, "http://") && !strings.HasPrefix(uri, "https://") {
		return fmt.Errorf("can't archive %q", uri)
	}
	return s.archiveRemote(ctx, uri, key, maxCoverBytes)
}

// copyBlob copies one stored blob to another key.
func (s *Service) copyBlob(ctx context.Context, from, to string) error {
	content, err := s.config.Store.Get(ctx, from)
	if err != nil {
		return err
	}
	defer content.Close()
	return s.config.Store.Put(ctx, to, content, -1, "")
}

// archiveRemote fetches url and stores at most maxBytes of its body at key.
func (s *Service) archiveRemote(ctx context.Context, url, key string, maxBytes int64) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", snapshotUserAgent)
	response, err := s.config.HTTPClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d fetching %s", response.StatusCode, url)
	}
	return s.config.Store.Put(ctx, key, io.LimitReader(response.Body, maxBytes), -1, response.Header.Get("Content-Type"))
}
