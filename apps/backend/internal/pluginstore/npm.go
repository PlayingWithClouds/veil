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
	"net/url"
	"path"
	"slices"
	"strings"
)

// maxTarballBytes bounds a downloaded package; a plugin bundle is well below.
const maxTarballBytes = 32 << 20

// Package is a plugin published on the registry.
type Package struct {
	Name        string
	Version     string
	Description string
}

// versionManifest is the registry's document for one published version.
type versionManifest struct {
	Manifest
	Dist struct {
		Tarball   string `json:"tarball"`
		Integrity string `json:"integrity"`
	} `json:"dist"`
}

// latest fetches the manifest of a package's latest published version.
func (s *Store) latest(ctx context.Context, packageName string) (*versionManifest, error) {
	escaped := strings.Replace(packageName, "/", "%2f", 1)
	var manifest versionManifest
	if err := s.getJSON(ctx, s.registryURL+"/"+escaped+"/latest", &manifest); err != nil {
		return nil, fmt.Errorf("%s: %w", packageName, err)
	}
	if !slices.Contains(manifest.Keywords, Keyword) {
		return nil, fmt.Errorf("%s is not an Veil plugin (no %q keyword)", packageName, Keyword)
	}
	if !validEntryName(manifest.Main) {
		return nil, fmt.Errorf("%s: unsupported main %q", packageName, manifest.Main)
	}
	return &manifest, nil
}

// Search lists plugins on the registry matching text (all when empty).
func (s *Store) Search(ctx context.Context, text string) ([]Package, error) {
	query := url.Values{}
	query.Set("text", strings.TrimSpace("keywords:"+Keyword+" "+text))
	query.Set("size", "250")
	var response struct {
		Objects []struct {
			Package struct {
				Name        string `json:"name"`
				Version     string `json:"version"`
				Description string `json:"description"`
			} `json:"package"`
		} `json:"objects"`
	}
	if err := s.getJSON(ctx, s.registryURL+"/-/v1/search?"+query.Encode(), &response); err != nil {
		return nil, fmt.Errorf("search plugins: %w", err)
	}
	packages := make([]Package, 0, len(response.Objects))
	for _, object := range response.Objects {
		packages = append(packages, Package{
			Name:        object.Package.Name,
			Version:     object.Package.Version,
			Description: object.Package.Description,
		})
	}
	return packages, nil
}

// Install installs the latest version of packageName, replacing an older
// install of the same package.
func (s *Store) Install(ctx context.Context, packageName string) (*Manifest, error) {
	manifest, err := s.latest(ctx, packageName)
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
	var updated []string
	var errs []error
	for _, plugin := range installed {
		if plugin.Manifest.LocalBuild || !slices.Contains(plugin.Manifest.Keywords, Keyword) {
			continue
		}
		manifest, err := s.latest(ctx, plugin.Manifest.Name)
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
		return fmt.Errorf("registry answered %s", response.Status)
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
		return nil, fmt.Errorf("registry answered %s", response.Status)
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

// verifyIntegrity checks data against an npm "sha512-<base64>" integrity string.
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

// extract reads the named files from the package root of a gzipped npm
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
