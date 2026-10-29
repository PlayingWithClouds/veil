package face

import (
	"context"
	"errors"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
)

// TestThreadMarksPlayableLinksAndLibraryNames checks that a thread is read
// through the plugin serving its URL, that links a scene plugin serves are
// marked playable, and that names the library has carry their performer.
func TestThreadMarksPlayableLinksAndLibraryNames(t *testing.T) {
	service, _, library, _ := testService(t, map[string]*plugins.FaceIdentifyResult{}, nil)
	ctx := context.Background()
	sceneID := ingestSceneWithPoster(t, library, "https://cdn.test/poster.jpg")
	if _, err := library.CreditPerformer(ctx, sceneID, "Siri"); err != nil {
		t.Fatal(err)
	}
	forum := &plugins.Plugin{Meta: plugins.PluginMeta{Name: "forum"}}
	service.threadPluginFor = func(threadURL string) *plugins.Plugin {
		if threadURL == "https://forum.test/post/1" {
			return forum
		}
		return nil
	}
	service.findPost = func(ctx context.Context, plugin *plugins.Plugin, threadURL string) (*plugins.ForumPost, error) {
		return &plugins.ForumPost{URL: threadURL, Title: "Who is she?", Comments: []plugins.ForumComment{
			{ID: "1", Text: "Siri, here", Names: []string{"Siri", "Nobody Known"}, Links: []string{"https://tube.test/v/1", "https://elsewhere.test/x"}},
		}}, nil
	}
	service.scenePluginFor = func(pageURL string) string {
		if pageURL == "https://tube.test/v/1" {
			return "tube"
		}
		return ""
	}

	thread, err := service.Thread(ctx, "https://forum.test/post/1")
	if err != nil {
		t.Fatal(err)
	}
	if thread.Plugin != "forum" || thread.Title != "Who is she?" || len(thread.MarkedComments) != 1 {
		t.Fatalf("thread = %+v", thread)
	}
	comment := thread.MarkedComments[0]
	if comment.MarkedLinks[0].ScenePlugin != "tube" || comment.MarkedLinks[1].ScenePlugin != "" {
		t.Errorf("links = %+v", comment.MarkedLinks)
	}
	if comment.MarkedNames[0].PerformerID == "" || comment.MarkedNames[1].PerformerID != "" {
		t.Errorf("names = %+v", comment.MarkedNames)
	}

	if _, err := service.Thread(ctx, "https://unknown.test/post/2"); err == nil {
		t.Error("a thread no plugin serves was read")
	}
}

// TestSearchThreadsSkipsFailingForums checks that hits keep their plugin and
// order, and that a failing forum leaves the others' hits.
func TestSearchThreadsSkipsFailingForums(t *testing.T) {
	service, _, _, _ := testService(t, map[string]*plugins.FaceIdentifyResult{}, nil)
	good := &plugins.Plugin{Meta: plugins.PluginMeta{Name: "forum"}}
	broken := &plugins.Plugin{Meta: plugins.PluginMeta{Name: "down"}}
	service.forums = func() []*plugins.Plugin { return []*plugins.Plugin{broken, good} }
	service.searchPosts = func(ctx context.Context, plugin *plugins.Plugin, args plugins.ForumSearchArgs) (*plugins.ForumSearchResult, error) {
		if plugin == broken {
			return nil, errors.New("site down")
		}
		if args.Query != "nicole" || args.Limit != 2 {
			t.Errorf("args = %+v", args)
		}
		return &plugins.ForumSearchResult{Posts: []plugins.ForumPostHit{
			{Title: "first", URL: "https://forum.test/post/1"},
			{Title: "second", URL: "https://forum.test/post/2", Solved: true},
		}}, nil
	}

	hits := service.SearchThreads(context.Background(), "nicole", 2)
	if len(hits) != 2 || hits[0].Title != "first" || hits[1].Plugin != "forum" || !hits[1].Solved {
		t.Fatalf("hits = %+v", hits)
	}
}
