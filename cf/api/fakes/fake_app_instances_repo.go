package fakes

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeAppInstancesRepo struct {
	GetInstancesAppGuid    string
	GetInstancesResponses  [][]models.AppInstanceFields
	GetInstancesErrorCodes []string
}

func (repo *FakeAppInstancesRepo) GetInstances(appGuid string) (instances []models.AppInstanceFields, apiErr error) {
	repo.GetInstancesAppGuid = appGuid

	if len(repo.GetInstancesResponses) > 0 {
		instances = repo.GetInstancesResponses[0]

		if len(repo.GetInstancesResponses) > 1 {
			repo.GetInstancesResponses = repo.GetInstancesResponses[1:]
		}
	}

	if len(repo.GetInstancesErrorCodes) > 0 {
		errorCode := repo.GetInstancesErrorCodes[0]

		// don't slice away the last one if this is all we have
		if len(repo.GetInstancesErrorCodes) > 1 {
			repo.GetInstancesErrorCodes = repo.GetInstancesErrorCodes[1:]
		}

		if errorCode != "" {
			apiErr = errors.NewHttpError(400, errorCode, "Error staging app")
		}
	}

	return
}
