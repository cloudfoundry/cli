package testhelpers

import (
	"cf"
	"cf/net"
)

type FakeServiceBrokerRepo struct {
	CreatedServiceBroker cf.ServiceBroker
}

func (repo *FakeServiceBrokerRepo) Create(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	repo.CreatedServiceBroker = serviceBroker
	return
}
