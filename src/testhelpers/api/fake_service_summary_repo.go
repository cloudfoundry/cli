package api

import (
	"cf/errors"
	"cf/models"
)

type FakeServiceSummaryRepo struct {
	GetSummariesInCurrentSpaceInstances []models.ServiceInstance
}

func (repo *FakeServiceSummaryRepo) GetSummariesInCurrentSpace() (instances []models.ServiceInstance, apiErr errors.Error) {
	instances = repo.GetSummariesInCurrentSpaceInstances
	return
}
