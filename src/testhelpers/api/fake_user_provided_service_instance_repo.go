package api

import (
	"cf"
	"cf/net"
)

type FakeUserProvidedServiceInstanceRepo struct {
	CreateServiceInstance cf.ServiceInstance

	UpdateServiceInstance cf.ServiceInstance
}

func (repo *FakeUserProvidedServiceInstanceRepo) Create(serviceInstance cf.ServiceInstance) (apiResponse net.ApiResponse) {
	repo.CreateServiceInstance = serviceInstance
	return
}

func (repo *FakeUserProvidedServiceInstanceRepo) Update(serviceInstance cf.ServiceInstance) (apiResponse net.ApiResponse) {
	repo.UpdateServiceInstance = serviceInstance
	return
}
