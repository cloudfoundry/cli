package manifest

import (
	"github.com/cloudfoundry/cli/cf/manifest"
)

type FakeManifestRepository struct {
	ReadManifestArgs struct {
		Path string
	}
	ReadManifestReturns struct {
		Manifest *manifest.Manifest
		Error    error
	}
}

func (repo *FakeManifestRepository) ReadManifest(inputPath string) (m *manifest.Manifest, err error) {
	repo.ReadManifestArgs.Path = inputPath
	if repo.ReadManifestReturns.Manifest != nil {
		m = repo.ReadManifestReturns.Manifest
	} else {
		m = manifest.NewEmptyManifest()
	}

	err = repo.ReadManifestReturns.Error
	return
}
