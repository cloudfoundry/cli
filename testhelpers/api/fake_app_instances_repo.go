package api

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"time"
)

type FakeAppInstancesRepo struct {
	GetInstancesAppGuid    string
	GetInstancesResponses  [][]models.AppInstanceFields
	GetInstancesErrorCodes []string
}

func (repo *FakeAppInstancesRepo) GetInstances(appGuid string) (instances []models.AppInstanceFields, apiErr error) {
	repo.GetInstancesAppGuid = appGuid
	time.Sleep(1 * time.Millisecond) //needed for Windows only, otherwise it thinks error codes are not assigned

	if len(repo.GetInstancesResponses) > 0 {
		instances = repo.GetInstancesResponses[0]
		repo.GetInstancesResponses = repo.GetInstancesResponses[1:]
	}

	if len(repo.GetInstancesErrorCodes) > 0 {
		errorCode := repo.GetInstancesErrorCodes[0]
		repo.GetInstancesErrorCodes = repo.GetInstancesErrorCodes[1:]
		if errorCode != "" {
			apiErr = errors.NewHttpError(400, errorCode, "Error staging app")
		}
	}

	return
}
