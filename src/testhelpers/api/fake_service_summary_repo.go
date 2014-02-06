package api

import (
"cf/models"
	"cf/net"
)

type FakeServiceSummaryRepo struct {
	GetSummariesInCurrentSpaceInstances []models.ServiceInstance
}

func (repo *FakeServiceSummaryRepo) GetSummariesInCurrentSpace() (instances []models.ServiceInstance, apiResponse net.ApiResponse) {
	instances = repo.GetSummariesInCurrentSpaceInstances
	return
}
