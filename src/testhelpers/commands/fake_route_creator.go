package commands

import "cf/models"

type FakeRouteCreator struct {
	CreateRouteHostname     string
	CreateRouteDomainFields models.DomainFields
	CreateRouteSpaceFields  models.SpaceFields
	ReservedRoute           models.Route
}

func (cmd *FakeRouteCreator) CreateRoute(hostName string, domain models.DomainFields, space models.SpaceFields) (reservedRoute models.Route, apiErr error) {
	cmd.CreateRouteHostname = hostName
	cmd.CreateRouteDomainFields = domain
	cmd.CreateRouteSpaceFields = space
	reservedRoute = cmd.ReservedRoute
	return
}
