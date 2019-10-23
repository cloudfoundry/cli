package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type Role ccv3.Role

func (actor Actor) CreateOrgRole(roleType constant.RoleType, userGUID string, orgGUID string) (Role, Warnings, error) {
	role, warnings, err := actor.CloudControllerClient.CreateRole(ccv3.Role{
		Type:     roleType,
		UserGUID: userGUID,
		OrgGUID:  orgGUID,
	})

	return Role(role), Warnings(warnings), err
}
