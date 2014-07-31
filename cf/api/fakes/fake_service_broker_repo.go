package fakes

import (
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
)

type FakeServiceBrokerRepo struct {
	FindByNameName          string
	FindByNameServiceBroker models.ServiceBroker
	FindByNameNotFound      bool

	FindByGuidGuid          string
	FindByGuidServiceBroker models.ServiceBroker
	FindByGuidNotFound      bool

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

func (repo *FakeServiceBrokerRepo) FindByName(name string) (serviceBroker models.ServiceBroker, apiErr error) {
	repo.FindByNameName = name
	serviceBroker = repo.FindByNameServiceBroker

	if repo.FindByNameNotFound {
		apiErr = errors.NewModelNotFoundError("Service Broker", name)
	}

	return
}

func (repo *FakeServiceBrokerRepo) FindByGuid(guid string) (serviceBroker models.ServiceBroker, apiErr error) {
	repo.FindByGuidGuid = guid
	serviceBroker = repo.FindByGuidServiceBroker

	if repo.FindByGuidNotFound {
		apiErr = errors.NewModelNotFoundError("Service Broker", guid)
	}

	return
}

func (repo *FakeServiceBrokerRepo) ListServiceBrokers(callback func(broker models.ServiceBroker) bool) error {
	for _, broker := range repo.ServiceBrokers {
		if !callback(broker) {
			break
		}
	}

	if repo.ListErr {
		return errors.New("Error finding service brokers")
	} else {
		return nil
	}
}

func (repo *FakeServiceBrokerRepo) Create(name, url, username, password string) (apiErr error) {
	repo.CreateName = name
	repo.CreateUrl = url
	repo.CreateUsername = username
	repo.CreatePassword = password
	return
}

func (repo *FakeServiceBrokerRepo) Update(serviceBroker models.ServiceBroker) (apiErr error) {
	repo.UpdatedServiceBroker = serviceBroker
	return
}

func (repo *FakeServiceBrokerRepo) Rename(guid, name string) (apiErr error) {
	repo.RenamedServiceBrokerGuid = guid
	repo.RenamedServiceBrokerName = name
	return
}

func (repo *FakeServiceBrokerRepo) Delete(guid string) (apiErr error) {
	repo.DeletedServiceBrokerGuid = guid
	return
}
