package library

import (
	"context"
	"fmt"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

// SegmentEdit is a new segment of a video or audio, or a change to one.
// Nil fields stay (or take their default); an empty title or tag clears it.
type SegmentEdit struct {
	StartSeconds *float64
	EndSeconds   *float64
	// ClearEnd turns a range into a marker (no end).
	ClearEnd     bool
	Title        *string
	TagID        *string
	PerformerIDs *[]string
}

// Segments lists a record's segments in time order: scenes, chapters and
// markers, the user's and the plugins'.
func (l *Library) Segments(ctx context.Context, kind schema.RecordKind, id string) ([]schema.Segment, error) {
	return db.QueryAs[schema.Segment](ctx, l.database,
		`SELECT * FROM "Segment" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL
		 ORDER BY "startSeconds", "endSeconds" IS NULL, "endSeconds"`,
		db.Vars{"kind": kind, "id": id})
}

// Segment loads one segment.
func (l *Library) Segment(ctx context.Context, id string) (*schema.Segment, error) {
	return db.Get[schema.Segment](ctx, l.database, "Segment", id)
}

// AddSegment adds the user's own segment to a video or audio.
func (l *Library) AddSegment(ctx context.Context, kind schema.RecordKind, id string, edit SegmentEdit) (*schema.Segment, error) {
	if kind != schema.RecordKindVideo && kind != schema.RecordKindAudio {
		return nil, fmt.Errorf("only videos and audio have segments")
	}
	if edit.StartSeconds == nil {
		return nil, fmt.Errorf("a segment needs a start")
	}
	fields := segmentFields(edit)
	if err := checkRange(fields); err != nil {
		return nil, err
	}
	fields["subjectKind"] = kind
	fields["subjectId"] = id
	fields["origin"] = schema.SegmentOriginUser
	fillListDefaults(fields, "performers")
	segmentID, err := l.database.Insert(ctx, "Segment", fields)
	if err != nil {
		return nil, err
	}
	return l.Segment(ctx, segmentID)
}

// EditSegment changes a segment; a plugin's segment becomes the user's, so
// the next scrape doesn't replace it.
func (l *Library) EditSegment(ctx context.Context, id string, edit SegmentEdit) (*schema.Segment, error) {
	current, err := l.Segment(ctx, id)
	if err != nil || current == nil {
		return nil, fmt.Errorf("no segment %s", id)
	}
	fields := segmentFields(edit)
	fields["origin"] = schema.SegmentOriginUser
	merged := db.Vars{"startSeconds": current.StartSeconds, "endSeconds": current.EndSeconds}
	for column, value := range fields {
		merged[column] = value
	}
	if err := checkRange(merged); err != nil {
		return nil, err
	}
	if _, err := l.recorder.Update(ctx, schema.RecordKindSegment, id, fields, provenance.User()); err != nil {
		return nil, err
	}
	return l.Segment(ctx, id)
}

// segmentFields turns a segment edit into columns.
func segmentFields(edit SegmentEdit) db.Vars {
	fields := db.Vars{}
	if edit.StartSeconds != nil {
		fields["startSeconds"] = *edit.StartSeconds
	}
	if edit.EndSeconds != nil {
		fields["endSeconds"] = *edit.EndSeconds
	}
	if edit.ClearEnd {
		fields["endSeconds"] = nil
	}
	setOptionalText(fields, "title", edit.Title)
	setOptionalText(fields, "tag", edit.TagID)
	if edit.PerformerIDs != nil {
		fields["performers"] = *edit.PerformerIDs
	}
	return fields
}

// checkRange fails for a negative start or an end before the start.
func checkRange(fields db.Vars) error {
	start, _ := fields["startSeconds"].(float64)
	if start < 0 {
		return fmt.Errorf("a segment can't start before 0")
	}
	end, hasEnd := fields["endSeconds"].(float64)
	if pointer, isPointer := fields["endSeconds"].(*float64); isPointer && pointer != nil {
		end, hasEnd = *pointer, true
	}
	if hasEnd && end <= start {
		return fmt.Errorf("a segment must end after it starts")
	}
	return nil
}

// ClipEdit is a new clip or a change to one.
type ClipEdit struct {
	StartSeconds *float64
	EndSeconds   *float64
	Title        *string
	CoverURI     *string
}

// Clips lists the user's clips, newest first, of one video or of all.
func (l *Library) Clips(ctx context.Context, videoID string, limit, offset int) ([]schema.Clip, error) {
	vars := db.Vars{"video": videoID}
	vars["limit"], vars["offset"] = pageBounds(limit, offset)
	condition := ``
	if videoID != "" {
		condition = ` AND "video" = $video`
	}
	return db.QueryAs[schema.Clip](ctx, l.database,
		`SELECT * FROM "Clip" WHERE "deletedAt" IS NULL`+condition+` ORDER BY "createdAt" DESC LIMIT $limit OFFSET $offset`, vars)
}

// Clip loads one clip.
func (l *Library) Clip(ctx context.Context, id string) (*schema.Clip, error) {
	return db.Get[schema.Clip](ctx, l.database, "Clip", id)
}

// AddClip saves a moment of a video as a clip.
func (l *Library) AddClip(ctx context.Context, videoID string, edit ClipEdit) (*schema.Clip, error) {
	if edit.StartSeconds == nil || edit.EndSeconds == nil {
		return nil, fmt.Errorf("a clip needs a start and an end")
	}
	if *edit.StartSeconds < 0 || *edit.EndSeconds <= *edit.StartSeconds {
		return nil, fmt.Errorf("a clip must end after it starts")
	}
	fields := clipFields(edit)
	fields["video"] = videoID
	id, err := l.database.Insert(ctx, "Clip", fields)
	if err != nil {
		return nil, err
	}
	return l.Clip(ctx, id)
}

// EditClip changes a clip's range, title or cover.
func (l *Library) EditClip(ctx context.Context, id string, edit ClipEdit) (*schema.Clip, error) {
	if _, err := l.recorder.Update(ctx, schema.RecordKindClip, id, clipFields(edit), provenance.User()); err != nil {
		return nil, err
	}
	return l.Clip(ctx, id)
}

// AttachClipFile records the exported file of a clip (made by a client,
// since the backend has no encoder).
func (l *Library) AttachClipFile(ctx context.Context, clipID, fileID string) (*schema.Clip, error) {
	if _, err := l.recorder.Update(ctx, schema.RecordKindClip, clipID, db.Vars{"file": fileID}, provenance.User()); err != nil {
		return nil, err
	}
	return l.Clip(ctx, clipID)
}

// clipFields turns a clip edit into columns.
func clipFields(edit ClipEdit) db.Vars {
	fields := db.Vars{}
	if edit.StartSeconds != nil {
		fields["startSeconds"] = *edit.StartSeconds
	}
	if edit.EndSeconds != nil {
		fields["endSeconds"] = *edit.EndSeconds
	}
	setOptionalText(fields, "title", edit.Title)
	if edit.CoverURI != nil {
		fields["cover"] = assetOf(*edit.CoverURI)
	}
	return fields
}

// Subtitles lists a record's subtitle tracks by language.
func (l *Library) Subtitles(ctx context.Context, kind schema.RecordKind, id string) ([]schema.Subtitle, error) {
	return db.QueryAs[schema.Subtitle](ctx, l.database,
		`SELECT * FROM "Subtitle" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL ORDER BY "language", "createdAt"`,
		db.Vars{"kind": kind, "id": id})
}

// AddSubtitle stores a subtitle track for a video or audio: an uploaded
// file or a URL, generated or translated by a client.
func (l *Library) AddSubtitle(ctx context.Context, kind schema.RecordKind, id string, subtitle schema.Subtitle) (*schema.Subtitle, error) {
	if strings.TrimSpace(subtitle.Asset.Uri) == "" {
		return nil, fmt.Errorf("a subtitle needs a file")
	}
	subtitle.SubjectKind = kind
	subtitle.SubjectId = id
	subtitle.Language = defaultText(subtitle.Language, "und")
	subtitle.Format = defaultText(strings.ToLower(subtitle.Format), "vtt")
	if subtitle.Origin == "" {
		subtitle.Origin = schema.SubtitleOriginGenerated
	}
	subtitleID, err := l.database.Insert(ctx, "Subtitle", subtitle)
	if err != nil {
		return nil, err
	}
	return db.Get[schema.Subtitle](ctx, l.database, "Subtitle", subtitleID)
}

// RemoveSubtitle trashes a subtitle track.
func (l *Library) RemoveSubtitle(ctx context.Context, id string) error {
	return l.database.Trash(ctx, "Subtitle", id)
}

// Scripts lists a record's motion scripts.
func (l *Library) Scripts(ctx context.Context, kind schema.RecordKind, id string) ([]schema.Script, error) {
	return db.QueryAs[schema.Script](ctx, l.database,
		`SELECT * FROM "Script" WHERE "subjectKind" = $kind AND "subjectId" = $id AND "deletedAt" IS NULL ORDER BY "axis", "createdAt"`,
		db.Vars{"kind": kind, "id": id})
}

// AddScript stores the user's funscript for a video or audio.
func (l *Library) AddScript(ctx context.Context, kind schema.RecordKind, id, assetURI string, axis *string) (*schema.Script, error) {
	if strings.TrimSpace(assetURI) == "" {
		return nil, fmt.Errorf("a script needs a file")
	}
	script := schema.Script{
		SubjectKind: kind, SubjectId: id, Kind: schema.ScriptKindFunscript,
		Asset: schema.Asset{Uri: strings.TrimSpace(assetURI)}, Axis: axis, WrittenBy: schema.ActorUser,
	}
	scriptID, err := l.database.Insert(ctx, "Script", script)
	if err != nil {
		return nil, err
	}
	return db.Get[schema.Script](ctx, l.database, "Script", scriptID)
}

// RemoveScript trashes a motion script.
func (l *Library) RemoveScript(ctx context.Context, id string) error {
	return l.database.Trash(ctx, "Script", id)
}

// SetTrickplay stores a video's scrubbing thumbnails, made by a client
// (the backend can't decode video); nil removes them.
func (l *Library) SetTrickplay(ctx context.Context, videoID string, trickplay *schema.Trickplay) (*schema.Video, error) {
	if trickplay != nil && (len(trickplay.Sheets) == 0 || trickplay.Columns <= 0 || trickplay.Rows <= 0 || trickplay.IntervalSeconds <= 0) {
		return nil, fmt.Errorf("trickplay needs sheets, a grid and an interval")
	}
	fields := db.Vars{"trickplay": nil}
	if trickplay != nil {
		fields["trickplay"] = trickplay
	}
	if _, err := l.recorder.Update(ctx, schema.RecordKindVideo, videoID, fields, provenance.System()); err != nil {
		return nil, err
	}
	return l.Video(ctx, videoID)
}

// SetStreamSpeed records the throughput a client measured on a stream, so
// streams rank by it.
func (l *Library) SetStreamSpeed(ctx context.Context, streamID string, kilobitsPerSecond int) error {
	if kilobitsPerSecond <= 0 {
		return fmt.Errorf("a speed is a positive number of kbit/s")
	}
	return l.database.Update(ctx, tableStream, streamID, db.Vars{"speedKbps": kilobitsPerSecond})
}

// Stream loads one stream.
func (l *Library) Stream(ctx context.Context, id string) (*schema.Stream, error) {
	return db.Get[schema.Stream](ctx, l.database, tableStream, id)
}

// MarkUnread marks every source of a record unread, so the next visit reads
// their pages again (fresh streams, credits and comments).
func (l *Library) MarkUnread(ctx context.Context, kind schema.RecordKind, id string) error {
	_, err := l.database.Exec(ctx,
		`UPDATE "Source" SET "fetchedAt" = NULL WHERE "subjectKind" = $kind AND "subjectId" = $id`,
		db.Vars{"kind": kind, "id": id})
	return err
}
