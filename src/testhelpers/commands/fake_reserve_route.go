package commands

import (
	"cf"
	"cf/net"
)

type FakeReserveRoute struct {
	CreateRouteHostname string
	CreateRouteDomain cf.Domain
	CreateRouteSpace cf.Space
	ReservedRoute cf.Route
}

func (cmd *FakeReserveRoute) CreateRoute(hostName string, domain cf.Domain, space cf.Space) (reservedRoute cf.Route, apiResponse net.ApiResponse) {
	cmd.CreateRouteHostname = hostName
	cmd.CreateRouteDomain = domain
	cmd.CreateRouteSpace = space
	reservedRoute = cmd.ReservedRoute
	return
}
