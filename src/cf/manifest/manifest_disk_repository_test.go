package manifest

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
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
