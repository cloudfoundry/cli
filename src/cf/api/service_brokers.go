package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"io"
	"strings"
)

type PaginatedServiceBrokerResources struct {
	ServiceBrokers []ServiceBrokerResource `json:"resources"`
}

type ServiceBrokerResource struct {
	Resource
	Entity ServiceBrokerEntity
}

type ServiceBrokerEntity struct {
	Guid     string
	Name     string
	Password string `json:"auth_password"`
	Username string `json:"auth_username"`
	Url      string `json:"broker_url"`
}

type ServiceBrokerRepository interface {
	FindAll() (serviceBrokers []cf.ServiceBroker, apiResponse net.ApiResponse)
	FindByName(name string) (serviceBroker cf.ServiceBroker, apiResponse net.ApiResponse)
	Create(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse)
	Update(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse)
	Rename(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse)
	Delete(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse)
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

func (repo CloudControllerServiceBrokerRepository) FindAll() (serviceBrokers []cf.ServiceBroker, apiResponse net.ApiResponse) {
	return repo.findAllWithPath(fmt.Sprintf("%s/v2/service_brokers", repo.config.Target))
}

func (repo CloudControllerServiceBrokerRepository) FindByName(name string) (serviceBroker cf.ServiceBroker, apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_brokers?q=name%%3A%s", repo.config.Target, name)
	serviceBrokers, apiResponse := repo.findAllWithPath(path)
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

func (repo CloudControllerServiceBrokerRepository) findAllWithPath(path string) (serviceBrokers []cf.ServiceBroker, apiResponse net.ApiResponse) {
	req, apiResponse := repo.gateway.NewRequest("GET", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	resources := new(PaginatedServiceBrokerResources)
	_, apiResponse = repo.gateway.PerformRequestForJSONResponse(req, resources)
	if apiResponse.IsNotSuccessful() {
		return
	}

	for _, resource := range resources.ServiceBrokers {
		serviceBroker := cf.ServiceBroker{
			Name:     resource.Entity.Name,
			Guid:     resource.Metadata.Guid,
			Url:      resource.Entity.Url,
			Username: resource.Entity.Username,
			Password: resource.Entity.Password,
		}

		serviceBrokers = append(serviceBrokers, serviceBroker)
	}
	return
}

func (repo CloudControllerServiceBrokerRepository) Create(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	body := fmt.Sprintf(
		`{"name":"%s","broker_url":"%s","auth_username":"%s","auth_password":"%s"}`,
		serviceBroker.Name, serviceBroker.Url, serviceBroker.Username, serviceBroker.Password,
	)
	return repo.createUpdateOrDelete(serviceBroker, "POST", strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Update(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	body := fmt.Sprintf(
		`{"broker_url":"%s","auth_username":"%s","auth_password":"%s"}`,
		serviceBroker.Url, serviceBroker.Username, serviceBroker.Password,
	)
	return repo.createUpdateOrDelete(serviceBroker, "PUT", strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Rename(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	body := fmt.Sprintf(`{"name":"%s"}`, serviceBroker.Name)

	return repo.createUpdateOrDelete(serviceBroker, "PUT", strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Delete(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	return repo.createUpdateOrDelete(serviceBroker, "DELETE", nil)
}

func (repo CloudControllerServiceBrokerRepository) createUpdateOrDelete(serviceBroker cf.ServiceBroker, verb string, body io.Reader) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_brokers", repo.config.Target)

	if serviceBroker.Guid != "" {
		path = fmt.Sprintf("%s/v2/service_brokers/%s", repo.config.Target, serviceBroker.Guid)
	}

	req, apiResponse := repo.gateway.NewRequest(verb, path, repo.config.AccessToken, body)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(req)
	return
}
