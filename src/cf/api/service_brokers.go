package api

import (
	"cf/configuration"
	"cf/errors"
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
	ListServiceBrokers(callback func(models.ServiceBroker) bool) error
	FindByName(name string) (serviceBroker models.ServiceBroker, apiErr error)
	Create(name, url, username, password string) (apiErr error)
	Update(serviceBroker models.ServiceBroker) (apiErr error)
	Rename(guid, name string) (apiErr error)
	Delete(guid string) (apiErr error)
}

type CloudControllerServiceBrokerRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerServiceBrokerRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerServiceBrokerRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerServiceBrokerRepository) ListServiceBrokers(callback func(models.ServiceBroker) bool) error {
	return repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		"/v2/service_brokers",
		ServiceBrokerResource{},
		func(resource interface{}) bool {
			callback(resource.(ServiceBrokerResource).ToFields())
			return true
		})
}

func (repo CloudControllerServiceBrokerRepository) FindByName(name string) (serviceBroker models.ServiceBroker, apiErr error) {
	foundBroker := false
	apiErr = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		repo.config.AccessToken(),
		fmt.Sprintf("/v2/service_brokers?q=%s", url.QueryEscape("name:"+name)),
		ServiceBrokerResource{},
		func(resource interface{}) bool {
			serviceBroker = resource.(ServiceBrokerResource).ToFields()
			foundBroker = true
			return false
		})

	if !foundBroker {
		apiErr = errors.NewModelNotFoundError("Service Broker", name)
	}

	return
}

func (repo CloudControllerServiceBrokerRepository) Create(name, url, username, password string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/service_brokers", repo.config.ApiEndpoint())
	body := fmt.Sprintf(
		`{"name":"%s","broker_url":"%s","auth_username":"%s","auth_password":"%s"}`, name, url, username, password,
	)
	return repo.gateway.CreateResource(path, repo.config.AccessToken(), strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Update(serviceBroker models.ServiceBroker) (apiErr error) {
	path := fmt.Sprintf("%s/v2/service_brokers/%s", repo.config.ApiEndpoint(), serviceBroker.Guid)
	body := fmt.Sprintf(
		`{"broker_url":"%s","auth_username":"%s","auth_password":"%s"}`,
		serviceBroker.Url, serviceBroker.Username, serviceBroker.Password,
	)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken(), strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Rename(guid, name string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/service_brokers/%s", repo.config.ApiEndpoint(), guid)
	body := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.UpdateResource(path, repo.config.AccessToken(), strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Delete(guid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/service_brokers/%s", repo.config.ApiEndpoint(), guid)
	return repo.gateway.DeleteResource(path, repo.config.AccessToken())
}
