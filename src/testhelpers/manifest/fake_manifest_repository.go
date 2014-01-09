package manifest

import (
	"cf/manifest"
)

type FakeManifestRepository struct {
	ReadManifestPath string
	ReadManifestErrors manifest.ManifestErrors
	ReadManifestManifest *manifest.Manifest
}

func (repo *FakeManifestRepository) ReadManifest(dir string) (m *manifest.Manifest, errs manifest.ManifestErrors) {
	repo.ReadManifestPath = dir
	errs = repo.ReadManifestErrors

	if repo.ReadManifestManifest != nil {
		m = repo.ReadManifestManifest
	} else {
		m = manifest.NewEmptyManifest()
	}
	return
}


