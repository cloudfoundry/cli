package deployment

import (
	biconfig "github.com/cloudfoundry/bosh-cli/config"
	birel "github.com/cloudfoundry/bosh-cli/release"
	bistemcell "github.com/cloudfoundry/bosh-cli/stemcell"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type Record interface {
	IsDeployed(manifestSHA string, releases []birel.Release, stemcell bistemcell.ExtractedStemcell) (bool, error)
	Clear() error
	Update(manifestSHA string, releases []birel.Release) error
}

type deploymentRecord struct {
	deploymentRepo biconfig.DeploymentRepo
	releaseRepo    biconfig.ReleaseRepo
	stemcellRepo   biconfig.StemcellRepo
}

func NewRecord(
	deploymentRepo biconfig.DeploymentRepo,
	releaseRepo biconfig.ReleaseRepo,
	stemcellRepo biconfig.StemcellRepo,
) Record {
	return &deploymentRecord{
		deploymentRepo: deploymentRepo,
		releaseRepo:    releaseRepo,
		stemcellRepo:   stemcellRepo,
	}
}

func (v *deploymentRecord) IsDeployed(newManifestSHA string, releases []birel.Release, stemcell bistemcell.ExtractedStemcell) (bool, error) {
	deployedManifestSHA, found, err := v.deploymentRepo.FindCurrent()
	if err != nil {
		return false, bosherr.WrapError(err, "Finding sha of currently deployed manifest")
	}

	if !found {
		return false, nil
	}

	if deployedManifestSHA != newManifestSHA {
		return false, nil
	}

	currentStemcell, found, err := v.stemcellRepo.FindCurrent()
	if err != nil {
		return false, bosherr.WrapError(err, "Finding currently deployed stemcell")
	}

	if !found {
		return false, nil
	}

	if currentStemcell.Name != stemcell.Manifest().Name || currentStemcell.Version != stemcell.Manifest().Version {
		return false, nil
	}

	currentReleaseRecords, err := v.releaseRepo.List()
	if err != nil {
		return false, bosherr.WrapError(err, "Finding currently deployed release")
	}

	if len(currentReleaseRecords) == 0 {
		return false, nil
	}

	if len(releases) != len(currentReleaseRecords) {
		return false, nil
	}

	for _, release := range releases {
		found := false
		for _, releaseRecord := range currentReleaseRecords {
			if releaseRecord.Name == release.Name() && releaseRecord.Version == release.Version() {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	return true, nil
}

func (v *deploymentRecord) Clear() error {
	err := v.deploymentRepo.UpdateCurrent("")
	if err != nil {
		return bosherr.WrapError(err, "Clearing sha of deployed manifest")
	}

	err = v.releaseRepo.Update([]birel.Release{})
	if err != nil {
		return bosherr.WrapError(err, "Clearing releases")
	}

	return nil
}

func (v *deploymentRecord) Update(manifestSHA string, releases []birel.Release) error {
	err := v.deploymentRepo.UpdateCurrent(manifestSHA)
	if err != nil {
		return bosherr.WrapError(err, "Saving sha of deployed manifest")
	}

	err = v.releaseRepo.Update(releases)
	if err != nil {
		return bosherr.WrapError(err, "Updating releases")
	}

	return nil
}
