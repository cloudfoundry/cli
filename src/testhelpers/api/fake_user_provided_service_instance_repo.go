package api

import (
	"cf"
	"cf/net"
)

type FakeUserProvidedServiceInstanceRepo struct {
	CreateName string
	CreateDrainUrl string
	CreateParams map[string]string

	UpdateServiceInstance cf.ServiceInstanceFields
}

func (repo *FakeUserProvidedServiceInstanceRepo) Create(name, drainUrl string, params map[string]string) (apiResponse net.ApiResponse) {
	repo.CreateName = name
	repo.CreateDrainUrl = drainUrl
	repo.CreateParams = params
	return
}

func (repo *FakeUserProvidedServiceInstanceRepo) Update(serviceInstance cf.ServiceInstanceFields) (apiResponse net.ApiResponse) {
	repo.UpdateServiceInstance = serviceInstance
	return
}
