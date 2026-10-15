package graphql

import (
	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/face"
)

// SetFaceIdentifier wires in the "Who is this?" service.
func (r *Resolver) SetFaceIdentifier(service *face.Service) { r.faces = service }

// faceIdentificationModel converts an identification run for the API.
func faceIdentificationModel(identification *face.Identification) *model.FaceIdentification {
	candidates := make([]*model.FaceCandidate, 0, len(identification.Candidates))
	for _, candidate := range identification.Candidates {
		candidates = append(candidates, faceCandidateModel(candidate))
	}
	return &model.FaceIdentification{
		Candidates: candidates,
		Sources:    nonNilStrings(identification.Sources),
		Failures:   nonNilStrings(identification.Failures),
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

// nonNilStrings returns values, or an empty list for nil (GraphQL non-null lists).
func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
