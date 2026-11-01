package library

import (
	"context"
	"fmt"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// upsertVideo stores a plugin scene as a video: a new record for a site
// identity seen the first time, else an update of the record it points at.
// fetched marks the source's page as read (a find, not a listing card).
func (l *Library) upsertVideo(ctx context.Context, plugin string, scene *plugins.Scene, fetched bool) (string, error) {
	site, ok := identity(plugin, scene.ExternalID, scene.SourceURL)
	if !ok {
		return "", fmt.Errorf("%s returned a video without id or URL", plugin)
	}
	fields, err := l.videoFields(ctx, plugin, scene, fetched)
	if err != nil {
		return "", err
	}
	source, err := l.findSource(ctx, site)
	if err != nil {
		return "", err
	}
	var id, sourceID string
	if source == nil {
		id, sourceID, err = l.createVideo(ctx, site, fields, scene)
	} else {
		id, sourceID = source.SubjectId, source.Id
		err = l.database.Update(ctx, tableVideo, id, fields)
	}
	if err != nil {
		return "", err
	}
	return id, l.afterVideoWrite(ctx, id, sourceID, scene, fetched)
}

// createVideo inserts a new video and its first source.
func (l *Library) createVideo(ctx context.Context, site siteIdentity, fields db.Vars, scene *plugins.Scene) (string, string, error) {
	fields["status"] = string(schema.MediaStatusAvailable)
	fillListDefaults(fields, "performers", "tags")
	if _, hasTitle := fields["title"]; !hasTitle {
		fields["title"] = site.URL
	}
	id, err := l.database.Insert(ctx, tableVideo, fields)
	if err != nil {
		return "", "", err
	}
	sourceID, err := l.addSource(ctx, schema.RecordKindVideo, id, site, scene.Title, assetOf(scene.PosterPath))
	return id, sourceID, err
}

// afterVideoWrite marks the source fetched and replaces its streams.
func (l *Library) afterVideoWrite(ctx context.Context, id, sourceID string, scene *plugins.Scene, fetched bool) error {
	if fetched {
		if err := l.markFetched(ctx, sourceID); err != nil {
			return err
		}
	}
	if len(scene.Downloads) == 0 {
		return nil
	}
	return l.replaceStreams(ctx, schema.RecordKindVideo, id, sourceID, scene.Downloads)
}

// videoFields are the columns a scene fills. A listing card (not fetched)
// only refreshes what it shows and never clears credits a page gave.
func (l *Library) videoFields(ctx context.Context, plugin string, scene *plugins.Scene, fetched bool) (db.Vars, error) {
	fields := db.Vars{}
	setText(fields, "title", scene.Title)
	setText(fields, "details", scene.Details)
	setText(fields, "date", isoDate(scene.Date))
	if scene.Duration > 0 {
		fields["durationSeconds"] = float64(scene.Duration)
	}
	setAsset(fields, "cover", posterOf(scene))
	setAsset(fields, "preview", scene.PreviewVideo)
	if err := l.setStudio(ctx, fields, plugin, scene.Studio); err != nil {
		return nil, err
	}
	if !fetched {
		return fields, nil
	}
	return fields, l.setCredits(ctx, fields, performerNames(scene.Performers), scene.Tags)
}

// setStudio sets the studio column from a studio reference, when there is one.
func (l *Library) setStudio(ctx context.Context, fields db.Vars, plugin string, reference *plugins.StudioRef) error {
	if reference == nil || reference.Name == "" {
		return nil
	}
	studioID, err := l.upsertStudio(ctx, plugin, &plugins.Studio{
		Name: reference.Name, ExternalID: reference.ExternalID, SourceURL: reference.SourceURL, ImagePath: reference.ImagePath,
	})
	if err != nil || studioID == "" {
		return err
	}
	fields["studio"] = studioID
	return nil
}

// setCredits sets the performers and tags columns from names.
func (l *Library) setCredits(ctx context.Context, fields db.Vars, performerNames, tagNames []string) error {
	performerIDs, err := l.performerIDs(ctx, performerNames)
	if err != nil {
		return err
	}
	tagIDs, err := l.tagIDs(ctx, tagNames)
	if err != nil {
		return err
	}
	fields["performers"] = performerIDs
	fields["tags"] = tagIDs
	return nil
}

// performerNames lists the credited names of a scene, in credit order.
func performerNames(credits []plugins.ScenePerformer) []string {
	result := make([]string, 0, len(credits))
	for _, credit := range credits {
		result = append(result, credit.Name)
	}
	return result
}

// posterOf returns the scene's cover image: its poster path, else the first
// image typed poster, else the first preview frame.
func posterOf(scene *plugins.Scene) string {
	if scene.PosterPath != "" {
		return scene.PosterPath
	}
	for _, image := range scene.Images {
		if image.Type == "poster" {
			return image.FilePath
		}
	}
	if len(scene.PreviewImages) > 0 {
		return scene.PreviewImages[0]
	}
	return ""
}
