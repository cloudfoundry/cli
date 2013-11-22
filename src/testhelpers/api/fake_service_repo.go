package api

import (
	"cf"
	"cf/net"
)

type FakeServiceRepo struct {
	ServiceOfferings []cf.ServiceOffering

	CreateServiceInstanceName string
	CreateServiceInstancePlanGuid string
	CreateServiceAlreadyExists bool

	FindInstanceByNameName string
	FindInstanceByNameServiceInstance cf.ServiceInstance
	FindInstanceByNameErr bool
	FindInstanceByNameNotFound bool

	DeleteServiceServiceInstance cf.ServiceInstance

	RenameServiceServiceInstance cf.ServiceInstance
	RenameServiceNewName string
}

func (repo *FakeServiceRepo) GetServiceOfferings() (offerings []cf.ServiceOffering, apiResponse net.ApiResponse) {
	offerings = repo.ServiceOfferings
	return
}

func (repo *FakeServiceRepo) CreateServiceInstance(name, planGuid string) (identicalAlreadyExists bool, apiResponse net.ApiResponse) {
	repo.CreateServiceInstanceName = name
	repo.CreateServiceInstancePlanGuid = planGuid
	identicalAlreadyExists = repo.CreateServiceAlreadyExists

	return
}

func (repo *FakeServiceRepo) FindInstanceByName(name string) (instance cf.ServiceInstance, apiResponse net.ApiResponse) {
	repo.FindInstanceByNameName = name
	instance = repo.FindInstanceByNameServiceInstance

	if repo.FindInstanceByNameErr {
		apiResponse = net.NewApiResponseWithMessage("Error finding instance")
	}

	if repo.FindInstanceByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","Service instance", name)
	}

	return
}

func (repo *FakeServiceRepo) DeleteService(instance cf.ServiceInstance) (apiResponse net.ApiResponse) {
	repo.DeleteServiceServiceInstance = instance
	return
}

func (repo *FakeServiceRepo) RenameService(instance cf.ServiceInstance, newName string) (apiResponse net.ApiResponse){
	repo.RenameServiceServiceInstance = instance
	repo.RenameServiceNewName = newName
	return
}
