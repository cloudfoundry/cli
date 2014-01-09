package manifest

import (
	"cf/manifest"
)

type FakeManifestRepository struct {
	ReadManifestDir string
	ReadManifestErrors manifest.ManifestErrors
	ReadManifestManifest *manifest.Manifest

	ManifestNotExists bool
}

func (repo *FakeManifestRepository) ReadManifest(dir string) (m *manifest.Manifest, errs manifest.ManifestErrors) {
	repo.ReadManifestDir = dir
	errs = repo.ReadManifestErrors

	if repo.ReadManifestManifest != nil {
		m = repo.ReadManifestManifest
	} else {
		m = manifest.NewEmptyManifest()
	}
	return
}

func (repo *FakeManifestRepository) ManifestExists(dir string) bool {
	return !repo.ManifestNotExists
}

