package apifakes

import "code.cloudfoundry.org/cli/v9/cf/models"

type OldFakeServiceSummaryRepo struct {
	GetSummariesInCurrentSpaceInstances []models.ServiceInstance
}

func (repo *OldFakeServiceSummaryRepo) GetSummariesInCurrentSpace() (instances []models.ServiceInstance, apiErr error) {
	instances = repo.GetSummariesInCurrentSpaceInstances
	return
}
