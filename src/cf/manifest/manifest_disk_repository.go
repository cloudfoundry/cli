package manifest

import (
	"os"
	"path/filepath"
)

type ManifestRepository interface {
	ReadManifest(dir string) (manifest *Manifest, errs ManifestErrors)
}

type ManifestDiskRepository struct {
}

func NewManifestDiskRepository() (repo ManifestRepository) {
	return ManifestDiskRepository{}
}

func (repo ManifestDiskRepository) ReadManifest(path string) (m *Manifest, errs ManifestErrors) {
	m = NewEmptyManifest()

	if os.Getenv("CF_MANIFEST") != "true" {
		return
	}

	fileInfo, err := os.Stat(path)

	if err != nil {
		return
	}

	var fullPath string
	if fileInfo.IsDir() {
		fullPath = filepath.Join(path, "manifest.yml")
	} else {
		fullPath = path
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return
	}

	m, errs = Parse(file)
	return
}
