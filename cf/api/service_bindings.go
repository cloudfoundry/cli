package api

import (
	"bytes"
	"encoding/json"

	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

//go:generate counterfeiter . ServiceBindingRepository

type ServiceBindingRepository interface {
	Create(instanceGUID, appGUID string, paramsMap map[string]interface{}) (apiErr error)
	Delete(instance models.ServiceInstance, appGUID string) (found bool, apiErr error)
}

type CloudControllerServiceBindingRepository struct {
	config  coreconfig.Reader
	gateway net.Gateway
}

func NewCloudControllerServiceBindingRepository(config coreconfig.Reader, gateway net.Gateway) (repo CloudControllerServiceBindingRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerServiceBindingRepository) Create(instanceGUID, appGUID string, paramsMap map[string]interface{}) (apiErr error) {
	path := "/v2/service_bindings"
	request := models.ServiceBindingRequest{
		AppGUID:             appGUID,
		ServiceInstanceGUID: instanceGUID,
		Params:              paramsMap,
	}

	jsonBytes, err := json.Marshal(request)
	if err != nil {
		return err
	}

	return repo.gateway.CreateResource(repo.config.APIEndpoint(), path, bytes.NewReader(jsonBytes))
}

func (repo CloudControllerServiceBindingRepository) Delete(instance models.ServiceInstance, appGUID string) (found bool, apiErr error) {
	var path string

	for _, binding := range instance.ServiceBindings {
		if binding.AppGUID == appGUID {
			path = binding.URL
			break
		}
	}

	if path == "" {
		return
	}

	found = true

	apiErr = repo.gateway.DeleteResource(repo.config.APIEndpoint(), path)
	return
}
