package api

import (
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"cf/net"
	"fmt"
	"strings"
)

type ServiceBindingRepository interface {
	Create(instanceGuid, appGuid string) (apiResponse errors.Error)
	Delete(instance models.ServiceInstance, appGuid string) (found bool, apiResponse errors.Error)
}

type CloudControllerServiceBindingRepository struct {
	config  configuration.Reader
	gateway net.Gateway
}

func NewCloudControllerServiceBindingRepository(config configuration.Reader, gateway net.Gateway) (repo CloudControllerServiceBindingRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerServiceBindingRepository) Create(instanceGuid, appGuid string) (apiResponse errors.Error) {
	path := fmt.Sprintf("%s/v2/service_bindings", repo.config.ApiEndpoint())
	body := fmt.Sprintf(
		`{"app_guid":"%s","service_instance_guid":"%s","async":true}`,
		appGuid, instanceGuid,
	)
	return repo.gateway.CreateResource(path, repo.config.AccessToken(), strings.NewReader(body))
}

func (repo CloudControllerServiceBindingRepository) Delete(instance models.ServiceInstance, appGuid string) (found bool, apiResponse errors.Error) {
	var path string

	for _, binding := range instance.ServiceBindings {
		if binding.AppGuid == appGuid {
			path = repo.config.ApiEndpoint() + binding.Url
			break
		}
	}

	if path == "" {
		return
	} else {
		found = true
	}

	apiResponse = repo.gateway.DeleteResource(path, repo.config.AccessToken())
	return
}
