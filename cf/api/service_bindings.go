package api

import (
	"bytes"
	"encoding/json"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type ServiceBindingRepository interface {
	Create(instanceGuid, appGuid string, paramsMap map[string]interface{}) (apiErr error)
	Delete(instance models.ServiceInstance, appGuid string) (found bool, apiErr error)
}

type CloudControllerServiceBindingRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerServiceBindingRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerServiceBindingRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerServiceBindingRepository) Create(instanceGuid, appGuid string, paramsMap map[string]interface{}) (apiErr error) {
	path := "/v2/service_bindings"
	request := models.ServiceBindingRequest{
		AppGuid:             appGuid,
		ServiceInstanceGuid: instanceGuid,
		Params:              paramsMap,
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	return repo.gateway.CreateResource(repo.config.ApiEndpoint(), path, bytes.NewReader(jsonBytes))
}

func (repo CloudControllerServiceBindingRepository) Delete(instance models.ServiceInstance, appGuid string) (found bool, apiErr error) {
	var path string

	for _, binding := range instance.ServiceBindings {
		if binding.AppGuid == appGuid {
			path = binding.Url
			break
		}
	}

	if path == "" {
		return
	}

	found = true

	apiErr = repo.gateway.DeleteResource(repo.config.ApiEndpoint(), path)
	return
}
