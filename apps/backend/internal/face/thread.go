package face

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// threadTimeout is how long reading one forum thread may take.
const threadTimeout = 30 * time.Second

// ThreadLink is a link a comment gives. ScenePlugin is the active scene
// plugin that serves the page, so the app can play it; empty when none does.
type ThreadLink struct {
	URL         string
	ScenePlugin string
}

// ThreadName is a performer a comment tags, with the library's performer of
// that name (empty when there is none).
type ThreadName struct {
	Name        string
	PerformerID string
}

// ThreadComment is one comment of a forum thread, its links and names marked.
type ThreadComment struct {
	plugins.ForumComment
	MarkedLinks []ThreadLink
	MarkedNames []ThreadName
}

// Thread is one forum thread (a post found by face search) as the app shows it.
type Thread struct {
	plugins.ForumPost
	// Plugin is the plugin that read the thread.
	Plugin string
	// MarkedComments are the post's comments in site order, links and names marked.
	MarkedComments []ThreadComment
}

// ThreadHit is a forum thread a text search found, with the plugin that found it.
type ThreadHit struct {
	plugins.ForumPostHit
	Plugin string
}

// SearchThreads asks every forum search plugin for threads matching query,
// concurrently, each with its own time limit. A plugin that fails is logged
// and left out: the threads only accompany a search. Hits keep each plugin's
// order, plugins in registry order, at most limit per plugin.
func (s *Service) SearchThreads(ctx context.Context, query string, limit int) []ThreadHit {
	forums := s.forums()
	answers := make([][]ThreadHit, len(forums))
	var group sync.WaitGroup
	for index, forum := range forums {
		group.Add(1)
		go func() {
			defer group.Done()
			searchContext, cancel := context.WithTimeout(ctx, threadTimeout)
			defer cancel()
			result, err := s.searchPosts(searchContext, forum, plugins.ForumSearchArgs{Query: query, Limit: limit})
			if err != nil {
				log.Printf("post:search via plugin %s: %v", forum.Meta.Name, err)
				return
			}
			for _, hit := range result.Posts {
				answers[index] = append(answers[index], ThreadHit{ForumPostHit: hit, Plugin: forum.Meta.Name})
			}
		}()
	}
	group.Wait()
	var hits []ThreadHit
	for _, answer := range answers {
		hits = append(hits, answer...)
	}
	return hits
}

// Thread reads the forum thread at threadURL through the post:find plugin
// serving it, marking each link the app can play and each name the library has.
func (s *Service) Thread(ctx context.Context, threadURL string) (*Thread, error) {
	plugin := s.threadPluginFor(threadURL)
	if plugin == nil {
		return nil, fmt.Errorf("no plugin reads threads of %s", threadURL)
	}
	threadContext, cancel := context.WithTimeout(ctx, threadTimeout)
	defer cancel()
	post, err := s.findPost(threadContext, plugin, threadURL)
	if err != nil {
		return nil, err
	}
	thread := &Thread{ForumPost: *post, Plugin: plugin.Meta.Name}
	for _, comment := range post.Comments {
		thread.MarkedComments = append(thread.MarkedComments, s.markComment(ctx, comment))
	}
	return thread, nil
}

// markComment names the scene plugin of each link and the library performer of each name.
func (s *Service) markComment(ctx context.Context, comment plugins.ForumComment) ThreadComment {
	marked := ThreadComment{ForumComment: comment}
	for _, link := range comment.Links {
		threadLink := ThreadLink{URL: link}
		if s.scenePluginFor != nil {
			threadLink.ScenePlugin = s.scenePluginFor(link)
		}
		marked.MarkedLinks = append(marked.MarkedLinks, threadLink)
	}
	for _, name := range comment.Names {
		marked.MarkedNames = append(marked.MarkedNames, ThreadName{Name: name, PerformerID: s.libraryPerformer(ctx, name)})
	}
	return marked
}
