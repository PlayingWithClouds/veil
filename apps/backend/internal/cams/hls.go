package cams

import (
	"bufio"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/playingwithclouds/veil/internal/livebuffer"
)

// playlistSegment, mediaPlaylist and the playlist helpers are shared with the
// live buffer, which reads the same playlists.
type (
	playlistSegment = livebuffer.Segment
	mediaPlaylist   = livebuffer.Playlist
)

// parseMediaPlaylist reads a media playlist, resolving every URI against base.
func parseMediaPlaylist(text string, base *url.URL) (*mediaPlaylist, error) {
	return livebuffer.ParsePlaylist(text, base)
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

// attributeValue returns the value of NAME in an HLS attribute list.
func attributeValue(line, name string) string {
	return livebuffer.AttributeValue(line, name)
}

// resolveReference makes reference absolute against base.
func resolveReference(base *url.URL, reference string) string {
	return livebuffer.ResolveReference(base, reference)
}

// urlIdentity is the URL's path, without host or query.
func urlIdentity(rawURL string) string {
	return livebuffer.URLIdentity(rawURL)
}
