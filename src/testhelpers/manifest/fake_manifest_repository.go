package manifest

import (
	"cf/manifest"
)

type FakeManifestRepository struct {
	ReadManifestPath string
	ReadManifestErrors manifest.ManifestErrors
	ReadManifestManifest *manifest.Manifest

	UserSpecifiedPath string
	ManifestDir string
	ManifestFilename string
	ManifestPathErr error
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

func (repo *FakeManifestRepository) ManifestPath(userSpecifiedPath string) (manifestDir, manifestFilename string, err error){
	repo.UserSpecifiedPath = userSpecifiedPath


	if repo.ManifestDir != "" {
		manifestDir = repo.ManifestDir
	} else {
		manifestDir = userSpecifiedPath
	}

	if repo.ManifestFilename != "" {
		manifestFilename = repo.ManifestFilename
	} else {
		manifestFilename = "manifest.yml"
	}

	err = repo.ManifestPathErr
	return
}
