package api

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	"net/url"
	"strings"
)

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
		"/v2/service_brokers",
		resources.ServiceBrokerResource{},
		func(resource interface{}) bool {
			callback(resource.(resources.ServiceBrokerResource).ToFields())
			return true
		})
}

func (repo CloudControllerServiceBrokerRepository) FindByName(name string) (serviceBroker models.ServiceBroker, apiErr error) {
	foundBroker := false
	apiErr = repo.gateway.ListPaginatedResources(
		repo.config.ApiEndpoint(),
		fmt.Sprintf("/v2/service_brokers?q=%s", url.QueryEscape("name:"+name)),
		resources.ServiceBrokerResource{},
		func(resource interface{}) bool {
			serviceBroker = resource.(resources.ServiceBrokerResource).ToFields()
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
	return repo.gateway.CreateResource(path, strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Update(serviceBroker models.ServiceBroker) (apiErr error) {
	path := fmt.Sprintf("%s/v2/service_brokers/%s", repo.config.ApiEndpoint(), serviceBroker.Guid)
	body := fmt.Sprintf(
		`{"broker_url":"%s","auth_username":"%s","auth_password":"%s"}`,
		serviceBroker.Url, serviceBroker.Username, serviceBroker.Password,
	)
	return repo.gateway.UpdateResource(path, strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Rename(guid, name string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/service_brokers/%s", repo.config.ApiEndpoint(), guid)
	body := fmt.Sprintf(`{"name":"%s"}`, name)
	return repo.gateway.UpdateResource(path, strings.NewReader(body))
}

func (repo CloudControllerServiceBrokerRepository) Delete(guid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/service_brokers/%s", repo.config.ApiEndpoint(), guid)
	return repo.gateway.DeleteResource(path)
}
