package api

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/net"
)

type RouteServiceBindingRepository interface {
	Bind(instanceGuid, routeGuid string, upsi bool) (apiErr error)
}

type CloudControllerRouteServiceBindingRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerRouteServiceBindingRepository(config core_config.Reader, gateway net.Gateway) (repo CloudControllerRouteServiceBindingRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerRouteServiceBindingRepository) Bind(instanceGuid, routeGuid string, upsi bool) (apiErr error) {
	resource := "service_instances"

	if upsi {
		resource = "user_provided_service_instances"
	}

	path := fmt.Sprintf("/v2/%s/%s/routes/%s", resource, instanceGuid, routeGuid)

	return repo.gateway.UpdateResourceSync(repo.config.ApiEndpoint(), path, strings.NewReader(""))
}
