package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type ServiceKeyRepository interface {
	CreateServiceKey(serviceKeyGuid string, keyName string, params map[string]interface{}) error
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

func (c CloudControllerServiceKeyRepository) CreateServiceKey(instanceGuid string, keyName string, params map[string]interface{}) error {
	path := "/v2/service_keys"

	request := models.ServiceKeyRequest{
		Name:                keyName,
		ServiceInstanceGuid: instanceGuid,
		Params:              params,
	}
	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	err = c.gateway.CreateResource(c.config.ApiEndpoint(), path, bytes.NewReader(jsonBytes))

	if httpErr, ok := err.(errors.HttpError); ok {
		switch httpErr.ErrorCode() {
		case errors.SERVICE_KEY_NAME_TAKEN:
			return errors.NewModelAlreadyExistsError("Service key", keyName)
		case errors.UNBINDABLE_SERVICE:
			return errors.NewUnbindableServiceError()
		default:
			return errors.New(httpErr.Error())
		}
	}

	return nil
}

func (c CloudControllerServiceKeyRepository) ListServiceKeys(instanceGuid string) ([]models.ServiceKey, error) {
	path := fmt.Sprintf("/v2/service_instances/%s/service_keys", instanceGuid)

	return c.listServiceKeys(path)
}

func (c CloudControllerServiceKeyRepository) GetServiceKey(instanceGuid string, keyName string) (models.ServiceKey, error) {
	path := fmt.Sprintf("/v2/service_instances/%s/service_keys?q=%s", instanceGuid, url.QueryEscape("name:"+keyName))

	serviceKeys, err := c.listServiceKeys(path)
	if err != nil || len(serviceKeys) == 0 {
		return models.ServiceKey{}, err
	}

	return serviceKeys[0], nil
}

func (c CloudControllerServiceKeyRepository) listServiceKeys(path string) ([]models.ServiceKey, error) {
	serviceKeys := []models.ServiceKey{}
	err := c.gateway.ListPaginatedResources(
		c.config.ApiEndpoint(),
		path,
		resources.ServiceKeyResource{},
		func(resource interface{}) bool {
			serviceKey := resource.(resources.ServiceKeyResource).ToModel()
			serviceKeys = append(serviceKeys, serviceKey)
			return true
		})

	if err != nil {
		if httpErr, ok := err.(errors.HttpError); ok && httpErr.ErrorCode() == errors.NOT_AUTHORIZED {
			return []models.ServiceKey{}, errors.NewNotAuthorizedError()
		}
		return []models.ServiceKey{}, err
	}

	return serviceKeys, nil
}

func (c CloudControllerServiceKeyRepository) DeleteServiceKey(serviceKeyGuid string) error {
	path := fmt.Sprintf("/v2/service_keys/%s", serviceKeyGuid)
	return c.gateway.DeleteResource(c.config.ApiEndpoint(), path)
}
