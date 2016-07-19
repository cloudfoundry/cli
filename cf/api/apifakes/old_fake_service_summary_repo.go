package apifakes

import "code.cloudfoundry.org/cli/cf/models"

type OldFakeServiceSummaryRepo struct {
	GetSummariesInCurrentSpaceInstances []models.ServiceInstance
}

func (repo *OldFakeServiceSummaryRepo) GetSummariesInCurrentSpace() (instances []models.ServiceInstance, apiErr error) {
	instances = repo.GetSummariesInCurrentSpaceInstances
	return
}
