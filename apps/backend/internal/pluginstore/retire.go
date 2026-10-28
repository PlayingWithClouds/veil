package pluginstore

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
)

// RetiredPackages maps packages that were folded into another plugin to the
// package that took their place. ukdevilz and mat6tube are copies of
// noodlemagazine, whose plugin now fetches from all three.
var RetiredPackages = map[string]string{
	"@playingwithclouds/veil-plugin-ukdevilz": "@playingwithclouds/veil-plugin-noodlemagazine",
	"@playingwithclouds/veil-plugin-mat6tube": "@playingwithclouds/veil-plugin-noodlemagazine",
}

// retire replaces installed retired packages with their successors: the
// successor is installed when no folder holds it yet, and the retired
// package's folder is removed once it does. Local builds and plugin sources
// are left alone (the dev bundler drops stale local builds itself). Returns
// the retired package names it removed.
func (s *Store) retire(ctx context.Context, installed []Installed, published *index) ([]string, error) {
	var removed []string
	var errs []error
	for _, plugin := range installed {
		successor, retired := RetiredPackages[plugin.Manifest.Name]
		if !retired || !managed(plugin) {
			continue
		}
		if err := s.ensureInstalled(ctx, installed, successor, published); err != nil {
			errs = append(errs, fmt.Errorf("retire %s: %w", plugin.Manifest.Name, err))
			continue
		}
		if err := s.Uninstall(plugin.Folder); err != nil {
			errs = append(errs, fmt.Errorf("retire %s: %w", plugin.Manifest.Name, err))
			continue
		}
		removed = append(removed, plugin.Manifest.Name)
	}
	return removed, errors.Join(errs...)
}

// managed reports whether the store installed the plugin and may replace it:
// a published package, not a local build or a plugin's source folder.
func managed(plugin Installed) bool {
	return !plugin.Manifest.LocalBuild && slices.Contains(plugin.Manifest.Keywords, Keyword)
}

// ensureInstalled installs packageName's latest version unless a folder
// (installed, or written since by an earlier call) already holds it.
func (s *Store) ensureInstalled(ctx context.Context, installed []Installed, packageName string, published *index) error {
	for _, plugin := range installed {
		if plugin.Manifest.Name == packageName {
			return nil
		}
	}
	if current, err := readManifest(filepath.Join(s.dir, FolderName(packageName))); err == nil && current.Name == packageName {
		return nil
	}
	manifest, err := published.latest(packageName)
	if err != nil {
		return err
	}
	return s.installVersion(ctx, manifest)
}
