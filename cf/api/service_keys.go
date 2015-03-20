package api

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type ServiceKeyRepository interface {
	CreateServiceKey(serviceKeyGuid string, keyName string) error
	ListServiceKeys(serviceKeyGuid string) ([]models.ServiceKey, error)
	GetServiceKey(serviceKeyGuid string, keyName string) (models.ServiceKey, error)
	DeleteServiceKey(serviceKeyGuid string) error
}

type CloudControllerServiceKeyRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerServiceKeyRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerServiceKeyRepository) {
	return CloudControllerServiceKeyRepository{
		config:  config,
		gateway: gateway,
	}
}

func (c CloudControllerServiceKeyRepository) CreateServiceKey(instanceGuid string, keyName string) error {
	path := "/v2/service_keys"
	data := fmt.Sprintf(`{"service_instance_guid":"%s","name":"%s"}`, instanceGuid, keyName)

	err := c.gateway.CreateResource(c.config.ApiEndpoint(), path, strings.NewReader(data))

	if httpErr, ok := err.(errors.HttpError); ok && httpErr.ErrorCode() == errors.SERVICE_KEY_NAME_TAKEN {
		return errors.NewModelAlreadyExistsError("Service key", keyName)
	} else if httpErr, ok := err.(errors.HttpError); ok && httpErr.ErrorCode() == errors.UNBINDABLE_SERVICE {
		return errors.NewUnbindableServiceError()
	} else if httpErr, ok := err.(errors.HttpError); ok && httpErr.ErrorCode() != "" {
		return errors.New(httpErr.Error())
	}

	return nil
}

func (c CloudControllerServiceKeyRepository) ListServiceKeys(instanceGuid string) ([]models.ServiceKey, error) {
	path := fmt.Sprintf("/v2/service_keys?q=service_instance_guid:%s", instanceGuid)

	return c.listServiceKeys(path)
}

func (c CloudControllerServiceKeyRepository) GetServiceKey(instanceId string, keyName string) (models.ServiceKey, error) {
	path := fmt.Sprintf("/v2/service_keys?q=%s", url.QueryEscape("service_instance_guid:"+instanceId+";name:"+keyName))

	serviceKeys, err := c.listServiceKeys(path)
	if err != nil || len(serviceKeys) == 0 {
		return models.ServiceKey{}, err
	}

	return serviceKeys[0], nil
}

func (c CloudControllerServiceKeyRepository) listServiceKeys(path string) ([]models.ServiceKey, error) {
	serviceKeys := []models.ServiceKey{}
	apiErr := c.gateway.ListPaginatedResources(
		c.config.ApiEndpoint(),
		path,
		resources.ServiceKeyResource{},
		func(resource interface{}) bool {
			serviceKey := resource.(resources.ServiceKeyResource).ToModel()
			serviceKeys = append(serviceKeys, serviceKey)
			return true
		})

	if apiErr != nil {
		return []models.ServiceKey{}, apiErr
	}

	return serviceKeys, nil
}

func (c CloudControllerServiceKeyRepository) DeleteServiceKey(serviceKeyGuid string) error {
	path := fmt.Sprintf("/v2/service_keys/%s", serviceKeyGuid)
	return c.gateway.DeleteResource(c.config.ApiEndpoint(), path)
}
