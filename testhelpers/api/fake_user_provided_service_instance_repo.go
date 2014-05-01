package api

import "github.com/cloudfoundry/cli/cf/models"

type FakeUserProvidedServiceInstanceRepo struct {
	CreateName     string
	CreateDrainUrl string
	CreateParams   map[string]string

	UpdateServiceInstance models.ServiceInstanceFields
}

func (repo *FakeUserProvidedServiceInstanceRepo) Create(name, drainUrl string, params map[string]string) (apiErr error) {
	repo.CreateName = name
	repo.CreateDrainUrl = drainUrl
	repo.CreateParams = params
	return
}

func (repo *FakeUserProvidedServiceInstanceRepo) Update(serviceInstance models.ServiceInstanceFields) (apiErr error) {
	repo.UpdateServiceInstance = serviceInstance
	return
}
