package cams

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/playingwithclouds/veil/internal/db"
	"github.com/playingwithclouds/veil/internal/provenance"
	"github.com/playingwithclouds/veil/internal/schema"
)

const (
	// minimumRecordingSeconds is the shortest recording kept as a video.
	minimumRecordingSeconds = 5.0
	// posterName is the room thumbnail kept with a recording.
	posterName = "poster.jpg"
	// liveCamTag is the tag every recording carries.
	liveCamTag = "Live cam"
	// recordingTagLimit caps how many of the room's own tags a recording takes.
	recordingTagLimit = 8
	// blobAssetScheme marks an asset stored in this device's blob store.
	blobAssetScheme = "blob://"
	// recordingMimeType is the type of a recording's playlist.
	recordingMimeType = "application/x-mpegURL"
)

// errTooShort is returned for a capture with too little footage to keep.
var errTooShort = errors.New("recording too short to keep")

// finish writes the final playlist and files the capture as a video, or
// deletes it when it holds too little footage. It returns the video's id.
func (s *Service) finish(ctx context.Context, recording *capture) (string, error) {
	duration, _ := recording.totals()
	if duration < minimumRecordingSeconds {
		if err := s.services.Blobs.DeletePrefix(ctx, recording.directory()); err != nil {
			log.Printf("cams: discard recording %s: %v", recording.jobID, err)
		}
		return "", errTooShort
	}
	if err := s.writePlaylist(ctx, recording, true); err != nil {
		return "", fmt.Errorf("write playlist: %w", err)
	}
	return s.fileVideo(ctx, recording)
}

// fileVideo creates the recording's video and media file, unless an earlier
// attempt of the job already did.
func (s *Service) fileVideo(ctx context.Context, recording *capture) (string, error) {
	location, err := s.libraryLocation(ctx)
	if err != nil {
		return "", err
	}
	filePath := recording.directory() + playlistName
	existing, err := s.services.Database.Strings(ctx,
		`SELECT "subjectId" FROM "MediaFile" WHERE "location" = $location AND "path" = $path AND "deletedAt" IS NULL`,
		db.Vars{"location": location.Id, "path": filePath})
	if err != nil || len(existing) > 0 {
		return firstOrEmpty(existing), err
	}
	fields, err := s.recordingFields(ctx, recording)
	if err != nil {
		return "", err
	}
	videoID, err := s.insertRecording(ctx, recording, fields, location.Id, filePath)
	if err != nil {
		return "", err
	}
	return videoID, s.services.Library.Recorder().Claim(ctx, schema.RecordKindVideo, videoID, columnsOf(fields), provenance.System())
}

// libraryLocation is this device's library storage location, where the blob
// store's files are filed.
func (s *Service) libraryLocation(ctx context.Context) (*schema.StorageLocation, error) {
	locations, err := s.services.Devices.Locations(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, location := range locations {
		if location.Kind == schema.StorageKindLibrary {
			return &location, nil
		}
	}
	return nil, errors.New("this device has no library storage location")
}

// insertRecording writes the video and its media file in one transaction.
func (s *Service) insertRecording(ctx context.Context, recording *capture, fields db.Vars, locationID, filePath string) (string, error) {
	duration, size := recording.totals()
	var videoID string
	err := s.services.Database.Tx(ctx, func(tx *db.DB) error {
		var err error
		videoID, err = tx.Insert(ctx, "Video", withRecordingColumns(fields, recording.camModel.Id))
		if err != nil {
			return err
		}
		_, err = tx.Insert(ctx, "MediaFile", schema.MediaFile{
			SubjectKind: schema.RecordKindVideo, SubjectId: videoID, Location: locationID, Path: filePath,
			SizeBytes: size, MimeType: recordingMimeType, Quality: optionalText(recording.quality()),
			Height: recording.knownHeight(), DurationSeconds: &duration,
		})
		return err
	})
	return videoID, err
}

// withRecordingColumns adds the columns that aren't descriptive (so no source
// owns them) to a recording's fields.
func withRecordingColumns(fields db.Vars, camModelID string) db.Vars {
	row := db.Vars{"status": schema.MediaStatusAvailable, "recordedFrom": camModelID}
	for column, value := range fields {
		row[column] = value
	}
	return row
}

// knownHeight is the recorded rendition's height, nil when the master didn't say.
func (recording *capture) knownHeight() *int32 {
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	if recording.height <= 0 {
		return nil
	}
	height := int32(recording.height)
	return &height
}

// recordingFields are the descriptive columns of a recording's video: a title
// naming the room and start, the cam site as studio, the room's linked
// performer, "Live cam" and some of the room's tags, its thumbnail as cover.
func (s *Service) recordingFields(ctx context.Context, recording *capture) (db.Vars, error) {
	camModel := s.latestModel(ctx, recording)
	duration, _ := recording.totals()
	fields := db.Vars{
		"title":           fmt.Sprintf("%s — live %s", camModel.DisplayName, recording.startedAt.Local().Format("2006-01-02 15:04")),
		"date":            recording.startedAt.UTC().Format("2006-01-02"),
		"durationSeconds": math.Round(duration),
		"performers":      creditedPerformers(camModel),
	}
	tagIDs, err := s.services.Library.TagIDs(ctx, recordingTags(camModel))
	if err != nil {
		return nil, err
	}
	fields["tags"] = tagIDs
	studioID, err := s.services.Library.StudioID(ctx, s.siteName(camModel.Plugin))
	if err != nil {
		return nil, err
	}
	if studioID != "" {
		fields["studio"] = studioID
	}
	if cover := s.storePoster(ctx, recording, camModel); cover != nil {
		fields["cover"] = cover
	}
	return fields, nil
}

// latestModel reloads the recorded room, whose name, tags and performer may
// have changed while it was recorded.
func (s *Service) latestModel(ctx context.Context, recording *capture) schema.CamModel {
	latest, err := s.load(ctx, recording.camModel.Id)
	if err != nil {
		return recording.camModel
	}
	return *latest
}

// creditedPerformers is the room's linked performer, if any.
func creditedPerformers(camModel schema.CamModel) []string {
	if camModel.Performer == nil {
		return []string{}
	}
	return []string{*camModel.Performer}
}

// recordingTags are the tag names of a recording: "Live cam" and some of
// the tags the room carried.
func recordingTags(camModel schema.CamModel) []string {
	tags := []string{liveCamTag}
	for _, tag := range camModel.Tags {
		if len(tags) > recordingTagLimit {
			break
		}
		tags = append(tags, tag)
	}
	return tags
}

// siteName is the cam site's shown name, the plugin name when it has none.
func (s *Service) siteName(pluginName string) string {
	plugin, ok := s.services.Sites.Plugin(pluginName)
	if !ok || plugin.Meta.DisplayName == "" {
		return pluginName
	}
	return plugin.Meta.DisplayName
}

// storePoster keeps the room's thumbnail with the recording, since site
// thumbnails change and expire, and returns it as the cover; nil when there
// is none to fetch.
func (s *Service) storePoster(ctx context.Context, recording *capture, camModel schema.CamModel) *schema.Asset {
	if camModel.Thumbnail == nil || !strings.HasPrefix(camModel.Thumbnail.Uri, "http") {
		return nil
	}
	data, err := s.fetchBytes(ctx, camModel.Thumbnail.Uri, nil)
	if err != nil || len(data) == 0 {
		return nil
	}
	key := recording.directory() + posterName
	if err := s.services.Blobs.Put(ctx, key, bytes.NewReader(data), int64(len(data)), "image/jpeg"); err != nil {
		return nil
	}
	mimeType := "image/jpeg"
	return &schema.Asset{Uri: blobAssetScheme + key, MimeType: &mimeType}
}

// columnsOf lists the columns fields sets.
func columnsOf(fields db.Vars) []string {
	columns := make([]string, 0, len(fields))
	for column := range fields {
		columns = append(columns, column)
	}
	return columns
}

// firstOrEmpty returns the first value, "" for none.
func firstOrEmpty(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}
