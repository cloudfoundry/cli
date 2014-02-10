package api

import (
	"cf/configuration"
	"cf/models"
	"cf/net"
	"fmt"
	"net/url"
	"strings"
)

type ServiceBrokerResource struct {
	Resource
	Entity ServiceBrokerEntity
}

func (resource ServiceBrokerResource) ToFields() (fields models.ServiceBroker) {
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
	ListServiceBrokers(callback func(models.ServiceBroker) bool) net.ApiResponse
	FindByName(name string) (serviceBroker models.ServiceBroker, apiResponse net.ApiResponse)
	Create(name, url, username, password string) (apiResponse net.ApiResponse)
	Update(serviceBroker models.ServiceBroker) (apiResponse net.ApiResponse)
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

func (repo CloudControllerServiceBrokerRepository) ListServiceBrokers(callback func(models.ServiceBroker) bool) net.ApiResponse {
	return repo.gateway.ListPaginatedResources(
		repo.config.Target,
		repo.config.AccessToken,
		"/v2/service_brokers",
		ServiceBrokerResource{},
		func(resource interface{}) bool {
			callback(resource.(ServiceBrokerResource).ToFields())
			return true
		})
}

func (repo CloudControllerServiceBrokerRepository) FindByName(name string) (serviceBroker models.ServiceBroker, apiResponse net.ApiResponse) {
	foundBroker := false
	apiResponse = repo.gateway.ListPaginatedResources(
		repo.config.Target,
		repo.config.AccessToken,
		fmt.Sprintf("/v2/service_brokers?q=%s", url.QueryEscape("name:"+name)),
		ServiceBrokerResource{},
		func(resource interface{}) bool {
			serviceBroker = resource.(ServiceBrokerResource).ToFields()
			foundBroker = true
			return false
		})

	if !foundBroker {
		apiResponse = net.NewNotFoundApiResponse("Service Broker '%s' not found", name)
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

func (repo CloudControllerServiceBrokerRepository) Update(serviceBroker models.ServiceBroker) (apiResponse net.ApiResponse) {
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
