package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestBranchingStory links the chapters a story page lists to their
// parents, then adds the choices found while reading one as its children,
// leaving a chapter already known under its first parent.
func TestBranchingStory(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	story := &plugins.Story{ExternalID: "cave", Title: "The cave", Chapters: []plugins.StoryChapter{
		{Title: "Start", URL: "https://cyoa.test/1", Text: "You stand at the mouth of a cave."},
		{URL: "https://cyoa.test/2", ParentURL: "https://cyoa.test/1", Choice: "Go in"},
		{URL: "https://cyoa.test/3", ParentURL: "https://cyoa.test/1", Choice: "Walk away"},
	}}
	ref, err := library.IngestFound(ctx, "site", []*plugins.ScrapeResult{{Type: plugins.MediaTypeStory, Story: story}})
	if err != nil {
		t.Fatal(err)
	}
	chapters, _ := library.Chapters(ctx, ref.Id)
	if len(chapters) != 3 || chapters[0].Parent != nil || !childOf(chapters[1], chapters[0], "Go in") || !childOf(chapters[2], chapters[0], "Walk away") {
		t.Fatalf("chapters = %+v", chapters)
	}

	read := plugins.StoryChapter{Text: "It's dark inside.", Choices: []plugins.StoryChapter{
		{URL: "https://cyoa.test/4", Choice: "Light a torch"},
		{URL: "https://cyoa.test/3", Choice: "Turn back"},
	}}
	if err := library.SetChapterText(ctx, chapters[1].Id, read); err != nil {
		t.Fatal(err)
	}
	chapters, _ = library.Chapters(ctx, ref.Id)
	if len(chapters) != 4 || chapters[3].Position != 3 || !childOf(chapters[3], chapters[1], "Light a torch") {
		t.Fatalf("after reading, chapters = %+v", chapters)
	}
	if !childOf(chapters[2], chapters[0], "Walk away") {
		t.Errorf("a known chapter moved: %+v", chapters[2])
	}
}

// childOf reports whether chapter hangs under parent through choice.
func childOf(chapter, parent schema.StoryChapter, choice string) bool {
	return chapter.Parent != nil && *chapter.Parent == parent.Id && chapter.Choice != nil && *chapter.Choice == choice
}
