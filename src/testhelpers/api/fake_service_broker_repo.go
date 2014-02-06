package api

import (
	"cf"
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

func (repo *FakeServiceBrokerRepo) ListServiceBrokers(stop chan bool) (serviceBrokersChan chan []models.ServiceBroker, statusChan chan net.ApiResponse) {
	serviceBrokersChan = make(chan []models.ServiceBroker, 4)
	statusChan = make(chan net.ApiResponse, 1)

	if repo.ListErr {
		statusChan <- net.NewApiResponseWithMessage("Error finding all routes")
		close(serviceBrokersChan)
		close(statusChan)
		return
	}

	go func() {
		serviceBrokersCount := len(repo.ServiceBrokers)
		for i := 0; i < serviceBrokersCount; i += 2 {
			select {
			case <-stop:
				break
			default:
				if serviceBrokersCount-i > 1 {
					serviceBrokersChan <- repo.ServiceBrokers[i : i+2]
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
