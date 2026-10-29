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

// identifyFrame identifies the face in a picture sent as base64: a video frame
// the user picked or a picture of their own.
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
	posts := make([]*model.FacePostMatch, 0, len(identification.Posts))
	for _, post := range identification.Posts {
		posts = append(posts, facePostModel(post))
	}
	postPerformers := make([]*model.FacePostPerformer, 0, len(identification.PostPerformers))
	for _, performer := range identification.PostPerformers {
		postPerformers = append(postPerformers, facePostPerformerModel(performer))
	}
	return &model.FaceIdentification{
		Candidates:     candidates,
		Videos:         videos,
		Posts:          posts,
		PostPerformers: postPerformers,
		Sources:        nonNilStrings(identification.Sources),
		Failures:       nonNilStrings(identification.Failures),
		PictureURL:     strPtrIfSet(identification.PictureURL),
	}
}

// facePostModel converts one matching forum post; an empty thumbnail becomes null.
func facePostModel(post face.PostMatch) *model.FacePostMatch {
	return &model.FacePostMatch{
		Title:        post.Title,
		URL:          post.URL,
		ThumbnailURL: strPtrIfSet(post.ThumbnailURL),
		Solved:       post.Solved,
		Names:        nonNilStrings(post.Names),
		Sources:      nonNilStrings(post.Sources),
	}
}

// facePostPerformerModel converts one performer named by matching posts; empty optional fields become null.
func facePostPerformerModel(performer face.PostPerformer) *model.FacePostPerformer {
	return &model.FacePostPerformer{
		Name:        performer.Name,
		Posts:       performer.Posts,
		ImageURL:    strPtrIfSet(performer.ImageURL),
		PostURL:     performer.PostURL,
		PerformerID: strPtrIfSet(performer.PerformerID),
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

// forumThreadModel converts a forum thread for the API; empty optional fields become null.
func forumThreadModel(thread *face.Thread) *model.ForumThread {
	comments := make([]*model.ForumComment, 0, len(thread.MarkedComments))
	for _, comment := range thread.MarkedComments {
		comments = append(comments, forumCommentModel(comment))
	}
	return &model.ForumThread{
		URL:         thread.URL,
		Title:       thread.Title,
		Description: strPtrIfSet(thread.Description),
		ImageURL:    strPtrIfSet(thread.ImageURL),
		VideoURL:    strPtrIfSet(thread.VideoURL),
		Author:      strPtrIfSet(thread.Author),
		Date:        strPtrIfSet(thread.Date),
		Tags:        nonNilStrings(thread.Tags),
		Solved:      thread.Solved,
		Plugin:      thread.Plugin,
		Comments:    comments,
	}
}

// forumCommentModel converts one comment with its marked links and names.
func forumCommentModel(comment face.ThreadComment) *model.ForumComment {
	links := make([]*model.ForumLink, 0, len(comment.MarkedLinks))
	for _, link := range comment.MarkedLinks {
		links = append(links, &model.ForumLink{URL: link.URL, ScenePlugin: strPtrIfSet(link.ScenePlugin)})
	}
	names := make([]*model.ForumName, 0, len(comment.MarkedNames))
	for _, name := range comment.MarkedNames {
		names = append(names, &model.ForumName{Name: name.Name, PerformerID: strPtrIfSet(name.PerformerID)})
	}
	posts := make([]*model.ForumPostRef, 0, len(comment.Posts))
	for _, post := range comment.Posts {
		posts = append(posts, &model.ForumPostRef{Title: post.Title, URL: post.URL, ThumbnailURL: strPtrIfSet(post.ThumbnailURL)})
	}
	return &model.ForumComment{
		Posts:    posts,
		ID:       comment.ID,
		Author:   strPtrIfSet(comment.Author),
		Date:     strPtrIfSet(comment.Date),
		Text:     comment.Text,
		Accepted: comment.Accepted,
		Likes:    comment.Likes,
		Names:    names,
		Titles:   nonNilStrings(comment.Titles),
		Links:    links,
		Images:   nonNilStrings(comment.Images),
	}
}
