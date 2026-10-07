package pluginstore

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Seed copies the plugins shipped in seedDir (the phone app's bundled set,
// laid out like the plugin folder) into the plugin folder. A shipped plugin
// is copied when it is newer than the version seeded before and than the one
// installed, so neither an npm update nor an uninstall is undone by a restart.
func (s *Store) Seed(seedDir string) error {
	entries, err := os.ReadDir(seedDir)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	current, err := s.loadState()
	if err != nil {
		return err
	}
	var errs []error
	for _, entry := range entries {
		if !entry.IsDir() || !validFolder(entry.Name()) {
			continue
		}
		if err := s.seedOne(filepath.Join(seedDir, entry.Name()), entry.Name(), current); err != nil {
			errs = append(errs, fmt.Errorf("seed %s: %w", entry.Name(), err))
		}
	}
	if err := s.saveState(current); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// seedOne copies one shipped plugin if it is new, recording it in current.
func (s *Store) seedOne(source, folder string, current *state) error {
	shipped, err := readManifest(source)
	if err != nil {
		return err
	}
	if !validEntryName(shipped.Main) {
		return fmt.Errorf("unsupported main %q", shipped.Main)
	}
	if seeded, ok := current.Seeded[shipped.Name]; ok && !newer(shipped.Version, seeded) {
		return nil
	}
	current.Seeded[shipped.Name] = shipped.Version

	installed, err := readManifest(filepath.Join(s.dir, folder))
	if err == nil && (installed.LocalBuild || !newer(shipped.Version, installed.Version)) {
		return nil
	}
	if err := s.checkReplace(folder, shipped.Name); err != nil {
		return err
	}
	manifest, err := os.ReadFile(filepath.Join(source, "package.json"))
	if err != nil {
		return err
	}
	main, err := os.ReadFile(filepath.Join(source, shipped.Main))
	if err != nil {
		return err
	}
	return s.writePlugin(folder, manifest, shipped.Main, main)
}

// InstallDefaults installs DefaultPackages once, on a backend that has never
// done so; plugins whose folder is already taken (seeded, or a local build)
// are left alone. Retried on the next run until every install succeeded.
func (s *Store) InstallDefaults(ctx context.Context) error {
	s.mu.Lock()
	current, err := s.loadState()
	s.mu.Unlock()
	if err != nil || current.DefaultsInstalled {
		return err
	}

	var errs []error
	for _, packageName := range DefaultPackages {
		if _, err := os.Stat(filepath.Join(s.dir, FolderName(packageName))); err == nil {
			continue
		}
		if _, err := s.Install(ctx, packageName); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	current, err = s.loadState()
	if err != nil {
		return err
	}
	current.DefaultsInstalled = true
	return s.saveState(current)
}

// Run installs the default plugins, then checks for plugin updates right away
// and every interval until ctx ends.
func (s *Store) Run(ctx context.Context, interval time.Duration) {
	if err := s.InstallDefaults(ctx); err != nil {
		log.Printf("plugin store: install defaults: %v", err)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		s.logUpdate(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// logUpdate runs one update check and logs its outcome.
func (s *Store) logUpdate(ctx context.Context) {
	updated, err := s.Update(ctx)
	for _, name := range updated {
		log.Printf("plugin store: updated %s", name)
	}
	if err != nil {
		log.Printf("plugin store: update: %v", err)
	}
}
