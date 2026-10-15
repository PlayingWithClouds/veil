package livebuffer

import (
	"bufio"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Segment is one media segment of a live playlist.
type Segment struct {
	// URL is the absolute segment URL, including any access token.
	URL string
	// Identity names the segment independently of the token and host of the
	// stream URL, so the same segment seen after a reconnect is recognised.
	Identity string
	Duration float64
	// InitURL is the absolute URL of the fMP4 initialization section the
	// segment needs, or "" for MPEG-TS.
	InitURL string
	// Discontinuity marks a break in timestamps or encoding before the segment.
	Discontinuity bool
}

// Playlist is a parsed HLS media playlist.
type Playlist struct {
	Segments       []Segment
	TargetDuration float64
	// MediaSequence is the sequence number of the first segment; the origin
	// numbers a stream's renditions alike, which players rely on to switch between them.
	MediaSequence int
	// Ended is set once the broadcaster closed the stream (EXT-X-ENDLIST).
	Ended bool
}

// ParsePlaylist reads a media playlist, resolving every URI against base.
// Encrypted streams are rejected because segments are kept as sent.
func ParsePlaylist(text string, base *url.URL) (*Playlist, error) {
	playlist := &Playlist{}
	var pendingDuration float64
	pendingDiscontinuity := false
	initURL := ""
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "":
		case strings.HasPrefix(line, "#EXT-X-KEY:"):
			if !strings.Contains(line, "METHOD=NONE") {
				return nil, fmt.Errorf("encrypted streams cannot be recorded")
			}
		case strings.HasPrefix(line, "#EXT-X-MAP:"):
			initURL = ResolveReference(base, AttributeValue(line, "URI"))
		case strings.HasPrefix(line, "#EXT-X-TARGETDURATION:"):
			playlist.TargetDuration, _ = strconv.ParseFloat(strings.TrimPrefix(line, "#EXT-X-TARGETDURATION:"), 64)
		case strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"):
			playlist.MediaSequence, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "#EXT-X-MEDIA-SEQUENCE:")))
		case strings.HasPrefix(line, "#EXT-X-DISCONTINUITY-SEQUENCE"):
		case strings.HasPrefix(line, "#EXT-X-DISCONTINUITY"):
			pendingDiscontinuity = true
		case strings.HasPrefix(line, "#EXTINF:"):
			durationText, _, _ := strings.Cut(strings.TrimPrefix(line, "#EXTINF:"), ",")
			pendingDuration, _ = strconv.ParseFloat(strings.TrimSpace(durationText), 64)
		case line == "#EXT-X-ENDLIST":
			playlist.Ended = true
		case strings.HasPrefix(line, "#"):
		default:
			segmentURL := ResolveReference(base, line)
			playlist.Segments = append(playlist.Segments, Segment{
				URL: segmentURL, Identity: URLIdentity(segmentURL), Duration: pendingDuration,
				InitURL: initURL, Discontinuity: pendingDiscontinuity,
			})
			pendingDuration = 0
			pendingDiscontinuity = false
		}
	}
	return playlist, scanner.Err()
}

// AttributeValue returns the value of NAME in an HLS attribute list, without
// quotes, or "" when absent.
func AttributeValue(line, name string) string {
	_, attributes, _ := strings.Cut(line, ":")
	for attributes != "" {
		key, rest, found := strings.Cut(attributes, "=")
		if !found {
			return ""
		}
		var value string
		if strings.HasPrefix(rest, `"`) {
			quoted, remainder, _ := strings.Cut(rest[1:], `"`)
			value = quoted
			attributes = strings.TrimPrefix(remainder, ",")
		} else {
			unquoted, remainder, _ := strings.Cut(rest, ",")
			value = unquoted
			attributes = remainder
		}
		if strings.TrimSpace(key) == name {
			return value
		}
	}
	return ""
}

// ResolveReference makes reference absolute against base.
func ResolveReference(base *url.URL, reference string) string {
	if reference == "" {
		return ""
	}
	parsed, err := url.Parse(reference)
	if err != nil {
		return reference
	}
	return base.ResolveReference(parsed).String()
}

// URLIdentity is the URL's path: no host (a refreshed stream URL may point at
// another edge server), no query and no fragment.
func URLIdentity(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return parsed.Path
}
