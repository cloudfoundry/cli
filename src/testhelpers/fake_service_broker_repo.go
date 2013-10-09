package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeServiceBrokerRepo struct {
	FindByNameName string
	FindByNameServiceBroker cf.ServiceBroker

	CreatedServiceBroker cf.ServiceBroker
	UpdatedServiceBroker cf.ServiceBroker
}

func (repo *FakeServiceBrokerRepo) FindByName(name string) (serviceBroker cf.ServiceBroker, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	serviceBroker = repo.FindByNameServiceBroker
	return
}

func (repo *FakeServiceBrokerRepo) Create(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	repo.CreatedServiceBroker = serviceBroker
	return
}

func (repo *FakeServiceBrokerRepo) Update(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	repo.UpdatedServiceBroker = serviceBroker
	return
}
