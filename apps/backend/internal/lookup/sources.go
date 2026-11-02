package lookup

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// askFace runs a plugin's face:identify: the people it names become
// performers, the similar videos video stubs, and solved forum posts that
// match the picture name performers with the post as evidence.
func (s *Service) askFace(ctx context.Context, plugin *plugins.Plugin, query Query) ([]finding, error) {
	mimeType := query.MimeType
	if mimeType == "" {
		mimeType = "image/jpeg"
	}
	result, err := s.runner.IdentifyFace(ctx, plugin, plugins.FaceIdentifyArgs{ImageBase64: encodeImage(query.Image), MimeType: mimeType})
	if err != nil {
		return nil, err
	}
	found := []finding{}
	for _, candidate := range result.Candidates {
		found = append(found, s.faceCandidate(ctx, plugin.Meta.Name, candidate)...)
	}
	for _, video := range result.Videos {
		found = append(found, s.similarVideo(ctx, plugin.Meta.Name, video)...)
	}
	for _, post := range result.Posts {
		found = append(found, s.postAnswer(ctx, plugin.Meta.Name, post.URL, post.Title, post.ThumbnailURL, post.Names)...)
	}
	return found, nil
}

// faceCandidate stores a named face match as a performer.
func (s *Service) faceCandidate(ctx context.Context, plugin string, candidate plugins.FaceCandidate) []finding {
	item := plugins.DiscoveredItem{MediaType: plugins.MediaTypePerformer, Title: candidate.Name, PosterPath: candidate.ImageURL}
	profile := ""
	if len(candidate.ProfileURLs) > 0 {
		profile = candidate.ProfileURLs[0]
		item.SourceURL = profile
	}
	ref, err := s.library.IngestListed(ctx, plugin, item)
	if err != nil || ref.Id == "" {
		return nil
	}
	detail := fmt.Sprintf("%d%% face match", int(candidate.Similarity*100))
	return []finding{{ref: ref, score: candidate.Similarity, evidence: evidenceOf(plugin, plugins.LookupEvidence{
		URL: profile, Title: candidate.Name, ImageURL: candidate.ImageURL, Detail: detail,
	})}}
}

// similarVideo stores a tube video whose face looks like the picture's, as
// a stub of the plugin that reads its site. Without one it can't be stored.
func (s *Service) similarVideo(ctx context.Context, plugin string, video plugins.FaceVideoMatch) []finding {
	ref, ok := s.storeVideoLink(ctx, video.URL, video.Title, video.ThumbnailURL)
	if !ok {
		return nil
	}
	detail := fmt.Sprintf("%d%% similar face", int(video.Similarity*100))
	return []finding{{ref: ref, score: video.Similarity, evidence: evidenceOf(plugin, plugins.LookupEvidence{
		URL: video.URL, Title: video.Title, ImageURL: video.ThumbnailURL, Detail: detail,
	})}}
}

// postAnswer turns the names a solved forum answer gives into performers
// with the thread as evidence.
func (s *Service) postAnswer(ctx context.Context, plugin, url, title, image string, names []string) []finding {
	found := []finding{}
	for _, name := range names {
		ref, err := s.library.IngestListed(ctx, plugin, plugins.DiscoveredItem{MediaType: plugins.MediaTypePerformer, Title: name})
		if err != nil || ref.Id == "" {
			continue
		}
		found = append(found, finding{ref: ref, score: answerScore, evidence: evidenceOf(plugin, plugins.LookupEvidence{
			URL: url, Title: title, ImageURL: image, Detail: "accepted answer",
		})})
	}
	return found
}

// askForum searches an identification forum for the text and reads the best
// threads: the names accepted answers give become performers, the video
// links they give video stubs.
func (s *Service) askForum(ctx context.Context, plugin *plugins.Plugin, text string) ([]finding, error) {
	result, err := s.runner.SearchForum(ctx, plugin, plugins.ForumSearchArgs{Query: text, Limit: forumThreads})
	if err != nil {
		return nil, err
	}
	found := []finding{}
	for index, hit := range result.Posts {
		if index >= forumThreads || !hit.Solved {
			continue
		}
		thread, err := s.runner.FindForumPost(ctx, plugin, hit.URL)
		if err != nil {
			continue
		}
		found = append(found, s.threadAnswers(ctx, plugin.Meta.Name, thread)...)
	}
	return found, nil
}

// threadAnswers collects what a thread's accepted comments name and link.
func (s *Service) threadAnswers(ctx context.Context, plugin string, thread *plugins.ForumPost) []finding {
	found := []finding{}
	for _, comment := range thread.Comments {
		if !comment.Accepted {
			continue
		}
		found = append(found, s.postAnswer(ctx, plugin, thread.URL, thread.Title, thread.ImageURL, comment.Names)...)
		for _, link := range comment.Links {
			ref, ok := s.storeVideoLink(ctx, link, firstOf(comment.Titles), "")
			if !ok {
				continue
			}
			found = append(found, finding{ref: ref, score: answerScore, evidence: evidenceOf(plugin, plugins.LookupEvidence{
				URL: thread.URL, Title: thread.Title, ImageURL: thread.ImageURL, Detail: "linked in the accepted answer",
			})})
		}
	}
	return found
}

// storeVideoLink stores a video page URL as a stub of the active plugin that
// reads that site's videos.
func (s *Service) storeVideoLink(ctx context.Context, url, title, thumbnail string) (schema.RecordRef, bool) {
	for _, plugin := range s.registry.WithCapability(plugins.CapabilitySceneFind) {
		if !plugin.ServesURL(url) {
			continue
		}
		if strings.TrimSpace(title) == "" {
			title = url
		}
		ref, err := s.library.IngestListed(ctx, plugin.Meta.Name, plugins.DiscoveredItem{
			MediaType: plugins.MediaTypeScene, SourceURL: url, Title: title, PosterPath: thumbnail,
		})
		return ref, err == nil && ref.Id != ""
	}
	return schema.RecordRef{}, false
}

// firstOf returns the first value, or "".
func firstOf(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
