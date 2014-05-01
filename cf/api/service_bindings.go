package api

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
	"strings"
)

type ServiceBindingRepository interface {
	Create(instanceGuid, appGuid string) (apiErr error)
	Delete(instance models.ServiceInstance, appGuid string) (found bool, apiErr error)
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

func (repo CloudControllerServiceBindingRepository) Create(instanceGuid, appGuid string) (apiErr error) {
	path := fmt.Sprintf("%s/v2/service_bindings", repo.config.ApiEndpoint())
	body := fmt.Sprintf(
		`{"app_guid":"%s","service_instance_guid":"%s","async":true}`,
		appGuid, instanceGuid,
	)
	return repo.gateway.CreateResource(path, strings.NewReader(body))
}

func (repo CloudControllerServiceBindingRepository) Delete(instance models.ServiceInstance, appGuid string) (found bool, apiErr error) {
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

	apiErr = repo.gateway.DeleteResource(path)
	return
}
