package faces

import (
	"context"
	"slices"
	"sort"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/embeddings"
	"github.com/playingwithclouds/veil/internal/schema"
)

// FaceMatch is a scanned video that probably shows a performer it doesn't
// credit.
type FaceMatch struct {
	VideoID    string
	Similarity float64
}

// PerformerMatches finds the scanned videos not crediting a performer where
// they win one of the video's people in the same one-to-one guessing the
// video's own page does, so a face that's more like someone else isn't
// theirs. Best first, at most limit.
func (s *Service) PerformerMatches(ctx context.Context, performerID string, limit int) ([]FaceMatch, error) {
	models, err := s.database.Strings(ctx,
		`SELECT DISTINCT "model" FROM "Embedding"
		 WHERE "subjectKind" = 'performer' AND "subjectId" = $id AND "kind" = 'face' AND "deletedAt" IS NULL`,
		db.Vars{"id": performerID})
	if err != nil {
		return nil, err
	}
	best := map[string]float64{}
	for _, model := range models {
		if err := s.collectMatches(ctx, performerID, model, best); err != nil {
			return nil, err
		}
	}
	matches := make([]FaceMatch, 0, len(best))
	for videoID, similarity := range best {
		matches = append(matches, FaceMatch{VideoID: videoID, Similarity: similarity})
	}
	sort.Slice(matches, func(first, second int) bool {
		return matches[first].Similarity > matches[second].Similarity
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}
	return matches, nil
}

// collectMatches adds the performer's matches among the videos scanned with
// model to best, keeping each video's highest similarity.
func (s *Service) collectMatches(ctx context.Context, performerID, model string, best map[string]float64) error {
	references, err := s.references(ctx, model, "")
	if err != nil {
		return err
	}
	own := referencesOf(references, performerID)
	if len(own) == 0 {
		return nil
	}
	scans, err := db.QueryAs[schema.FaceScan](ctx, s.database,
		`SELECT "FaceScan".* FROM "FaceScan"
		 JOIN "Video" ON "Video"."id" = "FaceScan"."subjectId" AND "Video"."deletedAt" IS NULL
		 WHERE "FaceScan"."subjectKind" = 'video' AND "FaceScan"."model" = $model AND "FaceScan"."deletedAt" IS NULL
		   AND $performer NOT IN (SELECT value FROM json_each("Video"."performers"))`,
		db.Vars{"model": model, "performer": performerID})
	if err != nil {
		return err
	}
	for index := range scans {
		similarity, found, err := s.scanMatch(ctx, &scans[index], performerID, own, references)
		if err != nil {
			return err
		}
		if found && similarity > best[scans[index].SubjectId] {
			best[scans[index].SubjectId] = similarity
		}
	}
	return nil
}

// scanMatch is how alike the performer is to the person they win in a
// video's scan, if they win one.
func (s *Service) scanMatch(ctx context.Context, scan *schema.FaceScan, performerID string, own, references []Reference) (float64, bool, error) {
	people, credited, err := s.videoPeople(ctx, scan.SubjectId, scan)
	if err != nil {
		return 0, false, err
	}
	// Most videos can't show them at all; skip ranking everyone there.
	if !anyLikely(people, own) {
		return 0, false, nil
	}
	// As on the video's page: faces learned from the video itself don't count.
	others := referencesBesides(references, scan.SubjectId)
	candidates := make([][]Candidate, len(people))
	for index, person := range people {
		candidates[index] = Candidates(person, others, credited)
	}
	for _, guess := range Assign(candidates) {
		if guess.PerformerID == performerID {
			return guess.Similarity, true, nil
		}
	}
	return 0, false, nil
}

// referencesOf keeps one performer's known faces.
func referencesOf(references []Reference, performerID string) []Reference {
	var kept []Reference
	for _, reference := range references {
		if reference.PerformerID == performerID {
			kept = append(kept, reference)
		}
	}
	return kept
}

// referencesBesides leaves out the faces learned from a video.
func referencesBesides(references []Reference, videoID string) []Reference {
	kept := make([]Reference, 0, len(references))
	for _, reference := range references {
		if reference.VideoID == nil || *reference.VideoID != videoID {
			kept = append(kept, reference)
		}
	}
	return kept
}

// anyLikely reports whether any person likely matches any of the faces.
func anyLikely(people []Person, references []Reference) bool {
	return slices.ContainsFunc(people, func(person Person) bool {
		return slices.ContainsFunc(references, func(reference Reference) bool {
			similarity, comparable := embeddings.Cosine(person.Centroid, reference.Vector)
			return comparable && similarity >= likelySimilarity
		})
	})
}
