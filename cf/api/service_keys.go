package api

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type ServiceKeyRepository interface {
	CreateServiceKey(instanceId string, keyName string) error
	ListServiceKeys(instanceId string) ([]models.ServiceKey, error)
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

func (c CloudControllerServiceKeyRepository) CreateServiceKey(instanceId string, keyName string) error {
	path := "/v2/service_keys"
	data := fmt.Sprintf(`{"service_instance_guid":"%s","name":"%s"}`, instanceId, keyName)

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

func (c CloudControllerServiceKeyRepository) ListServiceKeys(instanceId string) ([]models.ServiceKey, error) {
	path := fmt.Sprintf("/v2/service_keys?q=service_instance_guid:%s", instanceId)

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
