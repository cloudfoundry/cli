package api

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/net"
)

type ServiceKeyRepository interface {
	CreateServiceKey(instanceId string, keyName string) (apiErr error)
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

func (c CloudControllerServiceKeyRepository) CreateServiceKey(instanceId string, keyName string) (apiErr error) {
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
