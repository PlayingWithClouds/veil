package files

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/jobs"
	"github.com/playingwithclouds/veil/internal/schema"
	"github.com/playingwithclouds/veil/internal/storage"
)

// downloadDeferDelay is how long a download held back by playback waits
// before it checks again.
const downloadDeferDelay = 15 * time.Second

// maxManifestDepth bounds how far blobSize follows HLS manifests: a master
// playlist references variant playlists, which reference segments.
const maxManifestDepth = 2

// Download is a download job with the file it produced, once it has.
type Download struct {
	Job  jobs.Job
	File *schema.MediaFile
}

// Download queues a download of a record's stream (nil = its best one) and
// returns the job's id: one download per record, so a queued, running or
// finished one is returned as it is.
func (s *Service) Download(ctx context.Context, subject schema.RecordRef, streamID *string) (string, error) {
	if err := s.requireRecord(ctx, subject); err != nil {
		return "", err
	}
	finished, err := s.finishedDownload(ctx, subject)
	if err != nil || finished != "" {
		return finished, err
	}
	payload := map[string]any{"title": s.recordTitle(ctx, subject)}
	if streamID != nil {
		payload["stream"] = *streamID
	}
	return s.config.Queue.Enqueue(ctx, jobs.Job{
		Kind:      jobs.KindDownload,
		Subject:   &subject,
		Payload:   payload,
		DedupeKey: "download:" + string(subject.Kind) + ":" + subject.Id,
	})
}

// finishedDownload is the id of a succeeded download of subject whose file
// is still recorded, or "".
func (s *Service) finishedDownload(ctx context.Context, subject schema.RecordRef) (string, error) {
	ids, err := s.database.Strings(ctx,
		`SELECT "Job"."id" FROM "Job" JOIN "MediaFile" ON "MediaFile"."id" = json_extract("Job"."payload", '$.file')
		 WHERE "Job"."kind" = $kind AND "Job"."status" = $succeeded
		   AND json_extract("Job"."subject", '$.kind') = $subjectKind AND json_extract("Job"."subject", '$.id') = $subjectId
		   AND "MediaFile"."deletedAt" IS NULL
		 LIMIT 1`,
		db.Vars{"kind": jobs.KindDownload, "succeeded": jobs.StatusSucceeded, "subjectKind": subject.Kind, "subjectId": subject.Id})
	if err != nil || len(ids) == 0 {
		return "", err
	}
	return ids[0], nil
}

// recordTitle is a record's title, for the downloads list; "" when unknown.
func (s *Service) recordTitle(ctx context.Context, subject schema.RecordRef) string {
	titles, err := s.database.Strings(ctx, `SELECT "title" FROM `+db.Quote(mediaTable(subject.Kind))+` WHERE "id" = $id`,
		db.Vars{"id": subject.Id})
	if err != nil || len(titles) == 0 {
		return ""
	}
	return titles[0]
}

// DownloadTask runs download jobs: the stream is resolved and cached fully
// into the blob store (at the speed limit, held back while the user streams
// unless allowed), then recorded as a file of the record. The payload's
// download_url keeps the blob out of cache eviction.
func (s *Service) DownloadTask() jobs.Task {
	return func(ctx context.Context, job *jobs.Job, progress func(float64)) error {
		if job.Subject == nil {
			return fmt.Errorf("a download needs a subject")
		}
		if s.isHeldBack(ctx) {
			return s.deferDownload(ctx, job)
		}
		if s.config.Resolver == nil || s.config.StreamCache == nil {
			return fmt.Errorf("downloads need a stream resolver and cache")
		}
		return s.download(ctx, job, progress)
	}
}

// isHeldBack reports whether downloads must wait for playback to stop.
func (s *Service) isHeldBack(ctx context.Context) bool {
	if s.config.Settings == nil || s.config.StreamCache == nil {
		return false
	}
	if s.config.Settings.Cached().AllowDownloadsWhileStreaming {
		return false
	}
	return s.config.StreamCache.StreamActive(ctx)
}

// deferDownload puts a running download back in the queue for later. The
// claim counted an attempt; it is taken back, so a long playback doesn't use
// up the retries.
func (s *Service) deferDownload(ctx context.Context, job *jobs.Job) error {
	if err := s.config.Queue.Defer(ctx, job.ID, downloadDeferDelay); err != nil {
		return err
	}
	_, err := s.database.Exec(ctx, `UPDATE "Job" SET "attempts" = MAX("attempts" - 1, 0) WHERE "id" = $id`,
		db.Vars{"id": job.ID})
	return err
}

// download resolves, caches and records one download.
func (s *Service) download(ctx context.Context, job *jobs.Job, progress func(float64)) error {
	streamRow, err := s.downloadStream(ctx, *job.Subject, payloadString(job, "stream"))
	if err != nil {
		return err
	}
	result, err := s.config.Resolver.Resolve(ctx, streamRow.Url)
	if err != nil {
		return fmt.Errorf("resolve: %w", err)
	}
	headers := mergeHeaders(streamRow.Headers, result.Headers)
	blobURL, err := s.config.StreamCache.CacheNow(ctx, result, headers, s.downloadProgress(ctx, progress))
	if err != nil {
		return fmt.Errorf("cache to the blob store: %w", err)
	}
	file, err := s.recordDownload(ctx, *job.Subject, *streamRow, blobURL)
	if err != nil {
		return err
	}
	changes := map[string]any{"download_url": blobURL, "file": file.Id, "mime_type": file.MimeType, "stream": streamRow.Id}
	if err := s.setPayload(ctx, job, changes); err != nil {
		return err
	}
	s.fingerprintDownload(ctx, *file)
	return nil
}

// downloadStream is the stream a download fetches: the one named, else the
// subject's tallest, then fastest.
func (s *Service) downloadStream(ctx context.Context, subject schema.RecordRef, streamID string) (*schema.Stream, error) {
	vars := db.Vars{"kind": subject.Kind, "id": subject.Id, "stream": streamID}
	query := `SELECT * FROM "Stream" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL
		ORDER BY COALESCE("height", 0) DESC, COALESCE("speedKbps", 0) DESC, "createdAt" LIMIT 1`
	if streamID != "" {
		query = `SELECT * FROM "Stream" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "id" = $stream AND "deletedAt" IS NULL`
	}
	found, err := db.QueryOneAs[schema.Stream](ctx, s.database, query, vars)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return nil, fmt.Errorf("%s %s has no stream to download", subject.Kind, subject.Id)
	}
	return found, nil
}

// mergeHeaders is the stream's headers with the resolver's on top.
func mergeHeaders(streamHeaders, resolvedHeaders map[string]string) map[string]string {
	merged := map[string]string{}
	for name, value := range streamHeaders {
		merged[name] = value
	}
	for name, value := range resolvedHeaders {
		merged[name] = value
	}
	return merged
}

// downloadProgress is the stream cache's progress callback for a download:
// it paces the transfer to the speed limit and reports progress (0..1).
func (s *Service) downloadProgress(ctx context.Context, progress func(float64)) func(float64, int64, int64) {
	limiter := newPacer(s.speedLimitKBps())
	reporter := &steppedProgress{report: progress}
	return func(percent float64, received, total int64) {
		limiter.wait(ctx, received)
		reporter.set(percent / 100)
	}
}

// speedLimitKBps is the download speed limit setting; 0 = unlimited.
func (s *Service) speedLimitKBps() int {
	if s.config.Settings == nil {
		return 0
	}
	return s.config.Settings.Cached().DownloadSpeedLimitKBps
}

// recordDownload records a cached blob as a file of subject in this
// device's cache (or library) location, reusing the row of a blob already
// recorded.
func (s *Service) recordDownload(ctx context.Context, subject schema.RecordRef, streamRow schema.Stream, blobURL string) (*schema.MediaFile, error) {
	key, ok := storage.KeyFromURL(blobURL)
	if !ok {
		return nil, fmt.Errorf("download landed outside the blob store: %s", blobURL)
	}
	location, err := s.downloadLocation(ctx)
	if err != nil {
		return nil, err
	}
	existing, err := db.QueryOneAs[schema.MediaFile](ctx, s.database,
		`SELECT * FROM "MediaFile" WHERE "location" = $location AND "path" = $path AND "deletedAt" IS NULL`,
		db.Vars{"location": location.Id, "path": key})
	if err != nil || existing != nil {
		return existing, err
	}
	row := schema.MediaFile{
		SubjectKind: subject.Kind, SubjectId: subject.Id, Location: location.Id, Path: key,
		SizeBytes: s.blobSize(ctx, key, maxManifestDepth), MimeType: blobMimeType(key),
		Quality: streamRow.Quality, Height: streamRow.Height,
	}
	row.Id = db.NewID()
	if err := s.insertFile(ctx, row); err != nil {
		return nil, err
	}
	return db.Get[schema.MediaFile](ctx, s.database, tableMediaFile, row.Id)
}

// downloadLocation is this device's blob-store location downloads are
// recorded in: the cache one, else the library one.
func (s *Service) downloadLocation(ctx context.Context) (*schema.StorageLocation, error) {
	locations, err := s.config.Devices.Locations(ctx, "")
	if err != nil {
		return nil, err
	}
	var library *schema.StorageLocation
	for index := range locations {
		location := &locations[index]
		if !s.isBlobStore(*location) {
			continue
		}
		if location.Kind == schema.StorageKindCache {
			return location, nil
		}
		if location.Kind == schema.StorageKindLibrary {
			library = location
		}
	}
	if library == nil {
		return nil, fmt.Errorf("this device has no storage location for the blob store %s", s.config.BlobDir)
	}
	return library, nil
}

// blobMimeType is the type of a cached stream blob: an HLS manifest or MP4.
func blobMimeType(key string) string {
	if strings.HasSuffix(key, ".m3u8") {
		return "application/x-mpegURL"
	}
	return "video/mp4"
}

// blobSize is the stored size of a blob, for a manifest including the blobs
// it references (depth levels deep).
func (s *Service) blobSize(ctx context.Context, key string, depth int) int64 {
	listed, err := s.config.Store.List(ctx, key)
	if err != nil || len(listed) != 1 {
		return 0
	}
	size := listed[0].Size
	if depth == 0 || !strings.HasSuffix(key, ".m3u8") {
		return size
	}
	for _, reference := range s.manifestKeys(ctx, key) {
		size += s.blobSize(ctx, reference, depth-1)
	}
	return size
}

// manifestKeys lists the blob keys a cached manifest references.
func (s *Service) manifestKeys(ctx context.Context, key string) []string {
	content, err := s.config.Store.Get(ctx, key)
	if err != nil {
		return nil
	}
	defer content.Close()
	data, err := io.ReadAll(content)
	if err != nil {
		return nil
	}
	var keys []string
	for _, line := range strings.Split(string(data), "\n") {
		reference, ok := storage.KeyFromURL(strings.TrimSpace(line))
		if ok {
			keys = append(keys, reference)
		}
	}
	return keys
}

// setPayload merges changes into a job's payload. Writing the column
// replaces it, so the rest is carried over.
func (s *Service) setPayload(ctx context.Context, job *jobs.Job, changes map[string]any) error {
	payload := map[string]any{}
	for key, value := range job.Payload {
		payload[key] = value
	}
	for key, value := range changes {
		payload[key] = value
	}
	job.Payload = payload
	return s.database.Update(ctx, tableJob, job.ID, db.Vars{"payload": payload})
}

// fingerprintDownload queues fingerprinting a downloaded MP4. Manifests
// aren't media files, so HLS downloads get none.
func (s *Service) fingerprintDownload(ctx context.Context, file schema.MediaFile) {
	if file.MimeType != "video/mp4" {
		return
	}
	if _, err := s.FingerprintFile(ctx, file.Id); err != nil {
		log.Printf("download %s: queue fingerprint: %v", file.Id, err)
	}
}

// Downloads lists every download job, newest first, with its file.
func (s *Service) Downloads(ctx context.Context) ([]Download, error) {
	downloadJobs, err := db.QueryAs[jobs.Job](ctx, s.database,
		`SELECT * FROM "Job" WHERE "kind" = $kind ORDER BY "createdAt" DESC, rowid DESC`, db.Vars{"kind": jobs.KindDownload})
	if err != nil {
		return nil, err
	}
	downloads := make([]Download, 0, len(downloadJobs))
	for _, job := range downloadJobs {
		file, err := s.jobFile(ctx, job)
		if err != nil {
			return nil, err
		}
		downloads = append(downloads, Download{Job: job, File: file})
	}
	return downloads, nil
}

// jobFile is the file a download job recorded; nil before it finished.
func (s *Service) jobFile(ctx context.Context, job jobs.Job) (*schema.MediaFile, error) {
	fileID := payloadString(&job, "file")
	if fileID == "" {
		return nil, nil
	}
	return db.Get[schema.MediaFile](ctx, s.database, tableMediaFile, fileID)
}

// DeleteDownload removes a download: its job, its file row and the blob. A
// running download has to be cancelled (jobs.Runner.Cancel) first.
func (s *Service) DeleteDownload(ctx context.Context, jobID string) error {
	job, err := s.config.Queue.Get(ctx, jobID)
	if err != nil {
		return err
	}
	if job == nil || job.Kind != jobs.KindDownload {
		return fmt.Errorf("no download %s", jobID)
	}
	if job.Status == jobs.StatusRunning {
		return fmt.Errorf("download %s is running; cancel it first", jobID)
	}
	file, err := s.jobFile(ctx, *job)
	if err != nil {
		return err
	}
	if file != nil {
		if err := s.removeDownloadedFile(ctx, *file); err != nil {
			return err
		}
	}
	return s.config.Queue.Delete(ctx, jobID)
}

// removeDownloadedFile deletes a downloaded blob with its row and
// fingerprints. Segments of an HLS download stay as plain cache blobs, which
// eviction removes once unpinned.
func (s *Service) removeDownloadedFile(ctx context.Context, file schema.MediaFile) error {
	location, err := s.location(ctx, file.Location)
	if err != nil {
		return err
	}
	if !s.isBlobStore(*location) {
		return fmt.Errorf("file %s is not a download", file.Id)
	}
	if err := s.config.Store.Delete(ctx, file.Path); err != nil {
		return err
	}
	return s.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM "Fingerprint" WHERE "fileId" = $id`, db.Vars{"id": file.Id}); err != nil {
			return err
		}
		return tx.Delete(ctx, tableMediaFile, file.Id)
	})
}
