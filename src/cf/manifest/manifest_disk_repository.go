package manifest

import (
	"os"
	"path/filepath"
)

type ManifestRepository interface {
	ReadManifest(dir string) (manifest *Manifest, err error)
	ManifestExists(dir string) bool
}

type ManifestDiskRepository struct {
}

func NewManifestDiskRepository() (repo ManifestRepository) {
	return ManifestDiskRepository{}
}

func (repo ManifestDiskRepository) ReadManifest(dir string) (m *Manifest, err error) {
	file, err := os.Open(repo.filenameFromPath(dir))
	if err != nil {
		return
	}

	return Parse(file)
}

func (repo ManifestDiskRepository) ManifestExists(dir string) bool {
	if os.Getenv("CF_MANIFEST") != "true" {
		return false
	}

	_, err := os.Stat(repo.filenameFromPath(dir))
	return err == nil
}

func (repo ManifestDiskRepository) filenameFromPath(dir string) string {
	return filepath.Join(dir, "manifest.yml")
}
