package v7action

import "code.cloudfoundry.org/cli/v8/api/router"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . RoutingClient

type RoutingClient interface {
	GetRouterGroups() ([]router.RouterGroup, error)
	GetRouterGroupByName(name string) (router.RouterGroup, error)
}
