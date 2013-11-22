package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type ServiceBindingRepository interface {
	Create(instanceGuid, appGuid string) (apiResponse net.ApiResponse)
	Delete(instance cf.ServiceInstance, appGuid string) (found bool, apiResponse net.ApiResponse)
}

type CloudControllerServiceBindingRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerServiceBindingRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerServiceBindingRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerServiceBindingRepository) Create(instanceGuid, appGuid string) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_bindings", repo.config.Target)
	body := fmt.Sprintf(
		`{"app_guid":"%s","service_instance_guid":"%s"}`,
		appGuid, instanceGuid,
	)
	return repo.gateway.CreateResource(path, repo.config.AccessToken, strings.NewReader(body))
}

func (repo CloudControllerServiceBindingRepository) Delete(instance cf.ServiceInstance, appGuid string) (found bool, apiResponse net.ApiResponse) {
	var path string

	for _, binding := range instance.ServiceBindings {
		if binding.AppGuid == appGuid {
			path = repo.config.Target + binding.Url
			break
		}
	}

	if path == "" {
		return
	} else {
		found = true
	}

	apiResponse = repo.gateway.DeleteResource(path, repo.config.AccessToken)
	return
}
