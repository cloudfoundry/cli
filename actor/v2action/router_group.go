package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"

	"code.cloudfoundry.org/cli/api/router"
	"code.cloudfoundry.org/cli/api/router/routererror"
)

type RouterGroup router.RouterGroup

func (actor Actor) GetRouterGroupByName(routerGroupName string, client RouterClient) (RouterGroup, error) {
	routerGroup, err := client.GetRouterGroupByName(routerGroupName)
	if err != nil {
		if _, ok := err.(routererror.ResourceNotFoundError); ok {
			return RouterGroup{}, actionerror.RouterGroupNotFoundError{Name: routerGroupName}
		}
		return RouterGroup{}, err
	}

	return RouterGroup(routerGroup), nil
}
