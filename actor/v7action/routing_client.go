package v7action

import "code.cloudfoundry.org/cli/api/router"

//go:generate counterfeiter . RoutingClient

type RoutingClient interface {
	GetRouterGroups() ([]router.RouterGroup, error)
	GetRouterGroupByName(name string) (router.RouterGroup, error)
}
