package manifest_test

import (
	. "cf/manifest"
	"generic"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func assertFeatureFlag(t *testing.T) {
	if os.Getenv("CF_MANIFEST") != "true" {
		t.Fatal("CF_MANIFEST must be set to 'true' to run manifest tests")
	}
}

func TestReadManifestWithGoodPath(t *testing.T) {
	assertFeatureFlag(t)
	repo := NewManifestDiskRepository()
	manifest, errs := repo.ReadManifest("../../fixtures/different-manifest.yml")

	assert.True(t, errs.Empty())
	assert.Equal(t, len(manifest.Applications), 1)
	assert.Equal(t, manifest.Applications[0].Get("name").(string), "goodbyte")
}

func TestReadManifestWithBadPath(t *testing.T) {
	assertFeatureFlag(t)
	repo := NewManifestDiskRepository()
	_, errs := repo.ReadManifest("some/path/that/doesnt/exist/manifest.yml")

	assert.False(t, errs.Empty())
}

func TestManifestPathsDefaultsToCurrentDirectory(t *testing.T) {
	assertFeatureFlag(t)
	repo := NewManifestDiskRepository()

	cwd, err := os.Getwd()
	assert.NoError(t, err)

	path, filename, err := repo.ManifestPath("")

	assert.NoError(t, err)
	assert.Equal(t, path, cwd)
	assert.Equal(t, filename, "manifest.yml")
}

func TestAppAndManifestPathsIgnoreAppPathWhenManifestPathIsSpecified(t *testing.T) {
	assertFeatureFlag(t)
	repo := NewManifestDiskRepository()

	cwd, err := os.Getwd()
	assert.NoError(t, err)
	expectedDir := filepath.Join(cwd, "..")

	path, filename, err := repo.ManifestPath(expectedDir)

	assert.NoError(t, err)
	assert.Equal(t, path, expectedDir)
	assert.Equal(t, filename, "manifest.yml")
}

func TestAppAndManifestPathsManifestFileIsDroppedFromAppPath(t *testing.T) {
	assertFeatureFlag(t)
	repo := NewManifestDiskRepository()

	cwd, err := os.Getwd()
	assert.NoError(t, err)

	path, filename, err := repo.ManifestPath(filepath.Join(cwd, "example_manifest.yml"))

	assert.NoError(t, err)
	assert.Equal(t, path, cwd)
	assert.Equal(t, filename, "example_manifest.yml")
}

func TestManifestWithInheritance(t *testing.T) {
	assertFeatureFlag(t)
	repo := NewManifestDiskRepository()
	m, err := repo.ReadManifest("../../fixtures/inherited-manifest.yml")
	assert.NoError(t, err)
	assert.Equal(t, m.Applications[0].Get("name"), "base-app")
	assert.Equal(t, m.Applications[0].Get("services"), []string{"base-service"})
	assert.Equal(t, m.Applications[0].Get("env"), generic.NewMap(map[string]string{
		"foo":                "bar",
		"will-be-overridden": "my-value",
	}))

	assert.Equal(t, m.Applications[1].Get("name"), "my-app")

	env := generic.NewMap(m.Applications[1].Get("env"))
	assert.Equal(t, env.Get("will-be-overridden"), "my-value")
	assert.Equal(t, env.Get("foo"), "bar")

	services := m.Applications[1].Get("services")
	assert.Equal(t, services, []string{"base-service", "foo-service"})
}

func TestPushingWithAbsoluteAppPathFromManifestFile(t *testing.T) {
	assertFeatureFlag(t)

	if runtime.GOOS == "windows" {
		pushingWithAbsoluteWindowsPath(t)
	} else {
		pushingWithAbsoluteUnixPath(t)
	}
}

func pushingWithAbsoluteUnixPath(t *testing.T) {
	repo := NewManifestDiskRepository()
	m, err := repo.ReadManifest("../../fixtures/unix-manifest.yml")

	assert.NoError(t, err)
	assert.Equal(t, m.Applications[0].Get("path"), "/absolute/path/to/example-app")
}

func pushingWithAbsoluteWindowsPath(t *testing.T) {
	repo := NewManifestDiskRepository()
	m, err := repo.ReadManifest("../../fixtures/windows-manifest.yml")

	assert.NoError(t, err)
	assert.Equal(t, m.Applications[0].Get("path"), "C:\\path\\to\\my\\app")
}
