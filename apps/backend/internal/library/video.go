package library

import (
	"context"
	"log"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// upsertVideo stores a plugin scene as a video: a new record for a site
// identity seen the first time, else an update of the record it points at.
// fetched marks the source's page as read (a find, not a listing card).
func (l *Library) upsertVideo(ctx context.Context, plugin string, scene *plugins.Scene, fetched bool) (string, error) {
	site, err := identityOf(schema.RecordKindVideo, plugin, scene.ExternalID, scene.SourceURL)
	if err != nil {
		return "", err
	}
	fields, err := l.videoFields(ctx, plugin, scene, fetched)
	if err != nil {
		return "", err
	}
	id, sourceID, err := l.upsertMedia(ctx, mediaWrite{
		kind: schema.RecordKindVideo, plugin: plugin, site: site, fields: fields,
		label: scene.Title, image: assetOf(scene.PosterPath),
	})
	if err != nil {
		return "", err
	}
	return id, l.afterVideoWrite(ctx, id, sourceID, scene, fetched)
}

// afterVideoWrite marks the source fetched and replaces what the page lists
// with the video: streams, segments, subtitles, scripts and comments.
func (l *Library) afterVideoWrite(ctx context.Context, id, sourceID string, scene *plugins.Scene, fetched bool) error {
	page := pageExtras{
		kind: schema.RecordKindVideo, id: id, sourceID: sourceID, fetched: fetched,
		downloads: scene.Downloads, markers: scene.Markers, subtitles: scene.Subtitles,
		scripts: scene.Scripts, comments: scene.Comments,
	}
	if err := l.storeExtras(ctx, page); err != nil {
		return err
	}
	if scene.Duration > 0 {
		if err := l.DetectAlike(ctx, id); err != nil {
			log.Printf("alike %s: %v", id, err)
		}
	}
	return nil
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
	setFrameSize(fields, scene.Width, scene.Height)
	if scene.VR != nil {
		fields["vr"] = vrFormat(*scene.VR)
	}
	if err := l.setStudio(ctx, fields, plugin, scene.Studio); err != nil {
		return nil, err
	}
	if !fetched {
		return fields, nil
	}
	return fields, l.setCredits(ctx, fields, performerNames(scene.Performers), scene.Tags)
}

// setFrameSize sets width and height when the page gave both.
func setFrameSize(fields db.Vars, width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	fields["width"] = width
	fields["height"] = height
}

// vrFormat maps a plugin's VR format to the schema's; unknown projections
// and packings default to the most common VR video (180°, side by side).
func vrFormat(format plugins.VRFormat) schema.VrFormat {
	vr := schema.VrFormat{Projection: schema.VrProjectionEquirectangular180, Stereo: schema.VrStereoSideBySide}
	if knownProjections[format.Projection] {
		vr.Projection = schema.VrProjection(format.Projection)
	}
	if knownStereo[format.Stereo] {
		vr.Stereo = schema.VrStereo(format.Stereo)
	}
	if format.FieldOfView > 0 {
		fieldOfView := int32(format.FieldOfView)
		vr.FieldOfView = &fieldOfView
	}
	return vr
}

// knownProjections and knownStereo are the schema's VR enum values.
var (
	knownProjections = map[string]bool{"equirectangular180": true, "equirectangular360": true, "fisheye": true, "cubemap": true}
	knownStereo      = map[string]bool{"mono": true, "sideBySide": true, "topBottom": true}
)

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
