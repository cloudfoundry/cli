package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

type RouterGroup ccv3.RouterGroup

func (actor Actor) GetRouterGroups() ([]RouterGroup, Warnings, error) {
	var routerGroups []RouterGroup

	ccv3RouterGroups, warnings, err := actor.CloudControllerClient.GetRouterGroups()
	if err != nil {
		return nil, Warnings(warnings), err
	}

	for _, group := range ccv3RouterGroups {
		routerGroups = append(routerGroups, RouterGroup(group))
	}

	return routerGroups, Warnings(warnings), err
}
