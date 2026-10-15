package livebuffer

import (
	"fmt"
	"math"
	"strings"
)

// WindowSegment is one buffered segment of a Window.
type WindowSegment struct {
	// Name is the blob name to pass to Manager.Open.
	Name string
	// SourceURL is where the origin served the segment, the fallback when the
	// blob cannot be read.
	SourceURL     string
	Identity      string
	Duration      float64
	Discontinuity bool
	// InitName is the blob name of the fMP4 initialization section, "" if none.
	InitName string
}

// Window is the buffered stretch of one rendition, oldest segment first.
type Window struct {
	TrackKey              string
	Kind                  string
	Segments              []WindowSegment
	MediaSequence         int
	DiscontinuitySequence int
	TargetDuration        float64
	Ended                 bool
}

// window snapshots the rendition. Segments whose download failed are left out
// and the segment after one is marked discontinuous. The caller holds the lock.
func (rendition *track) window(buffered *room) *Window {
	snapshot := &Window{
		TrackKey: rendition.key, Kind: rendition.kind, TargetDuration: rendition.targetDuration,
		DiscontinuitySequence: rendition.discontinuitiesDropped, Ended: rendition.ended,
	}
	if len(rendition.entries) > 0 {
		snapshot.MediaSequence = rendition.entries[0].sequence
	}
	skipped := false
	for _, segment := range rendition.entries {
		if segment.file.failed {
			skipped = true
			continue
		}
		listed := WindowSegment{
			Name: segment.file.name, SourceURL: segment.sourceURL, Identity: segment.identity,
			Duration: segment.duration, Discontinuity: segment.discontinuity || skipped,
		}
		skipped = false
		if segment.initFile != nil {
			listed.InitName = segment.initFile.name
		}
		snapshot.Segments = append(snapshot.Segments, listed)
	}
	return snapshot
}

// Seconds is the footage the window holds.
func (window *Window) Seconds() float64 {
	total := 0.0
	for _, segment := range window.Segments {
		total += segment.Duration
	}
	return total
}

// NameOf returns the blob name of the buffered segment with the identity.
func (window *Window) NameOf(identity string) (string, bool) {
	for _, segment := range window.Segments {
		if segment.Identity == identity {
			return segment.Name, true
		}
	}
	return "", false
}

// Render writes the window as a sliding live playlist. uri turns a blob name
// (and the origin URL, "" for init sections) into the URI the player requests.
func (window *Window) Render(uri func(name, sourceURL string) string) string {
	longest := math.Max(window.TargetDuration, 1)
	for _, segment := range window.Segments {
		longest = math.Max(longest, segment.Duration)
	}
	var builder strings.Builder
	builder.WriteString("#EXTM3U\n#EXT-X-VERSION:7\n")
	fmt.Fprintf(&builder, "#EXT-X-TARGETDURATION:%d\n#EXT-X-MEDIA-SEQUENCE:%d\n#EXT-X-DISCONTINUITY-SEQUENCE:%d\n",
		int(math.Ceil(longest)), window.MediaSequence, window.DiscontinuitySequence)
	currentInit := ""
	for _, segment := range window.Segments {
		if segment.Discontinuity {
			builder.WriteString("#EXT-X-DISCONTINUITY\n")
			currentInit = ""
		}
		if segment.InitName != currentInit {
			currentInit = segment.InitName
			if currentInit != "" {
				fmt.Fprintf(&builder, "#EXT-X-MAP:URI=%q\n", uri(currentInit, ""))
			}
		}
		fmt.Fprintf(&builder, "#EXTINF:%.3f,\n%s\n", segment.Duration, uri(segment.Name, segment.SourceURL))
	}
	if window.Ended {
		builder.WriteString("#EXT-X-ENDLIST\n")
	}
	return builder.String()
}

// Windows lists the renditions of a room that currently hold segments.
func (m *Manager) Windows(roomID string) []*Window {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	buffered := m.rooms[roomIdentifier(roomID)]
	if buffered == nil {
		return nil
	}
	var windows []*Window
	for _, rendition := range buffered.tracks {
		snapshot := rendition.window(buffered)
		if len(snapshot.Segments) > 0 {
			windows = append(windows, snapshot)
		}
	}
	return windows
}

// BufferedSeconds is the longest stretch the room's video renditions hold,
// 0 when nothing is buffered.
func (m *Manager) BufferedSeconds(roomID string) float64 {
	best := 0.0
	for _, window := range m.Windows(roomID) {
		if window.Kind == KindAudio {
			continue
		}
		best = math.Max(best, window.Seconds())
	}
	return best
}

// Seed is the buffered tail a recording starts from.
type Seed struct {
	Video *Window
	// Audio is the separate audio rendition, nil when audio is muxed.
	Audio *Window
	// Seconds is the footage of the video part.
	Seconds float64
}

// SeedFrom picks the renditions to start a recording from and keeps the last
// secondsAgo seconds of them (whole segments, so the start can be up to one
// segment earlier). It returns nil when nothing is buffered. The video
// rendition is the one holding the most footage, the audio one the one observed last.
func (m *Manager) SeedFrom(roomID string, secondsAgo float64) *Seed {
	var video, audio *Window
	for _, window := range m.Windows(roomID) {
		if window.Kind == KindAudio {
			if audio == nil || m.observedAfter(roomID, window.TrackKey, audio.TrackKey) {
				audio = window
			}
			continue
		}
		if video == nil || window.Seconds() > video.Seconds() {
			video = window
		}
	}
	if video == nil || secondsAgo <= 0 {
		return nil
	}
	seed := &Seed{Video: video.tail(secondsAgo)}
	seed.Seconds = seed.Video.Seconds()
	if audio != nil {
		seed.Audio = audio.tail(seed.Seconds)
	}
	return seed
}

// observedAfter reports whether rendition a was observed more recently than b.
func (m *Manager) observedAfter(roomID, a, b string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	buffered := m.rooms[roomIdentifier(roomID)]
	if buffered == nil || buffered.tracks[a] == nil || buffered.tracks[b] == nil {
		return false
	}
	return buffered.tracks[a].lastObserved.After(buffered.tracks[b].lastObserved)
}

// tail keeps the newest segments covering at least seconds of footage.
func (window *Window) tail(seconds float64) *Window {
	kept := *window
	covered := 0.0
	start := len(window.Segments)
	for start > 0 && covered < seconds {
		start--
		covered += window.Segments[start].Duration
	}
	kept.Segments = window.Segments[start:]
	return &kept
}
