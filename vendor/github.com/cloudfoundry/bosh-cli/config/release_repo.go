package config

import (
	"github.com/cloudfoundry/bosh-cli/release"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshuuid "github.com/cloudfoundry/bosh-utils/uuid"
)

// ReleaseRepo persists releases metadata
type ReleaseRepo interface {
	List() ([]ReleaseRecord, error)
	Update([]release.Release) error
}

type releaseRepo struct {
	deploymentStateService DeploymentStateService
	uuidGenerator          boshuuid.Generator
}

func NewReleaseRepo(deploymentStateService DeploymentStateService, uuidGenerator boshuuid.Generator) ReleaseRepo {
	return releaseRepo{
		deploymentStateService: deploymentStateService,
		uuidGenerator:          uuidGenerator,
	}
}

func (r releaseRepo) Update(releases []release.Release) error {
	newRecordIDs := []string{}
	newRecords := []ReleaseRecord{}

	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return bosherr.WrapError(err, "Loading existing config")
	}

	for _, release := range releases {
		newRecord := ReleaseRecord{
			Name:    release.Name(),
			Version: release.Version(),
		}
		newRecord.ID, err = r.uuidGenerator.Generate()
		if err != nil {
			return bosherr.WrapError(err, "Generating release id")
		}
		newRecords = append(newRecords, newRecord)
		newRecordIDs = append(newRecordIDs, newRecord.ID)
	}

	deploymentState.CurrentReleaseIDs = newRecordIDs
	deploymentState.Releases = newRecords
	err = r.deploymentStateService.Save(deploymentState)
	if err != nil {
		return bosherr.WrapError(err, "Updating current release record")
	}
	return nil
}

func (r releaseRepo) List() ([]ReleaseRecord, error) {
	deploymentState, err := r.deploymentStateService.Load()
	if err != nil {
		return []ReleaseRecord{}, bosherr.WrapError(err, "Loading existing config")
	}
	return deploymentState.Releases, nil
}
