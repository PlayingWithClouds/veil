package pluginstore

import (
	"context"
	"testing"
)

const retiredTestPackage = "@playingwithclouds/veil-plugin-old"

// retireInto makes retiredTestPackage a retired package replaced by successor
// for the test's duration.
func retireInto(t *testing.T, successor string) {
	original := RetiredPackages
	RetiredPackages = map[string]string{retiredTestPackage: successor}
	t.Cleanup(func() { RetiredPackages = original })
}

func TestUpdateReplacesARetiredPackageWithItsSuccessor(t *testing.T) {
	retireInto(t, testPackage)
	registry := newFakeRegistry(t, "1.0.0")
	dir := t.TempDir()
	writeFolder(t, dir, "old", Manifest{Name: retiredTestPackage, Version: "0.3.0", Main: "plugin.js", Keywords: []string{Keyword}})

	removed, err := New(dir, registry.indexURL()).Update(context.Background())
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if len(removed) != 1 || removed[0] != retiredTestPackage {
		t.Errorf("update reported %v, want the retired package", removed)
	}
	if version := installedVersion(dir, "old"); version != "" {
		t.Errorf("retired package still installed at %s", version)
	}
	if version := installedVersion(dir, "example"); version != "1.0.0" {
		t.Errorf("successor installed at %q, want 1.0.0", version)
	}
}

func TestUpdateKeepsARetiredPackageWhoseSuccessorIsUnavailable(t *testing.T) {
	retireInto(t, "@playingwithclouds/veil-plugin-unpublished")
	registry := newFakeRegistry(t, "1.0.0")
	dir := t.TempDir()
	writeFolder(t, dir, "old", Manifest{Name: retiredTestPackage, Version: "0.3.0", Main: "plugin.js", Keywords: []string{Keyword}})

	if _, err := New(dir, registry.indexURL()).Update(context.Background()); err == nil {
		t.Error("update must report the successor it could not install")
	}
	if version := installedVersion(dir, "old"); version != "0.3.0" {
		t.Errorf("retired package removed without a successor (installed %q)", version)
	}
}

func TestUpdateLeavesARetiredLocalBuildAlone(t *testing.T) {
	retireInto(t, testPackage)
	registry := newFakeRegistry(t, "1.0.0")
	dir := t.TempDir()
	writeFolder(t, dir, "old", Manifest{Name: retiredTestPackage, Version: "0.3.0", Main: "plugin.js", Keywords: []string{Keyword}, LocalBuild: true})

	if _, err := New(dir, registry.indexURL()).Update(context.Background()); err != nil {
		t.Fatalf("update: %v", err)
	}
	if version := installedVersion(dir, "old"); version != "0.3.0" {
		t.Errorf("local build replaced (installed %q)", version)
	}
	if version := installedVersion(dir, "example"); version != "" {
		t.Errorf("successor installed for a local build: %s", version)
	}
}
