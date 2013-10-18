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
	resources := new(PaginatedServiceBrokerResources)

	apiResponse = repo.gateway.GetResource(path, repo.config.AccessToken, resources)
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
	path := fmt.Sprintf("%s/v2/service_brokers", repo.config.Target)
	body := fmt.Sprintf(
		`{"name":"%s","broker_url":"%s","auth_username":"%s","auth_password":"%s"}`,
		serviceBroker.Name, serviceBroker.Url, serviceBroker.Username, serviceBroker.Password,
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

func (repo CloudControllerServiceBrokerRepository) Rename(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_brokers/%s", repo.config.Target, serviceBroker.Guid)
	body := fmt.Sprintf(`{"name":"%s"}`, serviceBroker.Name)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Delete(serviceBroker cf.ServiceBroker) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_brokers/%s", repo.config.Target, serviceBroker.Guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken)
}
