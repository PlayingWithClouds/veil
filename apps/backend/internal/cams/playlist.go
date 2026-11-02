package cams

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net/url"
	"strings"
	"time"

	"github.com/playingwithclouds/veil/internal/livebuffer"
)

// playlistName is the recording's entry playlist, the one its MediaFile names.
const playlistName = "index.m3u8"

// flushPlaylist writes the playlists so far, so a crash keeps the footage.
func (s *Service) flushPlaylist(recording *capture) {
	if err := s.writePlaylist(recording.context, recording, false); err != nil {
		log.Printf("cams: record %s: write playlist: %v", recording.camModel.Username, err)
		return
	}
	recording.mutex.Lock()
	defer recording.mutex.Unlock()
	recording.video.flushedSegments = len(recording.video.segments)
	if recording.audio != nil {
		recording.audio.flushedSegments = len(recording.audio.segments)
	}
}

// writePlaylist stores the recording's VOD playlists; ended closes them. A
// stream with a separate audio rendition gets one playlist per track and a
// master (index.m3u8) joining them; otherwise index.m3u8 is the one playlist.
func (s *Service) writePlaylist(ctx context.Context, recording *capture, ended bool) error {
	recording.mutex.Lock()
	videoText := renderPlaylist(recording.video.segments, ended)
	audioText := ""
	if recording.audio != nil && len(recording.audio.segments) > 0 {
		audioText = renderPlaylist(recording.audio.segments, ended)
	}
	recording.mutex.Unlock()

	directory := recording.directory()
	if audioText == "" {
		return s.putText(ctx, directory+playlistName, videoText)
	}
	if err := s.putText(ctx, directory+recording.video.playlistName, videoText); err != nil {
		return err
	}
	if err := s.putText(ctx, directory+recording.audio.playlistName, audioText); err != nil {
		return err
	}
	return s.putText(ctx, directory+playlistName, renderMaster(recording.video.playlistName, recording.audio.playlistName))
}

// putText stores a playlist blob.
func (s *Service) putText(ctx context.Context, key, text string) error {
	return s.services.Blobs.Put(ctx, key, strings.NewReader(text), int64(len(text)), "application/x-mpegURL")
}

// renderMaster builds the master playlist joining a video and an audio playlist.
func renderMaster(videoPlaylist, audioPlaylist string) string {
	return "#EXTM3U\n#EXT-X-VERSION:7\n" +
		fmt.Sprintf("#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"audio\",NAME=\"Audio\",DEFAULT=YES,AUTOSELECT=YES,URI=%q\n", audioPlaylist) +
		"#EXT-X-STREAM-INF:BANDWIDTH=5000000,AUDIO=\"audio\"\n" + videoPlaylist + "\n"
}

// renderPlaylist builds an HLS playlist over stored segments with relative URIs.
func renderPlaylist(segments []storedSegment, ended bool) string {
	longest := 1.0
	for _, segment := range segments {
		longest = math.Max(longest, segment.duration)
	}
	var builder strings.Builder
	builder.WriteString("#EXTM3U\n#EXT-X-VERSION:7\n")
	fmt.Fprintf(&builder, "#EXT-X-TARGETDURATION:%d\n#EXT-X-MEDIA-SEQUENCE:0\n#EXT-X-PLAYLIST-TYPE:VOD\n", int(math.Ceil(longest)))
	currentInit := ""
	for _, segment := range segments {
		if segment.discontinuity {
			builder.WriteString("#EXT-X-DISCONTINUITY\n")
			currentInit = ""
		}
		if segment.initName != currentInit {
			currentInit = segment.initName
			writeMap(&builder, currentInit)
		}
		fmt.Fprintf(&builder, "#EXTINF:%.3f,\n%s\n", segment.duration, segment.name)
	}
	if ended {
		builder.WriteString("#EXT-X-ENDLIST\n")
	}
	return builder.String()
}

// writeMap writes the EXT-X-MAP tag of an init section, nothing for none.
func writeMap(builder *strings.Builder, initName string) {
	if initName == "" {
		return
	}
	fmt.Fprintf(builder, "#EXT-X-MAP:URI=%q\n", initName)
}

// flushedAt returns when the recording's playlist was last written, and false
// when it has none (nothing was flushed).
func (s *Service) flushedAt(ctx context.Context, jobID string) (time.Time, bool) {
	files, err := s.services.Blobs.List(ctx, blobDirectory(jobID))
	if err != nil {
		return time.Time{}, false
	}
	for _, file := range files {
		if file.Key == blobDirectory(jobID)+playlistName {
			return file.LastUsed, true
		}
	}
	return time.Time{}, false
}

// restoreSegments reads the flushed playlists of an interrupted attempt back
// into the capture, so it goes on from there.
func (s *Service) restoreSegments(ctx context.Context, recording *capture) error {
	directory := recording.directory()
	indexText, err := s.readText(ctx, directory+playlistName)
	if err != nil {
		return err
	}
	sizes := map[string]int64{}
	files, _ := s.services.Blobs.List(ctx, directory)
	for _, file := range files {
		sizes[strings.TrimPrefix(file.Key, directory)] = file.Size
	}
	if !strings.Contains(indexText, "#EXT-X-STREAM-INF") {
		return restoreTrack(recording.video, indexText, sizes)
	}
	return s.restoreTracks(ctx, recording, sizes)
}

// restoreTracks restores a capture stored as separate video and audio playlists.
func (s *Service) restoreTracks(ctx context.Context, recording *capture, sizes map[string]int64) error {
	directory := recording.directory()
	videoText, err := s.readText(ctx, directory+recording.video.playlistName)
	if err != nil {
		return err
	}
	if err := restoreTrack(recording.video, videoText, sizes); err != nil {
		return err
	}
	audio := newAudioTrack()
	audioText, err := s.readText(ctx, directory+audio.playlistName)
	if err != nil {
		return nil
	}
	recording.audio = audio
	return restoreTrack(audio, audioText, sizes)
}

// readText reads a stored text blob.
func (s *Service) readText(ctx context.Context, key string) (string, error) {
	reader, err := s.services.Blobs.Get(ctx, key)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	return string(data), err
}

// restoreTrack fills a track from the text of its flushed playlist. The
// restored footage ends in a break: whatever follows was recorded later.
func restoreTrack(restored *track, playlistText string, sizes map[string]int64) error {
	base, _ := url.Parse("http://recording.invalid/")
	playlist, err := livebuffer.ParsePlaylist(playlistText, base)
	if err != nil {
		return err
	}
	for _, segment := range playlist.Segments {
		name := strings.TrimPrefix(segment.Identity, "/")
		initName := ""
		if segment.InitURL != "" {
			initName = strings.TrimPrefix(livebuffer.URLIdentity(segment.InitURL), "/")
			restored.initNames[initName] = initName
		}
		restored.segments = append(restored.segments, storedSegment{
			name: name, duration: segment.Duration, initName: initName, discontinuity: segment.Discontinuity,
		})
		restored.durationSeconds += segment.Duration
		restored.sizeBytes += sizes[name]
	}
	restored.flushedSegments = len(restored.segments)
	restored.pendingBreak = len(restored.segments) > 0
	return nil
}
