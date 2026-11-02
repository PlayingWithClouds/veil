// Package faces guesses who is in a video from the face vectors clients
// compute: the backend runs no model (it also runs embedded on a phone, and
// has no video decoder), so clients scan a video's sampled frames or a
// performer's picture and record the scan with its faces. From videos
// crediting a performer the backend learns that performer's face; a video's
// faces are grouped into people, and each person is compared with every
// performer's known faces.
package faces

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/embeddings"
	"github.com/playingwithclouds/veil/internal/schema"
)

// Tables this package writes.
const (
	tableScan      = "FaceScan"
	tableEmbedding = "Embedding"
)

// maxRelearnedVideos bounds how many of a performer's videos learn again
// when a new picture of theirs is scanned.
const maxRelearnedVideos = 200

// Service records face scans, learns performers' faces and guesses.
type Service struct {
	database *db.DB
}

// New returns the service on database.
func New(database *db.DB) *Service {
	return &Service{database: database}
}

// Face is one face a client found: its vector, frame time and box.
type Face struct {
	Vector    []float32
	AtSeconds *float64
	Region    map[string]any
}

// Scan is what a client recorded: the faces it found in a video's sampled
// frames, or in one picture of a performer.
type Scan struct {
	Subject schema.RecordRef
	Model   string
	Picture *string
	Frames  int
	Faces   []Face
}

// RecordScan stores a scan and its faces, replacing the subject's earlier
// scan with the model (and picture), then learns the faces of the
// performers it concerns: a video's credited performers, or every scanned
// video of the pictured performer.
func (s *Service) RecordScan(ctx context.Context, scan Scan) (*schema.FaceScan, error) {
	if err := s.validate(ctx, &scan); err != nil {
		return nil, err
	}
	var scanID string
	err := s.database.Tx(ctx, func(tx *db.DB) error {
		if err := removeScans(ctx, tx, scan); err != nil {
			return err
		}
		id, err := insertScan(ctx, tx, scan)
		scanID = id
		return err
	})
	if err != nil {
		return nil, err
	}
	if err := s.learnAfter(ctx, scan); err != nil {
		return nil, err
	}
	return db.Get[schema.FaceScan](ctx, s.database, tableScan, scanID)
}

// validate checks a scan and trims its model and picture: a video's scan
// has no picture, a performer's has one, and the record must exist.
func (s *Service) validate(ctx context.Context, scan *Scan) error {
	scan.Model = strings.TrimSpace(scan.Model)
	if scan.Model == "" {
		return fmt.Errorf("a face scan needs its model's name")
	}
	if scan.Picture != nil {
		trimmed := strings.TrimSpace(*scan.Picture)
		scan.Picture = &trimmed
	}
	switch scan.Subject.Kind {
	case schema.RecordKindVideo:
		if scan.Picture != nil {
			return fmt.Errorf("a video's face scan is of its frames, not a picture")
		}
	case schema.RecordKindPerformer:
		if scan.Picture == nil || *scan.Picture == "" {
			return fmt.Errorf("a performer's face scan names the picture")
		}
	default:
		return fmt.Errorf("faces are scanned in videos and performers' pictures, not %s", scan.Subject.Kind)
	}
	return s.requireRecord(ctx, scan.Subject)
}

// requireRecord fails unless the record exists and isn't trashed.
func (s *Service) requireRecord(ctx context.Context, subject schema.RecordRef) error {
	table := "Video"
	if subject.Kind == schema.RecordKindPerformer {
		table = "Performer"
	}
	row, err := s.database.QueryRow(ctx, `SELECT "id" FROM `+db.Quote(table)+` WHERE "id" = $id AND "deletedAt" IS NULL`, db.Vars{"id": subject.Id})
	if err != nil {
		return err
	}
	if row == nil {
		return fmt.Errorf("no %s %s", subject.Kind, subject.Id)
	}
	return nil
}

// removeScans deletes the subject's scans with the scan's model and
// picture, with every vector pointing at them (its faces, and performers'
// faces learned from them).
func removeScans(ctx context.Context, tx *db.DB, scan Scan) error {
	vars := db.Vars{"kind": scan.Subject.Kind, "id": scan.Subject.Id, "model": scan.Model, "picture": scan.Picture}
	matching := `SELECT "id" FROM "FaceScan" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "model" = $model AND "picture" IS $picture`
	if _, err := tx.Exec(ctx, `DELETE FROM "Embedding" WHERE "scan" IN (`+matching+`)`, vars); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `DELETE FROM "FaceScan" WHERE "id" IN (`+matching+`)`, vars)
	return err
}

// insertScan stores the scan and its faces, returning the scan's id.
func insertScan(ctx context.Context, tx *db.DB, scan Scan) (string, error) {
	scanID, err := tx.Insert(ctx, tableScan, schema.FaceScan{
		SubjectKind: scan.Subject.Kind, SubjectId: scan.Subject.Id, Model: scan.Model,
		Picture: scan.Picture, Frames: int32(scan.Frames), Faces: int32(len(scan.Faces)),
	})
	if err != nil {
		return "", err
	}
	for _, face := range scan.Faces {
		if len(face.Vector) == 0 {
			return "", fmt.Errorf("a face needs its vector")
		}
		row := db.Vars{
			"subjectKind": scan.Subject.Kind, "subjectId": scan.Subject.Id, "kind": schema.EmbeddingKindFace,
			"model": scan.Model, "vector": embeddings.Encode(face.Vector), "scan": scanID,
		}
		if face.AtSeconds != nil {
			row["atSeconds"] = *face.AtSeconds
		}
		if face.Region != nil {
			row["region"] = face.Region
		}
		if _, err := tx.Insert(ctx, tableEmbedding, row); err != nil {
			return "", err
		}
	}
	return scanID, nil
}

// learnAfter learns from a recorded scan: a video's credited performers,
// or each scanned video of a pictured performer.
func (s *Service) learnAfter(ctx context.Context, scan Scan) error {
	if scan.Subject.Kind == schema.RecordKindVideo {
		_, err := s.Learn(ctx, scan.Subject.Id, scan.Model)
		return err
	}
	videoIDs, err := s.scannedVideosOf(ctx, scan.Subject.Id, scan.Model)
	if err != nil {
		return err
	}
	for _, videoID := range videoIDs {
		if _, err := s.Learn(ctx, videoID, scan.Model); err != nil {
			return err
		}
	}
	return nil
}

// scannedVideosOf lists the videos crediting a performer that have a scan
// with the model, newest scan first.
func (s *Service) scannedVideosOf(ctx context.Context, performerID, model string) ([]string, error) {
	rows, err := s.database.Query(ctx,
		`SELECT "Video"."id" AS "id" FROM "Video" JOIN "FaceScan" ON "FaceScan"."subjectKind" = 'video' AND "FaceScan"."subjectId" = "Video"."id"
		 WHERE "FaceScan"."model" = $model AND "FaceScan"."deletedAt" IS NULL AND "Video"."deletedAt" IS NULL
		   AND $performer IN (SELECT value FROM json_each("Video"."performers"))
		 ORDER BY "FaceScan"."createdAt" DESC LIMIT $limit`,
		db.Vars{"model": model, "performer": performerID, "limit": maxRelearnedVideos})
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, fmt.Sprint(row["id"]))
	}
	return ids, nil
}

// Scans lists a record's face scans, newest first.
func (s *Service) Scans(ctx context.Context, subject schema.RecordRef) ([]schema.FaceScan, error) {
	return db.QueryAs[schema.FaceScan](ctx, s.database,
		`SELECT * FROM "FaceScan" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL ORDER BY "createdAt" DESC`,
		db.Vars{"kind": subject.Kind, "id": subject.Id})
}

// videoScan is a video's newest scan with the model, or with any model
// when model is empty; nil when it has none.
func (s *Service) videoScan(ctx context.Context, videoID, model string) (*schema.FaceScan, error) {
	return db.QueryOneAs[schema.FaceScan](ctx, s.database,
		`SELECT * FROM "FaceScan" WHERE "subjectKind" = 'video' AND "subjectId" = $id AND ($model = '' OR "model" = $model)
		   AND "deletedAt" IS NULL ORDER BY "createdAt" DESC LIMIT 1`,
		db.Vars{"id": videoID, "model": model})
}

// Picture is a picture of a performer to scan for their face.
type Picture struct {
	URI     string
	Scanned bool
}

// Pictures lists a performer's pictures (their image, then each site
// profile's) and whether each is scanned with the model.
func (s *Service) Pictures(ctx context.Context, performerID, model string) ([]Picture, error) {
	rows, err := s.database.Query(ctx,
		`SELECT json_extract("image", '$.uri') AS "uri" FROM "Performer" WHERE "id" = $id AND "deletedAt" IS NULL
		 UNION ALL
		 SELECT json_extract("image", '$.uri') FROM "Source" WHERE "subjectKind" = 'performer' AND "subjectId" = $id AND "deletedAt" IS NULL`,
		db.Vars{"id": performerID})
	if err != nil {
		return nil, err
	}
	scanned, err := s.scannedPictures(ctx, performerID, model)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	pictures := []Picture{}
	for _, row := range rows {
		uri, isText := row["uri"].(string)
		if !isText || uri == "" || seen[uri] {
			continue
		}
		seen[uri] = true
		pictures = append(pictures, Picture{URI: uri, Scanned: scanned[uri]})
	}
	return pictures, nil
}

// scannedPictures is the set of a performer's pictures scanned with the
// model.
func (s *Service) scannedPictures(ctx context.Context, performerID, model string) (map[string]bool, error) {
	scans, err := db.QueryAs[schema.FaceScan](ctx, s.database,
		`SELECT * FROM "FaceScan" WHERE "subjectKind" = 'performer' AND "subjectId" = $id AND "model" = $model AND "deletedAt" IS NULL`,
		db.Vars{"id": performerID, "model": model})
	if err != nil {
		return nil, err
	}
	scanned := map[string]bool{}
	for _, scan := range scans {
		if scan.Picture != nil {
			scanned[*scan.Picture] = true
		}
	}
	return scanned, nil
}
