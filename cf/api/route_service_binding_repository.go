package api

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/net"
)

//go:generate counterfeiter -o fakes/fake_route_service_binding_repository.go . RouteServiceBindingRepository
type RouteServiceBindingRepository interface {
	Bind(instanceGuid, routeGuid string, userProvided bool) error
	Unbind(instanceGuid, routeGuid string, userProvided bool) error
}

type CloudControllerRouteServiceBindingRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

func NewCloudControllerRouteServiceBindingRepository(config core_config.Reader, gateway net.Gateway) CloudControllerRouteServiceBindingRepository {
	return CloudControllerRouteServiceBindingRepository{
		config:  config,
		gateway: gateway,
	}
}

func (repo CloudControllerRouteServiceBindingRepository) Bind(instanceGuid, routeGuid string, userProvided bool) error {
	path := getPath(instanceGuid, routeGuid, userProvided)
	return repo.gateway.UpdateResourceSync(repo.config.ApiEndpoint(), path, strings.NewReader(""))
}

func (repo CloudControllerRouteServiceBindingRepository) Unbind(instanceGuid, routeGuid string, userProvided bool) error {
	path := getPath(instanceGuid, routeGuid, userProvided)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), path)
}

func getPath(instanceGuid, routeGuid string, userProvided bool) string {
	var resource string
	if userProvided {
		resource = "user_provided_service_instances"
	} else {
		resource = "service_instances"
	}

	return fmt.Sprintf("/v2/%s/%s/routes/%s", resource, instanceGuid, routeGuid)
}
