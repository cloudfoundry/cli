package api

import (
	"cf"
	"cf/net"
)

type FakeServiceSummaryRepo struct{
	GetSummariesInCurrentSpaceInstances []cf.ServiceInstance
}

func (repo *FakeServiceSummaryRepo)GetSummariesInCurrentSpace() (instances []cf.ServiceInstance, apiResponse net.ApiResponse) {
	instances = repo.GetSummariesInCurrentSpaceInstances
	return
}
