package manifest

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func assertFeatureFlag(t *testing.T) {
	if os.Getenv("CF_MANIFEST") != "true" {
		t.Fatal("CF_MANIFEST must be set to 'true' to run manifest tests")
	}
}

func TestReadManifestWithDirectoryName(t *testing.T) {
	assertFeatureFlag(t)
	repo := NewManifestDiskRepository()
	manifest, err := repo.ReadManifest("../../fixtures/example-app")

	assert.NoError(t, err)
	assert.Equal(t, len(manifest.Applications), 1)
	assert.Equal(t, manifest.Applications[0].Get("name").(string), "hello")
}

func TestReadDifferentManifestFilename(t *testing.T) {
	assertFeatureFlag(t)
	repo := NewManifestDiskRepository()
	manifest, err := repo.ReadManifest("../../fixtures/different-manifest.yml")

	assert.NoError(t, err)
	assert.Equal(t, len(manifest.Applications), 1)
	assert.Equal(t, manifest.Applications[0].Get("name").(string), "goodbyte")
}

func TestReadManifestWithBadPath(t *testing.T) {
	assertFeatureFlag(t)
	repo := NewManifestDiskRepository()
	m, err := repo.ReadManifest("some/path/that/doesnt/exist")

	assert.NoError(t, err)
	assert.Equal(t, m, NewEmptyManifest())
}

func TestReadManifestWithValidPathAndNoManifest(t *testing.T) {
	assertFeatureFlag(t)
	repo := NewManifestDiskRepository()
	m, err := repo.ReadManifest(".")

	assert.NoError(t, err)
	assert.Equal(t, m, NewEmptyManifest())
}
