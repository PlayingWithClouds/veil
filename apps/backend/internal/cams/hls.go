package cams

import (
	"bufio"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// playlistSegment is one media segment of a live playlist.
type playlistSegment struct {
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

// mediaPlaylist is a parsed HLS media playlist.
type mediaPlaylist struct {
	Segments       []playlistSegment
	TargetDuration float64
	// Ended is set once the broadcaster closed the stream (EXT-X-ENDLIST).
	Ended bool
}

// parseMediaPlaylist reads a media playlist, resolving every URI against base.
// Encrypted streams are rejected because the recording keeps segments as sent.
func parseMediaPlaylist(text string, base *url.URL) (*mediaPlaylist, error) {
	playlist := &mediaPlaylist{}
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
			initURL = resolveReference(base, attributeValue(line, "URI"))
		case strings.HasPrefix(line, "#EXT-X-TARGETDURATION:"):
			playlist.TargetDuration, _ = strconv.ParseFloat(strings.TrimPrefix(line, "#EXT-X-TARGETDURATION:"), 64)
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
			segmentURL := resolveReference(base, line)
			playlist.Segments = append(playlist.Segments, playlistSegment{
				URL: segmentURL, Identity: urlIdentity(segmentURL), Duration: pendingDuration,
				InitURL: initURL, Discontinuity: pendingDiscontinuity,
			})
			pendingDuration = 0
			pendingDiscontinuity = false
		}
	}
	return playlist, scanner.Err()
}

// bestVariantURL returns the URL of the highest-bandwidth variant of a master
// playlist, or "" when text is already a media playlist.
func bestVariantURL(text string, base *url.URL) string {
	bestBandwidth := -1
	bestURL := ""
	expectingURI := false
	currentBandwidth := 0
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			expectingURI = true
			currentBandwidth, _ = strconv.Atoi(attributeValue(line, "BANDWIDTH"))
			continue
		}
		if !expectingURI || line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		expectingURI = false
		if currentBandwidth > bestBandwidth {
			bestBandwidth = currentBandwidth
			bestURL = resolveReference(base, line)
		}
	}
	return bestURL
}

// attributeValue returns the value of NAME in an HLS attribute list, without
// quotes, or "" when absent.
func attributeValue(line, name string) string {
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

// resolveReference makes reference absolute against base.
func resolveReference(base *url.URL, reference string) string {
	if reference == "" {
		return ""
	}
	parsed, err := url.Parse(reference)
	if err != nil {
		return reference
	}
	return base.ResolveReference(parsed).String()
}

// urlIdentity is the URL's path: no host (a refreshed stream URL may point at
// another edge server), no query and no fragment.
func urlIdentity(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return parsed.Path
}
