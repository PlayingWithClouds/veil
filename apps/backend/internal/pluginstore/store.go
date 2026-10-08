// Package pluginstore installs plugin bundles published on GitHub into the
// installed-plugins folder and keeps them up to date, so a plugin fix reaches
// every server and phone without a new backend or app release.
//
// Plugins are assets of the repository's rolling "plugins" release: an
// index.json naming every plugin's latest version, and per version an
// npm-style tarball holding one bundled CommonJS file and a package.json with
// the "veil-plugin" keyword (see scripts/publish-plugins.ts). A plugin lands
// in <plugins>/<folder>/, where the plugins registry picks it up through its
// file watcher.
package pluginstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/mod/semver"
)

// DefaultIndex is the plugin index on the repository's "plugins" release.
// Release downloads aren't subject to the GitHub API rate limit.
const DefaultIndex = "https://github.com/PlayingWithClouds/veil/releases/download/plugins/index.json"

// Keyword marks a package as a Veil plugin.
const Keyword = "veil-plugin"

// packagePrefix is stripped from package names to name their folder.
const packagePrefix = "veil-plugin-"

// stateFile records what the store has done in the plugin folder. Hidden, so
// the registry's watcher ignores it.
const stateFile = ".store.json"

// DefaultPackages are installed on a fresh backend that has no plugins yet.
var DefaultPackages = []string{
	"@playingwithclouds/veil-plugin-aylo",
	"@playingwithclouds/veil-plugin-eporner",
	"@playingwithclouds/veil-plugin-hqporner",
	"@playingwithclouds/veil-plugin-kvs",
	"@playingwithclouds/veil-plugin-missav",
	"@playingwithclouds/veil-plugin-pornpics",
	"@playingwithclouds/veil-plugin-spankbang",
	"@playingwithclouds/veil-plugin-tnaflix",
	"@playingwithclouds/veil-plugin-txxx-network",
	"@playingwithclouds/veil-plugin-xhamster",
}

// Manifest is the package.json of an installed plugin.
type Manifest struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Main        string   `json:"main"`
	Keywords    []string `json:"keywords,omitempty"`
	// LocalBuild marks bundles built from this checkout by the dev bundler;
	// the store never replaces them.
	LocalBuild bool `json:"localBuild,omitempty"`
}

// Installed is a plugin folder and its manifest.
type Installed struct {
	Folder   string
	Manifest Manifest
}

// state is the content of stateFile.
type state struct {
	// Seeded maps package name → the newest version copied in from a seed dir,
	// so a plugin the user uninstalled isn't seeded again until a newer one ships.
	Seeded map[string]string `json:"seeded"`
	// DefaultsInstalled is set once DefaultPackages were installed.
	DefaultsInstalled bool `json:"defaultsInstalled"`
}

// Store installs, updates and removes plugins in one plugin folder.
type Store struct {
	dir      string
	indexURL string
	client   *http.Client
	// mu serializes changes to the plugin folder.
	mu sync.Mutex
}

// New returns a store for the plugin folder dir, installing what the plugin
// index at indexURL lists (DefaultIndex when empty).
func New(dir, indexURL string) *Store {
	if indexURL == "" {
		indexURL = DefaultIndex
	}
	return &Store{
		dir:      dir,
		indexURL: indexURL,
		client:   &http.Client{Timeout: 2 * time.Minute},
	}
}

// FolderName is the plugin folder for a package:
// "@playingwithclouds/veil-plugin-eporner" → "eporner".
func FolderName(packageName string) string {
	name := packageName
	if index := strings.LastIndex(name, "/"); index != -1 {
		name = name[index+1:]
	}
	return strings.TrimPrefix(name, packagePrefix)
}

// Installed lists every plugin folder that holds a manifest.
func (s *Store) Installed() ([]Installed, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}
	var installed []Installed
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		manifest, err := readManifest(filepath.Join(s.dir, entry.Name()))
		if err != nil {
			continue
		}
		installed = append(installed, Installed{Folder: entry.Name(), Manifest: *manifest})
	}
	return installed, nil
}

// Uninstall removes the plugin in folder; the registry unloads it.
func (s *Store) Uninstall(folder string) error {
	if !validFolder(folder) {
		return fmt.Errorf("invalid plugin folder %q", folder)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return os.RemoveAll(filepath.Join(s.dir, folder))
}

// validFolder reports whether folder is a plain, visible directory name.
func validFolder(folder string) bool {
	return folder != "" && folder != "." && folder != ".." &&
		!strings.HasPrefix(folder, ".") && !strings.ContainsAny(folder, `/\`)
}

// readManifest reads a plugin folder's package.json.
func readManifest(folder string) (*Manifest, error) {
	data, err := os.ReadFile(filepath.Join(folder, "package.json"))
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parse package.json: %w", err)
	}
	return &manifest, nil
}

// checkReplace errors when folder already holds a different package, which
// installing packageName would overwrite.
func (s *Store) checkReplace(folder, packageName string) error {
	existing, err := readManifest(filepath.Join(s.dir, folder))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if existing.Name != packageName {
		return fmt.Errorf("plugin folder %q already holds %s", folder, existing.Name)
	}
	return nil
}

// writePlugin writes a plugin's entry file and manifest into its folder, each
// via rename so the registry never loads half a file. The manifest goes last:
// it is what makes the registry (re)load the plugin.
func (s *Store) writePlugin(folder string, manifest []byte, mainName string, main []byte) error {
	target := filepath.Join(s.dir, folder)
	if err := os.MkdirAll(target, 0o755); err != nil {
		return err
	}
	if err := writeAtomically(filepath.Join(target, mainName), main); err != nil {
		return err
	}
	return writeAtomically(filepath.Join(target, "package.json"), manifest)
}

// writeAtomically writes data to path through a hidden temporary file.
func writeAtomically(path string, data []byte) error {
	temporary := filepath.Join(filepath.Dir(path), "."+filepath.Base(path)+".tmp")
	if err := os.WriteFile(temporary, data, 0o644); err != nil {
		return err
	}
	return os.Rename(temporary, path)
}

// loadState reads stateFile; a missing file is an empty state.
func (s *Store) loadState() (*state, error) {
	current := &state{Seeded: map[string]string{}}
	data, err := os.ReadFile(filepath.Join(s.dir, stateFile))
	if errors.Is(err, os.ErrNotExist) {
		return current, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, current); err != nil {
		return nil, fmt.Errorf("parse %s: %w", stateFile, err)
	}
	if current.Seeded == nil {
		current.Seeded = map[string]string{}
	}
	return current, nil
}

// saveState writes stateFile.
func (s *Store) saveState(current *state) error {
	data, err := json.MarshalIndent(current, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomically(filepath.Join(s.dir, stateFile), data)
}

// newer reports whether version a is a higher semver than b. An unparsable
// b counts as older than anything valid.
func newer(a, b string) bool {
	return semver.Compare("v"+a, "v"+b) > 0
}

// validEntryName reports whether main names a file directly in the package
// root, the only layout the store installs.
func validEntryName(main string) bool {
	return validFolder(main)
}
