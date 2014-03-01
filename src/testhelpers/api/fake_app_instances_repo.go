package api

import (
	"cf/errors"
	"cf/models"
	"time"
)

type FakeAppInstancesRepo struct {
	GetInstancesAppGuid    string
	GetInstancesResponses  [][]models.AppInstanceFields
	GetInstancesErrorCodes []string
}

func (repo *FakeAppInstancesRepo) GetInstances(appGuid string) (instances []models.AppInstanceFields, apiErr errors.Error) {
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
			apiErr = errors.NewError("Error staging app", errorCode)
		}
	}

	return
}
