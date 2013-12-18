package manifest

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadManifestWithName(t *testing.T) {
	repo := NewManifestDiskRepository()
	manifest, err := repo.ReadManifest("../../fixtures/example-app")

	assert.NoError(t, err)
	assert.Equal(t, len(manifest.Applications), 1)
	assert.Equal(t, manifest.Applications[0].Get("name").(string), "hello")
}

func TestReadManifestWithBadPath(t *testing.T) {
	repo := NewManifestDiskRepository()
	_, err := repo.ReadManifest("some/path/that/doesnt/exist")

	assert.Error(t, err)
}
