package manifest

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestReadManifestWithName(t *testing.T) {
	if os.Getenv("CF_MANIFEST") != "true" {
		t.Fatal("CF_MANIFEST must be set to 'true' to run manifest tests")
	}
	repo := NewManifestDiskRepository()
	manifest, err := repo.ReadManifest("../../fixtures/example-app")

	assert.NoError(t, err)
	assert.Equal(t, len(manifest.Applications), 1)
	assert.Equal(t, manifest.Applications[0].Get("name").(string), "hello")
}

func TestReadManifestWithBadPath(t *testing.T) {
	if os.Getenv("CF_MANIFEST") != "true" {
		t.Fatal("CF_MANIFEST must be set to 'true' to run manifest tests")
	}
	repo := NewManifestDiskRepository()
	_, err := repo.ReadManifest("some/path/that/doesnt/exist")

	assert.Error(t, err)
}
