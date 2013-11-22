package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type PaginatedServiceBrokerResources struct {
	ServiceBrokers []ServiceBrokerResource `json:"resources"`
	NextUrl        string                  `json:"next_url"`
}

type ServiceBrokerResource struct {
	Resource
	Entity ServiceBrokerEntity
}

func (resource ServiceBrokerResource) ToFields() (fields cf.ServiceBroker) {
	fields.Name = resource.Entity.Name
	fields.Guid = resource.Metadata.Guid
	fields.Url = resource.Entity.Url
	fields.Username = resource.Entity.Username
	fields.Password = resource.Entity.Password
	return
}

type ServiceBrokerEntity struct {
	Guid     string
	Name     string
	Password string `json:"auth_password"`
	Username string `json:"auth_username"`
	Url      string `json:"broker_url"`
}

type ServiceBrokerRepository interface {
	ListServiceBrokers(stop chan bool) (serviceBrokersChan chan []cf.ServiceBroker, statusChan chan net.ApiResponse)
	FindByName(name string) (serviceBroker cf.ServiceBroker, apiResponse net.ApiResponse)
	Create(name, url, username, password string) (apiResponse net.ApiResponse)
	Update(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse)
	Rename(guid, name string) (apiResponse net.ApiResponse)
	Delete(guid string) (apiResponse net.ApiResponse)
}

type CloudControllerServiceBrokerRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerServiceBrokerRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerServiceBrokerRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerServiceBrokerRepository) ListServiceBrokers(stop chan bool) (serviceBrokersChan chan []cf.ServiceBroker, statusChan chan net.ApiResponse) {
	serviceBrokersChan = make(chan []cf.ServiceBroker, 4)
	statusChan = make(chan net.ApiResponse, 1)

	go func() {
		path := "/v2/service_brokers"

	loop:
		for path != "" {
			select {
			case <-stop:
				break loop
			default:
				var (
					serviceBrokers []cf.ServiceBroker
					apiResponse    net.ApiResponse
				)
				serviceBrokers, path, apiResponse = repo.findNextWithPath(path)
				if apiResponse.IsNotSuccessful() {
					statusChan <- apiResponse
					close(serviceBrokersChan)
					close(statusChan)
					return
				}

				if len(serviceBrokers) > 0 {
					serviceBrokersChan <- serviceBrokers
				}
			}
		}
		close(serviceBrokersChan)
		close(statusChan)
		cf.WaitForClose(stop)
	}()

	return
}

func (repo CloudControllerServiceBrokerRepository) FindByName(name string) (serviceBroker cf.ServiceBroker, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("/v2/service_brokers?q=name%%3A%s", name)
	serviceBrokers, _, apiResponse := repo.findNextWithPath(path)
	if apiResponse.IsNotSuccessful() {
		return
	}

	if len(serviceBrokers) == 0 {
		apiResponse = net.NewNotFoundApiResponse("Service Broker %s not found", name)
		return
	}

	serviceBroker = serviceBrokers[0]
	return
}

func (repo CloudControllerServiceBrokerRepository) findNextWithPath(path string) (serviceBrokers []cf.ServiceBroker, nextUrl string, apiResponse net.ApiResponse) {
	resources := new(PaginatedServiceBrokerResources)

	apiResponse = repo.gateway.GetResource(repo.config.Target+path, repo.config.AccessToken, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	nextUrl = resources.NextUrl

	for _, resource := range resources.ServiceBrokers {
		serviceBrokers = append(serviceBrokers, resource.ToFields())
	}
	return
}

func (repo CloudControllerServiceBrokerRepository) Create(name, url, username, password string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_brokers", repo.config.Target)
	body := fmt.Sprintf(
		`{"name":"%s","broker_url":"%s","auth_username":"%s","auth_password":"%s"}`, name, url, username, password,
	)
	return repo.gateway.CreateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Update(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_brokers/%s", repo.config.Target, serviceBroker.Guid)
	body := fmt.Sprintf(
		`{"broker_url":"%s","auth_username":"%s","auth_password":"%s"}`,
		serviceBroker.Url, serviceBroker.Username, serviceBroker.Password,
	)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Rename(guid, name string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_brokers/%s", repo.config.Target, guid)
	body := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Delete(guid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_brokers/%s", repo.config.Target, guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
