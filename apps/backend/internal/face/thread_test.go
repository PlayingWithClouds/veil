package face

import (
	"context"
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
