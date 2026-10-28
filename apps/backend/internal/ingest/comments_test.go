package ingest

import (
	"slices"
	"testing"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
)

// storedComments returns the scene's comments as "position:text>replyTo" in order.
func storedComments(t *testing.T, database *db.DB, sceneID string) []string {
	t.Helper()
	rows, err := database.Query(t.Context(),
		`SELECT position, text, reply_to FROM scene_comment WHERE scene = $scene ORDER BY position`,
		db.Vars{"scene": sceneID})
	if err != nil {
		t.Fatalf("query comments: %v", err)
	}
	comments := []string{}
	for _, row := range rows {
		entry := row["text"].(string)
		if row["reply_to"] != nil {
			entry += ">" + string(rune('0'+db.AsInt(row["reply_to"])))
		}
		comments = append(comments, entry)
	}
	return comments
}

func TestScrapedCommentsReplaceTheStoredOnes(t *testing.T) {
	service, database, ctx := openTestService(t)
	replyToFirst := 0
	replyToItself := 3
	first := ingestScene(t, ctx, service, "testplugin", &plugins.Scene{
		Type: plugins.MediaTypeScene, ExternalID: "test-1", SourceURL: testSceneURL, Title: "Visited",
		Comments: []plugins.SceneComment{
			{Text: "who is she?", Author: "asker"},
			{Text: "  ", Author: "blank"},
			{Text: "Anna Example", ReplyTo: &replyToFirst},
			{Text: "loops", ReplyTo: &replyToItself},
		},
	})
	got := storedComments(t, database, first.CanonicalID)
	if !slices.Equal(got, []string{"who is she?", "Anna Example>0", "loops"}) {
		t.Fatalf("first scrape = %v", got)
	}

	// A later scrape that found no comments keeps them; one that found some replaces them.
	ingestScene(t, ctx, service, "testplugin", &plugins.Scene{
		Type: plugins.MediaTypeScene, ExternalID: "test-1", SourceURL: testSceneURL, Title: "Visited",
	})
	if got := storedComments(t, database, first.CanonicalID); len(got) != 3 {
		t.Fatalf("after an empty scrape = %v", got)
	}
	ingestScene(t, ctx, service, "testplugin", &plugins.Scene{
		Type: plugins.MediaTypeScene, ExternalID: "test-1", SourceURL: testSceneURL, Title: "Visited",
		Comments: []plugins.SceneComment{{Text: "new one"}},
	})
	if got := storedComments(t, database, first.CanonicalID); !slices.Equal(got, []string{"new one"}) {
		t.Fatalf("after a rescrape = %v", got)
	}
}
