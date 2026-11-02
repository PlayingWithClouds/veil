package library

import (
	"context"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/names"
	"github.com/playingwithclouds/veil/internal/plugins"
	"github.com/playingwithclouds/veil/internal/schema"
)

// pageExtras is what a media page lists besides the record's own fields.
type pageExtras struct {
	kind      schema.RecordKind
	id        string
	sourceID  string
	fetched   bool
	downloads []plugins.Download
	markers   []plugins.SceneMarkerInput
	subtitles []plugins.SubtitleTrack
	scripts   []plugins.MotionScript
	comments  []plugins.SceneComment
}

// storeExtras stores a page's extras. A listing card only brings streams;
// a read page also marks its source fetched and brings segments, subtitles,
// scripts and comments.
func (l *Library) storeExtras(ctx context.Context, page pageExtras) error {
	if len(page.downloads) > 0 {
		if err := l.replaceStreams(ctx, page.kind, page.id, page.sourceID, page.downloads); err != nil {
			return err
		}
	}
	if !page.fetched {
		return nil
	}
	if err := l.markFetched(ctx, page.sourceID); err != nil {
		return err
	}
	if err := l.replacePluginSegments(ctx, page.kind, page.id, page.markers); err != nil {
		return err
	}
	if err := l.addSiteSubtitles(ctx, page.kind, page.id, page.subtitles); err != nil {
		return err
	}
	if err := l.addSiteScripts(ctx, page.kind, page.id, page.scripts); err != nil {
		return err
	}
	return l.replaceComments(ctx, page.kind, page.id, page.sourceID, page.comments)
}

// replacePluginSegments swaps the segments plugins gave a record for the
// page's markers. The user's own segments stay. A page without markers
// leaves them, since most sites have none.
func (l *Library) replacePluginSegments(ctx context.Context, kind schema.RecordKind, id string, markers []plugins.SceneMarkerInput) error {
	if len(markers) == 0 {
		return nil
	}
	segments, err := l.markerSegments(ctx, kind, id, markers)
	if err != nil {
		return err
	}
	return l.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM "Segment" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "origin" = $plugin`,
			db.Vars{"kind": kind, "id": id, "plugin": schema.SegmentOriginPlugin}); err != nil {
			return err
		}
		for _, segment := range segments {
			if _, err := tx.Insert(ctx, "Segment", segment); err != nil {
				return err
			}
		}
		return nil
	})
}

// markerSegments maps plugin markers to segments, finding or creating their
// tags by name. Tags are written before the segments' transaction opens,
// since SQLite has one writer.
func (l *Library) markerSegments(ctx context.Context, kind schema.RecordKind, id string, markers []plugins.SceneMarkerInput) ([]schema.Segment, error) {
	segments := make([]schema.Segment, 0, len(markers))
	for _, marker := range markers {
		tagID, err := l.findOrCreateNamed(ctx, tableTag, marker.Tag)
		if err != nil {
			return nil, err
		}
		segments = append(segments, schema.Segment{
			SubjectKind:  kind,
			SubjectId:    id,
			StartSeconds: marker.Seconds,
			EndSeconds:   marker.EndSeconds,
			Title:        optional(strings.TrimSpace(marker.Label)),
			Tag:          optional(tagID),
			Performers:   []string{},
			Origin:       schema.SegmentOriginPlugin,
		})
	}
	return segments, nil
}

// addSiteSubtitles stores the subtitle tracks a page offers that the record
// doesn't have yet. Tracks aren't tied to a source, so tracks another site
// gave stay too.
func (l *Library) addSiteSubtitles(ctx context.Context, kind schema.RecordKind, id string, tracks []plugins.SubtitleTrack) error {
	for _, track := range tracks {
		if track.URL == "" {
			continue
		}
		known, err := l.hasAsset(ctx, "Subtitle", kind, id, track.URL)
		if err != nil {
			return err
		}
		if known {
			continue
		}
		subtitle := schema.Subtitle{
			SubjectKind: kind, SubjectId: id, Language: defaultText(track.Language, "und"),
			Format: subtitleFormat(track), Asset: schema.Asset{Uri: track.URL}, Origin: schema.SubtitleOriginSite,
		}
		if _, err := l.database.Insert(ctx, "Subtitle", subtitle); err != nil {
			return err
		}
	}
	return nil
}

// subtitleFormat is a track's format, else its URL's extension, else vtt.
func subtitleFormat(track plugins.SubtitleTrack) string {
	if track.Format != "" {
		return strings.ToLower(track.Format)
	}
	extension := urlExtension(track.URL)
	if extension == "srt" || extension == "ass" {
		return extension
	}
	return "vtt"
}

// addSiteScripts stores the funscripts a page offers that the record
// doesn't have yet.
func (l *Library) addSiteScripts(ctx context.Context, kind schema.RecordKind, id string, scripts []plugins.MotionScript) error {
	for _, script := range scripts {
		if script.URL == "" {
			continue
		}
		known, err := l.hasAsset(ctx, "Script", kind, id, script.URL)
		if err != nil {
			return err
		}
		if known {
			continue
		}
		row := schema.Script{
			SubjectKind: kind, SubjectId: id, Kind: schema.ScriptKindFunscript,
			Asset: schema.Asset{Uri: script.URL}, Axis: optional(script.Axis), WrittenBy: schema.ActorPlugin,
		}
		if _, err := l.database.Insert(ctx, "Script", row); err != nil {
			return err
		}
	}
	return nil
}

// hasAsset reports whether the record already has a row of table (Subtitle,
// Script) with this asset URI.
func (l *Library) hasAsset(ctx context.Context, table string, kind schema.RecordKind, id, uri string) (bool, error) {
	count, err := l.database.Int(ctx,
		`SELECT COUNT(*) FROM `+db.Quote(table)+` WHERE "subjectKind" = $kind AND "subjectId" = $id
		 AND json_extract("asset", '$.uri') = $uri AND "deletedAt" IS NULL`,
		db.Vars{"kind": kind, "id": id, "uri": uri})
	return count > 0, err
}

// defaultText returns value, or fallback when it's blank.
func defaultText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// replaceComments swaps a source's comments for the ones its page shows now.
// Comments are the site's, so the old rows go for good. Library performers
// named in a comment are recorded as its mentions.
func (l *Library) replaceComments(ctx context.Context, kind schema.RecordKind, id, sourceID string, comments []plugins.SceneComment) error {
	finder, err := l.mentionFinder(ctx, kind, id, comments)
	if err != nil {
		return err
	}
	return l.database.Tx(ctx, func(tx *db.DB) error {
		if _, err := tx.Exec(ctx, `DELETE FROM "Comment" WHERE "source" = $source`, db.Vars{"source": sourceID}); err != nil {
			return err
		}
		ids := make([]string, len(comments))
		for index, comment := range comments {
			row := commentRow(sourceID, comment, ids, finder)
			inserted, err := tx.Insert(ctx, "Comment", row)
			if err != nil {
				return err
			}
			ids[index] = inserted
		}
		return nil
	})
}

// commentRow maps a plugin comment to a row; ids holds the ids of the
// comments stored before it, so a reply can point at its parent.
func commentRow(sourceID string, comment plugins.SceneComment, ids []string, finder *names.Finder) schema.Comment {
	row := schema.Comment{
		Source:   sourceID,
		Author:   optional(comment.Author),
		Text:     comment.Text,
		PostedAt: optional(comment.Date),
		Mentions: finder.Find(comment.Text),
	}
	if comment.Likes != nil {
		likes := int32(*comment.Likes)
		row.Likes = &likes
	}
	if comment.ReplyTo != nil && *comment.ReplyTo >= 0 && *comment.ReplyTo < len(ids) && ids[*comment.ReplyTo] != "" {
		row.ReplyTo = &ids[*comment.ReplyTo]
	}
	return row
}

// mentionFinder knows the names of the performers comments may mention:
// those credited on the record, and any whose name appears word for word
// (two to four words) in a comment.
func (l *Library) mentionFinder(ctx context.Context, kind schema.RecordKind, id string, comments []plugins.SceneComment) (*names.Finder, error) {
	finder := names.NewFinder()
	if len(comments) == 0 {
		return finder, nil
	}
	performers, err := db.QueryAs[schema.Performer](ctx, l.database,
		`SELECT * FROM "Performer" WHERE "deletedAt" IS NULL AND ("nameKey" IN (SELECT value FROM json_each($keys))
		   OR "id" IN (SELECT value FROM json_each((SELECT "performers" FROM `+db.Quote(tableFor(kind))+` WHERE "id" = $id))))`,
		db.Vars{"keys": nameCandidates(comments), "id": id})
	if err != nil {
		return nil, err
	}
	for _, performer := range performers {
		finder.Add(performer.Id, performer.Name)
		for _, alias := range performer.Aliases {
			finder.Add(performer.Id, alias)
		}
	}
	return finder, nil
}

// nameCandidates lists every run of two to four words in the comments,
// folded like name keys.
func nameCandidates(comments []plugins.SceneComment) []string {
	candidates := []string{}
	for _, comment := range comments {
		words := strings.Fields(names.Key(comment.Text))
		candidates = append(candidates, wordRuns(words, 2, 4)...)
	}
	return candidates
}

// wordRuns lists the runs of shortest to longest consecutive words.
func wordRuns(words []string, shortest, longest int) []string {
	runs := []string{}
	for start := range words {
		for length := shortest; length <= longest && start+length <= len(words); length++ {
			runs = append(runs, strings.Join(words[start:start+length], " "))
		}
	}
	return runs
}
