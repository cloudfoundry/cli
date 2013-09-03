package testhelpers

import (
	"cf"
	"cf/configuration"
)

type FakeServiceRepo struct {
	ServiceOfferings []cf.ServiceOffering

	CreateServiceInstanceName string
	CreateServiceInstancePlan cf.ServicePlan

	FindInstanceByNameName string
	FindInstanceByNameServiceInstance cf.ServiceInstance

	BindServiceServiceInstance cf.ServiceInstance
	BindServiceApplication cf.Application

	UnbindServiceServiceInstance cf.ServiceInstance
	UnbindServiceApplication cf.Application

	DeleteServiceServiceInstance cf.ServiceInstance
}

func (repo *FakeServiceRepo) GetServiceOfferings(config *configuration.Configuration) (offerings []cf.ServiceOffering, err error) {
	offerings = repo.ServiceOfferings
	return
}

func (repo *FakeServiceRepo) CreateServiceInstance(config *configuration.Configuration, name string, plan cf.ServicePlan) (err error) {
	repo.CreateServiceInstanceName = name
	repo.CreateServiceInstancePlan = plan
	return
}

func (repo *FakeServiceRepo) FindInstanceByName(config *configuration.Configuration, name string) (instance cf.ServiceInstance, err error) {
	repo.FindInstanceByNameName = name
	instance = repo.FindInstanceByNameServiceInstance
	return
}

func (repo *FakeServiceRepo) BindService(config *configuration.Configuration, instance cf.ServiceInstance, app cf.Application) (err error) {
	repo.BindServiceServiceInstance = instance
	repo.BindServiceApplication = app
	return
}

func (repo *FakeServiceRepo) UnbindService(config *configuration.Configuration, instance cf.ServiceInstance, app cf.Application) (err error) {
	repo.UnbindServiceServiceInstance = instance
	repo.UnbindServiceApplication = app
	return
}

func (repo *FakeServiceRepo) DeleteService(config *configuration.Configuration, instance cf.ServiceInstance) (err error) {
	repo.DeleteServiceServiceInstance = instance
	return
}
