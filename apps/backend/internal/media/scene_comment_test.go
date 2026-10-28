package media

import (
	"context"
	"testing"
)

func TestSceneCommentsLinkNamedPerformersAndReplies(t *testing.T) {
	repo, database := newTestRepository(t)
	ctx := context.Background()
	insertRow(t, database, "scene", "scene:one", map[string]any{"title": "One", "source_url": "https://example.test/1"})
	insertRow(t, database, "performer", "performer:anna", map[string]any{"name": "Anna Example"})
	insertRow(t, database, "performer", "performer:ann", map[string]any{"name": "Ann Ex"})
	insertRow(t, database, "performer", "performer:mono", map[string]any{"name": "Sky"})
	for _, statement := range []string{
		`INSERT INTO scene_comment (scene, plugin, position, author, text) VALUES ('scene:one', 'site', 0, 'asker', 'who is she? the sky is blue')`,
		`INSERT INTO scene_comment (scene, plugin, position, author, text, reply_to, likes) VALUES ('scene:one', 'site', 1, 'fan', 'That''s anna example!', 0, 4)`,
	} {
		if _, err := database.Exec(ctx, statement, nil); err != nil {
			t.Fatalf("insert comment: %v", err)
		}
	}

	comments, err := repo.SceneComments(ctx, "scene:one")
	if err != nil {
		t.Fatalf("scene comments: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("got %d comments", len(comments))
	}
	if len(comments[0].Mentions) != 0 {
		t.Errorf("a one-word name counted as a mention: %v", comments[0].Mentions[0].Name)
	}
	answer := comments[1]
	if answer.ReplyTo == nil || *answer.ReplyTo != comments[0].ID {
		t.Errorf("replyTo = %v, want %s", answer.ReplyTo, comments[0].ID)
	}
	// "Ann Ex" is inside "anna example" but not as whole words.
	if len(answer.Mentions) != 1 || answer.Mentions[0].ID != "performer:anna" {
		t.Errorf("mentions = %v", answer.Mentions)
	}
	if answer.Likes == nil || *answer.Likes != 4 {
		t.Errorf("likes = %v", answer.Likes)
	}
}
