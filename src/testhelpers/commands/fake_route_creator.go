package commands

import (
	"cf"
	"cf/net"
)

type FakeRouteCreator struct {
	CreateRouteHostname string
	CreateRouteDomainFields cf.DomainFields
	CreateRouteSpaceFields cf.SpaceFields
	ReservedRoute cf.Route
}

func (cmd *FakeRouteCreator) CreateRoute(hostName string, domain cf.DomainFields, space cf.SpaceFields) (reservedRoute cf.Route, apiResponse net.ApiResponse) {
	cmd.CreateRouteHostname = hostName
	cmd.CreateRouteDomainFields = domain
	cmd.CreateRouteSpaceFields = space
	reservedRoute = cmd.ReservedRoute
	return
}
