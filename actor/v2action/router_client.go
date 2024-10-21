package v2action

import "code.cloudfoundry.org/cli/v7/api/router"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . RouterClient

// RouterClient is a Router API client.
type RouterClient interface {
	GetRouterGroupByName(string) (router.RouterGroup, error)
}
