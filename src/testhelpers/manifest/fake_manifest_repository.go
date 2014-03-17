package manifest

import (
	"cf/manifest"
)

type FakeManifestRepository struct {
	ReadManifestArgs struct {
		Path string
	}
	ReadManifestReturns struct {
		Manifest *manifest.Manifest
		Errors   manifest.ManifestErrors
	}
}

func (repo *FakeManifestRepository) ReadManifest(inputPath string) (m *manifest.Manifest, errs manifest.ManifestErrors) {
	repo.ReadManifestArgs.Path = inputPath
	if repo.ReadManifestReturns.Manifest != nil {
		m = repo.ReadManifestReturns.Manifest
	} else {
		m = manifest.NewEmptyManifest()
	}
	errs = repo.ReadManifestReturns.Errors
	return
}
