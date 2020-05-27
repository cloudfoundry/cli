package v7action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/router"
	"code.cloudfoundry.org/cli/api/router/routererror"
)

type RouterGroup router.RouterGroup

func (actor Actor) GetRouterGroups() ([]RouterGroup, error) {
	var routerGroups []RouterGroup

	apiRouterGroups, err := actor.RoutingClient.GetRouterGroups()
	if err != nil {
		return nil, err
	}

	for _, group := range apiRouterGroups {
		routerGroups = append(routerGroups, RouterGroup(group))
	}

	return routerGroups, err
}

func (actor Actor) GetRouterGroupByName(name string) (RouterGroup, error) {
	apiRouterGroup, err := actor.RoutingClient.GetRouterGroupByName(name)
	if err != nil {
		if _, ok := err.(routererror.ResourceNotFoundError); ok {
			return RouterGroup{}, actionerror.RouterGroupNotFoundError{Name: name}
		}

		return RouterGroup{}, err
	}

	return RouterGroup(apiRouterGroup), nil
}
