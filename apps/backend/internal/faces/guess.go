package faces

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/embeddings"
	"github.com/playingwithclouds/veil/internal/schema"
)

// VideoFaces is who is probably in a video: its faces grouped into people,
// each with candidate performers, and the one-to-one guesses among them.
// Scan is nil when the video wasn't scanned (with the asked model).
type VideoFaces struct {
	Scan       *schema.FaceScan
	People     []Person
	Candidates [][]Candidate
	Guesses    []Guess
}

// VideoFaces groups a video's faces into people and compares each with
// every performer's known faces, leaving out what was learned from this
// video itself. An empty model takes the video's newest scan.
func (s *Service) VideoFaces(ctx context.Context, videoID, model string) (*VideoFaces, error) {
	scan, err := s.videoScan(ctx, videoID, model)
	if err != nil || scan == nil {
		return &VideoFaces{}, err
	}
	people, credited, err := s.videoPeople(ctx, videoID, scan)
	if err != nil {
		return nil, err
	}
	references, err := s.references(ctx, scan.Model, videoID)
	if err != nil {
		return nil, err
	}
	result := &VideoFaces{Scan: scan, People: people, Candidates: make([][]Candidate, len(people))}
	for index, person := range people {
		result.Candidates[index] = Candidates(person, references, credited)
	}
	result.Guesses = Assign(result.Candidates)
	return result, nil
}

// Learn works out which person of a scanned video each credited performer
// is and stores that person's mean face as the performer's, replacing what
// was learned from the video before; an empty model takes the newest scan.
// Returns how many performers were learned.
func (s *Service) Learn(ctx context.Context, videoID, model string) (int, error) {
	scan, err := s.videoScan(ctx, videoID, model)
	if err != nil || scan == nil {
		return 0, err
	}
	people, credited, err := s.videoPeople(ctx, videoID, scan)
	if err != nil {
		return 0, err
	}
	references, err := s.references(ctx, scan.Model, videoID)
	if err != nil {
		return 0, err
	}
	learned := Learn(people, references, credited)
	err = s.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM "Embedding" WHERE "scan" = $scan AND "subjectKind" = 'performer'`, db.Vars{"scan": scan.Id}); err != nil {
			return err
		}
		for person, performerID := range learned {
			if err := insertLearned(ctx, tx, scan, performerID, people[person].Centroid); err != nil {
				return err
			}
		}
		return nil
	})
	return len(learned), err
}

// insertLearned stores a performer's face learned from a video's scan.
func insertLearned(ctx context.Context, tx *db.DB, scan *schema.FaceScan, performerID string, centroid []float32) error {
	_, err := tx.Insert(ctx, tableEmbedding, db.Vars{
		"subjectKind": schema.RecordKindPerformer, "subjectId": performerID, "kind": schema.EmbeddingKindFace,
		"model": scan.Model, "vector": embeddings.Encode(centroid), "scan": scan.Id,
	})
	return err
}

// videoPeople loads a scan's faces grouped into people, and the video's
// credited performers.
func (s *Service) videoPeople(ctx context.Context, videoID string, scan *schema.FaceScan) ([]Person, []string, error) {
	video, err := db.Get[schema.Video](ctx, s.database, "Video", videoID)
	if err != nil {
		return nil, nil, err
	}
	if video == nil {
		return nil, nil, fmt.Errorf("no video %s", videoID)
	}
	stored, err := db.QueryAs[schema.Embedding](ctx, s.database,
		`SELECT * FROM "Embedding" WHERE "scan" = $scan AND "subjectKind" = 'video' AND "subjectId" = $id AND "deletedAt" IS NULL`,
		db.Vars{"scan": scan.Id, "id": videoID})
	if err != nil {
		return nil, nil, err
	}
	sightings := make([]Sighting, 0, len(stored))
	for _, embedding := range stored {
		sightings = append(sightings, Sighting{
			EmbeddingID: embedding.Id, Vector: embeddings.Decode(embedding.Vector),
			AtSeconds: embedding.AtSeconds, Region: embedding.Region,
		})
	}
	return Cluster(sightings, linkage), video.Performers, nil
}

// references loads every performer's known faces with the model: their
// pictures' vectors, and faces learned from videos that still credit them,
// except those learned from excludeVideo. Trashed performers are left out.
func (s *Service) references(ctx context.Context, model, excludeVideo string) ([]Reference, error) {
	rows, err := s.database.Query(ctx,
		`SELECT "Embedding"."subjectId" AS "performer", "Embedding"."vector" AS "vector",
		        "FaceScan"."subjectKind" AS "scanKind", "FaceScan"."subjectId" AS "scanSubject", "FaceScan"."picture" AS "picture"
		 FROM "Embedding"
		 JOIN "FaceScan" ON "FaceScan"."id" = "Embedding"."scan" AND "FaceScan"."deletedAt" IS NULL
		 JOIN "Performer" ON "Performer"."id" = "Embedding"."subjectId" AND "Performer"."deletedAt" IS NULL
		 WHERE "Embedding"."subjectKind" = 'performer' AND "Embedding"."kind" = 'face' AND "Embedding"."model" = $model
		   AND "Embedding"."deletedAt" IS NULL
		   AND ("FaceScan"."subjectKind" = 'performer' OR ("FaceScan"."subjectId" != $exclude AND EXISTS (
		     SELECT 1 FROM "Video" WHERE "Video"."id" = "FaceScan"."subjectId" AND "Video"."deletedAt" IS NULL
		       AND "Embedding"."subjectId" IN (SELECT value FROM json_each("Video"."performers")))))`,
		db.Vars{"model": model, "exclude": excludeVideo})
	if err != nil {
		return nil, err
	}
	references := make([]Reference, 0, len(rows))
	for _, row := range rows {
		references = append(references, referenceOf(row))
	}
	return references, nil
}

// referenceOf reads one row of references.
func referenceOf(row db.Row) Reference {
	reference := Reference{PerformerID: fmt.Sprint(row["performer"])}
	if vector, isBytes := row["vector"].([]byte); isBytes {
		reference.Vector = embeddings.Decode(vector)
	}
	if picture, isText := row["picture"].(string); isText {
		reference.Picture = &picture
	}
	if fmt.Sprint(row["scanKind"]) == string(schema.RecordKindVideo) {
		videoID := fmt.Sprint(row["scanSubject"])
		reference.VideoID = &videoID
	}
	return reference
}
