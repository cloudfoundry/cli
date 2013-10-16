package api

import (
	"cf"
	"cf/configuration"
	"cf/net"
	"fmt"
	"strings"
)

type ServiceBindingRepository interface {
	Create(instance cf.ServiceInstance, app cf.Application) (apiResponse net.ApiResponse)
	Delete(instance cf.ServiceInstance, app cf.Application) (found bool, apiResponse net.ApiResponse)
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

func (repo CloudControllerServiceBindingRepository) Create(instance cf.ServiceInstance, app cf.Application) (apiResponse net.ApiResponse) {
	path := fmt.Sprintf("%s/v2/service_bindings", repo.config.Target)
	body := fmt.Sprintf(
		`{"app_guid":"%s","service_instance_guid":"%s"}`,
		app.Guid, instance.Guid,
	)
	request, apiResponse := repo.gateway.NewRequest("POST", path, repo.config.AccessToken, strings.NewReader(body))
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}

func (repo CloudControllerServiceBindingRepository) Delete(instance cf.ServiceInstance, app cf.Application) (found bool, apiResponse net.ApiResponse) {
	var path string

	for _, binding := range instance.ServiceBindings {
		if binding.AppGuid == app.Guid {
			path = repo.config.Target + binding.Url
			break
		}
	}

	if path == "" {
		return
	} else {
		found = true
	}

	request, apiResponse := repo.gateway.NewRequest("DELETE", path, repo.config.AccessToken, nil)
	if apiResponse.IsNotSuccessful() {
		return
	}

	apiResponse = repo.gateway.PerformRequest(request)
	return
}
