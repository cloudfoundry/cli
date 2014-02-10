package api

import (
	"cf/models"
	"cf/net"
)

type FakeServiceBrokerRepo struct {
	FindByNameName          string
	FindByNameServiceBroker models.ServiceBroker
	FindByNameNotFound      bool

	CreateName     string
	CreateUrl      string
	CreateUsername string
	CreatePassword string

	UpdatedServiceBroker     models.ServiceBroker
	RenamedServiceBrokerGuid string
	RenamedServiceBrokerName string
	DeletedServiceBrokerGuid string

	ServiceBrokers []models.ServiceBroker
	ListErr        bool
}

func (repo *FakeServiceBrokerRepo) FindByName(name string) (serviceBroker models.ServiceBroker, apiResponse net.ApiResponse) {
	repo.FindByNameName = name
	serviceBroker = repo.FindByNameServiceBroker

	if repo.FindByNameNotFound {
		apiResponse = net.NewNotFoundApiResponse("%s %s not found", "Service Broker", name)
	}

	return
}

func (repo *FakeServiceBrokerRepo) ListServiceBrokers(callback func(broker models.ServiceBroker) bool) net.ApiResponse {
	for _, broker := range repo.ServiceBrokers {
		if !callback(broker) {
			break
		}
	}

	if repo.ListErr {
		return net.NewApiResponseWithMessage("Error finding service brokers")
	} else {
		return net.NewApiResponseWithStatusCode(200)
	}
}

func (repo *FakeServiceBrokerRepo) Create(name, url, username, password string) (apiResponse net.ApiResponse) {
	repo.CreateName = name
	repo.CreateUrl = url
	repo.CreateUsername = username
	repo.CreatePassword = password
	return
}

func (repo *FakeServiceBrokerRepo) Update(serviceBroker models.ServiceBroker) (apiResponse net.ApiResponse) {
	repo.UpdatedServiceBroker = serviceBroker
	return
}

func (repo *FakeServiceBrokerRepo) Rename(guid, name string) (apiResponse net.ApiResponse) {
	repo.RenamedServiceBrokerGuid = guid
	repo.RenamedServiceBrokerName = name
	return
}

func (repo *FakeServiceBrokerRepo) Delete(guid string) (apiResponse net.ApiResponse) {
	repo.DeletedServiceBrokerGuid = guid
	return
}
