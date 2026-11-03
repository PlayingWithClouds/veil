package library

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// upsertStory stores a plugin's story and, when its page lists them, its
// chapters.
func (l *Library) upsertStory(ctx context.Context, plugin string, story *plugins.Story, fetched bool) (string, error) {
	site, err := identityOf(schema.RecordKindStory, plugin, story.ExternalID, story.SourceURL)
	if err != nil {
		return "", err
	}
	fields, err := l.storyFields(ctx, plugin, story, fetched)
	if err != nil {
		return "", err
	}
	id, sourceID, err := l.upsertMedia(ctx, mediaWrite{
		kind: schema.RecordKindStory, plugin: plugin, site: site, fields: fields,
		defaults: db.Vars{"chapterCount": 0, "complete": false}, label: story.Title, image: assetOf(story.CoverPath),
	})
	if err != nil {
		return "", err
	}
	if fetched && len(story.Chapters) > 0 {
		if err := l.replaceChapters(ctx, id, story.Chapters); err != nil {
			return "", err
		}
	}
	return id, l.storeExtras(ctx, pageExtras{kind: schema.RecordKindStory, id: id, sourceID: sourceID, fetched: fetched, comments: story.Comments})
}

// storyFields are the columns a story fills; a listing leaves credits alone.
func (l *Library) storyFields(ctx context.Context, plugin string, story *plugins.Story, fetched bool) (db.Vars, error) {
	fields := db.Vars{}
	setText(fields, "title", story.Title)
	setText(fields, "details", story.Details)
	setText(fields, "date", isoDate(story.Date))
	setAsset(fields, "cover", story.CoverPath)
	if story.WordCount > 0 {
		fields["wordCount"] = story.WordCount
	}
	if err := l.setStudio(ctx, fields, plugin, story.Studio); err != nil {
		return nil, err
	}
	if !fetched {
		return fields, nil
	}
	fields["complete"] = story.Complete
	if len(story.Chapters) > 0 {
		fields["chapterCount"] = len(story.Chapters)
	}
	return fields, l.setCredits(ctx, fields, plugin, story.Performers, story.Tags)
}

// replaceChapters swaps a story's chapters for the ones its page lists,
// keeping the text already read for a chapter whose URL is unchanged, and
// links a branching story's chapters to their parents.
func (l *Library) replaceChapters(ctx context.Context, storyID string, chapters []plugins.StoryChapter) error {
	known, err := l.chapterTexts(ctx, storyID)
	if err != nil {
		return err
	}
	return l.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM "StoryChapter" WHERE "story" = $story`, db.Vars{"story": storyID}); err != nil {
			return err
		}
		ids, err := insertChapters(ctx, tx, storyID, chapters, known)
		if err != nil {
			return err
		}
		return linkParents(ctx, tx, chapters, ids)
	})
}

// insertChapters stores chapters in their order and maps each one's URL to
// its id.
func insertChapters(ctx context.Context, tx *db.DB, storyID string, chapters []plugins.StoryChapter, known map[string]string) (map[string]string, error) {
	ids := map[string]string{}
	for position, chapter := range chapters {
		id, err := tx.Insert(ctx, "StoryChapter", chapterRow(storyID, position, chapter, known))
		if err != nil {
			return nil, err
		}
		if chapter.URL != "" {
			ids[chapter.URL] = id
		}
	}
	return ids, nil
}

// linkParents points every chapter that names a parent at it; a parent the
// page didn't list leaves the chapter unlinked.
func linkParents(ctx context.Context, tx *db.DB, chapters []plugins.StoryChapter, ids map[string]string) error {
	for _, chapter := range chapters {
		parentID, listed := ids[chapter.ParentURL]
		if chapter.ParentURL == "" || chapter.URL == "" || !listed {
			continue
		}
		if err := tx.Update(ctx, "StoryChapter", ids[chapter.URL], db.Vars{"parent": parentID}); err != nil {
			return err
		}
	}
	return nil
}

// chapterTexts maps the URL of each chapter already read to its text.
func (l *Library) chapterTexts(ctx context.Context, storyID string) (map[string]string, error) {
	chapters, err := l.Chapters(ctx, storyID)
	if err != nil {
		return nil, err
	}
	texts := map[string]string{}
	for _, chapter := range chapters {
		if chapter.Url != nil && chapter.Text != "" {
			texts[*chapter.Url] = chapter.Text
		}
	}
	return texts, nil
}

// chapterRow maps a plugin chapter to a row at position.
func chapterRow(storyID string, position int, chapter plugins.StoryChapter, known map[string]string) schema.StoryChapter {
	text := chapter.Text
	if text == "" && chapter.URL != "" {
		text = known[chapter.URL]
	}
	return schema.StoryChapter{
		Story:     storyID,
		Position:  int32(position),
		Title:     optional(strings.TrimSpace(chapter.Title)),
		Text:      text,
		WordCount: int32(wordCountOf(text, chapter.WordCount)),
		Url:       optional(chapter.URL),
		Choice:    optional(strings.TrimSpace(chapter.Choice)),
	}
}

// wordCountOf is the given count, else the words in text.
func wordCountOf(text string, given int) int {
	if given > 0 {
		return given
	}
	return len(strings.Fields(text))
}

// Stories lists a page of stories matching the query's filter.
func (l *Library) Stories(ctx context.Context, query MediaQuery) ([]schema.Story, error) {
	statement, vars := mediaQuery(schema.RecordKindStory, query)
	return db.QueryAs[schema.Story](ctx, l.database, statement, vars)
}

// Story loads one story; nil when it doesn't exist or is trashed.
func (l *Library) Story(ctx context.Context, id string) (*schema.Story, error) {
	return db.Get[schema.Story](ctx, l.database, tableStory, id)
}

// SimilarStories lists up to limit other stories sharing the most tags with
// the story, newest first among equals; blocked and trashed ones are left out.
func (l *Library) SimilarStories(ctx context.Context, storyID string, limit int) ([]schema.Story, error) {
	story, err := l.Story(ctx, storyID)
	if err != nil || story == nil || len(story.Tags) == 0 {
		return []schema.Story{}, err
	}
	vars := db.Vars{"story": storyID, "tags": story.Tags}
	vars["limit"], _ = pageBounds(limit, 0)
	ids, err := l.database.Strings(ctx, `SELECT "other"."id" FROM "Story" AS "other", json_each("other"."tags") AS "tag"
		WHERE "other"."deletedAt" IS NULL AND "other"."id" != $story
			AND "tag"."value" IN (SELECT value FROM json_each($tags))
		GROUP BY "other"."id"
		ORDER BY COUNT(*) DESC, "other"."createdAt" DESC
		LIMIT $limit`, vars)
	if err != nil {
		return nil, err
	}
	visible, err := l.MatchingIDs(ctx, schema.RecordKindStory, schema.Filter{}, ids)
	if err != nil {
		return nil, err
	}
	return l.storiesInOrder(ctx, ids, visible)
}

// storiesInOrder loads the stories of ids that are also in visible, in the order of ids.
func (l *Library) storiesInOrder(ctx context.Context, ids, visible []string) ([]schema.Story, error) {
	stories := make([]schema.Story, 0, len(visible))
	for _, id := range ids {
		if !slices.Contains(visible, id) {
			continue
		}
		story, err := l.Story(ctx, id)
		if err != nil {
			return nil, err
		}
		if story != nil {
			stories = append(stories, *story)
		}
	}
	return stories, nil
}

// Chapters lists a story's chapters in reading order.
func (l *Library) Chapters(ctx context.Context, storyID string) ([]schema.StoryChapter, error) {
	return db.QueryAs[schema.StoryChapter](ctx, l.database,
		`SELECT * FROM "StoryChapter" WHERE "story" = $story AND "deletedAt" IS NULL ORDER BY "position"`,
		db.Vars{"story": storyID})
}

// Chapter loads one chapter by its story and position.
func (l *Library) Chapter(ctx context.Context, storyID string, position int) (*schema.StoryChapter, error) {
	return db.QueryOneAs[schema.StoryChapter](ctx, l.database,
		`SELECT * FROM "StoryChapter" WHERE "story" = $story AND "position" = $position AND "deletedAt" IS NULL`,
		db.Vars{"story": storyID, "position": position})
}

// SetChapterText stores a chapter's text once it was read from its URL,
// and the chapters its choices lead to.
func (l *Library) SetChapterText(ctx context.Context, chapterID string, chapter plugins.StoryChapter) error {
	fields := db.Vars{"text": chapter.Text, "wordCount": wordCountOf(chapter.Text, chapter.WordCount)}
	if strings.TrimSpace(chapter.Title) != "" {
		fields["title"] = strings.TrimSpace(chapter.Title)
	}
	if err := l.database.Update(ctx, "StoryChapter", chapterID, fields); err != nil {
		return err
	}
	return l.addChoices(ctx, chapterID, chapter.Choices)
}

// addChoices stores the chapters a chapter's choices lead to, found while
// reading it, as its children.
func (l *Library) addChoices(ctx context.Context, chapterID string, choices []plugins.StoryChapter) error {
	if len(choices) == 0 {
		return nil
	}
	parent, err := db.Get[schema.StoryChapter](ctx, l.database, "StoryChapter", chapterID)
	if err != nil || parent == nil {
		return err
	}
	return l.database.Tx(ctx, func(tx *db.DB) error {
		for _, choice := range choices {
			if err := addChoice(ctx, tx, parent, choice); err != nil {
				return err
			}
		}
		return nil
	})
}

// addChoice links the chapter one choice leads to. A chapter the story has
// already (same URL) keeps the parent it was first found under, so the
// story stays a tree; a new one goes at the end.
func addChoice(ctx context.Context, tx *db.DB, parent *schema.StoryChapter, choice plugins.StoryChapter) error {
	if choice.URL == "" {
		return nil
	}
	vars := db.Vars{"story": parent.Story, "url": choice.URL}
	known, err := tx.Strings(ctx, `SELECT "id" FROM "StoryChapter" WHERE "story" = $story AND "url" = $url AND "deletedAt" IS NULL`, vars)
	if err != nil || len(known) > 0 {
		return err
	}
	positions, err := tx.Strings(ctx, `SELECT CAST(COALESCE(MAX("position"), -1) + 1 AS TEXT) FROM "StoryChapter" WHERE "story" = $story`, vars)
	if err != nil {
		return err
	}
	position, err := strconv.Atoi(positions[0])
	if err != nil {
		return err
	}
	row := chapterRow(parent.Story, position, choice, nil)
	row.Parent = &parent.Id
	_, err = tx.Insert(ctx, "StoryChapter", row)
	return err
}
