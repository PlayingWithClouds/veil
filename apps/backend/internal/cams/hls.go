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

// streamVariant is a rendition of a master playlist.
type streamVariant struct {
	// URL is the variant's media playlist.
	URL string
	// AudioURL is the media playlist of the separate audio rendition the variant
	// plays with, or "" when its audio is muxed into the video segments.
	AudioURL string
	// Height is the video height in pixels, 0 when the master does not say.
	Height int
	// Bandwidth is the advertised bitrate in bits per second.
	Bandwidth int
}

// Label names the rendition the way players do: "720p", or its bitrate when
// the height is unknown.
func (variant streamVariant) Label() string {
	if variant.Height > 0 {
		return fmt.Sprintf("%dp", variant.Height)
	}
	return fmt.Sprintf("%d kbps", variant.Bandwidth/1000)
}

// parseVariants lists the renditions of a master playlist, each with its audio
// rendition; it returns nothing for a media playlist.
func parseVariants(text string, base *url.URL) []streamVariant {
	audioGroups := audioRenditions(text, base)
	var variants []streamVariant
	var pending *streamVariant
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			bandwidth, _ := strconv.Atoi(attributeValue(line, "BANDWIDTH"))
			pending = &streamVariant{
				Bandwidth: bandwidth, Height: resolutionHeight(attributeValue(line, "RESOLUTION")),
				AudioURL: audioGroups[attributeValue(line, "AUDIO")],
			}
			continue
		}
		if pending == nil || line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pending.URL = resolveReference(base, line)
		variants = append(variants, *pending)
		pending = nil
	}
	return variants
}

// resolutionHeight reads the height of a "1280x720" resolution, 0 when absent.
func resolutionHeight(resolution string) int {
	_, height, found := strings.Cut(resolution, "x")
	if !found {
		return 0
	}
	value, _ := strconv.Atoi(height)
	return value
}

// chooseVariant picks the rendition to record. targetHeight 0 asks for the best
// one (highest bitrate). Otherwise it is the tallest rendition at or below the
// target, or, when none is that small, the shortest one above it. Renditions
// without a known height only count when no rendition has one. It returns nil
// for a media playlist.
func chooseVariant(variants []streamVariant, targetHeight int) *streamVariant {
	if len(variants) == 0 {
		return nil
	}
	known := variantsWithHeight(variants)
	if targetHeight <= 0 || len(known) == 0 {
		return highestBandwidth(variants)
	}
	var atOrBelow, above []streamVariant
	for _, variant := range known {
		if variant.Height <= targetHeight {
			atOrBelow = append(atOrBelow, variant)
			continue
		}
		above = append(above, variant)
	}
	if len(atOrBelow) > 0 {
		return tallest(atOrBelow)
	}
	return shortest(above)
}

// variantsWithHeight keeps the renditions whose height is known.
func variantsWithHeight(variants []streamVariant) []streamVariant {
	var known []streamVariant
	for _, variant := range variants {
		if variant.Height > 0 {
			known = append(known, variant)
		}
	}
	return known
}

// highestBandwidth returns the rendition with the highest bitrate.
func highestBandwidth(variants []streamVariant) *streamVariant {
	best := variants[0]
	for _, variant := range variants[1:] {
		if variant.Bandwidth > best.Bandwidth {
			best = variant
		}
	}
	return &best
}

// tallest returns the highest rendition, the one with more bitrate on a tie.
func tallest(variants []streamVariant) *streamVariant {
	best := variants[0]
	for _, variant := range variants[1:] {
		if variant.Height > best.Height || (variant.Height == best.Height && variant.Bandwidth > best.Bandwidth) {
			best = variant
		}
	}
	return &best
}

// shortest returns the lowest rendition, the one with more bitrate on a tie.
func shortest(variants []streamVariant) *streamVariant {
	best := variants[0]
	for _, variant := range variants[1:] {
		if variant.Height < best.Height || (variant.Height == best.Height && variant.Bandwidth > best.Bandwidth) {
			best = variant
		}
	}
	return &best
}

// QualitySource asks for the best rendition the stream offers.
const QualitySource = "source"

// ParseQuality reads a quality: "source" (or empty) is the best rendition, "480"
// or "480p" a target height. It returns the target height, 0 for source.
func ParseQuality(quality string) (int, error) {
	trimmed := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(quality)), "p")
	if trimmed == "" || trimmed == QualitySource || trimmed == "best" {
		return 0, nil
	}
	height, err := strconv.Atoi(trimmed)
	if err != nil || height <= 0 {
		return 0, fmt.Errorf("invalid quality %q: use source or a height such as 480", quality)
	}
	return height, nil
}

// NormalizeQuality is the stored form of a quality: "source" or the height.
func NormalizeQuality(quality string) (string, error) {
	height, err := ParseQuality(quality)
	if err != nil {
		return "", err
	}
	if height == 0 {
		return QualitySource, nil
	}
	return strconv.Itoa(height), nil
}

// audioRenditions maps each audio group of a master playlist to the playlist of
// its default rendition (the first one when none is marked default).
func audioRenditions(text string, base *url.URL) map[string]string {
	renditions := map[string]string{}
	isDefault := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#EXT-X-MEDIA:") || attributeValue(line, "TYPE") != "AUDIO" {
			continue
		}
		uri := attributeValue(line, "URI")
		group := attributeValue(line, "GROUP-ID")
		if uri == "" {
			continue
		}
		_, taken := renditions[group]
		wanted := attributeValue(line, "DEFAULT") == "YES"
		if !taken || (wanted && !isDefault[group]) {
			renditions[group] = resolveReference(base, uri)
			isDefault[group] = wanted
		}
	}
	return renditions
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
