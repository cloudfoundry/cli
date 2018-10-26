package v2action

import "code.cloudfoundry.org/cli/api/router"

//go:generate counterfeiter . RouterClient

// RouterClient is a Router API client.
type RouterClient interface {
	GetRouterGroupsByName(string) ([]router.RouterGroup, error)
}
