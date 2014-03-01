package api

import (
	"cf/errors"
	"cf/models"
)

type FakeUserProvidedServiceInstanceRepo struct {
	CreateName     string
	CreateDrainUrl string
	CreateParams   map[string]string

	UpdateServiceInstance models.ServiceInstanceFields
}

func (repo *FakeUserProvidedServiceInstanceRepo) Create(name, drainUrl string, params map[string]string) (apiResponse errors.Error) {
	repo.CreateName = name
	repo.CreateDrainUrl = drainUrl
	repo.CreateParams = params
	return
}

func (repo *FakeUserProvidedServiceInstanceRepo) Update(serviceInstance models.ServiceInstanceFields) (apiResponse errors.Error) {
	repo.UpdateServiceInstance = serviceInstance
	return
}
