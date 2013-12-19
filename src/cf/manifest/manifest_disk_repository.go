package manifest

import (
	"os"
	"path/filepath"
)

type ManifestRepository interface {
	ReadManifest(dir string) (manifest *Manifest, err error)
}

type ManifestDiskRepository struct {
}

func NewManifestDiskRepository() (repo ManifestRepository) {
	return ManifestDiskRepository{}
}

func (repo ManifestDiskRepository) ReadManifest(dir string) (m *Manifest, err error) {
	if os.Getenv("CF_MANIFEST") != "true" {
		m = NewEmptyManifest()
		return
	}
	path := filepath.Join(dir, "manifest.yml")
	file, err := os.Open(path)
	if err != nil {
		return
	}

	return Parse(file)
}
