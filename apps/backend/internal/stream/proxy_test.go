package stream

import (
	"net/url"
	"strings"
	"testing"
)

// TestRewriteManifestRootRelative checks that proxied segment and variant URIs
// carry no host, so clients reaching the backend under any address can play them.
func TestRewriteManifestRootRelative(t *testing.T) {
	proxy := NewProxy(nil, nil, "http://localhost:8080")
	base, _ := url.Parse("https://cdn.example/video/master.m3u8")
	manifest := strings.Join([]string{
		"#EXTM3U",
		"#EXT-X-STREAM-INF:BANDWIDTH=800000",
		"720p/index.m3u8",
		"#EXTINF:4.0,",
		"seg-1.ts",
	}, "\n")

	lines := strings.Split(proxy.rewriteManifest(manifest, base, "session"), "\n")

	if !strings.HasPrefix(lines[2], "/api/stream/manifest?s=session&u=") {
		t.Errorf("variant line = %q, want root-relative manifest proxy URL", lines[2])
	}
	if !strings.HasPrefix(lines[4], "/api/stream/seg?s=session&h=") {
		t.Errorf("segment line = %q, want root-relative segment proxy URL", lines[4])
	}
}

// TestRewriteManifestTagURIs checks that files referenced from tag attributes
// (fMP4 init segments, alternate renditions) are proxied too.
func TestRewriteManifestTagURIs(t *testing.T) {
	proxy := NewProxy(nil, nil, "http://localhost:8080")
	base, _ := url.Parse("https://cdn.example/video/720p.av1.mp4/index.m3u8")
	manifest := strings.Join([]string{
		"#EXTM3U",
		`#EXT-X-MAP:URI="init-v1-a1.mp4"`,
		`#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",URI="audio/index.m3u8"`,
		"#EXT-X-TARGETDURATION:4",
	}, "\n")

	lines := strings.Split(proxy.rewriteManifest(manifest, base, "session"), "\n")

	wantInit := `#EXT-X-MAP:URI="/api/stream/seg?s=session&h=`
	if !strings.HasPrefix(lines[1], wantInit) || !strings.Contains(lines[1], "&u="+encURL("https://cdn.example/video/720p.av1.mp4/init-v1-a1.mp4")+`"`) {
		t.Errorf("init line = %q", lines[1])
	}
	wantMedia := `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="audio",URI="/api/stream/manifest?s=session&u=` + encURL("https://cdn.example/video/720p.av1.mp4/audio/index.m3u8") + `"`
	if lines[2] != wantMedia {
		t.Errorf("media line = %q, want %q", lines[2], wantMedia)
	}
	if lines[3] != "#EXT-X-TARGETDURATION:4" {
		t.Errorf("plain tag changed: %q", lines[3])
	}
}
