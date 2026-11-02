package library

import (
	"context"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// TestSceneExtras stores a scene page with markers, a subtitle, a script and
// comments, one of which names a credited performer and replies to another.
func TestSceneExtras(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	page := foundScene()
	end := 30.0
	replyTo := 0
	page[0].Scene.Markers = []plugins.SceneMarkerInput{{Seconds: 10, EndSeconds: &end, Tag: "Kissing", Label: "Intro"}}
	page[0].Scene.Subtitles = []plugins.SubtitleTrack{{URL: "https://cdn.test/v1.srt", Language: "en"}}
	page[0].Scene.Scripts = []plugins.MotionScript{{URL: "https://cdn.test/v1.funscript"}}
	page[0].Scene.Comments = []plugins.SceneComment{{Text: "who is she?"}, {Text: "That's Ada Lane", ReplyTo: &replyTo}}
	page[0].Scene.VR = &plugins.VRFormat{Projection: "equirectangular180", Stereo: "sideBySide"}
	ref, err := library.IngestFound(ctx, "site", page)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := library.IngestFound(ctx, "site", page); err != nil {
		t.Fatal(err)
	}
	video, _ := library.Video(ctx, ref.Id)
	if video.Vr == nil || video.Vr.Stereo != schema.VrStereoSideBySide {
		t.Errorf("vr = %+v", video.Vr)
	}
	expectCount(t, library.DB(), `SELECT COUNT(*) FROM "Segment" WHERE "subjectId" = $id AND "tag" IS NOT NULL`, ref.Id, 1)
	expectCount(t, library.DB(), `SELECT COUNT(*) FROM "Subtitle" WHERE "subjectId" = $id AND "format" = 'srt'`, ref.Id, 1)
	expectCount(t, library.DB(), `SELECT COUNT(*) FROM "Script" WHERE "subjectId" = $id`, ref.Id, 1)
	comments, err := db.QueryAs[schema.Comment](ctx, library.DB(), `SELECT * FROM "Comment" ORDER BY rowid`, nil)
	if err != nil || len(comments) != 2 {
		t.Fatalf("comments = %+v, err %v", comments, err)
	}
	if len(comments[1].Mentions) != 1 || comments[1].Mentions[0] != video.Performers[0] || comments[1].ReplyTo == nil || *comments[1].ReplyTo != comments[0].Id {
		t.Errorf("reply = %+v", comments[1])
	}
}

// expectCount runs a COUNT query for the record id and expects want.
func expectCount(t *testing.T, database *db.DB, query, id string, want int) {
	t.Helper()
	count, err := database.Int(context.Background(), query, db.Vars{"id": id})
	if err != nil || count != want {
		t.Errorf("%s: %d, err %v, want %d", query, count, err, want)
	}
}

// TestOtherMediaKinds stores an audio, a story with chapters and a post with
// an attachment, then lists them together.
func TestOtherMediaKinds(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	results := []*plugins.ScrapeResult{
		{Type: plugins.MediaTypeAudio, Audio: &plugins.Audio{ExternalID: "a1", Title: "Whispers", Duration: 900}},
		{Type: plugins.MediaTypeStory, Story: &plugins.Story{ExternalID: "s1", Title: "Summer", Chapters: []plugins.StoryChapter{
			{Title: "One", Text: "It was hot that day"}, {Title: "Two", URL: "https://stories.test/s1/2"},
		}}},
		{Type: plugins.MediaTypePost, Post: &plugins.Post{ExternalID: "p1", Text: "New set is up!", Creator: plugins.ScenePerformer{Name: "Cara Vale"},
			Attachments: []plugins.ScrapeResult{{Type: plugins.MediaTypeScene, Scene: &plugins.Scene{ExternalID: "p1-video", Title: "Behind the scenes"}}}}},
	}
	for _, result := range results {
		if _, err := library.IngestFound(ctx, "site", []*plugins.ScrapeResult{result}); err != nil {
			t.Fatal(err)
		}
	}
	stories, _ := library.Stories(ctx, MediaQuery{})
	chapters, _ := library.Chapters(ctx, stories[0].Id)
	if stories[0].ChapterCount != 2 || len(chapters) != 2 || chapters[0].WordCount != 5 || chapters[1].Url == nil {
		t.Errorf("story %+v, chapters %+v", stories[0], chapters)
	}
	posts, _ := library.Posts(ctx, MediaQuery{})
	if len(posts) != 1 || posts[0].Title != "New set is up!" || len(posts[0].Attachments) != 1 || posts[0].Creator == "" {
		t.Errorf("posts = %+v", posts)
	}
	all, err := library.Media(ctx, MediaQuery{Filter: schema.Filter{Sort: &schema.Sort{Field: schema.SortFieldTitle}}})
	if err != nil || len(all) != 4 {
		t.Fatalf("media = %+v, err %v", all, err)
	}
	if all[0].Kind != schema.RecordKindVideo || all[3].Kind != schema.RecordKindAudio {
		t.Errorf("media by title = %+v, want Behind…, New…, Summer, Whispers", all)
	}
	audioOnly, _ := library.Media(ctx, MediaQuery{Filter: schema.Filter{Kinds: []schema.MediaKind{schema.MediaKindAudio}}})
	if len(audioOnly) != 1 {
		t.Errorf("audio only = %+v", audioOnly)
	}
}

// TestGroupMembers stores a site collection whose members are one known and
// one unknown page, and expects both as the group's videos in order.
func TestGroupMembers(t *testing.T) {
	ctx := context.Background()
	library := openTestLibrary(t)
	ref, err := library.IngestFound(ctx, "site", foundScene())
	if err != nil {
		t.Fatal(err)
	}
	group := []*plugins.ScrapeResult{{Type: plugins.MediaTypeCollection, Collection: &plugins.Collection{
		SourceURL: "https://site.test/series/1", Name: "Series", MemberURLs: []string{"https://site.test/v2", "https://site.test/v1"},
	}}}
	groupRef, err := library.IngestFound(ctx, "site", group)
	if err != nil {
		t.Fatal(err)
	}
	videos, err := library.GroupVideos(ctx, groupRef.Id)
	if err != nil || len(videos) != 2 || videos[1].Id != ref.Id {
		t.Fatalf("group videos = %+v, err %v", videos, err)
	}
}
