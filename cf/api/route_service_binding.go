package api

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/net"
)

//go:generate counterfeiter -o fakes/fake_route_service_binding_repository.go . RouteServiceBindingRepository
type RouteServiceBindingRepository interface {
	Bind(instanceGuid, routeGuid string, upsi bool) (apiErr error)
	Unbind(instanceGuid, routeGuid string, upsi bool) (apiErr error)
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
	path := getPath(instanceGuid, routeGuid, upsi)
	return repo.gateway.UpdateResourceSync(repo.config.ApiEndpoint(), path, strings.NewReader(""))
}

func (repo CloudControllerRouteServiceBindingRepository) Unbind(instanceGuid, routeGuid string, upsi bool) (apiErr error) {
	path := getPath(instanceGuid, routeGuid, upsi)
	return repo.gateway.DeleteResource(repo.config.ApiEndpoint(), path)
}

func getPath(instanceGuid, routeGuid string, upsi bool) (path string) {
	resource := "service_instances"
	if upsi {
		resource = "user_provided_service_instances"
	}
	path = fmt.Sprintf("/v2/%s/%s/routes/%s", resource, instanceGuid, routeGuid)
	return
}
