package v7action

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

type Role ccv3.Role

func (actor Actor) CreateOrgRoleByUserGUID(roleType constant.RoleType, userGUID string, orgGUID string) (Role, Warnings, error) {
	role, warnings, err := actor.CloudControllerClient.CreateRole(ccv3.Role{
		Type:     roleType,
		UserGUID: userGUID,
		OrgGUID:  orgGUID,
	})

	return Role(role), Warnings(warnings), err
}

func (actor Actor) CreateOrgRoleByUserName(roleType constant.RoleType, userName string, origin string, orgGUID string) (Role, Warnings, error) {
	role, warnings, err := actor.CloudControllerClient.CreateRole(ccv3.Role{
		Type:     roleType,
		UserName: userName,
		Origin:   origin,
		OrgGUID:  orgGUID,
	})

	return Role(role), Warnings(warnings), err
}

func (actor Actor) CreateSpaceRoleByUserGUID(roleType constant.RoleType, userGUID string, spaceGUID string) (Role, Warnings, error) {
	role, warnings, err := actor.CloudControllerClient.CreateRole(ccv3.Role{
		Type:      roleType,
		UserGUID:  userGUID,
		SpaceGUID: spaceGUID,
	})

	return Role(role), Warnings(warnings), err
}

func (actor Actor) CreateSpaceRoleByUserName(roleType constant.RoleType, userName string, origin string, spaceGUID string) (Role, Warnings, error) {
	role, warnings, err := actor.CloudControllerClient.CreateRole(ccv3.Role{
		Type:      roleType,
		UserName:  userName,
		Origin:    origin,
		SpaceGUID: spaceGUID,
	})

	return Role(role), Warnings(warnings), err
}
