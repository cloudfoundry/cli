package api

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/net"
)

type routingApiRepository struct {
	config  core_config.Reader
	gateway net.Gateway
}

type RoutingApiRepository interface {
	ListRouterGroups(cb func(models.RouterGroup) bool) (apiErr error)
}

func NewRoutingApiRepository(config core_config.Reader, gateway net.Gateway) RoutingApiRepository {
	return routingApiRepository{
		config:  config,
		gateway: gateway,
	}
}

func (r routingApiRepository) ListRouterGroups(cb func(models.RouterGroup) bool) (apiErr error) {
	routerGroups := models.RouterGroups{}
	endpoint := fmt.Sprintf("%s/v1/router_groups", r.config.RoutingApiEndpoint())
	apiErr = r.gateway.GetResource(endpoint, &routerGroups)
	if apiErr != nil {
		return apiErr
	}

	for _, router := range routerGroups {
		if cb(router) == false {
			return
		}
	}
	return
}
