package api

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
	RenamedServiceBroker cf.ServiceBroker
	DeletedServiceBroker cf.ServiceBroker

	ServiceBrokers []cf.ServiceBroker
	ListErr bool
}

func (repo *FakeServiceBrokerRepo) FindByName(name string) (serviceBroker cf.ServiceBroker, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	serviceBroker = repo.FindByNameServiceBroker

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found","Service Broker", name)
	}

	return
}

func (repo *FakeServiceBrokerRepo) ListServiceBrokers(stop chan bool) (serviceBrokersChan chan []cf.ServiceBroker, statusChan chan net.ApiResponse) {
	serviceBrokersChan = make(chan []cf.ServiceBroker, 4)
	statusChan = make(chan net.ApiResponse, 1)

	if repo.ListErr {
		statusChan <- net.NewApiResponseWithMessage("Error finding all routes")
		close(serviceBrokersChan)
		close(statusChan)
		return
	}


	go func() {
		serviceBrokersCount := len(repo.ServiceBrokers)
		for i:= 0; i < serviceBrokersCount; i += 2 {
			select {
			case <-stop:
				break
			default:
				if serviceBrokersCount - i > 1 {
					serviceBrokersChan <- repo.ServiceBrokers[i:i+2]
				} else {
					serviceBrokersChan <- repo.ServiceBrokers[i:]
				}
			}
		}

		close(serviceBrokersChan)
		close(statusChan)

		cf.WaitForClose(stop)
	}()

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

func (repo *FakeServiceBrokerRepo) Rename(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	repo.RenamedServiceBroker = serviceBroker
	return
}

func (repo *FakeServiceBrokerRepo) Delete(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	repo.DeletedServiceBroker = serviceBroker
	return
}
