package api

import (
	"cf/errors"
	"cf/models"
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

func (repo *FakeServiceBrokerRepo) FindByName(name string) (serviceBroker models.ServiceBroker, apiResponse errors.Error) {
	repo.FindByNameName = name
	serviceBroker = repo.FindByNameServiceBroker

	if repo.FindByNameNotFound {
		apiResponse = errors.NewNotFoundError("%s %s not found", "Service Broker", name)
	}

	return
}

func (repo *FakeServiceBrokerRepo) ListServiceBrokers(callback func(broker models.ServiceBroker) bool) errors.Error {
	for _, broker := range repo.ServiceBrokers {
		if !callback(broker) {
			break
		}
	}

	if repo.ListErr {
		return errors.NewErrorWithMessage("Error finding service brokers")
	} else {
		return errors.NewErrorWithStatusCode(200)
	}
}

func (repo *FakeServiceBrokerRepo) Create(name, url, username, password string) (apiResponse errors.Error) {
	repo.CreateName = name
	repo.CreateUrl = url
	repo.CreateUsername = username
	repo.CreatePassword = password
	return
}

func (repo *FakeServiceBrokerRepo) Update(serviceBroker models.ServiceBroker) (apiResponse errors.Error) {
	repo.UpdatedServiceBroker = serviceBroker
	return
}

func (repo *FakeServiceBrokerRepo) Rename(guid, name string) (apiResponse errors.Error) {
	repo.RenamedServiceBrokerGuid = guid
	repo.RenamedServiceBrokerName = name
	return
}

func (repo *FakeServiceBrokerRepo) Delete(guid string) (apiResponse errors.Error) {
	repo.DeletedServiceBrokerGuid = guid
	return
}
