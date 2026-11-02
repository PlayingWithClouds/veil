package library

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// postTitleLength is how much of a post's text becomes its title when the
// post has none.
const postTitleLength = 80

// upsertPost stores a creator-platform post: its creator as a performer,
// and each attachment as a record of its own kind that the post points at.
func (l *Library) upsertPost(ctx context.Context, plugin string, post *plugins.Post, fetched bool) (string, error) {
	site, err := identityOf(schema.RecordKindPost, plugin, post.ExternalID, post.SourceURL)
	if err != nil {
		return "", err
	}
	creatorID, err := l.upsertPerformer(ctx, plugin, &plugins.Performer{
		Name: post.Creator.Name, ExternalID: post.Creator.ExternalID, SourceURL: post.Creator.SourceURL,
	})
	if err != nil {
		return "", err
	}
	if creatorID == "" {
		return "", fmt.Errorf("%s returned a post without a creator", plugin)
	}
	fields, err := l.postFields(ctx, plugin, post, creatorID, fetched)
	if err != nil {
		return "", err
	}
	id, sourceID, err := l.upsertMedia(ctx, mediaWrite{
		kind: schema.RecordKindPost, plugin: plugin, site: site, fields: fields,
		defaults: db.Vars{"attachments": []schema.RecordRef{}, "locked": false}, label: postTitle(post), image: assetOf(post.CoverPath),
	})
	if err != nil {
		return "", err
	}
	return id, l.storeExtras(ctx, pageExtras{kind: schema.RecordKindPost, id: id, sourceID: sourceID, fetched: fetched, comments: post.Comments})
}

// postFields are the columns a post fills. Attachments are stored first, so
// the post can point at them.
func (l *Library) postFields(ctx context.Context, plugin string, post *plugins.Post, creatorID string, fetched bool) (db.Vars, error) {
	fields := db.Vars{"creator": creatorID, "performers": []string{creatorID}, "locked": post.Locked}
	setText(fields, "title", postTitle(post))
	setText(fields, "text", post.Text)
	setText(fields, "date", isoDate(post.Date))
	setAsset(fields, "cover", post.CoverPath)
	setText(fields, "currency", post.Currency)
	if post.PriceCents > 0 {
		fields["priceCents"] = post.PriceCents
	}
	if !fetched {
		return fields, nil
	}
	tagIDs, err := l.tagIDs(ctx, post.Tags)
	if err != nil {
		return nil, err
	}
	fields["tags"] = tagIDs
	attachments, err := l.ingestAttachments(ctx, plugin, post.Attachments)
	if err != nil {
		return nil, err
	}
	if len(attachments) > 0 {
		fields["attachments"] = attachments
	}
	return fields, nil
}

// ingestAttachments stores a post's media and returns what they became. An
// attachment that fails is skipped, so one broken file doesn't lose the post.
func (l *Library) ingestAttachments(ctx context.Context, plugin string, attachments []plugins.ScrapeResult) ([]schema.RecordRef, error) {
	refs := []schema.RecordRef{}
	for index := range attachments {
		if attachments[index].Post != nil {
			continue
		}
		ref, err := l.ingestResult(ctx, plugin, &attachments[index])
		if err != nil || ref.Id == "" {
			continue
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

// postTitle is a post's title, else the start of its text.
func postTitle(post *plugins.Post) string {
	if strings.TrimSpace(post.Title) != "" {
		return post.Title
	}
	text := strings.Join(strings.Fields(post.Text), " ")
	if utf8.RuneCountInString(text) <= postTitleLength {
		return text
	}
	return string([]rune(text)[:postTitleLength]) + "…"
}

// Posts lists a page of posts matching the query's filter.
func (l *Library) Posts(ctx context.Context, query MediaQuery) ([]schema.Post, error) {
	statement, vars := mediaQuery(schema.RecordKindPost, query)
	return db.QueryAs[schema.Post](ctx, l.database, statement, vars)
}

// Post loads one post; nil when it doesn't exist or is trashed.
func (l *Library) Post(ctx context.Context, id string) (*schema.Post, error) {
	return db.Get[schema.Post](ctx, l.database, tablePost, id)
}
