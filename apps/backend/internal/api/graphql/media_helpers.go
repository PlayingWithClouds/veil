package graphql

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// valueOf returns the string a pointer holds, or "" (text is shadowed by
// arguments named text).
func valueOf(value *string) string {
	return text(value)
}

// readChapter returns a story's chapter, first reading its text from the
// site when the story listed it without: through story:chapter of the
// plugin of one of the story's sources.
func (r *Resolver) readChapter(ctx context.Context, storyID string, position int) (*schema.StoryChapter, error) {
	chapter, err := r.library.Chapter(ctx, storyID, position)
	if err != nil || chapter == nil {
		return chapter, err
	}
	if chapter.Text != "" || chapter.Url == nil {
		return chapter, nil
	}
	plugin, err := r.chapterReader(ctx, storyID)
	if err != nil {
		return nil, err
	}
	read, err := r.runner.StoryChapter(ctx, plugin, *chapter.Url)
	if err != nil {
		return nil, err
	}
	if err := r.library.SetChapterText(ctx, chapter.Id, *read); err != nil {
		return nil, err
	}
	return r.library.Chapter(ctx, storyID, position)
}

// chapterReader finds the active plugin of one of the story's sources that
// reads chapters.
func (r *Resolver) chapterReader(ctx context.Context, storyID string) (*plugins.Plugin, error) {
	sources, err := r.library.Sources(ctx, schema.RecordKindStory, storyID)
	if err != nil {
		return nil, err
	}
	for _, source := range sources {
		plugin, found := r.registry.Get(source.Plugin)
		if found && r.registry.IsActive(source.Plugin) && plugin.Serves(plugins.CapabilityStoryChapter) {
			return plugin, nil
		}
	}
	return nil, fmt.Errorf("no active plugin reads this story's chapters")
}
