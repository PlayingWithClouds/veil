package files

import (
	"context"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// matchingSubject returns the record that already has a file with one of
// hashes (same kind of media), or nil.
func (s *Service) matchingSubject(ctx context.Context, kind schema.RecordKind, hashes []hashValue) (*schema.RecordRef, error) {
	for _, hash := range hashes {
		candidates, err := db.QueryAs[schema.Fingerprint](ctx, s.database,
			`SELECT * FROM "Fingerprint" WHERE "algorithm" = $algorithm AND "value" = $value
			 AND "subjectKind" = $kind AND "deletedAt" IS NULL ORDER BY "createdAt"`,
			db.Vars{"algorithm": hash.algorithm, "value": hash.value, "kind": kind})
		if err != nil {
			return nil, err
		}
		subject, err := s.firstExisting(ctx, candidates)
		if err != nil || subject != nil {
			return subject, err
		}
	}
	return nil, nil
}

// firstExisting returns the subject of the first fingerprint whose record
// is still in the library (not trashed), or nil.
func (s *Service) firstExisting(ctx context.Context, fingerprints []schema.Fingerprint) (*schema.RecordRef, error) {
	for _, fingerprint := range fingerprints {
		subject := schema.RecordRef{Kind: fingerprint.SubjectKind, Id: fingerprint.SubjectId}
		exists, err := s.recordExists(ctx, subject)
		if err != nil {
			return nil, err
		}
		if exists {
			return &subject, nil
		}
	}
	return nil, nil
}

// createRecord creates the record a new file stands for: a video or audio
// titled after the file, or an image in its folder's gallery.
func (s *Service) createRecord(ctx context.Context, opened *openedRoot, file foundFile, fileID string) (*schema.RecordRef, error) {
	switch file.mediaType.kind {
	case schema.RecordKindVideo:
		return s.createMedia(ctx, schema.RecordKindVideo, titleFromName(path.Base(file.path)))
	case schema.RecordKindAudio:
		return s.createMedia(ctx, schema.RecordKindAudio, titleFromName(path.Base(file.path)))
	case schema.RecordKindImage:
		return s.createImage(ctx, opened, file, fileID)
	}
	return nil, fmt.Errorf("no record for a %s file", file.mediaType.kind)
}

// createMedia inserts an available video or audio with title. The system
// owns the title, so a scraper that later identifies the file replaces it.
func (s *Service) createMedia(ctx context.Context, kind schema.RecordKind, title string) (*schema.RecordRef, error) {
	id, err := s.database.Insert(ctx, mediaTable(kind), db.Vars{
		"title":      title,
		"status":     schema.MediaStatusAvailable,
		"performers": []string{},
		"tags":       []string{},
	})
	if err != nil {
		return nil, err
	}
	if err := s.recorder.Claim(ctx, kind, id, []string{"title"}, provenance.System()); err != nil {
		return nil, err
	}
	return &schema.RecordRef{Kind: kind, Id: id}, nil
}

// createImage adds an image to the gallery of the file's folder, creating
// the gallery for the folder's first image. The image's asset is the file.
func (s *Service) createImage(ctx context.Context, opened *openedRoot, file foundFile, fileID string) (*schema.RecordRef, error) {
	gallery, err := s.folderGallery(ctx, opened, path.Dir(file.path))
	if err != nil {
		return nil, err
	}
	asset := schema.Asset{Uri: s.fileRouteURL(fileID), MimeType: &file.mediaType.mimeType}
	imageID, err := s.database.Insert(ctx, "Image", db.Vars{"gallery": gallery.Id, "position": gallery.ImageCount, "asset": asset})
	if err != nil {
		return nil, err
	}
	fields := db.Vars{"imageCount": gallery.ImageCount + 1}
	if gallery.Cover == nil {
		fields["cover"] = asset
	}
	if _, err := s.recorder.Update(ctx, schema.RecordKindGallery, gallery.Id, fields, provenance.System()); err != nil {
		return nil, err
	}
	return &schema.RecordRef{Kind: schema.RecordKindImage, Id: imageID}, nil
}

// folderGallery returns the gallery the images of a folder (location path)
// belong to, creating it titled after the folder.
func (s *Service) folderGallery(ctx context.Context, opened *openedRoot, folder string) (*schema.Gallery, error) {
	folderPrefix := ""
	if folder != "." {
		folderPrefix = folder + "/"
	}
	// rtrim(path, path without slashes) leaves the folder with its trailing slash.
	gallery, err := db.QueryOneAs[schema.Gallery](ctx, s.database,
		`SELECT "Gallery".* FROM "MediaFile"
		 JOIN "Image" ON "Image"."id" = "MediaFile"."subjectId"
		 JOIN "Gallery" ON "Gallery"."id" = "Image"."gallery"
		 WHERE "MediaFile"."subjectKind" = 'image' AND "MediaFile"."location" = $location
		   AND rtrim("MediaFile"."path", replace("MediaFile"."path", '/', '')) = $folder
		   AND "MediaFile"."deletedAt" IS NULL AND "Gallery"."deletedAt" IS NULL
		 LIMIT 1`,
		db.Vars{"location": opened.location.Id, "folder": folderPrefix})
	if err != nil || gallery != nil {
		return gallery, err
	}
	return s.createGallery(ctx, galleryTitle(opened, folder))
}

// galleryTitle names a folder's gallery after the folder (the root's own
// folder when the images lie at the top).
func galleryTitle(opened *openedRoot, folder string) string {
	if folder == "." || folder == "" {
		folder = opened.directory
	}
	return titleFromName(path.Base(strings.ReplaceAll(folder, "\\", "/")))
}

// createGallery inserts an empty available gallery titled title.
func (s *Service) createGallery(ctx context.Context, title string) (*schema.Gallery, error) {
	id, err := s.database.Insert(ctx, "Gallery", db.Vars{
		"title":      title,
		"status":     schema.MediaStatusAvailable,
		"performers": []string{},
		"tags":       []string{},
		"imageCount": 0,
	})
	if err != nil {
		return nil, err
	}
	if err := s.recorder.Claim(ctx, schema.RecordKindGallery, id, []string{"title"}, provenance.System()); err != nil {
		return nil, err
	}
	return db.Get[schema.Gallery](ctx, s.database, "Gallery", id)
}

// separatorPattern matches the characters file names use instead of spaces.
var separatorPattern = regexp.MustCompile(`[._\s]+`)

// titleFromName turns a file name into a title: no extension, separators
// as single spaces.
func titleFromName(name string) string {
	base := strings.TrimSuffix(name, path.Ext(name))
	title := strings.TrimSpace(separatorPattern.ReplaceAllString(base, " "))
	if title == "" {
		return name
	}
	return title
}
