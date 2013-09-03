package testhelpers

import (
	"cf"
	"cf/configuration"
)

type FakeServiceRepo struct {
	ServiceOfferings []cf.ServiceOffering
	CreateServiceInstanceName string
	CreateServiceInstancePlan cf.ServicePlan
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



