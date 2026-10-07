package pluginstore

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testPackage = "@playingwithclouds/veil-plugin-example"

// fakeRegistry serves one package's latest version like registry.npmjs.org.
type fakeRegistry struct {
	server    *httptest.Server
	version   string
	tarball   []byte
	integrity string
}

// newFakeRegistry starts a registry publishing testPackage at version.
func newFakeRegistry(t *testing.T, version string) *fakeRegistry {
	registry := &fakeRegistry{}
	registry.server = httptest.NewServer(http.HandlerFunc(registry.serve))
	t.Cleanup(registry.server.Close)
	registry.publish(t, version)
	return registry
}

// publish replaces the latest version with a freshly packed one.
func (registry *fakeRegistry) publish(t *testing.T, version string) {
	registry.version = version
	registry.tarball = packTarball(t, map[string]string{
		"package/package.json": manifestJSON(t, Manifest{Name: testPackage, Version: version, Main: "plugin.js", Keywords: []string{Keyword}}),
		"package/plugin.js":    "module.exports = " + version,
	})
	sum := sha512.Sum512(registry.tarball)
	registry.integrity = "sha512-" + base64.StdEncoding.EncodeToString(sum[:])
}

// serve answers the version manifest and tarball requests.
func (registry *fakeRegistry) serve(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.RawPath == "/@playingwithclouds%2fveil-plugin-example/latest":
		manifest := map[string]any{
			"name": testPackage, "version": registry.version, "main": "plugin.js",
			"keywords": []string{Keyword},
			"dist": map[string]string{
				"tarball":   registry.server.URL + "/example.tgz",
				"integrity": registry.integrity,
			},
		}
		_ = json.NewEncoder(w).Encode(manifest)
	case r.URL.Path == "/example.tgz":
		_, _ = w.Write(registry.tarball)
	default:
		http.NotFound(w, r)
	}
}

// packTarball builds a gzipped tar from name → content.
func packTarball(t *testing.T, files map[string]string) []byte {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	tarWriter := tar.NewWriter(gzipWriter)
	for name, content := range files {
		header := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(content)), Typeflag: tar.TypeReg}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzipWriter.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

// manifestJSON encodes a manifest.
func manifestJSON(t *testing.T, manifest Manifest) string {
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// writeFolder writes a plugin folder (package.json + plugin.js) under base.
func writeFolder(t *testing.T, base, folder string, manifest Manifest) {
	target := filepath.Join(base, folder)
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "package.json"), []byte(manifestJSON(t, manifest)), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "plugin.js"), []byte("module.exports = "+manifest.Version), 0o644); err != nil {
		t.Fatal(err)
	}
}

// installedVersion reads the version installed in folder, "" when absent.
func installedVersion(dir, folder string) string {
	manifest, err := readManifest(filepath.Join(dir, folder))
	if err != nil {
		return ""
	}
	return manifest.Version
}

func TestFolderName(t *testing.T) {
	cases := map[string]string{
		"@playingwithclouds/veil-plugin-eporner": "eporner",
		"veil-plugin-txxx-network":          "txxx-network",
		"@someone/sitepack":                  "sitepack",
	}
	for packageName, want := range cases {
		if got := FolderName(packageName); got != want {
			t.Errorf("FolderName(%q) = %q, want %q", packageName, got, want)
		}
	}
}

func TestInstallAndUpdate(t *testing.T) {
	registry := newFakeRegistry(t, "1.0.0")
	dir := t.TempDir()
	store := New(dir, registry.server.URL)
	ctx := context.Background()

	if _, err := store.Install(ctx, testPackage); err != nil {
		t.Fatalf("install: %v", err)
	}
	main, err := os.ReadFile(filepath.Join(dir, "example", "plugin.js"))
	if err != nil || string(main) != "module.exports = 1.0.0" {
		t.Fatalf("plugin.js = %q, %v", main, err)
	}

	updated, err := store.Update(ctx)
	if err != nil || len(updated) != 0 {
		t.Fatalf("update without new version: %v, %v", updated, err)
	}

	registry.publish(t, "1.1.0")
	updated, err = store.Update(ctx)
	if err != nil || len(updated) != 1 {
		t.Fatalf("update: %v, %v", updated, err)
	}
	if version := installedVersion(dir, "example"); version != "1.1.0" {
		t.Fatalf("installed %s after update, want 1.1.0", version)
	}
}

func TestUpdateSkipsUnmanagedFolders(t *testing.T) {
	cases := map[string]Manifest{
		"local build":   {Name: testPackage, Version: "1.0.0", Main: "plugin.js", Keywords: []string{Keyword}, LocalBuild: true},
		"plugin source": {Name: testPackage, Version: "1.0.0", Main: "index.ts"},
	}
	for label, manifest := range cases {
		t.Run(label, func(t *testing.T) {
			registry := newFakeRegistry(t, "2.0.0")
			dir := t.TempDir()
			writeFolder(t, dir, "example", manifest)

			updated, err := New(dir, registry.server.URL).Update(context.Background())
			if err != nil || len(updated) != 0 {
				t.Fatalf("update: %v, %v", updated, err)
			}
			if version := installedVersion(dir, "example"); version != "1.0.0" {
				t.Fatalf("replaced by %s", version)
			}
		})
	}
}

func TestInstallRejectsTamperedTarball(t *testing.T) {
	registry := newFakeRegistry(t, "1.0.0")
	registry.integrity = "sha512-" + base64.StdEncoding.EncodeToString(make([]byte, sha512.Size))
	dir := t.TempDir()

	_, err := New(dir, registry.server.URL).Install(context.Background(), testPackage)
	if err == nil || !strings.Contains(err.Error(), "integrity") {
		t.Fatalf("install of tampered tarball: %v", err)
	}
	if installedVersion(dir, "example") != "" {
		t.Fatal("tampered plugin was written")
	}
}

func TestInstallRefusesForeignFolder(t *testing.T) {
	registry := newFakeRegistry(t, "1.0.0")
	dir := t.TempDir()
	writeFolder(t, dir, "example", Manifest{Name: "@other/veil-plugin-example", Version: "1.0.0", Main: "plugin.js"})

	if _, err := New(dir, registry.server.URL).Install(context.Background(), testPackage); err == nil {
		t.Fatal("install overwrote another package's folder")
	}
}

func TestSeed(t *testing.T) {
	seedDir := t.TempDir()
	dir := t.TempDir()
	store := New(dir, "")
	writeFolder(t, seedDir, "example", Manifest{Name: testPackage, Version: "1.0.0", Main: "plugin.js"})

	if err := store.Seed(seedDir); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if version := installedVersion(dir, "example"); version != "1.0.0" {
		t.Fatalf("seeded %q, want 1.0.0", version)
	}

	// An uninstalled plugin stays gone until a newer one ships.
	if err := store.Uninstall("example"); err != nil {
		t.Fatal(err)
	}
	if err := store.Seed(seedDir); err != nil {
		t.Fatalf("reseed: %v", err)
	}
	if version := installedVersion(dir, "example"); version != "" {
		t.Fatalf("uninstalled plugin seeded again (%s)", version)
	}

	writeFolder(t, seedDir, "example", Manifest{Name: testPackage, Version: "1.1.0", Main: "plugin.js"})
	if err := store.Seed(seedDir); err != nil {
		t.Fatalf("seed newer: %v", err)
	}
	if version := installedVersion(dir, "example"); version != "1.1.0" {
		t.Fatalf("seeded %q, want 1.1.0", version)
	}
}

func TestSeedKeepsNewerInstall(t *testing.T) {
	seedDir := t.TempDir()
	dir := t.TempDir()
	writeFolder(t, seedDir, "example", Manifest{Name: testPackage, Version: "1.0.0", Main: "plugin.js"})
	writeFolder(t, dir, "example", Manifest{Name: testPackage, Version: "1.2.0", Main: "plugin.js"})

	if err := New(dir, "").Seed(seedDir); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if version := installedVersion(dir, "example"); version != "1.2.0" {
		t.Fatalf("seed downgraded to %s", version)
	}
}
