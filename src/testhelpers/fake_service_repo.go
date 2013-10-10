package testhelpers

import (
	"cf"
	"cf/net"
	"net/http"
)

type FakeServiceRepo struct {
	ServiceOfferings []cf.ServiceOffering

	CreateServiceInstanceName string
	CreateServiceInstancePlan cf.ServicePlan
	CreateServiceAlreadyExists bool

	CreateUserProvidedServiceInstanceName string
	CreateUserProvidedServiceInstanceParameters map[string]string

	UpdateUserProvidedServiceInstanceServiceInstance cf.ServiceInstance
	UpdateUserProvidedServiceInstanceParameters map[string]string

	FindInstanceByNameName string
	FindInstanceByNameServiceInstance cf.ServiceInstance
	FindInstanceByNameErr bool
	FindInstanceByNameNotFound bool

	BindServiceServiceInstance cf.ServiceInstance
	BindServiceApplication cf.Application
	BindServiceErrorCode string

	UnbindServiceServiceInstance cf.ServiceInstance
	UnbindServiceApplication cf.Application
	UnbindServiceBindingNotFound bool

	DeleteServiceServiceInstance cf.ServiceInstance

	RenameServiceServiceInstance cf.ServiceInstance
	RenameServiceNewName string
}

func (repo *FakeServiceRepo) GetServiceOfferings() (offerings []cf.ServiceOffering, apiResponse net.ApiResponse) {
	offerings = repo.ServiceOfferings
	return
}

func (repo *FakeServiceRepo) CreateServiceInstance(name string, plan cf.ServicePlan) (identicalAlreadyExists bool, apiResponse net.ApiResponse) {
	repo.CreateServiceInstanceName = name
	repo.CreateServiceInstancePlan = plan
	identicalAlreadyExists = repo.CreateServiceAlreadyExists

	return
}

func (repo *FakeServiceRepo) CreateUserProvidedServiceInstance(name string, params map[string]string) (apiResponse net.ApiResponse) {
	repo.CreateUserProvidedServiceInstanceName = name
	repo.CreateUserProvidedServiceInstanceParameters = params
	return
}

func (repo *FakeServiceRepo) UpdateUserProvidedServiceInstance(serviceInstance cf.ServiceInstance, params map[string]string) (apiResponse net.ApiResponse) {
	repo.UpdateUserProvidedServiceInstanceServiceInstance = serviceInstance
	repo.UpdateUserProvidedServiceInstanceParameters = params
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

func (repo *FakeServiceRepo) BindService(instance cf.ServiceInstance, app cf.Application) (apiResponse net.ApiResponse) {
	repo.BindServiceServiceInstance = instance
	repo.BindServiceApplication = app

	if repo.BindServiceErrorCode != "" {
		apiResponse = net.NewApiResponse("Error binding service", repo.BindServiceErrorCode, http.StatusBadRequest)
	}

	return
}

func (repo *FakeServiceRepo) UnbindService(instance cf.ServiceInstance, app cf.Application) (found bool, apiResponse net.ApiResponse) {
	repo.UnbindServiceServiceInstance = instance
	repo.UnbindServiceApplication = app
	found = !repo.UnbindServiceBindingNotFound
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
