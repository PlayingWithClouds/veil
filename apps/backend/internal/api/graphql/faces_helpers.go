package graphql

import (
	"context"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/faces"
	"github.com/playingwithclouds/veil/internal/schema"
)

// faceScan turns a scan from the API into the service's.
func faceScan(input model.FaceScanInput) faces.Scan {
	scan := faces.Scan{Model: input.Model, Picture: input.Picture, Frames: input.Frames}
	if input.Subject != nil {
		scan.Subject = *input.Subject
	}
	for _, face := range input.Faces {
		found := faces.Face{Vector: float32s(face.Vector), AtSeconds: face.AtSeconds}
		if region, isObject := face.Region.(map[string]any); isObject {
			found.Region = region
		}
		scan.Faces = append(scan.Faces, found)
	}
	return scan
}

// facePictures lists a performer's pictures for the API.
func (r *Resolver) facePictures(ctx context.Context, performerID, faceModel string) ([]*model.FacePicture, error) {
	pictures, err := r.faces.Pictures(ctx, performerID, faceModel)
	if err != nil {
		return nil, err
	}
	listed := make([]*model.FacePicture, 0, len(pictures))
	for _, picture := range pictures {
		listed = append(listed, &model.FacePicture{URI: picture.URI, Scanned: picture.Scanned})
	}
	return listed, nil
}

// faceRecords loads the performers and videos guesses name, once each.
type faceRecords struct {
	resolver   *Resolver
	performers map[string]*schema.Performer
	videos     map[string]*schema.Video
}

// performer loads a performer, nil when it's gone.
func (records *faceRecords) performer(ctx context.Context, id string) (*schema.Performer, error) {
	if loaded, seen := records.performers[id]; seen {
		return loaded, nil
	}
	loaded, err := records.resolver.library.Performer(ctx, id)
	records.performers[id] = loaded
	return loaded, err
}

// video loads a video, nil when it's gone or not asked for.
func (records *faceRecords) video(ctx context.Context, id *string) (*schema.Video, error) {
	if id == nil {
		return nil, nil
	}
	if loaded, seen := records.videos[*id]; seen {
		return loaded, nil
	}
	loaded, err := records.resolver.library.Video(ctx, *id)
	records.videos[*id] = loaded
	return loaded, err
}

// defaultFaceMatches is how many face matches a performer lists by default.
const defaultFaceMatches = 40

// faceMatches loads the videos of a performer's face matches, skipping any
// gone since.
func (r *Resolver) faceMatches(ctx context.Context, found []faces.FaceMatch) ([]*model.FaceMatch, error) {
	matches := []*model.FaceMatch{}
	for _, match := range found {
		video, err := r.library.Video(ctx, match.VideoID)
		if err != nil {
			return nil, err
		}
		if video == nil {
			continue
		}
		matches = append(matches, &model.FaceMatch{Video: video, Similarity: match.Similarity})
	}
	return matches, nil
}

// videoFaces turns the service's people and guesses into the API's,
// loading the performers and videos they name.
func (r *Resolver) videoFaces(ctx context.Context, found *faces.VideoFaces) (*model.VideoFaces, error) {
	records := &faceRecords{resolver: r, performers: map[string]*schema.Performer{}, videos: map[string]*schema.Video{}}
	result := &model.VideoFaces{Scan: found.Scan, People: []*model.FacePerson{}, Guesses: []*model.FaceGuess{}}
	for index, person := range found.People {
		facePerson, err := records.person(ctx, person, found.Candidates[index])
		if err != nil {
			return nil, err
		}
		result.People = append(result.People, facePerson)
	}
	for _, guess := range found.Guesses {
		performer, err := records.performer(ctx, guess.PerformerID)
		if err != nil {
			return nil, err
		}
		if performer == nil {
			continue
		}
		result.Guesses = append(result.Guesses, &model.FaceGuess{Performer: performer, Similarity: guess.Similarity, Credited: guess.Credited, Person: guess.Person})
	}
	return result, nil
}

// person turns one person and their candidates into the API's.
func (records *faceRecords) person(ctx context.Context, person faces.Person, candidates []faces.Candidate) (*model.FacePerson, error) {
	result := &model.FacePerson{Sightings: []*model.FaceSighting{}, Candidates: []*model.FaceCandidate{}}
	for _, sighting := range person.Sightings {
		result.Sightings = append(result.Sightings, &model.FaceSighting{EmbeddingID: sighting.EmbeddingID, AtSeconds: sighting.AtSeconds, Region: sighting.Region})
	}
	for _, candidate := range candidates {
		performer, err := records.performer(ctx, candidate.PerformerID)
		if err != nil {
			return nil, err
		}
		if performer == nil {
			continue
		}
		video, err := records.video(ctx, candidate.VideoID)
		if err != nil {
			return nil, err
		}
		result.Candidates = append(result.Candidates, &model.FaceCandidate{
			Performer: performer, Similarity: candidate.Similarity, Likely: candidate.Likely,
			Credited: candidate.Credited, Picture: candidate.Picture, Video: video,
		})
	}
	return result, nil
}
