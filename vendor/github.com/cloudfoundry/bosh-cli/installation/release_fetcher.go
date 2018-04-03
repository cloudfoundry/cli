package installation

import (
	"fmt"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"

	"github.com/cloudfoundry/bosh-cli/installation/tarball"
	boshrel "github.com/cloudfoundry/bosh-cli/release"
	"github.com/cloudfoundry/bosh-cli/release/manifest"
	"github.com/cloudfoundry/bosh-cli/ui"
)

type ReleaseFetcher struct {
	tarballProvider tarball.Provider
	releaseReader   boshrel.Reader
	releaseManager  ReleaseManager
}

func NewReleaseFetcher(
	tarballProvider tarball.Provider,
	releaseReader boshrel.Reader,
	releaseManager ReleaseManager,
) ReleaseFetcher {
	return ReleaseFetcher{
		tarballProvider: tarballProvider,
		releaseReader:   releaseReader,
		releaseManager:  releaseManager,
	}
}

func (f ReleaseFetcher) DownloadAndExtract(releaseRef manifest.ReleaseRef, stage ui.Stage) error {
	releasePath, err := f.tarballProvider.Get(releaseRef, stage)
	if err != nil {
		return err
	}

	err = stage.Perform(fmt.Sprintf("Validating release '%s'", releaseRef.Name), func() error {
		release, err := f.releaseReader.Read(releasePath)
		if err != nil {
			return bosherr.WrapErrorf(err, "Extracting release '%s'", releasePath)
		}

		if release.Name() != releaseRef.Name {
			errMsg := "Release name '%s' does not match the name in release tarball '%s'"
			return bosherr.Errorf(errMsg, releaseRef.Name, release.Name())
		}

		f.releaseManager.Add(release)

		return nil
	})

	return err
}
