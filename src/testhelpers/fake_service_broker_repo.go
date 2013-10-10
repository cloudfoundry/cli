package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeServiceBrokerRepo struct {
	FindByNameName string
	FindByNameServiceBroker cf.ServiceBroker
	FindByNameNotFound bool

	CreatedServiceBroker cf.ServiceBroker
	UpdatedServiceBroker cf.ServiceBroker
	DeletedServiceBroker cf.ServiceBroker
}

func (repo *FakeServiceBrokerRepo) FindByName(name string) (serviceBroker cf.ServiceBroker, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	serviceBroker = repo.FindByNameServiceBroker

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("Service Broker", name)
	}

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

func (repo *FakeServiceBrokerRepo) Delete(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	repo.DeletedServiceBroker = serviceBroker
	return
}
