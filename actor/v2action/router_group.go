package v2action

import (
	"code.cloudfoundry.org/cli/actor/actionerror"

	"code.cloudfoundry.org/cli/api/router"
	"code.cloudfoundry.org/cli/api/router/routererror"
)

type RouterGroup router.RouterGroup

func (actor Actor) GetRouterGroupByName(routerGroupName string, client RouterClient) (RouterGroup, error) {
	routerGroups, err := client.GetRouterGroupsByName(routerGroupName)
	if err != nil {
		if rErr, ok := err.(routererror.ErrorResponse); ok {
			if rErr.Name == "ResourceNotFoundError" {
				return RouterGroup{}, actionerror.RouterGroupNotFoundError{Name: routerGroupName}
			}
		}
		return RouterGroup{}, err
	}

	for _, routerGroup := range routerGroups {
		if routerGroup.Name == routerGroupName {
			return RouterGroup(routerGroup), nil
		}
	}
	return RouterGroup{}, actionerror.RouterGroupNotFoundError{Name: routerGroupName}
}
