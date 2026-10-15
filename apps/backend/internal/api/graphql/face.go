package graphql

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/face"
)

// SetFaceIdentifier wires in the "Who is this?" service.
func (r *Resolver) SetFaceIdentifier(service *face.Service) { r.faces = service }

// identifyFrame identifies the face in a video frame the user picked, sent as base64.
func (r *Resolver) identifyFrame(ctx context.Context, frameBase64 string) (*model.FaceIdentification, error) {
	frame, err := base64.StdEncoding.DecodeString(frameBase64)
	if err != nil {
		return nil, fmt.Errorf("frame is not base64: %w", err)
	}
	identification, err := r.faces.IdentifyPicture(ctx, frame)
	if err != nil {
		return nil, err
	}
	return faceIdentificationModel(identification), nil
}

// faceIdentificationModel converts an identification run for the API.
func faceIdentificationModel(identification *face.Identification) *model.FaceIdentification {
	candidates := make([]*model.FaceCandidate, 0, len(identification.Candidates))
	for _, candidate := range identification.Candidates {
		candidates = append(candidates, faceCandidateModel(candidate))
	}
	videos := make([]*model.FaceVideoMatch, 0, len(identification.Videos))
	for _, video := range identification.Videos {
		videos = append(videos, faceVideoModel(video))
	}
	return &model.FaceIdentification{
		Candidates: candidates,
		Videos:     videos,
		Sources:    nonNilStrings(identification.Sources),
		Failures:   nonNilStrings(identification.Failures),
		PictureURL: strPtrIfSet(identification.PictureURL),
	}
}

// faceCandidateModel converts one candidate; empty optional fields become null.
func faceCandidateModel(candidate face.Candidate) *model.FaceCandidate {
	converted := &model.FaceCandidate{
		Name:        candidate.Name,
		Similarity:  candidate.Similarity,
		ProfileUrls: nonNilStrings(candidate.ProfileURLs),
		Sources:     nonNilStrings(candidate.Sources),
	}
	if candidate.ImageURL != "" {
		converted.ImageURL = &candidate.ImageURL
	}
	if candidate.PerformerID != "" {
		converted.PerformerID = &candidate.PerformerID
	}
	return converted
}

// faceVideoModel converts one similar video; an empty thumbnail becomes null.
func faceVideoModel(video face.VideoMatch) *model.FaceVideoMatch {
	converted := &model.FaceVideoMatch{
		Title:      video.Title,
		Similarity: video.Similarity,
		URL:        video.URL,
		Sources:    nonNilStrings(video.Sources),
	}
	if video.ThumbnailURL != "" {
		converted.ThumbnailURL = &video.ThumbnailURL
	}
	if video.ScenePlugin != "" {
		converted.ScenePlugin = &video.ScenePlugin
	}
	return converted
}

// nonNilStrings returns values, or an empty list for nil (GraphQL non-null lists).
func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
