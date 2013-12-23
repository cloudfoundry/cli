package manifest

import (
	"cf/manifest"
)

type FakeManifestRepository struct {
	ReadManifestDir string
	ReadManifestError error
	ReadManifestManifest *manifest.Manifest

	ManifestNotExists bool
}

func (repo *FakeManifestRepository) ReadManifest(dir string) (m *manifest.Manifest, err error) {
	repo.ReadManifestDir = dir
	err = repo.ReadManifestError
	if repo.ReadManifestManifest != nil {
		m = repo.ReadManifestManifest
	} else {
		m = manifest.NewEmptyManifest()
	}
	return
}

func (repo *FakeManifestRepository) ManifestExists(dir string) bool{
	return !repo.ManifestNotExists
}
