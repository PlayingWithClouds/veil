package library

import (
	"context"
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
	return fields, l.setCredits(ctx, fields, performerNames(story.Performers), story.Tags)
}

// replaceChapters swaps a story's chapters for the ones its page lists,
// keeping the text already read for a chapter whose URL is unchanged.
func (l *Library) replaceChapters(ctx context.Context, storyID string, chapters []plugins.StoryChapter) error {
	known, err := l.chapterTexts(ctx, storyID)
	if err != nil {
		return err
	}
	return l.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM "StoryChapter" WHERE "story" = $story`, db.Vars{"story": storyID}); err != nil {
			return err
		}
		for position, chapter := range chapters {
			if _, err := tx.Insert(ctx, "StoryChapter", chapterRow(storyID, position, chapter, known)); err != nil {
				return err
			}
		}
		return nil
	})
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

// SetChapterText stores a chapter's text once it was read from its URL.
func (l *Library) SetChapterText(ctx context.Context, chapterID string, chapter plugins.StoryChapter) error {
	fields := db.Vars{"text": chapter.Text, "wordCount": wordCountOf(chapter.Text, chapter.WordCount)}
	if strings.TrimSpace(chapter.Title) != "" {
		fields["title"] = strings.TrimSpace(chapter.Title)
	}
	return l.database.Update(ctx, "StoryChapter", chapterID, fields)
}
