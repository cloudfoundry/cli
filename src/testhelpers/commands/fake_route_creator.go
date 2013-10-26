package commands

import (
	"cf"
	"cf/net"
)

type FakeRouteCreator struct {
	CreateRouteHostname string
	CreateRouteDomain cf.Domain
	CreateRouteSpace cf.Space
	ReservedRoute cf.Route
}

func (cmd *FakeRouteCreator) CreateRoute(hostName string, domain cf.Domain, space cf.Space) (reservedRoute cf.Route, apiResponse net.ApiResponse) {
	cmd.CreateRouteHostname = hostName
	cmd.CreateRouteDomain = domain
	cmd.CreateRouteSpace = space
	reservedRoute = cmd.ReservedRoute
	return
}
