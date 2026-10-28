package media

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/playingwithclouds/veil/internal/api/graphql/model"
	"github.com/playingwithclouds/veil/internal/db"
)

// SceneComments returns the comments the scene's sites showed, each site's in
// its order, with the library performers every comment names.
func (r *Repository) SceneComments(ctx context.Context, sceneID string) ([]*model.SceneComment, error) {
	scene, err := parseMediaID(sceneID)
	if err != nil {
		return nil, err
	}
	rows, err := r.database.Query(ctx,
		`SELECT plugin, position, author, text, posted_at, likes, reply_to
		   FROM scene_comment WHERE scene = $scene ORDER BY plugin, position`,
		db.Vars{"scene": *scene})
	if err != nil {
		return nil, fmt.Errorf("scene comments: %w", err)
	}
	mentions, err := r.commentMentions(ctx, *scene)
	if err != nil {
		return nil, err
	}
	comments := make([]*model.SceneComment, 0, len(rows))
	for _, row := range rows {
		id := commentID(mString(row, "plugin"), mInt(row, "position"))
		comment := &model.SceneComment{
			ID:       id,
			Plugin:   mString(row, "plugin"),
			Author:   mStringPtr(row, "author"),
			Text:     mString(row, "text"),
			PostedAt: mStringPtr(row, "posted_at"),
			Likes:    mIntPtr(row, "likes"),
			Mentions: mentions[id],
		}
		if comment.Mentions == nil {
			comment.Mentions = []*model.Performer{}
		}
		if replyTo := mIntPtr(row, "reply_to"); replyTo != nil {
			target := commentID(comment.Plugin, *replyTo)
			comment.ReplyTo = &target
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

// commentID names a comment within its scene.
func commentID(plugin string, position int) string {
	return fmt.Sprintf("%s:%d", plugin, position)
}

// commentMentions finds, per comment id, the unblocked performers whose full
// name (two words or more) the comment contains as whole words. One-word names
// are left out: in prose they are mostly ordinary words.
func (r *Repository) commentMentions(ctx context.Context, scene db.RecordID) (map[string][]*model.Performer, error) {
	rows, err := r.database.Query(ctx, fmt.Sprintf(
		`SELECT scene_comment.plugin AS comment_plugin, scene_comment.position AS comment_position,
		        scene_comment.text AS comment_text, performer.*
		   FROM scene_comment
		   JOIN performer ON instr(lower(scene_comment.text), lower(performer.name)) > 0
		  WHERE scene_comment.scene = $scene AND trim(performer.name) LIKE '%% %%' AND %s`,
		notBlockedCondition("performer")), db.Vars{"scene": scene})
	if err != nil {
		return nil, fmt.Errorf("comment mentions: %w", err)
	}
	mentions := make(map[string][]*model.Performer)
	for _, row := range rows {
		if !containsWords(mString(row, "comment_text"), mString(row, "name")) {
			continue
		}
		id := commentID(mString(row, "comment_plugin"), mInt(row, "comment_position"))
		mentions[id] = append(mentions[id], r.performerFromMap(ctx, row, false))
	}
	return mentions, nil
}

// containsWords reports whether text contains phrase, case-insensitively, with
// no letter or digit right before or after it ("Ann Lee" is not in "Joann Leet").
func containsWords(text, phrase string) bool {
	haystack := []rune(strings.ToLower(text))
	needle := []rune(strings.ToLower(strings.TrimSpace(phrase)))
	if len(needle) == 0 {
		return false
	}
	for start := 0; start+len(needle) <= len(haystack); start++ {
		if string(haystack[start:start+len(needle)]) != string(needle) {
			continue
		}
		end := start + len(needle)
		if start > 0 && isWordRune(haystack[start-1]) {
			continue
		}
		if end < len(haystack) && isWordRune(haystack[end]) {
			continue
		}
		return true
	}
	return false
}

// isWordRune reports whether r is part of a word.
func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
