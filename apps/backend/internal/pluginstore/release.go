package pluginstore

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"slices"
	"strings"
)

// maxTarballBytes bounds a downloaded package; a plugin bundle is well below.
const maxTarballBytes = 32 << 20

// Package is a plugin published in the index.
type Package struct {
	Name        string
	Version     string
	Description string
}

// versionManifest is the index's entry for a package's latest version.
type versionManifest struct {
	Manifest
	Dist struct {
		Tarball   string `json:"tarball"`
		Integrity string `json:"integrity"`
	} `json:"dist"`
}

// index is the published index.json: the latest version of every plugin.
type index struct {
	Packages []versionManifest `json:"packages"`
}

// fetchIndex downloads the plugin index.
func (s *Store) fetchIndex(ctx context.Context) (*index, error) {
	var published index
	if err := s.getJSON(ctx, s.indexURL, &published); err != nil {
		return nil, fmt.Errorf("plugin index: %w", err)
	}
	return &published, nil
}

// latest finds a package's latest published version in the index.
func (published *index) latest(packageName string) (*versionManifest, error) {
	for index := range published.Packages {
		manifest := &published.Packages[index]
		if manifest.Name != packageName {
			continue
		}
		if !slices.Contains(manifest.Keywords, Keyword) {
			return nil, fmt.Errorf("%s is not a Veil plugin (no %q keyword)", packageName, Keyword)
		}
		if !validEntryName(manifest.Main) {
			return nil, fmt.Errorf("%s: unsupported main %q", packageName, manifest.Main)
		}
		return manifest, nil
	}
	return nil, fmt.Errorf("%s is not in the plugin index", packageName)
}

// Search lists published plugins whose name or description contains text
// (all when empty).
func (s *Store) Search(ctx context.Context, text string) ([]Package, error) {
	published, err := s.fetchIndex(ctx)
	if err != nil {
		return nil, err
	}
	needle := strings.ToLower(strings.TrimSpace(text))
	packages := []Package{}
	for _, manifest := range published.Packages {
		haystack := strings.ToLower(manifest.Name + " " + manifest.Description)
		if !strings.Contains(haystack, needle) {
			continue
		}
		packages = append(packages, Package{
			Name:        manifest.Name,
			Version:     manifest.Version,
			Description: manifest.Description,
		})
	}
	return packages, nil
}

// Install installs the latest version of packageName, replacing an older
// install of the same package.
func (s *Store) Install(ctx context.Context, packageName string) (*Manifest, error) {
	published, err := s.fetchIndex(ctx)
	if err != nil {
		return nil, err
	}
	manifest, err := published.latest(packageName)
	if err != nil {
		return nil, err
	}
	if err := s.installVersion(ctx, manifest); err != nil {
		return nil, err
	}
	return &manifest.Manifest, nil
}

// installVersion downloads, verifies and writes one published version.
func (s *Store) installVersion(ctx context.Context, manifest *versionManifest) error {
	folder := FolderName(manifest.Name)
	if !validFolder(folder) {
		return fmt.Errorf("%s: unusable package name", manifest.Name)
	}
	tarball, err := s.download(ctx, manifest.Dist.Tarball)
	if err != nil {
		return fmt.Errorf("download %s: %w", manifest.Name, err)
	}
	if err := verifyIntegrity(tarball, manifest.Dist.Integrity); err != nil {
		return fmt.Errorf("%s@%s: %w", manifest.Name, manifest.Version, err)
	}
	files, err := extract(tarball, "package.json", manifest.Main)
	if err != nil {
		return fmt.Errorf("%s@%s: %w", manifest.Name, manifest.Version, err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.checkReplace(folder, manifest.Name); err != nil {
		return err
	}
	return s.writePlugin(folder, files["package.json"], manifest.Main, files[manifest.Main])
}

// Update installs newer published versions of every installed plugin package
// except local builds, returning the package names it updated. Folders
// without the plugin keyword (e.g. plugin sources) are never touched. One
// failing package doesn't stop the others; their errors are joined.
func (s *Store) Update(ctx context.Context) ([]string, error) {
	installed, err := s.Installed()
	if err != nil {
		return nil, err
	}
	published, err := s.fetchIndex(ctx)
	if err != nil {
		return nil, err
	}
	var updated []string
	var errs []error
	for _, plugin := range installed {
		if plugin.Manifest.LocalBuild || !slices.Contains(plugin.Manifest.Keywords, Keyword) {
			continue
		}
		manifest, err := published.latest(plugin.Manifest.Name)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if !newer(manifest.Version, plugin.Manifest.Version) {
			continue
		}
		if err := s.installVersion(ctx, manifest); err != nil {
			errs = append(errs, err)
			continue
		}
		updated = append(updated, plugin.Manifest.Name)
	}
	return updated, errors.Join(errs...)
}

// getJSON fetches url and decodes its JSON body into target.
func (s *Store) getJSON(ctx context.Context, url string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("server answered %s", response.Status)
	}
	return json.NewDecoder(response.Body).Decode(target)
}

// download fetches a package tarball.
func (s *Store) download(ctx context.Context, url string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server answered %s", response.Status)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxTarballBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxTarballBytes {
		return nil, fmt.Errorf("tarball larger than %d bytes", maxTarballBytes)
	}
	return data, nil
}

// verifyIntegrity checks data against a "sha512-<base64>" (npm-style) integrity string.
func verifyIntegrity(data []byte, integrity string) error {
	encoded, ok := strings.CutPrefix(integrity, "sha512-")
	if !ok {
		return fmt.Errorf("unsupported integrity %q", integrity)
	}
	expected, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("bad integrity: %w", err)
	}
	actual := sha512.Sum512(data)
	if !bytes.Equal(actual[:], expected) {
		return errors.New("tarball does not match its integrity hash")
	}
	return nil
}

// extract reads the named files from the package root of a gzipped package
// tarball (entries under "package/"); every one must be present.
func extract(tarball []byte, names ...string) (map[string][]byte, error) {
	gzipReader, err := gzip.NewReader(bytes.NewReader(tarball))
	if err != nil {
		return nil, err
	}
	reader := tar.NewReader(gzipReader)
	files := map[string][]byte{}
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		name := packageRelative(header.Name)
		if header.Typeflag != tar.TypeReg || !slices.Contains(names, name) {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(reader, maxTarballBytes))
		if err != nil {
			return nil, err
		}
		files[name] = data
	}
	for _, name := range names {
		if _, ok := files[name]; !ok {
			return nil, fmt.Errorf("package has no %s", name)
		}
	}
	return files, nil
}

// packageRelative strips the tarball's top folder ("package/" by convention,
// though npm accepts any single top folder) from an entry name.
func packageRelative(entryName string) string {
	cleaned := path.Clean(entryName)
	_, rest, found := strings.Cut(cleaned, "/")
	if !found {
		return ""
	}
	return rest
}
